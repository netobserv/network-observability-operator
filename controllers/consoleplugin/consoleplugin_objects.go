package consoleplugin

import (
	"context"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	cfg "github.com/netobserv/network-observability-operator/controllers/consoleplugin/config"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/ebpf"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/loki"
	"github.com/netobserv/network-observability-operator/pkg/metrics"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
)

const secretName = "console-serving-cert"
const displayName = "NetObserv plugin"
const proxyAlias = "backend"

const configMapName = "console-plugin-config"
const configFile = "config.yaml"
const configVolume = "config-volume"
const configPath = "/opt/app-root/"
const metricsSvcName = constants.PluginName + "-metrics"
const metricsPort = 9002
const metricsPortName = "metrics"

type builder struct {
	info     *reconcilers.Instance
	labels   map[string]string
	selector map[string]string
	desired  *flowslatest.FlowCollectorSpec
	advanced *flowslatest.AdvancedPluginConfig
	volumes  volumes.Builder
}

func newBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec) builder {
	version := helper.ExtractVersion(info.Image)
	advanced := helper.GetAdvancedPluginConfig(desired.ConsolePlugin.Advanced)
	return builder{
		info: info,
		labels: map[string]string{
			"app":     constants.PluginName,
			"version": helper.MaxLabelLength(version),
		},
		selector: map[string]string{
			"app": constants.PluginName,
		},
		desired:  desired,
		advanced: &advanced,
	}
}

func (b *builder) consolePlugin() *osv1alpha1.ConsolePlugin {
	return &osv1alpha1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.PluginName,
		},
		Spec: osv1alpha1.ConsolePluginSpec{
			DisplayName: displayName,
			Service: osv1alpha1.ConsolePluginService{
				Name:      constants.PluginName,
				Namespace: b.info.Namespace,
				Port:      *b.advanced.Port,
				BasePath:  "/",
			},
			Proxy: []osv1alpha1.ConsolePluginProxy{{
				Type:      osv1alpha1.ProxyTypeService,
				Alias:     proxyAlias,
				Authorize: true,
				Service: osv1alpha1.ConsolePluginProxyServiceConfig{
					Name:      constants.PluginName,
					Namespace: b.info.Namespace,
					Port:      *b.advanced.Port,
				},
			}},
		},
	}
}

func (b *builder) serviceMonitor() *monitoringv1.ServiceMonitor {
	serverName := fmt.Sprintf("%s.%s.svc", constants.PluginName, b.info.Namespace)
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PluginName,
			Namespace: b.info.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     metricsPortName,
					Interval: "15s",
					Scheme:   "https",
					TLSConfig: &monitoringv1.TLSConfig{
						SafeTLSConfig: monitoringv1.SafeTLSConfig{
							ServerName: serverName,
							CA: monitoringv1.SecretOrConfigMap{
								ConfigMap: &corev1.ConfigMapKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "openshift-service-ca.crt",
									},
									Key: "service-ca.crt",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secretName,
									},
									Key: "tls.crt",
								},
							},
							KeySecret: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secretName,
								},
								Key: "tls.key",
							},
						},
					},
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					b.info.Namespace,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": constants.PluginName,
				},
			},
		},
	}
}

func (b *builder) deployment(cmDigest string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PluginName,
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: b.desired.ConsolePlugin.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.selector,
			},
			Template: *b.podTemplate(cmDigest),
		},
	}
}

