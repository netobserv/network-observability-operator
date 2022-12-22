package flowlogspipeline

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
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
	}
	info.nobjMngr.AddManagedObject(name, owned.daemonSet)
	info.nobjMngr.AddManagedObject(name, owned.serviceAccount)
	info.nobjMngr.AddManagedObject(promServiceName(ConfKafkaIngester), owned.promService)
	info.nobjMngr.AddManagedObject(RoleBindingName(ConfKafkaIngester), owned.roleBinding)
	info.nobjMngr.AddManagedObject(configMapName(ConfKafkaIngester), owned.configMap)
	if info.availableAPIs.HasSvcMonitor() {
		info.nobjMngr.AddManagedObject(serviceMonitorName(ConfKafkaIngester), owned.serviceMonitor)
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

// initStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *flpIngesterReconciler) initStaticResources(ctx context.Context) error {
	cr := buildClusterRoleIngester(r.useOpenShiftSCC)
	return r.ReconcileClusterRole(ctx, cr)
}

// PrepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *flpIngesterReconciler) prepareNamespaceChange(ctx context.Context) error {
	// Switching namespace => delete everything in the previous namespace
	r.nobjMngr.CleanupPreviousNamespace(ctx)
	cr := buildClusterRoleIngester(r.useOpenShiftSCC)
	return r.ReconcileClusterRole(ctx, cr)
}

func (r *flpIngesterReconciler) reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	// Retrieve current owned objects
	err := r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Ingester only used with Kafka and without eBPF
	if !desired.Spec.UseKafka() || desired.Spec.UseEBPF() {
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

	return r.reconcileDaemonSet(ctx, &desired.Spec.Processor, &builder, configDigest)
}

func (r *flpIngesterReconciler) reconcilePrometheusService(ctx context.Context, builder *ingestBuilder) error {
	if !r.nobjMngr.Exists(r.owned.promService) {
		if err := r.CreateOwned(ctx, builder.newPromService()); err != nil {
			return err
		}
		if r.availableAPIs.HasSvcMonitor() {
			if err := r.CreateOwned(ctx, builder.generic.serviceMonitor()); err != nil {
				return err
			}
		}
		return nil
	}
	newSVC := builder.fromPromService(r.owned.promService)
	if serviceNeedsUpdate(r.owned.promService, newSVC) {
		if err := r.UpdateOwned(ctx, r.owned.promService, newSVC); err != nil {
			return err
		}
	}
	return nil
}

func (r *flpIngesterReconciler) reconcileDaemonSet(ctx context.Context, desiredFLP *flpSpec, builder *ingestBuilder, configDigest string) error {
	if !r.nobjMngr.Exists(r.owned.daemonSet) {
		return r.CreateOwned(ctx, builder.daemonSet(configDigest))
	} else if daemonSetNeedsUpdate(r.owned.daemonSet, desiredFLP, r.image, configDigest) {
		return r.UpdateOwned(ctx, r.owned.daemonSet, builder.daemonSet(configDigest))
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

	desired := builder.clusterRoleBinding()
	if err := r.ClientHelper.ReconcileClusterRoleBinding(ctx, desired); err != nil {
		return err
	}
	return nil
}
