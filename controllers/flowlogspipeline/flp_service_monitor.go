package flowlogspipeline

import (
	"context"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddPrometheusServiceMonitor(ctx context.Context, b *builder, cl reconcilers.ClientHelper) error {
	serviceMonitorObject := buildPrometheusServiceMonitorObject(b)
	// apply object to kubernetes
	if err := cl.CreateOwned(ctx, serviceMonitorObject); err != nil {
		return err
	}
	return nil
}

func buildPrometheusServiceMonitorObject(b *builder) *monitoringv1.ServiceMonitor {
	flpServiceMonitorObject := monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPServiceMonitorName,
			Namespace: b.namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     prometheusServiceName,
					Interval: "15s",
					Scheme:   "http",
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					b.namespace,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": constants.FLPName,
				},
			},
		},
	}
	return &flpServiceMonitorObject
}
