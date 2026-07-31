package e2etests

import (
	"fmt"
	"os"
	filePath "path/filepath"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

var _ = g.Describe("[sig-netobserv] Network_Observability Multitenancy", g.Ordered, g.ContinueOnFailure, func() {
	defer g.GinkgoRecover()

	var (
		oc = compat_otp.NewCLI("netobserv", compat_otp.KubeConfigPath())

		kubeadminToken string
		namespace      string

		// Loki Operator variables
		lokiDir         = filePath.Join(baseDir, "loki")
		lokiPackageName = "loki-operator"
		lokiCatalog     = "redhat-operators"
		lokiSource      CatalogSourceObjects
		ls              *lokiStack
		Lokiexisting    = false
		lokiStackNS     = "netobserv-loki"
		LO              = SubscriptionObjects{
			OperatorName:  "loki-operator-controller-manager",
			Namespace:     loNS,
			PackageName:   lokiPackageName,
			Subscription:  filePath.Join(subscriptionDir, "sub-template.yaml"),
			OperatorGroup: filePath.Join(subscriptionDir, "allnamespace-og.yaml"),
			CatalogSource: &lokiSource,
		}

		// LokiStack variables
		ipStackType       string
		lokiStackTemplate = filePath.Join(lokiDir, "lokistack-simple.yaml")
		lokiTenant        = "openshift-network"

		// For FlowCollectorSlice tests
		flowSliceFixturePath = filePath.Join(baseDir, "flowcollectorSlice_v1alpha1_template.yaml")
	)

	g.BeforeAll(func() {
		oc.SetNamespace(netobservNS)
		namespace = oc.Namespace()

		g.By("Get kubeadmin token")
		kubeAdminPasswd := os.Getenv("QE_KUBEADMIN_PASSWORD")
		if kubeAdminPasswd == "" {
			g.Skip("no kubeAdminPasswd is provided in this profile, set QE_KUBEADMIN_PASSWORD env var")
		}
		serverURL, serverURLErr := oc.AsAdmin().WithoutNamespace().Run("whoami").Args("--show-server").Output()
		o.Expect(serverURLErr).NotTo(o.HaveOccurred())
		currentContext, currentContextErr := oc.WithoutNamespace().Run("config").Args("current-context").Output()
		o.Expect(currentContextErr).NotTo(o.HaveOccurred())
		kubeadminToken = getKubeAdminToken(oc, kubeAdminPasswd, serverURL, currentContext)
		o.Expect(kubeadminToken).NotTo(o.BeEmpty())

		ipStackType = checkIPStackType(oc)

		g.By("Deploy loki operator")
		if !validateInfraAndResourcesForLoki(oc, "10Gi", "6") {
			g.Skip("Current platform does not have enough resources available for this test!")
		}

		// check if Loki Operator exists
		var err error
		Lokiexisting, err = CheckOperatorStatus(oc, LO.Namespace, LO.PackageName)
		o.Expect(err).NotTo(o.HaveOccurred())

		lokiChannel, err := getOperatorChannel(oc, lokiCatalog, lokiPackageName)
		if err != nil || lokiChannel == "" {
			g.Skip("Loki channel not found, skip this case")
		}
		lokiSource = CatalogSourceObjects{lokiChannel, lokiCatalog, "openshift-marketplace"}

		if !Lokiexisting {
			ensureOperatorDeployed(oc, LO, lokiSource, "name="+LO.OperatorName)
		} else {
			channelName, err := checkOperatorChannel(oc, LO.Namespace, LO.PackageName)
			o.Expect(err).NotTo(o.HaveOccurred())
			if channelName != lokiChannel {
				e2e.Logf("found %s channel for loki operator, removing and reinstalling with %s channel instead", channelName, lokiSource.Channel)
				LO.uninstallOperator(oc)
				ensureOperatorDeployed(oc, LO, lokiSource, "name="+LO.OperatorName)
				Lokiexisting = false
			}
		}

		g.By("Deploy lokiStack")
		sc, err := getStorageClassName(oc)
		if err != nil || len(sc) == 0 {
			g.Skip("StorageClass not found in cluster, skip this case")
		}

		objectStorageType := getStorageType(oc)
		if len(objectStorageType) == 0 && ipStackType != "ipv6single" {
			g.Skip("Current cluster doesn't have a proper object storage for this test!")
		}

		oc.CreateSpecifiedNamespaceAsAdmin(lokiStackNS)

		ls = &lokiStack{
			Name:          "lokistack",
			Namespace:     lokiStackNS,
			TSize:         "1x.demo",
			StorageType:   objectStorageType,
			StorageSecret: "objectstore-secret",
			StorageClass:  sc,
			BucketName:    "netobserv-loki-" + getInfrastructureName(oc),
			Tenant:        lokiTenant,
			Template:      lokiStackTemplate,
		}

		if ipStackType == "ipv6single" {
			e2e.Logf("running IPv6 test")
			ls.EnableIPV6 = "true"
		}

		err = ls.prepareResourcesForLokiStack(oc)
		if err != nil {
			g.Skip("Skipping test since LokiStack resources were not deployed")
		}

		err = ls.deployLokiStack(oc)
		if err != nil {
			g.Skip("Skipping test since LokiStack was not deployed")
		}

		lokiStackResource := Resource{"lokistack", ls.Name, ls.Namespace}
		err = lokiStackResource.WaitForResourceToAppear(oc)
		if err != nil {
			g.Skip("Skipping test since LokiStack did not become ready")
		}

		err = ls.waitForLokiStackToBeReady(oc)
		if err != nil {
			g.Skip("Skipping test since LokiStack is not ready")
		}
		ls.Route = "https://" + getRouteAddress(oc, ls.Namespace, ls.Name)

		g.By("Create secret-watcher RoleBinding in LokiStack namespace")
		createSecretWatcherRB(oc, lokiStackNS)
	})

	g.AfterAll(func() {
		if ls != nil {
			ls.removeLokiStack(oc)
			ls.removeObjectStorage(oc)
		}
		oc.DeleteSpecifiedNamespaceAsAdmin(lokiStackNS)
		if !Lokiexisting {
			LO.uninstallOperator(oc)
		}
	})

	g.BeforeEach(func() {
		namespace = oc.Namespace()
	})

	g.It("Author:memodi-NonPreRelease-Longduration-High-63839-Verify-multi-tenancy [Disruptive][Slow]", func() {
		users, usersHTpassFile, htPassSecret := getNewUser(oc, 2)
		defer userCleanup(oc, users, usersHTpassFile, htPassSecret)

		g.By("Creating client server template and template CRBs for testusers")
		// create templates for testuser to be used later
		testUserstemplate := filePath.Join(baseDir, "testuser-client-server_template.yaml")
		stdout, stderr, err := oc.AsAdmin().Run("apply").Args("-f", testUserstemplate).Outputs()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(stderr).To(o.BeEmpty())
		templateResource := strings.Split(stdout, " ")[0]
		templateName := strings.Split(templateResource, "/")[1]
		defer removeTemplatePermissions(oc, users[0].Username)
		addTemplatePermissions(oc, users[0].Username)

		g.By("Deploy FlowCollector")
		flow := Flowcollector{
			Namespace:       namespace,
			Template:        flowFixturePath,
			LokiMode:        "LokiStack",
			LokiNamespace:   lokiStackNS,
			InstallDemoLoki: "false",
		}

		defer func() { _ = flow.DeleteFlowcollector(oc) }()
		flow.CreateFlowcollector(oc)

		g.By("Deploying test server and client pods")
		serverTemplate := filePath.Join(baseDir, "test-nginx-server_template.yaml")
		testServerTemplate := TestServerTemplate{
			ServerNS: "test-server-63839",
			Template: serverTemplate,
		}
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testServerTemplate.ServerNS)
		err = testServerTemplate.createServer(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		compat_otp.AssertAllPodsToBeReady(oc, testServerTemplate.ServerNS)

		clientTemplate := filePath.Join(baseDir, "test-nginx-client_template.yaml")

		testClientTemplate := TestClientTemplate{
			ServerNS: testServerTemplate.ServerNS,
			ClientNS: "test-client-63839",
			Template: clientTemplate,
		}

		defer oc.DeleteSpecifiedNamespaceAsAdmin(testClientTemplate.ClientNS)
		err = testClientTemplate.createClient(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		compat_otp.AssertAllPodsToBeReady(oc, testClientTemplate.ClientNS)

		// save original context
		origContxt, contxtErr := oc.AsAdmin().WithoutNamespace().Run("config").Args("current-context").Output()
		o.Expect(contxtErr).NotTo(o.HaveOccurred())
		e2e.Logf("orginal context is %v", origContxt)
		defer removeUserAsReader(oc, users[0].Username)
		addUserAsReader(oc, users[0].Username)
		origUser := oc.Username()

		e2e.Logf("current user is %s", origUser)
		defer func() { _ = oc.AsAdmin().WithoutNamespace().Run("config").Args("use-context", origContxt).Execute() }()
		defer oc.ChangeUser(origUser)
		oc.ChangeUser(users[0].Username)

		curUser := oc.Username()
		e2e.Logf("current user is %s", curUser)

		o.Expect(err).NotTo(o.HaveOccurred())
		user0Contxt, contxtErr := oc.WithoutNamespace().Run("config").Args("current-context").Output()
		o.Expect(contxtErr).NotTo(o.HaveOccurred())

		e2e.Logf("user0 context is %v", user0Contxt)

		g.By("Deploying test server and client pods as user0")
		var (
			testUserServerNS = fmt.Sprintf("%s-server", users[0].Username)
			testUserClientNS = fmt.Sprintf("%s-client", users[0].Username)
		)

		defer oc.DeleteSpecifiedNamespaceAsAdmin(testUserClientNS)
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testUserServerNS)
		configFile := compat_otp.ProcessTemplate(oc, "--ignore-unknown-parameters=true", templateName, "-p", "SERVER_NS="+testUserServerNS, "-p", "CLIENT_NS="+testUserClientNS)
		err = oc.WithoutNamespace().Run("create").Args("-f", configFile).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		// only required to getFlowLogs
		lokilabels := Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testUserServerNS,
			DstK8SNamespace: testUserClientNS,
			SrcK8SOwnerName: "nginx-service",
			FlowDirection:   "0",
		}

		user0token, err := oc.WithoutNamespace().Run("whoami").Args("-t").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Wait for a min before logs gets collected and written to loki")
		startTime := time.Now()
		time.Sleep(60 * time.Second)

		g.By("get flowlogs from loki")
		flowRecords, err := lokilabels.getLokiFlowLogs(user0token, ls.Route, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected number of flowRecords > 0")

		g.By("verify no logs are fetched from an NS that user is not admin for")
		lokilabels = Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testClientTemplate.ServerNS,
			DstK8SNamespace: testClientTemplate.ClientNS,
			SrcK8SOwnerName: "nginx-service",
			FlowDirection:   "0",
			AllowEmpty:      true,
		}
		flowRecords, err = lokilabels.getLokiFlowLogs(user0token, ls.Route, startTime)
		// Multi-tenancy verification: Loki Gateway returns permission errors for unauthorized namespace access
		if err != nil {
			o.Expect(err.Error()).To(o.ContainSubstring("permission"), "expected permission error for unauthorized namespace access, got: %v", err)
		} else {
			o.Expect(len(flowRecords)).To(o.Equal(0), "expected zero flowRecords for unauthorized namespace access")
		}
	})

	g.It("Author:aramesha-NonPreRelease-Longduration-High-87145-Verify FlowCollectorSlices multi-tenancy [Disruptive][Slow]", func() {
		g.By("Creating test users")
		users, usersHTpassFile, htPassSecret := getNewUser(oc, 1)
		defer userCleanup(oc, users, usersHTpassFile, htPassSecret)

		g.By("Deploy FlowCollector with Slices enabled")
		flow := Flowcollector{
			Namespace:       namespace,
			LokiMode:        "LokiStack",
			LokiNamespace:   lokiStackNS,
			InstallDemoLoki: "false",
			CollectionMode:  "AllowList",
			SlicesEnable:    "true",
			NamespacesAllow: []string{"\"/openshift-.*/\""},
			Template:        flowFixturePath,
		}

		defer func() { _ = flow.DeleteFlowcollector(oc) }()
		flow.CreateFlowcollector(oc)

		g.By("Verify FlowCollectorSlices ClusterRoles exist")
		clusterRoleOutput, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("clusterrole", "-o=jsonpath={.items[*].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(clusterRoleOutput).Should(o.ContainSubstring("flowcollectorslices.flows.netobserv.io-v1alpha1-admin"))
		o.Expect(clusterRoleOutput).Should(o.ContainSubstring("flowcollectorslices.flows.netobserv.io-v1alpha1-edit"))
		o.Expect(clusterRoleOutput).Should(o.ContainSubstring("flowcollectorslices.flows.netobserv.io-v1alpha1-view"))

		g.By("Deploy test server and client pods for test-a namespace")
		serverTemplate := filePath.Join(baseDir, "test-nginx-server_template.yaml")
		testServerTemplateA := TestServerTemplate{
			ServerNS: "test-a-server-87145",
			Template: serverTemplate,
		}
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testServerTemplateA.ServerNS)
		err = testServerTemplateA.createServer(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		compat_otp.AssertAllPodsToBeReady(oc, testServerTemplateA.ServerNS)

		clientTemplate := filePath.Join(baseDir, "test-nginx-client_template.yaml")
		testClientTemplateA := TestClientTemplate{
			ServerNS: testServerTemplateA.ServerNS,
			ClientNS: "test-a-client-87145",
			Template: clientTemplate,
		}
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testClientTemplateA.ClientNS)
		err = testClientTemplateA.createClient(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		compat_otp.AssertAllPodsToBeReady(oc, testClientTemplateA.ClientNS)

		// Save original context
		origContxt, contxtErr := oc.AsAdmin().WithoutNamespace().Run("config").Args("current-context").Output()
		o.Expect(contxtErr).NotTo(o.HaveOccurred())
		e2e.Logf("original context is %v", origContxt)
		defer func() { _ = oc.AsAdmin().WithoutNamespace().Run("config").Args("use-context", origContxt).Execute() }()

		origUser := oc.Username()
		e2e.Logf("original user is %s", origUser)
		defer oc.ChangeUser(origUser)

		testUserName := users[0].Username
		oc.ChangeUser(testUserName)
		e2e.Logf("switched to user: %s", testUserName)

		g.By("Create namespace test-a and grant testuser-0 admin permissions")
		testNSA := testClientTemplateA.ClientNS
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("rolebinding", "testuser-0-admin",
			"--clusterrole=flowcollectorslices.flows.netobserv.io-v1alpha1-admin",
			"--user="+testUserName, "-n", testNSA).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		// Grant testuser admin access to the server namespace as well for flow visibility
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("rolebinding", "testuser-0-admin-server",
			"--clusterrole=admin",
			"--user="+testUserName, "-n", testServerTemplateA.ServerNS).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		// Grant testuser admin access to client namespace
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("rolebinding", "testuser-0-admin-client",
			"--clusterrole=admin",
			"--user="+testUserName, "-n", testNSA).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		// Grant loki reader access for multi-tenancy
		defer removeUserAsReader(oc, testUserName)
		addUserAsReader(oc, testUserName)

		g.By("Verify testuser-0 can create flowcollectorslices in test-a")
		canCreate, err := oc.WithoutNamespace().Run("auth").Args("can-i", "create", "flowcollectorslices", "-n", testNSA).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(canCreate).Should(o.ContainSubstring("yes"))

		g.By("Create a FlowCollectorSlice in test-a namespace")
		flowSliceA := FlowcollectorSlice{
			Name:      "test-a-slice",
			Namespace: testNSA,
			Sampling:  "100",
			Template:  flowSliceFixturePath,
		}
		defer func() { _ = flowSliceA.DeleteFlowcollectorSlice(oc) }()
		flowSliceA.CreateFlowcollectorSlice(oc)
		flowSliceA.WaitForFlowcollectorSliceReady(oc)

		g.By("Verify testuser-0 can view the FlowCollectorSlice in test-a")
		sliceOutput, err := oc.WithoutNamespace().Run("get").Args("flowcollectorslice", "-n", testNSA, "-o=jsonpath={.items[*].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(sliceOutput).Should(o.ContainSubstring("test-a-slice"))

		// Verify sampling value
		samplingValue, err := oc.WithoutNamespace().Run("get").Args("flowcollectorslice", "test-a-slice", "-n", testNSA, "-o=jsonpath={.spec.sampling}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(samplingValue).Should(o.Equal("100"))

		g.By("Verify testuser-0 can update the FlowCollectorSlice in test-a")
		err = oc.WithoutNamespace().Run("patch").Args("flowcollectorslice", "test-a-slice", "-n", testNSA,
			"--type=merge", "-p={\"spec\":{\"sampling\":2}}").Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		// Verify sampling was updated
		samplingValue, err = oc.WithoutNamespace().Run("get").Args("flowcollectorslice", "test-a-slice", "-n", testNSA, "-o=jsonpath={.spec.sampling}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(samplingValue).Should(o.Equal("2"))

		g.By("Get testuser token for loki query")
		user0token, err := oc.WithoutNamespace().Run("whoami").Args("-t").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Wait for flows to be collected and written to loki")
		startTime := time.Now()
		time.Sleep(60 * time.Second)

		g.By("Verify testuser-0 can access flows from test-a namespace")
		lokilabels := Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testServerTemplateA.ServerNS,
			DstK8SNamespace: testNSA,
			SrcK8SOwnerName: "nginx-service",
			FlowDirection:   "0",
		}
		flowRecords, err := lokilabels.getLokiFlowLogs(user0token, ls.Route, startTime)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(len(flowRecords)).Should(o.BeNumerically(">", 0), "expected testuser to see flows from test-a namespace")

		g.By("Create namespace test-b and create a FlowCollectorSlice as kubeadmin")
		testServerTemplateB := TestServerTemplate{
			ServerNS: "test-b-server-87145",
			Template: serverTemplate,
		}
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testServerTemplateB.ServerNS)

		testClientTemplateB := TestClientTemplate{
			ServerNS: testServerTemplateB.ServerNS,
			ClientNS: "test-b-client-87145",
			Template: clientTemplate,
		}
		defer oc.DeleteSpecifiedNamespaceAsAdmin(testClientTemplateB.ClientNS)

		// Switch to admin to create test-b resources
		oc.ChangeUser(origUser)
		err = testServerTemplateB.createServer(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		compat_otp.AssertAllPodsToBeReady(oc, testServerTemplateB.ServerNS)

		err = testClientTemplateB.createClient(oc)
		o.Expect(err).NotTo(o.HaveOccurred())
		compat_otp.AssertAllPodsToBeReady(oc, testClientTemplateB.ClientNS)

		testNSB := testClientTemplateB.ClientNS
		flowSliceB := FlowcollectorSlice{
			Name:      "test-b-slice",
			Namespace: testNSB,
			Sampling:  "3",
			Template:  flowSliceFixturePath,
		}
		defer func() { _ = flowSliceB.DeleteFlowcollectorSlice(oc) }()
		flowSliceB.CreateFlowcollectorSlice(oc)
		flowSliceB.WaitForFlowcollectorSliceReady(oc)

		// Switch back to testuser
		oc.ChangeUser(testUserName)

		g.By("Verify testuser-0 cannot see test-b slice")
		sliceOutputB, err := oc.WithoutNamespace().Run("get").Args("flowcollectorslice", "-n", testNSB).Output()
		o.Expect(err).Should(o.HaveOccurred())
		o.Expect(sliceOutputB).Should(o.MatchRegexp(`User ".*" cannot list resource "flowcollectorslices"`))

		g.By("Verify testuser-0 cannot access flows from test-b namespace")
		lokilabels = Lokilabels{
			App:             "netobserv-flowcollector",
			SrcK8SNamespace: testServerTemplateB.ServerNS,
			DstK8SNamespace: testNSB,
			SrcK8SOwnerName: "nginx-service",
			FlowDirection:   "0",
			AllowEmpty:      true,
		}
		flowRecords, err = lokilabels.getLokiFlowLogs(user0token, ls.Route, startTime)
		// Multi-tenancy verification: Loki Gateway returns permission errors for unauthorized namespace access
		if err != nil {
			o.Expect(err.Error()).Should(o.ContainSubstring("permission"), "expected permission error for unauthorized namespace access, got: %v", err)
		} else {
			o.Expect(len(flowRecords)).Should(o.BeNumerically("==", 0), "expected testuser to NOT see flows from test-b namespace")
		}

		g.By("Add testuser-0 as viewer for test-b namespace")
		err = oc.AsAdmin().WithoutNamespace().Run("create").Args("rolebinding", "testuser-0-view",
			"--clusterrole=flowcollectorslices.flows.netobserv.io-v1alpha1-view",
			"--user="+testUserName, "-n", testNSB).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Verify testuser-0 can view FlowCollectorSlice in test-b")
		sliceOutput, err = oc.WithoutNamespace().Run("get").Args("flowcollectorslice", "-n", testNSB, "-o=jsonpath={.items[*].metadata.name}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(sliceOutput).Should(o.ContainSubstring("test-b-slice"))

		g.By("Verify testuser-0 cannot update FlowCollectorSlice in test-b (view-only)")
		patchOutput, err := oc.WithoutNamespace().Run("patch").Args("flowcollectorslice", "test-b-slice", "-n", testNSB,
			"--type=merge", "-p={\"spec\":{\"sampling\":25}}").Output()
		o.Expect(err).Should(o.HaveOccurred())
		o.Expect(patchOutput).Should(o.MatchRegexp(`User ".*" cannot patch resource "flowcollectorslices"`))

		g.By("Remove testuser-0's view access from test-b")
		err = oc.AsAdmin().WithoutNamespace().Run("delete").Args("rolebinding", "testuser-0-view", "-n", testNSB).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("Verify access to test-b FlowCollectorSlices is revoked")
		sliceOutputRevoked, err := oc.WithoutNamespace().Run("get").Args("flowcollectorslice", "-n", testNSB).Output()
		o.Expect(err).Should(o.HaveOccurred())
		o.Expect(sliceOutputRevoked).Should(o.MatchRegexp(`User ".*" cannot list resource "flowcollectorslices"`))
	})
})
