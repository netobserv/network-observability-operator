package flp

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
)

const (
	configVolume            = "config-volume"
	configPath              = "/etc/flowlogs-pipeline"
	configFile              = "config.json"
	healthServiceName       = "health"
	prometheusServiceName   = "prometheus"
	profilePortName         = "pprof"
	healthTimeoutSeconds    = 5
	livenessPeriodSeconds   = 10
	startupFailureThreshold = 5
	startupPeriodSeconds    = 10
	appLabel                = "app"
)

type Builder struct {
	info        *reconcilers.Instance
	appName     string
	labels      map[string]string
	selector    map[string]string
	desired     *flowslatest.FlowCollectorSpec
	flowMetrics *metricslatest.FlowMetricList
	promTLS     *flowslatest.CertificateReference
	volumes     volumes.Builder
	loki        *helper.LokiConfig
	pipeline    *PipelineBuilder
}

type builder = Builder

func newInProcessBuilder(info *reconcilers.Instance, appName string, desired *flowslatest.FlowCollectorSpec, flowMetrics *metricslatest.FlowMetricList) (*Builder, error) {
	b, err := newBuilder(info, appName, desired, flowMetrics)
	if err != nil {
		return nil, err
	}

	pipeline := b.createInProcessPipeline()
	if err = pipeline.AddProcessorStages(); err != nil {
		return nil, err
	}
	return &b, nil
}

func newKafkaConsumerBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec, flowMetrics *metricslatest.FlowMetricList) (*Builder, error) {
	b, err := newBuilder(info, constants.FLPName, desired, flowMetrics)
	if err != nil {
		return nil, err
	}
	pipeline := b.createKafkaPipeline()
	if err = pipeline.AddProcessorStages(); err != nil {
		return nil, err
	}
	return &b, nil
}

func newBuilder(info *reconcilers.Instance, appName string, desired *flowslatest.FlowCollectorSpec, flowMetrics *metricslatest.FlowMetricList) (Builder, error) {
	version := helper.ExtractVersion(info.Image)
	var promTLS *flowslatest.CertificateReference
	switch desired.Processor.Metrics.Server.TLS.Type {
	case flowslatest.ServerTLSProvided:
		promTLS = desired.Processor.Metrics.Server.TLS.Provided
		if promTLS == nil {
			return builder{}, fmt.Errorf("processor tls configuration set to provided but none is provided")
		}
	case flowslatest.ServerTLSAuto:
		promTLS = &flowslatest.CertificateReference{
			Type:     "secret",
			Name:     appName,
			CertFile: "tls.crt",
			CertKey:  "tls.key",
		}
	case flowslatest.ServerTLSDisabled:
		// nothing to do there
	}
	return builder{
		info:    info,
		appName: appName,
		labels: map[string]string{
			appLabel:  appName,
			"version": helper.MaxLabelLength(version),
		},
		selector: map[string]string{
			appLabel: appName,
		},
		desired:     desired,
		flowMetrics: flowMetrics,
		promTLS:     promTLS,
		loki:        info.Loki,
	}, nil
}

func promServiceName(appName string) string    { return appName + "-prom" }
func configMapName(appName string) string      { return appName + "-config" }
func serviceMonitorName(appName string) string { return appName + "-monitor" }
func prometheusRuleName(appName string) string { return appName + "-alert" }
func (b *builder) promServiceName() string     { return promServiceName(b.appName) }
func (b *builder) configMapName() string       { return configMapName(b.appName) }
func (b *builder) serviceMonitorName() string  { return serviceMonitorName(b.appName) }
func (b *builder) prometheusRuleName() string  { return prometheusRuleName(b.appName) }

func (b *builder) Pipeline() *PipelineBuilder { return b.pipeline }

