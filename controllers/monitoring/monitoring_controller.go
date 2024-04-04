package monitoring

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/metrics"
)

type Reconciler struct {
	client.Client
	mgr              *manager.Manager
	status           status.Instance
	currentNamespace string
}

func Start(ctx context.Context, mgr *manager.Manager) error {
	log := log.FromContext(ctx)
	log.Info("Starting Monitoring controller")
	r := Reconciler{
		Client: mgr.Client,
		mgr:    mgr,
		status: mgr.Status.ForComponent(status.Monitoring),
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&flowslatest.FlowCollector{}, reconcilers.IgnoreStatusChange).
		Named("monitoring").
		Owns(&corev1.Namespace{}).
		Watches(
			&metricslatest.FlowMetric{},
			handler.EnqueueRequestsFromMapFunc(func(_ context.Context, o client.Object) []reconcile.Request {
				if o.GetNamespace() == r.currentNamespace {
					return []reconcile.Request{{NamespacedName: constants.FlowCollectorName}}
				}
				return []reconcile.Request{}
			}),
		).
		Complete(&r)
}

// Reconcile is the controller entry point for reconciling current state with desired state.
// It manages the controller status at a high level. Business logic is delegated into `reconcile`.
func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	l := log.Log.WithName("monitoring") // clear context (too noisy)
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
		l.Error(err, "Monitoring reconcile failure")
		// Set status failure unless it was already set
		if !r.status.HasFailure() {
			r.status.SetFailure("MonitoringError", err.Error())
		}
		return ctrl.Result{}, err
	}

	r.status.SetReady()
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcile(ctx context.Context, clh *helper.Client, desired *flowslatest.FlowCollector) error {
	log := log.FromContext(ctx)
	ns := helper.GetNamespace(&desired.Spec)
	r.currentNamespace = ns

	// If namespace does not exist, we create it
	nsExist, err := r.namespaceExist(ctx, ns)
	if err != nil {
		return err
	}
	desiredNs := buildNamespace(ns, r.mgr.Config.DownstreamDeployment)
	if nsExist == nil {
		err = r.Create(ctx, desiredNs)
		if err != nil {
			return err
		}
	} else if !helper.IsSubSet(nsExist.ObjectMeta.Labels, desiredNs.ObjectMeta.Labels) {
		err = r.Update(ctx, desiredNs)
		if err != nil {
			return err
		}
	}
	if r.mgr.Config.DownstreamDeployment {
		desiredRole := buildRoleMonitoringReader()
		if err := reconcilers.ReconcileClusterRole(ctx, clh, desiredRole); err != nil {
			return err
		}
		desiredBinding := buildRoleBindingMonitoringReader(ns)
		if err := reconcilers.ReconcileClusterRoleBinding(ctx, clh, desiredBinding); err != nil {
			return err
		}
	}

	if r.mgr.HasSvcMonitor() {
		// List custom metrics
		fm := metricslatest.FlowMetricList{}
		if err := r.Client.List(ctx, &fm, &client.ListOptions{Namespace: ns}); err != nil {
			return r.status.Error("CantListFlowMetrics", err)
		}
		log.WithValues("items count", len(fm.Items)).Info("FlowMetrics loaded")

		allMetrics := metrics.MergePredefined(fm.Items, &desired.Spec)
		log.WithValues("metrics count", len(allMetrics)).Info("Merged metrics")

		desiredFlowDashboardCM, del, err := buildFlowMetricsDashboard(allMetrics)
		if err != nil {
			return err
		} else if err = reconcilers.ReconcileConfigMap(ctx, clh, desiredFlowDashboardCM, del); err != nil {
			return err
		}

		desiredHealthDashboardCM, del, err := buildHealthDashboard(ns)
		if err != nil {
			return err
		} else if err = reconcilers.ReconcileConfigMap(ctx, clh, desiredHealthDashboardCM, del); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) namespaceExist(ctx context.Context, nsName string) (*corev1.Namespace, error) {
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
