package networkpolicy

import (
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/pkg/cluster"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	"github.com/stretchr/testify/assert"

	ascv2 "k8s.io/api/autoscaling/v2"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var outputRecordTypes = flowslatest.LogTypeAll

func getLoki() flowslatest.FlowCollectorLoki {
	return flowslatest.FlowCollectorLoki{
		Mode: flowslatest.LokiModeManual,
		Manual: flowslatest.LokiManualParams{
			IngesterURL: "http://loki:3100/",
		},
		Enable: ptr.To(true),
		WriteBatchWait: &metav1.Duration{
			Duration: 1,
		},
		WriteBatchSize: 102400,
		Advanced: &flowslatest.AdvancedLokiConfig{
			WriteMinBackoff: &metav1.Duration{
				Duration: 1,
			},
			WriteMaxBackoff: &metav1.Duration{
				Duration: 300,
			},
			WriteMaxRetries: ptr.To(int32(10)),
			StaticLabels:    map[string]string{"app": "netobserv-flowcollector"},
		},
	}
}

func getConfig() flowslatest.FlowCollector {
	return flowslatest.FlowCollector{
		Spec: flowslatest.FlowCollectorSpec{
			DeploymentModel: flowslatest.DeploymentModelDirect,
			Agent:           flowslatest.FlowCollectorAgent{Type: flowslatest.AgentEBPF},
			Processor: flowslatest.FlowCollectorFLP{
				LogLevel:              "trace",
				KafkaConsumerReplicas: ptr.To(int32(1)),
				KafkaConsumerAutoscaler: flowslatest.FlowCollectorHPA{
					Status:  flowslatest.HPAStatusEnabled,
					Metrics: []ascv2.MetricSpec{},
				},
				LogTypes: &outputRecordTypes,
				Advanced: &flowslatest.AdvancedProcessorConfig{
					Port:       ptr.To(int32(2055)),
					HealthPort: ptr.To(int32(8080)),
				},
			},
			Loki: getLoki(),
			Kafka: flowslatest.FlowCollectorKafka{
				Address: "kafka",
				Topic:   "flp",
			},
		},
	}
}

func TestNpBuilder(t *testing.T) {
	assert := assert.New(t)

	desired := getConfig()
	mgr := &manager.Manager{ClusterInfo: &cluster.Info{}}

	desired.Spec.NetworkPolicy.Enable = nil
	name, np := buildMainNetworkPolicy(&desired, mgr)
	assert.Equal(netpolName, name.Name)
	assert.Equal("netobserv", name.Namespace)
	assert.Nil(np)
	name, np = buildPrivilegedNetworkPolicy(&desired, mgr)
	assert.Equal(netpolName, name.Name)
	assert.Equal("netobserv-privileged", name.Namespace)
	assert.Nil(np)

	desired.Spec.NetworkPolicy.Enable = ptr.To(false)
	_, np = buildMainNetworkPolicy(&desired, mgr)
	assert.Nil(np)
	_, np = buildPrivilegedNetworkPolicy(&desired, mgr)
	assert.Nil(np)

	desired.Spec.NetworkPolicy.Enable = ptr.To(true)
	name, np = buildMainNetworkPolicy(&desired, mgr)
	assert.NotNil(np)
	assert.Equal(np.ObjectMeta.Name, name.Name)
	assert.Equal(np.ObjectMeta.Namespace, name.Namespace)
	assert.Equal([]networkingv1.NetworkPolicyIngressRule{
		{From: []networkingv1.NetworkPolicyPeer{
			{PodSelector: &metav1.LabelSelector{}},
			{NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": "netobserv-privileged",
				},
			}},
		}},
	}, np.Spec.Ingress)

	name, np = buildPrivilegedNetworkPolicy(&desired, mgr)
	assert.NotNil(np)
	assert.Equal(np.ObjectMeta.Name, name.Name)
	assert.Equal(np.ObjectMeta.Namespace, name.Namespace)
	assert.Equal([]networkingv1.NetworkPolicyIngressRule{}, np.Spec.Ingress)

	desired.Spec.NetworkPolicy.AdditionalNamespaces = []string{"foo", "bar"}
	name, np = buildMainNetworkPolicy(&desired, mgr)
	assert.NotNil(np)
	assert.Equal(np.ObjectMeta.Name, name.Name)
	assert.Equal(np.ObjectMeta.Namespace, name.Namespace)
	assert.Equal([]networkingv1.NetworkPolicyIngressRule{
		{From: []networkingv1.NetworkPolicyPeer{
			{PodSelector: &metav1.LabelSelector{}},
			{NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": "netobserv-privileged",
				},
			}},
		}},
		{From: []networkingv1.NetworkPolicyPeer{
			{NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": "foo",
				},
			}},
		}},
		{From: []networkingv1.NetworkPolicyPeer{
			{NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": "bar",
				},
			}},
		}},
	}, np.Spec.Ingress)

	name, np = buildPrivilegedNetworkPolicy(&desired, mgr)
	assert.NotNil(np)
	assert.Equal(np.ObjectMeta.Name, name.Name)
	assert.Equal(np.ObjectMeta.Namespace, name.Namespace)
	assert.Equal([]networkingv1.NetworkPolicyIngressRule{}, np.Spec.Ingress)
}
