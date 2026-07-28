package e2etests

import (
	filePath "path/filepath"
	"regexp"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

var _ = g.Describe("[sig-netobserv] Network_Observability", func() {

	defer g.GinkgoRecover()
	var (
		oc              = compat_otp.NewCLI("netobserv", compat_otp.KubeConfigPath())
		flowmetricsPath = filePath.Join(baseDir, "flowmetrics_v1alpha1_template.yaml")
		flow            Flowcollector
	)

	g.BeforeEach(func() {
		flow = Flowcollector{
			Namespace:       oc.Namespace(),
			EBPFeatures:     []string{"\"FlowRTT\""},
			LokiEnable:      "false",
			InstallDemoLoki: "false",
			Template:        flowFixturePath,
		}
		flow.CreateFlowcollector(oc)
	})
	g.AfterEach(func() {
		_ = flow.DeleteFlowcollector(oc)
	})

	g.It("Author:memodi-High-73539-Create custom metrics and charts [Serial]", func() {
		namespace := oc.Namespace()
		customMetrics := CustomMetrics{
			Namespace: namespace,
			Template:  flowmetricsPath,
		}

		mainDashversion, err := getResourceVersion(oc, "cm", "netobserv-main", "openshift-config-managed")
		o.Expect(err).NotTo(o.HaveOccurred())
		curv, err := getResourceVersion(oc, "cm", "flowlogs-pipeline-config-dynamic", namespace)
		o.Expect(err).NotTo(o.HaveOccurred())

		customMetrics.createCustomMetrics(oc)
		waitForResourceGenerationUpdate(oc, "cm", "flowlogs-pipeline-config-dynamic", "resourceVersion", curv, namespace)

		customMetricsConfig := customMetrics.getCustomMetricConfigs()
		var allUniqueDash = make(map[string]bool)
		var uniqueDashboards []string
		for _, cmc := range customMetricsConfig {
			for _, dashboard := range cmc.DashboardNames {
				if _, ok := allUniqueDash[dashboard]; !ok {
					allUniqueDash[dashboard] = true
					uniqueDashboards = append(uniqueDashboards, dashboard)
				}
			}
			// verify custom metrics queries
			for _, query := range cmc.Queries {
				metricsQuery := strings.Replace(query, "$METRIC", "netobserv_"+cmc.MetricName, 1)
				metricVal := pollMetrics(oc, metricsQuery)
				e2e.Logf("metricsQuery %f for query %s", metricVal, metricsQuery)
			}
		}
		// verify dashboard exists
		for _, uniqDash := range uniqueDashboards {
			dashName := strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(uniqDash, "-"))
			if dashName == "main" {
				waitForResourceGenerationUpdate(oc, "cm", "netobserv-"+dashName, "resourceVersion", mainDashversion, "openshift-config-managed")
			}
			_, _ = checkResourceExists(oc, "cm", "netobserv-"+dashName, "openshift-config-managed")
		}
	})
})
