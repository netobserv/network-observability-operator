package static

import (
	"context"
	"fmt"
	"time"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	flowslatest "github.com/netobserv/netobserv-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/netobserv-operator/internal/controller/consoleplugin"
	"github.com/netobserv/netobserv-operator/internal/controller/constants"
	"github.com/netobserv/netobserv-operator/internal/controller/reconcilers"
	"github.com/netobserv/netobserv-operator/internal/pkg/helper"
	"github.com/netobserv/netobserv-operator/internal/pkg/manager"
	"github.com/netobserv/netobserv-operator/internal/pkg/manager/status"
	"github.com/netobserv/netobserv-operator/internal/pkg/retry"
)

var (
	retryBackoff = wait.Backoff{
		Steps:    6,
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
	}
	clog = log.Log.WithName("static-controller")
)

type Reconciler struct {
	client.Client
	mgr    *manager.Manager
	status status.Instance
}

func Start(ctx context.Context, mgr *manager.Manager) (manager.PostCreateHook, error) {
	log := log.FromContext(ctx)
	log.Info("Starting Static controller")
	r := Reconciler{
		Client: mgr.Client,
		mgr:    mgr,
		status: mgr.Status.ForComponent(status.StaticController),
	}

	b := ctrl.NewControllerManagedBy(mgr).
		For(&flowslatest.FlowCollector{}, reconcilers.IgnoreStatusChange).
		Named("staticPlugin")
	if mgr.Config.StaticPluginConfig.InheritTolerationFromSubscription != "" {
		b = b.Watches(
			&olm.Subscription{},
			handler.EnqueueRequestsFromMapFunc(func(_ context.Context, o client.Object) []reconcile.Request {
				if o.GetNamespace() == mgr.Config.Namespace && o.GetName() == mgr.Config.StaticPluginConfig.InheritTolerationFromSubscription {
					return []reconcile.Request{{NamespacedName: constants.FlowCollectorName}}
				}
				return []reconcile.Request{}
			}),
			reconcilers.IgnoreStatusChange,
		)
	}
	// Return initReconcile as a post-create hook
	return r.initReconcile, b.Complete(&r)
}

func (r *Reconciler) initReconcile(ctx context.Context) error {
	attempt := 0
	err := retry.OnError(ctx, retryBackoff, func(error) bool { return true }, func() error {
		attempt++
		if _, err := r.Reconcile(ctx, ctrl.Request{}); err != nil {
			clog.WithValues("attempt", attempt, "error", err).Info("Initial reconcile: attempt failed")
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed initial reconcile, all attempts failed: %w", err)
	}
	return nil
}

// Reconcile is the controller entry point for reconciling current state with desired state.
// It manages the controller status at a high level. Business logic is delegated into `reconcile`.
func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	ctx = log.IntoContext(ctx, clog)

	commit := r.status.Reset()
	defer commit(ctx, r.Client)

	if r.mgr.ClusterInfo.HasConsolePlugin() {
		// Only deploy static plugin on OpenShift 4.15+
		if !r.mgr.ClusterInfo.IsOpenShift() {
			clog.Info("Skipping static plugin reconciler (not OpenShift)")
		} else if supported, _, err := r.mgr.ClusterInfo.IsOpenShiftVersionAtLeast("4.15.0"); err != nil {
			return ctrl.Result{}, err
		} else if !supported {
			clog.Info("Skipping static plugin reconciler (OpenShift version < 4.15)")
		} else {
			scp, err := helper.NewControllerClientHelper(ctx, r.mgr.Config.Namespace, r.Client)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to get controller deployment: %w", err)
			}
			ri, err := r.newDefaultReconcilerInstance(scp)
			if err != nil {
				return ctrl.Result{}, r.status.Error("ConsolePluginImageError", fmt.Errorf("failed to resolve console plugin image: %w", err))
			}
			staticPluginReconciler := consoleplugin.NewStaticReconciler(ri, &r.mgr.Config.StaticPluginConfig)
			if err := staticPluginReconciler.ReconcileStaticPlugin(ctx, true); err != nil {
				clog.Error(err, "Static plugin reconcile failure")
				// Set status failure unless it was already set
				if !r.status.HasFailure() {
					r.status.SetFailure("StaticPluginError", err.Error())
				}
				return ctrl.Result{}, err
			}
		}
	}

	r.status.SetReady()
	return ctrl.Result{}, nil
}

func (r *Reconciler) newDefaultReconcilerInstance(clh *helper.Client) (*reconcilers.Instance, error) {
	// force default namespace
	reconcilersInfo := reconcilers.Common{
		Client:      *clh,
		Namespace:   r.mgr.Config.Namespace,
		ClusterInfo: r.mgr.ClusterInfo,
		Watcher:     nil,
		Loki:        &helper.LokiConfig{},
		Vendor:      r.mgr.Config.Vendor,
	}
	cpImage, err := r.mgr.Config.ResolveWebConsoleImage(r.mgr.ClusterInfo)
	if err != nil {
		return nil, err
	}
	return reconcilersInfo.NewInstance(map[reconcilers.ImageRef]string{
		reconcilers.MainImage: cpImage,
	}, r.status), nil
}
