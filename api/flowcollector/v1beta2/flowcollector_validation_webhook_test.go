package v1beta2

import (
	"context"
	"testing"

	"github.com/netobserv/network-observability-operator/internal/pkg/cluster"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestValidateAgent(t *testing.T) {
	tests := []struct {
		name             string
		fc               *FlowCollector
		ocpVersion       string
		expectedError    string
		expectedWarnings admission.Warnings
	}{
		{
			name: "Empty config is valid",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{},
			},
		},
		{
			name: "Valid configuration",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features:   []AgentFeature{DNSTracking, FlowRTT},
							Privileged: false,
							Sampling:   ptr.To(int32(100)),
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										Action:    "Accept",
										CIDR:      "0.0.0.0/0",
										Direction: "Egress",
										Protocol:  "TCP",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Invalid filter with duplicate CIDR",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features: []AgentFeature{DNSTracking, FlowRTT},
							Sampling: ptr.To(int32(100)),
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										Action:    "Accept",
										CIDR:      "0.0.0.0/0",
										PeerCIDR:  "1.1.1.1/24",
										Direction: "Egress",
										Protocol:  "TCP",
									},
									{
										Action:    "Accept",
										CIDR:      "0.0.0.0/0",
										PeerCIDR:  "1.1.1.1/24",
										Direction: "Egress",
										Protocol:  "UDP",
									},
								},
							},
						},
					},
				},
			},
			expectedError: "flow filter rule CIDR and PeerCIDR 0.0.0.0/0-1.1.1.1/24 already exists",
		},
		{
			name: "PacketDrop Can't detect environment",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features:   []AgentFeature{DNSTracking, FlowRTT, PacketDrop},
							Privileged: true,
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{"Unknown environment, cannot detect if the feature PacketDrop is supported"},
		},
		{
			name:       "PacketDrop on ocp 4.12 triggers warning",
			ocpVersion: "4.12.5",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features:   []AgentFeature{PacketDrop},
							Privileged: true,
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{"The PacketDrop feature requires OpenShift 4.14.0 or above (version detected: 4.12.5)"},
		},
		{
			name:       "PacketDrop is valid",
			ocpVersion: "4.16.0",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features:   []AgentFeature{PacketDrop},
							Privileged: true,
						},
					},
				},
			},
		},
		{
			name:       "PacketDrop without privilege triggers warning",
			ocpVersion: "4.16.0",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features: []AgentFeature{PacketDrop},
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{"The PacketDrop feature requires eBPF Agent to run in privileged mode, which is currently disabled in spec.agent.ebpf.privileged, or to use with eBPF Manager"},
		},
		{
			name:       "NetworkEvents on ocp 4.16 triggers warning",
			ocpVersion: "4.16.5",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features:   []AgentFeature{NetworkEvents},
							Privileged: true,
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{"The NetworkEvents feature requires OpenShift 4.19.0 or above (version detected: 4.16.5)"},
		},
		{
			name:       "NetworkEvents on ocp 4.19.0-0 doesn't trigger warnings",
			ocpVersion: "4.19.0-0.nightly-2025-03-20-063534",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features:   []AgentFeature{NetworkEvents},
							Privileged: true,
						},
					},
				},
			},
		},
		{
			name:       "NetworkEvents without privilege triggers warning",
			ocpVersion: "4.19.0",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features: []AgentFeature{NetworkEvents},
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{"The NetworkEvents feature requires eBPF Agent to run in privileged mode, which is currently disabled in spec.agent.ebpf.privileged"},
		},
		{
			name:       "UDNMapping on ocp 4.18.0-0 doesn't trigger warnings",
			ocpVersion: "4.18.0-0.nightly-2025-06-30-082606",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							Features:   []AgentFeature{UDNMapping},
							Privileged: true,
						},
					},
				},
			},
		},
		{
			name: "FlowFilter different ports configs are mutually exclusive",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										Action:      "Accept",
										CIDR:        "0.0.0.0/0",
										Ports:       intstr.FromInt(80),
										SourcePorts: intstr.FromInt(443),
									},
								},
							},
						},
					},
				},
			},
			expectedError: "cannot configure agent filter with ports and sourcePorts, they are mutually exclusive",
		},
		{
			name: "FlowFilter expect invalid ports",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										Ports: intstr.FromString("abcd"),
									},
								},
							},
						},
					},
				},
			},
			expectedError: "invalid port number",
		},
		{
			name: "FlowFilter expect valid ports range",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										Ports: intstr.FromString("80-255"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "FlowFilter expect invalid ports range (order)",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										Ports: intstr.FromString("255-80"),
									},
								},
							},
						},
					},
				},
			},
			expectedError: "start is greater or equal",
		},
		{
			name: "FlowFilter expect invalid ports range",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										Ports: intstr.FromString("80-?"),
									},
								},
							},
						},
					},
				},
			},
			expectedError: "invalid port number",
		},
		{
			name: "FlowFilter expect valid ports couple",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										Ports: intstr.FromString("255,80"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "FlowFilter expect invalid ports couple",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										Ports: intstr.FromString("80,100,250"),
									},
								},
							},
						},
					},
				},
			},
			expectedError: "expected two integers",
		},
		{
			name: "FlowFilter expect invalid CIDR",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{
						Type: AgentEBPF,
						EBPF: FlowCollectorEBPF{
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								Rules: []EBPFFlowFilterRule{
									{
										CIDR: "1.1.1.1",
									},
								},
							},
						},
					},
				},
			},
			expectedError: "invalid CIDR",
		},
	}

	CurrentClusterInfo = &cluster.Info{}
	for _, test := range tests {
		CurrentClusterInfo.Mock(test.ocpVersion, "")
		v := validator{fc: &test.fc.Spec}
		v.validateAgent()
		if test.expectedError == "" {
			assert.Empty(t, v.errors, test.name)
		} else {
			assert.Len(t, v.errors, 1, test.name)
			assert.ErrorContains(t, v.errors[0], test.expectedError, test.name)
		}
		assert.Equal(t, test.expectedWarnings, v.warnings, test.name)
	}
}

