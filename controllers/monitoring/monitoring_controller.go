package monitoring

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
)

type Reconciler struct {
	client.Client
	mgr    *manager.Manager
	status status.Instance
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
		Owns(&corev1.Namespace{}).
		Complete(&r)
}

func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	r.status.SetUnknown()
	defer r.status.Commit(ctx, r.Client)

	err := r.reconcile(ctx)
	if err != nil {
		log.Error(err, "Monitoring reconcile failure")
		// Set status failure unless it was already set
		if !r.status.HasFailure() {
			r.status.SetFailure("MonitoringError", err.Error())
		}
		return ctrl.Result{}, err
	}

	r.status.SetReady()
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcile(ctx context.Context) error {
	clh, desired, err := helper.NewFlowCollectorClientHelper(ctx, r.Client)
	if err != nil {
		return fmt.Errorf("failed to get FlowCollector: %w", err)
	} else if desired == nil {
		return nil
	}

	ns := helper.GetNamespace(&desired.Spec)

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
		names := helper.GetIncludeList(&desired.Spec)
		desiredFlowDashboardCM, del, err := buildFlowMetricsDashboard(ns, names)
		if err != nil {
			return err
		} else if err = reconcilers.ReconcileConfigMap(ctx, clh, desiredFlowDashboardCM, del); err != nil {
			return err
		}

		desiredHealthDashboardCM, del, err := buildHealthDashboard(ns, names)
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