func (b *builder) createKafkaPipeline() PipelineBuilder {
	decoder := api.Decoder{Type: "protobuf"}
	return b.initPipeline(config.NewKafkaPipeline("kafka-read", api.IngestKafka{
		Brokers:           []string{b.desired.Kafka.Address},
		Topic:             b.desired.Kafka.Topic,
		GroupId:           b.appName, // Without groupid, each message is delivered to each consumers
		Decoder:           decoder,
		TLS:               getKafkaTLS(&b.desired.Kafka.TLS, "kafka-cert", &b.volumes),
		SASL:              getKafkaSASL(&b.desired.Kafka.SASL, "kafka-ingest", &b.volumes),
		PullQueueCapacity: b.desired.Processor.KafkaConsumerQueueCapacity,
		PullMaxBytes:      b.desired.Processor.KafkaConsumerBatchSize,
	}))
}

func (b *builder) createInProcessPipeline() PipelineBuilder {
	return b.initPipeline(config.NewPresetIngesterPipeline())
}

func (b *builder) initPipeline(ingest config.PipelineBuilderStage) PipelineBuilder {
	pipeline := newPipelineBuilder(b.desired, b.flowMetrics, b.info.Loki, b.info.ClusterID, &b.volumes, &ingest)
	b.pipeline = &pipeline
	return pipeline
}

func (b *builder) podTemplate(annotations map[string]string) corev1.PodTemplateSpec {
	debugConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	var ports []corev1.ContainerPort

	ports = append(ports, corev1.ContainerPort{
		Name:          healthServiceName,
		ContainerPort: *debugConfig.HealthPort,
	})

	ports = append(ports, corev1.ContainerPort{
		Name:          prometheusServiceName,
		ContainerPort: b.desired.Processor.Metrics.Server.Port,
	})

	if debugConfig.ProfilePort != nil {
		ports = append(ports, corev1.ContainerPort{
			Name:          profilePortName,
			ContainerPort: *debugConfig.ProfilePort,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	volumeMounts := b.volumes.AppendMounts([]corev1.VolumeMount{{
		MountPath: configPath,
		Name:      configVolume,
	}})
	volumes := b.volumes.AppendVolumes([]corev1.Volume{{
		Name: configVolume,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: b.configMapName(),
				},
			},
		},
	}})

	var envs []corev1.EnvVar
	debugConfig = helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	// we need to sort env map to keep idempotency,
	// as equal maps could be iterated in different order
	for _, pair := range helper.KeySorted(debugConfig.Env) {
		envs = append(envs, corev1.EnvVar{Name: pair[0], Value: pair[1]})
	}
	envs = append(envs, constants.EnvNoHTTP2)

	container := corev1.Container{
		Name:            constants.FLPName,
		Image:           b.info.Image,
		ImagePullPolicy: corev1.PullPolicy(b.desired.Processor.ImagePullPolicy),
		Args:            []string{fmt.Sprintf(`--config=%s/%s`, configPath, configFile)},
		Resources:       *b.desired.Processor.Resources.DeepCopy(),
		VolumeMounts:    volumeMounts,
		Ports:           ports,
		Env:             envs,
		SecurityContext: helper.ContainerDefaultSecurityContext(),
	}
	if *debugConfig.EnableKubeProbes {
		container.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/live",
					Port: intstr.FromString(healthServiceName),
				},
			},
			TimeoutSeconds: healthTimeoutSeconds,
			PeriodSeconds:  livenessPeriodSeconds,
		}
		container.StartupProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/ready",
					Port: intstr.FromString(healthServiceName),
				},
			},
			TimeoutSeconds:   healthTimeoutSeconds,
			PeriodSeconds:    startupPeriodSeconds,
			FailureThreshold: startupFailureThreshold,
		}
	}
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/scrape_port"] = fmt.Sprint(b.desired.Processor.Metrics.Server.Port)

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      b.labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			Volumes:            volumes,
			Containers:         []corev1.Container{container},
			ServiceAccountName: b.appName,
			DNSPolicy:          corev1.DNSClusterFirst,
		},
	}
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap() (*corev1.ConfigMap, string, error) {
	configStr, err := b.GetJSONConfig()
	if err != nil {
		return nil, "", err
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.configMapName(),
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Data: map[string]string{
			configFile: configStr,
		},
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(configStr))
	digest := strconv.FormatUint(hasher.Sum64(), 36)
	return &configMap, digest, nil
}

