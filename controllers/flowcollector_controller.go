package controllers

import (
	"context"
	"fmt"
	"net"

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

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/consoleplugin"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/ebpf"
	"github.com/netobserv/network-observability-operator/controllers/flowlogspipeline"
	"github.com/netobserv/network-observability-operator/controllers/operator"
	"github.com/netobserv/network-observability-operator/controllers/ovs"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/cluster"
	"github.com/netobserv/network-observability-operator/pkg/conditions"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
)

const (
	ovsFlowsConfigMapName = "ovs-flows-config"
	flowsFinalizer        = "flows.netobserv.io/finalizer"
)

// FlowCollectorReconciler reconciles a FlowCollector object
type FlowCollectorReconciler struct {
	client.Client
	clusterInfo *cluster.Info
	Scheme      *runtime.Scheme
	config      *operator.Config
	watcher     *watchers.Watcher
	lookupIP    func(string) ([]net.IP, error)
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
//+kubebuilder:rbac:groups=core,resources=namespaces;services;serviceaccounts;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings;clusterroles;rolebindings;roles,verbs=get;list;create;delete;update;watch
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=get;create;delete;update;patch;list;watch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=consoles,verbs=get;update;list;update;watch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=hostnetwork,verbs=use
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=list;create;update;watch
//+kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,verbs=list;get;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors;prometheusrules,verbs=get;create;delete;update;patch;list;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch
//+kubebuilder:rbac:urls="/metrics",verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the FlowCollector object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *FlowCollectorReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	desired, err := r.getFlowCollector(ctx)
	if err != nil {
		log.Error(err, "Failed to get FlowCollector")
		return ctrl.Result{}, err
	} else if desired == nil {
		return ctrl.Result{}, nil
	}

	ns := getNamespaceName(desired)
	r.watcher.Reset(ns)
	if err := r.clusterInfo.CheckClusterInfo(ctx, r.Client); err != nil {
		return ctrl.Result{}, err
	}

	var didChange, isInProgress bool
	previousNamespace := desired.Status.Namespace
	reconcilersInfo := r.newCommonInfo(desired, ns, previousNamespace, func(b bool) { didChange = b }, func(b bool) { isInProgress = b })

	err = r.reconcileOperator(ctx, &reconcilersInfo, desired)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create reconcilers
	flpReconciler := flowlogspipeline.NewReconciler(&reconcilersInfo, r.config.FlowlogsPipelineImage)
	var cpReconciler consoleplugin.CPReconciler
	if r.clusterInfo.HasConsolePlugin() {
		cpReconciler = consoleplugin.NewReconciler(&reconcilersInfo, r.config.ConsolePluginImage)
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
	if r.clusterInfo.HasCNO() {
		ovsConfigController := ovs.NewFlowsConfigCNOController(&reconcilersInfo, desired.Spec.Agent.IPFIX.ClusterNetworkOperator.Namespace, ovsFlowsConfigMapName)
		if err := ovsConfigController.Reconcile(ctx, desired); err != nil {
			return ctrl.Result{}, r.failure(ctx, conditions.ReconcileCNOFailed(err), desired)
		}
	} else {
		ovsConfigController := ovs.NewFlowsConfigOVNKController(&reconcilersInfo, desired.Spec.Agent.IPFIX.OVNKubernetes)
		if err := ovsConfigController.Reconcile(ctx, desired); err != nil {
			return ctrl.Result{}, r.failure(ctx, conditions.ReconcileOVNKFailed(err), desired)
		}
	}

	// eBPF agent
	ebpfAgentController := ebpf.NewAgentController(&reconcilersInfo, r.config)
	if err := ebpfAgentController.Reconcile(ctx, desired); err != nil {
		return ctrl.Result{}, r.failure(ctx, conditions.ReconcileAgentFailed(err), desired)
	}

	// Console plugin
	if r.clusterInfo.HasConsolePlugin() {
		err := cpReconciler.Reconcile(ctx, desired)
		if err != nil {
			return ctrl.Result{}, r.failure(ctx, conditions.ReconcileConsolePluginFailed(err), desired)
		}
	}

	// Set readiness status
	var status *metav1.Condition
	if didChange {
		status = conditions.Updating()
	} else if isInProgress {
		status = conditions.DeploymentInProgress()
	} else {
		status = conditions.Ready()
	}
	r.updateCondition(ctx, status, desired)
	return ctrl.Result{}, nil
}

func (r *FlowCollectorReconciler) getFlowCollector(ctx context.Context) (*flowslatest.FlowCollector, error) {
	log := log.FromContext(ctx)
	desired := &flowslatest.FlowCollector{}
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
	desired *flowslatest.FlowCollector,
	flpReconciler *flowlogspipeline.FLPReconciler,
	cpReconciler *consoleplugin.CPReconciler,
) error {
	log := log.FromContext(ctx)
	if oldNS != "" {
		// Namespace updated, clean up previous namespace
		log.Info("FlowCollector namespace change detected: cleaning up previous namespace", "old namespace", oldNS, "new namepace", newNS)
		flpReconciler.CleanupNamespace(ctx)
		if r.clusterInfo.HasConsolePlugin() {
			cpReconciler.CleanupNamespace(ctx)
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
		For(&flowslatest.FlowCollector{}).
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

	r.watcher = watchers.RegisterWatcher(builder)

	return builder.Complete(r)
}

func (r *FlowCollectorReconciler) setupDiscovery(ctx context.Context, mgr ctrl.Manager, builder *builder.Builder) error {
	log := log.FromContext(ctx)
	dc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return fmt.Errorf("can't instantiate discovery client: %w", err)
	}
	info, err := cluster.NewInfo(dc)
	if err != nil {
		return fmt.Errorf("can't collect cluster info: %w", err)
	}
	if info.HasOCPSecurity() {
		builder.Owns(&securityv1.SecurityContextConstraints{})
	}
	r.clusterInfo = &info
	if info.HasConsolePlugin() {
		builder.Owns(&osv1alpha1.ConsolePlugin{})
	} else {
		log.Info("Console not detected: the console plugin is not available")
	}
	if !info.HasCNO() {
		log.Info("CNO not detected: if using IPFIX agent, will use ovnKubernetes config and reconciler")
	}
	return nil
}

func getNamespaceName(desired *flowslatest.FlowCollector) string {
	if desired.Spec.Namespace != "" {
		return desired.Spec.Namespace
	}
	return constants.DefaultOperatorNamespace
}

func (r *FlowCollectorReconciler) namespaceExist(ctx context.Context, nsName string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: nsName}, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		log.FromContext(ctx).Error(err, "Failed to get namespace")
		return nil, err
	}
	return ns, nil
}

