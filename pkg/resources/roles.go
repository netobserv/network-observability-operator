package resources

import (
	"github.com/netobserv/network-observability-operator/controllers/constants"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetRoleBindingName(app string, ref constants.RoleName) string {
	return string(ref) + "-" + app
}

func GetClusterRoleBindingName(app string, ref constants.ClusterRoleName) string {
	return string(ref) + "-" + app
}

func GetRoleBinding(namespace, app, sa string, ref constants.RoleName) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(ref) + "-" + app,
			Namespace: namespace,
			Labels:    map[string]string{"app": app},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     string(ref),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      sa,
			Namespace: namespace,
		}},
	}
}

func GetClusterRoleBinding(namespace, app, sa string, ref constants.ClusterRoleName) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   string(ref) + "-" + app,
			Labels: map[string]string{"app": app},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
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