func (b *builder) GetJSONConfig() (string, error) {
	metricsSettings := config.MetricsSettings{
		PromConnectionInfo: api.PromConnectionInfo{
			Port: int(b.desired.Processor.Metrics.Server.Port),
		},
		Prefix:  "netobserv_",
		NoPanic: true,
	}
	if b.desired.Processor.Metrics.Server.TLS.Type != flowslatest.ServerTLSDisabled {
		cert, key := b.volumes.AddCertificate(b.promTLS, "prom-certs")
		if cert != "" && key != "" {
			metricsSettings.TLS = &api.PromTLSConf{
				CertPath: cert,
				KeyPath:  key,
			}
		}
	}
	debugConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	config := map[string]interface{}{
		"log-level": b.desired.Processor.LogLevel,
		"health": map[string]interface{}{
			"port": *debugConfig.HealthPort,
		},
		"pipeline":        b.pipeline.GetStages(),
		"parameters":      b.pipeline.GetStageParams(),
		"metricsSettings": metricsSettings,
	}
	if debugConfig.ProfilePort != nil {
		config["profile"] = map[string]interface{}{
			"port": *debugConfig.ProfilePort,
		}
	}

	bs, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func (b *builder) promService() *corev1.Service {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.promServiceName(),
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: b.selector,
			Ports: []corev1.ServicePort{{
				Name:     prometheusServiceName,
				Port:     b.desired.Processor.Metrics.Server.Port,
				Protocol: corev1.ProtocolTCP,
				// Some Kubernetes versions might automatically set TargetPort to Port. We need to
				// explicitly set it here so the reconcile loop verifies that the owned service
				// is equal as the desired service
				TargetPort: intstr.FromInt32(b.desired.Processor.Metrics.Server.Port),
			}},
		},
	}
	if b.desired.Processor.Metrics.Server.TLS.Type == flowslatest.ServerTLSAuto {
		svc.ObjectMeta.Annotations = map[string]string{
			constants.OpenShiftCertificateAnnotation: b.promServiceName(),
		}
	}
	return &svc
}

func (b *builder) serviceMonitor() *monitoringv1.ServiceMonitor {
	serverName := fmt.Sprintf("%s.%s.svc", b.promServiceName(), b.info.Namespace)
	flpServiceMonitorObject := monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.serviceMonitorName(),
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     prometheusServiceName,
					Interval: "15s",
					Scheme:   "http",
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					b.info.Namespace,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: b.selector,
			},
		},
	}
	if b.desired.Processor.Metrics.Server.TLS.Type == flowslatest.ServerTLSAuto {
		flpServiceMonitorObject.Spec.Endpoints[0].Scheme = "https"
		flpServiceMonitorObject.Spec.Endpoints[0].TLSConfig = &monitoringv1.TLSConfig{
			SafeTLSConfig: monitoringv1.SafeTLSConfig{
				ServerName: serverName,
			},
			CAFile: "/etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt",
		}
	}

	if b.desired.Processor.Metrics.Server.TLS.Type == flowslatest.ServerTLSProvided {
		flpServiceMonitorObject.Spec.Endpoints[0].Scheme = "https"
		flpServiceMonitorObject.Spec.Endpoints[0].TLSConfig = &monitoringv1.TLSConfig{
			SafeTLSConfig: monitoringv1.SafeTLSConfig{
				ServerName:         serverName,
				InsecureSkipVerify: b.desired.Processor.Metrics.Server.TLS.InsecureSkipVerify,
			},
		}
		if !b.desired.Processor.Metrics.Server.TLS.InsecureSkipVerify &&
			b.desired.Processor.Metrics.Server.TLS.ProvidedCaFile != nil &&
			b.desired.Processor.Metrics.Server.TLS.ProvidedCaFile.File != "" {
			flpServiceMonitorObject.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.CA = helper.GetSecretOrConfigMap(b.desired.Processor.Metrics.Server.TLS.ProvidedCaFile)
		}
	}

	return &flpServiceMonitorObject
}

