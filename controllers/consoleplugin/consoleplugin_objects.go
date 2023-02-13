package consoleplugin

import (
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
)

const secretName = "console-serving-cert"
const displayName = "NetObserv plugin"
const proxyAlias = "backend"

const configMapName = "console-plugin-config"
const configFile = "config.yaml"
const configVolume = "config-volume"
const configPath = "/opt/app-root/"
const lokiCerts = "loki-certs"
const tokensPath = "/var/run/secrets/tokens/"

type builder struct {
	namespace   string
	labels      map[string]string
	selector    map[string]string
	desired     *flowsv1alpha1.FlowCollectorConsolePlugin
	desiredLoki *flowsv1alpha1.FlowCollectorLoki
	imageName   string
	cWatcher    *watchers.CertificatesWatcher
}

func newBuilder(ns, imageName string, desired *flowsv1alpha1.FlowCollectorConsolePlugin, desiredLoki *flowsv1alpha1.FlowCollectorLoki, cWatcher *watchers.CertificatesWatcher) builder {
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
		desired:     desired,
		desiredLoki: desiredLoki,
		imageName:   imageName,
		cWatcher:    cWatcher,
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
				Port:      b.desired.Port,
				BasePath:  "/",
			},
			Proxy: []osv1alpha1.ConsolePluginProxy{{
				Type:      osv1alpha1.ProxyTypeService,
				Alias:     proxyAlias,
				Authorize: true,
				Service: osv1alpha1.ConsolePluginProxyServiceConfig{
					Name:      constants.PluginName,
					Namespace: b.namespace,
					Port:      b.desired.Port,
				},
			}},
		},
	}
}

func (b *builder) serviceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PluginName,
			Namespace: b.namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "main",
					Interval: "15s",
					Scheme:   "https",
					TLSConfig: &monitoringv1.TLSConfig{
						SafeTLSConfig: monitoringv1.SafeTLSConfig{
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secretName,
									},
									Key: "tls.crt",
								},
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
			Replicas: &b.desired.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.selector,
			},
			Template: *b.podTemplate(cmDigest),
		},
	}
}

func tokenPath(desiredLoki *flowsv1alpha1.FlowCollectorLoki) string {
	if desiredLoki.UseHostToken() {
		return tokensPath + constants.PluginName
	}
	return ""
}

func buildArgs(desired *flowsv1alpha1.FlowCollectorConsolePlugin, desiredLoki *flowsv1alpha1.FlowCollectorLoki) []string {
	querierURL := querierURL(desiredLoki)
	statusURL := statusURL(desiredLoki)

	args := []string{
		"-cert", "/var/serving-cert/tls.crt",
		"-key", "/var/serving-cert/tls.key",
		"-loki", querierURL,
		"-loki-labels", strings.Join(constants.LokiIndexFields, ","),
		"-loki-tenant-id", desiredLoki.TenantID,
		"-loglevel", desired.LogLevel,
		"-frontend-config", filepath.Join(configPath, configFile),
	}

	if desiredLoki.ForwardUserToken() {
		args = append(args, "-loki-forward-user-token")
	}

	if querierURL != statusURL {
		args = append(args, "-loki-status", statusURL)
	}

	if desiredLoki.TLS.Enable {
		if desiredLoki.TLS.InsecureSkipVerify {
			args = append(args, "-loki-skip-tls")
		} else {
			args = append(args, "--loki-ca-path", helper.GetCACertPath(&desiredLoki.TLS, lokiCerts))
		}
	}
	if desiredLoki.UseHostToken() {
		args = append(args, "-loki-token-path", tokenPath(desiredLoki))
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

	args := buildArgs(b.desired, b.desiredLoki)
	if b.desiredLoki != nil && b.desiredLoki.TLS.Enable && !b.desiredLoki.TLS.InsecureSkipVerify {
		volumes, volumeMounts = helper.AppendCertVolumes(volumes, volumeMounts, &b.desiredLoki.TLS, lokiCerts, b.cWatcher)
	}

	if b.desiredLoki.UseHostToken() {
		volumes, volumeMounts = helper.AppendTokenVolume(volumes, volumeMounts, constants.PluginName, constants.PluginName)
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
				ImagePullPolicy: corev1.PullPolicy(b.desired.ImagePullPolicy),
				Resources:       *b.desired.Resources.DeepCopy(),
				VolumeMounts:    volumeMounts,
				Args:            args,
			}},
			Volumes:            volumes,
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
			MinReplicas: b.desired.Autoscaler.MinReplicas,
			MaxReplicas: b.desired.Autoscaler.MaxReplicas,
			Metrics:     b.desired.Autoscaler.Metrics,
		},
	}
}

func (b *builder) service(old *corev1.Service) *corev1.Service {
	if old == nil {
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
					Port:     b.desired.Port,
					Protocol: "TCP",
					Name:     "main",
				}},
			},
		}
	}
	// In case we're updating an existing service, we need to build from the old one to keep immutable fields such as clusterIP
	newService := old.DeepCopy()
	newService.Spec.Ports = []corev1.ServicePort{{
		Port:     b.desired.Port,
		Protocol: corev1.ProtocolUDP,
	}}
	return newService
}

func buildServiceAccount(ns string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PluginName,
			Namespace: ns,
			Labels: map[string]string{
				"app": constants.PluginName,
			},
		},
	}
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap() (*corev1.ConfigMap, string) {
	config := map[string]interface{}{
		"portNaming":      b.desired.PortNaming,
		"quickFilters":    b.desired.QuickFilters,
		"alertNamespaces": []string{b.namespace},
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
