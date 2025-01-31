package ebpf

import (
	"context"
	"encoding/binary"
	"fmt"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"

	bpfmaniov1alpha1 "github.com/bpfman/bpfman-operator/apis/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	netobservApp = "netobserv"
)

// bpfmanDetachNetobserv find BpfmanApplication object with all required ebpf hooks and detaches them using bpfman manager
func (c *AgentController) bpfmanDetachNetobserv(ctx context.Context) error {
	bpfApp := bpfmaniov1alpha1.ClusterBpfApplication{
		ObjectMeta: v1.ObjectMeta{
			Name: netobservApp,
		},
		TypeMeta: v1.TypeMeta{
			Kind: "BpfApplication",
		},
	}

	key := client.ObjectKey{Name: netobservApp}

	err := c.Get(ctx, key, &bpfApp)
	if err != nil {
		return fmt.Errorf("failed to get BpfApplication: %w", err)
	}

	err = c.deleteBpfApplication(ctx, &bpfApp)
	if err != nil {
		return fmt.Errorf("failed to delete BpfApplication: %w", err)
	}
	return nil
}

// bpfmanAttachNetobserv Creates BpfmanApplication object with all required ebpf hooks and attaches them using bpfman manager
func (c *AgentController) bpfmanAttachNetobserv(ctx context.Context, fc *flowslatest.FlowCollector) error {
	var err error
	bpfApp := bpfmaniov1alpha1.ClusterBpfApplication{
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

func prepareBpfApplication(bpfApp *bpfmaniov1alpha1.ClusterBpfApplication, fc *flowslatest.FlowCollector, netobservBCImage string) {
	samplingValue := make([]byte, 4)
	dnsPortValue := make([]byte, 2)
	var enableDNSValue, enableRTTValue, enableFLowFilterValue, enableNetworkEvents, traceValue, networkEventsGroupIDValue, enablePktTranslation, enableIPSecValue []byte

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

	if helper.IsEBPFFlowFilterEnabled(&fc.Spec.Agent.EBPF) {
		enableFLowFilterValue = append(enableFLowFilterValue, uint8(1))
	}

	if helper.IsNetworkEventsEnabled(&fc.Spec.Agent.EBPF) {
		enableNetworkEvents = append(enableNetworkEvents, uint8(1))
	}

	if helper.IsPacketTranslationEnabled(&fc.Spec.Agent.EBPF) {
		enablePktTranslation = append(enablePktTranslation, uint8(1))
	}

	if helper.IsIPSecEnabled(&fc.Spec.Agent.EBPF) {
		enableIPSecValue = append(enableIPSecValue, uint8(1))
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
		"enable_ipsec":                      enableIPSecValue,
	}

	bpfApp.Spec.BpfAppCommon.ByteCode = bpfmaniov1alpha1.ByteCodeSelector{
		Image: &bpfmaniov1alpha1.ByteCodeImage{
			Url:             netobservBCImage,
			ImagePullPolicy: bpfmaniov1alpha1.PullIfNotPresent,
		},
	}
	bpfApp.Spec.Programs = []bpfmaniov1alpha1.ClBpfApplicationProgram{
		{
			Name: "tcx_ingress_flow_parse",
			Type: bpfmaniov1alpha1.ProgTypeTCX,
			TCX: &bpfmaniov1alpha1.ClTcxProgramInfo{
				Links: []bpfmaniov1alpha1.ClTcxAttachInfo{
					{
						InterfaceSelector: bpfmaniov1alpha1.InterfaceSelector{
							InterfacesDiscoveryConfig: &bpfmaniov1alpha1.InterfaceDiscovery{
								InterfaceAutoDiscovery: ptr.To(true)},
						},
						Direction: bpfmaniov1alpha1.TCIngress,
					},
				},
			},
		},
		{
			Name: "tcx_egress_flow_parse",
			Type: bpfmaniov1alpha1.ProgTypeTCX,
			TCX: &bpfmaniov1alpha1.ClTcxProgramInfo{
				Links: []bpfmaniov1alpha1.ClTcxAttachInfo{
					{
						InterfaceSelector: bpfmaniov1alpha1.InterfaceSelector{
							InterfacesDiscoveryConfig: &bpfmaniov1alpha1.InterfaceDiscovery{
								InterfaceAutoDiscovery: ptr.To(true)},
						},
						Direction: bpfmaniov1alpha1.TCEgress,
					},
				},
			},
		},
	}

	if helper.IsFlowRTTEnabled(&fc.Spec.Agent.EBPF) {
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.ClBpfApplicationProgram{
			{
				Name: "tcp_rcv_fentry",
				Type: bpfmaniov1alpha1.ProgTypeFentry,
				FEntry: &bpfmaniov1alpha1.ClFentryProgramInfo{
					ClFentryLoadInfo: bpfmaniov1alpha1.ClFentryLoadInfo{
						Function: "tcp_rcv_established",
					},
					Links: []bpfmaniov1alpha1.ClFentryAttachInfo{
						{
							Mode: bpfmaniov1alpha1.Attach,
						},
					},
				},
			},
		}...)
	}

	if helper.IsNetworkEventsEnabled(&fc.Spec.Agent.EBPF) {
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.ClBpfApplicationProgram{
			{
				Name: "network_events_monitoring",
				Type: bpfmaniov1alpha1.ProgTypeKprobe,
				KProbe: &bpfmaniov1alpha1.ClKprobeProgramInfo{
					Links: []bpfmaniov1alpha1.ClKprobeAttachInfo{
						{
							Function: "psample_sample_packet",
						},
					},
				},
			},
		}...)
	}

	if helper.IsPktDropEnabled(&fc.Spec.Agent.EBPF) {
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.ClBpfApplicationProgram{
			{
				Name: "kfree_skb",
				Type: bpfmaniov1alpha1.ProgTypeTracepoint,
				TracePoint: &bpfmaniov1alpha1.ClTracepointProgramInfo{
					Links: []bpfmaniov1alpha1.ClTracepointAttachInfo{
						{
							Name: "skb/kfree_skb",
						},
					},
				},
			},
		}...)
	}

	if helper.IsPacketTranslationEnabled(&fc.Spec.Agent.EBPF) {
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.ClBpfApplicationProgram{
			{
				Name: "track_nat_manip_pkt",
				Type: bpfmaniov1alpha1.ProgTypeKprobe,
				KProbe: &bpfmaniov1alpha1.ClKprobeProgramInfo{
					Links: []bpfmaniov1alpha1.ClKprobeAttachInfo{
						{
							Function: "nf_nat_manip_pkt",
						},
					},
				},
			},
		}...)
	}

	if helper.IsIPSecEnabled(&fc.Spec.Agent.EBPF) {
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.ClBpfApplicationProgram{
			{
				Name: "xfrm_input_kprobe",
				Type: bpfmaniov1alpha1.ProgTypeKprobe,
				KProbe: &bpfmaniov1alpha1.ClKprobeProgramInfo{
					Links: []bpfmaniov1alpha1.ClKprobeAttachInfo{
						{
							Function: "xfrm_input",
						},
					},
				},
			},
		}...)
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.ClBpfApplicationProgram{
			{
				Name: "xfrm_input_kretprobe",
				Type: bpfmaniov1alpha1.ProgTypeKretprobe,
				KRetProbe: &bpfmaniov1alpha1.ClKretprobeProgramInfo{
					Links: []bpfmaniov1alpha1.ClKretprobeAttachInfo{
						{
							Function: "xfrm_input",
						},
					},
				},
			},
		}...)
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.ClBpfApplicationProgram{
			{
				Name: "xfrm_output_kprobe",
				Type: bpfmaniov1alpha1.ProgTypeKprobe,
				KProbe: &bpfmaniov1alpha1.ClKprobeProgramInfo{
					Links: []bpfmaniov1alpha1.ClKprobeAttachInfo{
						{
							Function: "xfrm_output",
						},
					},
				},
			},
		}...)
		bpfApp.Spec.Programs = append(bpfApp.Spec.Programs, []bpfmaniov1alpha1.ClBpfApplicationProgram{
			{
				Name: "xfrm_output_kretprobe",
				Type: bpfmaniov1alpha1.ProgTypeKretprobe,
				KRetProbe: &bpfmaniov1alpha1.ClKretprobeProgramInfo{
					Links: []bpfmaniov1alpha1.ClKretprobeAttachInfo{
						{
							Function: "xfrm_output",
						},
					},
				},
			},
		}...)
	}
}

func (c *AgentController) deleteBpfApplication(ctx context.Context, bpfApp *bpfmaniov1alpha1.ClusterBpfApplication) error {
	klog.Info("Deleting BpfApplication Object")
	return c.Delete(ctx, bpfApp)
}

func (c *AgentController) createBpfApplication(ctx context.Context, bpfApp *bpfmaniov1alpha1.ClusterBpfApplication) error {
	return c.CreateOwned(ctx, bpfApp)
}

func (c *AgentController) updateBpfApplication(ctx context.Context, bpfApp *bpfmaniov1alpha1.ClusterBpfApplication) error {
	return c.UpdateOwned(ctx, bpfApp, bpfApp)
}
