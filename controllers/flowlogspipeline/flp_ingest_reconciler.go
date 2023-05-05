package flowlogspipeline

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

// flpIngesterReconciler reconciles the current flowlogs-pipeline-ingester state with the desired configuration
type flpIngesterReconciler struct {
	singleReconciler
	reconcilersCommonInfo
	owned ingestOwnedObjects
}

type ingestOwnedObjects struct {
	daemonSet      *appsv1.DaemonSet
	promService    *corev1.Service
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	roleBinding    *rbacv1.ClusterRoleBinding
	serviceMonitor *monitoringv1.ServiceMonitor
	prometheusRule *monitoringv1.PrometheusRule
}

func newIngesterReconciler(info *reconcilersCommonInfo) *flpIngesterReconciler {
	name := name(ConfKafkaIngester)
	owned := ingestOwnedObjects{
		daemonSet:      &appsv1.DaemonSet{},
		promService:    &corev1.Service{},
		serviceAccount: &corev1.ServiceAccount{},
		configMap:      &corev1.ConfigMap{},
		roleBinding:    &rbacv1.ClusterRoleBinding{},
		serviceMonitor: &monitoringv1.ServiceMonitor{},
		prometheusRule: &monitoringv1.PrometheusRule{},
	}
	info.nobjMngr.AddManagedObject(name, owned.daemonSet)
	info.nobjMngr.AddManagedObject(name, owned.serviceAccount)
	info.nobjMngr.AddManagedObject(promServiceName(ConfKafkaIngester), owned.promService)
	info.nobjMngr.AddManagedObject(RoleBindingName(ConfKafkaIngester), owned.roleBinding)
	info.nobjMngr.AddManagedObject(configMapName(ConfKafkaIngester), owned.configMap)
	if info.availableAPIs.HasSvcMonitor() {
		info.nobjMngr.AddManagedObject(serviceMonitorName(ConfKafkaIngester), owned.serviceMonitor)
	}
	if info.availableAPIs.HasPromRule() {
		info.nobjMngr.AddManagedObject(prometheusRuleName(ConfKafkaIngester), owned.prometheusRule)
	}

	return &flpIngesterReconciler{
		reconcilersCommonInfo: *info,
		owned:                 owned,
	}
}

func (r *flpIngesterReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithValues(contextReconcilerName, "ingester")
	return log.IntoContext(ctx, l)
}

// cleanupNamespace cleans up old namespace
func (r *flpIngesterReconciler) cleanupNamespace(ctx context.Context) {
	r.nobjMngr.CleanupPreviousNamespace(ctx)
}

func (r *flpIngesterReconciler) reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	// Retrieve current owned objects
	err := r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Ingester only used with Kafka and without eBPF
	if !helper.UseKafka(&desired.Spec) || helper.UseEBPF(&desired.Spec) {
		r.nobjMngr.TryDeleteAll(ctx)
		return nil
	}

	builder := newIngestBuilder(r.nobjMngr.Namespace, r.image, &desired.Spec, r.useOpenShiftSCC)
	newCM, configDigest, err := builder.configMap()
	if err != nil {
		return err
	}
	if !r.nobjMngr.Exists(r.owned.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newCM.Data, r.owned.configMap.Data) {
		if err := r.UpdateOwned(ctx, r.owned.configMap, newCM); err != nil {
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

	return r.reconcileDaemonSet(ctx, builder.daemonSet(configDigest))
}

func (r *flpIngesterReconciler) reconcilePrometheusService(ctx context.Context, builder *ingestBuilder) error {
	report := helper.NewChangeReport("FLP prometheus service")
	defer report.LogIfNeeded(ctx)

	if err := reconcilers.ReconcileService(ctx, r.nobjMngr, &r.ClientHelper, r.owned.promService, builder.promService(), &report); err != nil {
		return err
	}
	if r.availableAPIs.HasSvcMonitor() {
		serviceMonitor := builder.generic.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.nobjMngr, &r.ClientHelper, r.owned.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	if r.availableAPIs.HasPromRule() {
		promRules := builder.generic.prometheusRule()
		if err := reconcilers.GenericReconcile(ctx, r.nobjMngr, &r.ClientHelper, r.owned.prometheusRule, promRules, &report, helper.PrometheusRuleChanged); err != nil {
			return err
		}
	}
	return nil
}

func (r *flpIngesterReconciler) reconcileDaemonSet(ctx context.Context, desiredDS *appsv1.DaemonSet) error {
	report := helper.NewChangeReport("FLP DaemonSet")
	defer report.LogIfNeeded(ctx)

	if !r.nobjMngr.Exists(r.owned.daemonSet) {
		return r.CreateOwned(ctx, desiredDS)
	} else if helper.PodChanged(&r.owned.daemonSet.Spec.Template, &desiredDS.Spec.Template, constants.FLPName, &report) {
		return r.UpdateOwned(ctx, r.owned.daemonSet, desiredDS)
	} else {
		// DaemonSet up to date, check if it's ready
		r.CheckDaemonSetInProgress(r.owned.daemonSet)
	}
	return nil
}

func (r *flpIngesterReconciler) reconcilePermissions(ctx context.Context, builder *ingestBuilder) error {
	if !r.nobjMngr.Exists(r.owned.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now

	cr := buildClusterRoleIngester(r.useOpenShiftSCC)
	if err := r.ReconcileClusterRole(ctx, cr); err != nil {
		return err
	}

	desired := builder.clusterRoleBinding()
	if err := r.ClientHelper.ReconcileClusterRoleBinding(ctx, desired); err != nil {
		return err
	}
	return nil
}
