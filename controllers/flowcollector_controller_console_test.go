package controllers

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	operatorsv1 "github.com/openshift/api/operator/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	. "github.com/netobserv/network-observability-operator/controllers/controllerstest"
)

// Because the simulated Kube server doesn't manage automatic resource cleanup like an actual Kube would do,
// we need either to cleanup all created resources manually, or to use different namespaces between tests
// For simplicity, we'll use a different namespace
const cpNamespace = "namespace-console-specs"

// nolint:cyclop
func flowCollectorConsolePluginSpecs() {
	cpKey := types.NamespacedName{
		Name:      "netobserv-plugin",
		Namespace: cpNamespace,
	}
	configKey := types.NamespacedName{
		Name:      "console-plugin-config",
		Namespace: cpNamespace,
	}
	crKey := types.NamespacedName{
		Name: "cluster",
	}
	consoleCRKey := types.NamespacedName{
		Name: "cluster",
	}

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
					Agent:           flowslatest.FlowCollectorAgent{Type: "IPFIX"},
					ConsolePlugin: flowslatest.FlowCollectorConsolePlugin{
						Port:            9001,
						ImagePullPolicy: "Never",
						Register:        true,
						Autoscaler: flowslatest.FlowCollectorHPA{
							Status:      flowslatest.HPAStatusEnabled,
							MinReplicas: pointer.Int32(1),
							MaxReplicas: 1,
							Metrics: []ascv2.MetricSpec{{
								Type: ascv2.ResourceMetricSourceType,
								Resource: &ascv2.ResourceMetricSource{
									Name: v1.ResourceCPU,
									Target: ascv2.MetricTarget{
										Type:               ascv2.UtilizationMetricType,
										AverageUtilization: pointer.Int32(90),
									},
								},
							}},
						},
						PortNaming: flowslatest.ConsolePluginPortConfig{
							Enable: true,
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
			By("Expecting to create the console plugin Deployment")
			Eventually(func() interface{} {
				dp := appsv1.Deployment{}
				if err := k8sClient.Get(ctx, cpKey, &dp); err != nil {
					return err
				}
				return *dp.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(1)))

			By("Expecting to create the console plugin Service")
			Eventually(func() interface{} {
				svc := v1.Service{}
				if err := k8sClient.Get(ctx, cpKey, &svc); err != nil {
					return err
				}
				return svc.Spec.Ports[0].Port
			}, timeout, interval).Should(Equal(int32(9001)))

			By("Creating the console plugin configmap")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, configKey, &ofc); err != nil {
					return err
				}
				return ofc.Data["config.yaml"]
			}, timeout, interval).Should(ContainSubstring("portNaming:\n  enable: true\n  portNames:\n    \"3100\": loki\nquickFilters:\n- name: Applications"))
		})

		It("Should update successfully", func() {
			Eventually(func() error {
				fc := flowslatest.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.ConsolePlugin.Port = 9099
				fc.Spec.ConsolePlugin.Replicas = 2
				fc.Spec.ConsolePlugin.Autoscaler.Status = flowslatest.HPAStatusDisabled
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting the console plugin Deployment to be scaled up")
			Eventually(func() interface{} {
				dp := appsv1.Deployment{}
				if err := k8sClient.Get(ctx, cpKey, &dp); err != nil {
					return err
				}
				return *dp.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(2)))

			By("Expecting the console plugin Service to be updated")
			Eventually(func() interface{} {
				svc := v1.Service{}
				if err := k8sClient.Get(ctx, cpKey, &svc); err != nil {
					return err
				}
				return svc.Spec.Ports[0].Port
			}, timeout, interval).Should(Equal(int32(9099)))
		})

		It("Should create desired objects when they're not found (e.g. case of an operator upgrade)", func() {
			sm := monitoringv1.ServiceMonitor{}

			By("Expecting ServiceMonitor to exist")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "netobserv-plugin",
					Namespace: cpNamespace,
				}, &sm)
			}, timeout, interval).Should(Succeed())

			// Manually delete ServiceMonitor
			By("Deleting ServiceMonitor")
			Eventually(func() error {
				return k8sClient.Delete(ctx, &sm)
			}, timeout, interval).Should(Succeed())

			// Do a dummy change that will trigger reconcile, and make sure SM is created again
			UpdateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Processor.LogLevel = "info"
			})
			By("Expecting ServiceMonitor to exist")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "netobserv-plugin",
					Namespace: cpNamespace,
				}, &sm)
			}, timeout, interval).Should(Succeed())
		})
	})

	Context("Configuring the Loki URL", func() {
		It("Should be initially configured with default Loki URL", func() {
			Eventually(getContainerArgumentAfter("netobserv-plugin", "-loki", cpKey),
				timeout, interval).Should(Equal("http://loki:3100/"))
		})
		It("Should update the Loki URL in the Console Plugin if it changes in the Spec", func() {
			UpdateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.URL = "http://loki.namespace:8888"
			})
			Eventually(getContainerArgumentAfter("netobserv-plugin", "-loki", cpKey),
				timeout, interval).Should(Equal("http://loki.namespace:8888"))
		})
		It("Should use the Loki Querier URL instead of the Loki URL, if the first is defined", func() {
			UpdateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.QuerierURL = "http://loki-querier:6789"
			})
			Eventually(getContainerArgumentAfter("netobserv-plugin", "-loki", cpKey),
				timeout, interval).Should(Equal("http://loki-querier:6789"))
		})
	})

	Context("Registering to the Console CR", func() {
		It("Should be registered by default", func() {
			Eventually(func() interface{} {
				cr := operatorsv1.Console{}
				if err := k8sClient.Get(ctx, consoleCRKey, &cr); err != nil {
					return err
				}
				return cr.Spec.Plugins
			}, timeout, interval).Should(Equal([]string{"netobserv-plugin"}))
		})

		It("Should be unregistered", func() {
			By("Update CR to unregister")
			UpdateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.ConsolePlugin.Register = false
			})

			By("Expecting the Console CR to not have plugin registered")
			Eventually(func() interface{} {
				cr := operatorsv1.Console{}
				if err := k8sClient.Get(ctx, consoleCRKey, &cr); err != nil {
					return err
				}
				return cr.Spec.Plugins
			}, timeout, interval).Should(BeEmpty())
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

		It("Should delete Console CR", func() {
			Eventually(func() error {
				return k8sClient.Delete(ctx, &operatorsv1.Console{
					ObjectMeta: metav1.ObjectMeta{
						Name: consoleCRKey.Name,
					},
				})
			}, timeout, interval).Should(Succeed())
		})

		It("Should be garbage collected", func() {
			By("Expecting console plugin deployment to be garbage collected")
			Eventually(func() interface{} {
				d := appsv1.Deployment{}
				_ = k8sClient.Get(ctx, cpKey, &d)
				return &d
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting console plugin service to be garbage collected")
			Eventually(func() interface{} {
				svc := v1.Service{}
				_ = k8sClient.Get(ctx, cpKey, &svc)
				return &svc
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting console plugin service account to be garbage collected")
			Eventually(func() interface{} {
				svcAcc := v1.ServiceAccount{}
				_ = k8sClient.Get(ctx, cpKey, &svcAcc)
				return &svcAcc
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))
		})
	})
}

func getContainerArgumentAfter(containerName, argName string, dpKey types.NamespacedName) func() interface{} {
	return func() interface{} {
		deployment := appsv1.Deployment{}
		if err := k8sClient.Get(ctx, dpKey, &deployment); err != nil {
			return err
		}
		for i := range deployment.Spec.Template.Spec.Containers {
			cnt := &deployment.Spec.Template.Spec.Containers[i]
			if cnt.Name == containerName {
				args := cnt.Args
				for len(args) > 0 {
					if args[0] == argName {
						if len(args) < 2 {
							return fmt.Errorf("container %q: arg %v has no value. Actual args: %v",
								containerName, argName, cnt.Args)
						}
						return args[1]
					}
					args = args[1:]
				}
				return fmt.Errorf("container %q: arg %v not found. Actual args: %v",
					containerName, argName, cnt.Args)
			}
		}
		return fmt.Errorf("container not found: %v", containerName)
	}
}