func (b *builder) podTemplate(cmDigest string) *corev1.PodTemplateSpec {
	volumes := []corev1.Volume{
		{
			Name: secretName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		}, {
			Name: configVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      secretName,
			MountPath: "/var/serving-cert",
			ReadOnly:  true,
		}, {
			Name:      configVolume,
			MountPath: configPath,
			ReadOnly:  true,
		},
	}

	// ensure volumes are up to date
	loki := b.info.Loki
	if loki.TLS.Enable && !loki.TLS.InsecureSkipVerify {
		b.volumes.AddCACertificate(&loki.TLS, "loki-certs")
	}
	if loki.StatusTLS.Enable && !loki.StatusTLS.InsecureSkipVerify {
		b.volumes.AddMutualTLSCertificates(&loki.StatusTLS, "loki-status-certs")
	}
	if loki.UseHostToken() {
		b.volumes.AddToken(constants.PluginName)
	}

	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: b.labels,
			Annotations: map[string]string{
				constants.PodConfigurationDigest: cmDigest,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:            constants.PluginName,
				Image:           b.info.Image,
				ImagePullPolicy: corev1.PullPolicy(b.desired.ConsolePlugin.ImagePullPolicy),
				Resources:       *b.desired.ConsolePlugin.Resources.DeepCopy(),
				VolumeMounts:    b.volumes.AppendMounts(volumeMounts),
				Env:             []corev1.EnvVar{constants.EnvNoHTTP2},
				Args: []string{

					"-loglevel", b.desired.ConsolePlugin.LogLevel,
					"-config", filepath.Join(configPath, configFile),
				},
				SecurityContext: helper.ContainerDefaultSecurityContext(),
			}},
			Volumes:            b.volumes.AppendVolumes(volumes),
			ServiceAccountName: constants.PluginName,
			NodeSelector:       b.advanced.Scheduling.NodeSelector,
			Tolerations:        b.advanced.Scheduling.Tolerations,
			Affinity:           b.advanced.Scheduling.Affinity,
			PriorityClassName:  b.advanced.Scheduling.PriorityClassName,
		},
	}
}

func (b *builder) autoScaler() *ascv2.HorizontalPodAutoscaler {
	return &ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PluginName,
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       constants.PluginName,
			},
			MinReplicas: b.desired.ConsolePlugin.Autoscaler.MinReplicas,
			MaxReplicas: b.desired.ConsolePlugin.Autoscaler.MaxReplicas,
			Metrics:     b.desired.ConsolePlugin.Autoscaler.Metrics,
		},
	}
}

func (b *builder) mainService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PluginName,
			Namespace: b.info.Namespace,
			Labels:    b.labels,
			Annotations: map[string]string{
				constants.OpenShiftCertificateAnnotation: "console-serving-cert",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: b.selector,
			Ports: []corev1.ServicePort{{
				Port:     *b.advanced.Port,
				Protocol: corev1.ProtocolTCP,
				// Some Kubernetes versions might automatically set TargetPort to Port. We need to
				// explicitly set it here so the reconcile loop verifies that the owned service
				// is equal as the desired service
				TargetPort: intstr.FromInt32(*b.advanced.Port),
			}},
		},
	}
}

func (b *builder) metricsService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricsSvcName,
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: b.selector,
			Ports: []corev1.ServicePort{{
				Port:     metricsPort,
				Protocol: corev1.ProtocolTCP,
				Name:     metricsPortName,
				// Some Kubernetes versions might automatically set TargetPort to Port. We need to
				// explicitly set it here so the reconcile loop verifies that the owned service
				// is equal as the desired service
				TargetPort: intstr.FromInt32(metricsPort),
			}},
		},
	}
}

