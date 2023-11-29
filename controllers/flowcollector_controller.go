package controllers

import (
	"context"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	securityv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/consoleplugin"
	"github.com/netobserv/network-observability-operator/controllers/ebpf"
	"github.com/netobserv/network-observability-operator/controllers/flowlogspipeline"
	"github.com/netobserv/network-observability-operator/controllers/ovs"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/cleanup"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
)

const (
	ovsFlowsConfigMapName = "ovs-flows-config"
	flowsFinalizer        = "flows.netobserv.io/finalizer"
)

// FlowCollectorReconciler reconciles a FlowCollector object
type FlowCollectorReconciler struct {
	client.Client
	mgr       *manager.Manager
	status    status.Instance
	watcher   *watchers.Watcher
	clusterID string
}

func Start(ctx context.Context, mgr *manager.Manager) error {
	log := log.FromContext(ctx)
	log.Info("Starting FlowCollector controller")
	r := FlowCollectorReconciler{
		Client: mgr.Client,
		mgr:    mgr,
		status: mgr.Status.ForComponent(status.FlowCollectorLegacy),
	}

	builder := ctrl.NewControllerManagedBy(mgr.Manager).
		For(&flowslatest.FlowCollector{}, reconcilers.IgnoreStatusChange).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&ascv2.HorizontalPodAutoscaler{}).
		Owns(&corev1.Namespace{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{})

	if mgr.IsOpenShift() {
		builder.Owns(&securityv1.SecurityContextConstraints{})
	}
	if mgr.HasConsolePlugin() {
		builder.Owns(&osv1alpha1.ConsolePlugin{})
	} else {
		log.Info("Console not detected: the console plugin is not available")
	}
	if !mgr.HasCNO() {
		log.Info("CNO not detected: using ovnKubernetes config and reconciler")
	}

	ctrl, err := builder.Build(&r)
	if err != nil {
		return err
	}
	r.watcher = watchers.NewWatcher(ctrl)

	return nil
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
//+kubebuilder:rbac:groups=loki.grafana.com,resources=network,resourceNames=logs,verbs=get;create
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
	// At the moment, status workflow is to start as ready then degrade if necessary
	// Later (when legacy controller is broken down into individual controllers), status should start as unknown and only on success finishes as ready
	r.status.SetReady()
	defer r.status.Commit(ctx, r.Client)

	err := r.reconcile(ctx)
	if err != nil {
		log.Error(err, "FlowCollector reconcile failure")
		// Set status failure unless it was already set
		if !r.status.HasFailure() {
			r.status.SetFailure("FlowCollectorGenericError", err.Error())
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *FlowCollectorReconciler) reconcile(ctx context.Context) error {
	clh, desired, err := helper.NewFlowCollectorClientHelper(ctx, r.Client)
	if err != nil {
		return fmt.Errorf("failed to get FlowCollector: %w", err)
	} else if desired == nil {
		return nil
	}

	ns := helper.GetNamespace(&desired.Spec)
	previousNamespace := desired.Status.Namespace
	loki := helper.NewLokiConfig(&desired.Spec.Loki, ns)
	reconcilersInfo := r.newCommonInfo(clh, ns, previousNamespace, &loki)

	if ret, err := r.checkFinalizer(ctx, desired, &reconcilersInfo); ret {
		return err
	}

	if err := cleanup.CleanPastReferences(ctx, r.Client, ns); err != nil {
		return err
	}
	r.watcher.Reset(ns)

	// obtain default cluster ID - api is specific to openshift
	if r.mgr.IsOpenShift() && r.clusterID == "" {
		cversion := &configv1.ClusterVersion{}
		key := client.ObjectKey{Name: "version"}
		if err := r.Client.Get(ctx, key, cversion); err != nil {
			log.FromContext(ctx).Error(err, "unable to obtain cluster ID")
		} else {
			r.clusterID = string(cversion.Spec.ClusterID)
		}
	}

	// Create reconcilers
	flpReconciler := flowlogspipeline.NewReconciler(&reconcilersInfo, r.mgr.Config.FlowlogsPipelineImage)
	var cpReconciler consoleplugin.CPReconciler
	if r.mgr.HasConsolePlugin() {
		cpReconciler = consoleplugin.NewReconciler(&reconcilersInfo, r.mgr.Config.ConsolePluginImage)
	}

	// Check namespace changed
	if ns != previousNamespace {
		if err := r.handleNamespaceChanged(ctx, previousNamespace, ns, desired, &flpReconciler, &cpReconciler); err != nil {
			return r.status.Error("CannotCreateNamespace", err)
		}
	}

	// Flowlogs-pipeline
	if err := flpReconciler.Reconcile(ctx, desired); err != nil {
		return r.status.Error("ReconcileFLPFailed", err)
	}

	// OVS config map for CNO
	if r.mgr.HasCNO() {
		ovsConfigController := ovs.NewFlowsConfigCNOController(&reconcilersInfo, desired.Spec.Agent.IPFIX.ClusterNetworkOperator.Namespace, ovsFlowsConfigMapName)
		if err := ovsConfigController.Reconcile(ctx, desired); err != nil {
			return r.status.Error("ReconcileCNOFailed", err)
		}
	} else {
		ovsConfigController := ovs.NewFlowsConfigOVNKController(&reconcilersInfo, desired.Spec.Agent.IPFIX.OVNKubernetes)
		if err := ovsConfigController.Reconcile(ctx, desired); err != nil {
			return r.status.Error("ReconcileOVNKFailed", err)
		}
	}

	// eBPF agent
	ebpfAgentController := ebpf.NewAgentController(&reconcilersInfo, r.mgr.Config.EBPFAgentImage)
	if err := ebpfAgentController.Reconcile(ctx, desired); err != nil {
		return r.status.Error("ReconcileAgentFailed", err)
	}

	// Console plugin
	if r.mgr.HasConsolePlugin() {
		err := cpReconciler.Reconcile(ctx, desired)
		if err != nil {
			return r.status.Error("ReconcileConsolePluginFailed", err)
		}
	}

	return nil
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
		if r.mgr.HasConsolePlugin() {
			cpReconciler.CleanupNamespace(ctx)
		}
	}

	// Update namespace in status
	log.Info("Updating status with new namespace " + newNS)
	desired.Status.Namespace = newNS
	return r.Status().Update(ctx, desired)
}

// checkFinalizer returns true (and/or error) if the calling function needs to return
func (r *FlowCollectorReconciler) checkFinalizer(ctx context.Context, desired *flowslatest.FlowCollector, info *reconcilers.Common) (bool, error) {
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(desired, flowsFinalizer) {
			// Run finalization logic
			if err := r.finalize(ctx, desired, info); err != nil {
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

func (r *FlowCollectorReconciler) finalize(ctx context.Context, desired *flowslatest.FlowCollector, info *reconcilers.Common) error {
	if !r.mgr.HasCNO() {
		ovsConfigController := ovs.NewFlowsConfigOVNKController(info, desired.Spec.Agent.IPFIX.OVNKubernetes)
		if err := ovsConfigController.Finalize(ctx, desired); err != nil {
			return fmt.Errorf("failed to finalize ovn-kubernetes reconciler: %w", err)
		}
	}
	return nil
}

func (r *FlowCollectorReconciler) newCommonInfo(clh *helper.Client, ns, prevNs string, loki *helper.LokiConfig) reconcilers.Common {
	return reconcilers.Common{
		Client:            *clh,
		Status:            r.status,
		Namespace:         ns,
		PreviousNamespace: prevNs,
		UseOpenShiftSCC:   r.mgr.IsOpenShift(),
		AvailableAPIs:     &r.mgr.AvailableAPIs,
		Watcher:           r.watcher,
		Loki:              loki,
		ClusterID:         r.clusterID,
	}
}
