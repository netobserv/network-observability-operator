package e2etests

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"

	g "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/onsi/gomega"
	exutil "github.com/openshift/origin/test/extended/util"
	e2eframework "k8s.io/kubernetes/test/e2e/framework"
)

func init() {
	// Initialize framework flags - must be done before flag.Parse()
	exutil.InitStandardFlags()
}

var _ = g.BeforeSuite(func() {
	// Parse flags
	flag.Parse()

	// Set up provider config after parsing flags
	e2eframework.AfterReadingAllFlags(&e2eframework.TestContext)

	// Control verbose event dumping on test failures via DUMP_EVENTS_ON_FAILURE env variable
	// Default: disabled (false)
	// Set DUMP_EVENTS_ON_FAILURE=true to enable verbose cluster-wide event dumps for debugging
	dumpEvents, err := strconv.ParseBool(os.Getenv("DUMP_EVENTS_ON_FAILURE"))
	if err != nil {
		dumpEvents = false
	}
	e2eframework.TestContext.DumpLogsOnFailure = dumpEvents

	// Initialize test
	gomega.Expect(exutil.InitTest(false)).NotTo(gomega.HaveOccurred())

	oc := exutil.NewCLIForMonitorTest("netobserv")
	_, err = GetOCPVersion(oc)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
})

func TestBackend(t *testing.T) {
	exutil.WithCleanup(func() {
		gomega.RegisterFailHandler(g.Fail)

		suiteConfig, reporterConfig := g.GinkgoConfiguration()

		// Apply focus filter

		if len(suiteConfig.FocusStrings) > 0 {
			combinedFocus := make([]string, len(suiteConfig.FocusStrings))
			for i, userFocus := range suiteConfig.FocusStrings {
				combinedFocus[i] = "sig-netobserv.*" + userFocus
			}
			suiteConfig.FocusStrings = combinedFocus
		} else {
			suiteConfig.FocusStrings = []string{"sig-netobserv"}
		}

		// Configure reporter - suppress default verbose output
		suiteConfig.EmitSpecProgress = true
		suiteConfig.OutputInterceptorMode = "none"
		reporterConfig.SilenceSkips = true // Hide the "S" characters for skipped tests
		reporterConfig.NoColor = true
		reporterConfig.Succinct = true
		reporterConfig.Verbose = false

		// Standard Ginkgo run with custom reporting via hooks
		g.RunSpecs(t, "Backend Suite", suiteConfig, reporterConfig)
	})
}

// Custom reporting hooks
var _ = g.ReportBeforeSuite(func(report g.Report) {
	fmt.Printf("Running Suite: %s - %s\n", report.SuiteDescription, report.SuitePath)
	fmt.Printf("==========================================================================================================\n")
	fmt.Printf("Random Seed: %d\n\n", report.SuiteConfig.RandomSeed)
	fmt.Printf("Will run %d specs\n", report.PreRunStats.SpecsThatWillRun)
})

var _ = g.ReportAfterEach(func(report g.SpecReport) {
	// Only report on specs that actually ran (not skipped on via focused filter)
	if report.State == types.SpecStateSkipped && report.Failure.Message == "" && report.RunTime <= 0 {
		return
	}

	if report.LeafNodeType != types.NodeTypeIt {
		return
	}

	// Print spec progress
	fmt.Printf("%s\n", report.FullText())
	fmt.Printf("%s\n", report.LeafNodeLocation.String())

	// Print result
	switch report.State {
	case types.SpecStatePassed:
		fmt.Printf("• PASSED [%.3f seconds]\n", report.RunTime.Seconds())
	case types.SpecStateSkipped:
		fmt.Printf("• SKIPPED [%.3f seconds]\n", report.RunTime.Seconds())
		if report.Failure.Message != "" {
			fmt.Printf("\n%s\n", report.Failure.Message)
			fmt.Printf("%s\n", report.Failure.Location.String())
		}
	case types.SpecStateFailed, types.SpecStatePanicked, types.SpecStateInvalid, types.SpecStateAborted, types.SpecStateInterrupted, types.SpecStatePending, types.SpecStateTimedout:
		fmt.Printf("• FAILED [%.3f seconds]\n", report.RunTime.Seconds())
		if report.Failure.Message != "" {
			fmt.Printf("\n%s\n", report.Failure.Message)
			fmt.Printf("%s\n", report.Failure.Location.String())
		}
	}
})

