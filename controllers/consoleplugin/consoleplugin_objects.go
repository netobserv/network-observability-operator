package consoleplugin

import (
	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
)

func buildLabels() map[string]string {
	return map[string]string{
		"app": pluginName,
	}
}

const secretName = "console-serving-cert"
const displayName = "Network Observability plugin"

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

func buildDeployment(desired *flowsv1alpha1.FlowCollectorConsolePlugin, ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pluginName,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &desired.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: buildLabels(),
			},
			Template: *buildPodTemplate(desired),
		},
	}
}

func buildPodTemplate(desired *flowsv1alpha1.FlowCollectorConsolePlugin) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: buildLabels(),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:            pluginName,
				Image:           desired.Image,
				ImagePullPolicy: corev1.PullPolicy(desired.ImagePullPolicy),
				Resources:       *desired.Resources.DeepCopy(),
				VolumeMounts: []corev1.VolumeMount{{
					Name:      secretName,
					MountPath: "/var/serving-cert",
					ReadOnly:  true,
				}},
				Args: []string{
					"--ssl",
					"--cert=/var/serving-cert/tls.crt",
					"--key=/var/serving-cert/tls.key",
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

func buildService(desired *flowsv1alpha1.FlowCollectorConsolePlugin, ns string) *corev1.Service {
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

func buildRBAC(ns string) []client.Object {
	return []client.Object{
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pluginName,
				Namespace: ns,
				Labels:    buildLabels(),
			},
		},
	}
}
