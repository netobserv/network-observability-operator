package monitoring

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	. "github.com/netobserv/network-observability-operator/controllers/controllerstest"
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

var updateCR = func(key types.NamespacedName, updater func(*flowslatest.FlowCollector)) {
	test.UpdateCR(ctx, k8sClient, key, updater)
}

// nolint:cyclop
func ControllerSpecs() {

	const operatorNamespace = "main-namespace"
	crKey := types.NamespacedName{
		Name: "cluster",
	}

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

			By("Expecting the monitoring dashboards configmap to be created")
			Eventually(func() interface{} {
				cm := v1.ConfigMap{}
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "grafana-dashboard-netobserv-flow-metrics",
					Namespace: "openshift-config-managed",
				}, &cm)
			}, timeout, interval).Should(Succeed())

			By("Expecting the infra health dashboards configmap to be created")
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

			By("Expecting the flow dashboards configmap to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "grafana-dashboard-netobserv-flow-metrics",
					Namespace: "openshift-config-managed",
				}, &v1.ConfigMap{})
			}, timeout, interval).Should(MatchError(`configmaps "grafana-dashboard-netobserv-flow-metrics" not found`))

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

	Context("Cleanup", func() {
		// Retrieve CR to get its UID
		flowCR := flowslatest.FlowCollector{}
		It("Should get CR", func() {
			Eventually(func() error {
				return k8sClient.Get(ctx, crKey, &flowCR)
			}, timeout, interval).Should(Succeed())
		})

		It("Should delete CR", func() {
			Eventually(func() error {
				return k8sClient.Delete(ctx, &flowCR)
			}, timeout, interval).Should(Succeed())
		})

		It("Should be garbage collected", func() {
			By("Expecting the health dashboards configmap to be garbage collected")
			Eventually(func() interface{} {
				cm := v1.ConfigMap{}
				_ = k8sClient.Get(ctx, types.NamespacedName{
					Name:      "grafana-dashboard-netobserv-health",
					Namespace: "openshift-config-managed",
				}, &cm)
				return &cm
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))
		})

		It("Should not get CR", func() {
			Eventually(func() bool {
				err := k8sClient.Get(ctx, crKey, &flowCR)
				return kerr.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
}
