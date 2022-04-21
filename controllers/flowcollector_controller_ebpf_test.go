package controllers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
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
	crKey := types.NamespacedName{Name: "cluster"}
	flpKey := types.NamespacedName{
		Name:      constants.FLPName,
		Namespace: operatorNamespace,
	}
	saKey := types.NamespacedName{
		Name:      constants.EBPFServiceAccount,
		Namespace: agentKey.Namespace,
	}
	nsKey := types.NamespacedName{Name: agentKey.Namespace}

	Context("Netobserv eBPF Agent Reconciler", func() {
		It("Should deploy when it does not exist", func() {
			desired := &flowsv1alpha1.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
				Spec: flowsv1alpha1.FlowCollectorSpec{
					Namespace: operatorNamespace,
					FlowlogsPipeline: flowsv1alpha1.FlowCollectorFLP{
						Kind:            "DaemonSet",
						Port:            9999,
						ImagePullPolicy: "Never",
						LogLevel:        "error",
						Image:           "testimg:latest",
					},
					EBPF: &flowsv1alpha1.FlowCollectorEBPF{
						Image:              "netobserv-ebpf-agent:latest",
						Sampling:           123,
						CacheActiveTimeout: "15s",
						CacheMaxFlows:      100,
						Interfaces:         []string{"veth0", "/^br-/"},
						ExcludeInterfaces:  []string{"br-3", "lo"},
						BuffersLength:      100,
						LogLevel:           "trace",
					},
				},
			}
			Expect(k8sClient.Create(ctx, desired)).Should(Succeed())

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
			Expect(spec.Containers[0].SecurityContext.Privileged).To(Not(BeNil()))
			Expect(*spec.Containers[0].SecurityContext.Privileged).To(BeTrue())
			env := spec.Containers[0].Env
			Expect(len(env)).To(Equal(9))
			Expect(env[0]).To(Equal(
				v1.EnvVar{Name: "CACHE_ACTIVE_TIMEOUT", Value: "15s"},
			))
			Expect(env[1]).To(Equal(
				v1.EnvVar{Name: "CACHE_MAX_FLOWS", Value: "100"},
			))
			Expect(env[2]).To(Equal(
				v1.EnvVar{Name: "LOG_LEVEL", Value: "trace"},
			))
			Expect(env[3]).To(Equal(
				v1.EnvVar{Name: "INTERFACES", Value: "veth0,/^br-/"},
			))
			Expect(env[4]).To(Equal(
				v1.EnvVar{Name: "EXCLUDE_INTERFACES", Value: "br-3,lo"},
			))
			Expect(env[5]).To(Equal(
				v1.EnvVar{Name: "BUFFERS_LENGTH", Value: "100"},
			))
			Expect(env[6]).To(Equal(
				v1.EnvVar{Name: "SAMPLING", Value: "123"},
			))
			Expect(env[7].Name).To(Equal("FLOWS_TARGET_HOST"))
			Expect(env[7].ValueFrom.FieldRef.FieldPath).To(Equal("status.hostIP"))
			Expect(env[8]).To(Equal(
				v1.EnvVar{Name: "FLOWS_TARGET_PORT", Value: "9999"},
			))

			ns := v1.Namespace{}
			By("expecting to create the network-observability-privileged namespace")
			Expect(k8sClient.Get(ctx, nsKey, &ns)).To(Succeed())
			Expect(ns.Labels).To(Satisfy(func(labels map[string]string) bool {
				return helper.IsSubSet(ns.Labels, map[string]string{
					"app":                                "network-observability-operator",
					"pod-security.kubernetes.io/enforce": "privileged",
					"pod-security.kubernetes.io/audit":   "privileged",
				})
			}))

			By("expecting to create the netobserv-ebpf-agent service account")
			Expect(k8sClient.Get(ctx, saKey, &v1.ServiceAccount{})).To(Succeed())
		})

		It("should undeploy everything when deleted", func() {
			// Retrieve CR to get its UID
			flowCR := &flowsv1alpha1.FlowCollector{}
			Eventually(func() error {
				return k8sClient.Get(ctx, crKey, flowCR)
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &flowsv1alpha1.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
			})).Should(Succeed())

			By("expecting to delete the netobserv-ebpf-agent")
			Eventually(func() error {
				return k8sClient.Get(ctx,
					types.NamespacedName{Name: crKey.Name},
					&flowsv1alpha1.FlowCollector{},
				)
			}).WithTimeout(timeout).WithPolling(interval).
				Should(Satisfy(errors.IsNotFound))

			By("expecting to delete the flowlogs-pipeline deployment")
			Eventually(func() error {
				return k8sClient.Get(ctx, flpKey, &flowsv1alpha1.FlowCollector{})
			}).WithTimeout(timeout).WithPolling(interval).
				Should(Satisfy(errors.IsNotFound))

			By("expecting to delete netobserv-ebpf-agent daemonset")
			Eventually(func() interface{} {
				ds := &appsv1.DaemonSet{}
				if err := k8sClient.Get(ctx, agentKey, ds); err != nil {
					return err
				}
				return ds
			}).WithTimeout(timeout).WithPolling(interval).
				Should(BeGarbageCollectedBy(flowCR))

			By("expecting to delete the network-observability-privileged namespace")
			Eventually(func() interface{} {
				ns := &v1.Namespace{}
				if err := k8sClient.Get(ctx, nsKey, ns); err != nil {
					return err
				}
				return ns
			}).WithTimeout(timeout).WithPolling(interval).
				Should(BeGarbageCollectedBy(flowCR))

			By("expecting to delete the netobserv-ebpf-agent service account")
			Eventually(func() interface{} {
				sa := &v1.ServiceAccount{}
				if err := k8sClient.Get(ctx, saKey, sa); err != nil {
					return err
				}
				return sa
			}).WithTimeout(timeout).WithPolling(interval).
				Should(BeGarbageCollectedBy(flowCR))
		})
	})
}
