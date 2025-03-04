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
	roleBindingIn    *rbacv1.ClusterRoleBinding
	roleBindingTr    *rbacv1.ClusterRoleBinding
	serviceMonitor   *monitoringv1.ServiceMonitor
	prometheusRule   *monitoringv1.PrometheusRule
}

func newMonolithReconciler(cmn *reconcilers.Instance) *monolithReconciler {
	name := name(ConfMonolith)
	rec := monolithReconciler{
		Instance:         cmn,
		daemonSet:        cmn.Managed.NewDaemonSet(name),
		promService:      cmn.Managed.NewService(promServiceName(ConfMonolith)),
		serviceAccount:   cmn.Managed.NewServiceAccount(name),
		staticConfigMap:  cmn.Managed.NewConfigMap(staticConfigMapName(ConfMonolith)),
		dynamicConfigMap: cmn.Managed.NewConfigMap(dynamicConfigMapName(ConfMonolith)),
		roleBindingIn:    cmn.Managed.NewCRB(RoleBindingMonoName(ConfKafkaIngester)),
		roleBindingTr:    cmn.Managed.NewCRB(RoleBindingMonoName(ConfKafkaTransformer)),
	}
	if cmn.ClusterInfo.HasSvcMonitor() {
		rec.serviceMonitor = cmn.Managed.NewServiceMonitor(serviceMonitorName(ConfMonolith))
	}
	if cmn.ClusterInfo.HasPromRule() {
		rec.prometheusRule = cmn.Managed.NewPrometheusRule(prometheusRuleName(ConfMonolith))
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
	newSCM, configDigest, err := builder.staticConfigMap()
	if err != nil {
		return err
	}
	annotations := map[string]string{
		constants.PodConfigurationDigest: configDigest,
	}
	if !r.Managed.Exists(r.staticConfigMap) {
		if err := r.CreateOwned(ctx, newSCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newSCM.Data, r.staticConfigMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.staticConfigMap, newSCM); err != nil {
			return err
		}
	}

	if err := r.reconcileDynamicConfigMap(ctx, &builder); err != nil {
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

func (r *monolithReconciler) reconcileDynamicConfigMap(ctx context.Context, builder *monolithBuilder) error {
	newDCM, err := builder.dynamicConfigMap()
	if err != nil {
		return err
	}
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
		serviceMonitor := builder.generic.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	if r.ClusterInfo.HasPromRule() {
		promRules := builder.generic.prometheusRule()
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

	roles := []constants.ClusterRoleName{
		constants.FLPInformersRole,
	}
	if r.ClusterInfo.IsOpenShift() {
		roles = append(roles, constants.HostNetworkRole)
	}
	if helper.UseLoki(builder.generic.desired) && builder.generic.desired.Loki.Mode == flowslatest.LokiModeLokiStack {
		roles = append(roles, constants.LokiWriterRole)
	}
	bindings, crBindings := resources.GetAllBindings(
		r.Namespace,
		builder.generic.name(),
		builder.generic.name(),
		[]constants.RoleName{constants.ConfigWatcherRole},
		roles,
	)
	if err := r.ReconcileClusterRoleBindings(ctx, crBindings); err != nil {
		return err
	}
	return r.ReconcileRoleBindings(ctx, bindings)
}
