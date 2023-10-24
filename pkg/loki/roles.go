package loki

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
)

func ClusterRoles(mode flowslatest.LokiMode) []rbacv1.ClusterRole {
	if mode == flowslatest.LokiModeLokiStack {
		return []rbacv1.ClusterRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.LokiCRWriter,
				},
				Rules: []rbacv1.PolicyRule{{
					APIGroups:     []string{"loki.grafana.com"},
					Resources:     []string{"network"},
					ResourceNames: []string{"logs"},
					Verbs:         []string{"create"},
				}},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.LokiCRReader,
				},
				Rules: []rbacv1.PolicyRule{{
					APIGroups:     []string{"loki.grafana.com"},
					Resources:     []string{"network"},
					ResourceNames: []string{"logs"},
					Verbs:         []string{"get"},
				}},
			},
		}
	}
	return []rbacv1.ClusterRole{}
}

func ClusterRoleBinding(appName, saName, namespace string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.LokiCRBWriter,
			Labels: map[string]string{
				"app": appName,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     constants.LokiCRWriter,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      saName,
			Namespace: namespace,
		}},
	}
}
