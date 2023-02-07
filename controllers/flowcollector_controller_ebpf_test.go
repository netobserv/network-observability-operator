package controllers

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	. "github.com/netobserv/network-observability-operator/controllers/controllerstest"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

func flowCollectorEBPFSpecs() {
	// Because the simulated Kube server doesn't manage automatic resource cleanup like an actual Kube would do,
	// we need either to cleanup all created resources manually, or to use different namespaces between tests
	// For simplicity, we'll use a different namespace
	operatorNamespace := "namespace-ebpf-specs"
	agentKey := types.NamespacedName{
		Name:      "netobserv-ebpf-agent",
		Namespace: operatorNamespace + "-privileged",
	}
	operatorNamespace2 := "namespace-ebpf-specs2"
	agentKey2 := types.NamespacedName{
		Name:      "netobserv-ebpf-agent",
		Namespace: operatorNamespace2 + "-privileged",
	}
	crKey := types.NamespacedName{Name: "cluster"}
	flpKey2 := types.NamespacedName{
		Name:      constants.FLPName,
		Namespace: operatorNamespace2,
	}
	saKey := types.NamespacedName{
		Name:      constants.EBPFServiceAccount,
		Namespace: agentKey.Namespace,
	}
	saKey2 := types.NamespacedName{
		Name:      constants.EBPFServiceAccount,
		Namespace: agentKey2.Namespace,
	}
	nsKey := types.NamespacedName{Name: agentKey.Namespace}
	nsKey2 := types.NamespacedName{Name: agentKey2.Namespace}

	Context("Netobserv eBPF Agent Reconciler", func() {
		It("Should deploy when it does not exist", func() {
			desired := &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
				Spec: flowslatest.FlowCollectorSpec{
					Namespace:       operatorNamespace,
					DeploymentModel: flowslatest.DeploymentModelDirect,
					Processor: flowslatest.FlowCollectorFLP{
						Port:            9999,
						ImagePullPolicy: "Never",
						LogLevel:        "error",
					},
					Agent: flowslatest.FlowCollectorAgent{
						Type: "EBPF",
						EBPF: flowslatest.FlowCollectorEBPF{
							Sampling:           pointer.Int32Ptr(123),
							CacheActiveTimeout: "15s",
							CacheMaxFlows:      100,
							Interfaces:         []string{"veth0", "/^br-/"},
							ExcludeInterfaces:  []string{"br-3", "lo"},
							LogLevel:           "trace",
							Debug: flowslatest.DebugConfig{
								Env: map[string]string{
									// we'll test that multiple variables are reordered
									"GOGC":           "400",
									"BUFFERS_LENGTH": "100",
								},
							},
						},
					},
				},
			}
			Eventually(func() interface{} {
				return k8sClient.Create(ctx, desired)
			}).WithTimeout(timeout).WithPolling(interval).Should(Succeed())

			ds := appsv1.DaemonSet{}
			By("Expecting to create the netobserv-ebpf-agent DaemonSet")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, agentKey, &ds)
			}).WithTimeout(timeout).WithPolling(interval).Should(Succeed())

			spec := ds.Spec.Template.Spec
			By("expecting that the netobserv-ebpf-agent daemonset is properly configured")
			Expect(spec.HostNetwork).To(BeTrue())
			Expect(spec.DNSPolicy).To(Equal(v1.DNSClusterFirstWithHostNet))
			Expect(spec.ServiceAccountName).To(Equal(constants.EBPFServiceAccount))
			Expect(len(spec.Containers)).To(Equal(1))
			Expect(spec.Containers[0].SecurityContext.Privileged).To(BeNil())
			Expect(spec.Containers[0].SecurityContext.Capabilities.Add).To(ContainElements(
				[]v1.Capability{"BPF", "PERFMON", "NET_ADMIN", "SYS_RESOURCE"},
			))
			Expect(spec.Containers[0].SecurityContext.RunAsUser).To(Not(BeNil()))
			Expect(*spec.Containers[0].SecurityContext.RunAsUser).To(Equal(int64(0)))
			Expect(spec.Containers[0].Env).To(ContainElements(
				v1.EnvVar{Name: "EXPORT", Value: "grpc"},
				v1.EnvVar{Name: "CACHE_ACTIVE_TIMEOUT", Value: "15s"},
				v1.EnvVar{Name: "CACHE_MAX_FLOWS", Value: "100"},
				v1.EnvVar{Name: "LOG_LEVEL", Value: "trace"},
				v1.EnvVar{Name: "INTERFACES", Value: "veth0,/^br-/"},
				v1.EnvVar{Name: "EXCLUDE_INTERFACES", Value: "br-3,lo"},
				v1.EnvVar{Name: "BUFFERS_LENGTH", Value: "100"},
				v1.EnvVar{Name: "GOGC", Value: "400"},
				v1.EnvVar{Name: "SAMPLING", Value: "123"},
				v1.EnvVar{Name: "FLOWS_TARGET_PORT", Value: "9999"},
			))
			hostFound := false
			for _, env := range spec.Containers[0].Env {
				if env.Name == "FLOWS_TARGET_HOST" {
					if env.ValueFrom == nil ||
						env.ValueFrom.FieldRef == nil ||
						env.ValueFrom.FieldRef.FieldPath != "status.hostIP" {
						Fail(fmt.Sprintf("FLOWS_TARGET_HOST expected to refer to \"status.hostIP\"."+
							" Got: %+v", env.ValueFrom))
					} else {
						hostFound = true
						break
					}
				}
			}
			Expect(hostFound).To(BeTrue(),
				fmt.Sprintf("expected FLOWS_TARGET_HOST env var in %+v", spec.Containers[0].Env))

			ns := v1.Namespace{}
			By("expecting to create the netobserv-privileged namespace")
			Expect(k8sClient.Get(ctx, nsKey, &ns)).To(Succeed())
			Expect(ns.Labels).To(Satisfy(func(labels map[string]string) bool {
				return helper.IsSubSet(ns.Labels, map[string]string{
					"app":                                constants.OperatorName,
					"pod-security.kubernetes.io/enforce": "privileged",
					"pod-security.kubernetes.io/audit":   "privileged",
				})
			}))

			By("expecting to create the netobserv-ebpf-agent service account")
			Expect(k8sClient.Get(ctx, saKey, &v1.ServiceAccount{})).To(Succeed())
		})

		It("Should update fields that have changed", func() {
			UpdateCR(crKey, func(fc *flowslatest.FlowCollector) {
				Expect(*fc.Spec.Agent.EBPF.Sampling).To(Equal(int32(123)))
				*fc.Spec.Agent.EBPF.Sampling = 4
				fc.Spec.Agent.EBPF.Privileged = true
			})

			ds := appsv1.DaemonSet{}
			By("expecting that the daemonset spec has eventually changed")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, agentKey, &ds); err != nil {
					return err
				}
				expected := v1.EnvVar{Name: "SAMPLING", Value: "4"}
				for _, env := range ds.Spec.Template.Spec.Containers[0].Env {
					if env == expected {
						return nil
					}
				}
				return fmt.Errorf("unexpected env vars: %#v",
					ds.Spec.Template.Spec.Containers[0].Env)
			}).WithTimeout(timeout).WithPolling(interval).Should(Succeed())

			container := ds.Spec.Template.Spec.Containers[0]
			Expect(container.SecurityContext.Privileged).To(Not(BeNil()))
			Expect(*container.SecurityContext.Privileged).To(BeTrue())
			Expect(container.SecurityContext.Capabilities).To(BeNil())
		})

		It("Should redeploy all when changing namespace", func() {
			UpdateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Namespace = operatorNamespace2
			})

			By("Expecting daemonset in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, agentKey, &appsv1.DaemonSet{})
			}, timeout, interval).Should(MatchError(`daemonsets.apps "netobserv-ebpf-agent" not found`))

			By("Expecting service account in previous namespace to be deleted")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, saKey, &v1.ServiceAccount{})
			}, timeout, interval).Should(MatchError(`serviceaccounts "netobserv-ebpf-agent" not found`))

			By("Expecting daemonset to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, agentKey2, &appsv1.DaemonSet{})
			}, timeout, interval).Should(Succeed())

			By("Expecting service account to be created in new namespace")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, saKey2, &v1.ServiceAccount{})
			}, timeout, interval).Should(Succeed())
		})

		It("Should undeploy everything when deleted", func() {
			// Retrieve CR to get its UID
			flowCR := &flowslatest.FlowCollector{}
			Eventually(func() error {
				return k8sClient.Get(ctx, crKey, flowCR)
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
			})).Should(Succeed())

			By("expecting to delete the flowcollector")
			Eventually(func() error {
				return k8sClient.Get(ctx,
					types.NamespacedName{Name: crKey.Name},
					&flowslatest.FlowCollector{},
				)
			}).WithTimeout(timeout).WithPolling(interval).
				Should(Satisfy(errors.IsNotFound))

			By("expecting to delete the flowlogs-pipeline deployment")
			Eventually(func() error {
				return k8sClient.Get(ctx, flpKey2, &flowslatest.FlowCollector{})
			}).WithTimeout(timeout).WithPolling(interval).
				Should(Satisfy(errors.IsNotFound))

			By("expecting to delete netobserv-ebpf-agent daemonset")
			Eventually(func() interface{} {
				ds := &appsv1.DaemonSet{}
				if err := k8sClient.Get(ctx, agentKey2, ds); err != nil {
					return err
				}
				return ds
			}).WithTimeout(timeout).WithPolling(interval).
				Should(BeGarbageCollectedBy(flowCR))

			By("expecting to delete the netobserv-privileged namespace")
			Eventually(func() interface{} {
				ns := &v1.Namespace{}
				if err := k8sClient.Get(ctx, nsKey2, ns); err != nil {
					return err
				}
				return ns
			}).WithTimeout(timeout).WithPolling(interval).
				Should(BeGarbageCollectedBy(flowCR))

			By("expecting to delete the netobserv-ebpf-agent service account")
			Eventually(func() interface{} {
				sa := &v1.ServiceAccount{}
				if err := k8sClient.Get(ctx, saKey2, sa); err != nil {
					return err
				}
				return sa
			}).WithTimeout(timeout).WithPolling(interval).
				Should(BeGarbageCollectedBy(flowCR))
		})
	})
}

