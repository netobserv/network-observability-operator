package controllers

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/test"
)

func flowCollectorEBPFSpecs() {
	// Because the simulated Kube server doesn't manage automatic resource cleanup like an actual Kube would do,
	// we need either to cleanup all created resources manually, or to use different namespaces between tests
	// For simplicity, we'll use a different namespace
	operatorNamespace := "namespace-ebpf-specs"
	operatorPrivilegedNamespace := operatorNamespace + "-privileged"
	agentKey := types.NamespacedName{
		Name:      "netobserv-ebpf-agent",
		Namespace: operatorPrivilegedNamespace,
	}
	operatorNamespace2 := "namespace-ebpf-specs2"
	operatorPrivilegedNamespace2 := operatorNamespace2 + "-privileged"

	dsRef := test.DaemonSet(constants.EBPFAgentName)
	saRef := test.ServiceAccount(constants.EBPFServiceAccount)
	svcMetricsRef := test.Service(constants.EBPFAgentMetricsSvcName)
	svcFLPMetricsRef := test.Service("netobserv-ebpf-agent-prom")
	smRef := test.ServiceMonitor(constants.EBPFAgentMetricsSvcMonitoringName)
	smFLPRef := test.ServiceMonitor(constants.EBPFAgentName + "-monitor")
	ruleFLPRef := test.PrometheusRule(constants.EBPFAgentName + "-alert")
	nsRef := test.Namespace(operatorPrivilegedNamespace)

	crKey := types.NamespacedName{Name: "cluster"}

	Context("Netobserv eBPF Agent Reconciler", func() {
		It("Should deploy when it does not exist", func() {
			desired := &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
				Spec: flowslatest.FlowCollectorSpec{
					Namespace:       operatorNamespace,
					DeploymentModel: flowslatest.DeploymentModelDirect,
					Processor: flowslatest.FlowCollectorFLP{
						ImagePullPolicy: "Never",
						LogLevel:        "error",
					},
					Agent: flowslatest.FlowCollectorAgent{
						Type: "eBPF",
						EBPF: flowslatest.FlowCollectorEBPF{
							Sampling:           ptr.To(int32(123)),
							CacheActiveTimeout: "15s",
							CacheMaxFlows:      100,
							Interfaces:         []string{"veth0", "/^br-/"},
							ExcludeInterfaces:  []string{"br-3", "lo"},
							LogLevel:           "trace",
							Advanced: &flowslatest.AdvancedAgentConfig{
								Env: map[string]string{"GOGC": "400", "BUFFERS_LENGTH": "100"},
							},
							Metrics: flowslatest.EBPFMetrics{
								Enable: ptr.To(true),
							},
						},
					},
				},
			}
			Eventually(func() interface{} {
				return k8sClient.Create(ctx, desired)
			}).WithTimeout(timeout).WithPolling(interval).Should(Succeed())

			objs := expectCreation(operatorPrivilegedNamespace,
				dsRef,
				saRef,
				svcMetricsRef,
				svcFLPMetricsRef,
				smRef,
				smFLPRef,
				ruleFLPRef,
				nsRef,
			)
			Expect(objs).To(HaveLen(8))

			spec := objs[0].(*appsv1.DaemonSet).Spec.Template.Spec
			ns := objs[7].(*v1.Namespace)

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
				v1.EnvVar{Name: "EXPORT", Value: "direct-flp"},
				v1.EnvVar{Name: "CACHE_ACTIVE_TIMEOUT", Value: "15s"},
				v1.EnvVar{Name: "CACHE_MAX_FLOWS", Value: "100"},
				v1.EnvVar{Name: "LOG_LEVEL", Value: "trace"},
				v1.EnvVar{Name: "INTERFACES", Value: "veth0,/^br-/"},
				v1.EnvVar{Name: "EXCLUDE_INTERFACES", Value: "br-3,lo"},
				v1.EnvVar{Name: "BUFFERS_LENGTH", Value: "100"},
				v1.EnvVar{Name: "GOGC", Value: "400"},
				v1.EnvVar{Name: "SAMPLING", Value: "123"},
			))
			var flpConfig string
			for _, env := range spec.Containers[0].Env {
				if env.Name == "FLP_CONFIG" {
					flpConfig = env.Value
				}
			}
			Expect(flpConfig).NotTo(BeEmpty())

			// Parse config
			var cfs config.ConfigFileStruct
			err := yaml.Unmarshal([]byte(flpConfig), &cfs)
			Expect(err).To(BeNil())
			Expect(cfs.Pipeline).To(Equal([]config.Stage{
				{Name: "enrich", Follows: "preset-ingester"},
				{Name: "loki", Follows: "enrich"},
				{Name: "prometheus", Follows: "enrich"},
			}))

			Expect(ns.Labels).To(Satisfy(func(labels map[string]string) bool {
				return helper.IsSubSet(ns.Labels, map[string]string{
					"app":                                constants.OperatorName,
					"pod-security.kubernetes.io/enforce": "privileged",
					"pod-security.kubernetes.io/audit":   "privileged",
				})
			}))
		})

		It("Should update fields that have changed", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				Expect(*fc.Spec.Agent.EBPF.Sampling).To(Equal(int32(123)))
				*fc.Spec.Agent.EBPF.Sampling = 4
				fc.Spec.Agent.EBPF.Privileged = true
				fc.Spec.Agent.EBPF.Metrics.Enable = ptr.To(false)
				fc.Spec.DeploymentModel = flowslatest.DeploymentModelKafka
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

			expectDeletion(operatorNamespace+"-privileged",
				svcMetricsRef,
				smRef,
				svcFLPMetricsRef,
				smFLPRef,
				ruleFLPRef,
			)
		})

		It("Should redeploy all when changing namespace", func() {
			updateCR(crKey, func(fc *flowslatest.FlowCollector) {
				fc.Spec.Namespace = operatorNamespace2
			})

			expectDeletion(operatorPrivilegedNamespace,
				dsRef,
				saRef,
			)
			expectCreation(operatorPrivilegedNamespace2,
				dsRef,
				saRef,
			)
		})

		It("Should be garbage collected", func() {
			expectOwnership(operatorPrivilegedNamespace2,
				dsRef,
				test.Namespace(operatorPrivilegedNamespace2),
				saRef,
			)
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
		})
	})
}

