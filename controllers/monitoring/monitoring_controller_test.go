package monitoring

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/pkg/dashboards"
	"github.com/netobserv/network-observability-operator/pkg/test"
)

const (
	timeout                     = test.Timeout
	interval                    = test.Interval
	conntrackEndTimeout         = 10 * time.Second
	conntrackTerminatingTimeout = 5 * time.Second
	conntrackHeartbeatInterval  = 30 * time.Second
)

var (
	updateCR = func(key types.NamespacedName, updater func(*flowslatest.FlowCollector)) {
		test.UpdateCR(ctx, k8sClient, key, updater)
	}
	cleanupCR = func(key types.NamespacedName) {
		test.CleanupCR(ctx, k8sClient, key)
	}
	expectCreation = func(namespace string, objs ...test.ResourceRef) []client.Object {
		GinkgoHelper()
		return test.ExpectCreation(ctx, k8sClient, namespace, objs...)
	}
	expectDeletion = func(namespace string, objs ...test.ResourceRef) {
		GinkgoHelper()
		test.ExpectDeletion(ctx, k8sClient, namespace, objs...)
	}
	expectOwnership = func(namespace string, objs ...test.ResourceRef) {
		GinkgoHelper()
		test.ExpectOwnership(ctx, k8sClient, namespace, objs...)
	}
)

// nolint:cyclop
func ControllerSpecs() {

	const operatorNamespace = "main-namespace"
	crKey := types.NamespacedName{Name: "cluster"}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Installing CR", func() {
		It("Should create successfully", func() {
			created := &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: crKey.Name,
				},
				Spec: flowslatest.FlowCollectorSpec{
					Namespace:       operatorNamespace,
					DeploymentModel: flowslatest.DeploymentModelDirect,
				},
			}

			// Create
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())

			objs := expectCreation("openshift-config-managed",
				test.ConfigMap("grafana-dashboard-netobserv-flow-metrics"),
				test.ConfigMap("grafana-dashboard-netobserv-health"),
			)
			Expect(objs).To(HaveLen(2))
			healthCM := objs[1].(*v1.ConfigMap)
			d, err := dashboards.FromBytes([]byte(healthCM.Data["netobserv-health-metrics.json"]))
			Expect(err).To(BeNil())
			Expect(d.Titles()).To(Equal([]string{
				"Flows",
				"Flows Overhead",
				"Top flow rates per source and destination namespaces",
				"Agents",
				"Processor",
				"Operator",
			}))
		})

		It("Should update successfully", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Processor = flowslatest.FlowCollectorFLP{
					Metrics: flowslatest.FLPMetrics{
						IncludeList:   &[]flowslatest.FLPMetric{},
						DisableAlerts: []flowslatest.FLPAlert{flowslatest.AlertLokiError},
					},
				}
			})

			expectDeletion("openshift-config-managed",
				test.ConfigMap("grafana-dashboard-netobserv-flow-metrics"),
			)

			By("Expecting the health dashboards rows to be filtered")
			Eventually(func() interface{} {
				cm := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "grafana-dashboard-netobserv-health",
					Namespace: "openshift-config-managed",
				}, &cm); err != nil {
					return err
				}
				d, err := dashboards.FromBytes([]byte(cm.Data["netobserv-health-metrics.json"]))
				if err != nil {
					return err
				}
				return d.Titles()
			}, timeout, interval).Should(Equal([]string{
				"Flows",
				"Agents",
				"Processor",
				"Operator",
			}))
		})
	})

	Context("Checking CR ownership", func() {
		It("Should be garbage collected", func() {
			expectOwnership("openshift-config-managed",
				test.ConfigMap("grafana-dashboard-netobserv-health"),
			)
		})
	})

	Context("Cleanup", func() {
		It("Should delete CR", func() {
			cleanupCR(crKey)
		})
	})
}
