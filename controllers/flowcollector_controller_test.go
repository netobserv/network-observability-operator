package controllers

import (
	. "github.com/onsi/ginkgo/v2"
	operatorsv1 "github.com/openshift/api/operator/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/pkg/test"
)

const (
	timeout  = test.Timeout
	interval = test.Interval
)

var (
	updateCR = func(key types.NamespacedName, updater func(*flowslatest.FlowCollector)) {
		test.UpdateCR(ctx, k8sClient, key, updater)
	}
	cleanupCR = func(key types.NamespacedName) {
		test.CleanupCR(ctx, k8sClient, key)
	}
	installResources = func(objs ...client.Object) []client.Object {
		GinkgoHelper()
		return test.InstallResources(ctx, k8sClient, objs)
	}
	cleanupResources = func(objs []client.Object) {
		GinkgoHelper()
		test.CleanupResources(ctx, k8sClient, objs)
	}
	expectPresence = func(namespace string, objs ...test.ResourceRef) []client.Object {
		GinkgoHelper()
		return test.ExpectPresence(ctx, k8sClient, namespace, objs...)
	}
	expectAbsence = func(namespace string, objs ...test.ResourceRef) {
		GinkgoHelper()
		test.ExpectAbsence(ctx, k8sClient, namespace, objs...)
	}
	expectOwnership = func(namespace string, objs ...test.ResourceRef) {
		GinkgoHelper()
		test.ExpectOwnership(ctx, k8sClient, namespace, objs...)
	}
)

type testCase struct {
	name   string
	using  *flowslatest.FlowCollectorSpec
	expect []test.ResourceRef
}