func flowCollectorEBPFKafkaSpecs() {
	operatorNamespace := "ebpf-kafka-specs"
	operatorPrivilegedNamespace := operatorNamespace + "-privileged"
	dsRef := test.DaemonSet(constants.EBPFAgentName)
	saRef := test.ServiceAccount(constants.EBPFServiceAccount)
	flpRef := test.Deployment(constants.FLPName)
	flpSvcRef := test.Service(constants.FLPName + "-prom")
	flpSMRef := test.ServiceMonitor(constants.FLPName + "-monitor")
	crKey := types.NamespacedName{Name: "cluster"}

	Context("Netobserv eBPF Agent Reconciler", func() {
		It("Should deploy the agent with the proper configuration", func() {
			descriptor := &flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
				Spec: flowslatest.FlowCollectorSpec{
					Namespace:       operatorNamespace,
					Agent:           flowslatest.FlowCollectorAgent{Type: "eBPF"},
					DeploymentModel: flowslatest.DeploymentModelKafka,
					Kafka: flowslatest.FlowCollectorKafka{
						Address: "kafka-cluster-kafka-bootstrap",
						Topic:   "network-flows",
					},
				},
			}
			Expect(k8sClient.Create(ctx, descriptor)).Should(Succeed())

			objs := expectCreation(operatorPrivilegedNamespace,
				dsRef,
				saRef,
			)
			Expect(objs).To(HaveLen(2))

			spec := objs[0].(*appsv1.DaemonSet).Spec.Template.Spec

			Expect(len(spec.Containers)).To(Equal(1))
			Expect(spec.Containers[0].Env).To(ContainElements(
				v1.EnvVar{Name: "EXPORT", Value: "kafka"},
				v1.EnvVar{Name: "KAFKA_BROKERS", Value: "kafka-cluster-kafka-bootstrap"},
				v1.EnvVar{Name: "KAFKA_TOPIC", Value: "network-flows"},
			))
		})

		It("Should properly deploy flowlogs-pipeline", func() {
			objs := expectCreation(operatorNamespace,
				flpRef,
				flpSvcRef,
				flpSMRef,
			)
			Expect(objs).To(HaveLen(3))
		})

		It("Should be garbage collected", func() {
			expectOwnership(operatorNamespace,
				dsRef,
				flpRef,
				saRef,
			)
		})

		It("Should correctly undeploy", func() {
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
		})
	})
}
