package flowlogspipeline

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/discover"
)

// flpMonolithReconciler reconciles the current flowlogs-pipeline monolith state with the desired configuration
type flpMonolithReconciler struct {
	singleReconciler
	reconcilers.ClientHelper
	nobjMngr        *reconcilers.NamespacedObjectManager
	owned           monolithOwnedObjects
	useOpenShiftSCC bool
	image           string
}

type monolithOwnedObjects struct {
	daemonSet      *appsv1.DaemonSet
	promService    *corev1.Service
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	roleBindingIn  *rbacv1.ClusterRoleBinding
	roleBindingTr  *rbacv1.ClusterRoleBinding
}

func newMonolithReconciler(ctx context.Context, cl reconcilers.ClientHelper, ns, prevNS, image string, permissionsVendor *discover.Permissions) *flpMonolithReconciler {
	name := name(ConfMonolith)
	owned := monolithOwnedObjects{
		daemonSet:      &appsv1.DaemonSet{},
		promService:    &corev1.Service{},
		serviceAccount: &corev1.ServiceAccount{},
		configMap:      &corev1.ConfigMap{},
		roleBindingIn:  &rbacv1.ClusterRoleBinding{},
		roleBindingTr:  &rbacv1.ClusterRoleBinding{},
	}
	nobjMngr := reconcilers.NewNamespacedObjectManager(cl, ns, prevNS)
	nobjMngr.AddManagedObject(name, owned.daemonSet)
	nobjMngr.AddManagedObject(name, owned.serviceAccount)
	nobjMngr.AddManagedObject(promServiceName(ConfMonolith), owned.promService)
	nobjMngr.AddManagedObject(RoleBindingMonoName(ConfKafkaIngester), owned.roleBindingIn)
	nobjMngr.AddManagedObject(RoleBindingMonoName(ConfKafkaTransformer), owned.roleBindingTr)
	nobjMngr.AddManagedObject(configMapName(ConfMonolith), owned.configMap)

	openshift := permissionsVendor.Vendor(ctx) == discover.VendorOpenShift

	return &flpMonolithReconciler{
		ClientHelper:    cl,
		nobjMngr:        nobjMngr,
		owned:           owned,
		useOpenShiftSCC: openshift,
		image:           image,
	}
}

func (r *flpMonolithReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithValues(contextReconcilerName, "monolith")
	return log.IntoContext(ctx, l)
}

// initStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *flpMonolithReconciler) initStaticResources(ctx context.Context) error {
	// Nothing to do here: monolith FLP uses cluster roles defined by Transformer and Ingester reconcilers
	return nil
}

// prepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *flpMonolithReconciler) prepareNamespaceChange(ctx context.Context) error {
	// Switching namespace => delete everything in the previous namespace
	r.nobjMngr.CleanupPreviousNamespace(ctx)
	return nil
}

func (r *flpMonolithReconciler) reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	// Retrieve current owned objects
	err := r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Monolith only used without Kafka
	if desired.Spec.UseKafka() {
		r.nobjMngr.TryDeleteAll(ctx)
		return nil
	}

	builder := newMonolithBuilder(r.nobjMngr.Namespace, r.image, &desired.Spec, r.useOpenShiftSCC)
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

func (r *flpMonolithReconciler) reconcilePrometheusService(ctx context.Context, builder *monolithBuilder) error {
	if !r.nobjMngr.Exists(r.owned.promService) {
		if err := AddPrometheusServiceMonitor(ctx, &builder.generic, r.ClientHelper); err != nil {
			return err
		}
		return r.CreateOwned(ctx, builder.newPromService())
	}
	newSVC := builder.fromPromService(r.owned.promService)
	if serviceNeedsUpdate(r.owned.promService, newSVC) {
		if err := r.UpdateOwned(ctx, r.owned.promService, newSVC); err != nil {
			return err
		}
	}
	return nil
}

func (r *flpMonolithReconciler) reconcileDaemonSet(ctx context.Context, desiredFLP *flpSpec, builder *monolithBuilder, configDigest string) error {
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

func (r *flpMonolithReconciler) reconcilePermissions(ctx context.Context, builder *monolithBuilder) error {
	if !r.nobjMngr.Exists(r.owned.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now

	// Monolith uses ingester + transformer cluster roles
	for _, kind := range []ConfKind{ConfKafkaIngester, ConfKafkaTransformer} {
		desired := builder.clusterRoleBinding(kind)
		if err := r.ClientHelper.ReconcileClusterRoleBinding(ctx, desired); err != nil {
			return err
		}
	}
	return nil
}
