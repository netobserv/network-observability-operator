package e2etests

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	filePath "path/filepath"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = g.Describe("[sig-netobserv] Network_Observability", func() {

	defer g.GinkgoRecover()
	var (
		namespace string

		OtelNS = OperatorNamespace{
			Name:              "openshift-opentelemetry-operator",
			NamespaceTemplate: filePath.Join(subscriptionDir, "namespace.yaml"),
		}

		OTELSource = CatalogSourceObjects{"stable", "redhat-operators", "openshift-marketplace"}

		OTEL = SubscriptionObjects{
			OperatorName:  "opentelemetry-operator",
			Namespace:     OtelNS.Name,
			PackageName:   "opentelemetry-product",
			Subscription:  filePath.Join(subscriptionDir, "sub-template.yaml"),
			OperatorGroup: filePath.Join(subscriptionDir, "allnamespace-og.yaml"),
			CatalogSource: &OTELSource,
		}
	)

	g.BeforeEach(func() {
		oc := NewCLI()
		namespace = oc.Namespace()
	})

	g.It("Author:aramesha-High-64156-Verify IPFIX-exporter [Serial]", func() {
		clusterArch, err := getNodeArchitecture()
		o.Expect(err).NotTo(o.HaveOccurred())
		if !strings.Contains(clusterArch, "amd64") {
			g.Skip("IPFIX collector image only supports amd64 architecture. Skip this test!")
		}

		g.By("Create IPFIX namespace")
		ipfixCollectorTemplatePath := filePath.Join(baseDir, "exporters", "ipfix-collector.yaml")
		IPFIXns := "ipfix"
		defer deleteNamespace(IPFIXns)
		err = createNamespace(IPFIXns)
		o.Expect(err).NotTo(o.HaveOccurred())
		_ = setNamespacePrivileged(IPFIXns)

		g.By("Deploy IPFIX collector")
		createResourceFromFile(IPFIXns, ipfixCollectorTemplatePath)
		WaitForPodsReadyWithLabel(IPFIXns, "app=ipfix-collector")

		g.By("Wait for IPFIX collector TCP listener to initialize")
		time.Sleep(10 * time.Second)

		IPFIXconfig := map[string]any{
			"ipfix": map[string]any{
				"targetHost":   "ipfix-collector.ipfix.svc.cluster.local",
				"targetPort":   2055,
				"transport":    "TCP",
				"enterpriseID": 0},
			"type": "IPFIX",
		}

		config, err := json.Marshal(IPFIXconfig)
		o.Expect(err).ToNot(o.HaveOccurred())
		IPFIXexporter := string(config)
		additionalNamespaces := fmt.Sprintf("\"%s\"", IPFIXns)
		samplingValue := 3

		g.By("Deploy FlowCollector with IPFIX exporter and sampling")
		flow := Flowcollector{
			Namespace:                         namespace,
			Template:                          flowFixturePath,
			LokiEnable:                        "false",
			InstallDemoLoki:                   "false",
			Exporters:                         []string{IPFIXexporter},
			NetworkPolicyAdditionalNamespaces: []string{additionalNamespaces},
			Sampling:                          strconv.Itoa(samplingValue),
		}

		defer func() { _ = flow.DeleteFlowcollector() }()
		flow.CreateFlowcollector()

		g.By("Verify flowcollector is deployed with IPFIX exporter")
		fcObj, err := getDynamicResource("flowcollector", "cluster", "")
		o.Expect(err).ToNot(o.HaveOccurred())
		exporters, _, _ := unstructured.NestedSlice(fcObj.Object, "spec", "exporters")
		o.Expect(exporters).NotTo(o.BeEmpty())
		exporter0 := exporters[0].(map[string]interface{})
		o.Expect(exporter0["type"]).To(o.Equal("IPFIX"))

		g.By("Get IPFIX collector pod")
		collectorPod, err := getPodNameWithLabel(IPFIXns, "app=ipfix-collector")
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Wait for IPFIX flows to be collected")
		time.Sleep(60 * time.Second)

		g.By("Retrieve and parse IPFIX flow records from collector API")
		flowRecords, err := getIPFIXFlowRecordsFromAPI(IPFIXns, collectorPod)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "No IPFIX flow records found in collector")

		g.By("Verify all IPFIX fields are present and valid")
		for _, record := range flowRecords {
			record.Flowlog.verifyIPFIXFields()
		}

		g.By("Verify sampling value matches FlowCollector configuration (NETOBSERV-2706)")
		for _, record := range flowRecords {
			o.Expect(record.Flowlog.Sampling).Should(o.BeNumerically("==", samplingValue),
				fmt.Sprintf("Expected Sampling=%d, got %d", samplingValue, record.Flowlog.Sampling))
		}
	})

	g.It("Author:memodi-High-74977-Verify OTEL exporter with TLS [Serial]", func() {
		// don't delete the OTEL Operator at the end of the test
		g.By("Subscribe to OTEL Operator")
		OtelNS.DeployOperatorNamespace()
		OTEL.SubscribeOperator()
		WaitForPodsReadyWithLabel(OTEL.Namespace, "app.kubernetes.io/name="+OTEL.OperatorName)
		OTELStatus, err := CheckOperatorStatus(OTEL.Namespace, OTEL.PackageName)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect((OTELStatus)).To(o.BeTrue())

		g.By("Create OTEL Collector with TLS enabled")
		otelCollectorTemplatePath := filePath.Join(baseDir, "exporters", "otel-collector-tls.yaml")
		otlpEndpoint := 4317
		promEndpoint := "8889"
		collectorname := "otel"
		err = applyNsResourceFromTemplateByAdmin(namespace, "-f", otelCollectorTemplatePath, "-p", "NAME="+collectorname, "OTLP_GRPC_ENDPOINT="+strconv.Itoa(otlpEndpoint), "OTLP_PROM_PORT="+promEndpoint)
		o.Expect(err).NotTo(o.HaveOccurred())
		otelPodLabel := "app.kubernetes.io/component=opentelemetry-collector"
		defer func() {
			_ = deleteDynamicResource("opentelemetrycollector", collectorname, namespace)
			_ = k8sClient.CoreV1().Services(namespace).Delete(context.Background(), collectorname+"-collector", metav1.DeleteOptions{})
			_ = k8sClient.CoreV1().ConfigMaps(namespace).Delete(context.Background(), "service-ca", metav1.DeleteOptions{})
		}()
		WaitForPodsReadyWithLabel(namespace, otelPodLabel)

		g.By("Wait for service-ca configmap to be injected with CA bundle")
		waitForConfigMapDataInjection(namespace, "service-ca", "service-ca.crt")

		targetHost := fmt.Sprintf("%s-collector.%s.svc", collectorname, namespace)
		otel_config := map[string]any{
			"openTelemetry": map[string]any{
				"logs": map[string]bool{"enable": true},
				"metrics": map[string]any{"enable": true,
					"pushTimeInterval": "20s"},
				"targetHost": targetHost,
				"targetPort": otlpEndpoint,
				"protocol":   "grpc",
				"tls": map[string]any{
					"enable":             true,
					"insecureSkipVerify": false,
					"caCert": map[string]any{
						"type":     "configmap",
						"name":     "service-ca",
						"certFile": "service-ca.crt",
					},
				},
			},
			"type": "OpenTelemetry",
		}
		config, err := json.Marshal(otel_config)
		o.Expect(err).NotTo(o.HaveOccurred())
		config_str := string(config)

		g.By("Deploy FlowCollector with OTEL TLS exporter and Loki disabled")
		flow := Flowcollector{
			Namespace:       namespace,
			Template:        flowFixturePath,
			LokiEnable:      "false",
			InstallDemoLoki: "false",
			Exporters:       []string{config_str},
		}

		defer func() { _ = flow.DeleteFlowcollector() }()
		flow.CreateFlowcollector()

		// wait for 60 seconds to ensure we collected enough logs to grep from
		time.Sleep(60 * time.Second)

		g.By("Verify OTEL collector is receiving TLS-encrypted flows")
		otelCollectorPod, err := getPodNameWithLabel(namespace, otelPodLabel)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Verify OTEL flowlogs are seen in collector pod logs")
		textToExist := "Attributes:"
		textToNotExist := "INVALID"

		podLogs, err := getPodLogs(namespace, otelCollectorPod)
		o.Expect(err).ToNot(o.HaveOccurred())

		o.Expect(podLogs).To(o.ContainSubstring(textToExist))
		o.Expect(podLogs).ToNot(o.ContainSubstring(textToNotExist))

		g.By("Verify OTEL prometheus has metrics")
		// Get the service IP for the service with label operator.opentelemetry.io/collector-service-type=base
		svcIP, err := getServiceClusterIP(namespace, "operator.opentelemetry.io/collector-service-type=base")
		o.Expect(err).NotTo(o.HaveOccurred())

		// Get one of the flowlogs-pipeline pods
		flowlogsPipelinePod, err := getPodNameWithLabel(namespace, "app=flowlogs-pipeline")
		o.Expect(err).NotTo(o.HaveOccurred())

		// Use the flowlogs-pipeline pod to curl the metrics endpoint of the otel collector service
		command := fmt.Sprintf("curl -s http://%s:%s/metrics | grep 'netobserv_workload_flows_total{' | head -1 | awk '{print $2}'", svcIP, promEndpoint)
		cmd := []string{"/bin/sh", "-c", command}
		count, err := execInPod(namespace, flowlogsPipelinePod, cmd)
		o.Expect(err).ToNot(o.HaveOccurred())
		nCount, err := strconv.Atoi(strings.Trim(count, "\n"))
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(nCount).To(o.BeNumerically(">", 0))
	})
})
