package controllers

import (
	"fmt"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	. "github.com/netobserv/network-observability-operator/controllers/controllerstest"
	"github.com/netobserv/network-observability-operator/controllers/flowlogspipeline"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

const timeout = time.Second * 10
const interval = 50 * time.Millisecond

// nolint:cyclop
func flowCollectorControllerSpecs() {
	const operatorNamespace = "main-namespace"
	const otherNamespace = "other-namespace"
	ipResolver.On("LookupIP", constants.FLPName+"."+operatorNamespace).
		Return([]net.IP{net.IPv4(11, 22, 33, 44)}, nil)
	ipResolver.On("LookupIP", constants.FLPName+"."+otherNamespace).
		Return([]net.IP{net.IPv4(111, 122, 133, 144)}, nil)
	crKey := types.NamespacedName{
		Name: "cluster",
	}
	ovsConfigMapKey := types.NamespacedName{
		Name:      "ovs-flows-config",
		Namespace: "openshift-network-operator",
	}
	gfKey1 := types.NamespacedName{
		Name:      constants.FLPName,
		Namespace: operatorNamespace,
	}
	gfKey2 := types.NamespacedName{
		Name:      constants.FLPName,
		Namespace: otherNamespace,
	}
	gfKeyKafkaIngestor := types.NamespacedName{
		Name:      constants.FLPName + flowlogspipeline.FlpConfSuffix[flowlogspipeline.ConfKafkaIngestor],
		Namespace: operatorNamespace,
	}
	gfKeyKafkaTransformer := types.NamespacedName{
		Name:      constants.FLPName + flowlogspipeline.FlpConfSuffix[flowlogspipeline.ConfKafkaTransformer],
		Namespace: operatorNamespace,
	}
	cpKey1 := types.NamespacedName{
		Name:      "network-observability-plugin",
		Namespace: operatorNamespace,
	}
	cpKey2 := types.NamespacedName{
		Name:      "network-observability-plugin",
		Namespace: otherNamespace,
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
	Context("Deployment with autho-scaling", func() {
		var oldDigest string
		It("Should create successfully", func() {

			created := &flowsv1alpha1.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: crKey.Name,
				},
				Spec: flowsv1alpha1.FlowCollectorSpec{
					Namespace: operatorNamespace,
					FlowlogsPipeline: flowsv1alpha1.FlowCollectorFLP{
						Kind:            "Deployment",
						Port:            9999,
						ImagePullPolicy: "Never",
						LogLevel:        "error",
						Image:           "testimg:latest",
						HPA: &flowsv1alpha1.FlowCollectorHPA{
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
					},
					Agent: "ipfix",
					IPFIX: flowsv1alpha1.FlowCollectorIPFIX{
						Sampling: 200,
					},
					ConsolePlugin: flowsv1alpha1.FlowCollectorConsolePlugin{
						Port:            9001,
						ImagePullPolicy: "Never",
						Image:           "testimg:latest",
						HPA: &flowsv1alpha1.FlowCollectorHPA{
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

			By("Expecting to create the flowlogs-pipeline Deployment")
			Eventually(func() interface{} {
				dp := appsv1.Deployment{}
				if err := k8sClient.Get(ctx, gfKey1, &dp); err != nil {
					return err
				}
				oldDigest = dp.Spec.Template.Annotations[flowlogspipeline.PodConfigurationDigest]
				if oldDigest == "" {
					return fmt.Errorf("%q annotation can't be empty", flowlogspipeline.PodConfigurationDigest)
				}

				return *dp.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(1)))

			svc := v1.Service{}
			By("Expecting to create the flowlogs-pipeline Service")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, gfKey1, &svc); err != nil {
					return err
				}
				return svc
			}, timeout, interval).Should(Satisfy(func(svc v1.Service) bool {
				return svc.Labels != nil && svc.Labels["app"] == constants.FLPName &&
					svc.Spec.Selector != nil && svc.Spec.Selector["app"] == constants.FLPName &&
					len(svc.Spec.Ports) == 1 &&
					svc.Spec.Ports[0].Protocol == v1.ProtocolUDP &&
					svc.Spec.Ports[0].Port == 9999
			}), "unexpected service contents", helper.AsyncJSON{Ptr: svc})

			By("Expecting to create the flowlogs-pipeline ServiceAccount")
			Eventually(func() interface{} {
				svcAcc := v1.ServiceAccount{}
				if err := k8sClient.Get(ctx, gfKey1, &svcAcc); err != nil {
					return err
				}
				return svcAcc
			}, timeout, interval).Should(Satisfy(func(svcAcc v1.ServiceAccount) bool {
				return svcAcc.Labels != nil && svcAcc.Labels["app"] == constants.FLPName
			}))

			By("Creating the ovn-flows-configmap with the configuration from the FlowCollector")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, ovsConfigMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "200",
				"sharedTarget":       "11.22.33.44:9999",
				"cacheMaxFlows":      "100",
				"cacheActiveTimeout": "60s",
			}))
		})

		It("Should update successfully", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.IPFIX.CacheActiveTimeout = "30s"
				fc.Spec.IPFIX.Sampling = 1234
				fc.Spec.FlowlogsPipeline.Port = 1999
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting updated flowlogs-pipeline Service port")
			Eventually(func() interface{} {
				svc := v1.Service{}
				if err := k8sClient.Get(ctx, gfKey1, &svc); err != nil {
					return err
				}
				return svc.Spec.Ports[0].Port
			}, timeout, interval).Should(Equal(int32(1999)))

			By("Expecting that ovn-flows-configmap is updated accordingly")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, ovsConfigMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "1234",
				"sharedTarget":       "11.22.33.44:1999",
				"cacheMaxFlows":      "100",
				"cacheActiveTimeout": "30s",
			}))
		})

		It("Should redeploy if the spec doesn't change but the external flowlogs-pipeline-config does", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.Loki.MaxRetries = 7
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting that the flowlogsPipeline.PodConfigurationDigest attribute has changed")
			Eventually(func() error {
				dp := appsv1.Deployment{}
				if err := k8sClient.Get(ctx, gfKey1, &dp); err != nil {
					return err
				}
				currentConfigDigest := dp.Spec.Template.Annotations[flowlogspipeline.PodConfigurationDigest]
				if currentConfigDigest == oldDigest {
					return fmt.Errorf("annotation %v %q was expected to change",
						flowlogspipeline.PodConfigurationDigest, currentConfigDigest)
				}
				return nil
			}).Should(Succeed())
		})

		It("Should autoscale when the HPA options change", func() {
			hpa := ascv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, gfKey1, &hpa)).To(Succeed())
			Expect(*hpa.Spec.MinReplicas).To(Equal(int32(1)))
			Expect(hpa.Spec.MaxReplicas).To(Equal(int32(1)))
			Expect(*hpa.Spec.Metrics[0].Resource.Target.AverageUtilization).To(Equal(int32(90)))
			// update FlowCollector and verify that HPA spec also changed
			fc := flowsv1alpha1.FlowCollector{}
			Expect(k8sClient.Get(ctx, crKey, &fc)).To(Succeed())
			fc.Spec.FlowlogsPipeline.HPA.MinReplicas = pointer.Int32(2)
			fc.Spec.FlowlogsPipeline.HPA.MaxReplicas = 2
			Expect(k8sClient.Update(ctx, &fc)).To(Succeed())

			By("Changing the Horizontal Pod Autoscaler instance")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, gfKey1, &hpa); err != nil {
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

	Context("Deploying as DaemonSet", func() {
		var oldConfigDigest string
		It("Should update successfully", func() {
			fc := flowsv1alpha1.FlowCollector{}
			Expect(k8sClient.Get(ctx, crKey, &fc)).Should(Succeed())
			fc.Spec.FlowlogsPipeline = flowsv1alpha1.FlowCollectorFLP{
				Kind:            "DaemonSet",
				Port:            7891,
				ImagePullPolicy: "Never",
				LogLevel:        "error",
				Image:           "testimg:latest",
			}
			fc.Spec.Loki = flowsv1alpha1.FlowCollectorLoki{}
			fc.Spec.IPFIX = flowsv1alpha1.FlowCollectorIPFIX{
				Sampling: 200,
			}
			// Update
			Expect(k8sClient.Update(ctx, &fc)).Should(Succeed())

			By("Expecting to create the ovn-flows-configmap with the configuration from the FlowCollector")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, ovsConfigMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "200",
				"nodePort":           "7891",
				"cacheMaxFlows":      "100",
				"cacheActiveTimeout": "60s",
			}))

			ds := appsv1.DaemonSet{}
			Expect(k8sClient.Get(ctx, gfKey1, &ds)).To(Succeed())

			oldConfigDigest = ds.Spec.Template.Annotations[flowlogspipeline.PodConfigurationDigest]
			Expect(oldConfigDigest).ToNot(BeEmpty())

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
			})

			By("Allocating the proper toleration to allow its placement in the master nodes", func() {
				Expect(ds.Spec.Template.Spec.Tolerations).
					To(ContainElement(v1.Toleration{Operator: v1.TolerationOpExists}))
			})
		})
		It("Should redeploy if the spec doesn't change but the external flowlogs-pipeline-config does", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.Loki.MaxRetries = 7
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting that the flowlogsPipeline.PodConfigurationDigest attribute has changed")
			Eventually(func() error {
				dp := appsv1.DaemonSet{}
				if err := k8sClient.Get(ctx, gfKey1, &dp); err != nil {
					return err
				}
				currentConfigDigest := dp.Spec.Template.Annotations[flowlogspipeline.PodConfigurationDigest]
				if currentConfigDigest == oldConfigDigest {
					return fmt.Errorf("annotation %v %q was expected to change (was %q)",
						flowlogspipeline.PodConfigurationDigest, currentConfigDigest, oldConfigDigest)
				}
				return nil
			}).Should(Succeed())
		})
	})

	Context("Changing kafka config", func() {
		It("Should update kafka config successfully", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.Kafka = flowsv1alpha1.FlowCollectorKafka{Enable: true, Address: "loaclhost:9092", Topic: "FLP"}
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())
		})

		It("Should deploy kafka ingestor and transformer", func() {
			By("Expecting ingestor daemonset to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKeyKafkaIngestor, &appsv1.DaemonSet{})
			}, timeout, interval).Should(Succeed())

			By("Expecting transformer deployment to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKeyKafkaTransformer, &appsv1.Deployment{})
			}, timeout, interval).Should(Succeed())

			By("Not Expecting transformer service to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKeyKafkaTransformer, &v1.Service{})
			}, timeout, interval).Should(MatchError(`services "flowlogs-pipeline-ktransform" not found`))
		})

		It("Should delete previous flp deployment", func() {
			By("Expecting deployment to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey1, &appsv1.DaemonSet{})
			}, timeout, interval).Should(MatchError(`daemonsets.apps "flowlogs-pipeline" not found`))

			By("Expecting service to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey1, &v1.Service{})
			}, timeout, interval).Should(MatchError(`services "flowlogs-pipeline" not found`))
		})

		It("Should remove kafka config successfully", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.Kafka.Enable = false
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())
		})

		It("Should deploy single flp again", func() {
			By("Expecting daemonset to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey1, &appsv1.DaemonSet{})
			}, timeout, interval).Should(Succeed())
		})

		It("Should delete kafka ingestor and transformer", func() {
			By("Expecting ingestor daemonset to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKeyKafkaIngestor, &appsv1.DaemonSet{})
			}, timeout, interval).Should(MatchError(`daemonsets.apps "flowlogs-pipeline-kingestor" not found`))

			By("Expecting transformer deployment to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKeyKafkaTransformer, &appsv1.Deployment{})
			}, timeout, interval).Should(MatchError(`deployments.apps "flowlogs-pipeline-ktransform" not found`))
		})

	})

	Context("Changing namespace", func() {
		It("Should update namespace successfully", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, crKey, &fc); err != nil {
					return err
				}
				fc.Spec.FlowlogsPipeline.Kind = "Deployment"
				fc.Spec.FlowlogsPipeline.Port = 9999
				fc.Spec.Namespace = otherNamespace
				fc.Spec.IPFIX = flowsv1alpha1.FlowCollectorIPFIX{
					Sampling: 200,
				}
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())
		})

		It("Should redeploy goglow-kube in new namespace", func() {
			By("Expecting daemonset in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey1, &appsv1.DaemonSet{})
			}, timeout, interval).Should(MatchError(`daemonsets.apps "flowlogs-pipeline" not found`))

			By("Expecting deployment in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey1, &appsv1.Deployment{})
			}, timeout, interval).Should(MatchError(`deployments.apps "flowlogs-pipeline" not found`))

			By("Expecting service in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey1, &v1.Service{})
			}, timeout, interval).Should(MatchError(`services "flowlogs-pipeline" not found`))

			By("Expecting service account in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey1, &v1.ServiceAccount{})
			}, timeout, interval).Should(MatchError(`serviceaccounts "flowlogs-pipeline" not found`))

			By("Expecting deployment to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey2, &appsv1.Deployment{})
			}, timeout, interval).Should(Succeed())

			By("Expecting service to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey2, &v1.Service{})
			}, timeout, interval).Should(Succeed())

			By("Expecting service account to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, gfKey2, &v1.ServiceAccount{})
			}, timeout, interval).Should(Succeed())
		})

		It("Should update ovn-flows-configmap with new IP", func() {
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, ovsConfigMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "200",
				"sharedTarget":       "111.122.133.144:9999",
				"cacheMaxFlows":      "100",
				"cacheActiveTimeout": "60s",
			}))
		})

		It("Should redeploy console plugin in new namespace", func() {
			By("Expecting deployment in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cpKey1, &appsv1.Deployment{})
			}, timeout, interval).Should(MatchError(`deployments.apps "network-observability-plugin" not found`))

			By("Expecting service in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cpKey1, &v1.Service{})
			}, timeout, interval).Should(MatchError(`services "network-observability-plugin" not found`))

			By("Expecting service account in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, cpKey1, &v1.ServiceAccount{})
			}, timeout, interval).Should(MatchError(`serviceaccounts "network-observability-plugin" not found`))

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
			By("Expecting flowlogs-pipeline deployment to be garbage collected")
			Eventually(func() interface{} {
				d := appsv1.Deployment{}
				_ = k8sClient.Get(ctx, gfKey2, &d)
				return &d
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting flowlogs-pipeline service to be garbage collected")
			Eventually(func() interface{} {
				svc := v1.Service{}
				_ = k8sClient.Get(ctx, gfKey2, &svc)
				return &svc
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting flowlogs-pipeline service account to be garbage collected")
			Eventually(func() interface{} {
				svcAcc := v1.ServiceAccount{}
				_ = k8sClient.Get(ctx, gfKey2, &svcAcc)
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
	})
}
