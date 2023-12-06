package flp

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
)

type monolithReconciler struct {
	*reconcilers.Instance
	daemonSet      *appsv1.DaemonSet
	promService    *corev1.Service
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	roleBindingIn  *rbacv1.ClusterRoleBinding
	roleBindingTr  *rbacv1.ClusterRoleBinding
	serviceMonitor *monitoringv1.ServiceMonitor
	prometheusRule *monitoringv1.PrometheusRule
}

func newMonolithReconciler(cmn *reconcilers.Instance) *monolithReconciler {
	name := name(ConfMonolith)
	rec := monolithReconciler{
		Instance:       cmn,
		daemonSet:      cmn.Managed.NewDaemonSet(name),
		promService:    cmn.Managed.NewService(promServiceName(ConfMonolith)),
		serviceAccount: cmn.Managed.NewServiceAccount(name),
		configMap:      cmn.Managed.NewConfigMap(configMapName(ConfMonolith)),
		roleBindingIn:  cmn.Managed.NewCRB(RoleBindingMonoName(ConfKafkaIngester)),
		roleBindingTr:  cmn.Managed.NewCRB(RoleBindingMonoName(ConfKafkaTransformer)),
	}
	if cmn.AvailableAPIs.HasSvcMonitor() {
		rec.serviceMonitor = cmn.Managed.NewServiceMonitor(serviceMonitorName(ConfMonolith))
	}
	if cmn.AvailableAPIs.HasPromRule() {
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

func (r *monolithReconciler) reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
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

	builder, err := newMonolithBuilder(r.Instance, &desired.Spec)
	if err != nil {
		return err
	}
	newCM, configDigest, err := builder.configMap()
	if err != nil {
		return err
	}
	annotations := map[string]string{
		constants.PodConfigurationDigest: configDigest,
	}
	if !r.Managed.Exists(r.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newCM.Data, r.configMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.configMap, newCM); err != nil {
			return err
		}
	}

	if err := r.reconcilePermissions(ctx, &builder); err != nil {
		return err
	}

	err = r.reconcilePrometheusService(ctx, &builder)
	if err != nil {
		return err
	}

	// Watch for Loki certificate if necessary; we'll ignore in that case the returned digest, as we don't need to restart pods on cert rotation
	// because certificate is always reloaded from file
	if _, err = r.Watcher.ProcessCACert(ctx, r.Client, &r.Loki.TLS, r.Namespace); err != nil {
		return err
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

func (r *monolithReconciler) reconcilePrometheusService(ctx context.Context, builder *monolithBuilder) error {
	report := helper.NewChangeReport("FLP prometheus service")
	defer report.LogIfNeeded(ctx)

	if err := r.ReconcileService(ctx, r.promService, builder.promService(), &report); err != nil {
		return err
	}
	if r.AvailableAPIs.HasSvcMonitor() {
		serviceMonitor := builder.generic.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	if r.AvailableAPIs.HasPromRule() {
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

	cr := buildClusterRoleIngester(r.UseOpenShiftSCC)
	if err := r.ReconcileClusterRole(ctx, cr); err != nil {
		return err
	}
	cr = BuildClusterRoleTransformer()
	if err := r.ReconcileClusterRole(ctx, cr); err != nil {
		return err
	}
	// Monolith uses ingester + transformer cluster roles
	for _, kind := range []ConfKind{ConfKafkaIngester, ConfKafkaTransformer} {
		desired := builder.clusterRoleBinding(kind)
		if err := r.ReconcileClusterRoleBinding(ctx, desired); err != nil {
			return err
		}
	}

	return reconcileLokiRoles(ctx, r.Common, &builder.generic)
}
