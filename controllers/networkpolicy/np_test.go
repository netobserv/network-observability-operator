package networkpolicy

import (
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/stretchr/testify/assert"

	ascv2 "k8s.io/api/autoscaling/v2"
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

	ns := "namespace1"
	desired := getConfig()
	additionalNs := []string{}

	desired.Spec.NetworkPolicy.Enable = nil
	name, np := buildNetworkPolicy(ns, &desired, additionalNs)
	assert.Equal(name.Name, constants.OperatorName)
	assert.Equal(name.Namespace, ns)
	assert.Nil(np)

	desired.Spec.NetworkPolicy.Enable = ptr.To(false)
	name, np = buildNetworkPolicy(ns, &desired, additionalNs)
	assert.Nil(np)

	desired.Spec.NetworkPolicy.Enable = ptr.To(true)
	name, np = buildNetworkPolicy(ns, &desired, additionalNs)
	assert.NotNil(np)
	assert.Equal(np.ObjectMeta.Name, name.Name)
	assert.Equal(np.ObjectMeta.Namespace, name.Namespace)
	assert.Len(np.Spec.Ingress, 1)
	assert.Len(np.Spec.Ingress[0].From, 1)

	name, np = buildNetworkPolicy(ns, &desired, []string{"foo", "bar"})
	assert.NotNil(np)
	assert.Equal(np.ObjectMeta.Name, name.Name)
	assert.Equal(np.ObjectMeta.Namespace, name.Namespace)
	assert.Len(np.Spec.Ingress, 1)
	assert.Len(np.Spec.Ingress[0].From, 3)
}
