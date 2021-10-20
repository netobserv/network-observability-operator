package controllers

import (
	"context"
	"time"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("FlowCollector Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

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

			key := types.NamespacedName{
				Name:      "test-cluster",
				Namespace: operatorNamespace,
			}

			created := &flowsv1alpha1.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
				},
				Spec: flowsv1alpha1.FlowCollectorSpec{
					GoflowKube: flowsv1alpha1.FlowCollectorGoflowKube{
						Kind:            "Deployment",
						Port:            999,
						ImagePullPolicy: "Never",
						LogLevel:        "error",
						Image:           "testimg:latest",
					},
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &flowsv1alpha1.FlowCollector{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())
		})
	})
})
