package controllers

import (
	_ "embed"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	downstreamLabelKey       = "openshift.io/cluster-monitoring"
	downstreamLabelValue     = "true"
	roleSuffix               = "-metrics-reader"
	monitoringServiceAccount = "prometheus-k8s"
	monitoringNamespace      = "openshift-monitoring"
)

func buildNamespace(ns string, isDownstream bool) *corev1.Namespace {
	labels := map[string]string{}
	if isDownstream {
		labels[downstreamLabelKey] = downstreamLabelValue
	}
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ns,
			Labels: labels,
		},
	}
}

func buildRoleMonitoringReader() *rbacv1.ClusterRole {
	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.OperatorName + roleSuffix,
		},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{""},
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{"pods", "services", "endpoints"},
			},
			{
				NonResourceURLs: []string{"/metrics"},
				Verbs:           []string{"get"},
			},
		},
	}
	return &cr
}

func buildRoleBindingMonitoringReader(ns string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.OperatorName + roleSuffix,
			Namespace: ns,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     constants.OperatorName + roleSuffix,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      monitoringServiceAccount,
			Namespace: monitoringNamespace,
		}},
	}
}

//go:embed infra_health_dashboard.json
var healthDashboardEmbed string

const (
	healthDashboardCMName       = "grafana-dashboard-netobserv-health"
	healthDashboardCMNamespace  = "openshift-config-managed"
	healthDashboardCMAnnotation = "console.openshift.io/dashboard"
	healthDashboardCMFile       = "netobserv-health-metrics.json"
)

func buildHealthDashboard() *corev1.ConfigMap {
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      healthDashboardCMName,
			Namespace: healthDashboardCMNamespace,
			Labels: map[string]string{
				healthDashboardCMAnnotation: "true",
			},
		},
		Data: map[string]string{
			healthDashboardCMFile: healthDashboardEmbed,
		},
	}
	return &configMap
}
