package e2etests

import (
	"fmt"
	"strings"
	"time"

	filePath "path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

var _ = g.Describe("[sig-netobserv] Network_Observability with VMs", g.Ordered, g.ContinueOnFailure, func() {
	defer g.GinkgoRecover()

	var (
		oc        = compat_otp.NewCLI("netobserv", compat_otp.KubeConfigPath())
		namespace string

		// virt operator vars
		VOexisting                 = false
		virtOperatorNS             = "openshift-cnv"
		virtualizationDir, _       = filePath.Abs("testdata/virtualization")
		kubevirtHyperconvergedPath = filePath.Join(virtualizationDir, "kubevirt-hyperconverged.yaml")
		virtCatsrc                 = Resource{"catsrc", "redhat-operators", "openshift-marketplace"}
		virtPackageName            = "kubevirt-hyperconverged"
		virtSource                 = CatalogSourceObjects{"stable", virtCatsrc.Name, virtCatsrc.Namespace}
		VO                         = SubscriptionObjects{
			OperatorName:  "kubevirt-hyperconverged",
			Namespace:     virtOperatorNS,
			PackageName:   virtPackageName,
			Subscription:  filePath.Join(subscriptionDir, "sub-template.yaml"),
			OperatorGroup: filePath.Join(subscriptionDir, "singlenamespace-og.yaml"),
			CatalogSource: &virtSource,
		}
		networkingDir = filePath.Join(baseDir, "networking")
		ipStackType   string
	)

	g.BeforeAll(func() {
		oc.SetNamespace(netobservNS)
		ipStackType = checkIPStackType(oc)

		// Architecture check
		clusterArch, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("nodes", "-o=jsonpath={.items[0].status.nodeInfo.architecture}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		if strings.Contains(clusterArch, "ppc64le") {
			g.Skip("Virtualization operator is not supported on ppc64le architecture.")
		}
		isMetal, err := isClusterBareMetal(oc)
		o.Expect(err).ToNot(o.HaveOccurred())
		if !isMetal && !hasMetalWorkerNodes(oc) {
			g.Skip("Cluster does not have baremetal workers.")
		}

		// Deploy virt operator + HyperConverged
		VOexisting, err = CheckOperatorStatus(oc, VO.Namespace, VO.PackageName)
		o.Expect(err).NotTo(o.HaveOccurred())
		if !VOexisting {
			ensureOperatorDeployed(oc, VO, virtSource, "name=virt-operator")
		}
		_, err = oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", kubevirtHyperconvergedPath).Output()
		o.Expect(err).ToNot(o.HaveOccurred())
		waitUntilHyperConvergedReady(oc, "kubevirt-hyperconverged", virtOperatorNS)
		WaitForPodsReadyWithLabel(oc, virtOperatorNS, "app.kubernetes.io/managed-by=virt-operator")
		// Wait for kubemacpool service endpoints to be ready to avoid race condition when creating VMs
		waitForServiceEndpoints(oc, virtOperatorNS, "kubemacpool-service")
	})

	g.AfterAll(func() {
		deleteResource(oc, "hyperconverged", "kubevirt-hyperconverged", virtOperatorNS)
		if !VOexisting {
			VO.uninstallOperator(oc)
		}
	})

	g.BeforeEach(func() {
		namespace = oc.Namespace()
	})

	g.It("Author:aramesha-NonPreRelease-Longduration-High-76537-Verify flow enrichment for VM's secondary interfaces [Disruptive][Slow]", func() {
		var (
			testNS = "test-76537"
			// NAD vars
			networkName   = "l2-network"
			layer2NadPath = filePath.Join(virtualizationDir, "layer2-nad.yaml")
			// VM vars
			testVMStaticIPTemplatePath = filePath.Join(virtualizationDir, "test-vm-static-IP_template.yaml")
		)

		g.By("Deploy Network Attachment Definition in test-76537 namespace")
		defer deleteNamespace(oc, testNS)
		defer deleteResource(oc, "net-attach-def", networkName, testNS)
		_, err := oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", layer2NadPath).Output()
		o.Expect(err).ToNot(o.HaveOccurred())
		// Wait a min for NAD to come up
		time.Sleep(60 * time.Second)
		checkNAD(oc, networkName, testNS)

		g.By("Deploy test VM1")
		testVM1 := TestVMStaticIPTemplate{
			Name:        "test-vm1",
			Namespace:   testNS,
			NetworkName: networkName,
			Mac:         "02:00:00:00:00:01",
			StaticIP:    "10.10.10.15/24",
			Template:    testVMStaticIPTemplatePath,
		}
		defer deleteResource(oc, "vm", testVM1.Name, testNS)
		err = testVM1.createVMStaticIP(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		waitUntilVMReady(oc, testVM1.Name, testVM1.Namespace)

		g.By("Wait for VM1 to get IP assigned")
		vm1Ip, err := waitForVMIPAssignment(oc, testVM1.Name, testVM1.Namespace, 1)
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(vm1Ip).To(o.Equal("10.10.10.15"))

		startTime := time.Now()

		g.By("Deploy test VM2")
		testVM2 := TestVMStaticIPTemplate{
			Name:        "test-vm2",
			Namespace:   testNS,
			NetworkName: networkName,
			Mac:         "02:00:00:00:00:02",
			StaticIP:    "10.10.10.14/24",
			RunCmd:      fmt.Sprintf("[[ping, %s]]", vm1Ip),
			Template:    testVMStaticIPTemplatePath,
		}
		defer deleteResource(oc, "vm", testVM2.Name, testNS)
		err = testVM2.createVMStaticIP(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		waitUntilVMReady(oc, testVM2.Name, testVM2.Namespace)

		g.By("Wait for VM2 to get IP assigned")
		vm2Ip, err := waitForVMIPAssignment(oc, testVM2.Name, testVM2.Namespace, 1)
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(vm2Ip).To(o.Equal("10.10.10.14"))

		g.By("Deploy FlowCollector")
		flow := Flowcollector{
			Namespace:      namespace,
			Template:       flowFixturePath,
			EBPFPrivileged: "true",
		}

		defer func() { _ = flow.DeleteFlowcollector(oc) }()
		flow.CreateFlowcollector(oc)

		g.By("Wait for a min before logs gets collected and written to loki")
		time.Sleep(60 * time.Second)

		lokilabels := Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testNS,
			SrcK8SOwnerName: testVM2.Name,
			DstK8SNamespace: testNS,
			DstK8SOwnerName: testVM1.Name,
		}
		parameters := []string{"DstAddr=\"10.10.10.15\"", "SrcAddr=\"10.10.10.14\""}

		g.By("Verify flows are written to loki")
		flowRecords, err := lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime, parameters...)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of flows written to loki > 0")

		g.By("Verify flow logs are enriched")
		// Get VM1 pod name and node
		vm1PodName, err := compat_otp.GetAllPodsWithLabel(oc, testNS, "vm.kubevirt.io/name="+testVM1.Name)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(vm1PodName).NotTo(o.BeEmpty())
		vm1node, err := compat_otp.GetPodNodeName(oc, testNS, vm1PodName[0])
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(vm1node).NotTo(o.BeEmpty())

		// Get vm2 pod name and node
		vm2PodName, err := compat_otp.GetAllPodsWithLabel(oc, testNS, "vm.kubevirt.io/name="+testVM2.Name)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(vm2PodName).NotTo(o.BeEmpty())
		vm2node, err := compat_otp.GetPodNodeName(oc, testNS, vm2PodName[0])
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(vm2node).NotTo(o.BeEmpty())

		for _, r := range flowRecords {
			o.Expect(r.Flowlog.DstK8SName).Should(o.ContainSubstring(vm1PodName[0]))
			o.Expect(r.Flowlog.SrcK8SName).Should(o.ContainSubstring(vm2PodName[0]))
			o.Expect(r.Flowlog.DstK8SOwnerType).Should(o.ContainSubstring("VirtualMachineInstance"))
			o.Expect(r.Flowlog.SrcK8SOwnerType).Should(o.ContainSubstring("VirtualMachineInstance"))
			o.Expect(r.Flowlog.DstK8SNetworkName).Should(o.ContainSubstring("test-76537/l2-network"))
			o.Expect(r.Flowlog.SrcK8SNetworkName).Should(o.ContainSubstring("test-76537/l2-network"))
		}
	})

	g.It("Author:aramesha-NonPreRelease-Longduration-Medium-85887-Verify UDN with VMs [Disruptive][Slow]", func() {
		SkipIfOCPBelow("v4.18")
		var (
			// UDN vars
			udnNS   = "netobserv-udn-85887"
			udnName = "udn-network-85887"
			// VM vars
			testVMUDNTemplatePath = filePath.Join(virtualizationDir, "test-vm-UDN_template.yaml")
		)

		g.By("Deploy UDN in UDN ns")
		var cidr, ipv4cidr, ipv6cidr string
		defer deleteNamespace(oc, udnNS)
		oc.CreateSpecificNamespaceUDN(udnNS)

		if ipStackType == "ipv4single" {
			cidr = "10.151.0.0/16"
		} else {
			if ipStackType == "ipv6single" {
				cidr = "2011:100:200::0/48"
			} else {
				ipv4cidr = "10.151.0.0/16"
				ipv6cidr = "2011:100:200::0/48"
			}
		}
		createGeneralUDNCRD(oc, udnNS, udnName, ipv4cidr, ipv6cidr, cidr, "layer2")

		g.By("Deploy test VM3")
		testVM3 := TestVMUDNTemplate{
			Name:        "test-vm3",
			Namespace:   udnNS,
			NetworkName: udnName,
			Template:    testVMUDNTemplatePath,
		}
		defer deleteResource(oc, "vm", testVM3.Name, testVM3.Namespace)
		err := testVM3.createVMUDN(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		waitUntilVMReady(oc, testVM3.Name, testVM3.Namespace)

		g.By("Wait for VM3 to get IP assigned")
		vm3Ip, err := waitForVMIPAssignment(oc, testVM3.Name, testVM3.Namespace, 0)
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(vm3Ip).NotTo(o.BeEmpty())

		startTime := time.Now()

		g.By("Deploy test VM4")
		testVM4 := TestVMUDNTemplate{
			Name:        "test-vm4",
			Namespace:   udnNS,
			NetworkName: udnName,
			RunCmd:      fmt.Sprintf("[[ping, %s]]", vm3Ip),
			Template:    testVMUDNTemplatePath,
		}
		defer deleteResource(oc, "vm", testVM4.Name, testVM4.Namespace)
		err = testVM4.createVMUDN(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		waitUntilVMReady(oc, testVM4.Name, testVM4.Namespace)

		g.By("Wait for VM4 to get IP assigned")
		vm4Ip, err := waitForVMIPAssignment(oc, testVM4.Name, testVM4.Namespace, 0)
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(vm4Ip).NotTo(o.BeEmpty())

		g.By("Deploy FlowCollector with UDNMapping feature enabled with eBPF in privileged mode")
		flow := Flowcollector{
			Namespace:      namespace,
			EBPFPrivileged: "true",
			EBPFeatures:    []string{"\"UDNMapping\""},
			Template:       flowFixturePath,
		}

		defer func() { _ = flow.DeleteFlowcollector(oc) }()
		flow.CreateFlowcollector(oc)

		g.By("Wait for a min before logs gets collected and written to loki")
		time.Sleep(60 * time.Second)

		lokilabels := Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: udnNS,
			SrcK8SOwnerName: testVM4.Name,
			DstK8SNamespace: udnNS,
			DstK8SOwnerName: testVM3.Name,
		}

		g.By("Verify flows are written to loki")
		flowRecords, err := lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of flows written to loki > 0")

		g.By("Verify flow logs are enriched")
		// Get VM3 launcher pod name
		vm3podname, err := compat_otp.GetAllPodsWithLabel(oc, udnNS, "vm.kubevirt.io/name="+testVM3.Name)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(vm3podname).NotTo(o.BeEmpty())
		// Get VM4 launcher pod name
		vm4podname, err := compat_otp.GetAllPodsWithLabel(oc, udnNS, "vm.kubevirt.io/name="+testVM4.Name)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(vm4podname).NotTo(o.BeEmpty())

		for _, r := range flowRecords {
			o.Expect(r.Flowlog.DstK8SName).Should(o.ContainSubstring(vm3podname[0]))
			o.Expect(r.Flowlog.SrcK8SName).Should(o.ContainSubstring(vm4podname[0]))
			o.Expect(r.Flowlog.DstK8SOwnerType).Should(o.ContainSubstring("VirtualMachineInstance"))
			o.Expect(r.Flowlog.SrcK8SOwnerType).Should(o.ContainSubstring("VirtualMachineInstance"))
			o.Expect(r.Flowlog.Udns).Should(o.ContainElement(fmt.Sprintf("%s/%s", udnNS, udnName)))
			o.Expect(r.Flowlog.DstK8SNetworkName).Should(o.ContainSubstring(fmt.Sprintf("%s/%s", udnNS, udnName)))
			o.Expect(r.Flowlog.SrcK8SNetworkName).Should(o.ContainSubstring(fmt.Sprintf("%s/%s", udnNS, udnName)))
		}
	})

	g.It("Author:aramesha-High-85935-Validate CUDN with Localnet and VM's[Serial]", func() {
		SkipIfOCPBelow("v4.19")
		var (
			// NMstate operator vars
			opNamespace       = "openshift-nmstate"
			nmStateDir        = filePath.Join(networkingDir, "nmstate")
			nmstateCRTemplate = filePath.Join(nmStateDir, "nmstate-cr-template.yaml")
			nmstateCR         = nmstateCRResource{
				name:     "nmstate",
				template: nmstateCRTemplate,
			}
			nodeSelectLabel          = "node-role.kubernetes.io/worker"
			ovnMappingPolicyTemplate = filePath.Join(nmStateDir, "ovn-mapping-policy-template.yaml")
			ovnMappingPolicy         = ovnMappingPolicyResource{
				name:       "bridge-mapping-85935",
				nodelabel:  nodeSelectLabel,
				labelvalue: "",
				localnet1:  "mylocalnet",
				bridge1:    "br-ex",
				template:   ovnMappingPolicyTemplate,
			}
			// CUDN vars
			matchLabelKey              = "test.io"
			matchValue                 = "cudn-network-" + getRandomString()
			secondaryCUDNName          = "secondary-localnet-85935"
			cudnNS                     = []string{"netobserv-cudn1-85935", "netobserv-cudn2-85935"}
			testVMLocalnetTemplatePath = filePath.Join(virtualizationDir, "test-vm-localnet_template.yaml")
		)

		g.By("Check the platform and network plugin type if it is suitable for running the test")
		networkType := checkNetworkType(oc)
		if !(isPlatformSuitableForNMState(oc)) || !strings.Contains(networkType, "ovn") {
			g.Skip("Skipping for unsupported platform or non-OVN network plugin type!")
		}
		nmstateCatsrc := Resource{"catalogsource", "redhat-operators", "openshift-marketplace"}
		nmstateSource := CatalogSourceObjects{"stable", nmstateCatsrc.Name, nmstateCatsrc.Namespace}
		NMS := SubscriptionObjects{
			OperatorName:  "kubernetes-nmstate-operator",
			Namespace:     opNamespace,
			PackageName:   "kubernetes-nmstate-operator",
			Subscription:  filePath.Join(subscriptionDir, "sub-template.yaml"),
			OperatorGroup: filePath.Join(subscriptionDir, "singlenamespace-og.yaml"),
			CatalogSource: &nmstateSource,
		}
		ensureOperatorDeployed(oc, NMS, nmstateSource, "name="+NMS.OperatorName)

		workerNode, getNodeErr := compat_otp.GetFirstWorkerNode(oc)
		o.Expect(getNodeErr).NotTo(o.HaveOccurred())
		o.Expect(workerNode).NotTo(o.BeEmpty())

		compat_otp.By("Create NMState CR")
		defer deleteNMStateCR(oc, nmstateCR)
		result, crErr := createNMStateCR(oc, nmstateCR, opNamespace)
		compat_otp.AssertWaitPollNoErr(crErr, "create nmstate cr failed")
		o.Expect(result).To(o.BeTrue())
		e2e.Logf("SUCCESS - NMState CR Created")

		compat_otp.By("Configure NNCP for creating OvnMapping NMstate Feature")
		defer deleteNNCP(oc, ovnMappingPolicy.name)
		defer func() {
			ovnmapping, deferErr := compat_otp.DebugNodeWithChroot(oc, workerNode, "ovs-vsctl", "get", "Open_vSwitch", ".", "external_ids:ovn-bridge-mappings")
			o.Expect(deferErr).NotTo(o.HaveOccurred())
			if strings.Contains(ovnmapping, ovnMappingPolicy.localnet1) {
				// ovs-vsctl can only use "set" to reserve some fields
				_, err := compat_otp.DebugNodeWithChroot(oc, workerNode, "ovs-vsctl", "set", "Open_vSwitch", ".", "external_ids:ovn-bridge-mappings=\"physnet:br-ex\"")
				o.Expect(err).NotTo(o.HaveOccurred())
			}
		}()
		configErr3 := ovnMappingPolicy.configNNCP(oc)
		o.Expect(configErr3).NotTo(o.HaveOccurred())
		nncpErr3 := checkNNCPStatus(oc, ovnMappingPolicy.name, "Available")
		compat_otp.AssertWaitPollNoErr(nncpErr3, fmt.Sprintf("%s policy applied failed", ovnMappingPolicy.name))

		compat_otp.By("Create two namespaces and label them")
		for _, ns := range cudnNS {
			defer oc.DeleteSpecifiedNamespaceAsAdmin(ns)
			oc.CreateSpecifiedNamespaceAsAdmin(ns)
			err := oc.AsAdmin().WithoutNamespace().Run("label").Args("ns", ns, fmt.Sprintf("%s=%s", matchLabelKey, matchValue)).Execute()
			o.Expect(err).NotTo(o.HaveOccurred())
		}

		compat_otp.By("Create secondary localnet CUDN")
		defer removeResource(oc, true, true, "clusteruserdefinednetwork", secondaryCUDNName)
		_, err := applyLocalnetCUDNtoMatchLabelNS(oc, matchLabelKey, matchValue, secondaryCUDNName, "mylocalnet", "192.168.200.0/24", "192.168.200.1/32", false)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Deploy test VM5")
		testVM5 := TestVMUDNTemplate{
			Name:        "test-vm5",
			Namespace:   cudnNS[0],
			NetworkName: secondaryCUDNName,
			Template:    testVMLocalnetTemplatePath,
		}
		defer deleteResource(oc, "vm", testVM5.Name, testVM5.Namespace)
		err = testVM5.createVMUDN(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		waitUntilVMReady(oc, testVM5.Name, testVM5.Namespace)

		// Even though VM comes up as Ready, the IP assignment takes some time
		g.By("Wait for VM5 to get IP assigned")
		vm5Ip, err := waitForVMIPAssignment(oc, testVM5.Name, testVM5.Namespace, 1)
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(vm5Ip).NotTo(o.BeEmpty())

		startTime := time.Now()

		g.By("Deploy test VM6")
		testVM6 := TestVMUDNTemplate{
			Name:        "test-vm6",
			Namespace:   cudnNS[1],
			NetworkName: secondaryCUDNName,
			RunCmd:      fmt.Sprintf("[[ping, %s]]", vm5Ip),
			Template:    testVMLocalnetTemplatePath,
		}
		defer deleteResource(oc, "vm", testVM6.Name, testVM6.Namespace)
		err = testVM6.createVMUDN(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		waitUntilVMReady(oc, testVM6.Name, testVM6.Namespace)

		g.By("Wait for VM6 to get IP assigned")
		vm6Ip, err := waitForVMIPAssignment(oc, testVM6.Name, testVM6.Namespace, 1)
		o.Expect(err).ToNot(o.HaveOccurred())
		o.Expect(vm6Ip).NotTo(o.BeEmpty())

		g.By("Deploy FlowCollector with UDNMapping feature enabled with eBPF in privileged mode")
		flow := Flowcollector{
			Namespace:      namespace,
			EBPFPrivileged: "true",
			EBPFeatures:    []string{"\"UDNMapping\""},
			Template:       flowFixturePath,
		}

		defer func() { _ = flow.DeleteFlowcollector(oc) }()
		flow.CreateFlowcollector(oc)

		g.By("Wait for a min before logs gets collected and written to loki")
		time.Sleep(60 * time.Second)

		g.By("Verify CUDN Localnet flows")
		lokilabels := Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: cudnNS[1],
			SrcK8SOwnerName: testVM6.Name,
			DstK8SNamespace: cudnNS[0],
			DstK8SOwnerName: testVM5.Name,
		}

		flowRecords, err := lokilabels.GetMonolithicLokiFlowLogs(oc, flow.Namespace, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of CUDN Localnet flows > 0")
		for _, r := range flowRecords {
			o.Expect(r.Flowlog.Udns).Should(o.ContainElement(secondaryCUDNName))
			o.Expect(r.Flowlog.DstK8SNetworkName).Should(o.ContainSubstring(secondaryCUDNName))
			o.Expect(r.Flowlog.SrcK8SNetworkName).Should(o.ContainSubstring(secondaryCUDNName))
		}
	})
})
