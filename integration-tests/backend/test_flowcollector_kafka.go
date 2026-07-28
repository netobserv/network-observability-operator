package e2etests

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	filePath "path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

var _ = g.Describe("[sig-netobserv] Network_Observability with Kafka", g.Ordered, g.ContinueOnFailure, func() {
	defer g.GinkgoRecover()

	var (
		oc        = compat_otp.NewCLI("netobserv", compat_otp.KubeConfigPath())
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
		kafkaAddress                                = fmt.Sprintf("%s-kafka-bootstrap.%s:9093", kafkaClusterName, kafkaNs)
		additionalNamespaces                        = fmt.Sprintf("\"%s\"", kafkaNs)
	)

	g.BeforeAll(func() {
		oc.SetNamespace(netobservNS)

		// Set up Kafka paths
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

		kafkaChannel, err := getOperatorChannel(oc, kafkaSource.SourceName, amq.PackageName)
		if err != nil || kafkaChannel == "" {
			g.Skip("Kafka channel not found, skip this case")
		}

		// check if AMQ Streams Operator is already present
		AMQexisting, err = CheckOperatorStatus(oc, amq.Namespace, amq.PackageName)
		o.Expect(err).NotTo(o.HaveOccurred())
		if !AMQexisting {
			ensureOperatorDeployed(oc, amq, kafkaSource, "name="+amq.OperatorName)
			// before creating kafka, check the existence of crd kafkas.kafka.strimzi.io
			checkResource(oc, true, true, "kafka.strimzi.io", []string{"crd", "kafkas.kafka.strimzi.io", "-ojsonpath={.spec.group}"})
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
		oc.CreateSpecifiedNamespaceAsAdmin(kafkaNs)
		kafkaMetrics.deployKafkaMetrics(oc)
		kafkaNodePool.deployKafkaNodePool(oc)
		kafka.deployKafka(oc)
		kafkaTopic.deployKafkaTopic(oc)
		kafkaUser.deployKafkaUser(oc)

		g.By("Check if Kafka and Kafka topic are ready")
		// wait for KafkaNodePool, Kafka and KafkaTopic to be ready
		WaitForPodsReadyWithLabel(oc, kafka.Namespace, "strimzi.io/pool-name=kafka-pool")
		waitForKafkaReady(oc, kafka.Name, kafka.Namespace)
		waitForKafkaTopicReady(oc, kafkaTopic.TopicName, kafkaTopic.Namespace)
	})

	g.AfterAll(func() {
		// Clean up Kafka resources in reverse order
		kafkaUser.deleteKafkaUser(oc)
		kafkaTopic.deleteKafkaTopic(oc)
		kafka.deleteKafka(oc)
		kafkaNodePool.deleteKafkaNodePool(oc)
		oc.DeleteSpecifiedNamespaceAsAdmin(kafkaNs)

		// Uninstall AMQ operator if it was not pre-existing
		if !AMQexisting {
			amq.uninstallOperator(oc)
		}
	})

	g.BeforeEach(func() {
		namespace = oc.Namespace()
	})

	g.It("Author:aramesha-NonPreRelease-Longduration-Critical-56362-High-53597-High-56326-High-64880-High-75340-Verify network flows are captured with Kafka with TLS [Serial][Slow]", func() {

		g.By("Deploy FlowCollector with Kafka TLS")
		flow := Flowcollector{
			Namespace:                         namespace,
			DeploymentModel:                   "Kafka",
			Template:                          flowFixturePath,
			KafkaAddress:                      kafkaAddress,
			KafkaTLSEnable:                    "true",
			KafkaNamespace:                    kafkaNs,
			NetworkPolicyAdditionalNamespaces: []string{additionalNamespaces},
		}

		defer func() { _ = flow.DeleteFlowcollector(oc) }()
		flow.CreateFlowcollector(oc)

		g.By("Ensure secrets are synced")
		// ensure certs are synced to privileged NS
		secrets, err := getSecrets(oc, namespace+"-privileged")
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(secrets).To(o.And(o.ContainSubstring(kafkaUser.UserName), o.ContainSubstring(kafka.Name+"-cluster-ca-cert")))

		g.By("Verify prometheus is able to scrape metrics for FLP-Kafka")
		flpPrpmSM := "flowlogs-pipeline-transformer-monitor"
		tlsScheme, err := getMetricsScheme(oc, flpPrpmSM, flow.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		tlsScheme = strings.Trim(tlsScheme, "'")
		o.Expect(tlsScheme).To(o.Equal("https"))

		serverName, err := getMetricsServerName(oc, flpPrpmSM, flow.Namespace)
		serverName = strings.Trim(serverName, "'")
		o.Expect(err).NotTo(o.HaveOccurred())
		flpPromSA := "flowlogs-pipeline-transformer-prom"
		expectedServerName := fmt.Sprintf("%s.%s.svc", flpPromSA, namespace)
		o.Expect(serverName).To(o.Equal(expectedServerName))

		// verify FLP metrics are being populated with Kafka
		// Sleep before making any metrics request
		g.By("Verify prometheus is able to scrape FLP metrics")
		time.Sleep(30 * time.Second)
		verifyFLPMetrics(oc)

		// verify logs
		g.By("Wait for a min before logs gets collected and written to loki")
		startTime := time.Now()
		time.Sleep(60 * time.Second)

		g.By("Get flowlogs from loki")
		err = verifyMonolithicLokilogsTime(oc, flow.Namespace, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())
	})

	g.It("Author:aramesha-NonPreRelease-Longduration-High-57397-High-65116-Low-89197-Verify network-flows export with Kafka and netobserv installation without Loki[Serial]", func() {
		g.By("Deploy kafka Topic for export")
		// deploy kafka topic for export
		kafkaTopic2 := KafkaTopic{
			TopicName: "network-flows-export",
			Name:      kafka.Name,
			Namespace: kafkaNs,
			Template:  kafkaTopicPath,
		}

		defer kafkaTopic2.deleteKafkaTopic(oc)
		kafkaTopic2.deployKafkaTopic(oc)
		waitForKafkaTopicReady(oc, kafkaTopic2.TopicName, kafkaTopic2.Namespace)

		kafkaExporterConfig := map[string]interface{}{
			"kafka": map[string]interface{}{
				"address": kafkaAddress,
				"tls": map[string]interface{}{
					"caCert": map[string]interface{}{
						"certFile":  "ca.crt",
						"name":      "kafka-cluster-cluster-ca-cert",
						"namespace": kafkaNs,
						"type":      "secret"},
					"enable":             true,
					"insecureSkipVerify": false,
					"userCert": map[string]interface{}{
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

		// Use random compression to test different options across periodic runs
		validCompressions := []string{"gzip", "snappy", "lz4", "zstd"}
		randomCompression := validCompressions[g.GinkgoRandomSeed()%int64(len(validCompressions))]
		e2e.Logf("Using Kafka compression: %s (seed: %d)", randomCompression, g.GinkgoRandomSeed())

		g.By("Deploy FlowCollector with Kafka TLS")
		flow := Flowcollector{
			Namespace:                         namespace,
			DeploymentModel:                   "Kafka",
			Template:                          flowFixturePath,
			KafkaAddress:                      kafkaAddress,
			KafkaTLSEnable:                    "true",
			KafkaNamespace:                    kafkaNs,
			KafkaCompression:                  randomCompression,
			Exporters:                         []string{kafkaConfig},
			NetworkPolicyAdditionalNamespaces: []string{additionalNamespaces},
		}

		defer func() { _ = flow.DeleteFlowcollector(oc) }()
		flow.CreateFlowcollector(oc)

		// Scenario1: Verify flows are exported with Kafka DeploymentModel and with Loki enabled
		g.By("Verify flowcollector is deployed with KAFKA exporter")
		exporterType, err := oc.AsAdmin().Run("get").Args("flowcollector", "cluster", "-o", "jsonpath='{.spec.exporters[0].type}'").Output()
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(exporterType).To(o.Equal(`'Kafka'`))

		g.By("Ensure flows are observed, all pods are running and secrets are synced and plugin pod is deployed")
		// ensure certs are synced to privileged NS
		secrets, err := getSecrets(oc, namespace+"-privileged")
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(secrets).To(o.And(o.ContainSubstring(kafkaUser.UserName), o.ContainSubstring(kafka.Name+"-cluster-ca-cert")))

		// verify logs
		g.By("Wait for a min before logs gets collected and written to loki")
		startTime := time.Now()
		time.Sleep(60 * time.Second)

		g.By("Get flowlogs from loki")
		err = verifyMonolithicLokilogsTime(oc, flow.Namespace, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Deploy Kafka consumer pod")
		// using amq-streams/kafka-34-rhel8:2.5.2 version. Update if imagePull issues are observed
		consumerTemplate := filePath.Join(kafkaDir, "topic-consumer-tls.yaml")
		consumer := Resource{"job", kafkaTopic2.TopicName + "-consumer", kafkaNs}
		defer func() { _ = consumer.clear(oc) }()
		err = consumer.applyFromTemplate(oc, "-n", consumer.Namespace, "-f", consumerTemplate, "-p", "NAME="+consumer.Name, "NAMESPACE="+consumer.Namespace, "KAFKA_TOPIC="+kafkaTopic2.TopicName, "CLUSTER_NAME="+kafka.Name, "KAFKA_USER="+kafkaUser.UserName)
		o.Expect(err).NotTo(o.HaveOccurred())

		WaitForPodsReadyWithLabel(oc, consumer.Namespace, "job-name="+consumer.Name)

		consumerPodName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-n", consumer.Namespace, "-l", "job-name="+consumer.Name, "-o=jsonpath={.items[0].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Verify Kafka consumer pod logs")
		podLogs, err := compat_otp.WaitAndGetSpecificPodLogs(oc, consumer.Namespace, "", consumerPodName, `'{"AgentIP":'`)
		compat_otp.AssertWaitPollNoErr(err, "Did not get log for the pod with job-name=network-flows-export-consumer label")
		verifyFlowRecordFromLogs(podLogs)

		g.By("Verify NetObserv can be installed without Loki")
		_ = flow.DeleteFlowcollector(oc)
		// Ensure FLP and eBPF pods are deleted
		checkPodDeleted(oc, namespace, "app=flowlogs-pipeline", "flowlogs-pipeline")
		checkPodDeleted(oc, namespace+"-privileged", "app=netobserv-ebpf-agent", "netobserv-ebpf-agent")

		flow.DeploymentModel = "Service"
		flow.LokiEnable = "false"
		flow.InstallDemoLoki = "false"
		flow.CreateFlowcollector(oc)

		g.By("Verify new Kafka consumer records after FlowCollector recreation")
		podLogs, err = waitForNewPodLogs(oc, consumer.Namespace, consumerPodName, `{"AgentIP":`)
		compat_otp.AssertWaitPollNoErr(err, "No new records appeared in Kafka consumer after FlowCollector recreation")
		verifyFlowRecordFromLogs(podLogs)

		g.By("Verify console plugin pod is not deployed when its disabled in flowcollector")
		_ = flow.DeleteFlowcollector(oc)
		// Ensure FLP and eBPF pods are deleted
		checkPodDeleted(oc, namespace, "app=flowlogs-pipeline", "flowlogs-pipeline")
		checkPodDeleted(oc, namespace+"-privileged", "app=netobserv-ebpf-agent", "netobserv-ebpf-agent")

		flow.PluginEnable = "false"
		flow.CreateFlowcollector(oc)

		// Scenario3: Verify all pods except plugin pod are present with only Plugin disabled in flowcollector
		g.By("Ensure all pods except consolePlugin pod are deployed")
		consolePod, err := compat_otp.GetAllPodsWithLabel(oc, namespace, "app=netobserv-plugin")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(consolePod)).To(o.Equal(0))
	})

	// Add future NetObserv + demoLoki + Kafka test-cases here
})
