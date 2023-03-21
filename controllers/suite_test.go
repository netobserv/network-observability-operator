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
	"net"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	operatorsv1 "github.com/openshift/api/operator/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/mock"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	flowsv1beta1 "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/operator"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const testCnoNamespace = "openshift-network-operator"

var (
	ctx        context.Context
	k8sManager manager.Manager
	k8sClient  client.Client
	testEnv    *envtest.Environment
	cancel     context.CancelFunc
	ipResolver ipResolverMock
)

func TestAPIs(t *testing.T) {
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
})

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			// We need to install the ConsolePlugin CRD to test setup of our Network Console Plugin
			filepath.Join("..", "vendor", "github.com", "openshift", "api", "console", "v1alpha1"),
			filepath.Join("..", "vendor", "github.com", "openshift", "api", "config", "v1"),
			filepath.Join("..", "vendor", "github.com", "openshift", "api", "operator", "v1"),
			filepath.Join("..", "test-assets"),
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = flowsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = flowsv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = osv1alpha1.Install(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = configv1.Install(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apiregv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = ascv2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = operatorsv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = monitoringv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	Expect(prepareNamespaces()).NotTo(HaveOccurred())

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sManager).NotTo(BeNil())

	err = NewTestFlowCollectorReconciler(k8sManager.GetClient(), k8sManager.GetScheme()).
		SetupWithManager(ctx, k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func prepareNamespaces() error {
	if err := k8sClient.Create(ctx, &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: testCnoNamespace},
	}); err != nil {
		return err
	}
	if err := k8sClient.Create(ctx, &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "openshift-config-managed"},
	}); err != nil {
		return err
	}
	return nil
}

// NewTestFlowCollectorReconciler allows mocking the IP resolutor of a
// FlowCollectorReconciler
func NewTestFlowCollectorReconciler(client client.Client, scheme *runtime.Scheme) *FlowCollectorReconciler {
	return &FlowCollectorReconciler{
		Client:   client,
		Scheme:   scheme,
		lookupIP: ipResolver.LookupIP,
		config: &operator.Config{
			EBPFAgentImage:        "registry-proxy.engineering.redhat.com/rh-osbs/network-observability-ebpf-agent@sha256:6481481ba23375107233f8d0a4f839436e34e50c2ec550ead0a16c361ae6654e",
			FlowlogsPipelineImage: "registry-proxy.engineering.redhat.com/rh-osbs/network-observability-flowlogs-pipeline@sha256:6481481ba23375107233f8d0a4f839436e34e50c2ec550ead0a16c361ae6654e",
			ConsolePluginImage:    "registry-proxy.engineering.redhat.com/rh-osbs/network-observability-console-plugin@sha256:6481481ba23375107233f8d0a4f839436e34e50c2ec550ead0a16c361ae6654e",
			DownstreamDeployment:  false,
		},
	}
}

type ipResolverMock struct {
	mock.Mock
}

func (ipr *ipResolverMock) LookupIP(host string) ([]net.IP, error) {
	m := ipr.Called(host)
	return m.Get(0).([]net.IP), m.Error(1)
}
