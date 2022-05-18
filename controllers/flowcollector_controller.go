package controllers

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/netobserv/network-observability-operator/controllers/ebpf"
	"github.com/netobserv/network-observability-operator/controllers/ovs"
	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	securityv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/consoleplugin"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/flowlogspipeline"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/discover"
)

const ovsFlowsConfigMapName = "ovs-flows-config"

// FlowCollectorReconciler reconciles a FlowCollector object
type FlowCollectorReconciler struct {
	client.Client
	permissions      discover.Permissions
	Scheme           *runtime.Scheme
	consoleAvailable bool
	lookupIP         func(string) ([]net.IP, error)
}

func NewFlowCollectorReconciler(client client.Client, scheme *runtime.Scheme) *FlowCollectorReconciler {
	return &FlowCollectorReconciler{
		Client:           client,
		Scheme:           scheme,
		consoleAvailable: false,
		lookupIP:         net.LookupIP,
	}
}

//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces;services;serviceaccounts;configmaps,verbs=get;list;watch;create;update;patch;delete
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

	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("No need to reconcile status of a FlowCollector that is being deleted. Ignoring")
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
			log.Error(err, "Failed to create Namespace")
			return ctrl.Result{}, err
		}
	}

	clientHelper := reconcilers.ClientHelper{
		Client: r.Client,
		SetControllerReference: func(obj client.Object) error {
			return ctrl.SetControllerReference(desired, obj, r.Scheme)
		},
	}
	previousNamespace := desired.Status.Namespace

	// Create reconcilers
	flpReconciler := flowlogspipeline.NewReconciler(clientHelper, ns, previousNamespace)
	var cpReconciler consoleplugin.CPReconciler
	if r.consoleAvailable {
		cpReconciler = consoleplugin.NewReconciler(clientHelper, ns, previousNamespace)
	}

	// Check namespace changed
	if ns != previousNamespace {
		if err := r.handleNamespaceChanged(ctx, previousNamespace, ns, desired, &flpReconciler, &cpReconciler); err != nil {
			log.Error(err, "Failed to handle namespace change")
			return ctrl.Result{}, err
		}
	}

	// Flowlogs-pipeline
	if err := flpReconciler.Reconcile(ctx, desired); err != nil {
		log.Error(err, "Failed to reconcile flowlogs-pipeline")
		return ctrl.Result{}, err
	}

	// OVS config map for CNO
	ovsConfigController := ovs.NewFlowsConfigController(clientHelper,
		ns,
		desired.Spec.ClusterNetworkOperator.Namespace,
		ovsFlowsConfigMapName,
		r.lookupIP)
	if err := ovsConfigController.Reconcile(ctx, desired, flpReconciler.GetServiceName(desired.Spec.Kafka)); err != nil {
		return ctrl.Result{},
			fmt.Errorf("failed to reconcile ovs-flows-config ConfigMap: %w", err)
	}

	// eBPF agent
	ebpfAgentController := ebpf.NewAgentController(clientHelper, ns, &r.permissions)
	if err := ebpfAgentController.Reconcile(ctx, desired); err != nil {
		return ctrl.Result{},
			fmt.Errorf("failed to reconcile eBPF Netobserv Agent: %w", err)
	}

	// Console plugin
	if r.consoleAvailable {
		err := cpReconciler.Reconcile(ctx, desired)
		if err != nil {
			log.Error(err, "Failed to reconcile console plugin")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
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
		if r.consoleAvailable {
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
		if r.consoleAvailable {
			err := cpReconciler.PrepareNamespaceChange(ctx)
			if err != nil {
				return err
			}
		}
	}

	// Update namespace in status
	log.Info("Updating status with new namespace " + newNS)
	desired.Status.Namespace = newNS
	return r.Status().Update(ctx, desired)
}

func isConsoleAvailable(mgr ctrl.Manager) (bool, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return false, err
	}
	groupsList, err := discoveryClient.ServerGroups()
	if err != nil {
		return false, err
	}
	for i := range groupsList.Groups {
		if strings.HasSuffix(groupsList.Groups[i].Name, osv1alpha1.GroupName) {
			return true, nil
		}
	}
	return false, nil
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

	if err := r.setupOpenshift(ctx, mgr, builder); err != nil {
		return err
	}

	var err error
	r.consoleAvailable, err = isConsoleAvailable(mgr)
	if err != nil {
		return err
	}
	if r.consoleAvailable {
		builder = builder.Owns(&osv1alpha1.ConsolePlugin{})
	}
	return builder.Complete(r)
}

func (r *FlowCollectorReconciler) setupOpenshift(ctx context.Context, mgr ctrl.Manager, builder *builder.Builder) error {
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
