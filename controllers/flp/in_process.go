package flp

import (
	"context"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type inProcessObjects struct {
	*reconcilers.Instance
	appName        string
	promService    *corev1.Service
	roleBinding    *rbacv1.ClusterRoleBinding
	serviceMonitor *monitoringv1.ServiceMonitor
	prometheusRule *monitoringv1.PrometheusRule
}

type InProcessInfo struct {
	JSONConfig  string
	Annotations map[string]string
	Volumes     volumes.Builder
}

func ReconcileInProcess(ctx context.Context, parent *reconcilers.Instance, desired *flowslatest.FlowCollector, appName string) (*InProcessInfo, error) {
	objs := newInProcessObjects(parent, appName)
	return objs.reconcile(ctx, desired)
}

func newInProcessObjects(parent *reconcilers.Instance, appName string) *inProcessObjects {
	cloneInfo := *parent.Common
	cloneInfo.Namespace = parent.PrivilegedNamespace()
	inst := cloneInfo.NewInstance(parent.Image, parent.Status)

	objs := inProcessObjects{
		Instance:    inst,
		appName:     appName,
		promService: inst.Managed.NewService(promServiceName(appName)),
		roleBinding: inst.Managed.NewCRB(appName),
	}
	if parent.AvailableAPIs.HasSvcMonitor() {
		objs.serviceMonitor = inst.Managed.NewServiceMonitor(serviceMonitorName(appName))
	}
	if parent.AvailableAPIs.HasPromRule() {
		objs.prometheusRule = inst.Managed.NewPrometheusRule(prometheusRuleName(appName))
	}
	return &objs
}

func (i *inProcessObjects) reconcile(ctx context.Context, desired *flowslatest.FlowCollector) (*InProcessInfo, error) {
	// Retrieve current owned objects
	err := i.Managed.FetchAll(ctx)
	if err != nil {
		return nil, err
	}

	if helper.UseKafka(&desired.Spec) {
		// No in-process with Kafka; remove attached resource
		i.Managed.TryDeleteAll(ctx)
		return nil, nil
	}

	fm := metricslatest.FlowMetricList{}
	if err := i.List(ctx, &fm, &client.ListOptions{Namespace: desired.Namespace}); err != nil {
		return nil, i.Status.Error("CantListFlowMetrics", err)
	}

	builder, err := newInProcessBuilder(i.Instance, i.appName, &desired.Spec, &fm)
	if err != nil {
		return nil, err
	}

	cfg, err := builder.GetJSONConfig()
	if err != nil {
		return nil, err
	}

	if err := reconcileRBAC(ctx, i.Common, constants.EBPFAgentName, constants.EBPFServiceAccount, i.Namespace, builder.desired.Loki.Mode); err != nil {
		return nil, err
	}

	if err := reconcilePrometheusService(ctx, i.Instance, i.promService, i.serviceMonitor, i.prometheusRule, builder); err != nil {
		return nil, err
	}

	annotations := map[string]string{}
	// Watch for Loki certificate if necessary; we'll ignore in that case the returned digest, as we don't need to restart pods on cert rotation
	// because certificate is always reloaded from file
	if _, err := i.Watcher.ProcessCACert(ctx, i.Client, &i.Loki.TLS, i.Namespace); err != nil {
		return nil, err
	}
	// Watch for Kafka exporter certificate if necessary; need to restart pods in case of cert rotation
	if err := annotateKafkaExporterCerts(ctx, i.Common, desired.Spec.Exporters, annotations); err != nil {
		return nil, err
	}
	// Watch for monitoring caCert
	if err := reconcileMonitoringCerts(ctx, i.Common, &desired.Spec.Processor.Metrics.Server.TLS, i.Namespace); err != nil {
		return nil, err
	}

	return &InProcessInfo{
		Annotations: annotations,
		Volumes:     builder.volumes,
		JSONConfig:  cfg,
	}, nil
}
