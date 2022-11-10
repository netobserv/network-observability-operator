package controllers

import (
	"context"
	"fmt"
	"net"
	"reflect"

	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	securityv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/consoleplugin"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/ebpf"
	"github.com/netobserv/network-observability-operator/controllers/flowlogspipeline"
	"github.com/netobserv/network-observability-operator/controllers/operators"
	"github.com/netobserv/network-observability-operator/controllers/ovs"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/controllers/secrets"
	"github.com/netobserv/network-observability-operator/pkg/conditions"
	"github.com/netobserv/network-observability-operator/pkg/discover"
)

const (
	ovsFlowsConfigMapName = "ovs-flows-config"
	flowsFinalizer        = "flows.netobserv.io/finalizer"
)

// FlowCollectorReconciler reconciles a FlowCollector object
type FlowCollectorReconciler struct {
	client.Client
	permissions   discover.Permissions
	availableAPIs *discover.AvailableAPIs
	Scheme        *runtime.Scheme
	lookupIP      func(string) ([]net.IP, error)
}

func NewFlowCollectorReconciler(client client.Client, scheme *runtime.Scheme) *FlowCollectorReconciler {
	return &FlowCollectorReconciler{
		Client:   client,
		Scheme:   scheme,
		lookupIP: net.LookupIP,
	}
}

//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces;services;serviceaccounts;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;create;delete;watch;list
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;create;delete;update;watch
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=get;create;delete;update;patch;list
//+kubebuilder:rbac:groups=operator.openshift.io,resources=consoles,verbs=get;update;list;update;watch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=hostnetwork,verbs=use
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=list;create;update;watch
//+kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,verbs=list;get;watch
//+kubebuilder:rbac:groups=operators.coreos.com,resources=operatorgroups;subscriptions,verbs=create;delete;patch;update;get;watch;list
//+kubebuilder:rbac:groups=kafka.strimzi.io,resources=kafkas;kafkatopics;kafkausers,verbs=create;delete;patch;update;get;watch;list
//+kubebuilder:rbac:groups=loki.grafana.com,resources=lokistacks;network,verbs=create;delete;patch;update;get;watch;list
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=list;get;watch

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

	desired, err := r.getFlowCollector(ctx, req)
	if err != nil {
		log.Error(err, "Failed to get FlowCollector")
		return ctrl.Result{}, err
	} else if desired == nil {
		return ctrl.Result{}, nil
	}

	ns := getNamespaceName(desired)
	// If namespace does not exist, we create it
	nsExist, err := r.namespaceExist(ctx, ns)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !nsExist {
		err = r.Create(ctx, buildNamespace(ns))
		if err != nil {
			return ctrl.Result{}, r.failure(ctx, conditions.CannotCreateNamespace(err), desired)
		}
	}

	clientHelper := r.newClientHelper(desired)

	// reconcile operator content only when CRD / Owned object has changed
	// this avoid watched objects like secret to call unecessary reconcile steps
	if req.Name == constants.Cluster {
		err := r.reconcileComponents(ctx, clientHelper, desired, ns)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// secrets sync between namespaces
	secretsReconciler := secrets.NewReconciler(clientHelper, ns, req)
	if err := secretsReconciler.Reconcile(ctx, &desired.Spec); err != nil {
		return ctrl.Result{}, r.failure(ctx, conditions.ReconcileSecretsFailed(err), desired)
	}

	// dependent operators
	operatorsReconciler := operators.NewReconciler(ctx, clientHelper, &desired.Spec, ns, req, r.permissions.Vendor(ctx) == discover.VendorOpenShift)
	conds, err := operatorsReconciler.Reconcile(clientHelper)
	if err != nil {
		return ctrl.Result{}, r.failure(ctx, conditions.ReconcileDependentOperatorsFailed(err), desired)
	}

	// manage conditions
	clientCondition := clientHelper.GetClientCondition()
	if clientCondition.Type != conditions.TypeReady || conditions.DependenciesReady(conds) {
		conds = append(conds, clientCondition)
	}
	r.updateConditions(ctx, &conds, desired)

	return ctrl.Result{}, nil
}

