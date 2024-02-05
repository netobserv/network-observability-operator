/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/netobserv/network-observability-operator/pkg/test"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var namespacesToPrepare = []string{"openshift-network-operator", "openshift-config-managed", "loki-namespace", "kafka-exporter-namespace", "main-namespace", "main-namespace-privileged"}

var (
	ctx       context.Context
	k8sClient client.Client
	testEnv   *envtest.Environment
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	// Uncomment and edit next line to run/debug from IDE (get the path by running: `bin/setup-envtest use 1.23 -p path`); you may need to override the test timeout in your settings.
	// os.Setenv("KUBEBUILDER_ASSETS", "/home/jotak/.local/share/kubebuilder-envtest/k8s/1.23.5-linux-amd64")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

// go test ./... runs always Ginkgo test suites in parallel and they would interfere
// this way we make sure that both test sub-suites are executed serially
var _ = Describe("FlowCollector Controller", Ordered, Serial, func() {
	flowCollectorControllerSpecs()
	flowCollectorConsolePluginSpecs()
	flowCollectorEBPFSpecs()
	flowCollectorEBPFKafkaSpecs()
	flowCollectorMinimalSpecs()
	flowCollectorIsoSpecs()
	flowCollectorCertificatesSpecs()
})

var _ = BeforeSuite(func() {
	ctx, k8sClient, testEnv, cancel = test.PrepareEnvTest(Registerers, namespacesToPrepare, ".")
})

var _ = AfterSuite(func() {
	test.TeardownEnvTest(testEnv, cancel)
})
