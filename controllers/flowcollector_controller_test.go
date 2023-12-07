package controllers

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	. "github.com/netobserv/network-observability-operator/controllers/controllerstest"
	"github.com/netobserv/network-observability-operator/pkg/test"
)

const (
	timeout                     = test.Timeout
	interval                    = test.Interval
	conntrackEndTimeout         = 10 * time.Second
	conntrackTerminatingTimeout = 5 * time.Second
	conntrackHeartbeatInterval  = 30 * time.Second
)

var outputRecordTypes = flowslatest.LogTypeAll

var updateCR = func(key types.NamespacedName, updater func(*flowslatest.FlowCollector)) {
	test.UpdateCR(ctx, k8sClient, key, updater)
}

// nolint:cyclop
func flowCollectorControllerSpecs() {

	const operatorNamespace = "main-namespace"
	const otherNamespace = "other-namespace"
	crKey := types.NamespacedName{
		Name: "cluster",
	}
	ovsConfigMapKey := types.NamespacedName{
		Name:      "ovs-flows-config",
		Namespace: "openshift-network-operator",
	}
	cpKey1 := types.NamespacedName{
		Name:      "netobserv-plugin",
		Namespace: operatorNamespace,
	}
	cpKey2 := types.NamespacedName{
		Name:      "netobserv-plugin",
		Namespace: otherNamespace,
	}
	rbKeyPlugin := types.NamespacedName{Name: constants.PluginName}

	// Created objects to cleanup
	cleanupList := []client.Object{}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Without Kafka", func() {
		It("Should create successfully", func() {
			created := &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: crKey.Name,
				},
				Spec: flowslatest.FlowCollectorSpec{
					Namespace:       operatorNamespace,
					DeploymentModel: flowslatest.DeploymentModelDirect,
					Agent: flowslatest.FlowCollectorAgent{
						Type: "IPFIX",
						IPFIX: flowslatest.FlowCollectorIPFIX{
							Sampling: 200,
						},
					},
					Processor: flowslatest.FlowCollectorFLP{
						Port: 9999,
					},
					ConsolePlugin: flowslatest.FlowCollectorConsolePlugin{
						Enable:          ptr.To(true),
						Port:            9001,
						ImagePullPolicy: "Never",
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
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())

			By("Expecting to create console plugin role binding")
			rb3 := rbacv1.ClusterRoleBinding{}
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, rbKeyPlugin, &rb3)
			}, timeout, interval).Should(Succeed())
			Expect(rb3.Subjects).Should(HaveLen(1))
			Expect(rb3.Subjects[0].Name).Should(Equal("netobserv-plugin"))
			Expect(rb3.RoleRef.Name).Should(Equal("netobserv-plugin"))

			By("Creating the ovn-flows-configmap with the configuration from the FlowCollector")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, ovsConfigMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "200",
				"nodePort":           "9999",
				"cacheMaxFlows":      "400",
				"cacheActiveTimeout": "20s",
			}))
		})

		It("Should update successfully", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Processor = flowslatest.FlowCollectorFLP{
					Port: 7891,
				}
				fc.Spec.Loki = flowslatest.FlowCollectorLoki{}
				fc.Spec.Agent.IPFIX = flowslatest.FlowCollectorIPFIX{
					Sampling:           400,
					CacheActiveTimeout: "30s",
					CacheMaxFlows:      1000,
				}
			})

			By("Expecting to create the ovn-flows-configmap with the configuration from the FlowCollector", func() {
				Eventually(func() interface{} {
					ofc := v1.ConfigMap{}
					if err := k8sClient.Get(ctx, ovsConfigMapKey, &ofc); err != nil {
						return err
					}
					return ofc.Data
				}, timeout, interval).Should(Equal(map[string]string{
					"sampling":           "400",
					"nodePort":           "7891",
					"cacheMaxFlows":      "1000",
					"cacheActiveTimeout": "30s",
				}))
			})
		})

		It("Should prevent undesired sampling-everything", func() {
			Eventually(func() error {
				fc := flowslatest.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.Agent.IPFIX.Sampling = 1
				return k8sClient.Update(ctx, &fc)
			}).Should(Satisfy(func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "spec.agent.ipfix.sampling: Invalid value: 1")
			}), "Error expected for invalid sampling value")

			Eventually(func() error {
				fc := flowslatest.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.Agent.IPFIX.Sampling = 10
				fc.Spec.Agent.IPFIX.ForceSampleAll = true
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting that ovn-flows-configmap is updated with sampling=1")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, ovsConfigMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data["sampling"]
			}, timeout, interval).Should(Equal("1"))
		})
	})

	Context("Changing namespace", func() {
		It("Should update namespace successfully", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Namespace = otherNamespace
			})
		})

		It("Should redeploy console plugin in new namespace", func() {
			By("Expecting deployment in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cpKey1, &appsv1.Deployment{})
			}, timeout, interval).Should(MatchError(`deployments.apps "netobserv-plugin" not found`))

			By("Expecting service in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cpKey1, &v1.Service{})
			}, timeout, interval).Should(MatchError(`services "netobserv-plugin" not found`))

			By("Expecting service account in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cpKey1, &v1.ServiceAccount{})
			}, timeout, interval).Should(MatchError(`serviceaccounts "netobserv-plugin" not found`))

			By("Expecting deployment to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cpKey2, &appsv1.Deployment{})
			}, timeout, interval).Should(Succeed())

			By("Expecting service to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cpKey2, &v1.Service{})
			}, timeout, interval).Should(Succeed())

			By("Expecting service account to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cpKey2, &v1.ServiceAccount{})
			}, timeout, interval).Should(Succeed())
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
			By("Expecting console plugin deployment to be garbage collected")
			Eventually(func() interface{} {
				d := appsv1.Deployment{}
				_ = k8sClient.Get(ctx, cpKey2, &d)
				return &d
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting console plugin service to be garbage collected")
			Eventually(func() interface{} {
				svc := v1.Service{}
				_ = k8sClient.Get(ctx, cpKey2, &svc)
				return &svc
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting console plugin service account to be garbage collected")
			Eventually(func() interface{} {
				svcAcc := v1.ServiceAccount{}
				_ = k8sClient.Get(ctx, cpKey2, &svcAcc)
				return &svcAcc
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting ovn-flows-configmap to be garbage collected")
			Eventually(func() interface{} {
				cm := v1.ConfigMap{}
				_ = k8sClient.Get(ctx, ovsConfigMapKey, &cm)
				return &cm
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))
		})

		It("Should not get CR", func() {
			Eventually(func() bool {
				err := k8sClient.Get(ctx, crKey, &flowCR)
				return kerr.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})

		It("Should cleanup other data", func() {
			for _, obj := range cleanupList {
				Eventually(func() error {
					return k8sClient.Delete(ctx, obj)
				}, timeout, interval).Should(Succeed())
			}
		})
	})
}
