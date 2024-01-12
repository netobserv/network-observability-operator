package flp

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/netobserv/network-observability-operator/pkg/manager"
	"github.com/netobserv/network-observability-operator/pkg/test"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var namespacesToPrepare = []string{"main-namespace", "other-namespace"}

var (
	ctx       context.Context
	k8sClient client.Client
	testEnv   *envtest.Environment
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FLP Controller Suite")
}

// go test ./... runs always Ginkgo test suites in parallel and they would interfere
// this way we make sure that both test sub-suites are executed serially
var _ = Describe("FLP Controller", Ordered, Serial, func() {
	ControllerSpecs()
	ControllerFlowMetricsSpecs()
})

var _ = BeforeSuite(func() {
	ctx, k8sClient, testEnv, cancel = test.PrepareEnvTest([]manager.Registerer{Start}, namespacesToPrepare, "..")
})

var _ = AfterSuite(func() {
	test.TeardownEnvTest(testEnv, cancel)
})
