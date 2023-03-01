package controllers

import (
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

func buildRoleMonitoringReader(ns string) *rbacv1.Role {
	cr := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.OperatorName + roleSuffix,
			Namespace: ns,
		},
		Rules: []rbacv1.PolicyRule{{APIGroups: []string{""},
			Verbs:     []string{"get", "list", "watch"},
			Resources: []string{"pods", "services", "endpoints"},
		},
		},
	}
	return &cr
}

func buildRoleBindingMonitoringReader(ns string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.OperatorName + roleSuffix,
			Namespace: ns,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     constants.OperatorName + roleSuffix,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      monitoringServiceAccount,
			Namespace: monitoringNamespace,
		}},
	}
}
