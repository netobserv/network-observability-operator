package controllers

import (
	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Flow Collector with eBPF agent", Serial, func() {
	const namespace, agentName = "network-observability-privileged", "netobserv-agent"
	agentKey := types.NamespacedName{Name: agentName, Namespace: namespace}
	crKey := types.NamespacedName{Name: "cluster", Namespace: namespace}

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

			//By("Expecting to create the netobserv-agent DaemonSet")
			//Eventually(func() interface{} {
			//	ds := appsv1.DaemonSet{}
			//	if err := k8sClient.Get(ctx, agentKey, &ds); err != nil {
			//		return err
			//	}
			//	return ds
			//}, timeout, interval).Should(Equal("foo"))
			By("expecting to create the network-observability-privileged namespace")
			By("expecting to create the netobserv-agent service account")
		})

		It("should restore the managed objects if manually changed", func() {
			By("expecting to restore the network-observability-privileged namespace")
		})

		It("should undeploy everything when deleted", func() {
			//Expect(k8sClient.Delete(ctx, &flowsv1alpha1.FlowCollector{
			//	ObjectMeta: metav1.ObjectMeta{Name: crKey.Name},
			//})).Should(Succeed())

			By("expecting to delete the netobserv-agent")
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
