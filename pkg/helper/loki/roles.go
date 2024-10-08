package loki

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/resources"
)

func ClusterRoles(appName, saName, namespace string) ([]rbacv1.ClusterRole, []rbacv1.ClusterRoleBinding) {
	crb := writerBinding(appName, saName, namespace)
	return []rbacv1.ClusterRole{resources.LokiWriterCR(), resources.LokiReaderCR()}, []rbacv1.ClusterRoleBinding{*crb}
}

func writerBinding(appName, saName, namespace string) *rbacv1.ClusterRoleBinding {
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
