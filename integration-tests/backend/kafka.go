package e2etests

import (
	"context"
	"fmt"
	"time"

	o "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// Kafka struct to handle default Kafka installation
type Kafka struct {
	Name      string
	Namespace string
	Template  string
}

// KafkaMetrics struct to handle kafka metrics config deployment
type KafkaMetrics struct {
	Namespace string
	Template  string
}

// KafkaNodePool struct handles creation of kafka node pool
type KafkaNodePool struct {
	Namespace    string
	NodePoolName string
	Name         string
	Template     string
}

// KafkaTopic struct handles creation of kafka topic
type KafkaTopic struct {
	Namespace string
	TopicName string
	Name      string
	Template  string
}

type KafkaUser struct {
	Namespace string
	UserName  string
	Name      string
	Template  string
}

// deploys default Kafka
func (kafka *Kafka) deployKafka() {
	e2e.Logf("Deploy Default Kafka")
	parameters := []string{"--ignore-unknown-parameters=true", "-f", kafka.Template, "-p", "NAMESPACE=" + kafka.Namespace}

	if kafka.Name != "" {
		parameters = append(parameters, "NAME="+kafka.Name)
	}

	err := applyNsResourceFromTemplateByAdmin(kafka.Namespace, parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// deploys Kafka Metrics
func (kafkaMetrics *KafkaMetrics) deployKafkaMetrics() {
	e2e.Logf("Deploy Kafka metrics")
	parameters := []string{"--ignore-unknown-parameters=true", "-f", kafkaMetrics.Template, "-p", "NAMESPACE=" + kafkaMetrics.Namespace}

	err := applyNsResourceFromTemplateByAdmin(kafkaMetrics.Namespace, parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// creates a Kafka topic
func (kafkaTopic *KafkaTopic) deployKafkaTopic() {
	e2e.Logf("Create Kafka topic")
	parameters := []string{"--ignore-unknown-parameters=true", "-f", kafkaTopic.Template, "-p", "NAMESPACE=" + kafkaTopic.Namespace}

	if kafkaTopic.Name != "" {
		parameters = append(parameters, "NAME="+kafkaTopic.Name)
	}

	if kafkaTopic.TopicName != "" {
		parameters = append(parameters, "TOPIC="+kafkaTopic.TopicName)
	}

	err := applyNsResourceFromTemplateByAdmin(kafkaTopic.Namespace, parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// creates a Kafka nodePool
func (kafkaNodePool *KafkaNodePool) deployKafkaNodePool() {
	e2e.Logf("Create Kafka nodePool")
	parameters := []string{"--ignore-unknown-parameters=true", "-f", kafkaNodePool.Template, "-p", "NAMESPACE=" + kafkaNodePool.Namespace}

	if kafkaNodePool.Name != "" {
		parameters = append(parameters, "NAME="+kafkaNodePool.Name)
	}

	if kafkaNodePool.NodePoolName != "" {
		parameters = append(parameters, "NODEPOOL="+kafkaNodePool.NodePoolName)
	}

	err := applyNsResourceFromTemplateByAdmin(kafkaNodePool.Namespace, parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// deploys KafkaUser
func (kafkaUser *KafkaUser) deployKafkaUser() {
	e2e.Logf("Create Kafka User")
	parameters := []string{"--ignore-unknown-parameters=true", "-f", kafkaUser.Template, "-p", "NAMESPACE=" + kafkaUser.Namespace}

	if kafkaUser.UserName != "" {
		parameters = append(parameters, "USER_NAME="+kafkaUser.UserName)
	}

	if kafkaUser.Name != "" {
		parameters = append(parameters, "NAME="+kafkaUser.Name)
	}

	err := applyNsResourceFromTemplateByAdmin(kafkaUser.Namespace, parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// deletes kafkaUser
func (k *KafkaUser) deleteKafkaUser() {
	e2e.Logf("Deleting Kafka user")
	err := deleteDynamicResource("kafkauser", k.UserName, k.Namespace)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// deletes kafkaTopic
func (kafkaTopic *KafkaTopic) deleteKafkaTopic() {
	e2e.Logf("Deleting Kafka topic")
	err := deleteDynamicResource("kafkatopic", kafkaTopic.TopicName, kafkaTopic.Namespace)
	o.Expect(err).NotTo(o.HaveOccurred())
	topic := Resource{"kafkatopic", kafkaTopic.TopicName, kafkaTopic.Namespace}
	err = topic.WaitUntilResourceIsGone()
	o.Expect(err).NotTo(o.HaveOccurred())
}

// deletes kafkaNodePool
func (kafkaNodePool *KafkaNodePool) deleteKafkaNodePool() {
	e2e.Logf("Deleting KafkaNodePool")
	err := deleteDynamicResource("kafkanodepool", kafkaNodePool.NodePoolName, kafkaNodePool.Namespace)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// deletes kafka
func (kafka *Kafka) deleteKafka() {
	e2e.Logf("Deleting Kafka")
	err := deleteDynamicResource("kafka", kafka.Name, kafka.Namespace)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// Poll to wait for kafka to be ready
func waitForKafkaReady(kafkaName string, kafkaNS string) {
	err := wait.PollUntilContextTimeout(context.Background(), 6*time.Second, 360*time.Second, false, func(ctx context.Context) (done bool, err error) {
		obj, getErr := getDynamicResource("kafka", kafkaName, kafkaNS)
		if getErr != nil {
			e2e.Logf("kafka status ready error: %v", getErr)
			return false, nil
		}
		condStatus, _, _ := getConditionStatus(obj, "Ready")
		return condStatus == "True", nil
	})
	assertWaitPollNoErr(err, fmt.Sprintf("resource kafka/%s did not appear", kafkaName))
}

// Poll to wait for kafka Topic to be ready
func waitForKafkaTopicReady(kafkaTopicName string, kafkaTopicNS string) {
	err := wait.PollUntilContextTimeout(context.Background(), 6*time.Second, 360*time.Second, false, func(ctx context.Context) (done bool, err error) {
		obj, getErr := getDynamicResource("kafkatopic", kafkaTopicName, kafkaTopicNS)
		if getErr != nil {
			e2e.Logf("kafka Topic status ready error: %v", getErr)
			return false, nil
		}
		condStatus, _, _ := getConditionStatus(obj, "Ready")
		e2e.Logf("Waiting for kafka topic status %s", condStatus)
		return condStatus == "True", nil
	})
	assertWaitPollNoErr(err, fmt.Sprintf("resource kafkaTopic/%s did not appear", kafkaTopicName))
}