func TestValidateConntrack(t *testing.T) {
	tests := []struct {
		name             string
		fc               *FlowCollector
		expectedError    string
		expectedWarnings admission.Warnings
	}{
		{
			name: "Conntrack with Loki is valid",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Processor: FlowCollectorFLP{
						LogTypes: ptr.To(LogTypeConversations),
					},
					Loki: FlowCollectorLoki{
						Enable: ptr.To(true),
					},
				},
			},
		},
		{
			name: "Conntrack ALL is not recommended",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Processor: FlowCollectorFLP{
						LogTypes: ptr.To(LogTypeAll),
					},
					Loki: FlowCollectorLoki{
						Enable: ptr.To(true),
					},
				},
			},
			expectedWarnings: admission.Warnings{"Enabling all log types (in spec.processor.logTypes) has a high impact on resources footprint"},
		},
		{
			name: "Conntrack without Loki is not recommended",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Processor: FlowCollectorFLP{
						LogTypes: ptr.To(LogTypeConversations),
					},
					Loki: FlowCollectorLoki{
						Enable: ptr.To(false),
					},
				},
			},
			expectedError: "enabling conversation tracking without Loki is not allowed, as it generates extra processing for no benefit",
		},
		{
			name: "Conntrack not allowed with deploymentModel Service",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					DeploymentModel: DeploymentModelService,
					Processor: FlowCollectorFLP{
						LogTypes: ptr.To(LogTypeConversations),
					},
				},
			},
			expectedError: "cannot enable conversation tracking when spec.deploymentModel is Service: you must disable it, or change the deployment model",
		},
	}

	r := FlowCollector{}
	CurrentClusterInfo = &cluster.Info{}
	CurrentClusterInfo.Mock("", "")
	for _, test := range tests {
		warnings, err := r.Validate(context.TODO(), test.fc)
		if test.expectedError == "" {
			assert.NoError(t, err, test.name)
		} else {
			assert.ErrorContains(t, err, test.expectedError, test.name)
		}
		assert.Equal(t, test.expectedWarnings, warnings, test.name)
	}
}

