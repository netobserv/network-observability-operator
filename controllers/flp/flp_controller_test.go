package flp

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
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
	outputRecordTypes = flowslatest.LogTypeAll
	updateCR          = func(key types.NamespacedName, updater func(*flowslatest.FlowCollector)) {
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
	expectNoCreation = func(namespace string, objs ...test.ResourceRef) {
		GinkgoHelper()
		test.ExpectNoCreation(ctx, k8sClient, namespace, objs...)
	}
	expectOwnership = func(namespace string, objs ...test.ResourceRef) {
		GinkgoHelper()
		test.ExpectOwnership(ctx, k8sClient, namespace, objs...)
	}
)

// nolint:cyclop
func ControllerSpecs() {
	const operatorNamespace = "main-namespace"
	const otherNamespace = "other-namespace"
	crKey := types.NamespacedName{
		Name: "cluster",
	}
	deplRef := test.Deployment(constants.FLPName)
	cmRef := test.ConfigMap(constants.FLPName + "-config")
	saRef := test.ServiceAccount(constants.FLPName)
	crbRef := test.ClusterRoleBinding(constants.FLPName)

	// Created objects to cleanup
	cleanupList := []client.Object{}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Direct mode / direct-flp", func() {
		It("Should create CR successfully", func() {
			created := &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
				Spec: flowslatest.FlowCollectorSpec{
					Namespace:       operatorNamespace,
					DeploymentModel: flowslatest.DeploymentModelDirect,
					Processor: flowslatest.FlowCollectorFLP{
						ImagePullPolicy: "Never",
						LogLevel:        "error",
						Advanced: &flowslatest.AdvancedProcessorConfig{
							Env: map[string]string{
								"GOGC": "200",
							},
							ConversationHeartbeatInterval: &metav1.Duration{
								Duration: conntrackHeartbeatInterval,
							},
							ConversationEndTimeout: &metav1.Duration{
								Duration: conntrackEndTimeout,
							},
							ConversationTerminatingTimeout: &metav1.Duration{
								Duration: conntrackTerminatingTimeout,
							},
						},
						LogTypes: &outputRecordTypes,
						Metrics: flowslatest.FLPMetrics{
							IncludeList: &[]flowslatest.FLPMetric{"node_ingress_bytes_total", "namespace_ingress_bytes_total", "workload_ingress_bytes_total"},
						},
					},
				},
			}

			// Create
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())
		})

		It("Should not create flowlogs-pipeline when using agent direct-flp", func() {
			expectNoCreation(operatorNamespace,
				deplRef,
				cmRef,
				test.DaemonSet(constants.FLPName),
			)
		})
	})

	Context("With Kafka", func() {
		It("Should update kafka config successfully", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.DeploymentModel = flowslatest.DeploymentModelKafka
				fc.Spec.Kafka = flowslatest.FlowCollectorKafka{
					Address: "localhost:9092",
					Topic:   "FLP",
					TLS: flowslatest.ClientTLS{
						CACert: flowslatest.CertificateReference{
							Type:     "secret",
							Name:     "some-secret",
							CertFile: "ca.crt",
						},
					},
				}
			})
		})

		var depl *appsv1.Deployment
		var digest string

		It("Should deploy kafka transformer", func() {
			objs := expectCreation(operatorNamespace,
				deplRef,
				cmRef,
				saRef,
				crbRef,
			)
			Expect(objs).To(HaveLen(4))
			depl = objs[0].(*appsv1.Deployment)
			digest = depl.Spec.Template.Annotations[constants.PodConfigurationDigest]
			Expect(digest).NotTo(BeEmpty())

			rb := objs[3].(*rbacv1.ClusterRoleBinding)
			Expect(rb.Subjects).Should(HaveLen(1))
			Expect(rb.Subjects[0].Name).Should(Equal("flowlogs-pipeline"))
			Expect(rb.RoleRef.Name).Should(Equal("flowlogs-pipeline"))
		})

		It("Should update successfully", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Processor = flowslatest.FlowCollectorFLP{
					ImagePullPolicy: "Never",
					LogLevel:        "error",
					Advanced: &flowslatest.AdvancedProcessorConfig{
						Env: map[string]string{
							// we'll test that env vars are sorted, to keep idempotency
							"GOMAXPROCS": "33",
							"GOGC":       "400",
						},
						Port: ptr.To(int32(7891)),
						ConversationHeartbeatInterval: &metav1.Duration{
							Duration: conntrackHeartbeatInterval,
						},
						ConversationEndTimeout: &metav1.Duration{
							Duration: conntrackEndTimeout,
						},
						ConversationTerminatingTimeout: &metav1.Duration{
							Duration: conntrackTerminatingTimeout,
						},
					},
					LogTypes: &outputRecordTypes,
					Metrics: flowslatest.FLPMetrics{
						IncludeList:   &[]flowslatest.FLPMetric{"node_ingress_bytes_total"},
						DisableAlerts: []flowslatest.FLPAlert{flowslatest.AlertLokiError},
					},
				}
				fc.Spec.Loki = flowslatest.FlowCollectorLoki{}
			})

			By("CR updated", func() {
				Eventually(func() error {
					err := k8sClient.Get(ctx, deplRef.GetKey(operatorNamespace), depl)
					if err != nil {
						return err
					}
					return checkDigestUpdate(&digest, depl.Spec.Template.Annotations)
				}, timeout, interval).Should(Succeed())
			})
		})

		It("Should redeploy if the spec doesn't change but the external flowlogs-pipeline-config does", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Advanced = &flowslatest.AdvancedLokiConfig{
					WriteMaxRetries: ptr.To(int32(7)),
				}
			})

			By("Expecting that the flowlogsPipeline.PodConfigurationDigest attribute has changed")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, deplRef.GetKey(operatorNamespace), depl); err != nil {
					return err
				}
				return checkDigestUpdate(&digest, depl.Spec.Template.Annotations)
			}).Should(Succeed())
		})
	})

	Context("Adding auto-scaling", func() {
		hpa := ascv2.HorizontalPodAutoscaler{}
		It("Should update with HPA", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Processor.KafkaConsumerAutoscaler = flowslatest.FlowCollectorHPA{
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
				}
			})
		})

		It("Should have HPA installed", func() {
			By("Expecting HPA to be created")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, deplRef.GetKey(operatorNamespace), &hpa)
			}, timeout, interval).Should(Succeed())
			Expect(*hpa.Spec.MinReplicas).To(Equal(int32(1)))
			Expect(hpa.Spec.MaxReplicas).To(Equal(int32(1)))
			Expect(*hpa.Spec.Metrics[0].Resource.Target.AverageUtilization).To(Equal(int32(90)))
		})

		It("Should autoscale when the HPA options change", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Processor.KafkaConsumerAutoscaler.MinReplicas = ptr.To(int32(2))
				fc.Spec.Processor.KafkaConsumerAutoscaler.MaxReplicas = 2
			})

			By("Changing the Horizontal Pod Autoscaler instance")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, deplRef.GetKey(operatorNamespace), &hpa); err != nil {
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

	Context("Checking monitoring resources", func() {
		It("Should create desired objects when they're not found (e.g. case of an operator upgrade)", func() {
			objs := expectCreation(operatorNamespace,
				test.Service("flowlogs-pipeline-prom"),
				test.ServiceMonitor("flowlogs-pipeline-monitor"),
				test.PrometheusRule("flowlogs-pipeline-alert"),
			)
			Expect(objs).To(HaveLen(3))
			sm := objs[1].(*monitoringv1.ServiceMonitor)
			pr := objs[2].(*monitoringv1.PrometheusRule)
			Expect(pr.Spec.Groups).Should(HaveLen(1))
			Expect(pr.Spec.Groups[0].Rules).Should(HaveLen(1))

			// Manually delete ServiceMonitor
			By("Deleting ServiceMonitor")
			Eventually(func() error { return k8sClient.Delete(ctx, sm) }, timeout, interval).Should(Succeed())

			// Do a dummy change that will trigger reconcile, and make sure SM is created again
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Processor.LogLevel = "trace"
			})

			By("Expecting ServiceMonitor to exist")
			expectCreation(operatorNamespace, test.ServiceMonitor("flowlogs-pipeline-monitor"))

			// Manually delete Rule
			By("Deleting prom rule")
			Eventually(func() error { return k8sClient.Delete(ctx, pr) }, timeout, interval).Should(Succeed())

			// Do a dummy change that will trigger reconcile, and make sure Rule is created again
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Processor.LogLevel = "debug"
			})
			By("Expecting PrometheusRule to exist")
			expectCreation(operatorNamespace, test.PrometheusRule("flowlogs-pipeline-alert"))
		})
	})

	Context("Using certificates with loki manual mode", func() {
		flpKey := deplRef.GetKey(operatorNamespace)
		depl := appsv1.Deployment{}
		It("Should update Loki to use TLS", func() {
			// Create CM certificate
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "loki-ca",
					Namespace: operatorNamespace,
				},
				Data: map[string]string{"ca.crt": "certificate data"},
			}
			cleanupList = append(cleanupList, cm)
			Expect(k8sClient.Create(ctx, cm)).Should(Succeed())
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Mode = flowslatest.LokiModeManual
				fc.Spec.Loki.Manual.TLS = flowslatest.ClientTLS{
					Enable: true,
					CACert: flowslatest.CertificateReference{
						Type:     flowslatest.RefTypeConfigMap,
						Name:     "loki-ca",
						CertFile: "ca.crt",
					},
				}
			})
		})

		It("Should have certificate mounted", func() {
			By("Expecting certificate mounted")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &depl); err != nil {
					return err
				}
				return depl.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(2))
			Expect(depl.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
			Expect(depl.Spec.Template.Spec.Volumes[1].Name).To(Equal("loki-certs-ca"))
		})

		It("Should restore no TLS config", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Manual.TLS = flowslatest.ClientTLS{
					Enable: false,
				}
			})
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &depl); err != nil {
					return err
				}
				return depl.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(1))
			Expect(depl.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
		})
	})

	Context("Using certificates with loki distributed mode", func() {
		flpKey := deplRef.GetKey(operatorNamespace)
		depl := appsv1.Deployment{}
		It("Should update Loki to use TLS", func() {
			// Create CM certificate
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "loki-distri-ca",
					Namespace: operatorNamespace,
				},
				Data: map[string]string{"ca.crt": "certificate data"},
			}
			cleanupList = append(cleanupList, cm)
			Expect(k8sClient.Create(ctx, cm)).Should(Succeed())
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Mode = flowslatest.LokiModeMicroservices
				fc.Spec.Loki.Microservices = flowslatest.LokiMicroservicesParams{
					IngesterURL: "http://loki-ingested:3100/",
					QuerierURL:  "http://loki-queries:3100/",
					TLS: flowslatest.ClientTLS{
						Enable: true,
						CACert: flowslatest.CertificateReference{
							Type:     flowslatest.RefTypeConfigMap,
							Name:     "loki-distri-ca",
							CertFile: "ca.crt",
						},
					},
				}
			})
		})

		It("Should have certificate mounted", func() {
			By("Expecting certificate mounted")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &depl); err != nil {
					return err
				}
				return depl.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(2))
			Expect(depl.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
			Expect(depl.Spec.Template.Spec.Volumes[1].Name).To(Equal("loki-certs-ca"))
		})

		It("Should restore no TLS config", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Microservices.TLS = flowslatest.ClientTLS{
					Enable: false,
				}
			})
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &depl); err != nil {
					return err
				}
				return depl.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(1))
			Expect(depl.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
		})
	})

	Context("Using certificates with loki monolithic mode", func() {
		flpKey := deplRef.GetKey(operatorNamespace)
		depl := appsv1.Deployment{}
		It("Should update Loki to use TLS", func() {
			// Create CM certificate
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "loki-mono-ca",
					Namespace: operatorNamespace,
				},
				Data: map[string]string{"ca.crt": "certificate data"},
			}
			cleanupList = append(cleanupList, cm)
			Expect(k8sClient.Create(ctx, cm)).Should(Succeed())
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Mode = flowslatest.LokiModeMonolithic
				fc.Spec.Loki.Monolithic = flowslatest.LokiMonolithParams{
					URL: "http://loki-mono:3100/",
					TLS: flowslatest.ClientTLS{
						Enable: true,
						CACert: flowslatest.CertificateReference{
							Type:     flowslatest.RefTypeConfigMap,
							Name:     "loki-mono-ca",
							CertFile: "ca.crt",
						},
					},
				}
			})
		})

		It("Should have certificate mounted", func() {
			By("Expecting certificate mounted")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &depl); err != nil {
					return err
				}
				return depl.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(2))
			Expect(depl.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
			Expect(depl.Spec.Template.Spec.Volumes[1].Name).To(Equal("loki-certs-ca"))
		})

		It("Should restore no TLS config", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Monolithic.TLS = flowslatest.ClientTLS{
					Enable: false,
				}
			})
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &depl); err != nil {
					return err
				}
				return depl.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(1))
			Expect(depl.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
		})
	})

	Context("Using Certificates With Loki in LokiStack Mode", func() {
		flpKey := deplRef.GetKey(operatorNamespace)
		depl := appsv1.Deployment{}
		It("Should update Loki config successfully", func() {
			// Create CM certificate
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lokistack-gateway-ca-bundle",
					Namespace: operatorNamespace,
				},
				Data: map[string]string{"service-ca.crt": "certificate data"},
			}
			cleanupList = append(cleanupList, cm)
			Expect(k8sClient.Create(ctx, cm)).Should(Succeed())
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Mode = flowslatest.LokiModeLokiStack
				fc.Spec.Loki.LokiStack = flowslatest.LokiStackRef{
					Name:      "lokistack",
					Namespace: operatorNamespace,
				}
			})
		})

		It("Should have certificate mounted", func() {
			By("Expecting certificate mounted")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &depl); err != nil {
					return err
				}
				return depl.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(3))
			Expect(depl.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
			Expect(depl.Spec.Template.Spec.Volumes[1].Name).To(Equal("flowlogs-pipeline"))
			Expect(depl.Spec.Template.Spec.Volumes[2].Name).To(Equal("loki-certs-ca"))
		})

		It("Should deploy Loki roles", func() {
			By("Expecting Writer ClusterRole")
			Eventually(func() interface{} {
				var cr rbacv1.ClusterRole
				return k8sClient.Get(ctx, types.NamespacedName{Name: constants.LokiCRWriter}, &cr)
			}, timeout, interval).Should(Succeed())
			By("Expecting Reader ClusterRole")
			Eventually(func() interface{} {
				var cr rbacv1.ClusterRole
				return k8sClient.Get(ctx, types.NamespacedName{Name: constants.LokiCRReader}, &cr)
			}, timeout, interval).Should(Succeed())
			By("Expecting FLP Writer ClusterRoleBinding")
			Eventually(func() interface{} {
				var crb rbacv1.ClusterRoleBinding
				return k8sClient.Get(ctx, types.NamespacedName{Name: constants.LokiCRBWriter}, &crb)
			}, timeout, interval).Should(Succeed())
		})

		It("Should restore no TLS config in manual mode", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Loki.Mode = flowslatest.LokiModeManual
				fc.Spec.Loki.Manual.TLS = flowslatest.ClientTLS{
					Enable: false,
				}
			})
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &depl); err != nil {
					return err
				}
				return depl.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(1))
			Expect(depl.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
		})
	})

	Context("Changing namespace", func() {
		It("Should update namespace successfully", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Namespace = otherNamespace
			})
		})

		It("Should redeploy FLP in new namespace", func() {
			By("Expecting resources in previous namespace to be deleted")
			expectDeletion(operatorNamespace,
				deplRef,
				cmRef,
				saRef,
			)

			objs := expectCreation(otherNamespace,
				deplRef,
				cmRef,
				saRef,
				crbRef,
			)
			Expect(objs).To(HaveLen(4))
			crb := objs[3].(*rbacv1.ClusterRoleBinding)
			Expect(crb.Subjects).To(HaveLen(1))
			Expect(crb.Subjects[0].Namespace).To(Equal(otherNamespace))
		})
	})

	Context("Checking CR ownership", func() {
		It("Should be garbage collected", func() {
			expectOwnership(otherNamespace,
				deplRef,
				cmRef,
				saRef,
			)
		})
	})

	Context("Back without Kafka", func() {
		It("Should remove kafka config successfully", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.DeploymentModel = flowslatest.DeploymentModelDirect
			})
		})

		It("Should delete kafka transformer", func() {
			expectDeletion(otherNamespace,
				deplRef,
				cmRef,
				saRef,
			)
		})
	})

	Context("Cleanup", func() {
		It("Should delete CR", func() {
			cleanupCR(crKey)
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

func checkDigestUpdate(oldDigest *string, annots map[string]string) error {
	newDigest := annots[constants.PodConfigurationDigest]
	if newDigest == "" {
		return fmt.Errorf("%q annotation can't be empty", constants.PodConfigurationDigest)
	} else if newDigest == *oldDigest {
		return fmt.Errorf("expect digest to change, but is still %s", *oldDigest)
	}
	*oldDigest = newDigest
	return nil
}
