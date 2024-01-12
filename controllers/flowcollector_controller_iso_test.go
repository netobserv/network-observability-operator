package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ascv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/pkg/test"
)

// nolint:cyclop
func flowCollectorIsoSpecs() {
	const operatorNamespace = "main-namespace"
	crKey := types.NamespacedName{
		Name: "cluster",
	}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// This aims to verify that the CRD is preserved / unchanged upon successive serialization/deserialization
	// A typical pitfall is go's empty values messing up with CRD-defined default values, e.g:
	// If field A (bool) has "omitempty" and defaults to true, and is explicitely set to false,
	// then serialization would remove that field (bc go says false ~= empty), and subsequent kube's deserialization would set it to true.
	// The typical workaround is to use pointers on fields (*bool) so 0-value is nil.
	Context("Isomorphic ser/de", func() {
		zero := int32(0)
		defaultTLS := flowslatest.ClientTLS{
			Enable:             false,
			InsecureSkipVerify: false,
			CACert: flowslatest.CertificateReference{
				Type:     "configmap",
				Name:     "",
				CertFile: "",
				CertKey:  "",
			},
			UserCert: flowslatest.CertificateReference{
				Type:     "configmap",
				Name:     "",
				CertFile: "",
				CertKey:  "",
			},
		}

		specInput := flowslatest.FlowCollectorSpec{
			Namespace:       operatorNamespace,
			DeploymentModel: flowslatest.DeploymentModelDirect,
			Processor: flowslatest.FlowCollectorFLP{
				Port:                           12345,
				HealthPort:                     12346,
				ProfilePort:                    0,
				ImagePullPolicy:                "Always",
				LogLevel:                       "trace",
				Resources:                      v1.ResourceRequirements{Limits: nil, Requests: nil},
				KafkaConsumerReplicas:          &zero,
				KafkaConsumerAutoscaler:        flowslatest.FlowCollectorHPA{Status: "Disabled", MinReplicas: &zero, MaxReplicas: zero, Metrics: []ascv2.MetricSpec{}},
				KafkaConsumerQueueCapacity:     int(zero),
				KafkaConsumerBatchSize:         int(zero),
				ConversationHeartbeatInterval:  &metav1.Duration{Duration: time.Second},
				ConversationEndTimeout:         &metav1.Duration{Duration: time.Second},
				ConversationTerminatingTimeout: &metav1.Duration{Duration: time.Second},
				MultiClusterDeployment:         ptr.To(true),
				ClusterName:                    "testCluster",
				Debug:                          flowslatest.DebugConfig{},
				LogTypes:                       &outputRecordTypes,
				Metrics: flowslatest.FLPMetrics{
					Server: flowslatest.MetricsServerConfig{
						Port: 12347,
						TLS: flowslatest.ServerTLS{
							Type:     "Disabled",
							Provided: nil,
						},
					},
					DisableAlerts: []flowslatest.FLPAlert{},
				},
				EnableKubeProbes: ptr.To(false),
				DropUnusedFields: ptr.To(false),
			},
			Agent: flowslatest.FlowCollectorAgent{
				Type: "eBPF",
				IPFIX: flowslatest.FlowCollectorIPFIX{
					Sampling:           2, // 0 is forbidden here
					CacheActiveTimeout: "5s",
					CacheMaxFlows:      100,
					ForceSampleAll:     false,
					ClusterNetworkOperator: flowslatest.ClusterNetworkOperatorConfig{
						Namespace: "test",
					},
					OVNKubernetes: flowslatest.OVNKubernetesConfig{
						Namespace:     "test",
						DaemonSetName: "test",
						ContainerName: "test",
					},
				},
				EBPF: flowslatest.FlowCollectorEBPF{
					Sampling:           &zero,
					CacheActiveTimeout: "5s",
					CacheMaxFlows:      100,
					ImagePullPolicy:    "Always",
					Debug:              flowslatest.DebugConfig{},
					LogLevel:           "trace",
					Resources:          v1.ResourceRequirements{Limits: nil, Requests: nil},
					Interfaces:         []string{},
					ExcludeInterfaces:  []string{},
					Privileged:         false,
					KafkaBatchSize:     0,
					Features:           nil,
				},
			},
			ConsolePlugin: flowslatest.FlowCollectorConsolePlugin{
				Enable:          ptr.To(true),
				Register:        ptr.To(false),
				Replicas:        &zero,
				Port:            12345,
				ImagePullPolicy: "Always",
				Resources:       v1.ResourceRequirements{Limits: nil, Requests: nil},
				LogLevel:        "trace",
				Autoscaler:      flowslatest.FlowCollectorHPA{Status: "Disabled", MinReplicas: &zero, MaxReplicas: zero, Metrics: []ascv2.MetricSpec{}},
				PortNaming: flowslatest.ConsolePluginPortConfig{
					Enable:    ptr.To(false),
					PortNames: map[string]string{},
				},
				QuickFilters: []flowslatest.QuickFilter{},
			},
			Loki: flowslatest.FlowCollectorLoki{
				Enable: ptr.To(true),
				Mode:   flowslatest.LokiModeManual,
				Manual: flowslatest.LokiManualParams{
					IngesterURL: "http://loki",
					QuerierURL:  "http://loki",
					StatusURL:   "",
					TenantID:    "test",
					AuthToken:   "Disabled",
					TLS:         defaultTLS,
					StatusTLS:   defaultTLS,
				},
				Microservices: flowslatest.LokiMicroservicesParams{
					IngesterURL: "http://loki-distributor:3100/",
					QuerierURL:  "http://loki-query-frontend:3100/",
					TenantID:    "netobserv",
					TLS:         defaultTLS,
				},
				Monolithic: flowslatest.LokiMonolithParams{
					URL:      "http://loki:3100/",
					TenantID: "netobserv",
					TLS:      defaultTLS,
				},
				LokiStack: flowslatest.LokiStackRef{
					Name:      "loki",
					Namespace: "",
				},
				BatchWait:    &metav1.Duration{Duration: time.Second},
				BatchSize:    100,
				Timeout:      &metav1.Duration{Duration: time.Second},
				MinBackoff:   &metav1.Duration{Duration: time.Second},
				MaxBackoff:   &metav1.Duration{Duration: time.Second},
				MaxRetries:   &zero,
				StaticLabels: map[string]string{},
			},
			Kafka: flowslatest.FlowCollectorKafka{
				Address: "http://kafka",
				Topic:   "topic",
				TLS:     defaultTLS,
				SASL: flowslatest.SASLConfig{
					Type: "Disabled",
					ClientIDReference: flowslatest.FileReference{
						Type:      "configmap",
						Name:      "",
						Namespace: "",
						File:      "",
					},
					ClientSecretReference: flowslatest.FileReference{
						Type:      "configmap",
						Name:      "",
						Namespace: "",
						File:      "",
					},
				},
			},
			Exporters: []*flowslatest.FlowCollectorExporter{},
		}

		It("Should create CR successfully", func() {
			Eventually(func() interface{} {
				return k8sClient.Create(ctx, &flowslatest.FlowCollector{
					ObjectMeta: metav1.ObjectMeta{
						Name: crKey.Name,
					},
					Spec: specInput,
				})
			}, timeout, interval).Should(Succeed())
		})

		It("Should not have modified input CR values", func() {
			cr := test.GetCR(ctx, k8sClient, crKey)

			// For easier debugging, we check CR parts one by one
			Expect(cr.Spec.Processor).Should(Equal(specInput.Processor))
			Expect(cr.Spec.Agent).Should(Equal(specInput.Agent))
			Expect(cr.Spec.ConsolePlugin).Should(Equal(specInput.ConsolePlugin))
			Expect(cr.Spec.Loki).Should(Equal(specInput.Loki))
			Expect(cr.Spec.Kafka).Should(Equal(specInput.Kafka))
			Expect(cr.Spec.Exporters).Should(Equal(specInput.Exporters))

			// Catch-all in case we missed something
			Expect(cr.Spec).Should(Equal(specInput))
		})
	})

	Context("Cleanup", func() {
		// Retrieve CR to get its UID
		flowCR := flowslatest.FlowCollector{}
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
	})
}
