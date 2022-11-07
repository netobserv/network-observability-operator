package flowlogspipeline

import (
	"context"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func AddPrometheusServiceMonitor(ctx context.Context, b *builder, cl reconcilers.ClientHelper) error {
	logger := log.FromContext(ctx)
	svcKey := types.NamespacedName{
		Name:      b.promServiceName(),
		Namespace: b.namespace,
	}
	svc := v1.Service{}
	err := cl.Client.Get(ctx, svcKey, &svc)
	if errors.IsNotFound(err) {
		logger.Info("flowlogs-pipeline prom service not found; not creating the service monitor")
		return nil
	}
	crd := apiextensionsv1.CustomResourceDefinition{}
	crdKey := types.NamespacedName{
		Name: "servicemonitors.monitoring.coreos.com",
	}
	err = cl.Client.Get(ctx, crdKey, &crd)
	if err != nil {
		logger.Info("Service Monitor crd not found; not creating the service monitor")
		return nil
	}
	serviceMonitorObject := buildPrometheusServiceMonitorObject(b)
	// apply object to kubernetes
	if err = cl.CreateOwned(ctx, serviceMonitorObject); err != nil {
		return err
	}
	return nil
}

func buildPrometheusServiceMonitorObject(b *builder) *monitoringv1.ServiceMonitor {
	flpServiceMonitorObject := monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPName,
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
