package flp

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/resources"
)

type monolithReconciler struct {
	*reconcilers.Instance
	daemonSet        *appsv1.DaemonSet
	promService      *corev1.Service
	serviceAccount   *corev1.ServiceAccount
	staticConfigMap  *corev1.ConfigMap
	dynamicConfigMap *corev1.ConfigMap
	rbConfigWatcher  *rbacv1.RoleBinding
	rbHostNetwork    *rbacv1.ClusterRoleBinding
	rbLokiWriter     *rbacv1.ClusterRoleBinding
	rbInformer       *rbacv1.ClusterRoleBinding
	serviceMonitor   *monitoringv1.ServiceMonitor
	prometheusRule   *monitoringv1.PrometheusRule
}

func newMonolithReconciler(cmn *reconcilers.Instance) *monolithReconciler {
	rec := monolithReconciler{
		Instance:         cmn,
		daemonSet:        cmn.Managed.NewDaemonSet(monoName),
		promService:      cmn.Managed.NewService(monoPromService),
		serviceAccount:   cmn.Managed.NewServiceAccount(monoName),
		staticConfigMap:  cmn.Managed.NewConfigMap(monoConfigMap),
		dynamicConfigMap: cmn.Managed.NewConfigMap(monoDynConfigMap),
		rbConfigWatcher:  cmn.Managed.NewRB(resources.GetRoleBindingName(monoShortName, constants.ConfigWatcherRole)),
		rbHostNetwork:    cmn.Managed.NewCRB(resources.GetClusterRoleBindingName(monoShortName, constants.HostNetworkRole)),
		rbLokiWriter:     cmn.Managed.NewCRB(resources.GetClusterRoleBindingName(monoShortName, constants.LokiWriterRole)),
		rbInformer:       cmn.Managed.NewCRB(resources.GetClusterRoleBindingName(monoShortName, constants.FLPInformersRole)),
	}
	if cmn.ClusterInfo.HasSvcMonitor() {
		rec.serviceMonitor = cmn.Managed.NewServiceMonitor(monoServiceMonitor)
	}
	if cmn.ClusterInfo.HasPromRule() {
		rec.prometheusRule = cmn.Managed.NewPrometheusRule(monoPromRule)
	}
	return &rec
}

func (r *monolithReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithName("monolith")
	return log.IntoContext(ctx, l)
}

// cleanupNamespace cleans up old namespace
func (r *monolithReconciler) cleanupNamespace(ctx context.Context) {
	r.Managed.CleanupPreviousNamespace(ctx)
}

func (r *monolithReconciler) getStatus() *status.Instance {
	return &r.Status
}

