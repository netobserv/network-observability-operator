package ebpf

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/sirupsen/logrus"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"

	bpfmaniov1alpha1 "github.com/bpfman/bpfman-operator/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	netobservApp = "netobserv"
)

// bpfmanAttachNetobserv Creates BpfmanApplication object with all required ebpf hooks and attaches them using bpfman manager
func (c *AgentController) bpfmanAttachNetobserv(ctx context.Context, fc *flowslatest.FlowCollector) error {
	var err error
	bpfApp := bpfmaniov1alpha1.BpfApplication{
		ObjectMeta: v1.ObjectMeta{
			Name: netobservApp,
		},
		TypeMeta: v1.TypeMeta{
			Kind: "BpfApplication",
		},
	}

	key := client.ObjectKey{Name: netobservApp}

	err = c.Get(ctx, key, &bpfApp)
	if err != nil {
		if errors.IsNotFound(err) {
			prepareBpfApplication(&bpfApp, fc, c.Images[constants.EBPFAgentByteCodeImageIndex])
			err = c.createBpfApplication(ctx, &bpfApp)
			if err != nil {
				return fmt.Errorf("failed to create BpfApplication: %w for obj: %s", err, fc.Name)
			}
		} else {
			return fmt.Errorf("failed to get BpfApplication: %w for obj: %s", err, fc.Name)
		}
	} else {
		// object exists repopulate it with the new configuration and update it
		prepareBpfApplication(&bpfApp, fc, c.Images[constants.EBPFAgentByteCodeImageIndex])
		err = c.updateBpfApplication(ctx, &bpfApp)
		if err != nil {
			return fmt.Errorf("failed to update BpfApplication: %w for obj: %s", err, fc.Name)
		}
	}

	return err
}