func (r *FlowCollectorReconciler) reconcileComponents(ctx context.Context, clientHelper reconcilers.ClientHelper, desired *flowsv1alpha1.FlowCollector, ns string) error {
	previousNamespace := desired.Status.Namespace

	// Create reconcilers
	flpReconciler := flowlogspipeline.NewReconciler(ctx, clientHelper, ns, previousNamespace, &r.permissions)
	var cpReconciler consoleplugin.CPReconciler
	if r.availableAPIs.HasConsole() {
		cpReconciler = consoleplugin.NewReconciler(clientHelper, ns, previousNamespace)
	}

	// Check namespace changed
	if ns != previousNamespace {
		if err := r.handleNamespaceChanged(ctx, previousNamespace, ns, desired, &flpReconciler, &cpReconciler); err != nil {
			return r.failure(ctx, conditions.CannotCreateNamespace(err), desired)
		}
	}

	// Flowlogs-pipeline
	if err := flpReconciler.Reconcile(ctx, desired); err != nil {
		return r.failure(ctx, conditions.ReconcileFLPFailed(err), desired)
	}

	// OVS config map for CNO
	if r.availableAPIs.HasCNO() {
		ovsConfigController := ovs.NewFlowsConfigCNOController(clientHelper,
			ns,
			desired.Spec.Agent.IPFIX.ClusterNetworkOperator.Namespace,
			ovsFlowsConfigMapName,
			r.lookupIP)
		if err := ovsConfigController.Reconcile(ctx, desired); err != nil {
			return r.failure(ctx, conditions.ReconcileCNOFailed(err), desired)
		}
	} else {
		ovsConfigController := ovs.NewFlowsConfigOVNKController(clientHelper,
			ns,
			desired.Spec.Agent.IPFIX.OVNKubernetes,
			r.lookupIP)
		if err := ovsConfigController.Reconcile(ctx, desired); err != nil {
			return r.failure(ctx, conditions.ReconcileOVNKFailed(err), desired)
		}
	}

	// eBPF agent
	ebpfAgentController := ebpf.NewAgentController(clientHelper, ns, &r.permissions)
	if err := ebpfAgentController.Reconcile(ctx, desired); err != nil {
		return r.failure(ctx, conditions.ReconcileAgentFailed(err), desired)
	}

	// Console plugin
	if r.availableAPIs.HasConsole() {
		err := cpReconciler.Reconcile(ctx, desired)
		if err != nil {
			return r.failure(ctx, conditions.ReconcileConsolePluginFailed(err), desired)
		}
	}

	return nil
}

func (r *FlowCollectorReconciler) getFlowCollector(ctx context.Context, req ctrl.Request) (*flowsv1alpha1.FlowCollector, error) {
	log := log.FromContext(ctx)
	desired := &flowsv1alpha1.FlowCollector{}
	if err := r.Get(ctx, types.NamespacedName{
		Name: constants.Cluster,
	}, desired); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			if req.Name == constants.Cluster {
				log.Info("FlowCollector resource not found. Ignoring since object must be deleted")
			}
			return nil, nil
		}
		// Error reading the object - requeue the request.
		return nil, err
	}

	if ret, err := r.checkFinalizer(ctx, desired); ret {
		return nil, err
	}
	return desired, nil
}

func (r *FlowCollectorReconciler) handleNamespaceChanged(
	ctx context.Context,
	oldNS, newNS string,
	desired *flowsv1alpha1.FlowCollector,
	flpReconciler *flowlogspipeline.FLPReconciler,
	cpReconciler *consoleplugin.CPReconciler,
) error {
	log := log.FromContext(ctx)
	if oldNS == "" {
		// First install: create one-shot resources
		log.Info("FlowCollector first install: creating initial resources")
		err := flpReconciler.InitStaticResources(ctx)
		if err != nil {
			return err
		}
		if r.availableAPIs.HasConsole() {
			err := cpReconciler.InitStaticResources(ctx)
			if err != nil {
				return err
			}
		}
	} else {
		// Namespace updated, clean up previous namespace
		log.Info("FlowCollector namespace change detected: cleaning up previous namespace and preparing next one", "old namespace", oldNS, "new namepace", newNS)
		err := flpReconciler.PrepareNamespaceChange(ctx)
		if err != nil {
			return err
		}
		if r.availableAPIs.HasConsole() {
			err := cpReconciler.PrepareNamespaceChange(ctx)
			if err != nil {
				return err
			}
		}
	}

	// Update namespace in status
	log.Info("Updating status with new namespace " + newNS)
	desired.Status.Namespace = newNS
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlowCollectorReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&flowsv1alpha1.FlowCollector{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&ascv2.HorizontalPodAutoscaler{}).
		Owns(&corev1.Namespace{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(
				predicate.Funcs{
					// trigger reconcile loop when secret match
					// based on name or labels containing either "loki" or "kafka"
					CreateFunc: func(e event.CreateEvent) bool {
						secret := e.Object.(*corev1.Secret)
						return reconcilers.MatchingSecret(secret)
					},
					// trigger reconcile loop when old / new secret match
					// and data is updated
					UpdateFunc: func(e event.UpdateEvent) bool {
						oldSecret := e.ObjectOld.(*corev1.Secret)
						secret := e.ObjectNew.(*corev1.Secret)
						return (reconcilers.MatchingSecret(secret) ||
							reconcilers.MatchingSecret(oldSecret)) &&
							!reflect.DeepEqual(oldSecret.Data, secret.Data)
					},
					DeleteFunc: func(e event.DeleteEvent) bool {
						return false
					},
					GenericFunc: func(e event.GenericEvent) bool {
						return false
					},
				},
			),
		).
		Watches(
			&source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(
				predicate.Funcs{
					// trigger reconcile loop on matching CRD names
					// dependent operators are created when CRD are available
					// but not owned by this operator
					CreateFunc: func(e event.CreateEvent) bool {
						crd := e.Object.(*apiextensionsv1.CustomResourceDefinition)
						return reconcilers.MatchingCRD(crd.Name)
					},
					// trigger reconcile when old / new CRD match
					// and status is updated
					UpdateFunc: func(e event.UpdateEvent) bool {
						oldCrd := e.ObjectOld.(*apiextensionsv1.CustomResourceDefinition)
						crd := e.ObjectNew.(*apiextensionsv1.CustomResourceDefinition)
						return (reconcilers.MatchingCRD(oldCrd.Name) ||
							reconcilers.MatchingCRD(crd.Name)) &&
							!reflect.DeepEqual(oldCrd.Status, crd.Status)
					},
					DeleteFunc: func(e event.DeleteEvent) bool {
						return false
					},
					GenericFunc: func(e event.GenericEvent) bool {
						return false
					},
				},
			),
		)

	if err := r.setupDiscovery(ctx, mgr, builder); err != nil {
		return err
	}
	return builder.Complete(r)
}

func (r *FlowCollectorReconciler) setupDiscovery(ctx context.Context, mgr ctrl.Manager, builder *builder.Builder) error {
	log := log.FromContext(ctx)
	dc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return fmt.Errorf("can't instantiate discovery client: %w", err)
	}
	r.permissions = discover.Permissions{
		Client: dc,
	}
	if r.permissions.Vendor(ctx) == discover.VendorOpenShift {
		builder.Owns(&securityv1.SecurityContextConstraints{})
	}
	apis, err := discover.NewAvailableAPIs(dc)
	if err != nil {
		return fmt.Errorf("can't discover available APIs: %w", err)
	}
	r.availableAPIs = apis
	if apis.HasConsole() {
		builder.Owns(&osv1alpha1.ConsolePlugin{})
	} else {
		log.Info("Console not detected: the console plugin is not available")
	}
	if !apis.HasCNO() {
		log.Info("CNO not detected: using ovnKubernetes config and reconciler")
	}
	return nil
}

