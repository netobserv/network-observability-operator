package v1beta2

import (
	"context"
	"testing"

	"github.com/netobserv/network-observability-operator/internal/pkg/cluster"
	"github.com/stretchr/testify/assert"
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
		CurrentClusterInfo.MockOpenShiftVersion(test.ocpVersion)
		warnings, errs := test.fc.validateAgent(context.TODO(), &test.fc.Spec)
		if test.expectedError == "" {
			assert.Empty(t, errs, test.name)
		} else {
			assert.Len(t, errs, 1, test.name)
			assert.ErrorContains(t, errs[0], test.expectedError, test.name)
		}
		assert.Equal(t, test.expectedWarnings, warnings, test.name)
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
	}

	CurrentClusterInfo = &cluster.Info{}
	for _, test := range tests {
		warnings, err := test.fc.Validate(context.TODO(), test.fc)
		if test.expectedError == "" {
			assert.Nil(t, err, test.name)
		} else {
			assert.ErrorContains(t, err, test.expectedError, test.name)
		}
		assert.Equal(t, test.expectedWarnings, warnings, test.name)
	}
}

func TestValidateFLPQueries(t *testing.T) {
	tests := []struct {
		name             string
		fc               *FlowCollector
		expectedError    string
		expectedWarnings admission.Warnings
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
	}

	CurrentClusterInfo = &cluster.Info{}
	for _, test := range tests {
		warnings, err := test.fc.Validate(context.TODO(), test.fc)
		if test.expectedError == "" {
			assert.Nil(t, err, test.name)
		} else {
			assert.ErrorContains(t, err, test.expectedError, test.name)
		}
		assert.Equal(t, test.expectedWarnings, warnings, test.name)
	}
}
