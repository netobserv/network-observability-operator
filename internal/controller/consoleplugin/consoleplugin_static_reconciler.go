package consoleplugin

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/netobserv-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/netobserv-operator/internal/controller/constants"
	"github.com/netobserv/netobserv-operator/internal/controller/reconcilers"
	"github.com/netobserv/netobserv-operator/internal/pkg/helper"
	osv1 "github.com/openshift/api/console/v1"
)

const (
	staticPluginFinalizer = "staticplugin.netobserv.io/finalizer"
)

func NewStaticReconciler(cmn *reconcilers.Instance) CPReconciler {
	return CPReconciler{
		Instance:       cmn,
		deployment:     cmn.Managed.NewDeployment(constants.StaticPluginName),
		service:        cmn.Managed.NewService(constants.StaticPluginName),
		serviceAccount: cmn.Managed.NewServiceAccount(constants.StaticPluginName),
	}
}

func (r *CPReconciler) ReconcileStaticPlugin(ctx context.Context, enable bool) error {
	// Fake a FlowCollector to create console plugin and expose forms
	return r.reconcileStatic(ctx, &flowslatest.FlowCollector{
		Spec: flowslatest.FlowCollectorSpec{
			ConsolePlugin: flowslatest.FlowCollectorConsolePlugin{
				Enable:   ptr.To(enable),
				LogLevel: "info",
			},
		},
	})
}

// Reconcile is the reconciler entry point to reconcile the static plugin state with the desired configuration
func (r *CPReconciler) reconcileStatic(ctx context.Context, desired *flowslatest.FlowCollector) error {
	l := log.FromContext(ctx).WithName("static-console-plugin")
	ctx = log.IntoContext(ctx, l)

	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if r.ClusterInfo.HasConsolePlugin() {
		if r.IsDeleting {
			// Process cluster-scope resources cleanup
			if err := r.cleanupClusterScope(ctx); err != nil {
				return err
			}
			// Ignore delete request for namespace-scope resources, as it is managed via owner reference
			return nil
		} else if err := r.AddFinalizer(ctx, staticPluginFinalizer); err != nil {
			return err
		}

		// Create object builder
		builder := newBuilder(r.Instance, &desired.Spec, constants.StaticPluginName)

		if err = r.reconcilePlugin(ctx, &builder, &desired.Spec, constants.StaticPluginName, "NetObserv static plugin"); err != nil {
			return err
		}

		if err = r.reconcileDeployment(ctx, &builder, &desired.Spec, constants.StaticPluginName, ""); err != nil {
			return err
		}

		if err = r.reconcileServices(ctx, &builder, constants.StaticPluginName); err != nil {
			return err
		}
	} else {
		// delete any existing owned object
		r.Managed.TryDeleteAll(ctx)
	}

	return nil
}

func (r *CPReconciler) cleanupClusterScope(ctx context.Context) error {
	plg := osv1.ConsolePlugin{}
	if err := r.Get(ctx, types.NamespacedName{Name: constants.StaticPluginName}, &plg); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	} else if helper.IsManaged(&plg) {
		if err := r.Client.Delete(ctx, &plg); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return r.RemoveFinalizer(ctx, staticPluginFinalizer)
}
