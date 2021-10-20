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

func goflowKubeLabels() map[string]string {
	return map[string]string{
		"app": gfkName,
	}
}

func goflowKubeDeployment(desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gfkName,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &desired.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: goflowKubeLabels(),
			},
			Template: *goflowKubePodTemplate(desired),
		},
	}
}

func goflowKubeDaemonSet(desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gfkName,
			Namespace: ns,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: goflowKubeLabels(),
			},
			Template: *goflowKubePodTemplate(desired),
		},
	}
}

func goflowKubePodTemplate(desired *flowsv1alpha1.FlowCollectorGoflowKube) *corev1.PodTemplateSpec {
	cmd := goflowKubeMainCommand(desired)
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: goflowKubeLabels(),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:            gfkName,
				Image:           desired.Image,
				ImagePullPolicy: corev1.PullPolicy(desired.ImagePullPolicy),
				Command:         []string{"/bin/sh", "-c", cmd},
			}},
			ServiceAccountName: gfkName,
		},
	}
}

func goflowKubeMainCommand(desired *flowsv1alpha1.FlowCollectorGoflowKube) string {
	return fmt.Sprintf(`/kube-enricher -loglevel %s -stdinsourceformat pb -listen "netflow://:%d"`, desired.LogLevel, desired.Port)
}

func goflowKubeService(desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gfkName,
			Namespace: ns,
			Labels:    goflowKubeLabels(),
		},
		Spec: corev1.ServiceSpec{
			Selector: goflowKubeLabels(),
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

func goflowKubeRBAC(ns string) []client.Object {
	return []client.Object{
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gfkName,
				Namespace: ns,
				Labels:    goflowKubeLabels(),
			},
		},
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:   gfkName,
				Labels: goflowKubeLabels(),
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
				Labels: goflowKubeLabels(),
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
