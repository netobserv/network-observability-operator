package goflowkube

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
)

const configMapName = "goflow-kube-config"
const configVolume = "config-volume"
const configPath = "/etc/goflow-kube"
const configFile = "config.yaml"

func buildLabels() map[string]string {
	return map[string]string{
		"app": gfkName,
	}
}

func buildDeployment(desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gfkName,
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

func buildDaemonSet(desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gfkName,
			Namespace: ns,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: buildLabels(),
			},
			Template: *buildPodTemplate(desired),
		},
	}
}

func buildPodTemplate(desired *flowsv1alpha1.FlowCollectorGoflowKube) *corev1.PodTemplateSpec {
	cmd := buildMainCommand(desired)
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: buildLabels(),
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: configVolume,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapName,
						},
					},
				},
			}},
			Containers: []corev1.Container{{
				Name:            gfkName,
				Image:           desired.Image,
				ImagePullPolicy: corev1.PullPolicy(desired.ImagePullPolicy),
				Command:         []string{"/bin/sh", "-c", cmd},
				Resources:       *desired.Resources.DeepCopy(),
				VolumeMounts: []corev1.VolumeMount{{
					MountPath: configPath,
					Name:      configVolume,
				}},
			}},
			ServiceAccountName: gfkName,
		},
	}
}

func buildMainCommand(desired *flowsv1alpha1.FlowCollectorGoflowKube) string {
	return fmt.Sprintf(`/goflow-kube -loglevel "%s" -config %s/%s`, desired.LogLevel, configPath, configFile)
}

func buildConfigMap(desiredGoflowKube *flowsv1alpha1.FlowCollectorGoflowKube,
	desiredLoki *flowsv1alpha1.FlowCollectorLoki, ns string) *corev1.ConfigMap {
	data := fmt.Sprintf(`
{
listen: "netflow://:%d",
loki: {
	url: %s,
	labels: ["SrcNamespace","SrcWorkload","DstNamespace","DstWorkload"],
}}
`, desiredGoflowKube.Port, desiredLoki.URL)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: ns,
			Labels:    buildLabels(),
		},
		Data: map[string]string{
			configFile: data,
		},
	}
}

func buildService(desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gfkName,
			Namespace: ns,
			Labels:    buildLabels(),
		},
		Spec: corev1.ServiceSpec{
			Selector: buildLabels(),
			Ports: []corev1.ServicePort{{
				Port:     desired.Port,
				Protocol: "UDP",
			}},
		},
	}
}

// The operator needs to have at least the same permissions as goflow-kube in order to grant them
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=pods;services,verbs=get;list;watch

func buildRBAC(ns string) []client.Object {
	return []client.Object{
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gfkName,
				Namespace: ns,
				Labels:    buildLabels(),
			},
		},
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:   gfkName,
				Labels: buildLabels(),
			},
			Rules: []rbacv1.PolicyRule{{
				APIGroups: []string{""},
				Verbs:     []string{"list", "get", "watch"},
				Resources: []string{"pods", "services"},
			}, {
				APIGroups: []string{"apps"},
				Verbs:     []string{"list", "get", "watch"},
				Resources: []string{"replicasets"},
			}},
		},
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:   gfkName,
				Labels: buildLabels(),
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     gfkName,
			},
			Subjects: []rbacv1.Subject{{
				Kind:      "ServiceAccount",
				Name:      gfkName,
				Namespace: ns,
			}},
		},
	}
}
