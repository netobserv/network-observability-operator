package static

import (
	"testing"

	flowslatest "github.com/netobserv/netobserv-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/netobserv-operator/internal/controller/constants"
	"github.com/netobserv/netobserv-operator/internal/controller/reconcilers"
	"github.com/netobserv/netobserv-operator/internal/pkg/cluster"
	"github.com/netobserv/netobserv-operator/internal/pkg/manager"
	"github.com/netobserv/netobserv-operator/internal/pkg/manager/status"
	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func TestBuildNetworkPolicy_OVN_Knobs(t *testing.T) {
	cni := flowslatest.OVNKubernetes
	clust := cluster.Mock(cluster.WithCNI(cni))
	info := reconcilers.Common{Namespace: "noo", ClusterInfo: clust}
	rec := operatorReconciler{cfg: &manager.Config{}, Instance: info.NewInstance(nil, status.Instance{})}

	np := rec.buildNetworkPolicy(cni)
	assert.NotNil(t, np)

	rec.cfg.DeployOperatorNetworkPolicy = ptr.To(false)

	np = rec.buildNetworkPolicy(cni)
	assert.Nil(t, np)

	rec.cfg.DeployOperatorNetworkPolicy = ptr.To(true)

	np = rec.buildNetworkPolicy(cni)
	assert.NotNil(t, np)
}

func TestBuildNetworkPolicy_SDN_Knobs(t *testing.T) {
	cni := flowslatest.OpenShiftSDN
	clust := cluster.Mock(cluster.WithCNI(cni))
	info := reconcilers.Common{Namespace: "noo", ClusterInfo: clust}
	rec := operatorReconciler{cfg: &manager.Config{}, Instance: info.NewInstance(nil, status.Instance{})}

	np := rec.buildNetworkPolicy(cni)
	assert.Nil(t, np)

	rec.cfg.DeployOperatorNetworkPolicy = ptr.To(false)

	np = rec.buildNetworkPolicy(cni)
	assert.Nil(t, np)

	rec.cfg.DeployOperatorNetworkPolicy = ptr.To(true)

	np = rec.buildNetworkPolicy(cni)
	assert.Nil(t, np)
}

func TestBuildNetworkPolicy_Kindnet_Knobs(t *testing.T) {
	cni := flowslatest.Kindnet
	clust := cluster.Mock(cluster.WithCNI(cni))
	info := reconcilers.Common{Namespace: "noo", ClusterInfo: clust}
	rec := operatorReconciler{cfg: &manager.Config{}, Instance: info.NewInstance(nil, status.Instance{})}

	np := rec.buildNetworkPolicy(cni)
	assert.NotNil(t, np)

	rec.cfg.DeployOperatorNetworkPolicy = ptr.To(false)

	np = rec.buildNetworkPolicy(cni)
	assert.Nil(t, np)

	rec.cfg.DeployOperatorNetworkPolicy = ptr.To(true)

	np = rec.buildNetworkPolicy(cni)
	assert.NotNil(t, np)
}

func TestBuildNetworkPolicy_Unknown_Knobs(t *testing.T) {
	cni := flowslatest.NetworkType("")
	clust := cluster.Mock(cluster.WithCNI(cni))
	info := reconcilers.Common{Namespace: "noo", ClusterInfo: clust}
	rec := operatorReconciler{cfg: &manager.Config{}, Instance: info.NewInstance(nil, status.Instance{})}

	np := rec.buildNetworkPolicy(cni)
	assert.Nil(t, np)

	rec.cfg.DeployOperatorNetworkPolicy = ptr.To(false)

	np = rec.buildNetworkPolicy(cni)
	assert.Nil(t, np)

	rec.cfg.DeployOperatorNetworkPolicy = ptr.To(true)

	np = rec.buildNetworkPolicy(cni)
	assert.NotNil(t, np)
}

func TestBuildNetworkPolicy_Kindnet(t *testing.T) {
	cni := flowslatest.Kindnet
	clust := cluster.Mock(cluster.WithCNI(cni), cluster.WithKubeIPs("172.0.0.1"), cluster.WithKubePorts(443))
	info := reconcilers.Common{Namespace: "noo", ClusterInfo: clust}
	rec := operatorReconciler{cfg: &manager.Config{}, Instance: info.NewInstance(nil, status.Instance{})}

	np := rec.buildNetworkPolicy(cni)
	assert.NotNil(t, np)

	assert.Equal(t, []networkingv1.NetworkPolicyIngressRule{
		{
			From: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"}}},
			}, Ports: []networkingv1.NetworkPolicyPort{{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromInt(9443))}},
		},
	}, np.Spec.Ingress)

	assert.Equal(t, []networkingv1.NetworkPolicyEgressRule{
		{
			To: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"},
				}},
				{IPBlock: &networkingv1.IPBlock{
					CIDR: "172.0.0.1/32",
				}},
			}, Ports: []networkingv1.NetworkPolicyPort{
				{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromInt(443))},
			},
		},
		{
			To: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"},
				}},
			}, Ports: []networkingv1.NetworkPolicyPort{
				{Protocol: ptr.To(v1.ProtocolUDP), Port: ptr.To(intstr.FromString("dns"))},
				{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromString("dns-tcp"))},
			},
		},
	}, np.Spec.Egress)
}

