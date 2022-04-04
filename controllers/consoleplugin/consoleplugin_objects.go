package consoleplugin

import (
	"hash/fnv"
	"strconv"
	"strings"

	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

const secretName = "console-serving-cert"
const displayName = "Network Observability plugin"
const proxyAlias = "backend"

const configMapName = "console-plugin-config"
const configFile = "config.yaml"
const configVolume = "config-volume"
const configPath = "/opt/app-root/"

// PodConfigurationDigest is an annotation name to facilitate pod restart after
// any external configuration change
const PodConfigurationDigest = "flows.netobserv.io/" + configMapName

type builder struct {
	namespace   string
	labels      map[string]string
	selector    map[string]string
	desired     *flowsv1alpha1.FlowCollectorConsolePlugin
	desiredLoki *flowsv1alpha1.FlowCollectorLoki
}

func newBuilder(ns string, desired *flowsv1alpha1.FlowCollectorConsolePlugin, desiredLoki *flowsv1alpha1.FlowCollectorLoki) builder {
	version := helper.ExtractVersion(desired.Image)
	return builder{
		namespace: ns,
		labels: map[string]string{
			"app":     constants.PluginName,
			"version": version,
		},
		selector: map[string]string{
			"app": constants.PluginName,
		},
		desired:     desired,
		desiredLoki: desiredLoki,
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
				Type:  osv1alpha1.ProxyTypeService,
				Alias: proxyAlias,
				Service: osv1alpha1.ConsolePluginProxyServiceConfig{
					Name:      constants.PluginName,
					Namespace: b.namespace,
					Port:      b.desired.Port,
				},
			}},
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

func buildArgs(desired *flowsv1alpha1.FlowCollectorConsolePlugin, desiredLoki *flowsv1alpha1.FlowCollectorLoki) []string {
	return []string{
		"-cert", "/var/serving-cert/tls.crt",
		"-key", "/var/serving-cert/tls.key",
		"-loki", querierURL(desiredLoki),
		"-loki-labels", strings.Join(constants.LokiIndexFields, ","),
		"-loglevel", desired.LogLevel,
		"-frontend-config", configPath + configFile,
	}
}

func (b *builder) podTemplate(cmDigest string) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: b.labels,
			Annotations: map[string]string{
				PodConfigurationDigest: cmDigest,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:            constants.PluginName,
				Image:           b.desired.Image,
				ImagePullPolicy: corev1.PullPolicy(b.desired.ImagePullPolicy),
				Resources:       *b.desired.Resources.DeepCopy(),
				VolumeMounts: []corev1.VolumeMount{{
					Name:      secretName,
					MountPath: "/var/serving-cert",
					ReadOnly:  true,
				},
					{
						Name:      configVolume,
						MountPath: configPath,
						ReadOnly:  true,
					}},
				Args: []string{
					"-cert", "/var/serving-cert/tls.crt",
					"-key", "/var/serving-cert/tls.key",
					"-loki", querierURL(b.desiredLoki),
					"-loki-labels", strings.Join(constants.LokiIndexFields, ","),
					"-loglevel", b.desired.LogLevel,
					"-frontend-config", configPath + configFile,
				},
			}},
			Volumes: []corev1.Volume{{
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
			},
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
				Kind:       constants.DeploymentKind,
				Name:       constants.PluginName,
				APIVersion: "apps/v1",
			},
			MinReplicas: b.desired.HPA.MinReplicas,
			MaxReplicas: b.desired.HPA.MaxReplicas,
			Metrics:     b.desired.HPA.Metrics,
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
					"service.alpha.openshift.io/serving-cert-secret-name": "console-serving-cert",
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: b.selector,
				Ports: []corev1.ServicePort{{
					Port:     b.desired.Port,
					Protocol: "TCP",
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
		"portNaming": b.desired.PortNaming,
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
