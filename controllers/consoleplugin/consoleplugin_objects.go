package consoleplugin

import (
	"fmt"
	"hash/fnv"
	"path/filepath"
	"strconv"
	"strings"

	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
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
	namespace           string
	labels              map[string]string
	selector            map[string]string
	desired             *flowslatest.FlowCollectorSpec
	imageName           string
	volumes             volumes.Builder
	availableDashboards []string
}

func newBuilder(ns, imageName string, desired *flowslatest.FlowCollectorSpec, dashboards []string) builder {
	version := helper.ExtractVersion(imageName)
	return builder{
		namespace: ns,
		labels: map[string]string{
			"app":     constants.PluginName,
			"version": helper.MaxLabelLength(version),
		},
		selector: map[string]string{
			"app": constants.PluginName,
		},
		desired:             desired,
		imageName:           imageName,
		availableDashboards: dashboards,
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
				Port:      b.desired.ConsolePlugin.Port,
				BasePath:  "/",
			},
			Proxy: []osv1alpha1.ConsolePluginProxy{{
				Type:      osv1alpha1.ProxyTypeService,
				Alias:     proxyAlias,
				Authorize: true,
				Service: osv1alpha1.ConsolePluginProxyServiceConfig{
					Name:      constants.PluginName,
					Namespace: b.namespace,
					Port:      b.desired.ConsolePlugin.Port,
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

func (b *builder) buildArgs(desired *flowslatest.FlowCollectorSpec) []string {
	querierURL := querierURL(&desired.Loki)
	statusURL := statusURL(&desired.Loki)

	// check for connection traking to list indexes
	indexFields := constants.LokiIndexFields
	if desired.Processor.LogTypes != nil && *desired.Processor.LogTypes != flowslatest.LogTypeFlows {
		indexFields = append(indexFields, constants.LokiConnectionIndexFields...)
	}

	args := []string{
		"-cert", "/var/serving-cert/tls.crt",
		"-key", "/var/serving-cert/tls.key",
		"-loki", querierURL,
		"-loki-labels", strings.Join(indexFields, ","),
		"-loki-tenant-id", desired.Loki.TenantID,
		"-loglevel", desired.ConsolePlugin.LogLevel,
		"-frontend-config", filepath.Join(configPath, configFile),
	}

	if helper.LokiForwardUserToken(&desired.Loki) {
		args = append(args, "-loki-forward-user-token")
	}

	if querierURL != statusURL {
		args = append(args, "-loki-status", statusURL)
	}

	if desired.Loki.TLS.Enable {
		if desired.Loki.TLS.InsecureSkipVerify {
			args = append(args, "-loki-skip-tls")
		} else {
			caPath := b.volumes.AddCACertificate(&desired.Loki.TLS, "loki-certs")
			if caPath != "" {
				args = append(args, "-loki-ca-path", caPath)
			}
		}
	}

	statusTLS := helper.GetLokiStatusTLS(&desired.Loki)
	if statusTLS.Enable {
		if statusTLS.InsecureSkipVerify {
			args = append(args, "-loki-status-skip-tls")
		} else {
			statusCaPath, userCertPath, userKeyPath := b.volumes.AddMutualTLSCertificates(&statusTLS, "loki-status-certs")
			if statusCaPath != "" {
				args = append(args, "-loki-status-ca-path", statusCaPath)
			}
			if userCertPath != "" && userKeyPath != "" {
				args = append(args, "-loki-status-user-cert-path", userCertPath)
				args = append(args, "-loki-status-user-key-path", userKeyPath)
			}
		}
	}

	if helper.LokiUseHostToken(&desired.Loki) {
		tokenPath := b.volumes.AddToken(constants.PluginName)
		args = append(args, "-loki-token-path", tokenPath)
	}
	return args
}

func (b *builder) podTemplate(cmDigest string) *corev1.PodTemplateSpec {
	volumes := []corev1.Volume{{
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

	volumeMounts := []corev1.VolumeMount{{
		Name:      secretName,
		MountPath: "/var/serving-cert",
		ReadOnly:  true,
	}, {
		Name:      configVolume,
		MountPath: configPath,
		ReadOnly:  true,
	},
	}

	args := b.buildArgs(b.desired)

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
				Args:            args,
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
				Port:     b.desired.ConsolePlugin.Port,
				Protocol: corev1.ProtocolTCP,
				// Some Kubernetes versions might automatically set TargetPort to Port. We need to
				// explicitly set it here so the reconcile loop verifies that the owned service
				// is equal as the desired service
				TargetPort: intstr.FromInt(int(b.desired.ConsolePlugin.Port)),
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
				TargetPort: intstr.FromInt(metricsPort),
			}},
		},
	}
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap() (*corev1.ConfigMap, string) {
	outputRecordTypes := helper.GetRecordTypes(&b.desired.Processor)

	var features []string
	if helper.UseEBPF(b.desired) {
		if helper.IsPktDropEnabled(b.desired) {
			features = append(features, "pktDrop")
		}
		if helper.IsDNSTrackingEnabled(b.desired) {
			features = append(features, "dnsTracking")
		}
	}

	config := map[string]interface{}{
		"recordTypes":         outputRecordTypes,
		"portNaming":          b.desired.ConsolePlugin.PortNaming,
		"quickFilters":        b.desired.ConsolePlugin.QuickFilters,
		"alertNamespaces":     []string{b.namespace},
		"sampling":            helper.GetSampling(b.desired),
		"features":            features,
		"availableDashboards": b.availableDashboards,
	}

	configStr := "{}"
	if bs, err := yaml.Marshal(config); err == nil {
		configStr = string(bs)
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
	_, _ = hasher.Write([]byte(configStr))
	digest := strconv.FormatUint(hasher.Sum64(), 36)
	return &configMap, digest
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
