package controllers

import (
	"context"

	"github.com/netobserv/network-observability-operator/pkg/helper"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/netobserv/network-observability-operator/controllers/ovs"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/goflowkube"
)

// Make sure it always matches config/default/kustomization.yaml:namespace
// See also https://github.com/operator-framework/operator-lib/issues/74
const operatorNamespace = "network-observability"
const cnoNamespace = "openshift-network-operator"
const ovsFlowsConfigMapName = "ovs-flows-config"
const finalizerName = "flows.netobserv.io/finalizer"

// FlowCollectorReconciler reconciles a FlowCollector object
type FlowCollectorReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	ovsConfigController ovs.FlowsConfigController
}

func NewFlowCollectorReconciler(client client.Client, scheme *runtime.Scheme) *FlowCollectorReconciler {
	return &FlowCollectorReconciler{
		Client: client,
		Scheme: scheme,
		ovsConfigController: ovs.NewFlowsConfigController(client,
			operatorNamespace, cnoNamespace, ovsFlowsConfigMapName),
	}
}

//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services;configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;create;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;create;delete
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the FlowCollector object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *FlowCollectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	desired := &flowsv1alpha1.FlowCollector{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("FlowCollector resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get FlowCollector")
		return ctrl.Result{}, err
	}

	if err := r.checkFinalizerStatus(ctx, desired); err != nil {
		return ctrl.Result{}, err
	}

	// Goflow
	gfReconciler := goflowkube.Reconciler{
		Client: r.Client,
		SetControllerReference: func(obj client.Object) error {
			return ctrl.SetControllerReference(desired, obj, r.Scheme)
		},
		OperatorNamespace: operatorNamespace,
	}
	if err := gfReconciler.Reconcile(ctx, &desired.Spec.GoflowKube, &desired.Spec.Loki); err != nil {
		log.Error(err, "Failed to get FlowCollector")
		return ctrl.Result{}, err
	}
	// ovs-flows-config map for CNO
	if err := r.ovsConfigController.Reconcile(ctx, desired); err != nil {
		log.Error(err, "Failed to reconcile ovs-flows-config ConfigMap")
	}

	return ctrl.Result{}, nil
}

func (r *FlowCollectorReconciler) checkFinalizerStatus(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	log := log.FromContext(ctx)
	if desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// the object is not being deleted. Register the required finalizer if not already there
		if !helper.ContainsString(desired.GetFinalizers(), finalizerName) {
			log.Info("registering finalizer, if not already present", "finalizerName", finalizerName)
			controllerutil.AddFinalizer(desired, finalizerName)
			return r.Update(ctx, desired)
		}
	} else {
		// the object is being deleted. Execute finalizer, if present, and unregister it
		if helper.ContainsString(desired.GetFinalizers(), finalizerName) {
			log.Info("deleting external resources", "finalizerName", finalizerName)
			if err := r.deleteExternalResources(ctx); err != nil {
				return err
			}
			controllerutil.RemoveFinalizer(desired, finalizerName)
			return r.Update(ctx, desired)
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlowCollectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&flowsv1alpha1.FlowCollector{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *FlowCollectorReconciler) deleteExternalResources(ctx context.Context) error {
	return r.ovsConfigController.Finalize(ctx)
}
