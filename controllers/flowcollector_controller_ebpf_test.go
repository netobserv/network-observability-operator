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

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	. "github.com/netobserv/network-observability-operator/controllers/controllerstest"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

func flowCollectorEBPFSpecs() {
	agentKey := types.NamespacedName{
		Name:      "netobserv-ebpf-agent",
		Namespace: "network-observability-privileged",
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
			Expect(spec.Containers[0].Env).To(ContainElements(
				v1.EnvVar{Name: "CACHE_ACTIVE_TIMEOUT", Value: "15s"},
				v1.EnvVar{Name: "CACHE_MAX_FLOWS", Value: "100"},
				v1.EnvVar{Name: "LOG_LEVEL", Value: "trace"},
				v1.EnvVar{Name: "INTERFACES", Value: "veth0,/^br-/"},
				v1.EnvVar{Name: "EXCLUDE_INTERFACES", Value: "br-3,lo"},
				v1.EnvVar{Name: "BUFFERS_LENGTH", Value: "100"},
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

		It("Should update fields that have changed", func() {
			updated := flowsv1alpha1.FlowCollector{}
			Expect(k8sClient.Get(ctx, crKey, &updated)).Should(Succeed())
			Expect(updated.Spec.EBPF.Sampling).To(Equal(int32(123)))
			updated.Spec.EBPF.Sampling = 4
			Expect(k8sClient.Update(ctx, &updated)).Should(Succeed())

			By("expecting that the daemonset spec has eventually changed")
			Eventually(func() interface{} {
				ds := appsv1.DaemonSet{}
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
		})

		It("Should undeploy everything when deleted", func() {
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
