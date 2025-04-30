package static

import (
	"context"
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/consoleplugin"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
)

const (
	initReconcileAttempts = 5
)

type Reconciler struct {
	client.Client
	mgr    *manager.Manager
	status status.Instance
}

func Start(ctx context.Context, mgr *manager.Manager) error {
	log := log.FromContext(ctx)
	log.Info("Starting Static controller")
	r := Reconciler{
		Client: mgr.Client,
		mgr:    mgr,
		status: mgr.Status.ForComponent(status.StaticPlugin),
	}

	// force reconcile at startup
	go r.InitReconcile(ctx)

	return ctrl.NewControllerManagedBy(mgr).
		For(&flowslatest.FlowCollector{}, reconcilers.IgnoreStatusChange).
		Named("staticPlugin").
		Complete(&r)
}

func (r *Reconciler) InitReconcile(ctx context.Context) {
	log := log.FromContext(ctx)
	log.Info("Initializing resources...")

	for attempt := range initReconcileAttempts {
		// delay the reconcile calls to let some time to the cache to load
		time.Sleep(5 * time.Second)
		_, err := r.Reconcile(ctx, ctrl.Request{})
		if err != nil {
			log.Error(err, "Error while doing initial reconcile", "attempt", attempt)
		} else {
			return
		}
	}
}

// Reconcile is the controller entry point for reconciling current state with desired state.
// It manages the controller status at a high level. Business logic is delegated into `reconcile`.
func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	l := log.Log.WithName("staticPlugin") // clear context (too noisy)
	ctx = log.IntoContext(ctx, l)

	r.status.SetUnknown()
	defer r.status.Commit(ctx, r.Client)

	// always reconcile static console plugin
	scp, err := helper.NewControllerClientHelper(r.mgr.Config.Namespace, ctx, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get controller deployment: %w", err)
	}
	staticPluginReconciler := consoleplugin.NewStaticReconciler(r.newDefaultReconcilerInstance(scp))
	if err := staticPluginReconciler.ReconcileStaticPlugin(ctx, true); err != nil {
		l.Error(err, "Static plugin reconcile failure")
		// Set status failure unless it was already set
		if !r.status.HasFailure() {
			r.status.SetFailure("StaticPluginError", err.Error())
		}
		return ctrl.Result{}, err
	}

	r.status.SetReady()
	return ctrl.Result{}, nil
}

func (r *Reconciler) newDefaultReconcilerInstance(clh *helper.Client) *reconcilers.Instance {
	// force default namespace
	reconcilersInfo := reconcilers.Common{
		Client:       *clh,
		Namespace:    r.mgr.Config.Namespace,
		ClusterInfo:  r.mgr.ClusterInfo,
		Watcher:      nil,
		Loki:         &helper.LokiConfig{},
		IsDownstream: r.mgr.Config.DownstreamDeployment,
	}
	return reconcilersInfo.NewInstance([]string{r.mgr.Config.ConsolePluginImage}, r.status)
}
