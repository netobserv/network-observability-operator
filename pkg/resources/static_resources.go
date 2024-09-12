package resources

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netobserv/network-observability-operator/controllers/constants"
)

var LokiWriterCR = rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: constants.LokiCRWriter,
	},
	Rules: []rbacv1.PolicyRule{{
		APIGroups:     []string{"loki.grafana.com"},
		Resources:     []string{"network"},
		ResourceNames: []string{"logs"},
		Verbs:         []string{"create"},
	}},
}

var LokiReaderCR = rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: constants.LokiCRReader,
	},
	Rules: []rbacv1.PolicyRule{{
		APIGroups:     []string{"loki.grafana.com"},
		Resources:     []string{"network"},
		ResourceNames: []string{"logs"},
		Verbs:         []string{"get"},
	}},
}

var PromReaderCR = rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: constants.PromCRReader,
	},
	Rules: []rbacv1.PolicyRule{{
		APIGroups: []string{"metrics.k8s.io"},
		Resources: []string{"pods"},
		Verbs:     []string{"create"},
	}},
}