func getNamespaceName(desired *flowsv1alpha1.FlowCollector) string {
	if desired.Spec.Namespace != "" {
		return desired.Spec.Namespace
	}
	return constants.DefaultOperatorNamespace
}

func (r *FlowCollectorReconciler) namespaceExist(ctx context.Context, nsName string) (bool, error) {
	err := r.Get(ctx, types.NamespacedName{Name: nsName}, &corev1.Namespace{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		log.FromContext(ctx).Error(err, "Failed to get namespace")
		return false, err
	}
	return true, nil
}

// checkFinalizer returns true (and/or error) if the calling function needs to return
func (r *FlowCollectorReconciler) checkFinalizer(ctx context.Context, desired *flowsv1alpha1.FlowCollector) (bool, error) {
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(desired, flowsFinalizer) {
			// Run finalization logic
			if err := r.finalize(ctx, desired); err != nil {
				return true, err
			}
			// Remove finalizer
			controllerutil.RemoveFinalizer(desired, flowsFinalizer)
			err := r.Update(ctx, desired)
			return true, err
		}
		return true, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(desired, flowsFinalizer) {
		controllerutil.AddFinalizer(desired, flowsFinalizer)
		if err := r.Update(ctx, desired); err != nil {
			return true, err
		}
	}

	return false, nil
}

func (r *FlowCollectorReconciler) finalize(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	if !r.availableAPIs.HasCNO() {
		ns := getNamespaceName(desired)
		clientHelper := r.newClientHelper(desired)
		ovsConfigController := ovs.NewFlowsConfigOVNKController(clientHelper,
			ns,
			desired.Spec.Agent.IPFIX.OVNKubernetes,
			r.lookupIP)
		if err := ovsConfigController.Finalize(ctx, desired); err != nil {
			return fmt.Errorf("failed to finalize ovn-kubernetes reconciler: %w", err)
		}
	}
	return nil
}

func (r *FlowCollectorReconciler) newClientHelper(desired *flowsv1alpha1.FlowCollector) reconcilers.ClientHelper {
	return reconcilers.ClientHelper{
		Client: r.Client,
		SetControllerReference: func(obj client.Object) error {
			return ctrl.SetControllerReference(desired, obj, r.Scheme)
		},
	}
}

func (r *FlowCollectorReconciler) failure(ctx context.Context, errcond *conditions.ErrorCondition, fc *flowsv1alpha1.FlowCollector) error {
	log := log.FromContext(ctx)
	log.Error(errcond.Error, errcond.Message)
	r.updateCondition(ctx, &errcond.Condition, fc)
	return errcond.Error
}

func (r *FlowCollectorReconciler) updateCondition(ctx context.Context, cond *metav1.Condition, fc *flowsv1alpha1.FlowCollector) {
	r.updateConditions(ctx, &[]metav1.Condition{*cond}, fc)
}

func (r *FlowCollectorReconciler) updateConditions(ctx context.Context, conds *[]metav1.Condition, fc *flowsv1alpha1.FlowCollector) {
	conditions.SetNewConditions(&fc.Status.Conditions, conds)

	if err := r.Status().Update(ctx, fc); err != nil {
		log.FromContext(ctx).Error(err, "Set conditions failed")
		// Do not propagate this update failure: if update failed it's generally because it was modified concurrently:
		// in that case, it will anyway trigger new reconcile loops so the conditions will be updated soon.
	}
}
