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
		if c.AvailableAPIs.HasPromRule() {
			c.Managed.TryDelete(ctx, c.prometheusRule)
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

	if c.AvailableAPIs.HasPromRule() {
		promRules := c.agentPrometheusRule(target)
		if err := reconcilers.GenericReconcile(ctx, c.Managed, &c.Client, c.prometheusRule, promRules, &report, helper.PrometheusRuleChanged); err != nil {
			return err
		}
	}
	return nil
}

func (c *AgentController) promService(target *flowslatest.FlowCollectorEBPF) *corev1.Service {
	port := helper.GetEBPFMetricsPort(target)
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
				Port:       port,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt32(port),
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

func (c *AgentController) agentPrometheusRule(target *flowslatest.FlowCollectorEBPF) *monitoringv1.PrometheusRule {
	rules := []monitoringv1.Rule{}
	d := monitoringv1.Duration("10m")

	// EBPF hashmap table is full Not receiving any new flows
	if shouldAddAlert(flowslatest.AlertDroppedFlows, target.Metrics.DisableAlerts) {

		rules = append(rules, monitoringv1.Rule{
			Alert: string(flowslatest.AlertDroppedFlows),
			Annotations: map[string]string{
				"description": "NetObserv eBPF agent is not able to process new flows as it's hashmap is full. Hashmap table size can be increased by increasing cacheMaxFlows value in Flowcollector resource.",
				"summary":     "NetObserv eBPF is not able to process any new flows",
			},
			Expr: intstr.FromString("sum(rate(netobserv_agent_dropped_flows_total[1m])) > 0"),
			For:  &d,
			Labels: map[string]string{
				"severity": "warning",
				"app":      "netobserv",
			},
		})
	}

	prometheusRuleObject := monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.EBPFAgentPromoAlertRule,
			Labels: map[string]string{
				"app": constants.EBPFAgentName,
			},
			Namespace: c.PrivilegedNamespace(),
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name:  "NetobservEBPFAgentAlerts",
					Rules: rules,
				},
			},
		},
	}
	return &prometheusRuleObject
}

func shouldAddAlert(name flowslatest.EBPFAgentAlert, disabledList []flowslatest.EBPFAgentAlert) bool {
	for _, disabledAlert := range disabledList {
		if name == disabledAlert {
			return false
		}
	}
	return true
}
