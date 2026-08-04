package static

import (
	"context"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowslatest "github.com/netobserv/netobserv-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/netobserv-operator/internal/controller/constants"
	"github.com/netobserv/netobserv-operator/internal/controller/reconcilers"
	"github.com/netobserv/netobserv-operator/internal/pkg/helper"
	"github.com/netobserv/netobserv-operator/internal/pkg/manager"
	"github.com/netobserv/netobserv-operator/internal/pkg/netpol"
)

type operatorReconciler struct {
	*reconcilers.Instance
	cfg    *manager.Config
	netpol *networkingv1.NetworkPolicy
}

func newOperatorReconciler(cmn *reconcilers.Instance, cfg *manager.Config) *operatorReconciler {
	return &operatorReconciler{
		Instance: cmn,
		cfg:      cfg,
		netpol:   cmn.Managed.NewNetworkPolicy(constants.OperatorName),
	}
}

func (r *operatorReconciler) reconcile(ctx context.Context) error {
	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	return r.reconcileNetpol(ctx)
}

func (r *operatorReconciler) reconcileNetpol(ctx context.Context) error {
	cni, err := r.ClusterInfo.GetCNI()
	if err != nil {
		return err
	}

	specStub := flowslatest.FlowCollectorSpec{NetworkPolicy: flowslatest.NetworkPolicy{Enable: r.cfg.DeployOperatorNetworkPolicy}}

	if !specStub.DeployNetworkPolicy(cni) {
		r.Managed.TryDelete(ctx, r.netpol)
		return nil
	}

	ingress := []networkingv1.NetworkPolicyIngressRule{
		netpol.AllowToWebhooks(r.ClusterInfo, r.Vendor),
	}
	egress := []networkingv1.NetworkPolicyEgressRule{
		netpol.AllowToAPIServer(r.ClusterInfo, r.Vendor),
		netpol.AllowDNS(r.Vendor),
	}

	if r.ClusterInfo.IsOpenShift() {
		ingress = append(ingress, netpol.AllowVendorPrometheusScrape(r.Vendor, constants.OperatorMetricsPort))
	}

	desired := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.OperatorName,
			Namespace: r.cfg.Namespace,
			Labels: map[string]string{
				"part-of": constants.OperatorName,
				"app":     constants.OperatorName,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": constants.OperatorName,
				},
			},
			Ingress: ingress,
			Egress:  egress,
		},
	}
	nsname := helper.NamespacedName(desired)

	return reconcilers.ReconcileNetworkPolicy(ctx, &r.Client, nsname, desired)
}
