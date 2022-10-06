package flowlogspipeline

import (
	"embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"strconv"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/confgen"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	promConfig "github.com/prometheus/common/config"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/filters"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

const (
	configVolume            = "config-volume"
	configPath              = "/etc/flowlogs-pipeline"
	configFile              = "config.json"
	metricsConfigDir        = "metrics_definitions"
	kafkaCerts              = "kafka-certs"
	lokiCerts               = "loki-certs"
	promCerts               = "prom-certs"
	lokiToken               = "loki-token"
	healthServiceName       = "health"
	prometheusServiceName   = "prometheus"
	profilePortName         = "pprof"
	healthTimeoutSeconds    = 5
	livenessPeriodSeconds   = 10
	startupFailureThreshold = 5
	startupPeriodSeconds    = 10
)

type ConfKind string

const (
	ConfMonolith         ConfKind = "allInOne"
	ConfKafkaIngester    ConfKind = "kafkaIngester"
	ConfKafkaTransformer ConfKind = "kafkaTransformer"
)

var FlpConfSuffix = map[ConfKind]string{
	ConfMonolith:         "",
	ConfKafkaIngester:    "-ingester",
	ConfKafkaTransformer: "-transformer",
}

// PodConfigurationDigest is an annotation name to facilitate pod restart after
// any external configuration change
const PodConfigurationDigest = "flows.netobserv.io/configDigest"

type builder struct {
	namespace       string
	labels          map[string]string
	selector        map[string]string
	desired         *flowsv1alpha1.FlowCollectorSpec
	promTLS         *flowsv1alpha1.CertificateReference
	confKind        ConfKind
	useOpenShiftSCC bool
}

func newBuilder(ns string, desired *flowsv1alpha1.FlowCollectorSpec, ck ConfKind, useOpenShiftSCC bool) builder {
	version := helper.ExtractVersion(desired.Processor.Image)
	name := name(ck)
	var promTLS *flowsv1alpha1.CertificateReference
	if desired.Processor.MetricsServer != nil {
		switch desired.Processor.MetricsServer.TLS.Type {
		case flowsv1alpha1.ServerTLSProvided:
			promTLS = desired.Processor.MetricsServer.TLS.Provided
		case flowsv1alpha1.ServerTLSAuto:
			promTLS = &flowsv1alpha1.CertificateReference{
				Type:     "secret",
				Name:     promServiceName(ck),
				CertFile: "tls.crt",
				CertKey:  "tls.key",
			}
		}
	}
	return builder{
		namespace: ns,
		labels: map[string]string{
			"app":     name,
			"version": version,
		},
		selector: map[string]string{
			"app": name,
		},
		desired:         desired,
		confKind:        ck,
		useOpenShiftSCC: useOpenShiftSCC,
		promTLS:         promTLS,
	}
}

func name(ck ConfKind) string                { return constants.FLPName + FlpConfSuffix[ck] }
func RoleBindingName(ck ConfKind) string     { return name(ck) + "-role" }
func RoleBindingMonoName(ck ConfKind) string { return name(ck) + "-role-mono" }
func promServiceName(ck ConfKind) string     { return name(ck) + "-prom" }
func configMapName(ck ConfKind) string       { return name(ck) + "-config" }
func (b *builder) name() string              { return name(b.confKind) }
func (b *builder) promServiceName() string   { return promServiceName(b.confKind) }
func (b *builder) configMapName() string     { return configMapName(b.confKind) }

func (b *builder) portProtocol() corev1.Protocol {
	if b.desired.UseEBPF() {
		return corev1.ProtocolTCP
	}
	return corev1.ProtocolUDP
}

func (b *builder) podTemplate(hasHostPort, hasLokiInterface, hostNetwork bool, configDigest string) corev1.PodTemplateSpec {
	var ports []corev1.ContainerPort
	var tolerations []corev1.Toleration
	if hasHostPort {
		ports = []corev1.ContainerPort{{
			Name:          constants.FLPPortName,
			HostPort:      b.desired.Processor.Port,
			ContainerPort: b.desired.Processor.Port,
			Protocol:      b.portProtocol(),
		}}
		// This allows deploying an instance in the master node, the same technique used in the
		// companion ovnkube-node daemonset definition
		tolerations = []corev1.Toleration{{Operator: corev1.TolerationOpExists}}
	}

	ports = append(ports, corev1.ContainerPort{
		Name:          healthServiceName,
		ContainerPort: b.desired.Processor.HealthPort,
	})

	if b.desired.Processor.MetricsServer != nil {
		ports = append(ports, corev1.ContainerPort{
			Name:          prometheusServiceName,
			ContainerPort: b.desired.Processor.MetricsServer.Port,
		})
	}

	if b.desired.Processor.ProfilePort > 0 {
		ports = append(ports, corev1.ContainerPort{
			Name:          profilePortName,
			ContainerPort: b.desired.Processor.ProfilePort,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	volumeMounts := []corev1.VolumeMount{{
		MountPath: configPath,
		Name:      configVolume,
	}}
	volumes := []corev1.Volume{{
		Name: configVolume,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: b.configMapName(),
				},
			},
		},
	}}

	if b.desired.UseKafka() && b.desired.Kafka.TLS.Enable {
		volumes, volumeMounts = helper.AppendCertVolumes(volumes, volumeMounts, &b.desired.Kafka.TLS, kafkaCerts)
	}

	if hasLokiInterface {
		if b.desired.Loki.TLS.Enable {
			volumes, volumeMounts = helper.AppendCertVolumes(volumes, volumeMounts, &b.desired.Loki.TLS, lokiCerts)
		}
		if b.desired.Loki.UseHostToken() || b.desired.Loki.ForwardUserToken() {
			volumes, volumeMounts = helper.AppendTokenVolume(volumes, volumeMounts, lokiToken, constants.FLPName)
		}
	}

	if b.desired.Processor.MetricsServer != nil &&
		b.desired.Processor.MetricsServer.TLS.Type != flowsv1alpha1.ServerTLSDisabled {
		volumes, volumeMounts = helper.AppendSingleCertVolumes(volumes, volumeMounts, b.promTLS, promCerts)
	}

	var envs []corev1.EnvVar
	for k, v := range b.desired.Processor.Env {
		envs = append(envs, corev1.EnvVar{Name: k, Value: v})
	}

	container := corev1.Container{
		Name:            constants.FLPName,
		Image:           b.desired.Processor.Image,
		ImagePullPolicy: corev1.PullPolicy(b.desired.Processor.ImagePullPolicy),
		Args:            []string{fmt.Sprintf(`--config=%s/%s`, configPath, configFile)},
		Resources:       *b.desired.Processor.Resources.DeepCopy(),
		VolumeMounts:    volumeMounts,
		Ports:           ports,
		Env:             envs,
	}
	if b.desired.Processor.EnableKubeProbes {
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
	dnsPolicy := corev1.DNSClusterFirst
	if hostNetwork {
		dnsPolicy = corev1.DNSClusterFirstWithHostNet
	}

	var annotations map[string]string
	if b.desired.Processor.MetricsServer != nil {
		annotations = map[string]string{
			PodConfigurationDigest:      configDigest,
			"prometheus.io/scrape":      "true",
			"prometheus.io/scrape_port": fmt.Sprint(b.desired.Processor.MetricsServer.Port),
		}
	} else {
		annotations = map[string]string{
			PodConfigurationDigest: configDigest,
		}
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      b.labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			Tolerations:        tolerations,
			Volumes:            volumes,
			Containers:         []corev1.Container{container},
			ServiceAccountName: b.name(),
			HostNetwork:        hostNetwork,
			DNSPolicy:          dnsPolicy,
		},
	}
}

//go:embed metrics_definitions
var metricsConfigEmbed embed.FS

func (b *builder) obtainMetricsConfiguration() (api.PromMetricsItems, error) {
	entries, err := metricsConfigEmbed.ReadDir(metricsConfigDir)
	if err != nil {
		return nil, fmt.Errorf("failed to access metrics_definitions directory: %w", err)
	}

	cg := confgen.NewConfGen(&confgen.Options{
		GenerateStages: []string{"encode_prom"},
		SkipWithTags:   b.desired.Processor.IgnoreMetrics,
	})

	for _, entry := range entries {
		fileName := entry.Name()
		srcPath := filepath.Join(metricsConfigDir, fileName)

		input, err := metricsConfigEmbed.ReadFile(srcPath)
		if err != nil {
			return nil, fmt.Errorf("error reading metrics file %s; %w", srcPath, err)
		}
		err = cg.ParseDefinition(fileName, input)
		if err != nil {
			return nil, fmt.Errorf("error parsing metrics file %s; %w", srcPath, err)
		}
	}

	stages := cg.GenerateTruncatedConfig()
	if len(stages) != 1 {
		return nil, fmt.Errorf("error generating truncated config, 1 stage expected in %v", stages)
	}
	if stages[0].Encode == nil || stages[0].Encode.Prom == nil {
		return nil, fmt.Errorf("error generating truncated config, Encode expected in %v", stages)
	}
	return stages[0].Encode.Prom.Metrics, nil
}

func (b *builder) addTransformStages(stage *config.PipelineBuilderStage) error {
	lastStage := *stage
	// Filter-out unused fields?
	if b.desired.Processor.DropUnusedFields {
		if b.desired.UseIPFIX() {
			lastStage = lastStage.TransformFilter("filter", api.TransformFilter{
				Rules: filters.GetOVSGoflowUnusedRules(),
			})
		}
		// Else: nothing for eBPF at the moment
	}
	// enrich stage (transform) configuration
	enrichedStage := lastStage.TransformNetwork("enrich", api.TransformNetwork{
		Rules: api.NetworkTransformRules{{
			Input:  "SrcAddr",
			Output: "SrcK8S",
			Type:   api.AddKubernetesRuleType,
		}, {
			Input:  "DstAddr",
			Output: "DstK8S",
			Type:   api.AddKubernetesRuleType,
		}},
	})

	// loki stage (write) configuration
	lokiWrite := api.WriteLoki{
		Labels:         constants.LokiIndexFields,
		BatchSize:      int(b.desired.Loki.BatchSize),
		BatchWait:      b.desired.Loki.BatchWait.ToUnstructured().(string),
		MaxBackoff:     b.desired.Loki.MaxBackoff.ToUnstructured().(string),
		MaxRetries:     int(b.desired.Loki.MaxRetries),
		MinBackoff:     b.desired.Loki.MinBackoff.ToUnstructured().(string),
		StaticLabels:   model.LabelSet{},
		Timeout:        b.desired.Loki.Timeout.ToUnstructured().(string),
		URL:            b.desired.Loki.URL,
		TimestampLabel: "TimeFlowEndMs",
		TimestampScale: "1ms",
		TenantID:       b.desired.Loki.TenantID,
	}

	for k, v := range b.desired.Loki.StaticLabels {
		lokiWrite.StaticLabels[model.LabelName(k)] = model.LabelValue(v)
	}

	var authorization *promConfig.Authorization
	if b.desired.Loki.UseHostToken() || b.desired.Loki.ForwardUserToken() {
		authorization = &promConfig.Authorization{
			Type:            "Bearer",
			CredentialsFile: helper.TokensPath + constants.FLPName,
		}
	}

	if b.desired.Loki.TLS.Enable {
		if b.desired.Loki.TLS.InsecureSkipVerify {
			lokiWrite.ClientConfig = &promConfig.HTTPClientConfig{
				Authorization: authorization,
				TLSConfig: promConfig.TLSConfig{
					InsecureSkipVerify: true,
				},
			}
		} else {
			lokiWrite.ClientConfig = &promConfig.HTTPClientConfig{
				Authorization: authorization,
				TLSConfig: promConfig.TLSConfig{
					CAFile: helper.GetCACertPath(&b.desired.Loki.TLS, lokiCerts),
				},
			}
		}
	} else {
		lokiWrite.ClientConfig = &promConfig.HTTPClientConfig{
			Authorization: authorization,
		}
	}
	enrichedStage.WriteLoki("loki", lokiWrite)

	// write on Stdout if logging trace enabled
	if b.desired.Processor.LogLevel == "trace" {
		enrichedStage.WriteStdout("stdout", api.WriteStdout{Format: "json"})
	}

	// obtain encode_prometheus stage from metrics_definitions
	promMetrics, err := b.obtainMetricsConfiguration()
	if err != nil {
		return err
	}

	// prometheus stage (encode) configuration
	if b.desired.Processor.MetricsServer != nil {
		promEncode := api.PromEncode{
			Port:    int(b.desired.Processor.MetricsServer.Port),
			Prefix:  "netobserv_",
			Metrics: promMetrics,
		}

		if b.desired.Processor.MetricsServer.TLS.Type != flowsv1alpha1.ServerTLSDisabled {
			promEncode.TLS = &api.PromTLSConf{
				CertPath: helper.GetSingleCertPath(b.promTLS, promCerts),
				KeyPath:  helper.GetSingleKeyPath(b.promTLS, promCerts),
			}
		}

		enrichedStage.EncodePrometheus("prometheus", promEncode)
	}
	return nil
}

func (b *builder) getKafkaTLS() *api.ClientTLS {
	if b.desired.Kafka.TLS.Enable {
		return &api.ClientTLS{
			InsecureSkipVerify: b.desired.Kafka.TLS.InsecureSkipVerify,
			CACertPath:         helper.GetCACertPath(&b.desired.Kafka.TLS, kafkaCerts),
			UserCertPath:       helper.GetUserCertPath(&b.desired.Kafka.TLS, kafkaCerts),
			UserKeyPath:        helper.GetUserKeyPath(&b.desired.Kafka.TLS, kafkaCerts),
		}
	}
	return nil
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap(stages []config.Stage, parameters []config.StageParam) (*corev1.ConfigMap, string, error) {
	config := map[string]interface{}{
		"log-level": b.desired.Processor.LogLevel,
		"health": map[string]interface{}{
			"port": b.desired.Processor.HealthPort,
		},
		"pipeline":         stages,
		"parameters":       parameters,
		"metrics-settings": config.MetricsSettings{Prefix: "netobserv_", NoPanic: true},
	}
	if b.desired.Processor.ProfilePort > 0 {
		config["profile"] = map[string]interface{}{
			"port": b.desired.Processor.ProfilePort,
		}
	}

	bs, err := json.Marshal(config)
	if err != nil {
		return nil, "", err
	}
	configStr := string(bs)

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.configMapName(),
			Namespace: b.namespace,
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

func (b *builder) newPromService() *corev1.Service {
	if b.desired.Processor.MetricsServer == nil {
		return nil
	}
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.promServiceName(),
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Spec: corev1.ServiceSpec{Selector: b.selector},
	}
	b.fillPromService(&service)
	return &service
}

func (b *builder) fromPromService(old *corev1.Service) *corev1.Service {
	if b.desired.Processor.MetricsServer == nil {
		return nil
	}
	svc := old.DeepCopy()
	b.fillPromService(svc)
	return svc
}

func (b *builder) fillPromService(svc *corev1.Service) {
	if b.desired.Processor.MetricsServer == nil {
		return
	}
	svc.Spec.Ports = []corev1.ServicePort{{
		Name:     prometheusServiceName,
		Port:     b.desired.Processor.MetricsServer.Port,
		Protocol: corev1.ProtocolTCP,
	}}
	if b.desired.Processor.MetricsServer.TLS.Type == flowsv1alpha1.ServerTLSAuto {
		if svc.ObjectMeta.Annotations == nil {
			svc.ObjectMeta.Annotations = map[string]string{}
		}
		svc.ObjectMeta.Annotations[constants.OpenShiftCertificateAnnotation] = b.promServiceName()
	} else if svc.ObjectMeta.Annotations != nil {
		delete(svc.ObjectMeta.Annotations, constants.OpenShiftCertificateAnnotation)
	}
}

func (b *builder) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name(),
			Namespace: b.namespace,
			Labels: map[string]string{
				"app": b.name(),
			},
		},
	}
}

func (b *builder) clusterRoleBinding(ck ConfKind, mono bool) *rbacv1.ClusterRoleBinding {
	var rbName string
	if mono {
		rbName = RoleBindingMonoName(ck)
	} else {
		rbName = RoleBindingName(ck)
	}
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: rbName,
			Labels: map[string]string{
				"app": b.name(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     name(ck),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      b.name(),
			Namespace: b.namespace,
		}},
	}
}
