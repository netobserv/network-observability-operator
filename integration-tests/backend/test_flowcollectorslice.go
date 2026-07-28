package e2etests

import (
	"encoding/json"
	"time"

	filePath "path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"
	e2eoutput "k8s.io/kubernetes/test/e2e/framework/pod/output"
)

var _ = g.Describe("[sig-netobserv] Network_Observability", func() {

	defer g.GinkgoRecover()
	var (
		oc                   = compat_otp.NewCLI("netobserv", compat_otp.KubeConfigPath())
		flowSliceFixturePath = filePath.Join(baseDir, "flowcollectorSlice_v1alpha1_template.yaml")
		namespace            string
		ipStackType          string
	)

	g.BeforeEach(func() {
		namespace = oc.Namespace()
		ipStackType = checkIPStackType(oc)
	})

	g.It("Author:aramesha-Critical-86388-Verify flowCollectorSlice collectionMode: AlwaysCollect [Serial]", func() {
		pingPodsTemplate := filePath.Join(baseDir, "test-ping-pods_template.yaml")
		testPingPodsTemplate := TestPingPodsTemplate{
			ServerNS:    "test-ping-server-86388-always",
			ClientNS:    "test-ping-client-86388-always",
			PingTargets: "192.168.1.0 8.8.8.8",
			Template:    pingPodsTemplate,
		}

		subnetLabelsConfig := []map[string]interface{}{
			{
				"name": "external-api",
				"cidrs": []string{
					"8.8.8.8/32",
					"1.1.1.1/32",
				},
			},
			{
				"name": "internal-service",
				"cidrs": []string{
					"192.168.1.0/24",
				},
			},
		}

		config, err := json.Marshal(subnetLabelsConfig)
		o.Expect(err).ToNot(o.HaveOccurred())
		subnetLabels := string(config)

		g.By("Deploy FlowCollectorSlice")
		startTime := time.Now()
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testPingPodsTemplate.ClientNS)
		oc.CreateSpecifiedNamespaceAsAdmin(testPingPodsTemplate.ClientNS)
		flowSlice := FlowcollectorSlice{
			Name:         "subnet-label-slice",
			Namespace:    testPingPodsTemplate.ClientNS,
			SubnetLabels: subnetLabels,
			Template:     flowSliceFixturePath,
		}

		defer func() { _ = flowSlice.DeleteFlowcollectorSlice(oc) }()
		flowSlice.CreateFlowcollectorSlice(oc)

		g.By("Deploy FlowCollector with SlicesEnabled in AlwaysCollect mode")
		flow := Flowcollector{
			Namespace:      namespace,
			CollectionMode: "AlwaysCollect",
			SlicesEnable:   "true",
			Template:       flowFixturePath,
		}

		defer func() { _ = flow.DeleteFlowcollector(oc) }()
		flow.CreateFlowcollector(oc)
		flowSlice.WaitForFlowcollectorSliceReady(oc)

		g.By("Deploy test ping server and client pods")
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testPingPodsTemplate.ServerNS)
		err = testPingPodsTemplate.createPingPods(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		compat_otp.AssertAllPodsToBeReady(oc, testPingPodsTemplate.ServerNS)
		compat_otp.AssertAllPodsToBeReady(oc, testPingPodsTemplate.ClientNS)

		g.By("Wait for a min before logs gets collected and written to loki")
		time.Sleep(60 * time.Second)

		// Scenario1: Internal IP subnetLabel
		g.By("Verify flows with internal-service subnetLabel")
		lokilabels := Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testPingPodsTemplate.ClientNS,
		}
		parameters := []string{"DstAddr=\"192.168.1.0\""}

		flowRecords, err := lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime, parameters...)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of flows from client NS to internal-service > 0")
		for _, r := range flowRecords {
			o.Expect(r.Flowlog.DstSubnetLabel).Should(o.ContainSubstring("internal-service"))
		}

		// Scenario2: External IP subnetLabel
		g.By("Verify flows with external-api subnetLabel")
		parameters = []string{"DstAddr=\"8.8.8.8\""}

		flowRecords, err = lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime, parameters...)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of flows from client NS to external-api > 0")
		for _, r := range flowRecords {
			o.Expect(r.Flowlog.DstSubnetLabel).Should(o.ContainSubstring("external-api"))
		}

		// Scenario3: Flows are collected from namespaces without Slice deployed too
		g.By("Verify flows having no subnet label")
		lokilabels = Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testPingPodsTemplate.ServerNS,
		}

		flowRecords, err = lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime, parameters...)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of flows from server NS > 0")
		for _, r := range flowRecords {
			o.Expect(r.Flowlog.DstSubnetLabel).Should(o.ContainSubstring("external-api"))
		}
	})

	g.It("Author:aramesha-Critical-86388-Verify flowCollectorSlice collectionMode: AllowList [Serial]", func() {

		pingPodsTemplate := filePath.Join(baseDir, "test-ping-pods_template.yaml")
		testPingPodsTemplate := TestPingPodsTemplate{
			ServerNS:    "test-ping-server-86388-allowlist",
			ClientNS:    "test-ping-client-86388-allowlist",
			PingTargets: "8.8.8.8",
			Template:    pingPodsTemplate,
		}

		g.By("Deploy FlowCollectorSlice")
		startTime := time.Now()
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testPingPodsTemplate.ClientNS)
		oc.CreateSpecifiedNamespaceAsAdmin(testPingPodsTemplate.ClientNS)
		flowSlice := FlowcollectorSlice{
			Name:      "namespace-slice",
			Namespace: testPingPodsTemplate.ClientNS,
			Sampling:  "3",
			Template:  flowSliceFixturePath,
		}

		defer func() { _ = flowSlice.DeleteFlowcollectorSlice(oc) }()
		flowSlice.CreateFlowcollectorSlice(oc)

		g.By("Deploy FlowCollector with Slices enabled in AllowList mode")
		flow := Flowcollector{
			Namespace:       namespace,
			CollectionMode:  "AllowList",
			SlicesEnable:    "true",
			NamespacesAllow: []string{"\"/openshift-.*/\""},
			Template:        flowFixturePath,
		}

		defer func() { _ = flow.DeleteFlowcollector(oc) }()
		flow.CreateFlowcollector(oc)
		flowSlice.WaitForFlowcollectorSliceReady(oc)

		g.By("Deploy test ping server and client pods")
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testPingPodsTemplate.ServerNS)
		err := testPingPodsTemplate.createPingPods(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		compat_otp.AssertAllPodsToBeReady(oc, testPingPodsTemplate.ServerNS)
		compat_otp.AssertAllPodsToBeReady(oc, testPingPodsTemplate.ClientNS)

		g.By("Wait for a min before logs gets collected and written to loki")
		time.Sleep(60 * time.Second)

		// Scenario1: Ping from namespace where flowCollectorSlice is deployed
		g.By("Verify flows from client NS")
		lokilabels := Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testPingPodsTemplate.ClientNS,
		}
		parameters := []string{"DstAddr=\"8.8.8.8\""}

		flowRecords, err := lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime, parameters...)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of flows from client NS > 0")
		for _, r := range flowRecords {
			o.Expect(r.Flowlog.Sampling).Should(o.BeNumerically("==", 3))
		}

		// Scenario2: Ping from namespace where flowCollectorSlice is NOT deployed
		g.By("Verify NO flows are seen from server NS")
		lokilabels = Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testPingPodsTemplate.ServerNS,
			AllowEmpty:      true,
		}

		flowRecords, err = lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime, parameters...)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically("==", 0), "expected number of flows from server NS = 0")

		// Scenario3: Flows from namespace in allowedNamespaces section of flowcollector
		g.By("Verify flows are seen to openshift-dns")
		lokilabels = Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: "openshift-dns",
		}

		flowRecords, err = lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of flows from openshift-dns NS > 0")

		// Scenario4: Flows between namespaces with one in allowedNamespaces section should still be collected
		g.By("Verify flows between namespaces")
		startTime = time.Now()
		serverPodIP, _ := getPodIP(oc, testPingPodsTemplate.ServerNS, "ping-server", ipStackType)

		_, err = e2eoutput.RunHostCmd(testPingPodsTemplate.ClientNS, "ping-client", "ping -w 300 "+serverPodIP)
		o.Expect(err).NotTo(o.HaveOccurred())
		time.Sleep(60 * time.Second)

		lokilabels = Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testPingPodsTemplate.ClientNS,
			DstK8SNamespace: testPingPodsTemplate.ServerNS,
		}

		flowRecords, err = lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of flows between test namespaces > 0")
	})
})
