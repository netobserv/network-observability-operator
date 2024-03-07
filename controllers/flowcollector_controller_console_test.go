package controllers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	operatorsv1 "github.com/openshift/api/operator/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/test"
)

// Because the simulated Kube server doesn't manage automatic resource cleanup like an actual Kube would do,
// we need either to cleanup all created resources manually, or to use different namespaces between tests
// For simplicity, we'll use a different namespace
const cpNamespace = "namespace-console-specs"

// nolint:cyclop
func flowCollectorConsolePluginSpecs() {
	cpDepl := test.Deployment(constants.PluginName)
	cpCM := test.ConfigMap("console-plugin-config")
	cpSvc := test.Service(constants.PluginName)
	cpSA := test.ServiceAccount(constants.PluginName)
	cpCRB := test.ClusterRoleBinding(constants.PluginName)
	cpSM := test.ServiceMonitor(constants.PluginName)
	crKey := types.NamespacedName{Name: "cluster"}
	consoleCRKey := types.NamespacedName{Name: "cluster"}
	configKey := cpCM.GetKey(cpNamespace)

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Console plugin test init", func() {
		It("Should create Console CR", func() {
			created := &operatorsv1.Console{
				ObjectMeta: metav1.ObjectMeta{
					Name: consoleCRKey.Name,
				},
				Spec: operatorsv1.ConsoleSpec{
					OperatorSpec: operatorsv1.OperatorSpec{
						ManagementState: operatorsv1.Unmanaged,
					},
				},
			}

			// Create
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())
		})

		It("Should create CR successfully", func() {
			created := &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: crKey.Name,
				},
				Spec: flowslatest.FlowCollectorSpec{
					Namespace:       cpNamespace,
					DeploymentModel: flowslatest.DeploymentModelDirect,
					Agent:           flowslatest.FlowCollectorAgent{Type: flowslatest.AgentEBPF},
					ConsolePlugin: flowslatest.FlowCollectorConsolePlugin{
						Enable:          ptr.To(true),
						ImagePullPolicy: "Never",
						Advanced: &flowslatest.AdvancedPluginConfig{
							Register: ptr.To(false),
						},
						Autoscaler: flowslatest.FlowCollectorHPA{
							Status:      flowslatest.HPAStatusEnabled,
							MinReplicas: ptr.To(int32(1)),
							MaxReplicas: 1,
							Metrics: []ascv2.MetricSpec{{
								Type: ascv2.ResourceMetricSourceType,
								Resource: &ascv2.ResourceMetricSource{
									Name: v1.ResourceCPU,
									Target: ascv2.MetricTarget{
										Type:               ascv2.UtilizationMetricType,
										AverageUtilization: ptr.To(int32(90)),
									},
								},
							}},
						},
						PortNaming: flowslatest.ConsolePluginPortConfig{
							Enable: ptr.To(true),
							PortNames: map[string]string{
								"3100": "loki",
							},
						},
					},
				},
			}

			// Create
			Eventually(func() interface{} {
				return k8sClient.Create(ctx, created)
			}, timeout, interval).Should(Succeed())
		})
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Deploying the console plugin", func() {
		It("Should create successfully", func() {

			objs := expectCreation(cpNamespace,
				cpDepl,
				cpSvc,
				cpCM,
				cpSA,
				cpCRB,
				cpSM,
			)
			Expect(objs).To(HaveLen(6))
			Expect(*objs[0].(*appsv1.Deployment).Spec.Replicas).To(Equal(int32(1)))
			Expect(objs[1].(*v1.Service).Spec.Ports[0].Port).To(Equal(int32(9001)))

			By("Creating the console plugin configmap")
			Eventually(getConfigMapData(configKey),
				timeout, interval).Should(ContainSubstring("url: http://loki:3100/"))

			By("Expecting to create console plugin role binding")
			rb := objs[4].(*rbacv1.ClusterRoleBinding)
			Expect(rb.Subjects).Should(HaveLen(1))
			Expect(rb.Subjects[0].Name).Should(Equal("netobserv-plugin"))
			Expect(rb.RoleRef.Name).Should(Equal("netobserv-plugin"))
		})

		It("Should update successfully", func() {
			Eventually(func() error {
				fc := flowslatest.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.ConsolePlugin.Advanced.Port = ptr.To(int32(9099))
				fc.Spec.ConsolePlugin.Replicas = ptr.To(int32(2))
				fc.Spec.ConsolePlugin.Autoscaler.Status = flowslatest.HPAStatusDisabled
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting the console plugin Deployment to be scaled up")
			Eventually(func() interface{} {
				dp := appsv1.Deployment{}
				if err := k8sClient.Get(ctx, cpDepl.GetKey(cpNamespace), &dp); err != nil {
					return err
				}
				return *dp.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(2)))

			By("Expecting the console plugin Service to be updated")
			Eventually(func() interface{} {
				svc := v1.Service{}
				if err := k8sClient.Get(ctx, cpSvc.GetKey(cpNamespace), &svc); err != nil {
					return err
				}
				return svc.Spec.Ports[0].Port
			}, timeout, interval).Should(Equal(int32(9099)))
		})

		It("Should create desired objects when they're not found (e.g. case of an operator upgrade)", func() {
			// Manually delete ServiceMonitor
			By("Deleting ServiceMonitor")
			Eventually(func() error { return k8sClient.Delete(ctx, cpSM.Resource) }, timeout, interval).Should(Succeed())

			// Do a dummy change that will trigger reconcile, and make sure SM is created again
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Processor.LogLevel = "trace"
			})

			By("Expecting ServiceMonitor to exist")
			expectCreation(cpNamespace, cpSM)
		})
	})

	Context("Configuring the Loki URL", func() {
		It("Should be initially configured with default Loki URL", func() {
			Eventually(getConfigMapData(configKey),
				timeout, interval).Should(ContainSubstring("url: http://loki:3100/"))
		})
		It("Should update the Loki URL in the Console Plugin if it changes in the Spec", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Monolithic.URL = "http://loki.namespace:8888"
			})
			Eventually(getConfigMapData(configKey),
				timeout, interval).Should(ContainSubstring("url: http://loki.namespace:8888"))
		})
		It("Should use the Loki Querier URL instead of the Loki URL, when switching to manual mode", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Mode = flowslatest.LokiModeManual
				fc.Spec.Loki.Manual.QuerierURL = "http://loki-querier:6789"
			})
			Eventually(getConfigMapData(configKey),
				timeout, interval).Should(ContainSubstring("url: http://loki-querier:6789"))
		})
	})

	Context("Registering to the Console CR", func() {
		It("Should start unregistered", func() {
			Eventually(func() interface{} {
				cr := operatorsv1.Console{}
				if err := k8sClient.Get(ctx, consoleCRKey, &cr); err != nil {
					return err
				}
				return cr.Spec.Plugins
			}, timeout, interval).Should(BeEmpty())
		})

		It("Should be registered", func() {
			By("Update CR to registered")
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.ConsolePlugin.Advanced.Register = ptr.To(true)
			})

			By("Expecting the Console CR to not have plugin registered")
			Eventually(func() interface{} {
				cr := operatorsv1.Console{}
				if err := k8sClient.Get(ctx, consoleCRKey, &cr); err != nil {
					return err
				}
				return cr.Spec.Plugins
			}, timeout, interval).Should(Equal([]string{"netobserv-plugin"}))
		})
	})

	Context("Update enable option", func() {
		It("Should be initially enabled", func() {
			Eventually(func() interface{} {
				fc := flowslatest.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				return *fc.Spec.ConsolePlugin.Enable
			}, timeout, interval).Should(Equal(true))
		})

		It("Should cleanup console plugin if disabled", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.ConsolePlugin.Enable = ptr.To(false)
			})

			expectDeletion(cpNamespace,
				cpDepl,
				cpSvc,
				cpSA,
				cpCM,
				cpSM,
			)
		})

		It("Should recreate console plugin if enabled back", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.ConsolePlugin.Enable = ptr.To(true)
			})

			expectCreation(cpNamespace,
				cpDepl,
				cpSvc,
				cpSA,
				cpCM,
				cpSM,
			)
		})
	})

	Context("Checking CR ownership", func() {
		It("Should be garbage collected", func() {
			expectOwnership(cpNamespace,
				cpDepl,
				cpSvc,
				cpSA,
				cpCM,
				cpSM,
			)
		})
	})

	Context("Changing namespace", func() {
		const otherNamespace = "other-namespace"

		It("Should update namespace successfully", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Namespace = otherNamespace
			})
		})

		It("Should redeploy console plugin in new namespace", func() {
			expectDeletion(cpNamespace,
				cpDepl,
				cpSvc,
				cpSA,
				cpCM,
				cpSM,
			)

			expectCreation(otherNamespace,
				cpDepl,
				cpSvc,
				cpSA,
				cpCM,
				cpSM,
			)
		})
	})

	Context("Cleanup", func() {
		It("Should delete CR", func() {
			cleanupCR(crKey)
		})

		It("Should delete Console CR", func() {
			Eventually(func() error {
				return k8sClient.Delete(ctx, &operatorsv1.Console{
					ObjectMeta: metav1.ObjectMeta{
						Name: consoleCRKey.Name,
					},
				})
			}, timeout, interval).Should(Succeed())
		})
	})
}

func getConfigMapData(configKey types.NamespacedName) func() interface{} {
	return func() interface{} {
		ofc := v1.ConfigMap{}
		if err := k8sClient.Get(ctx, configKey, &ofc); err != nil {
			return err
		}
		return ofc.Data["config.yaml"]
	}
}
