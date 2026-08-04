package consoleplugin

import (
	"context"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/netobserv-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/netobserv-operator/internal/controller/constants"
	"github.com/netobserv/netobserv-operator/internal/controller/reconcilers"
	"github.com/netobserv/netobserv-operator/internal/pkg/helper"
	"github.com/netobserv/netobserv-operator/internal/pkg/manager"
	"github.com/netobserv/netobserv-operator/internal/pkg/netpol"
)

type StaticReconciler struct {
	CPReconciler
	managerConfig *manager.Config
	netpol        *networkingv1.NetworkPolicy
}

func NewStaticReconciler(cmn *reconcilers.Instance, cfg *manager.Config) StaticReconciler {
	return StaticReconciler{
		CPReconciler: CPReconciler{
			Instance:       cmn,
			deployment:     cmn.Managed.NewDeployment(constants.StaticPluginName),
			service:        cmn.Managed.NewService(constants.StaticPluginName),
			serviceAccount: cmn.Managed.NewServiceAccount(constants.StaticPluginName),
		},
		netpol:        cmn.Managed.NewNetworkPolicy(constants.StaticPluginName),
		managerConfig: cfg,
	}
}

func (r *StaticReconciler) ReconcileStaticPlugin(ctx context.Context, enable bool) error {
	// Retrieve toleration from subscription
	var sched *flowslatest.SchedulingConfig
	subName := r.managerConfig.StaticPluginConfig.InheritTolerationFromSubscription
	if subName != "" {
		sub := olm.Subscription{}
		if err := r.Client.Get(ctx, types.NamespacedName{Name: subName, Namespace: r.Namespace}, &sub); err != nil {
			return err
		}
		if sub.Spec != nil && sub.Spec.Config != nil {
			sched = &flowslatest.SchedulingConfig{
				Tolerations:  sub.Spec.Config.Tolerations,
				NodeSelector: sub.Spec.Config.NodeSelector,
				Affinity:     sub.Spec.Config.Affinity,
			}
		}
	}

	// Fake a FlowCollector to create console plugin and expose forms
	return r.reconcileStatic(ctx, &flowslatest.FlowCollector{
		Spec: flowslatest.FlowCollectorSpec{
			NetworkPolicy: flowslatest.NetworkPolicy{
				Enable: r.managerConfig.DeployOperatorNetworkPolicy,
			},
			ConsolePlugin: flowslatest.FlowCollectorConsolePlugin{
				Enable:   ptr.To(enable),
				LogLevel: "info",
				Advanced: &flowslatest.AdvancedPluginConfig{
					Register:   ptr.To(true),
					Scheduling: sched,
				},
			},
		},
	})
}

// Reconcile is the reconciler entry point to reconcile the static plugin state with the desired configuration
func (r *StaticReconciler) reconcileStatic(ctx context.Context, desired *flowslatest.FlowCollector) error {
	l := log.FromContext(ctx).WithName("static-console-plugin")
	ctx = log.IntoContext(ctx, l)

	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if r.ClusterInfo.HasConsolePlugin() {
		r.checkAutoPatch(ctx, desired, constants.StaticPluginName)

		// Create object builder
		builder := newBuilder(r.Instance, &desired.Spec, constants.StaticPluginName)

		if err = r.reconcilePlugin(ctx, &builder, constants.StaticPluginName, "NetObserv static plugin"); err != nil {
			return err
		}

		if err = r.reconcileDeployment(ctx, &builder, &desired.Spec, constants.StaticPluginName, ""); err != nil {
			return err
		}

		if err = r.reconcileServices(ctx, &builder, constants.StaticPluginName); err != nil {
			return err
		}

		if err = r.reconcileNetpol(ctx, desired); err != nil {
			return err
		}
	} else {
		// delete any existing owned object
		r.Managed.TryDeleteAll(ctx)
	}

	return nil
}

func (r *StaticReconciler) reconcileNetpol(ctx context.Context, desired *flowslatest.FlowCollector) error {
	cni, err := r.ClusterInfo.GetCNI()
	if err != nil {
		return err
	}

	if !desired.Spec.DeployNetworkPolicy(cni) {
		r.Managed.TryDelete(ctx, r.netpol)
		return nil
	}

	ingress := []networkingv1.NetworkPolicyIngressRule{
		netpol.AllowFromOpenShiftConsole(&desired.Spec.ConsolePlugin),
	}

	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.StaticPluginName,
			Namespace: r.Namespace,
			Labels: map[string]string{
				"part-of": constants.OperatorName,
				"app":     constants.StaticPluginName,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": constants.StaticPluginName,
				},
			},
			Ingress: ingress,
			Egress:  nil,
		},
	}
	nsname := helper.NamespacedName(policy)

	return reconcilers.ReconcileNetworkPolicy(ctx, &r.Client, nsname, policy)
}