func (b *builder) getLokiConfig() cfg.LokiConfig {
	if !helper.UseLoki(b.desired) {
		// Empty config/URL will disable Loki in the console plugin
		return cfg.LokiConfig{}
	}
	lk := b.info.Loki
	lconf := cfg.LokiConfig{
		URL:              lk.QuerierURL,
		Labels:           loki.GetLokiLabels(b.desired),
		Timeout:          api.Duration{Duration: 30 * time.Second},
		TenantID:         lk.TenantID,
		ForwardUserToken: lk.UseForwardToken(),
	}
	if lk.QuerierURL != lk.StatusURL {
		lconf.StatusURL = lk.StatusURL
	}
	if b.desired.Loki.ReadTimeout != nil {
		lconf.Timeout = api.Duration{Duration: b.desired.Loki.ReadTimeout.Duration}
	}
	if lk.TLS.Enable {
		if lk.TLS.InsecureSkipVerify {
			lconf.SkipTLS = true
		} else {
			caPath := b.volumes.AddCACertificate(&lk.TLS, "loki-certs")
			if caPath != "" {
				lconf.CAPath = caPath
			}
		}
	}
	if lk.StatusTLS.Enable {
		if lk.StatusTLS.InsecureSkipVerify {
			lconf.StatusSkipTLS = true
		} else {
			statusCaPath, userCertPath, userKeyPath := b.volumes.AddMutualTLSCertificates(&lk.StatusTLS, "loki-status-certs")
			if statusCaPath != "" {
				lconf.StatusCAPath = statusCaPath
			}
			if userCertPath != "" && userKeyPath != "" {
				lconf.StatusUserCertPath = userCertPath
				lconf.StatusUserKeyPath = userKeyPath
			}
		}
	}
	if lk.UseHostToken() {
		lconf.TokenPath = b.volumes.AddToken(constants.PluginName)
	}
	return lconf
}

func (b *builder) getPromConfig(ctx context.Context) cfg.PrometheusConfig {
	if !helper.UsePrometheus(b.desired) {
		return cfg.PrometheusConfig{}
	}

	// Default config = manual
	tls := b.desired.Prometheus.Querier.Manual.TLS
	config := cfg.PrometheusConfig{
		URL:              b.desired.Prometheus.Querier.Manual.URL,
		ForwardUserToken: b.desired.Prometheus.Querier.Manual.ForwardUserToken,
	}
	if b.desired.Prometheus.Querier.Mode == flowslatest.PromModeAuto {
		if b.info.UseOpenShiftSCC /* aka IsOpenShift */ {
			config = cfg.PrometheusConfig{
				URL:              "https://thanos-querier.openshift-monitoring.svc:9091/",
				ForwardUserToken: true,
			}
			tls = flowslatest.ClientTLS{
				Enable: true,
				CACert: flowslatest.CertificateReference{
					Type:     flowslatest.RefTypeConfigMap,
					Name:     "openshift-service-ca.crt",
					CertFile: "service-ca.crt",
				},
			}
		} else {
			log.FromContext(ctx).Info("Could not configure Prometheus querier automatically. Using manual configuration.")
		}
	}

	if tls.Enable {
		if tls.InsecureSkipVerify {
			config.SkipTLS = true
		} else {
			caPath := b.volumes.AddCACertificate(&tls, "prom-certs")
			if caPath != "" {
				config.CAPath = caPath
			}
		}
	}

	if b.desired.Prometheus.Querier.Timeout != nil {
		config.Timeout = api.Duration{Duration: b.desired.Prometheus.Querier.Timeout.Duration}
	} else {
		config.Timeout = api.Duration{Duration: 30 * time.Second}
	}
	config.TokenPath = b.volumes.AddToken(constants.PluginName)

	allMetricNames := metrics.GetAllNames()
	includeList := metrics.GetIncludeList(b.desired)
	allMetrics := metrics.GetDefinitions(allMetricNames)
	for i := range allMetrics {
		mSpec := allMetrics[i].Spec
		enabled := slices.Contains(includeList, mSpec.MetricName)
		config.Metrics = append(config.Metrics, cfg.MetricInfo{
			Enabled:    enabled,
			Name:       "netobserv_" + mSpec.MetricName,
			Type:       string(mSpec.Type),
			ValueField: mSpec.ValueField,
			Direction:  string(mSpec.Direction),
			Labels:     mSpec.Labels,
		})
	}

	return config
}

