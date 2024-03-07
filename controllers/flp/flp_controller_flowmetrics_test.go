package flp

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
)

// nolint:cyclop
func ControllerFlowMetricsSpecs() {
	const operatorNamespace = "main-namespace"
	const otherNamespace = "other-namespace"
	crKey := types.NamespacedName{
		Name: "cluster",
	}
	flpKey := types.NamespacedName{
		Name:      constants.FLPName,
		Namespace: operatorNamespace,
	}
	cmKey := types.NamespacedName{
		Name:      "flowlogs-pipeline-config",
		Namespace: operatorNamespace,
	}
	metric1 := metricslatest.FlowMetric{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metric-1",
			Namespace: operatorNamespace,
		},
		Spec: metricslatest.FlowMetricSpec{
			MetricName: "m_1",
			Type:       metricslatest.CounterMetric,
		},
	}
	metric2 := metricslatest.FlowMetric{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metric-2",
			Namespace: operatorNamespace,
		},
		Spec: metricslatest.FlowMetricSpec{
			MetricName: "m_2",
			Type:       metricslatest.CounterMetric,
		},
	}
	metricUnwatched := metricslatest.FlowMetric{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metric-unwatched",
			Namespace: otherNamespace,
		},
		Spec: metricslatest.FlowMetricSpec{
			MetricName: "m_3",
			Type:       metricslatest.CounterMetric,
		},
	}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Deploying default FLP", func() {
		depl := appsv1.Deployment{}
		cm := v1.ConfigMap{}
		It("Should create successfully", func() {
			created := &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
				Spec: flowslatest.FlowCollectorSpec{
					Namespace:       operatorNamespace,
					DeploymentModel: flowslatest.DeploymentModelKafka,
				},
			}

			// Create
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())

			By("Expecting to create the flowlogs-pipeline Deployment")
			Eventually(func() error {
				return k8sClient.Get(ctx, flpKey, &depl)
			}, timeout, interval).Should(Succeed())

			By("Expecting flowlogs-pipeline-config configmap to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cmKey, &cm)
			}, timeout, interval).Should(Succeed())

			metrics, err := getConfiguredMetrics(&cm)
			Expect(err).NotTo(HaveOccurred())
			Expect(metrics).To(HaveLen(3)) // only default metrics
		})
	})

	Context("Creating FlowMetrics", func() {
		It("Should create successfully", func() {
			Expect(k8sClient.Create(ctx, &metric1)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &metricUnwatched)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &metric2)).Should(Succeed())
		})

		It("Should update configmap with custom metrics", func() {
			Eventually(func() interface{} {
				cm := v1.ConfigMap{}
				err := k8sClient.Get(ctx, cmKey, &cm)
				if err != nil {
					return err
				}
				metrics, err := getConfiguredMetrics(&cm)
				if err != nil {
					return err
				}
				return metrics
			}, timeout, interval).Should(Satisfy(func(metrics api.MetricsItems) bool {
				names := getSortedMetricsNames(metrics)
				return slices.Contains(names, metric1.Spec.MetricName) &&
					slices.Contains(names, metric2.Spec.MetricName) &&
					!slices.Contains(names, metricUnwatched.Spec.MetricName)
			}))
		})
	})

	Context("Updating a FlowMetric", func() {
		It("Should update successfully", func() {
			metric1.Spec.MetricName = "m_1_bis"
			Expect(k8sClient.Update(ctx, &metric1)).Should(Succeed())
		})

		It("Should update configmap with custom metrics", func() {
			Eventually(func() interface{} {
				cm := v1.ConfigMap{}
				err := k8sClient.Get(ctx, cmKey, &cm)
				if err != nil {
					return err
				}
				metrics, err := getConfiguredMetrics(&cm)
				if err != nil {
					return err
				}
				return metrics
			}, timeout, interval).Should(Satisfy(func(metrics api.MetricsItems) bool {
				names := getSortedMetricsNames(metrics)
				return slices.Contains(names, "m_1_bis") &&
					slices.Contains(names, metric2.Spec.MetricName) &&
					!slices.Contains(names, metricUnwatched.Spec.MetricName)
			}))
		})
	})

	Context("Cleanup", func() {
		It("Should delete CR", func() {
			cleanupCR(crKey)
		})
	})
}