// nolint:cyclop
func checkInstalledResources() {
	cases := []testCase{
		{
			name:  "Minimal CR",
			using: &flowslatest.FlowCollectorSpec{},
			expect: []test.ResourceRef{
				test.AgentDS,
				test.AgentSA,
				test.AgentFLPMetricsSvc,
				test.AgentFLPSM,
				test.AgentFLPRule,
				test.AgentFLPCRB,
				test.AgentNS,
				test.PluginDepl,
				test.PluginCM,
				test.PluginSvc,
				test.PluginSA,
				test.PluginCRB,
				test.PluginSM,
			},
		},
		{
			name: "With agent metrics",
			using: &flowslatest.FlowCollectorSpec{
				Agent: flowslatest.FlowCollectorAgent{
					EBPF: flowslatest.FlowCollectorEBPF{
						Metrics: flowslatest.EBPFMetrics{
							Enable: ptr.To(true),
						},
					},
				},
			},
			expect: []test.ResourceRef{
				test.AgentDS,
				test.AgentSA,
				test.AgentFLPMetricsSvc,
				test.AgentFLPSM,
				test.AgentFLPRule,
				test.AgentFLPCRB,
				test.AgentMetricsSvc,
				test.AgentSM,
				test.AgentNS,
				test.PluginDepl,
				test.PluginCM,
				test.PluginSvc,
				test.PluginSA,
				test.PluginCRB,
				test.PluginSM,
			},
		},
		{
			name: "With Kafka",
			using: &flowslatest.FlowCollectorSpec{
				DeploymentModel: flowslatest.DeploymentModelKafka,
			},
			expect: []test.ResourceRef{
				test.AgentDS,
				test.AgentSA,
				test.AgentNS,
				test.FLPDepl,
				test.FLPCM,
				test.FLPSA,
				test.FLPMetricsSvc,
				test.FLPSM,
				test.FLPRule,
				test.FLPCRB,
				test.PluginDepl,
				test.PluginCM,
				test.PluginSvc,
				test.PluginSA,
				test.PluginCRB,
				test.PluginSM,
			},
		},
		{
			name: "Without Console plugin",
			using: &flowslatest.FlowCollectorSpec{
				DeploymentModel: flowslatest.DeploymentModelKafka,
				ConsolePlugin: flowslatest.FlowCollectorConsolePlugin{
					Enable: ptr.To(false),
				},
			},
			expect: []test.ResourceRef{
				test.AgentDS,
				test.AgentSA,
				test.AgentNS,
				test.FLPDepl,
				test.FLPCM,
				test.FLPSA,
				test.FLPMetricsSvc,
				test.FLPSM,
				test.FLPRule,
				test.FLPCRB,
			},
		},
		{
			name: "With LokiStack",
			using: &flowslatest.FlowCollectorSpec{
				Loki: flowslatest.FlowCollectorLoki{
					Mode: flowslatest.LokiModeLokiStack,
					LokiStack: flowslatest.LokiStackRef{
						Name:      "loki",
						Namespace: "default",
					},
				},
			},
			expect: []test.ResourceRef{
				test.AgentDS,
				test.AgentSA,
				test.AgentFLPMetricsSvc,
				test.AgentFLPSM,
				test.AgentFLPRule,
				test.AgentFLPCRB,
				test.AgentNS,
				test.PluginDepl,
				test.PluginCM,
				test.PluginSvc,
				test.PluginSA,
				test.PluginCRB,
				test.PluginSM,
				test.LokiReaderCR,
				test.LokiWriterCR,
				test.LokiWriterCRB,
			},
		},
	}

	var installed []client.Object
	It("Should install initial resources", func() {
		installed = installResources(
			&operatorsv1.Console{
				ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
				Spec: operatorsv1.ConsoleSpec{
					OperatorSpec: operatorsv1.OperatorSpec{
						ManagementState: operatorsv1.Unmanaged,
					},
				},
			},
			&flowslatest.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
				Spec:       flowslatest.FlowCollectorSpec{Namespace: test.TestNamespace},
			},
			&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "loki-gateway-ca-bundle",
					Namespace: "default",
				},
				Data: map[string]string{"service-ca.crt": "certificate data"},
			},
		)
	})

	Context("Iterating over test cases", func() {
		for _, c := range cases {
			checkCase(c)
		}
	})

	Context("Iterating in reverse order", func() {
		for i := len(cases) - 1; i >= 0; i-- {
			checkCase(cases[i])
		}
	})

	Context("Cleanup", func() {
		It("Should delete initial resources", func() {
			cleanupResources(installed)
		})

		It("Should cleanup other data", func() {
			for _, obj := range test.ClusterResources {
				_ = k8sClient.Delete(ctx, obj.Resource)
			}
		})
	})
}

func checkCase(c testCase) {
	It("Running case: "+c.name, func() {
		updateCR(types.NamespacedName{Name: "cluster"}, func(fc *flowslatest.FlowCollector) {
			fc.Spec = *c.using
			fc.Spec.Namespace = test.TestNamespace
		})
		clusterResources := test.GetClusterResourcesIn(c.expect)
		resourcesMainNamespace := append(
			test.GetFLPResourcesIn(c.expect),
			test.GetPluginResourcesIn(c.expect)...,
		)
		resourcesPrivilegedNamespace := test.GetAgentResourcesIn(c.expect)

		// Ensure presence
		expectPresence("", clusterResources...)
		expectPresence(test.TestNamespace, resourcesMainNamespace...)
		expectPresence(test.TestNamespace+"-privileged", resourcesPrivilegedNamespace...)
		expectOwnership(test.TestNamespace, resourcesMainNamespace...)
		expectOwnership(test.TestNamespace+"-privileged", resourcesPrivilegedNamespace...)
		// Ensure absence
		unusedResourcesMainNamespace := append(
			test.GetFLPResourcesNotIn(c.expect),
			test.GetPluginResourcesNotIn(c.expect)...,
		)
		unusedResourcesPrivilegedNamespace := test.GetAgentResourcesNotIn(c.expect)
		expectAbsence(test.TestNamespace, unusedResourcesMainNamespace...)
		expectAbsence(test.TestNamespace+"-privileged", unusedResourcesPrivilegedNamespace...)
	})
}