func flowCollectorEBPFKafkaSpecs() {
	operatorNamespace := "ebpf-kafka-specs"
	agentKey := types.NamespacedName{
		Name:      "netobserv-ebpf-agent",
		Namespace: operatorNamespace + "-privileged",
	}
	crKey := types.NamespacedName{Name: "cluster"}
	flpIngesterKey := types.NamespacedName{
		Name:      constants.FLPName + "-ingester",
		Namespace: operatorNamespace,
	}
	flpTransformerKey := types.NamespacedName{
		Name:      constants.FLPName + "-transformer",
		Namespace: operatorNamespace,
	}
	Context("Netobserv eBPF Agent Reconciler", func() {
		It("Should deploy the agent with the proper configuration", func() {
			descriptor := &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
				Spec: flowslatest.FlowCollectorSpec{
					Namespace:       operatorNamespace,
					Agent:           flowslatest.FlowCollectorAgent{Type: "EBPF"},
					DeploymentModel: flowslatest.DeploymentModelKafka,
					Kafka: flowslatest.FlowCollectorKafka{
						Address: "kafka-cluster-kafka-bootstrap",
						Topic:   "network-flows",
					},
				},
			}
			Expect(k8sClient.Create(ctx, descriptor)).Should(Succeed())

			ds := appsv1.DaemonSet{}
			By("making sure that the proper environment variables have been passed to the agent")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, agentKey, &ds)
			}).WithTimeout(timeout).WithPolling(interval).Should(Succeed())

			spec := ds.Spec.Template.Spec
			Expect(len(spec.Containers)).To(Equal(1))
			Expect(spec.Containers[0].Env).To(ContainElements(
				v1.EnvVar{Name: "EXPORT", Value: "kafka"},
				v1.EnvVar{Name: "KAFKA_BROKERS", Value: "kafka-cluster-kafka-bootstrap"},
				v1.EnvVar{Name: "KAFKA_TOPIC", Value: "network-flows"},
			))
		})
		It("Should properly deploy flowlogs-pipeline", func() {
			By("deploying flowlogs-pipeline-transformer")
			Eventually(func() interface{} {
				return k8sClient.Get(ctx, flpTransformerKey, &appsv1.Deployment{})
			}, timeout, interval).Should(Succeed())

			By("not deploying flowlogs-pipeline-ingester")
			Expect(k8sClient.Get(ctx, flpIngesterKey, &appsv1.DaemonSet{})).
				Should(Not(Succeed()))
		})
		It("Should correctly undeploy", func() {
			// Retrieve CR to get its UID
			flowCR := &flowslatest.FlowCollector{}
			Eventually(func() error {
				return k8sClient.Get(ctx, crKey, flowCR)
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
			})).Should(Succeed())

			By("expecting to delete the flowcollector")
			Eventually(func() error {
				return k8sClient.Get(ctx,
					types.NamespacedName{Name: crKey.Name},
					&flowslatest.FlowCollector{},
				)
			}).WithTimeout(timeout).WithPolling(interval).
				Should(Satisfy(errors.IsNotFound))

			By("expecting to delete the flowlogs-pipeline-transformer deployment")
			Eventually(func() interface{} {
				dp := &appsv1.Deployment{}
				if err := k8sClient.Get(ctx, flpTransformerKey, dp); err != nil {
					return err
				}
				return dp
			}).WithTimeout(timeout).WithPolling(interval).
				Should(BeGarbageCollectedBy(flowCR))

			By("expecting to delete netobserv-ebpf-agent daemonset")
			Eventually(func() interface{} {
				ds := &appsv1.DaemonSet{}
				if err := k8sClient.Get(ctx, agentKey, ds); err != nil {
					return err
				}
				return ds
			}).WithTimeout(timeout).WithPolling(interval).
				Should(BeGarbageCollectedBy(flowCR))
		})
	})
}
