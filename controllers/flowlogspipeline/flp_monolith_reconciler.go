package flowlogspipeline

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
)

// flpMonolithReconciler reconciles the current flowlogs-pipeline monolith state with the desired configuration
type flpMonolithReconciler struct {
	*reconcilers.Instance
	owned monolithOwnedObjects
}

type monolithOwnedObjects struct {
	daemonSet      *appsv1.DaemonSet
	promService    *corev1.Service
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	roleBindingIn  *rbacv1.ClusterRoleBinding
	roleBindingTr  *rbacv1.ClusterRoleBinding
	serviceMonitor *monitoringv1.ServiceMonitor
	prometheusRule *monitoringv1.PrometheusRule
}

func newMonolithReconciler(cmn *reconcilers.Instance) *flpMonolithReconciler {
	name := name(ConfMonolith)
	owned := monolithOwnedObjects{
		daemonSet:      &appsv1.DaemonSet{},
		promService:    &corev1.Service{},
		serviceAccount: &corev1.ServiceAccount{},
		configMap:      &corev1.ConfigMap{},
		roleBindingIn:  &rbacv1.ClusterRoleBinding{},
		roleBindingTr:  &rbacv1.ClusterRoleBinding{},
		serviceMonitor: &monitoringv1.ServiceMonitor{},
		prometheusRule: &monitoringv1.PrometheusRule{},
	}
	cmn.Managed.AddManagedObject(name, owned.daemonSet)
	cmn.Managed.AddManagedObject(name, owned.serviceAccount)
	cmn.Managed.AddManagedObject(promServiceName(ConfMonolith), owned.promService)
	cmn.Managed.AddManagedObject(RoleBindingMonoName(ConfKafkaIngester), owned.roleBindingIn)
	cmn.Managed.AddManagedObject(RoleBindingMonoName(ConfKafkaTransformer), owned.roleBindingTr)
	cmn.Managed.AddManagedObject(configMapName(ConfMonolith), owned.configMap)
	if cmn.AvailableAPIs.HasSvcMonitor() {
		cmn.Managed.AddManagedObject(serviceMonitorName(ConfMonolith), owned.serviceMonitor)
	}
	if cmn.AvailableAPIs.HasPromRule() {
		cmn.Managed.AddManagedObject(prometheusRuleName(ConfMonolith), owned.prometheusRule)
	}

	return &flpMonolithReconciler{
		Instance: cmn,
		owned:    owned,
	}
}

func (r *flpMonolithReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithValues(contextReconcilerName, "monolith")
	return log.IntoContext(ctx, l)
}

// cleanupNamespace cleans up old namespace
func (r *flpMonolithReconciler) cleanupNamespace(ctx context.Context) {
	r.Managed.CleanupPreviousNamespace(ctx)
}

func (r *flpMonolithReconciler) reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Monolith only used without Kafka
	if helper.UseKafka(&desired.Spec) {
		r.Managed.TryDeleteAll(ctx)
		return nil
	}

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
	if !r.Managed.Exists(r.owned.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newCM.Data, r.owned.configMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.owned.configMap, newCM); err != nil {
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

func (r *flpMonolithReconciler) reconcilePrometheusService(ctx context.Context, builder *monolithBuilder) error {
	report := helper.NewChangeReport("FLP prometheus service")
	defer report.LogIfNeeded(ctx)

	if err := r.ReconcileService(ctx, r.owned.promService, builder.promService(), &report); err != nil {
		return err
	}
	if r.AvailableAPIs.HasSvcMonitor() {
		serviceMonitor := builder.generic.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.owned.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	if r.AvailableAPIs.HasPromRule() {
		promRules := builder.generic.prometheusRule()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.owned.prometheusRule, promRules, &report, helper.PrometheusRuleChanged); err != nil {
			return err
		}
	}
	return nil
}

func (r *flpMonolithReconciler) reconcileDaemonSet(ctx context.Context, desiredDS *appsv1.DaemonSet) error {
	report := helper.NewChangeReport("FLP DaemonSet")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileDaemonSet(
		ctx,
		r.Instance,
		r.owned.daemonSet,
		desiredDS,
		constants.FLPName,
		&report,
	)
}

func (r *flpMonolithReconciler) reconcilePermissions(ctx context.Context, builder *monolithBuilder) error {
	if !r.Managed.Exists(r.owned.serviceAccount) {
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