func shouldAddAlert(name flowslatest.FLPAlert, disabledList []flowslatest.FLPAlert) bool {
	for _, disabledAlert := range disabledList {
		if name == disabledAlert {
			return false
		}
	}
	return true
}

func (b *builder) prometheusRule() *monitoringv1.PrometheusRule {
	rules := []monitoringv1.Rule{}
	d := monitoringv1.Duration("10m")

	// Not receiving flows
	if shouldAddAlert(flowslatest.AlertNoFlows, b.desired.Processor.Metrics.DisableAlerts) {
		rules = append(rules, monitoringv1.Rule{
			Alert: string(flowslatest.AlertNoFlows),
			Annotations: map[string]string{
				"description": "NetObserv flowlogs-pipeline is not receiving any flow, this is either a connection issue with the agent, or an agent issue",
				"summary":     "NetObserv flowlogs-pipeline is not receiving any flow",
			},
			Expr: intstr.FromString("sum(rate(netobserv_ingest_flows_processed[1m])) == 0"),
			For:  &d,
			Labels: map[string]string{
				"severity": "warning",
				appLabel:   "netobserv",
			},
		})
	}

	// Flows getting dropped by loki library
	if shouldAddAlert(flowslatest.AlertLokiError, b.desired.Processor.Metrics.DisableAlerts) {
		rules = append(rules, monitoringv1.Rule{
			Alert: string(flowslatest.AlertLokiError),
			Annotations: map[string]string{
				"description": "NetObserv flowlogs-pipeline is dropping flows because of loki errors, loki may be down or having issues ingesting every flows. Please check loki and flowlogs-pipeline logs.",
				"summary":     "NetObserv flowlogs-pipeline is dropping flows because of loki errors",
			},
			Expr: intstr.FromString("sum(rate(netobserv_loki_dropped_entries_total[1m])) > 0"),
			For:  &d,
			Labels: map[string]string{
				"severity": "warning",
				appLabel:   "netobserv",
			},
		})
	}

	flpPrometheusRuleObject := monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.prometheusRuleName(),
			Labels:    b.labels,
			Namespace: b.info.Namespace,
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name:  "NetobservFlowLogsPipeline",
					Rules: rules,
				},
			},
		},
	}
	return &flpPrometheusRuleObject
}

func (b *Builder) deployment(annotations map[string]string) *appsv1.Deployment {
	pod := b.podTemplate(annotations)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.appName,
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: b.desired.Processor.KafkaConsumerReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.selector,
			},
			Template: pod,
		},
	}
}

func (b *Builder) autoScaler() *ascv2.HorizontalPodAutoscaler {
	return &ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.appName,
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       b.appName,
			},
			MinReplicas: b.desired.Processor.KafkaConsumerAutoscaler.MinReplicas,
			MaxReplicas: b.desired.Processor.KafkaConsumerAutoscaler.MaxReplicas,
			Metrics:     b.desired.Processor.KafkaConsumerAutoscaler.Metrics,
		},
	}
}

// The operator needs to have at least the same permissions as flowlogs-pipeline in order to grant them
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=pods;services;nodes,verbs=get;list;watch

func rbacInfo(appName, saName, saNamespace string) (*corev1.ServiceAccount, *rbacv1.ClusterRole, *rbacv1.ClusterRoleBinding) {
	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: saNamespace,
			Labels: map[string]string{
				appLabel: appName,
			},
		},
	}
	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.FLPName,
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Verbs:     []string{"list", "get", "watch"},
			Resources: []string{"pods", "services", "nodes"},
		}, {
			APIGroups: []string{"apps"},
			Verbs:     []string{"list", "get", "watch"},
			Resources: []string{"replicasets"},
		}},
	}
	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.FLPName,
			Labels: map[string]string{
				appLabel: appName,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     cr.Name,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      saName,
			Namespace: saNamespace,
		}},
	}
	return &sa, &cr, &crb
}
