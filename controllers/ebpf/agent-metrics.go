package ebpf

import (
	"context"
	"fmt"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (c *AgentController) reconcileMetricsService(ctx context.Context, target *flowslatest.FlowCollectorEBPF) error {
	report := helper.NewChangeReport("EBPF Agent prometheus service")
	defer report.LogIfNeeded(ctx)

	if !helper.IsEBPFMetricsEnabled(target) {
		c.Managed.TryDelete(ctx, c.promSvc)
		if c.AvailableAPIs.HasSvcMonitor() {
			c.Managed.TryDelete(ctx, c.serviceMonitor)
		}
		return nil
	}

	if err := c.ReconcileService(ctx, c.promSvc, c.promService(target), &report); err != nil {
		return err
	}
	if c.AvailableAPIs.HasSvcMonitor() {
		serviceMonitor := c.promServiceMonitoring(target)
		if err := reconcilers.GenericReconcile(ctx, c.Managed, &c.Client, c.serviceMonitor,
			serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	return nil
}

func (c *AgentController) promService(target *flowslatest.FlowCollectorEBPF) *corev1.Service {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFAgentMetricsSvcName,
			Namespace: c.PrivilegedNamespace(),
			Labels: map[string]string{
				"app": constants.EBPFAgentName,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": constants.EBPFAgentName,
			},
			Ports: []corev1.ServicePort{{
				Name:       "metrics",
				Port:       target.Metrics.Server.Port,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt32(target.Metrics.Server.Port),
			}},
		},
	}
	if target.Metrics.Server.TLS.Type == flowslatest.ServerTLSAuto {
		svc.ObjectMeta.Annotations = map[string]string{
			constants.OpenShiftCertificateAnnotation: constants.EBPFAgentMetricsSvcName,
		}
	}
	return &svc
}

func (c *AgentController) promServiceMonitoring(target *flowslatest.FlowCollectorEBPF) *monitoringv1.ServiceMonitor {
	serverName := fmt.Sprintf("%s.%s.svc", constants.EBPFAgentMetricsSvcName, c.PrivilegedNamespace())
	scheme, smTLS := helper.GetServiceMonitorTLSConfig(&target.Metrics.Server.TLS, serverName, c.IsDownstream)
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFAgentMetricsSvcMonitoringName,
			Namespace: c.PrivilegedNamespace(),
			Labels: map[string]string{
				"app": constants.EBPFAgentName,
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:      "metrics",
					Interval:  "30s",
					Scheme:    scheme,
					TLSConfig: smTLS,
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					c.PrivilegedNamespace(),
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": constants.EBPFAgentName,
				},
			},
		},
	}
}
