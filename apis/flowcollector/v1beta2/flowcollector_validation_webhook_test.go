package v1beta2

import (
	"context"
	"testing"

	"github.com/netobserv/network-observability-operator/pkg/cluster"
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
							Features:   []AgentFeature{DNSTracking, FlowRTT, PacketDrop},
							Privileged: true,
							Sampling:   ptr.To(int32(100)),
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								FlowFilterRules: []EBPFFlowFilterRule{
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
							Features:   []AgentFeature{DNSTracking, FlowRTT, PacketDrop},
							Privileged: true,
							Sampling:   ptr.To(int32(100)),
							FlowFilter: &EBPFFlowFilter{
								Enable: ptr.To(true),
								FlowFilterRules: []EBPFFlowFilterRule{
									{
										Action:    "Accept",
										CIDR:      "0.0.0.0/0",
										Direction: "Egress",
										Protocol:  "TCP",
									},
									{
										Action:    "Accept",
										CIDR:      "0.0.0.0/0",
										Direction: "Egress",
										Protocol:  "UDP",
									},
								},
							},
						},
					},
				},
			},
			expectedError: "flow filter rule CIDR 0.0.0.0/0 already exists",
		},
		{
			name: "PacketDrop without privilege triggers warning",
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
			expectedWarnings: admission.Warnings{"The PacketDrop feature requires eBPF Agent to run in privileged mode"},
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
			expectedWarnings: admission.Warnings{"The NetworkEvents feature requires OpenShift 4.18 or above (version detected: 4.16.5)"},
		},
		{
			name:       "NetworkEvents without privilege triggers warning",
			ocpVersion: "4.18.0",
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
			expectedWarnings: admission.Warnings{"The NetworkEvents feature requires eBPF Agent to run in privileged mode"},
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
								FlowFilterRules: []EBPFFlowFilterRule{
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
								FlowFilterRules: []EBPFFlowFilterRule{
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
								FlowFilterRules: []EBPFFlowFilterRule{
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
								FlowFilterRules: []EBPFFlowFilterRule{
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
								FlowFilterRules: []EBPFFlowFilterRule{
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
								FlowFilterRules: []EBPFFlowFilterRule{
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
								FlowFilterRules: []EBPFFlowFilterRule{
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
								FlowFilterRules: []EBPFFlowFilterRule{
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
		warnings, errs := test.fc.validateAgent(context.TODO(), test.fc)
		if test.expectedError == "" {
			assert.Empty(t, errs, test.name)
		} else {
			assert.Len(t, errs, 1, test.name)
			assert.ErrorContains(t, errs[0], test.expectedError, test.name)
		}
		assert.Equal(t, test.expectedWarnings, warnings, test.name)
	}
}
