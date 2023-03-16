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

// flpMonolithReconciler reconciles the current flowlogs-pipeline monolith state with the desired configuration
type flpMonolithReconciler struct {
	singleReconciler
	reconcilersCommonInfo
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

func newMonolithReconciler(info *reconcilersCommonInfo) *flpMonolithReconciler {
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
	info.nobjMngr.AddManagedObject(name, owned.daemonSet)
	info.nobjMngr.AddManagedObject(name, owned.serviceAccount)
	info.nobjMngr.AddManagedObject(promServiceName(ConfMonolith), owned.promService)
	info.nobjMngr.AddManagedObject(RoleBindingMonoName(ConfKafkaIngester), owned.roleBindingIn)
	info.nobjMngr.AddManagedObject(RoleBindingMonoName(ConfKafkaTransformer), owned.roleBindingTr)
	info.nobjMngr.AddManagedObject(configMapName(ConfMonolith), owned.configMap)
	if info.availableAPIs.HasSvcMonitor() {
		info.nobjMngr.AddManagedObject(serviceMonitorName(ConfMonolith), owned.serviceMonitor)
	}
	if info.availableAPIs.HasPromRule() {
		info.nobjMngr.AddManagedObject(prometheusRuleName(ConfMonolith), owned.prometheusRule)
	}

	return &flpMonolithReconciler{
		reconcilersCommonInfo: *info,
		owned:                 owned,
	}
}

func (r *flpMonolithReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithValues(contextReconcilerName, "monolith")
	return log.IntoContext(ctx, l)
}

// cleanupNamespace cleans up old namespace
func (r *flpMonolithReconciler) cleanupNamespace(ctx context.Context) {
	r.nobjMngr.CleanupPreviousNamespace(ctx)
}

func (r *flpMonolithReconciler) reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	// Retrieve current owned objects
	err := r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Monolith only used without Kafka
	if helper.UseKafka(&desired.Spec) {
		r.nobjMngr.TryDeleteAll(ctx)
		return nil
	}

	builder := newMonolithBuilder(r.nobjMngr.Namespace, r.image, &desired.Spec, r.useOpenShiftSCC, r.CertWatcher)
	newCM, configDigest, dbConfigMap, err := builder.configMap()
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
	if r.reconcilersCommonInfo.availableAPIs.HasConsoleConfig() {
		if err := r.reconcileDashboardConfig(ctx, dbConfigMap); err != nil {
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

func (r *flpMonolithReconciler) reconcilePrometheusService(ctx context.Context, builder *monolithBuilder) error {
	report := helper.NewChangeReport("FLP prometheus service")
	defer report.LogIfNeeded(ctx)

	if !r.nobjMngr.Exists(r.owned.promService) {
		if err := r.CreateOwned(ctx, builder.newPromService()); err != nil {
			return err
		}
	} else {
		newSVC := builder.fromPromService(r.owned.promService)
		if helper.ServiceChanged(r.owned.promService, newSVC, &report) {
			if err := r.UpdateOwned(ctx, r.owned.promService, newSVC); err != nil {
				return err
			}
		}
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

func (r *flpMonolithReconciler) reconcileDaemonSet(ctx context.Context, desiredDS *appsv1.DaemonSet) error {
	report := helper.NewChangeReport("FLP DaemonSet")
	defer report.LogIfNeeded(ctx)

	// Annotate pod with certificate reference so that it is reloaded if modified
	if err := r.CertWatcher.AnnotatePod(ctx, r.Client, &desiredDS.Spec.Template, lokiCerts, kafkaCerts); err != nil {
		return err
	}
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

func (r *flpMonolithReconciler) reconcilePermissions(ctx context.Context, builder *monolithBuilder) error {
	if !r.nobjMngr.Exists(r.owned.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now

	cr := buildClusterRoleIngester(r.useOpenShiftSCC)
	if err := r.ReconcileClusterRole(ctx, cr); err != nil {
		return err
	}
	cr = buildClusterRoleTransformer()
	if err := r.ReconcileClusterRole(ctx, cr); err != nil {
		return err
	}
	// Monolith uses ingester + transformer cluster roles
	for _, kind := range []ConfKind{ConfKafkaIngester, ConfKafkaTransformer} {
		desired := builder.clusterRoleBinding(kind)
		if err := r.ClientHelper.ReconcileClusterRoleBinding(ctx, desired); err != nil {
			return err
		}
	}
	return nil
}
