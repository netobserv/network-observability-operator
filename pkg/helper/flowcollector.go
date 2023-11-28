package helper

import (
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSampling(spec *flowslatest.FlowCollectorSpec) int {
	if UseEBPF(spec) {
		return int(*spec.Agent.EBPF.Sampling)
	}
	return int(spec.Agent.IPFIX.Sampling)
}

func UseEBPF(spec *flowslatest.FlowCollectorSpec) bool {
	return spec.Agent.Type == flowslatest.AgentEBPF
}

func UseIPFIX(spec *flowslatest.FlowCollectorSpec) bool {
	return spec.Agent.Type == flowslatest.AgentIPFIX
}

func UseKafka(spec *flowslatest.FlowCollectorSpec) bool {
	return spec.DeploymentModel == flowslatest.DeploymentModelKafka
}

func HasKafkaExporter(spec *flowslatest.FlowCollectorSpec) bool {
	for _, ex := range spec.Exporters {
		if ex.Type == flowslatest.KafkaExporter {
			return true
		}
	}
	return false
}

func HPADisabled(spec *flowslatest.FlowCollectorHPA) bool {
	return spec == nil || spec.Status == flowslatest.HPAStatusDisabled
}

func HPAEnabled(spec *flowslatest.FlowCollectorHPA) bool {
	return spec != nil && spec.Status == flowslatest.HPAStatusEnabled
}

func GetRecordTypes(processor *flowslatest.FlowCollectorFLP) []string {
	outputRecordTypes := []string{constants.FlowLogType}
	if processor.LogTypes != nil {
		switch *processor.LogTypes {
		case flowslatest.LogTypeFlows:
			outputRecordTypes = []string{
				constants.FlowLogType,
			}
		case flowslatest.LogTypeConversations:
			outputRecordTypes = []string{
				constants.NewConnectionType,
				constants.HeartbeatType,
				constants.EndConnectionType,
			}
		case flowslatest.LogTypeEndedConversations:
			outputRecordTypes = []string{
				constants.EndConnectionType,
			}
		case flowslatest.LogTypeAll:
			outputRecordTypes = []string{
				constants.FlowLogType,
				constants.NewConnectionType,
				constants.HeartbeatType,
				constants.EndConnectionType,
			}
		}
	}
	return outputRecordTypes
}

func UseSASL(cfg *flowslatest.SASLConfig) bool {
	return cfg.Type == flowslatest.SASLPlain || cfg.Type == flowslatest.SASLScramSHA512
}

func UseLoki(spec *flowslatest.FlowCollectorSpec) bool {
	// nil should fallback to default value, which is "true"
	return spec.Loki.Enable == nil || *spec.Loki.Enable
}

func UseConsolePlugin(spec *flowslatest.FlowCollectorSpec) bool {
	return UseLoki(spec) &&
		// nil should fallback to default value, which is "true"
		(spec.ConsolePlugin.Enable == nil || *spec.ConsolePlugin.Enable)
}

func IsFeatureEnabled(spec *flowslatest.FlowCollectorEBPF, feature flowslatest.AgentFeature) bool {
	for _, f := range spec.Features {
		if f == feature {
			return true
		}
	}
	return false
}

func IsPrivileged(spec *flowslatest.FlowCollectorEBPF) bool {
	return spec.Privileged
}

func IsPktDropEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	if IsPrivileged(spec) && IsFeatureEnabled(spec, flowslatest.PacketDrop) {
		return true
	}
	return false
}

func IsDNSTrackingEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return IsFeatureEnabled(spec, flowslatest.DNSTracking)
}

func IsFlowRTTEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return IsFeatureEnabled(spec, flowslatest.FlowRTT)
}

func PtrBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func PtrInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func IsOwned(obj client.Object) bool {
	refs := obj.GetOwnerReferences()
	return len(refs) > 0 && strings.HasPrefix(refs[0].APIVersion, flowslatest.GroupVersion.Group)
}

func GetIncludeList(spec *flowslatest.FlowCollectorSpec) []string {
	var list []string
	if spec.Processor.Metrics.IncludeList == nil {
		list = metrics.DefaultIncludeList
	} else {
		for _, m := range *spec.Processor.Metrics.IncludeList {
			list = append(list, string(m))
		}
	}
	if !UseEBPF(spec) || !IsPktDropEnabled(&spec.Agent.EBPF) {
		list = removeMetricsByPattern(list, "_drop_")
	}
	if !UseEBPF(spec) || !IsFlowRTTEnabled(&spec.Agent.EBPF) {
		list = removeMetricsByPattern(list, "_rtt_")
	}
	if !UseEBPF(spec) || !IsDNSTrackingEnabled(&spec.Agent.EBPF) {
		list = removeMetricsByPattern(list, "_dns_")
	}
	return list
}

func removeMetricsByPattern(list []string, search string) []string {
	var filtered []string
	for _, m := range list {
		if !strings.Contains(m, search) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func GetNamespace(spec *flowslatest.FlowCollectorSpec) string {
	if spec.Namespace != "" {
		return spec.Namespace
	}
	return constants.DefaultOperatorNamespace
}
