//nolint:revive
package vanillafs

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/netobserv/netobserv-operator/internal/controller/networkpolicy"
	"github.com/netobserv/netobserv-operator/internal/controller/networkpolicy/envtest"
	"github.com/netobserv/netobserv-operator/internal/pkg/manager"
	"github.com/netobserv/netobserv-operator/internal/pkg/test"
)

const (
	env = test.EnvVanillaFullStack
)

var (
	ctx          context.Context
	k8sClient    client.Client
	suiteContext *test.SuiteContext
)

func TestAPIsVanillaFullStack(t *testing.T) {
	// Uncomment and edit next line to run/debug from IDE (get the path by running: `bin/setup-envtest use 1.23 -p path`); you may need to override the test timeout in your settings.
	// os.Setenv("KUBEBUILDER_ASSETS", "/home/jotak/.local/share/kubebuilder-envtest/k8s/1.23.5-linux-amd64")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Networkpolicy Controller Suite - Vanilla Full Stack")
}

// go test ./... runs always Ginkgo test suites in parallel and they would interfere
// this way we make sure that both test sub-suites are executed serially
var _ = Describe("Networkpolicy Controller - Vanilla Full Stack", Ordered, Serial, func() {
	ctxGetter := func() (context.Context, client.Client) { return ctx, k8sClient }
	envtest.ControllerSpecs(ctxGetter)
})

var _ = BeforeSuite(func() {
	ctx, k8sClient, suiteContext = test.PrepareEnvTest(
		env,
		[]manager.Registerer{networkpolicy.Start},
		"main-namespace",
		[]string{"other-namespace"},
		"../../..",
	)
})

var _ = AfterSuite(func() {
	test.TeardownEnvTest(suiteContext)
})
