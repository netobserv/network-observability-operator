package networkpolicy

import (
	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/internal/controller/constants"
	"github.com/netobserv/network-observability-operator/internal/pkg/helper"
	"github.com/netobserv/network-observability-operator/internal/pkg/manager"
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
	ns := desired.Spec.GetNamespace()

	name := types.NamespacedName{Name: netpolName, Namespace: ns}
	if !desired.Spec.DeployNetworkPolicy() {
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
				},
			}},
			Egress: []networkingv1.NetworkPolicyEgressRule{{
				To: []networkingv1.NetworkPolicyPeer{
					{
						// Setting empty namespace selector will authorize every pod from the same namespace
						PodSelector: &metav1.LabelSelector{},
					},
				},
			}},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					// TODO: remove this restiction when LokiStack implements network policy by default
					"part-of": constants.OperatorName,
				},
			},
		},
	}
	// Allow traffic from the eBPF agents
	np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			peerInNamespace(privNs),
		},
	})
	np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			peerInNamespace(privNs),
		},
	})

	if mgr.ClusterInfo.IsOpenShift() {
		if desired.Spec.UseConsolePlugin() && mgr.ClusterInfo.HasConsolePlugin() {
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
			np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
				To: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(constants.ConsoleNamespace),
				},
				Ports: []networkingv1.NetworkPolicyPort{{
					Protocol: ptr.To(corev1.ProtocolTCP),
					Port:     ptr.To(intstr.FromInt32(*advanced.Port)),
				}},
			})
		}
		np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
			// Console plugin pod needs access to cluster monitoring, see its configured URL, even with upstream deployment
			To: []networkingv1.NetworkPolicyPeer{
				peerInNamespace(constants.MonitoringNamespace),
			},
		})
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
		// Allow apiserver/host
		np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				{
					NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"policy-group.network.openshift.io/host-network": "",
					}},
				},
			},
			Ports: []networkingv1.NetworkPolicyPort{{
				Protocol: ptr.To(corev1.ProtocolTCP),
				Port:     ptr.To(intstr.FromInt32(constants.WebhookPort)),
			}},
		})
		// Allow host
		np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
			To: []networkingv1.NetworkPolicyPeer{},
			Ports: []networkingv1.NetworkPolicyPort{{
				Protocol: ptr.To(corev1.ProtocolTCP),
				Port:     ptr.To(intstr.FromInt32(constants.K8sAPIServerPort)),
			}},
		})
		// Allow apiserver
		np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				{
					NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"policy-group.network.openshift.io/host-network": "",
					}},
				},
			},
			Ports: []networkingv1.NetworkPolicyPort{{
				Protocol: ptr.To(corev1.ProtocolTCP),
				Port:     ptr.To(intstr.FromInt32(constants.WebhookPort)),
			}},
		})
		np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
			To: []networkingv1.NetworkPolicyPeer{
				peerInNamespace(constants.DNSNamespace),
			},
		})
		if desired.Spec.UseLoki() && desired.Spec.Loki.Mode == flowslatest.LokiModeLokiStack {
			np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
				To: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(desired.Spec.Loki.LokiStack.Namespace),
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
		np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
			To: []networkingv1.NetworkPolicyPeer{
				peerInNamespace(aNs),
			},
		})

	}

	return name, &np
}

func buildPrivilegedNetworkPolicy(desired *flowslatest.FlowCollector, mgr *manager.Manager) (types.NamespacedName, *networkingv1.NetworkPolicy) {
	mainNs := desired.Spec.GetNamespace()
	privNs := mainNs + constants.EBPFPrivilegedNSSuffix

	name := types.NamespacedName{Name: netpolName, Namespace: privNs}
	if !desired.Spec.DeployNetworkPolicy() {
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
			Egress:  []networkingv1.NetworkPolicyEgressRule{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
		},
	}

	np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			peerInNamespace(mainNs),
		},
	})

	if mgr.ClusterInfo.IsOpenShift() {
		if mgr.Config.DownstreamDeployment {
			np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
				From: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(constants.MonitoringNamespace),
				},
			})
			np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
				To: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(constants.MonitoringNamespace),
				},
			})

		} else {
			np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
				From: []networkingv1.NetworkPolicyPeer{
					peerInNamespace(constants.UWMonitoringNamespace),
				},
			})
			np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
				To: []networkingv1.NetworkPolicyPeer{
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
		np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
			To: []networkingv1.NetworkPolicyPeer{
				peerInNamespace(aNs),
			},
		})

	}

	return name, &np
}