func (b *builder) setFrontendConfig(fconf *cfg.FrontendConfig) error {
	var err error
	dedupJustMark, err := strconv.ParseBool(ebpf.DedupeJustMarkDefault)
	if err != nil {
		return err
	}
	dedupMerge, err := strconv.ParseBool(ebpf.DedupeMergeDefault)
	if err != nil {
		return err
	}
	if helper.IsPktDropEnabled(&b.desired.Agent.EBPF) {
		fconf.Features = append(fconf.Features, "pktDrop")
	}

	if helper.IsDNSTrackingEnabled(&b.desired.Agent.EBPF) {
		fconf.Features = append(fconf.Features, "dnsTracking")
	}

	if helper.IsFlowRTTEnabled(&b.desired.Agent.EBPF) {
		fconf.Features = append(fconf.Features, "flowRTT")
	}

	if b.desired.Agent.EBPF.Advanced != nil {
		if v, ok := b.desired.Agent.EBPF.Advanced.Env[ebpf.EnvDedupeJustMark]; ok {
			dedupJustMark, err = strconv.ParseBool(v)
			if err != nil {
				return err
			}
		}

		if v, ok := b.desired.Agent.EBPF.Advanced.Env[ebpf.EnvDedupeMerge]; ok {
			dedupMerge, err = strconv.ParseBool(v)
			if err != nil {
				return err
			}
		}
	}
	fconf.RecordTypes = helper.GetRecordTypes(&b.desired.Processor)
	fconf.PortNaming = b.desired.ConsolePlugin.PortNaming
	fconf.QuickFilters = b.desired.ConsolePlugin.QuickFilters
	fconf.AlertNamespaces = []string{b.info.Namespace}
	fconf.Sampling = helper.GetSampling(b.desired)
	fconf.Deduper = cfg.Deduper{
		Mark:  dedupJustMark,
		Merge: dedupMerge,
	}
	if helper.IsMultiClusterEnabled(&b.desired.Processor) {
		fconf.Features = append(fconf.Features, "multiCluster")
	}
	if helper.IsZoneEnabled(&b.desired.Processor) {
		fconf.Features = append(fconf.Features, "zones")
	}
	if helper.IsSubnetLabelsEnabled(&b.desired.Processor) {
		fconf.Features = append(fconf.Features, "subnetLabels")
	}
	return nil
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap(ctx context.Context) (*corev1.ConfigMap, string, error) {
	config := cfg.PluginConfig{}
	// configure server
	config.Server.CertPath = "/var/serving-cert/tls.crt"
	config.Server.KeyPath = "/var/serving-cert/tls.key"
	config.Server.Port = int(*b.advanced.Port)

	// configure loki
	config.Loki = b.getLokiConfig()

	// configure prometheus
	config.Prometheus = b.getPromConfig(ctx)

	// configure frontend from embedded static file
	var err error
	config.Frontend, err = cfg.LoadStaticFrontendConfig()
	if err != nil {
		return nil, "", err
	}
	err = b.setFrontendConfig(&config.Frontend)
	if err != nil {
		return nil, "", err
	}

	var configStr string
	bs, err := yaml.Marshal(config)
	if err == nil {
		configStr = string(bs)
	} else {
		return nil, "", err
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Data: map[string]string{
			configFile: configStr,
		},
	}
	hasher := fnv.New64a()
	_, err = hasher.Write([]byte(configStr))
	if err != nil {
		return nil, "", err
	}
	digest := strconv.FormatUint(hasher.Sum64(), 36)
	return &configMap, digest, nil
}

func (b *builder) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PluginName,
			Namespace: b.info.Namespace,
			Labels: map[string]string{
				"app": constants.PluginName,
			},
		},
	}
}

// The operator needs to have at least the same permissions as flowlogs-pipeline in order to grant them
//+kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create

func buildClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.PluginName,
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"authentication.k8s.io"},
			Verbs:     []string{"create"},
			Resources: []string{"tokenreviews"},
		}},
	}
}

func (b *builder) clusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.PluginName,
			Labels: map[string]string{
				"app": constants.PluginName,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     constants.PluginName,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      constants.PluginName,
			Namespace: b.info.Namespace,
		}},
	}
}
