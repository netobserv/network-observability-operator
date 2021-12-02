package consoleplugin

import (
	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
)

func buildLabels() map[string]string {
	return map[string]string{
		"app": pluginName,
	}
}

const secretName = "console-serving-cert"
const displayName = "Network Observability plugin"

// lokiURLAnnotation contains the used Loki querier URL, facilitating the change management
const lokiURLAnnotation = "flows.netobserv.io/loki-url"

func buildConsolePlugin(desired *flowsv1alpha1.FlowCollectorConsolePlugin, ns string) *osv1alpha1.ConsolePlugin {
	return &osv1alpha1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: pluginName,
		},
		Spec: osv1alpha1.ConsolePluginSpec{
			DisplayName: displayName,
			Service: osv1alpha1.ConsolePluginService{
				Name:      pluginName,
				Namespace: ns,
				Port:      desired.Port,
				BasePath:  "/",
			},
		},
	}
}

func buildDeployment(desired *flowsv1alpha1.FlowCollectorSpec, ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pluginName,
			Namespace: ns,
			Annotations: map[string]string{
				lokiURLAnnotation: querierURL(&desired.Loki),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &desired.ConsolePlugin.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: buildLabels(),
			},
			Template: *buildPodTemplate(desired),
		},
	}
}

func buildPodTemplate(desired *flowsv1alpha1.FlowCollectorSpec) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: buildLabels(),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:            pluginName,
				Image:           desired.ConsolePlugin.Image,
				ImagePullPolicy: corev1.PullPolicy(desired.ConsolePlugin.ImagePullPolicy),
				Resources:       *desired.ConsolePlugin.Resources.DeepCopy(),
				VolumeMounts: []corev1.VolumeMount{{
					Name:      secretName,
					MountPath: "/var/serving-cert",
					ReadOnly:  true,
				}},
				Args: []string{
					"-cert", "/var/serving-cert/tls.crt",
					"-key", "/var/serving-cert/tls.key",
					"-loki", querierURL(&desired.Loki),
				},
			}},
			Volumes: []corev1.Volume{{
				Name: secretName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}},
			ServiceAccountName: pluginName,
		},
	}
}

func buildService(old *corev1.Service, desired *flowsv1alpha1.FlowCollectorConsolePlugin, ns string) *corev1.Service {
	if old == nil {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pluginName,
				Namespace: ns,
				Labels:    buildLabels(),
				Annotations: map[string]string{
					"service.alpha.openshift.io/serving-cert-secret-name": "console-serving-cert",
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: buildLabels(),
				Ports: []corev1.ServicePort{{
					Port:     desired.Port,
					Protocol: "TCP",
				}},
			},
		}
	}
	// In case we're updating an existing service, we need to build from the old one to keep immutable fields such as clusterIP
	old.Spec.Ports = []corev1.ServicePort{{
		Port:     desired.Port,
		Protocol: corev1.ProtocolUDP,
	}}
	return old
}

func buildServiceAccount(ns string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pluginName,
			Namespace: ns,
			Labels:    buildLabels(),
		},
	}
}
