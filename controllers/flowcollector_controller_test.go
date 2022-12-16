package controllers

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	. "github.com/netobserv/network-observability-operator/controllers/controllerstest"
	"github.com/netobserv/network-observability-operator/controllers/flowlogspipeline"
	"github.com/netobserv/network-observability-operator/pkg/conditions"
)

const timeout = time.Second * 10
const interval = 50 * time.Millisecond

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
	flpKey1 := types.NamespacedName{
		Name:      constants.FLPName,
		Namespace: operatorNamespace,
	}
	flpKey2 := types.NamespacedName{
		Name:      constants.FLPName,
		Namespace: otherNamespace,
	}
	flpKeyKafkaIngester := types.NamespacedName{
		Name:      constants.FLPName + flowlogspipeline.FlpConfSuffix[flowlogspipeline.ConfKafkaIngester],
		Namespace: operatorNamespace,
	}
	flpKeyKafkaTransformer := types.NamespacedName{
		Name:      constants.FLPName + flowlogspipeline.FlpConfSuffix[flowlogspipeline.ConfKafkaTransformer],
		Namespace: operatorNamespace,
	}
	cpKey1 := types.NamespacedName{
		Name:      "netobserv-plugin",
		Namespace: operatorNamespace,
	}
	cpKey2 := types.NamespacedName{
		Name:      "netobserv-plugin",
		Namespace: otherNamespace,
	}
	rbKeyIngest := types.NamespacedName{Name: flowlogspipeline.RoleBindingName(flowlogspipeline.ConfKafkaIngester)}
	rbKeyTransform := types.NamespacedName{Name: flowlogspipeline.RoleBindingName(flowlogspipeline.ConfKafkaTransformer)}
	rbKeyIngestMono := types.NamespacedName{Name: flowlogspipeline.RoleBindingMonoName(flowlogspipeline.ConfKafkaIngester)}
	rbKeyTransformMono := types.NamespacedName{Name: flowlogspipeline.RoleBindingMonoName(flowlogspipeline.ConfKafkaTransformer)}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Deploying as DaemonSet", func() {
		var digest string
		ds := appsv1.DaemonSet{}
		It("Should create successfully", func() {
			created := &flowsv1alpha1.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: crKey.Name,
				},
				Spec: flowsv1alpha1.FlowCollectorSpec{
					Namespace:       operatorNamespace,
					DeploymentModel: flowsv1alpha1.DeploymentModelDirect,
					Processor: flowsv1alpha1.FlowCollectorFLP{
						Port:            9999,
						ImagePullPolicy: "Never",
						LogLevel:        "error",
						Debug: flowsv1alpha1.DebugConfig{
							Env: map[string]string{
								"GOGC": "200",
							},
						},
					},
					Agent: flowsv1alpha1.FlowCollectorAgent{
						Type: "IPFIX",
						IPFIX: flowsv1alpha1.FlowCollectorIPFIX{
							Sampling: 200,
						},
					},
					ConsolePlugin: flowsv1alpha1.FlowCollectorConsolePlugin{
						Port:            9001,
						ImagePullPolicy: "Never",
						PortNaming: flowsv1alpha1.ConsolePluginPortConfig{
							Enable: true,
							PortNames: map[string]string{
								"3100": "loki",
							},
						},
					},
				},
			}

			// Create
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())

			By("Expecting to create the flowlogs-pipeline DaemonSet")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, flpKey1, &ds); err != nil {
					return err
				}
				digest = ds.Spec.Template.Annotations[flowlogspipeline.PodConfigurationDigest]
				if digest == "" {
					return fmt.Errorf("%q annotation can't be empty", flowlogspipeline.PodConfigurationDigest)
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("Expecting to create the flowlogs-pipeline ServiceAccount")
			Eventually(func() interface{} {
				svcAcc := v1.ServiceAccount{}
				if err := k8sClient.Get(ctx, flpKey1, &svcAcc); err != nil {
					return err
				}
				return svcAcc
			}, timeout, interval).Should(Satisfy(func(svcAcc v1.ServiceAccount) bool {
				return svcAcc.Labels != nil && svcAcc.Labels["app"] == constants.FLPName
			}))

			By("Expecting to create two flowlogs-pipeline role binding")
			rb1 := rbacv1.ClusterRoleBinding{}
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, rbKeyIngestMono, &rb1)
			}, timeout, interval).Should(Succeed())
			Expect(rb1.Subjects).Should(HaveLen(1))
			Expect(rb1.Subjects[0].Name).Should(Equal("flowlogs-pipeline"))
			Expect(rb1.RoleRef.Name).Should(Equal("flowlogs-pipeline-ingester"))

			rb2 := rbacv1.ClusterRoleBinding{}
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, rbKeyTransformMono, &rb2)
			}, timeout, interval).Should(Succeed())
			Expect(rb2.Subjects).Should(HaveLen(1))
			Expect(rb2.Subjects[0].Name).Should(Equal("flowlogs-pipeline"))
			Expect(rb2.RoleRef.Name).Should(Equal("flowlogs-pipeline-transformer"))

			By("Not expecting transformer role binding")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, rbKeyIngest, &rbacv1.ClusterRoleBinding{})
			}, timeout, interval).Should(MatchError(`clusterrolebindings.rbac.authorization.k8s.io "flowlogs-pipeline-ingester-role" not found`))

			By("Not expecting ingester role binding")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, rbKeyTransform, &rbacv1.ClusterRoleBinding{})
			}, timeout, interval).Should(MatchError(`clusterrolebindings.rbac.authorization.k8s.io "flowlogs-pipeline-transformer-role" not found`))

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

			By("Expecting flowlogs-pipeline-config configmap to be created")
			Eventually(func() interface{} {
				cm := v1.ConfigMap{}
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "flowlogs-pipeline-config",
					Namespace: operatorNamespace,
				}, &cm)
			}, timeout, interval).Should(Succeed())
		})

		It("Should update successfully", func() {
			UpdateCR(crKey, func(fc *flowsv1alpha1.FlowCollector) {
				fc.Spec.Processor = flowsv1alpha1.FlowCollectorFLP{
					Port:            7891,
					ImagePullPolicy: "Never",
					LogLevel:        "error",
					Debug: flowsv1alpha1.DebugConfig{
						Env: map[string]string{
							// we'll test that env vars are sorted, to keep idempotency
							"GOMAXPROCS": "33",
							"GOGC":       "400",
						},
					},
				}
				fc.Spec.Loki = flowsv1alpha1.FlowCollectorLoki{}
				fc.Spec.Agent.IPFIX = flowsv1alpha1.FlowCollectorIPFIX{
					Sampling:           400,
					CacheActiveTimeout: "30s",
					CacheMaxFlows:      1000,
				}
			})

			By("CR updated", func() {
				Eventually(func() error {
					err := k8sClient.Get(ctx, flpKey1, &ds)
					if err != nil {
						return err
					}
					return checkDigestUpdate(&digest, ds.Spec.Template.Annotations)
				}, timeout, interval).Should(Succeed())
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

			By("Creating the required HostPort to access flowlogs-pipeline through the NodeIP", func() {
				var cnt *v1.Container
				for i := range ds.Spec.Template.Spec.Containers {
					if ds.Spec.Template.Spec.Containers[i].Name == constants.FLPName {
						cnt = &ds.Spec.Template.Spec.Containers[i]
						break
					}
				}
				Expect(cnt).ToNot(BeNil(), "can't find a container named", constants.FLPName)
				var cp *v1.ContainerPort
				for i := range cnt.Ports {
					if cnt.Ports[i].Name == constants.FLPPortName {
						cp = &cnt.Ports[i]
						break
					}
				}
				Expect(cp).
					ToNot(BeNil(), "can't find a container port named", constants.FLPPortName)
				Expect(*cp).To(Equal(v1.ContainerPort{
					Name:          constants.FLPPortName,
					HostPort:      7891,
					ContainerPort: 7891,
					Protocol:      "UDP",
				}))
				Expect(cnt.Env).To(Equal([]v1.EnvVar{
					{Name: "GOGC", Value: "400"}, {Name: "GOMAXPROCS", Value: "33"},
				}))
			})

			By("Allocating the proper toleration to allow its placement in the master nodes", func() {
				Expect(ds.Spec.Template.Spec.Tolerations).
					To(ContainElement(v1.Toleration{Operator: v1.TolerationOpExists}))
			})
		})

		It("Should redeploy if the spec doesn't change but the external flowlogs-pipeline-config does", func() {
			UpdateCR(crKey, func(fc *flowsv1alpha1.FlowCollector) {
				fc.Spec.Loki.MaxRetries = 7
			})

			By("Expecting that the flowlogsPipeline.PodConfigurationDigest attribute has changed")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, flpKey1, &ds); err != nil {
					return err
				}
				return checkDigestUpdate(&digest, ds.Spec.Template.Annotations)
			}).Should(Succeed())
		})

		It("Should prevent undesired sampling-everything", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.Agent.IPFIX.Sampling = 1
				return k8sClient.Update(ctx, &fc)
			}).Should(Satisfy(func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "spec.agent.ipfix.sampling: Invalid value: 1")
			}), "Error expected for invalid sampling value")

			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
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

	Context("With Kafka", func() {
		It("Should update kafka config successfully", func() {
			UpdateCR(crKey, func(fc *flowsv1alpha1.FlowCollector) {
				fc.Spec.DeploymentModel = flowsv1alpha1.DeploymentModelKafka
				fc.Spec.Kafka = flowsv1alpha1.FlowCollectorKafka{
					Address: "localhost:9092",
					Topic:   "FLP",
					TLS: flowsv1alpha1.ClientTLS{
						CACert: flowsv1alpha1.CertificateReference{
							Type:     "secret",
							Name:     "some-secret",
							CertFile: "ca.crt",
						},
					},
				}
			})
		})

		It("Should deploy kafka ingester and transformer", func() {
			By("Expecting ingester daemonset to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKeyKafkaIngester, &appsv1.DaemonSet{})
			}, timeout, interval).Should(Succeed())

			By("Expecting transformer deployment to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKeyKafkaTransformer, &appsv1.Deployment{})
			}, timeout, interval).Should(Succeed())

			By("Not Expecting transformer service to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKeyKafkaTransformer, &v1.Service{})
			}, timeout, interval).Should(MatchError(`services "flowlogs-pipeline-transformer" not found`))

			By("Expecting to create two different flowlogs-pipeline role bindings")
			rb1 := rbacv1.ClusterRoleBinding{}
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, rbKeyIngest, &rb1)
			}, timeout, interval).Should(Succeed())
			Expect(rb1.Subjects).Should(HaveLen(1))
			Expect(rb1.Subjects[0].Name).Should(Equal("flowlogs-pipeline-ingester"))
			Expect(rb1.RoleRef.Name).Should(Equal("flowlogs-pipeline-ingester"))

			rb2 := rbacv1.ClusterRoleBinding{}
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, rbKeyTransform, &rb2)
			}, timeout, interval).Should(Succeed())
			Expect(rb2.Subjects).Should(HaveLen(1))
			Expect(rb2.Subjects[0].Name).Should(Equal("flowlogs-pipeline-transformer"))
			Expect(rb2.RoleRef.Name).Should(Equal("flowlogs-pipeline-transformer"))

			By("Not expecting mono-transformer role binding")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, rbKeyIngestMono, &rbacv1.ClusterRoleBinding{})
			}, timeout, interval).Should(MatchError(`clusterrolebindings.rbac.authorization.k8s.io "flowlogs-pipeline-ingester-role-mono" not found`))

			By("Not expecting mono-ingester role binding")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, rbKeyTransformMono, &rbacv1.ClusterRoleBinding{})
			}, timeout, interval).Should(MatchError(`clusterrolebindings.rbac.authorization.k8s.io "flowlogs-pipeline-transformer-role-mono" not found`))
		})

		It("Should delete previous flp deployment", func() {
			By("Expecting monolith to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKey1, &appsv1.DaemonSet{})
			}, timeout, interval).Should(MatchError(`daemonsets.apps "flowlogs-pipeline" not found`))
		})
	})

	Context("Adding auto-scaling", func() {
		hpa := ascv2.HorizontalPodAutoscaler{}
		It("Should update with HPA", func() {
			UpdateCR(crKey, func(fc *flowsv1alpha1.FlowCollector) {
				fc.Spec.Processor.KafkaConsumerAutoscaler = flowsv1alpha1.FlowCollectorHPA{
					Status:      flowsv1alpha1.HPAStatusEnabled,
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
				}
			})
		})

		It("Should have HPA installed", func() {
			By("Expecting HPA to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKeyKafkaTransformer, &hpa)
			}, timeout, interval).Should(Succeed())
			Expect(*hpa.Spec.MinReplicas).To(Equal(int32(1)))
			Expect(hpa.Spec.MaxReplicas).To(Equal(int32(1)))
			Expect(*hpa.Spec.Metrics[0].Resource.Target.AverageUtilization).To(Equal(int32(90)))
		})

		It("Should autoscale when the HPA options change", func() {
			UpdateCR(crKey, func(fc *flowsv1alpha1.FlowCollector) {
				fc.Spec.Processor.KafkaConsumerAutoscaler.MinReplicas = pointer.Int32(2)
				fc.Spec.Processor.KafkaConsumerAutoscaler.MaxReplicas = 2
			})

			By("Changing the Horizontal Pod Autoscaler instance")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, flpKeyKafkaTransformer, &hpa); err != nil {
					return err
				}
				if *hpa.Spec.MinReplicas != int32(2) || hpa.Spec.MaxReplicas != int32(2) ||
					*hpa.Spec.Metrics[0].Resource.Target.AverageUtilization != int32(90) {
					return fmt.Errorf("expected {2, 2, 90}: Got %v, %v, %v",
						*hpa.Spec.MinReplicas, hpa.Spec.MaxReplicas,
						*hpa.Spec.Metrics[0].Resource.Target.AverageUtilization)
				}
				return nil
			}, timeout, interval).Should(Succeed())
		})
	})

	Context("Back without Kafka", func() {
		It("Should remove kafka config successfully", func() {
			UpdateCR(crKey, func(fc *flowsv1alpha1.FlowCollector) {
				fc.Spec.DeploymentModel = flowsv1alpha1.DeploymentModelDirect
			})
		})

		It("Should deploy single flp again", func() {
			By("Expecting daemonset to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKey1, &appsv1.DaemonSet{})
			}, timeout, interval).Should(Succeed())
		})

		It("Should delete kafka ingester and transformer", func() {
			By("Expecting ingester daemonset to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKeyKafkaIngester, &appsv1.DaemonSet{})
			}, timeout, interval).Should(MatchError(`daemonsets.apps "flowlogs-pipeline-ingester" not found`))

			By("Expecting transformer deployment to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKeyKafkaTransformer, &appsv1.Deployment{})
			}, timeout, interval).Should(MatchError(`deployments.apps "flowlogs-pipeline-transformer" not found`))
		})
	})

	Context("Changing namespace", func() {
		It("Should update namespace successfully", func() {
			UpdateCR(crKey, func(fc *flowsv1alpha1.FlowCollector) {
				fc.Spec.Processor.Port = 9999
				fc.Spec.Namespace = otherNamespace
				fc.Spec.Agent.IPFIX = flowsv1alpha1.FlowCollectorIPFIX{
					Sampling: 200,
				}
			})
		})

		It("Should redeploy FLP in new namespace", func() {
			By("Expecting daemonset in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKey1, &appsv1.DaemonSet{})
			}, timeout, interval).Should(MatchError(`daemonsets.apps "flowlogs-pipeline" not found`))

			By("Expecting deployment in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKey1, &appsv1.Deployment{})
			}, timeout, interval).Should(MatchError(`deployments.apps "flowlogs-pipeline" not found`))

			By("Expecting service account in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKey1, &v1.ServiceAccount{})
			}, timeout, interval).Should(MatchError(`serviceaccounts "flowlogs-pipeline" not found`))

			By("Expecting daemonset to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKey2, &appsv1.DaemonSet{})
			}, timeout, interval).Should(Succeed())

			By("Expecting service account to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpKey2, &v1.ServiceAccount{})
			}, timeout, interval).Should(Succeed())
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
		flowCR := flowsv1alpha1.FlowCollector{}
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
			By("Expecting flowlogs-pipeline daemonset to be garbage collected")
			Eventually(func() interface{} {
				d := appsv1.DaemonSet{}
				_ = k8sClient.Get(ctx, flpKey2, &d)
				return &d
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting flowlogs-pipeline service account to be garbage collected")
			Eventually(func() interface{} {
				svcAcc := v1.ServiceAccount{}
				_ = k8sClient.Get(ctx, flpKey2, &svcAcc)
				return &svcAcc
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

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

			By("Expecting flowlogs-pipeline configmap to be garbage collected")
			Eventually(func() interface{} {
				cm := v1.ConfigMap{}
				_ = k8sClient.Get(ctx, types.NamespacedName{
					Name:      "flowlogs-pipeline-config",
					Namespace: otherNamespace,
				}, &cm)
				return &cm
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))
		})

		It("Should not get CR", func() {
			Eventually(func() bool {
				err := k8sClient.Get(ctx, crKey, &flowCR)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
}

func GetReadyCR(key types.NamespacedName) *flowsv1alpha1.FlowCollector {
	cr := flowsv1alpha1.FlowCollector{}
	Eventually(func() error {
		err := k8sClient.Get(ctx, key, &cr)
		if err != nil {
			return err
		}
		cond := meta.FindStatusCondition(cr.Status.Conditions, conditions.TypeReady)
		if cond.Status == metav1.ConditionFalse {
			return fmt.Errorf("CR is not ready: %s - %v", cond.Reason, cond.Message)
		}
		return nil
	}).Should(Succeed())
	return &cr
}

func UpdateCR(key types.NamespacedName, updater func(*flowsv1alpha1.FlowCollector)) {
	cr := GetReadyCR(key)
	Eventually(func() error {
		updater(cr)
		return k8sClient.Update(ctx, cr)
	}).Should(Succeed())
}

func checkDigestUpdate(oldDigest *string, annots map[string]string) error {
	newDigest := annots[flowlogspipeline.PodConfigurationDigest]
	if newDigest == "" {
		return fmt.Errorf("%q annotation can't be empty", flowlogspipeline.PodConfigurationDigest)
	} else if newDigest == *oldDigest {
		return fmt.Errorf("expect digest to change, but is still %s", *oldDigest)
	}
	*oldDigest = newDigest
	return nil
}