func TestBuildNetworkPolicy_OVN(t *testing.T) {
	cni := flowslatest.OVNKubernetes
	clust := cluster.Mock(cluster.WithCNI(cni), cluster.WithKubeIPs("172.0.0.1"), cluster.WithKubePorts(6443))
	info := reconcilers.Common{Namespace: "noo", ClusterInfo: clust}
	rec := operatorReconciler{cfg: &manager.Config{}, Instance: info.NewInstance(nil, status.Instance{})}

	np := rec.buildNetworkPolicy(cni)
	assert.NotNil(t, np)

	assert.Equal(t, []networkingv1.NetworkPolicyIngressRule{
		{
			From: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": "ovn-host-network"}}},
			}, Ports: []networkingv1.NetworkPolicyPort{{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromInt(9443))}},
		},
	}, np.Spec.Ingress)

	assert.Equal(t, []networkingv1.NetworkPolicyEgressRule{
		{
			To: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"},
				}},
				{IPBlock: &networkingv1.IPBlock{
					CIDR: "172.0.0.1/32",
				}},
			}, Ports: []networkingv1.NetworkPolicyPort{
				{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromInt(443))},
				{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromInt(6443))},
			},
		},
		{
			To: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"},
				}},
			}, Ports: []networkingv1.NetworkPolicyPort{
				{Protocol: ptr.To(v1.ProtocolUDP), Port: ptr.To(intstr.FromString("dns"))},
				{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromString("dns-tcp"))},
			},
		},
	}, np.Spec.Egress)
}

func TestBuildNetworkPolicy_OVN_OpenShift(t *testing.T) {
	cni := flowslatest.OVNKubernetes
	clust := cluster.Mock(cluster.WithOpenShiftVersion("4.20.0"), cluster.WithCNI(cni), cluster.WithKubeIPs("172.0.0.1"), cluster.WithKubePorts(6443))
	info := reconcilers.Common{Namespace: "noo", ClusterInfo: clust, Vendor: constants.VendorOpenShiftDownstream}
	rec := operatorReconciler{cfg: &manager.Config{}, Instance: info.NewInstance(nil, status.Instance{})}

	np := rec.buildNetworkPolicy(cni)
	assert.NotNil(t, np)

	assert.Equal(t, []networkingv1.NetworkPolicyIngressRule{
		{
			From: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"policy-group.network.openshift.io/host-network": ""}}},
			}, Ports: []networkingv1.NetworkPolicyPort{{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromInt(9443))}},
		},
		{
			From: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-monitoring"}}},
			}, Ports: []networkingv1.NetworkPolicyPort{{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromInt(8443))}},
		},
	}, np.Spec.Ingress)

	assert.Equal(t, []networkingv1.NetworkPolicyEgressRule{
		{
			To: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{{
						Key:      "kubernetes.io/metadata.name",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"openshift-apiserver", "openshift-kube-apiserver"},
					}},
				}},
				{IPBlock: &networkingv1.IPBlock{
					CIDR: "172.0.0.1/32",
				}},
			}, Ports: []networkingv1.NetworkPolicyPort{
				{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromInt(6443))},
			},
		},
		{
			To: []networkingv1.NetworkPolicyPeer{
				{NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"kubernetes.io/metadata.name": "openshift-dns"},
				}},
			}, Ports: []networkingv1.NetworkPolicyPort{
				{Protocol: ptr.To(v1.ProtocolUDP), Port: ptr.To(intstr.FromString("dns"))},
				{Protocol: ptr.To(v1.ProtocolTCP), Port: ptr.To(intstr.FromString("dns-tcp"))},
			},
		},
	}, np.Spec.Egress)
}