func (r *FlowCollectorReconciler) reconcileOperator(ctx context.Context, cmn *reconcilers.Common, desired *flowslatest.FlowCollector) error {
	// If namespace does not exist, we create it
	nsExist, err := r.namespaceExist(ctx, cmn.Namespace)
	if err != nil {
		return err
	}
	desiredNs := buildNamespace(cmn.Namespace, r.config.DownstreamDeployment)
	if nsExist == nil {
		err = r.Create(ctx, desiredNs)
		if err != nil {
			return r.failure(ctx, conditions.CannotCreateNamespace(err), desired)
		}
	} else if !helper.IsSubSet(nsExist.ObjectMeta.Labels, desiredNs.ObjectMeta.Labels) {
		err = r.Update(ctx, desiredNs)
		if err != nil {
			return err
		}
	}
	if r.config.DownstreamDeployment {
		desiredRole := buildRoleMonitoringReader()
		if err := cmn.ReconcileClusterRole(ctx, desiredRole); err != nil {
			return err
		}
		desiredBinding := buildRoleBindingMonitoringReader(cmn.Namespace)
		if err := cmn.ReconcileClusterRoleBinding(ctx, desiredBinding); err != nil {
			return err
		}
	}

	if r.clusterInfo.HasSvcMonitor() {
		desiredFlowDashboardCM, del, err := buildFlowMetricsDashboard(cmn.Namespace, desired.Spec.Processor.Metrics.IgnoreTags)
		if err != nil {
			return err
		} else if err = cmn.ReconcileConfigMap(ctx, desiredFlowDashboardCM, del); err != nil {
			return err
		}

		desiredHealthDashboardCM, del, err := buildHealthDashboard(desired.Spec.Processor.Metrics.IgnoreTags)
		if err != nil {
			return err
		} else if err = cmn.ReconcileConfigMap(ctx, desiredHealthDashboardCM, del); err != nil {
			return err
		}
	}
	return nil
}

// checkFinalizer returns true (and/or error) if the calling function needs to return
func (r *FlowCollectorReconciler) checkFinalizer(ctx context.Context, desired *flowslatest.FlowCollector) (bool, error) {
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

func (r *FlowCollectorReconciler) finalize(ctx context.Context, desired *flowslatest.FlowCollector) error {
	if !r.clusterInfo.HasCNO() {
		ns := getNamespaceName(desired)
		info := r.newCommonInfo(desired, ns, ns, func(b bool) {}, func(b bool) {})
		ovsConfigController := ovs.NewFlowsConfigOVNKController(&info, desired.Spec.Agent.IPFIX.OVNKubernetes)
		if err := ovsConfigController.Finalize(ctx, desired); err != nil {
			return fmt.Errorf("failed to finalize ovn-kubernetes reconciler: %w", err)
		}
	}
	return nil
}

func (r *FlowCollectorReconciler) newCommonInfo(desired *flowslatest.FlowCollector, ns, prevNs string, changeHook, inProgressHook func(bool)) reconcilers.Common {
	return reconcilers.Common{
		Client: helper.Client{
			Client: r.Client,
			SetControllerReference: func(obj client.Object) error {
				return ctrl.SetControllerReference(desired, obj, r.Scheme)
			},
			SetChanged:    changeHook,
			SetInProgress: inProgressHook,
		},
		Namespace:         ns,
		PreviousNamespace: prevNs,
		ClusterInfo:       r.clusterInfo,
		Watcher:           r.watcher,
	}
}

func (r *FlowCollectorReconciler) failure(ctx context.Context, errcond *conditions.ErrorCondition, fc *flowslatest.FlowCollector) error {
	log := log.FromContext(ctx)
	log.Error(errcond.Error, errcond.Message)
	meta.SetStatusCondition(&fc.Status.Conditions, errcond.Condition)
	if errUpdate := r.Status().Update(ctx, fc); errUpdate != nil {
		log.Error(errUpdate, "Set conditions failed")
	}
	return errcond.Error
}

func (r *FlowCollectorReconciler) updateCondition(ctx context.Context, cond *metav1.Condition, fc *flowslatest.FlowCollector) {
	meta.SetStatusCondition(&fc.Status.Conditions, *cond)
	if err := r.Status().Update(ctx, fc); err != nil {
		log.FromContext(ctx).Error(err, "Set conditions failed")
		// Do not propagate this update failure: if update failed it's generally because it was modified concurrently:
		// in that case, it will anyway trigger new reconcile loops so the conditions will be updated soon.
	}
}
