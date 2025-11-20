package networkpolicy

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/internal/controller/reconcilers"
	"github.com/netobserv/network-observability-operator/internal/pkg/helper"
	"github.com/netobserv/network-observability-operator/internal/pkg/manager"
	"github.com/netobserv/network-observability-operator/internal/pkg/manager/status"
)

type Reconciler struct {
	client.Client
	mgr    *manager.Manager
	status status.Instance
}

// enqueueFlowCollectorOnEndpointChange is a handler that triggers reconciliation when kubernetes service endpoints change
func enqueueFlowCollectorOnEndpointChange(ctx context.Context, obj client.Object) []reconcile.Request {
	// Only watch the kubernetes service in default namespace
	if obj.GetNamespace() == kubernetesServiceNamespace && obj.GetName() == kubernetesServiceName {
		log.FromContext(ctx).V(1).Info("Kubernetes service endpoint changed, triggering reconciliation")
		// Trigger reconciliation for all FlowCollectors
		return []reconcile.Request{{}}
	}
	return nil
}

func Start(ctx context.Context, mgr *manager.Manager) error {
	log := log.FromContext(ctx)
	log.Info("Starting Network Policy controller")
	r := Reconciler{
		Client: mgr.Client,
		mgr:    mgr,
		status: mgr.Status.ForComponent(status.NetworkPolicy),
	}

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&flowslatest.FlowCollector{}, reconcilers.IgnoreStatusChange).
		Named("networkPolicy").
		Owns(&networkingv1.NetworkPolicy{}, reconcilers.UpdateOrDeleteOnlyPred)

	// Watch EndpointSlice if available (preferred, k8s >= 1.21), otherwise fallback to Endpoints
	if isEndpointSliceAvailable(mgr) {
		log.V(1).Info("EndpointSlice API available, watching for kubernetes service changes")
		builder = builder.Watches(&discoveryv1.EndpointSlice{}, handler.EnqueueRequestsFromMapFunc(enqueueFlowCollectorOnEndpointChange))
	} else {
		log.Info("EndpointSlice API not available (requires k8s >= 1.21), using Endpoints API")
		//nolint:staticcheck // SA1019: Endpoints is deprecated but used as fallback for k8s < 1.21
		builder = builder.Watches(&corev1.Endpoints{}, handler.EnqueueRequestsFromMapFunc(enqueueFlowCollectorOnEndpointChange))
	}

	return builder.Complete(&r)
}

// isEndpointSliceAvailable checks if the EndpointSlice API (discovery.k8s.io/v1) is available
// in the cluster. This API was introduced in Kubernetes 1.21.
func isEndpointSliceAvailable(mgr *manager.Manager) bool {
	gvk := discoveryv1.SchemeGroupVersion.WithKind("EndpointSlice")
	restMapper := mgr.GetRESTMapper()

	_, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	return err == nil
}

// Reconcile is the controller entry point for reconciling current state with desired state.
// It manages the controller status at a high level. Business logic is delegated into `reconcile`.
func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	l := log.Log.WithName("networkpolicy") // clear context (too noisy)
	ctx = log.IntoContext(ctx, l)

	// Get flowcollector & create dedicated client
	clh, desired, err := helper.NewFlowCollectorClientHelper(ctx, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get FlowCollector: %w", err)
	} else if desired == nil {
		// Delete case
		return ctrl.Result{}, nil
	}

	r.status.SetUnknown()
	defer r.status.Commit(ctx, r.Client)

	err = r.reconcile(ctx, clh, desired)
	if err != nil {
		l.Error(err, "Network policy reconcile failure")
		// Set status failure unless it was already set
		if !r.status.HasFailure() {
			r.status.SetFailure("NetworkPolicyError", err.Error())
		}
		return ctrl.Result{}, err
	}

	r.status.SetReady()
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcile(ctx context.Context, clh *helper.Client, desired *flowslatest.FlowCollector) error {
	l := log.FromContext(ctx)

	cni, err := r.mgr.ClusterInfo.GetCNI()
	if err != nil {
		return err
	}

	// Get API server endpoint IPs for network policy
	var apiServerIPs []string
	if r.mgr.ClusterInfo.IsOpenShift() {
		apiServerIPs, err = GetAPIServerEndpointIPs(ctx, r.Client)
		if err != nil {
			l.Error(err, "Failed to get API server endpoint IPs, network policy will allow all IPs on port 6443")
			// Continue without IPs - will fallback to allowing all IPs on port 6443
			apiServerIPs = nil
		}
	}

	npName, desiredNp := buildMainNetworkPolicy(desired, r.mgr, cni, apiServerIPs)
	if err := reconcilers.ReconcileNetworkPolicy(ctx, clh, npName, desiredNp); err != nil {
		return err
	}

	privilegedNpName, desiredPrivilegedNp := buildPrivilegedNetworkPolicy(desired, r.mgr, cni)
	return reconcilers.ReconcileNetworkPolicy(ctx, clh, privilegedNpName, desiredPrivilegedNp)
}
