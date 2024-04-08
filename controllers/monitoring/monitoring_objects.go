package monitoring

import (
	"regexp"
	"strings"

	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/dashboards"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	downstreamLabelKey    = "openshift.io/cluster-monitoring"
	downstreamLabelValue  = "true"
	roleSuffix            = "-metrics-reader"
	dashboardCMNamespace  = "openshift-config-managed"
	dashboardCMAnnotation = "console.openshift.io/dashboard"

	flowDashboardCMFile = "netobserv-flow-metrics.json"

	healthDashboardCMName = "grafana-dashboard-netobserv-health"
	healthDashboardCMFile = "netobserv-health-metrics.json"
)

var k8sInvalidChar = regexp.MustCompile(`[^a-z0-9\-]`)

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
			{
				APIGroups: []string{""},
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
			Name:      constants.MonitoringServiceAccount,
			Namespace: constants.MonitoringNamespace,
		}},
	}
}

func buildFlowMetricsDashboards(metrics []metricslatest.FlowMetric) []*corev1.ConfigMap {
	var cms []*corev1.ConfigMap
	dash := dashboards.CreateFlowMetricsDashboards(metrics)

	for name, json := range dash {
		k8sName := "netobserv-" + k8sInvalidChar.ReplaceAllString(strings.ToLower(name), "-")
		configMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      k8sName,
				Namespace: dashboardCMNamespace,
				Labels: map[string]string{
					dashboardCMAnnotation: "true",
				},
			},
			Data: map[string]string{
				flowDashboardCMFile: json,
			},
		}
		cms = append(cms, &configMap)
	}
	return cms
}

func buildHealthDashboard(namespace string) (*corev1.ConfigMap, bool, error) {
	dashboard, err := dashboards.CreateHealthDashboard(namespace)
	if err != nil {
		return nil, false, err
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      healthDashboardCMName,
			Namespace: dashboardCMNamespace,
			Labels: map[string]string{
				dashboardCMAnnotation: "true",
			},
		},
		Data: map[string]string{
			healthDashboardCMFile: dashboard,
		},
	}
	return &configMap, len(dashboard) == 0, nil
}