func TestValidateFLP(t *testing.T) {
	tests := []struct {
		name             string
		fc               *FlowCollector
		expectedError    string
		expectedWarnings admission.Warnings
		ocpVersion       string
	}{
		{
			name: "Valid FLP query",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Processor: FlowCollectorFLP{
						Filters: []FLPFilterSet{
							{
								Query: `SrcK8S_Namespace="foo" and without(DstK8S_Namespace)`,
							},
						},
					},
				},
			},
		},
		{
			name: "Invalid FLP query",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Processor: FlowCollectorFLP{
						Filters: []FLPFilterSet{
							{
								Query: `SrcK8S_Namespace="foo" and without(DstK8S_Namespace)`,
							},
							{
								Query: `invalid query`,
							},
						},
					},
				},
			},
			expectedError: "cannot parse spec.processor.filters[1].query: syntax error",
		},
		{
			name: "Missing feature for alerts",
			fc: &FlowCollector{
				Spec: FlowCollectorSpec{
					Processor: FlowCollectorFLP{
						Metrics: FLPMetrics{
							HealthRules: &[]FLPHealthRule{
								{
									Template: HealthRulePacketDropsByKernel,
									Variants: []HealthRuleVariant{},
								},
							},
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{"HealthRule PacketDropsByKernel requires the PacketDrop agent feature to be enabled"},
		},
		{
			name:       "No missing metrics for alerts by default",
			ocpVersion: "4.18.0",
			fc: &FlowCollector{
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{EBPF: FlowCollectorEBPF{
						Features:   []AgentFeature{PacketDrop},
						Privileged: true,
					}},
					Processor: FlowCollectorFLP{
						Metrics: FLPMetrics{
							HealthRules: &[]FLPHealthRule{
								{
									Template: HealthRulePacketDropsByKernel,
									Variants: []HealthRuleVariant{
										{
											GroupBy: GroupByNode,
											Thresholds: HealthRuleThresholds{
												Info: "5.5",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "Missing metrics for alerts",
			ocpVersion: "4.18.0",
			fc: &FlowCollector{
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{EBPF: FlowCollectorEBPF{
						Features:   []AgentFeature{PacketDrop},
						Privileged: true,
					}},
					Processor: FlowCollectorFLP{
						Metrics: FLPMetrics{
							HealthRules: &[]FLPHealthRule{
								{
									Template: HealthRulePacketDropsByKernel,
									Variants: []HealthRuleVariant{
										{
											GroupBy: GroupByNode,
											Thresholds: HealthRuleThresholds{
												Info: "5",
											},
										},
									},
								},
							},
							IncludeList: &[]FLPMetric{"node_ingress_bytes_total"},
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{
				"HealthRule PacketDropsByKernel/Node requires enabling at least one metric from this list: node_drop_packets_total",
				"HealthRule PacketDropsByKernel/Node requires enabling at least one metric from this list: node_ingress_packets_total, node_egress_packets_total",
			},
		},
		{
			name:       "Invalid alert threshold",
			ocpVersion: "4.18.0",
			fc: &FlowCollector{
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{EBPF: FlowCollectorEBPF{
						Features:   []AgentFeature{PacketDrop},
						Privileged: true,
					}},
					Processor: FlowCollectorFLP{
						Metrics: FLPMetrics{
							HealthRules: &[]FLPHealthRule{
								{
									Template: HealthRulePacketDropsByKernel,
									Variants: []HealthRuleVariant{
										{
											Thresholds: HealthRuleThresholds{
												Info: "nope",
											},
										},
									},
								},
							},
							IncludeList: &[]FLPMetric{"node_drop_packets_total", "node_ingress_packets_total"},
						},
					},
				},
			},
			expectedError: `cannot parse info threshold as float in spec.processor.metrics.healthRules[0].variants[0]: "nope"`,
		},
		{
			name:       "Invalid alert threshold severities",
			ocpVersion: "4.18.0",
			fc: &FlowCollector{
				Spec: FlowCollectorSpec{
					Agent: FlowCollectorAgent{EBPF: FlowCollectorEBPF{
						Features:   []AgentFeature{PacketDrop},
						Privileged: true,
					}},
					Processor: FlowCollectorFLP{
						Metrics: FLPMetrics{
							HealthRules: &[]FLPHealthRule{
								{
									Template: HealthRulePacketDropsByKernel,
									Variants: []HealthRuleVariant{
										{
											Thresholds: HealthRuleThresholds{
												Info:     "5",
												Warning:  "50",
												Critical: "10",
											},
										},
									},
								},
							},
							IncludeList: &[]FLPMetric{"node_drop_packets_total", "node_ingress_packets_total"},
						},
					},
				},
			},
			expectedError: `warning threshold must be lower than 10, which is defined for a higher severity`,
		},
	}

	CurrentClusterInfo = &cluster.Info{}
	r := FlowCollector{}
	for _, test := range tests {
		CurrentClusterInfo.Mock(test.ocpVersion, "")
		warnings, err := r.Validate(context.TODO(), test.fc)
		if test.expectedError == "" {
			assert.NoError(t, err, test.name)
		} else {
			assert.ErrorContains(t, err, test.expectedError, test.name)
		}
		assert.Equal(t, test.expectedWarnings, warnings, test.name)
	}
}

func TestValidateScheduling(t *testing.T) {
	mismatchWarning := "Mismatch detected between spec.agent.ebpf.advanced.scheduling and spec.processor.advanced.scheduling. In Direct mode, it can lead to inconsistent pod scheduling that would result in errors in the flow collection process."
	tests := []struct {
		name             string
		fc               *FlowCollector
		expectedError    string
		expectedWarnings admission.Warnings
		ocpVersion       string
	}{
		{
			name: "Valid default (Direct)",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					DeploymentModel: DeploymentModelDirect,
				},
			},
		},
		{
			name: "Invalid Agent scheduling",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					DeploymentModel: DeploymentModelDirect,
					Agent: FlowCollectorAgent{
						EBPF: FlowCollectorEBPF{
							Advanced: &AdvancedAgentConfig{
								Scheduling: &SchedulingConfig{
									Tolerations: []corev1.Toleration{{Key: "key"}},
								},
							},
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{mismatchWarning},
		},
		{
			name: "Invalid FLP scheduling",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					DeploymentModel: DeploymentModelDirect,
					Processor: FlowCollectorFLP{
						Advanced: &AdvancedProcessorConfig{
							Scheduling: &SchedulingConfig{
								Tolerations: []corev1.Toleration{{Key: "key"}},
							},
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{mismatchWarning},
		},
		{
			name: "Invalid FLP and Agent scheduling",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					DeploymentModel: DeploymentModelDirect,
					Agent: FlowCollectorAgent{
						EBPF: FlowCollectorEBPF{
							Advanced: &AdvancedAgentConfig{
								Scheduling: &SchedulingConfig{
									Tolerations: []corev1.Toleration{{Key: "key1"}},
								},
							},
						},
					},
					Processor: FlowCollectorFLP{
						Advanced: &AdvancedProcessorConfig{
							Scheduling: &SchedulingConfig{
								Tolerations: []corev1.Toleration{{Key: "key2"}},
							},
						},
					},
				},
			},
			expectedWarnings: admission.Warnings{mismatchWarning},
		},
		{
			name: "Valid FLP and Agent scheduling",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					DeploymentModel: DeploymentModelDirect,
					Agent: FlowCollectorAgent{
						EBPF: FlowCollectorEBPF{
							Advanced: &AdvancedAgentConfig{
								Scheduling: &SchedulingConfig{
									Tolerations: []corev1.Toleration{{Key: "same_key"}},
								},
							},
						},
					},
					Processor: FlowCollectorFLP{
						Advanced: &AdvancedProcessorConfig{
							Scheduling: &SchedulingConfig{
								Tolerations: []corev1.Toleration{{Key: "same_key"}},
							},
						},
					},
				},
			},
		},
		{
			name: "Valid default (Kafka)",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					DeploymentModel: DeploymentModelKafka,
				},
			},
		},
		{
			name: "No inconsistent scheduling with Kafka",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					Processor: FlowCollectorFLP{
						Advanced: &AdvancedProcessorConfig{
							Scheduling: &SchedulingConfig{
								Tolerations: []corev1.Toleration{{Key: "key"}},
							},
						},
					},
				},
			},
		},
	}

	CurrentClusterInfo = &cluster.Info{}
	r := FlowCollector{}
	for _, test := range tests {
		CurrentClusterInfo.Mock(test.ocpVersion, "")
		warnings, err := r.Validate(context.TODO(), test.fc)
		if test.expectedError == "" {
			assert.NoError(t, err, test.name)
		} else {
			assert.ErrorContains(t, err, test.expectedError, test.name)
		}
		assert.Equal(t, test.expectedWarnings, warnings, test.name)
	}
}

func TestElligibleMetrics(t *testing.T) {
	met, tot := GetElligibleMetricsForHealthRule(HealthRulePacketDropsByKernel, &HealthRuleVariant{
		GroupBy: GroupByNamespace,
	})
	assert.Equal(t, []string{"namespace_drop_packets_total", "workload_drop_packets_total"}, met)
	assert.Equal(t, []string{"namespace_ingress_packets_total", "workload_ingress_packets_total", "namespace_egress_packets_total", "workload_egress_packets_total"}, tot)

	met, tot = GetElligibleMetricsForHealthRule(HealthRulePacketDropsByKernel, &HealthRuleVariant{
		GroupBy: GroupByWorkload,
	})
	assert.Equal(t, []string{"workload_drop_packets_total"}, met)
	assert.Equal(t, []string{"workload_ingress_packets_total", "workload_egress_packets_total"}, tot)

	met, tot = GetElligibleMetricsForHealthRule(HealthRulePacketDropsByKernel, &HealthRuleVariant{
		GroupBy: GroupByNode,
	})
	assert.Equal(t, []string{"node_drop_packets_total"}, met)
	assert.Equal(t, []string{"node_ingress_packets_total", "node_egress_packets_total"}, tot)

	met, tot = GetElligibleMetricsForHealthRule(HealthRulePacketDropsByKernel, &HealthRuleVariant{})
	assert.Equal(t, []string{"namespace_drop_packets_total", "workload_drop_packets_total", "node_drop_packets_total"}, met)
	assert.Equal(t, []string{"namespace_ingress_packets_total", "workload_ingress_packets_total", "node_ingress_packets_total", "namespace_egress_packets_total", "workload_egress_packets_total", "node_egress_packets_total"}, tot)
}

func TestValidateNetPol(t *testing.T) {
	tests := []struct {
		name             string
		fc               *FlowCollector
		cni              cluster.NetworkType
		expectedError    string
		expectedWarnings admission.Warnings
	}{
		{
			name: "Empty config is valid for ovn-k",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{},
			},
			cni: cluster.OVNKubernetes,
		},
		{
			name: "Empty config is valid for sdn",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{},
			},
			cni: cluster.OpenShiftSDN,
		},
		{
			name: "Empty config is valid for unknown",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{},
			},
			cni: "unknown",
		},
		{
			name: "Enabled netpol is valid for ovn-k",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					NetworkPolicy: NetworkPolicy{Enable: ptr.To(true)},
				},
			},
			cni: cluster.OVNKubernetes,
		},
		{
			name: "Enabled netpol triggers warning for sdn",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					NetworkPolicy: NetworkPolicy{Enable: ptr.To(true)},
				},
			},
			cni:              cluster.OpenShiftSDN,
			expectedWarnings: admission.Warnings{"OpenShiftSDN detected with unsupported setting: spec.networkPolicy.enable; this setting will be ignored; to remove this warning set spec.networkPolicy.enable to false."},
		},
		{
			name: "Enabled netpol triggers warning for unknown",
			fc: &FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: FlowCollectorSpec{
					NetworkPolicy: NetworkPolicy{Enable: ptr.To(true)},
				},
			},
			cni:              "unknown",
			expectedWarnings: admission.Warnings{"Network policy is enabled via spec.networkPolicy.enable, despite not running OVN-Kubernetes: this configuration has not been tested; to remove this warning set spec.networkPolicy.enable to false."},
		},
	}

	CurrentClusterInfo = &cluster.Info{}
	for _, test := range tests {
		CurrentClusterInfo.Mock("4.20.0", test.cni)
		v := validator{fc: &test.fc.Spec}
		v.validateNetPol()
		if test.expectedError == "" {
			assert.Empty(t, v.errors, test.name)
		} else {
			assert.Len(t, v.errors, 1, test.name)
			assert.ErrorContains(t, v.errors[0], test.expectedError, test.name)
		}
		assert.Equal(t, test.expectedWarnings, v.warnings, test.name)
	}
}
