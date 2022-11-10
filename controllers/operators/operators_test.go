package operators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestK8SKafkaSubscription(t *testing.T) {
	assert := assert.New(t)

	u, err := loadYaml("embed/k8s/kafka_subscription.yaml", nil)
	assert.Nil(err)
	assert.Equal("operators.coreos.com/v1alpha1", u.GetAPIVersion())
	assert.Equal("Subscription", u.GetKind())
	assert.Equal("strimzi", u.GetName())
}

func TestOpenshiftKafkaSubscription(t *testing.T) {
	assert := assert.New(t)

	u, err := loadYaml("embed/openshift/kafka_subscription.yaml", nil)
	assert.Nil(err)
	assert.Equal("operators.coreos.com/v1alpha1", u.GetAPIVersion())
	assert.Equal("Subscription", u.GetKind())
	assert.Equal("amq-streams", u.GetName())
}

func TestKafkaInstance(t *testing.T) {
	assert := assert.New(t)

	u, err := loadYaml(kafkaInstancePath, nil)
	assert.Nil(err)
	assert.Equal("kafka.strimzi.io/v1beta2", u.GetAPIVersion())
	assert.Equal("Kafka", u.GetKind())
	assert.Equal("kafka-cluster", u.GetName())
}

func TestKafkaTopic(t *testing.T) {
	assert := assert.New(t)

	u, err := loadYaml(kafkaTopicPath, nil)
	assert.Nil(err)
	assert.Equal("kafka.strimzi.io/v1beta2", u.GetAPIVersion())
	assert.Equal("KafkaTopic", u.GetKind())
	assert.Equal("network-flows", u.GetName())
}

func TestK8SLokiSubscription(t *testing.T) {
	assert := assert.New(t)

	u, err := loadYaml("embed/k8s/loki_subscription.yaml", nil)
	assert.Nil(err)
	assert.Equal("operators.coreos.com/v1alpha1", u.GetAPIVersion())
	assert.Equal("Subscription", u.GetKind())
	assert.Equal("loki-operator", u.GetName())
}

func TestOpenshiftLokiSubscription(t *testing.T) {
	assert := assert.New(t)

	u, err := loadYaml("embed/openshift/loki_subscription.yaml", nil)
	assert.Nil(err)
	assert.Equal("operators.coreos.com/v1alpha1", u.GetAPIVersion())
	assert.Equal("Subscription", u.GetKind())
	assert.Equal("loki-operator", u.GetName())
}

func TestLokiInstance(t *testing.T) {
	assert := assert.New(t)

	u, err := loadYaml(lokiInstancePath, nil)
	assert.Nil(err)
	assert.Equal("loki.grafana.com/v1", u.GetAPIVersion())
	assert.Equal("LokiStack", u.GetKind())
	assert.Equal("lokistack", u.GetName())
}

func TestInject(t *testing.T) {
	assert := assert.New(t)

	u, err := loadYaml(operatorGroupPath, &JSONInterface{
		"metadata": JSONInterface{
			"name":      "test",
			"namespace": "test-ns",
		},
	})

	assert.Nil(err)
	assert.Equal("test", u.GetName())
	assert.Equal("test-ns", u.GetNamespace())

}