func prepareBpfApplication(bpfApp *bpfmaniov1alpha1.BpfApplication, fc *flowslatest.FlowCollector, netobservBCImage string) {
	interfaces := fc.Spec.Agent.EBPF.Interfaces

	samplingValue := make([]byte, 4)
	dnsPortValue := make([]byte, 2)
	var enableDNSValue, enableRTTValue, enableFLowFilterValue, enableNetworkEvents, traceValue, networkEventsGroupIDValue, enablePktTranslation []byte

	binary.NativeEndian.PutUint32(samplingValue, uint32(*fc.Spec.Agent.EBPF.Sampling))

	if fc.Spec.Agent.EBPF.LogLevel == logrus.TraceLevel.String() || fc.Spec.Agent.EBPF.LogLevel == logrus.DebugLevel.String() {
		traceValue = append(traceValue, uint8(1))
	}

	if helper.IsDNSTrackingEnabled(&fc.Spec.Agent.EBPF) {
		enableDNSValue = append(enableDNSValue, uint8(1))
	}

	if helper.IsFlowRTTEnabled(&fc.Spec.Agent.EBPF) {
		enableRTTValue = append(enableRTTValue, uint8(1))
	}

	if helper.IsEBFPFlowFilterEnabled(&fc.Spec.Agent.EBPF) {
		enableFLowFilterValue = append(enableFLowFilterValue, uint8(1))
	}

	if helper.IsNetworkEventsEnabled(&fc.Spec.Agent.EBPF) {
		enableNetworkEvents = append(enableNetworkEvents, uint8(1))
	}

	if helper.IsPacketTranslationEnabled(&fc.Spec.Agent.EBPF) {
		enablePktTranslation = append(enablePktTranslation, uint8(1))
	}

	bpfApp.Labels = map[string]string{
		"app": netobservApp,
	}

	if fc.Spec.Agent.EBPF.Advanced != nil {
		advancedConfig := helper.GetAdvancedAgentConfig(fc.Spec.Agent.EBPF.Advanced)
		for _, pair := range helper.KeySorted(advancedConfig.Env) {
			k, v := pair[0], pair[1]
			if k == envDNSTrackingPort {
				dnsPortValue = []byte(v)
			} else if k == envNetworkEventsGroupID {
				networkEventsGroupIDValue = []byte(v)
			}
		}

		if advancedConfig.Scheduling != nil {
			bpfApp.Spec.NodeSelector = v1.LabelSelector{MatchLabels: fc.Spec.Agent.EBPF.Advanced.Scheduling.NodeSelector}
		}
	}

	bpfApp.Spec.BpfAppCommon.GlobalData = map[string][]byte{
		"sampling":                          samplingValue,
		"trace_messages":                    traceValue,
		"enable_rtt":                        enableRTTValue,
		"enable_dns_tracking":               enableDNSValue,
		"dns_port":                          dnsPortValue,
		"enable_flows_filtering":            enableFLowFilterValue,
		"enable_network_events_monitoring":  enableNetworkEvents,
		"network_events_monitoring_groupid": networkEventsGroupIDValue,
		"enable_pkt_translation_tracking":   enablePktTranslation,
	}

	bpfApp.Spec.BpfAppCommon.ByteCode = bpfmaniov1alpha1.BytecodeSelector{
		Image: &bpfmaniov1alpha1.BytecodeImage{
			Url:             netobservBCImage,
			ImagePullPolicy: bpfmaniov1alpha1.PullIfNotPresent,
		},
	}
	bpfApp.Spec.Programs = []bpfmaniov1alpha1.BpfApplicationProgram{
		{
			Type: bpfmaniov1alpha1.ProgTypeTCX,
			TCX: &bpfmaniov1alpha1.TcxProgramInfo{
				BpfProgramCommon: bpfmaniov1alpha1.BpfProgramCommon{
					BpfFunctionName: "tcx_ingress_flow_parse",
				},
				InterfaceSelector: bpfmaniov1alpha1.InterfaceSelector{Interfaces: &interfaces},
				Direction:         "ingress",
			},
		},
		{
			Type: bpfmaniov1alpha1.ProgTypeTCX,
			TCX: &bpfmaniov1alpha1.TcxProgramInfo{
				BpfProgramCommon: bpfmaniov1alpha1.BpfProgramCommon{
					BpfFunctionName: "tcx_egress_flow_parse",
				},
				InterfaceSelector: bpfmaniov1alpha1.InterfaceSelector{Interfaces: &interfaces},
				Direction:         "egress",
			},
		},
	}

	if helper.IsFlowRTTEnabled(&fc.Spec.Agent.EBPF) {
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.BpfApplicationProgram{
			{
				Type: bpfmaniov1alpha1.ProgTypeFentry,
				Fentry: &bpfmaniov1alpha1.FentryProgramInfo{
					BpfProgramCommon: bpfmaniov1alpha1.BpfProgramCommon{
						BpfFunctionName: "tcp_rcv_fentry",
					},
					FunctionName: "tcp_rcv_established",
				},
			},
			{
				Type: bpfmaniov1alpha1.ProgTypeKprobe,
				Kprobe: &bpfmaniov1alpha1.KprobeProgramInfo{
					BpfProgramCommon: bpfmaniov1alpha1.BpfProgramCommon{
						BpfFunctionName: "tcp_rcv_kprobe",
					},
					FunctionName: "tcp_rcv_established",
					RetProbe:     false,
				},
			},
		}...)
	}

	if helper.IsNetworkEventsEnabled(&fc.Spec.Agent.EBPF) {
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.BpfApplicationProgram{
			{
				Type: bpfmaniov1alpha1.ProgTypeKprobe,
				Kprobe: &bpfmaniov1alpha1.KprobeProgramInfo{
					BpfProgramCommon: bpfmaniov1alpha1.BpfProgramCommon{
						BpfFunctionName: "rh_network_events_monitoring",
					},
					FunctionName: "rh_psample_sample_packet",
					RetProbe:     false,
				},
			},
		}...)
	}

	if helper.IsPktDropEnabled(&fc.Spec.Agent.EBPF) {
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.BpfApplicationProgram{
			{
				Type: bpfmaniov1alpha1.ProgTypeTracepoint,
				Tracepoint: &bpfmaniov1alpha1.TracepointProgramInfo{
					BpfProgramCommon: bpfmaniov1alpha1.BpfProgramCommon{
						BpfFunctionName: "kfree_skb",
					},
					Names: []string{"skb/kfree_skb"},
				},
			},
		}...)
	}

	if helper.IsPacketTranslationEnabled(&fc.Spec.Agent.EBPF) {
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.BpfApplicationProgram{
			{
				Type: bpfmaniov1alpha1.ProgTypeKprobe,
				Kprobe: &bpfmaniov1alpha1.KprobeProgramInfo{
					BpfProgramCommon: bpfmaniov1alpha1.BpfProgramCommon{
						BpfFunctionName: "track_nat_manip_pkt",
					},
					FunctionName: "nf_nat_manip_pkt",
					RetProbe:     false,
				},
			},
		}...)
	}
}

func (c *AgentController) createBpfApplication(ctx context.Context, bpfApp *bpfmaniov1alpha1.BpfApplication) error {
	return c.CreateOwned(ctx, bpfApp)
}

func (c *AgentController) updateBpfApplication(ctx context.Context, bpfApp *bpfmaniov1alpha1.BpfApplication) error {
	return c.UpdateOwned(ctx, bpfApp, bpfApp)
}
