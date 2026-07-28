package e2etests

import (
	"encoding/json"
	"fmt"
	filePath "path/filepath"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

var _ = g.Describe("[sig-netobserv] Network_Observability with Kafka", g.Ordered, g.ContinueOnFailure, func() {
	defer g.GinkgoRecover()

	var (
		namespace string

		// Kafka-specific variables
		kafkaDir, kafkaTopicPath, kafkaNodePoolPath string
		AMQexisting                                 = false
		amq                                         SubscriptionObjects
		kafkaMetrics                                KafkaMetrics
		kafka                                       Kafka
		kafkaTopic                                  KafkaTopic
		kafkaNodePool                               KafkaNodePool
		kafkaUser                                   KafkaUser
		kafkaNs                                     = "netobserv-kafka"
		kafkaClusterName                            = "kafka-cluster"
		kafkaAddress                                string
		additionalNamespaces                        string
	)

	g.BeforeAll(func() {
		kafkaAddress = fmt.Sprintf("%s-kafka-bootstrap.%s:9093", kafkaClusterName, kafkaNs)
		additionalNamespaces = fmt.Sprintf("\"%s\"", kafkaNs)

		kafkaDir, _ = filePath.Abs("testdata/kafka")
		kafkaNodePoolPath = filePath.Join(kafkaDir, "kafka-node-pool.yaml")
		kafkaTopicPath = filePath.Join(kafkaDir, "kafka-topic.yaml")
		kafkaTLSPath := filePath.Join(kafkaDir, "kafka-tls.yaml")
		kafkaMetricsPath := filePath.Join(kafkaDir, "kafka-metrics-config.yaml")
		kafkaUserPath := filePath.Join(kafkaDir, "kafka-user.yaml")

		g.By("Subscribe to AMQ operator")
		kafkaSource := CatalogSourceObjects{"stable", "redhat-operators", "openshift-marketplace"}
		amq = SubscriptionObjects{
			OperatorName:  "amq-streams-cluster-operator",
			Namespace:     "openshift-operators",
			PackageName:   "amq-streams",
			Subscription:  filePath.Join(subscriptionDir, "sub-template.yaml"),
			CatalogSource: &kafkaSource,
		}

		kafkaChannel, err := getOperatorChannel(kafkaSource.SourceName, amq.PackageName)
		if err != nil || kafkaChannel == "" {
			g.Skip("Kafka channel not found, skip this case")
		}

		AMQexisting, err = CheckOperatorStatus(amq.Namespace, amq.PackageName)
		o.Expect(err).NotTo(o.HaveOccurred())
		if !AMQexisting {
			ensureOperatorDeployed(amq, kafkaSource, "name="+amq.OperatorName)
			checkResource(true, true, "kafka.strimzi.io", "crd", "kafkas.kafka.strimzi.io", ".spec.group")
		}

		kafkaMetrics = KafkaMetrics{
			Namespace: kafkaNs,
			Template:  kafkaMetricsPath,
		}

		kafka = Kafka{
			Name:      kafkaClusterName,
			Namespace: kafkaNs,
			Template:  kafkaTLSPath,
		}

		kafkaNodePool = KafkaNodePool{
			NodePoolName: "kafka-pool",
			Namespace:    kafkaNs,
			Name:         kafka.Name,
			Template:     kafkaNodePoolPath,
		}

		kafkaTopic = KafkaTopic{
			TopicName: "network-flows",
			Name:      kafka.Name,
			Namespace: kafkaNs,
			Template:  kafkaTopicPath,
		}

		kafkaUser = KafkaUser{
			UserName:  "flp-kafka",
			Name:      kafka.Name,
			Namespace: kafkaNs,
			Template:  kafkaUserPath,
		}

		g.By("Deploy Kafka with TLS")
		err = createNamespace(kafkaNs)
		o.Expect(err).NotTo(o.HaveOccurred())
		kafkaMetrics.deployKafkaMetrics()
		kafkaNodePool.deployKafkaNodePool()
		kafka.deployKafka()
		kafkaTopic.deployKafkaTopic()
		kafkaUser.deployKafkaUser()

		g.By("Check if Kafka and Kafka topic are ready")
		WaitForPodsReadyWithLabel(kafka.Namespace, "strimzi.io/pool-name=kafka-pool")
		waitForKafkaReady(kafka.Name, kafka.Namespace)
		waitForKafkaTopicReady(kafkaTopic.TopicName, kafkaTopic.Namespace)
	})

	g.AfterAll(func() {
		kafkaUser.deleteKafkaUser()
		kafkaTopic.deleteKafkaTopic()
		kafka.deleteKafka()
		kafkaNodePool.deleteKafkaNodePool()
		deleteNamespace(kafkaNs)
		if !AMQexisting {
			amq.uninstallOperator()
		}
	})

	g.BeforeEach(func() {
		oc = NewCLI()
		namespace = oc.Namespace()
	})

	g.It("Author:aramesha-NonPreRelease-Longduration-Critical-56362-High-53597-High-56326-High-64880-High-75340-Verify network flows are captured with Kafka with TLS [Serial][Slow]", func() {
		g.By("Deploy FlowCollector with Kafka TLS")
		flow := Flowcollector{
			Namespace:                         namespace,
			DeploymentModel:                   "Kafka",
			Template:                          flowFixturePath,
			MonolithicLokiURL:                 fmt.Sprintf("http://loki.%s.svc:3100/", namespace),
			KafkaAddress:                      kafkaAddress,
			KafkaTLSEnable:                    "true",
			KafkaNamespace:                    kafkaNs,
			NetworkPolicyAdditionalNamespaces: []string{additionalNamespaces},
		}

		defer func() { _ = flow.DeleteFlowcollector() }()
		flow.CreateFlowcollector()

		g.By("Ensure secrets are synced")
		secrets, err := getSecrets(namespace + "-privileged")
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(secrets).To(o.And(o.ContainSubstring(kafkaUser.UserName), o.ContainSubstring(kafka.Name+"-cluster-ca-cert")))

		g.By("Verify prometheus is able to scrape metrics for FLP-Kafka")
		flpPrpmSM := "flowlogs-pipeline-transformer-monitor"
		tlsScheme, err := getMetricsScheme(flpPrpmSM, flow.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		tlsScheme = strings.Trim(tlsScheme, "'")
		o.Expect(tlsScheme).To(o.Equal("https"))

		serverName, err := getMetricsServerName(flpPrpmSM, flow.Namespace)
		serverName = strings.Trim(serverName, "'")
		o.Expect(err).NotTo(o.HaveOccurred())
		flpPromSA := "flowlogs-pipeline-transformer-prom"
		expectedServerName := fmt.Sprintf("%s.%s.svc", flpPromSA, namespace)
		o.Expect(serverName).To(o.Equal(expectedServerName))

		g.By("Verify prometheus is able to scrape FLP metrics")
		time.Sleep(30 * time.Second)
		verifyFLPMetrics()

		g.By("Wait for a min before logs gets collected and written to loki")
		startTime := time.Now()
		time.Sleep(60 * time.Second)

		g.By("Get flowlogs from loki")
		err = verifyMonolithicLokilogsTime(flow.MonolithicLokiURL, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())
	})

	g.It("Author:aramesha-NonPreRelease-Longduration-High-57397-High-65116-Low-89197-Verify network-flows export with Kafka and netobserv installation without Loki[Serial]", func() {
		g.By("Deploy kafka Topic for export")
		kafkaTopic2 := KafkaTopic{
			TopicName: "network-flows-export",
			Name:      kafka.Name,
			Namespace: kafkaNs,
			Template:  kafkaTopicPath,
		}

		defer kafkaTopic2.deleteKafkaTopic()
		kafkaTopic2.deployKafkaTopic()
		waitForKafkaTopicReady(kafkaTopic2.TopicName, kafkaTopic2.Namespace)

		kafkaExporterConfig := map[string]any{
			"kafka": map[string]any{
				"address": kafkaAddress,
				"tls": map[string]any{
					"caCert": map[string]any{
						"certFile":  "ca.crt",
						"name":      "kafka-cluster-cluster-ca-cert",
						"namespace": kafkaNs,
						"type":      "secret"},
					"enable":             true,
					"insecureSkipVerify": false,
					"userCert": map[string]any{
						"certFile":  "user.crt",
						"certKey":   "user.key",
						"name":      kafkaUser.UserName,
						"namespace": kafkaNs,
						"type":      "secret"},
				},
				"topic": kafkaTopic2.TopicName},
			"type": "Kafka",
		}

		config, err := json.Marshal(kafkaExporterConfig)
		o.Expect(err).ToNot(o.HaveOccurred())
		kafkaConfig := string(config)

		validCompressions := []string{"gzip", "snappy", "lz4", "zstd"}
		randomCompression := validCompressions[g.GinkgoRandomSeed()%int64(len(validCompressions))]
		e2e.Logf("Using Kafka compression: %s (seed: %d)", randomCompression, g.GinkgoRandomSeed())

		g.By("Deploy FlowCollector with Kafka TLS")
		flow := Flowcollector{
			Namespace:                         namespace,
			DeploymentModel:                   "Kafka",
			Template:                          flowFixturePath,
			MonolithicLokiURL:                 fmt.Sprintf("http://loki.%s.svc:3100/", namespace),
			KafkaAddress:                      kafkaAddress,
			KafkaTLSEnable:                    "true",
			KafkaNamespace:                    kafkaNs,
			KafkaCompression:                  randomCompression,
			Exporters:                         []string{kafkaConfig},
			NetworkPolicyAdditionalNamespaces: []string{additionalNamespaces},
		}

		defer func() { _ = flow.DeleteFlowcollector() }()
		flow.CreateFlowcollector()

		g.By("Verify flowcollector is deployed with KAFKA exporter")
		fcObj, err := getDynamicResource("flowcollector", "cluster", "")
		o.Expect(err).ToNot(o.HaveOccurred())
		exporters, _, _ := unstructured.NestedSlice(fcObj.Object, "spec", "exporters")
		o.Expect(exporters).NotTo(o.BeEmpty())
		exporter0 := exporters[0].(map[string]interface{})
		o.Expect(exporter0["type"]).To(o.Equal("Kafka"))

		g.By("Ensure flows are observed, all pods are running and secrets are synced and plugin pod is deployed")
		secrets, err := getSecrets(namespace + "-privileged")
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(secrets).To(o.And(o.ContainSubstring(kafkaUser.UserName), o.ContainSubstring(kafka.Name+"-cluster-ca-cert")))

		g.By("Wait for a min before logs gets collected and written to loki")
		startTime := time.Now()
		time.Sleep(60 * time.Second)

		g.By("Get flowlogs from loki")
		err = verifyMonolithicLokilogsTime(flow.MonolithicLokiURL, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Deploy Kafka consumer pod")
		consumerTemplate := filePath.Join(kafkaDir, "topic-consumer-tls.yaml")
		consumer := Resource{"job", kafkaTopic2.TopicName + "-consumer", kafkaNs}
		defer func() { _ = consumer.clear() }()
		err = consumer.applyFromTemplate("-n", consumer.Namespace, "-f", consumerTemplate, "-p", "NAME="+consumer.Name, "NAMESPACE="+consumer.Namespace, "KAFKA_TOPIC="+kafkaTopic2.TopicName, "CLUSTER_NAME="+kafka.Name, "KAFKA_USER="+kafkaUser.UserName)
		o.Expect(err).NotTo(o.HaveOccurred())

		WaitForPodsReadyWithLabel(consumer.Namespace, "job-name="+consumer.Name)

		consumerPodName, err := getPodNameWithLabel(consumer.Namespace, "job-name="+consumer.Name)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Verify Kafka consumer pod logs")
		podLogs, err := waitAndGetSpecificPodLogs(consumer.Namespace, consumerPodName, `{"AgentIP":`)
		assertWaitPollNoErr(err, "Did not get log for the pod with job-name=network-flows-export-consumer label")
		verifyFlowRecordFromLogs(podLogs)

		g.By("Verify NetObserv can be installed without Loki")
		_ = flow.DeleteFlowcollector()
		checkPodDeleted(namespace, "app=flowlogs-pipeline", "flowlogs-pipeline")
		checkPodDeleted(namespace+"-privileged", "app=netobserv-ebpf-agent", "netobserv-ebpf-agent")

		flow.DeploymentModel = "Service"
		flow.LokiEnable = "false"
		flow.InstallDemoLoki = "false"
		flow.CreateFlowcollector()

		g.By("Verify new Kafka consumer records after FlowCollector recreation")
		podLogs, err = waitForNewPodLogs(consumer.Namespace, consumerPodName, `{"AgentIP":`)
		assertWaitPollNoErr(err, "No new records appeared in Kafka consumer after FlowCollector recreation")
		verifyFlowRecordFromLogs(podLogs)

		g.By("Verify console plugin pod is not deployed when its disabled in flowcollector")
		_ = flow.DeleteFlowcollector()
		checkPodDeleted(namespace, "app=flowlogs-pipeline", "flowlogs-pipeline")
		checkPodDeleted(namespace+"-privileged", "app=netobserv-ebpf-agent", "netobserv-ebpf-agent")

		flow.PluginEnable = "false"
		flow.CreateFlowcollector()

		g.By("Ensure all pods except consolePlugin pod are deployed")
		consolePod, err := getAllPodsWithLabel(namespace, "app=netobserv-plugin")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(consolePod)).To(o.Equal(0))
	})
})
