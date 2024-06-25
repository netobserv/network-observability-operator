package networkpolicy

import (
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func buildNetworkPolicy(ns string, desired *flowslatest.FlowCollector, additionalNs []string) (types.NamespacedName, *networkingv1.NetworkPolicy) {
	name := types.NamespacedName{Name: constants.OperatorName, Namespace: ns}
	if desired.Spec.NetworkPolicy.Enable == nil || !*desired.Spec.NetworkPolicy.Enable {
		return name, nil
	}
	np := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.OperatorName,
			Namespace: ns,
		},
		Spec: networkingv1.NetworkPolicySpec{
			Ingress: []networkingv1.NetworkPolicyIngressRule{{
				From: []networkingv1.NetworkPolicyPeer{
					{PodSelector: &metav1.LabelSelector{}}, //Setting empty namespace selector will authorize every pod from the same namespace
				},
			}},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
		},
	}
	for _, aNs := range additionalNs {
		np.Spec.Ingress[0].From = append(np.Spec.Ingress[0].From, networkingv1.NetworkPolicyPeer{PodSelector: &metav1.LabelSelector{},
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": aNs,
				},
			},
		})
	}
	return name, &np
}
