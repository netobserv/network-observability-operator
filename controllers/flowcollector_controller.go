package controllers

import (
	"context"
	"fmt"
	"net"

	"github.com/netobserv/network-observability-operator/controllers/operator"
	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	securityv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/consoleplugin"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/ebpf"
	"github.com/netobserv/network-observability-operator/controllers/flowlogspipeline"
	"github.com/netobserv/network-observability-operator/controllers/ovs"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/conditions"
	"github.com/netobserv/network-observability-operator/pkg/discover"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
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
	config        *operator.Config
	certWatcher   *watchers.CertificatesWatcher
	lookupIP      func(string) ([]net.IP, error)
}

func NewFlowCollectorReconciler(client client.Client, scheme *runtime.Scheme, config *operator.Config) *FlowCollectorReconciler {
	return &FlowCollectorReconciler{
		Client:   client,
		Scheme:   scheme,
		lookupIP: net.LookupIP,
		config:   config,
	}
}

//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces;services;serviceaccounts;configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;create;delete;watch;list
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;create;delete;update;watch
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=get;create;delete;update;patch;list;watch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=consoles,verbs=get;update;list;update;watch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=hostnetwork,verbs=use
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=list;create;update;watch
//+kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,verbs=list;get;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors;prometheusrules,verbs=get;create;delete;update;patch;list;watch

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
	desired, err := r.getFlowCollector(ctx)
	if err != nil {
		log.Error(err, "Failed to get FlowCollector")
		return ctrl.Result{}, err
	} else if desired == nil {
		return ctrl.Result{}, nil
	}

	ns := getNamespaceName(desired)
	r.certWatcher.Reset(ns)
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
	previousNamespace := desired.Status.Namespace

	// Create reconcilers
	flpReconciler := flowlogspipeline.NewReconciler(ctx, clientHelper, ns, previousNamespace, r.config.FlowlogsPipelineImage, &r.permissions, r.availableAPIs)
	var cpReconciler consoleplugin.CPReconciler
	if r.availableAPIs.HasConsole() {
		cpReconciler = consoleplugin.NewReconciler(clientHelper, ns, previousNamespace, r.config.ConsolePluginImage, r.availableAPIs)
	}

	// Check namespace changed
	if ns != previousNamespace {
		if err := r.handleNamespaceChanged(ctx, previousNamespace, ns, desired, &flpReconciler, &cpReconciler); err != nil {
			return ctrl.Result{}, r.failure(ctx, conditions.CannotCreateNamespace(err), desired)
		}
	}

	// Flowlogs-pipeline
	if err := flpReconciler.Reconcile(ctx, desired); err != nil {
		return ctrl.Result{}, r.failure(ctx, conditions.ReconcileFLPFailed(err), desired)
	}

	// OVS config map for CNO
	if r.availableAPIs.HasCNO() {
		ovsConfigController := ovs.NewFlowsConfigCNOController(clientHelper,
			ns,
			desired.Spec.Agent.IPFIX.ClusterNetworkOperator.Namespace,
			ovsFlowsConfigMapName,
			r.lookupIP)
		if err := ovsConfigController.Reconcile(ctx, desired); err != nil {
			return ctrl.Result{}, r.failure(ctx, conditions.ReconcileCNOFailed(err), desired)
		}
	} else {
		ovsConfigController := ovs.NewFlowsConfigOVNKController(clientHelper,
			ns,
			desired.Spec.Agent.IPFIX.OVNKubernetes,
			r.lookupIP)
		if err := ovsConfigController.Reconcile(ctx, desired); err != nil {
			return ctrl.Result{}, r.failure(ctx, conditions.ReconcileOVNKFailed(err), desired)
		}
	}

	// eBPF agent
	ebpfAgentController := ebpf.NewAgentController(clientHelper, ns, previousNamespace, &r.permissions, r.config)
	if err := ebpfAgentController.Reconcile(ctx, desired); err != nil {
		return ctrl.Result{}, r.failure(ctx, conditions.ReconcileAgentFailed(err), desired)
	}

	// Console plugin
	if r.availableAPIs.HasConsole() {
		err := cpReconciler.Reconcile(ctx, desired)
		if err != nil {
			return ctrl.Result{}, r.failure(ctx, conditions.ReconcileConsolePluginFailed(err), desired)
		}
	}

	// Set readiness status
	var status *metav1.Condition
	if clientHelper.DidChange() {
		status = conditions.Updating()
	} else if clientHelper.IsInProgress() {
		status = conditions.DeploymentInProgress()
	} else {
		status = conditions.Ready()
	}
	r.updateCondition(ctx, status, desired)
	return ctrl.Result{}, nil
}

func (r *FlowCollectorReconciler) getFlowCollector(ctx context.Context) (*flowsv1alpha1.FlowCollector, error) {
	log := log.FromContext(ctx)
	desired := &flowsv1alpha1.FlowCollector{}
	if err := r.Get(ctx, constants.FlowCollectorName, desired); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("FlowCollector resource not found. Ignoring since object must be deleted")
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
	r.updateCondition(ctx, conditions.Updating(), desired)
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
		Owns(&corev1.ServiceAccount{})

	if err := r.setupDiscovery(ctx, mgr, builder); err != nil {
		return err
	}

	r.certWatcher = watchers.RegisterCertificatesWatcher(builder)

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
		CertWatcher: r.certWatcher,
		Client:      r.Client,
		SetControllerReference: func(obj client.Object) error {
			return ctrl.SetControllerReference(desired, obj, r.Scheme)
		},
	}
}

func (r *FlowCollectorReconciler) failure(ctx context.Context, errcond *conditions.ErrorCondition, fc *flowsv1alpha1.FlowCollector) error {
	log := log.FromContext(ctx)
	log.Error(errcond.Error, errcond.Message)
	meta.SetStatusCondition(&fc.Status.Conditions, errcond.Condition)
	if errUpdate := r.Status().Update(ctx, fc); errUpdate != nil {
		log.Error(errUpdate, "Set conditions failed")
	}
	return errcond.Error
}

func (r *FlowCollectorReconciler) updateCondition(ctx context.Context, cond *metav1.Condition, fc *flowsv1alpha1.FlowCollector) {
	meta.SetStatusCondition(&fc.Status.Conditions, *cond)
	if err := r.Status().Update(ctx, fc); err != nil {
		log.FromContext(ctx).Error(err, "Set conditions failed")
		// Do not propagate this update failure: if update failed it's generally because it was modified concurrently:
		// in that case, it will anyway trigger new reconcile loops so the conditions will be updated soon.
	}
}
