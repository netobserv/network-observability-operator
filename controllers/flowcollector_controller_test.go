package controllers

import (
	"net"
	"time"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	v1 "k8s.io/api/core/v1"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("FlowCollector Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1
	const flowCollectorPort = 999
	ipResolver.On("LookupIP", constants.GoflowKubeName+"."+operatorNamespace).
		Return([]net.IP{net.IPv4(11, 22, 33, 44)}, nil)
	expectedSharedTarget := "11.22.33.44:999"
	configMapKey := types.NamespacedName{
		Name:      "ovs-flows-config",
		Namespace: cnoNamespace,
	}
	key := types.NamespacedName{
		Name:      "test-cluster",
		Namespace: operatorNamespace,
	}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Cluster with autho-scaling", func() {
		It("Should create successfully", func() {

			created := &flowsv1alpha1.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
				},
				Spec: flowsv1alpha1.FlowCollectorSpec{
					GoflowKube: flowsv1alpha1.FlowCollectorGoflowKube{
						Kind:            "Deployment",
						Port:            flowCollectorPort,
						ImagePullPolicy: "Never",
						LogLevel:        "error",
						Image:           "testimg:latest",
					},
					IPFIX: flowsv1alpha1.FlowCollectorIPFIX{
						Sampling: 200,
					},
				},
			}

			// Create
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())

			By("Expecting to create the ovn-flows-configmap with the configuration from the FlowCollector")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, configMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "200",
				"sharedTarget":       expectedSharedTarget,
				"cacheMaxFlows":      "100",
				"cacheActiveTimeout": "10s",
			}))
		})

		It("Should update successfully", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, key, &fc); err != nil {
					return err
				}
				fc.Spec.IPFIX.CacheActiveTimeout = "30s"
				fc.Spec.IPFIX.Sampling = 1234
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting that ovn-flows-configmap is updated accordingly")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, configMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "1234",
				"sharedTarget":       expectedSharedTarget,
				"cacheMaxFlows":      "100",
				"cacheActiveTimeout": "30s",
			}))
		})

		It("Should delete successfully", func() {
			Eventually(func() error {
				f := &flowsv1alpha1.FlowCollector{}
				_ = k8sClient.Get(ctx, key, f)
				return k8sClient.Delete(ctx, f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete the ovn-flows-configmap")
			Eventually(func() error {
				return k8sClient.Get(ctx, configMapKey, &v1.ConfigMap{})
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