var _ = g.ReportAfterSuite("NetObserv Summary", func(report g.Report) {
	var passedSpecs []g.SpecReport
	var failedSpecs []g.SpecReport
	var skippedSpecs []g.SpecReport
	ranTests := 0

	// Get only the test specs (not setup/teardown)
	specs := report.SpecReports.WithLeafNodeType(types.NodeTypeIt)

	for _, specReport := range specs {
		switch specReport.State {
		case types.SpecStatePassed:
			passedSpecs = append(passedSpecs, specReport)
			ranTests++
		case types.SpecStateFailed, types.SpecStatePanicked, types.SpecStateInvalid, types.SpecStateAborted, types.SpecStateInterrupted, types.SpecStatePending, types.SpecStateTimedout:
			failedSpecs = append(failedSpecs, specReport)
			ranTests++
		case types.SpecStateSkipped:
			// Skip filtered-out specs: they have State==Skipped but RunTime==0 and no failure info
			// Explicitly skipped specs (via Skip()) have State==Skipped but were actually evaluated
			if specReport.Failure.Message != "" || specReport.RunTime > 0 {
				// Explicitly skipped - test body was evaluated
				skippedSpecs = append(skippedSpecs, specReport)
			}
		}
	}

	// Total specs evaluated (passed + failed + explicitly skipped)
	totalEvaluated := ranTests + len(skippedSpecs)

	fmt.Printf("==========================================================================================================\n")
	if report.SuiteSucceeded {
		fmt.Printf("Backend Suite - %d/%d specs • SUCCESS! [%.3f seconds]\n",
			len(passedSpecs), totalEvaluated, report.RunTime.Seconds())
	} else {
		fmt.Printf("Backend Suite - %d/%d specs • FAILURE! [%.3f seconds]\n",
			len(passedSpecs), totalEvaluated, report.RunTime.Seconds())
	}

	fmt.Printf("\nRan %d tests\n", ranTests)
	fmt.Printf("Passed: %d, Failed: %d, Skipped: %d\n", len(passedSpecs), len(failedSpecs), len(skippedSpecs))
	fmt.Printf("==========================================================================================================\n\n")

	// Print failed tests section
	if len(failedSpecs) > 0 {
		fmt.Printf("FAILED TESTS (%d):\n", len(failedSpecs))
		fmt.Printf("----------------------------------------------------------------------------------------------------------\n")
		for i, spec := range failedSpecs {
			fmt.Printf("%d. %s\n", i+1, spec.FullText())
			fmt.Printf("   %s [%.3f seconds]\n", spec.LeafNodeLocation.String(), spec.RunTime.Seconds())
			if spec.Failure.Message != "" {
				fmt.Printf("   Error: %s\n", spec.Failure.Message)
			}
		}
		fmt.Printf("\n")
	}

	// Print passed tests section
	if len(passedSpecs) > 0 {
		fmt.Printf("PASSED TESTS (%d):\n", len(passedSpecs))
		fmt.Printf("----------------------------------------------------------------------------------------------------------\n")
		for i, spec := range passedSpecs {
			fmt.Printf("%d. %s [%.3f seconds]\n", i+1, spec.FullText(), spec.RunTime.Seconds())
		}
		fmt.Printf("\n")
	}

	// Print skipped tests section
	if len(skippedSpecs) > 0 {
		fmt.Printf("SKIPPED TESTS (%d):\n", len(skippedSpecs))
		fmt.Printf("----------------------------------------------------------------------------------------------------------\n")
		for i, spec := range skippedSpecs {
			fmt.Printf("%d. %s [%.3f seconds]\n", i+1, spec.FullText(), spec.RunTime.Seconds())
			if spec.Failure.Message != "" {
				fmt.Printf("   Reason: %s\n", spec.Failure.Message)
			}
		}
		fmt.Printf("\n")
	}
})
