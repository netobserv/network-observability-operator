package loki

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/resources"
)

func ClusterRoles(mode flowslatest.LokiMode) []rbacv1.ClusterRole {
	if mode == flowslatest.LokiModeLokiStack {
		return []rbacv1.ClusterRole{resources.NetObservWriterCR, resources.NetObservReaderCR}
	}
	return []rbacv1.ClusterRole{resources.NetObservReaderCR}
}

func ClusterRoleBinding(appName, saName, namespace string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.CRBWriter,
			Labels: map[string]string{
				"app": appName,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     constants.CRWriter,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      saName,
			Namespace: namespace,
		}},
	}
}
