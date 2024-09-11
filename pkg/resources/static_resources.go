package resources

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netobserv/network-observability-operator/controllers/constants"
)

var NetObservWriterCR = rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: constants.CRWriter,
	},
	Rules: []rbacv1.PolicyRule{{
		APIGroups:     []string{"loki.grafana.com"},
		Resources:     []string{"network"},
		ResourceNames: []string{"logs"},
		Verbs:         []string{"create"},
	}},
}

var NetObservReaderCR = rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: constants.CRReader,
	},
	Rules: []rbacv1.PolicyRule{{
		APIGroups: []string{"metrics.k8s.io"},
		Resources: []string{"pods"},
		Verbs:     []string{"create"},
	}, {
		APIGroups:     []string{"loki.grafana.com"},
		Resources:     []string{"network"},
		ResourceNames: []string{"logs"},
		Verbs:         []string{"get"},
	}},
}
