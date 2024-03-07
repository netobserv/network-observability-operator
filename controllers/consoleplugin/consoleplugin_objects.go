package consoleplugin

import (
	_ "embed"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"strconv"

	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	config "github.com/netobserv/network-observability-operator/controllers/consoleplugin/config"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/loki"
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
	namespace string
	labels    map[string]string
	selector  map[string]string
	desired   *flowslatest.FlowCollectorSpec
	advanced  *flowslatest.AdvancedPluginConfig
	imageName string
	volumes   volumes.Builder
	loki      *helper.LokiConfig
}

func newBuilder(ns, imageName string, desired *flowslatest.FlowCollectorSpec, loki *helper.LokiConfig) builder {
	version := helper.ExtractVersion(imageName)
	advanced := helper.GetAdvancedPluginConfig(desired.ConsolePlugin.Advanced)
	return builder{
		namespace: ns,
		labels: map[string]string{
			"app":     constants.PluginName,
			"version": helper.MaxLabelLength(version),
		},
		selector: map[string]string{
			"app": constants.PluginName,
		},
		desired:   desired,
		advanced:  &advanced,
		imageName: imageName,
		loki:      loki,
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
				Namespace: b.namespace,
				Port:      *b.advanced.Port,
				BasePath:  "/",
			},
			Proxy: []osv1alpha1.ConsolePluginProxy{{
				Type:      osv1alpha1.ProxyTypeService,
				Alias:     proxyAlias,
				Authorize: true,
				Service: osv1alpha1.ConsolePluginProxyServiceConfig{
					Name:      constants.PluginName,
					Namespace: b.namespace,
					Port:      *b.advanced.Port,
				},
			}},
		},
	}
}

func (b *builder) serviceMonitor() *monitoringv1.ServiceMonitor {
	serverName := fmt.Sprintf("%s.%s.svc", constants.PluginName, b.namespace)
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PluginName,
			Namespace: b.namespace,
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
					b.namespace,
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
			Namespace: b.namespace,
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
	if b.loki.TLS.Enable && !b.loki.TLS.InsecureSkipVerify {
		b.volumes.AddCACertificate(&b.loki.TLS, "loki-certs")
	}
	if b.loki.StatusTLS.Enable && !b.loki.StatusTLS.InsecureSkipVerify {
		b.volumes.AddMutualTLSCertificates(&b.loki.StatusTLS, "loki-status-certs")
	}
	if b.loki.UseHostToken() {
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
				Image:           b.imageName,
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
		},
	}
}

func (b *builder) autoScaler() *ascv2.HorizontalPodAutoscaler {
	return &ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PluginName,
			Namespace: b.namespace,
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
			Namespace: b.namespace,
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
			Namespace: b.namespace,
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

func (b *builder) setLokiConfig(lconf *config.LokiConfig) {
	lconf.URL = b.loki.QuerierURL
	statusURL := b.loki.StatusURL
	if lconf.URL != statusURL {
		lconf.StatusURL = statusURL
	}
	lconf.Labels = loki.GetLokiLabels(b.desired)
	if b.desired.Loki.ReadTimeout != nil {
		lconf.Timeout = helper.UnstructuredDuration(b.desired.Loki.ReadTimeout)
	} else {
		lconf.Timeout = "30s"
	}
	lconf.TenantID = b.loki.TenantID
	lconf.ForwardUserToken = b.loki.UseForwardToken()
	if b.loki.TLS.Enable {
		if b.loki.TLS.InsecureSkipVerify {
			lconf.SkipTLS = true
		} else {
			caPath := b.volumes.AddCACertificate(&b.loki.TLS, "loki-certs")
			if caPath != "" {
				lconf.CAPath = caPath
			}
		}
	}
	if b.loki.StatusTLS.Enable {
		if b.loki.StatusTLS.InsecureSkipVerify {
			lconf.StatusSkipTLS = true
		} else {
			statusCaPath, userCertPath, userKeyPath := b.volumes.AddMutualTLSCertificates(&b.loki.StatusTLS, "loki-status-certs")
			if statusCaPath != "" {
				lconf.StatusCAPath = statusCaPath
			}
			if userCertPath != "" && userKeyPath != "" {
				lconf.StatusUserCertPath = userCertPath
				lconf.StatusUserKeyPath = userKeyPath
			}
		}
	}
	if b.loki.UseHostToken() {
		lconf.TokenPath = b.volumes.AddToken(constants.PluginName)
	}
}

func (b *builder) setFrontendConfig(fconf *config.FrontendConfig) error {
	if helper.IsPktDropEnabled(&b.desired.Agent.EBPF) {
		fconf.Features = append(fconf.Features, "pktDrop")
	}

	if helper.IsDNSTrackingEnabled(&b.desired.Agent.EBPF) {
		fconf.Features = append(fconf.Features, "dnsTracking")
	}

	if helper.IsFlowRTTEnabled(&b.desired.Agent.EBPF) {
		fconf.Features = append(fconf.Features, "flowRTT")
	}

	fconf.RecordTypes = helper.GetRecordTypes(&b.desired.Processor)
	fconf.PortNaming = b.desired.ConsolePlugin.PortNaming
	fconf.QuickFilters = b.desired.ConsolePlugin.QuickFilters
	fconf.AlertNamespaces = []string{b.namespace}
	fconf.Sampling = helper.GetSampling(b.desired)
	fconf.Deduper = config.Deduper{
		Mark:  helper.UseDedupJustMark(b.desired),
		Merge: helper.UseDedupMerge(b.desired),
	}
	if helper.IsMultiClusterEnabled(&b.desired.Processor) {
		fconf.Features = append(fconf.Features, "multiCluster")
	}
	if helper.IsZoneEnabled(&b.desired.Processor) {
		fconf.Features = append(fconf.Features, "zones")
	}
	return nil
}

//go:embed config/static-frontend-config.yaml
var staticFrontendConfig []byte

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap() (*corev1.ConfigMap, string, error) {
	config := config.PluginConfig{}
	// configure server
	config.Server.CertPath = "/var/serving-cert/tls.crt"
	config.Server.KeyPath = "/var/serving-cert/tls.key"
	config.Server.Port = int(*b.advanced.Port)

	// configure loki
	b.setLokiConfig(&config.Loki)

	// configure frontend from embedded static file
	err := yaml.Unmarshal(staticFrontendConfig, &config.Frontend)
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
			Namespace: b.namespace,
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
			Namespace: b.namespace,
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
			Namespace: b.namespace,
		}},
	}
}
