package controllers

import (
	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func flowCollectorEBPFSpecs() {
	Describe("Flow Collector with eBPF agent", Ordered, Serial, func() {
		agentKey := types.NamespacedName{
			Name:      "netobserv-agent",
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
							Image: "netobserv-agent:latest",
						},
					},
				}
				Expect(k8sClient.Create(ctx, desired)).Should(Succeed())

				ds := appsv1.DaemonSet{}
				By("Expecting to create the netobserv-agent DaemonSet")
				Eventually(func() interface{} {
					return k8sClient.Get(ctx, agentKey, &ds)
				}).WithTimeout(timeout).WithPolling(interval).Should(Succeed())

				spec := ds.Spec.Template.Spec
				By("expecting that the netobserv-agent daemonset is properly configured")
				Expect(spec.HostNetwork).To(BeTrue())
				Expect(spec.DNSPolicy).To(Equal(v1.DNSClusterFirstWithHostNet))
				Expect(spec.ServiceAccountName).To(Equal(constants.EBPFServiceAccount))
				Expect(len(spec.Containers)).To(Equal(1))
				Expect(spec.Containers[0].SecurityContext.Privileged).To(Not(BeNil()))
				Expect(*spec.Containers[0].SecurityContext.Privileged).To(BeTrue())
				Expect(spec.Containers[0].Env[0].Name).To(Equal("FLOWS_TARGET_HOST"))
				Expect(spec.Containers[0].Env[0].ValueFrom.FieldRef.FieldPath).To(Equal("status.hostIP"))
				Expect(spec.Containers[0].Env[1].Name).To(Equal("FLOWS_TARGET_PORT"))
				Expect(spec.Containers[0].Env[1].Value).To(Equal("9999"))

				ns := v1.Namespace{}
				By("expecting to create the network-observability-privileged namespace")
				Expect(k8sClient.Get(ctx,
					types.NamespacedName{Name: "network-observability-privileged"},
					&ns)).To(Succeed())
				Expect(ns.Labels).To(Satisfy(func(labels map[string]string) bool {
					return helper.IsSubSet(ns.Labels, map[string]string{
						"app":                                "network-observability-operator",
						"pod-security.kubernetes.io/enforce": "privileged",
						"pod-security.kubernetes.io/audit":   "privileged",
					})
				}))

				sa := v1.ServiceAccount{}
				By("expecting to create the netobserv-agent service account")
				Expect(k8sClient.Get(ctx, saKey, &sa)).To(Succeed())
			})

			It("should restore the managed objects if manually changed", func() {
				By("expecting to restore the network-observability-privileged namespace")
			})

			It("should undeploy everything when deleted", func() {
				Expect(k8sClient.Delete(ctx, &flowsv1alpha1.FlowCollector{
					ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
				})).Should(Succeed())

				By("expecting to delete the netobserv-agent")
				Eventually(k8sClient.Get(ctx,
					types.NamespacedName{Name: crKey.Name},
					&flowsv1alpha1.FlowCollector{},
				), timeout, interval).Should(Satisfy(errors.IsNotFound))

				By("expecting to delete the flowlogs-pipeline deployment")
				Eventually(k8sClient.Get(ctx, flpKey, &flowsv1alpha1.FlowCollector{}),
					timeout, interval).Should(Satisfy(errors.IsNotFound))

				By("expecting to delete the network-observability-privileged namespace")
				By("expecting to delete the netobserv-agent service account")

			})
		})

		/*
			 Nuevos tests
				- Crear flowcollector ebpf y ver que agente y security stufff se ha creado
				- Modificar namespaces/agente/serviceaccount/securitycontextconstraints y ver que vuelven a su lugar
			    - Borrar flowcollector y ver que todo se ha borrado
			//*/

	})
}
