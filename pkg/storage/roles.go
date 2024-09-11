package storage

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netobserv/network-observability-operator/controllers/constants"
)

func ClusterRoles(createLokiRoles bool, createPromRoles bool) []rbacv1.ClusterRole {
	clusterRoles := []rbacv1.ClusterRole{}
	readerCR := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.CRReader,
		},
		Rules: []rbacv1.PolicyRule{},
	}

	if createLokiRoles {
		// add loki rule to reader
		readerCR.Rules = append(readerCR.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{"loki.grafana.com"},
			Resources:     []string{"network"},
			ResourceNames: []string{"logs"},
			Verbs:         []string{"get"},
		})

		// add writer role
		clusterRoles = append(clusterRoles, rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.CRWriter,
			},
			Rules: []rbacv1.PolicyRule{{
				APIGroups:     []string{"loki.grafana.com"},
				Resources:     []string{"network"},
				ResourceNames: []string{"logs"},
				Verbs:         []string{"create"},
			}},
		})
	}

	// add prometheus rule to reader
	if createPromRoles {
		readerCR.Rules = append(readerCR.Rules, rbacv1.PolicyRule{
			APIGroups: []string{"metrics.k8s.io"},
			Resources: []string{"pods"},
			// TODO: remove "create" verb when https://issues.redhat.com/browse/OCPBUGS-41158 is fixed
			Verbs: []string{"get", "create"},
		})
	}

	// add reader role if at least one rule is set
	if len(readerCR.Rules) > 0 {
		clusterRoles = append(clusterRoles, readerCR)
	}
	return clusterRoles
}

func ClusterRoleWriterBinding(appName, saName, namespace string) *rbacv1.ClusterRoleBinding {
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
