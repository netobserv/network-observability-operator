package resources

import (
	"github.com/netobserv/netobserv-operator/internal/controller/constants"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetRoleBindingName(shortName string, ref constants.RoleName) string {
	return string(ref) + "-" + shortName
}

func GetRoleBinding(namespace, shortName, app, sa string, ref constants.RoleName, fromClusterRole bool) *rbacv1.RoleBinding {
	roleKind := "Role"
	if fromClusterRole {
		roleKind = "ClusterRole"
	}
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(ref) + "-" + shortName,
			Namespace: namespace,
			Labels: map[string]string{
				"part-of": constants.OperatorName,
				"app":     app,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleKind,
			Name:     string(ref),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      sa,
			Namespace: namespace,
		}},
	}
}

func GetExposeMetricsRoleBinding(ns string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(constants.ExposeMetricsRole),
			Namespace: ns,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     string(constants.ExposeMetricsRole),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      constants.MonitoringServiceAccount,
			Namespace: constants.MonitoringNamespace,
		}},
	}
}
