package flowlogspipeline

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/discover"
)

// flpTransformerReconciler reconciles the current flowlogs-pipeline-transformer state with the desired configuration
type flpTransformerReconciler struct {
	singleReconciler
	reconcilers.ClientHelper
	nobjMngr        *reconcilers.NamespacedObjectManager
	owned           transfoOwnedObjects
	useOpenShiftSCC bool
	image           string
}

type transfoOwnedObjects struct {
	deployment     *appsv1.Deployment
	promService    *corev1.Service
	hpa            *ascv2.HorizontalPodAutoscaler
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	roleBinding    *rbacv1.ClusterRoleBinding
}

func newTransformerReconciler(ctx context.Context, cl reconcilers.ClientHelper, ns, prevNS, image string, permissionsVendor *discover.Permissions) *flpTransformerReconciler {
	name := name(ConfKafkaTransformer)
	owned := transfoOwnedObjects{
		deployment:     &appsv1.Deployment{},
		promService:    &corev1.Service{},
		hpa:            &ascv2.HorizontalPodAutoscaler{},
		serviceAccount: &corev1.ServiceAccount{},
		configMap:      &corev1.ConfigMap{},
		roleBinding:    &rbacv1.ClusterRoleBinding{},
	}
	nobjMngr := reconcilers.NewNamespacedObjectManager(cl, ns, prevNS)
	nobjMngr.AddManagedObject(name, owned.deployment)
	nobjMngr.AddManagedObject(name, owned.hpa)
	nobjMngr.AddManagedObject(name, owned.serviceAccount)
	nobjMngr.AddManagedObject(promServiceName(ConfKafkaTransformer), owned.promService)
	nobjMngr.AddManagedObject(RoleBindingName(ConfKafkaTransformer), owned.roleBinding)
	nobjMngr.AddManagedObject(configMapName(ConfKafkaTransformer), owned.configMap)

	openshift := permissionsVendor.Vendor(ctx) == discover.VendorOpenShift

	return &flpTransformerReconciler{
		ClientHelper:    cl,
		nobjMngr:        nobjMngr,
		owned:           owned,
		useOpenShiftSCC: openshift,
		image:           image,
	}
}

func (r *flpTransformerReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithValues(contextReconcilerName, "transformer")
	return log.IntoContext(ctx, l)
}

// initStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *flpTransformerReconciler) initStaticResources(ctx context.Context) error {
	cr := buildClusterRoleTransformer()
	return r.ReconcileClusterRole(ctx, cr)
}

// prepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *flpTransformerReconciler) prepareNamespaceChange(ctx context.Context) error {
	// Switching namespace => delete everything in the previous namespace
	r.nobjMngr.CleanupPreviousNamespace(ctx)
	cr := buildClusterRoleTransformer()
	return r.ReconcileClusterRole(ctx, cr)
}

func (r *flpTransformerReconciler) reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	// Retrieve current owned objects
	err := r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Transformer only used with Kafka
	if !desired.Spec.UseKafka() {
		r.nobjMngr.TryDeleteAll(ctx)
		return nil
	}

	builder := newTransfoBuilder(r.nobjMngr.Namespace, r.image, &desired.Spec, r.useOpenShiftSCC)
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

	return r.reconcileDeployment(ctx, &desired.Spec.Processor, &builder, configDigest)
}

func (r *flpTransformerReconciler) reconcileDeployment(ctx context.Context, desiredFLP *flpSpec, builder *transfoBuilder, configDigest string) error {
	ns := r.nobjMngr.Namespace

	if !r.nobjMngr.Exists(r.owned.deployment) {
		if err := r.CreateOwned(ctx, builder.deployment(configDigest)); err != nil {
			return err
		}
	} else if r.deploymentNeedsUpdate(r.owned.deployment, desiredFLP, configDigest) {
		if err := r.UpdateOwned(ctx, r.owned.deployment, builder.deployment(configDigest)); err != nil {
			return err
		}
	} else {
		// Deployment up to date, check if it's ready
		r.CheckDeploymentInProgress(r.owned.deployment)
	}

	// Delete or Create / Update Autoscaler according to HPA option
	if desiredFLP.KafkaConsumerAutoscaler.Disabled() {
		r.nobjMngr.TryDelete(ctx, r.owned.hpa)
	} else {
		newASC := builder.autoScaler()
		if !r.nobjMngr.Exists(r.owned.hpa) {
			if err := r.CreateOwned(ctx, newASC); err != nil {
				return err
			}
		} else if autoScalerNeedsUpdate(r.owned.hpa, desiredFLP.KafkaConsumerAutoscaler, ns) {
			if err := r.UpdateOwned(ctx, r.owned.hpa, newASC); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *flpTransformerReconciler) reconcilePrometheusService(ctx context.Context, builder *transfoBuilder) error {
	if !r.nobjMngr.Exists(r.owned.promService) {
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

func (r *flpTransformerReconciler) reconcilePermissions(ctx context.Context, builder *transfoBuilder) error {
	if !r.nobjMngr.Exists(r.owned.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now

	desired := builder.clusterRoleBinding()
	if err := r.ReconcileClusterRoleBinding(ctx, desired); err != nil {
		return err
	}
	return nil
}

func (r *flpTransformerReconciler) deploymentNeedsUpdate(depl *appsv1.Deployment, desired *flpSpec, configDigest string) bool {
	return containerNeedsUpdate(&depl.Spec.Template.Spec, desired, r.image) ||
		configChanged(&depl.Spec.Template, configDigest) ||
		(desired.KafkaConsumerAutoscaler.Disabled() && *depl.Spec.Replicas != desired.KafkaConsumerReplicas)
}

func autoScalerNeedsUpdate(asc *ascv2.HorizontalPodAutoscaler, desired flowsv1alpha1.FlowCollectorHPA, ns string) bool {
	if asc.Namespace != ns {
		return true
	}
	differentPointerValues := func(a, b *int32) bool {
		return (a == nil && b != nil) || (a != nil && b == nil) || (a != nil && *a != *b)
	}
	if asc.Spec.MaxReplicas != desired.MaxReplicas ||
		differentPointerValues(asc.Spec.MinReplicas, desired.MinReplicas) {
		return true
	}
	if !equality.Semantic.DeepDerivative(desired.Metrics, asc.Spec.Metrics) {
		return true
	}
	return false
}