func (r *monolithReconciler) reconcile(ctx context.Context, desired *flowslatest.FlowCollector, flowMetrics *metricslatest.FlowMetricList, detectedSubnets []flowslatest.SubnetLabel) error {
	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if helper.UseKafka(&desired.Spec) {
		r.Status.SetUnused("Monolith only used without Kafka")
		r.Managed.TryDeleteAll(ctx)
		return nil
	}

	r.Status.SetReady() // will be overidden if necessary, as error or pending

	builder, err := newMonolithBuilder(r.Instance, &desired.Spec, flowMetrics, detectedSubnets)
	if err != nil {
		return err
	}
	staticCM, configDigest, dynCM, err := builder.configMaps()
	if err != nil {
		return err
	}
	annotations := map[string]string{
		constants.PodConfigurationDigest: configDigest,
	}
	if !r.Managed.Exists(r.staticConfigMap) {
		if err := r.CreateOwned(ctx, staticCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(staticCM.Data, r.staticConfigMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.staticConfigMap, staticCM); err != nil {
			return err
		}
	}

	if err := r.reconcileDynamicConfigMap(ctx, dynCM); err != nil {
		return err
	}

	if err := r.reconcilePermissions(ctx, &builder); err != nil {
		return err
	}

	err = r.reconcilePrometheusService(ctx, &builder)
	if err != nil {
		return err
	}

	if helper.UseLoki(&desired.Spec) {
		// Watch for Loki certificate if necessary; we'll ignore in that case the returned digest, as we don't need to restart pods on cert rotation
		// because certificate is always reloaded from file
		if _, err = r.Watcher.ProcessCACert(ctx, r.Client, &r.Loki.TLS, r.Namespace); err != nil {
			return err
		}
	}

	// Watch for Kafka exporter certificate if necessary; need to restart pods in case of cert rotation
	if err = annotateKafkaExporterCerts(ctx, r.Common, desired.Spec.Exporters, annotations); err != nil {
		return err
	}

	// Watch for monitoring caCert
	if err = reconcileMonitoringCerts(ctx, r.Common, &desired.Spec.Processor.Metrics.Server.TLS, r.Namespace); err != nil {
		return err
	}

	return r.reconcileDaemonSet(ctx, builder.daemonSet(annotations))
}

func (r *monolithReconciler) reconcileDynamicConfigMap(ctx context.Context, newDCM *corev1.ConfigMap) error {
	if !r.Managed.Exists(r.dynamicConfigMap) {
		if err := r.CreateOwned(ctx, newDCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newDCM.Data, r.dynamicConfigMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.dynamicConfigMap, newDCM); err != nil {
			return err
		}
	}
	return nil
}

func (r *monolithReconciler) reconcilePrometheusService(ctx context.Context, builder *monolithBuilder) error {
	report := helper.NewChangeReport("FLP prometheus service")
	defer report.LogIfNeeded(ctx)

	if err := r.ReconcileService(ctx, r.promService, builder.promService(), &report); err != nil {
		return err
	}
	if r.ClusterInfo.HasSvcMonitor() {
		serviceMonitor := builder.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	if r.ClusterInfo.HasPromRule() {
		promRules := builder.prometheusRule()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.prometheusRule, promRules, &report, helper.PrometheusRuleChanged); err != nil {
			return err
		}
	}
	return nil
}

func (r *monolithReconciler) reconcileDaemonSet(ctx context.Context, desiredDS *appsv1.DaemonSet) error {
	report := helper.NewChangeReport("FLP DaemonSet")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileDaemonSet(
		ctx,
		r.Instance,
		r.daemonSet,
		desiredDS,
		constants.FLPName,
		&report,
	)
}

func (r *monolithReconciler) reconcilePermissions(ctx context.Context, builder *monolithBuilder) error {
	if !r.Managed.Exists(r.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now

	// Informers
	r.rbInformer = resources.GetClusterRoleBinding(r.Namespace, monoShortName, monoName, monoName, constants.FLPInformersRole)
	if err := r.ReconcileClusterRoleBinding(ctx, r.rbInformer); err != nil {
		return err
	}

	// Host network
	if r.ClusterInfo.IsOpenShift() {
		r.rbHostNetwork = resources.GetClusterRoleBinding(r.Namespace, monoShortName, monoName, monoName, constants.HostNetworkRole)
		if err := r.ReconcileClusterRoleBinding(ctx, r.rbHostNetwork); err != nil {
			return err
		}
	} else {
		r.Managed.TryDelete(ctx, r.rbHostNetwork)
	}

	// Loki writer
	if helper.UseLoki(builder.desired) && builder.desired.Loki.Mode == flowslatest.LokiModeLokiStack {
		r.rbLokiWriter = resources.GetClusterRoleBinding(r.Namespace, monoShortName, monoName, monoName, constants.LokiWriterRole)
		if err := r.ReconcileClusterRoleBinding(ctx, r.rbLokiWriter); err != nil {
			return err
		}
	} else {
		r.Managed.TryDelete(ctx, r.rbLokiWriter)
	}

	// Config watcher
	r.rbConfigWatcher = resources.GetRoleBinding(r.Namespace, monoShortName, monoName, monoName, constants.ConfigWatcherRole, true)
	if err := r.ReconcileRoleBinding(ctx, r.rbConfigWatcher); err != nil {
		return err
	}

	return nil
}
