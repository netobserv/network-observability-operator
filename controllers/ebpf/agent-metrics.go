package ebpf

import (
	"context"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (c *AgentController) ReconcileMetricsService(ctx context.Context, target *flowslatest.FlowCollectorEBPF) error {
	report := helper.NewChangeReport("EBPF Agent prometheus service")
	defer report.LogIfNeeded(ctx)

	if err := c.ReconcileService(ctx, c.promSvc, c.promService(target), &report); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	if c.AvailableAPIs.HasSvcMonitor() {
		serviceMonitor := c.promServiceMonitoring()
		if err := reconcilers.GenericReconcile(ctx, c.Managed, &c.Client, c.serviceMonitor,
			serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil && !errors.IsAlreadyExists(err) {
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
	return &svc
}

func (c *AgentController) promServiceMonitoring() *monitoringv1.ServiceMonitor {
	agentServiceMonitorObject := monitoringv1.ServiceMonitor{
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
					Port:     "metrics",
					Interval: "30s",
					Scheme:   "http",
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

	return &agentServiceMonitorObject
}
