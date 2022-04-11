package controllers

import (
	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func flowCollectorEBPFSpecs() {
	Describe("Flow Collector with eBPF agent", Ordered, Serial, func() {
		const fcNamespace = "network-observability"
		const agentNamespace, agentName = "network-observability-privileged", "netobserv-agent"
		agentKey := types.NamespacedName{Name: agentName, Namespace: agentNamespace}
		crKey := types.NamespacedName{Name: "cluster", Namespace: fcNamespace}
		flpKey := types.NamespacedName{Name: constants.FLPName, Namespace: operatorNamespace}
		Context("Netobserv eBPF Agent Reconciler", func() {
			It("Should deploy when it does not exist", func() {
				desired := &flowsv1alpha1.FlowCollector{
					ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
					Spec: flowsv1alpha1.FlowCollectorSpec{
						FlowlogsPipeline: flowsv1alpha1.FlowCollectorFLP{
							Kind:            "Deployment",
							Port:            9999,
							ImagePullPolicy: "Never",
							LogLevel:        "error",
							Image:           "testimg:latest",
						},
						EBPF: &flowsv1alpha1.FlowCollectorEBPF{},
					},
				}
				_, _ = desired, agentKey
				//Create
				Expect(k8sClient.Create(ctx, desired)).Should(Succeed())

				ds := appsv1.DaemonSet{}
				By("Expecting to create the netobserv-agent DaemonSet")
				Eventually(k8sClient.Get(ctx, agentKey, &ds), timeout, interval).
					Should(Succeed())

				By("expecting that the netobserv-agent daemonset is properly configured")

				By("expecting to create the network-observability-privileged namespace")
				By("expecting to create the netobserv-agent service account")
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
