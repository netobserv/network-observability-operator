package networkpolicy

import (
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

const netpolName = "netobserv"

func peerInNamespace(ns string) networkingv1.NetworkPolicyPeer {
	return networkingv1.NetworkPolicyPeer{
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"kubernetes.io/metadata.name": ns},
		},
	}
}

func buildMainNetworkPolicy(desired *flowslatest.FlowCollector, mgr *manager.Manager) (types.NamespacedName, *networkingv1.NetworkPolicy) {
	ns := helper.GetNamespace(&desired.Spec)

	name := types.NamespacedName{Name: netpolName, Namespace: ns}
	if desired.Spec.NetworkPolicy.Enable == nil || !*desired.Spec.NetworkPolicy.Enable {
		return name, nil
	}

	privNs := ns + constants.EBPFPrivilegedNSSuffix

	np := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      netpolName,
			Namespace: ns,
		},
		Spec: networkingv1.NetworkPolicySpec{
			Ingress: []networkingv1.NetworkPolicyIngressRule{{
				From: []networkingv1.NetworkPolicyPeer{
					{
						// Setting empty namespace selector will authorize every pod from the same namespace
						PodSelector: &metav1.LabelSelector{},
					},
					// Allow traffic from the eBPF agents
					peerInNamespace(privNs),
				},
			}},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
		},
	}

	if mgr.ClusterInfo.IsOpenShift() {
		if helper.UseConsolePlugin(&desired.Spec) && mgr.ClusterInfo.HasConsolePlugin() {
			advanced := helper.GetAdvancedPluginConfig(desired.Spec.ConsolePlugin.Advanced)
			np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
				From: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(constants.ConsoleNamespace),
				},
				Ports: []networkingv1.NetworkPolicyPort{{
					Protocol: ptr.To(corev1.ProtocolTCP),
					Port:     ptr.To(intstr.FromInt32(*advanced.Port)),
				}},
			})
		}
		if mgr.Config.DownstreamDeployment {
			np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
				From: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(constants.MonitoringNamespace),
				},
			})
		} else {
			np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
				From: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(constants.UWMonitoringNamespace),
				},
			})
		}
	}

	for _, aNs := range desired.Spec.NetworkPolicy.AdditionalNamespaces {
		np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				peerInNamespace(aNs),
			},
		})
	}

	return name, &np
}

func buildPrivilegedNetworkPolicy(desired *flowslatest.FlowCollector, mgr *manager.Manager) (types.NamespacedName, *networkingv1.NetworkPolicy) {
	mainNs := helper.GetNamespace(&desired.Spec)
	privNs := mainNs + constants.EBPFPrivilegedNSSuffix

	name := types.NamespacedName{Name: netpolName, Namespace: privNs}
	if desired.Spec.NetworkPolicy.Enable == nil || !*desired.Spec.NetworkPolicy.Enable {
		return name, nil
	}

	np := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      netpolName,
			Namespace: privNs,
		},
		Spec: networkingv1.NetworkPolicySpec{
			// Start with no allowed traffic
			Ingress: []networkingv1.NetworkPolicyIngressRule{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
		},
	}

	if mgr.ClusterInfo.IsOpenShift() {
		if mgr.Config.DownstreamDeployment {
			np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
				From: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(constants.MonitoringNamespace),
				},
			})
		} else {
			np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
				From: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(constants.UWMonitoringNamespace),
				},
			})
		}
	}

	return name, &np
}
