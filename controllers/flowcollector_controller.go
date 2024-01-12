package controllers

import (
	"context"
	"fmt"

	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	securityv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/consoleplugin"
	"github.com/netobserv/network-observability-operator/controllers/ebpf"
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
	mgr     *manager.Manager
	status  status.Instance
	watcher *watchers.Watcher
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
		Named("legacy").
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
	l := log.Log.WithName("legacy") // clear context (too noisy)
	ctx = log.IntoContext(ctx, l)
	// At the moment, status workflow is to start as ready then degrade if necessary
	// Later (when legacy controller is broken down into individual controllers), status should start as unknown and only on success finishes as ready
	r.status.SetReady()
	defer r.status.Commit(ctx, r.Client)

	err := r.reconcile(ctx)
	if err != nil {
		l.Error(err, "FlowCollector reconcile failure")
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
	previousNamespace := r.status.GetDeployedNamespace(desired)
	loki := helper.NewLokiConfig(&desired.Spec.Loki, ns)
	reconcilersInfo := r.newCommonInfo(clh, ns, previousNamespace, &loki)

	if ret, err := r.checkFinalizer(ctx, desired, &reconcilersInfo); ret {
		return err
	}

	if err := cleanup.CleanPastReferences(ctx, r.Client, ns); err != nil {
		return err
	}
	r.watcher.Reset(ns)

	// Create reconcilers
	var cpReconciler consoleplugin.CPReconciler
	if r.mgr.HasConsolePlugin() {
		cpReconciler = consoleplugin.NewReconciler(reconcilersInfo.NewInstance(r.mgr.Config.ConsolePluginImage, r.status))
	}

	// Check namespace changed
	if ns != previousNamespace {
		if previousNamespace != "" && r.mgr.HasConsolePlugin() {
			// Namespace updated, clean up previous namespace
			log.FromContext(ctx).
				Info("FlowCollector namespace change detected: cleaning up previous namespace", "old", previousNamespace, "new", ns)
			cpReconciler.CleanupNamespace(ctx)
		}

		// Update namespace in status
		if err := r.status.SetDeployedNamespace(ctx, r.Client, ns); err != nil {
			return r.status.Error("ChangeNamespaceError", err)
		}
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
	ebpfAgentController := ebpf.NewAgentController(reconcilersInfo.NewInstance(r.mgr.Config.EBPFAgentImage, r.status))
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
		Namespace:         ns,
		PreviousNamespace: prevNs,
		UseOpenShiftSCC:   r.mgr.IsOpenShift(),
		AvailableAPIs:     &r.mgr.AvailableAPIs,
		Watcher:           r.watcher,
		Loki:              loki,
	}
}
