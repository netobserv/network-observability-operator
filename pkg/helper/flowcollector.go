package helper

import (
	"strconv"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSampling(spec *flowslatest.FlowCollectorSpec) int {
	if spec.Agent.EBPF.Sampling == nil {
		return 50
	}
	return int(*spec.Agent.EBPF.Sampling)
}

func UseDedupJustMark(spec *flowslatest.FlowCollectorSpec) bool {
	if spec.Agent.EBPF.Advanced != nil {
		if v, ok := spec.Agent.EBPF.Advanced.Env["DEDUPER_JUST_MARK"]; ok {
			b, _ := strconv.ParseBool(v)
			return b
		}
	}
	// default true
	return true
}

func UseDedupMerge(spec *flowslatest.FlowCollectorSpec) bool {
	if spec.Agent.EBPF.Advanced != nil {
		if v, ok := spec.Agent.EBPF.Advanced.Env["DEDUPER_MERGE"]; ok {
			b, _ := strconv.ParseBool(v)
			return b
		}
	}
	// default false
	return false
}

func UseKafka(spec *flowslatest.FlowCollectorSpec) bool {
	return spec.DeploymentModel == flowslatest.DeploymentModelKafka
}

func UseMergedAgentFLP(spec *flowslatest.FlowCollectorSpec) bool {
	return spec.DeploymentModel == flowslatest.DeploymentModelDirect && spec.Agent.Type == flowslatest.AgentEBPF
}

func HasKafkaExporter(spec *flowslatest.FlowCollectorSpec) bool {
	for _, ex := range spec.Exporters {
		if ex.Type == flowslatest.KafkaExporter {
			return true
		}
	}
	return false
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

func IsAgentFeatureEnabled(spec *flowslatest.FlowCollectorEBPF, feature flowslatest.AgentFeature) bool {
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
	if IsPrivileged(spec) && IsAgentFeatureEnabled(spec, flowslatest.PacketDrop) {
		return true
	}
	return false
}

func IsDNSTrackingEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return IsAgentFeatureEnabled(spec, flowslatest.DNSTracking)
}

func IsFlowRTTEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return IsAgentFeatureEnabled(spec, flowslatest.FlowRTT)
}

func IsMultiClusterEnabled(spec *flowslatest.FlowCollectorFLP) bool {
	return spec.MultiClusterDeployment != nil && *spec.MultiClusterDeployment
}

func IsZoneEnabled(spec *flowslatest.FlowCollectorFLP) bool {
	return spec.AddZone != nil && *spec.AddZone
}

func IsEBPFMetricsEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return spec.Metrics.Enable != nil && *spec.Metrics.Enable
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

func GetNamespace(spec *flowslatest.FlowCollectorSpec) string {
	if spec.Namespace != "" {
		return spec.Namespace
	}
	return constants.DefaultOperatorNamespace
}

func GetAdvancedAgentConfig(specConfig *flowslatest.AdvancedAgentConfig) flowslatest.AdvancedAgentConfig {
	debugConfig := flowslatest.AdvancedAgentConfig{
		Env: map[string]string{},
	}

	if specConfig != nil {
		if len(specConfig.Env) > 0 {
			debugConfig.Env = specConfig.Env
		}
	}

	return debugConfig
}

func GetAdvancedProcessorConfig(specConfig *flowslatest.AdvancedProcessorConfig) flowslatest.AdvancedProcessorConfig {
	debugConfig := flowslatest.AdvancedProcessorConfig{
		Env:                            map[string]string{},
		Port:                           ptr.To(GetFieldDefaultInt32(ProcessorAdvancedPath, "port")),
		HealthPort:                     ptr.To(GetFieldDefaultInt32(ProcessorAdvancedPath, "healthPort")),
		ProfilePort:                    ptr.To(GetFieldDefaultInt32(ProcessorAdvancedPath, "profilePort")),
		EnableKubeProbes:               ptr.To(GetFieldDefaultBool(ProcessorAdvancedPath, "enableKubeProbes")),
		DropUnusedFields:               ptr.To(GetFieldDefaultBool(ProcessorAdvancedPath, "dropUnusedFields")),
		ConversationHeartbeatInterval:  ptr.To(GetFieldDefaultDuration(ProcessorAdvancedPath, "conversationHeartbeatInterval")),
		ConversationEndTimeout:         ptr.To(GetFieldDefaultDuration(ProcessorAdvancedPath, "conversationEndTimeout")),
		ConversationTerminatingTimeout: ptr.To(GetFieldDefaultDuration(ProcessorAdvancedPath, "conversationTerminatingTimeout")),
	}

	if specConfig != nil {
		if len(specConfig.Env) > 0 {
			debugConfig.Env = specConfig.Env
		}
		if specConfig.Port != nil && *specConfig.Port > 0 {
			debugConfig.Port = specConfig.Port
		}
		if specConfig.HealthPort != nil && *specConfig.HealthPort > 0 {
			debugConfig.HealthPort = specConfig.HealthPort
		}
		if specConfig.ProfilePort != nil && *specConfig.ProfilePort > 0 {
			debugConfig.ProfilePort = specConfig.ProfilePort
		}
		if specConfig.EnableKubeProbes != nil {
			debugConfig.EnableKubeProbes = specConfig.EnableKubeProbes
		}
		if specConfig.DropUnusedFields != nil {
			debugConfig.DropUnusedFields = specConfig.DropUnusedFields
		}
		if specConfig.ConversationHeartbeatInterval != nil {
			debugConfig.ConversationHeartbeatInterval = specConfig.ConversationHeartbeatInterval
		}
		if specConfig.ConversationEndTimeout != nil {
			debugConfig.ConversationEndTimeout = specConfig.ConversationEndTimeout
		}
		if specConfig.ConversationTerminatingTimeout != nil {
			debugConfig.ConversationTerminatingTimeout = specConfig.ConversationTerminatingTimeout
		}
	}

	return debugConfig
}

func GetAdvancedLokiConfig(specConfig *flowslatest.AdvancedLokiConfig) flowslatest.AdvancedLokiConfig {
	debugConfig := flowslatest.AdvancedLokiConfig{
		WriteMinBackoff: ptr.To(GetFieldDefaultDuration(LokiAdvancedPath, "writeMinBackoff")),
		WriteMaxBackoff: ptr.To(GetFieldDefaultDuration(LokiAdvancedPath, "writeMaxBackoff")),
		WriteMaxRetries: ptr.To(GetFieldDefaultInt32(LokiAdvancedPath, "writeMaxRetries")),
		StaticLabels:    GetFieldDefaultMapString(LokiAdvancedPath, "staticLabels"),
	}

	if specConfig != nil {
		if specConfig.WriteMinBackoff != nil {
			debugConfig.WriteMinBackoff = specConfig.WriteMinBackoff
		}
		if specConfig.WriteMaxBackoff != nil {
			debugConfig.WriteMaxBackoff = specConfig.WriteMaxBackoff
		}
		if specConfig.WriteMaxRetries != nil {
			debugConfig.WriteMaxRetries = specConfig.WriteMaxRetries
		}
		if specConfig.StaticLabels != nil {
			debugConfig.StaticLabels = specConfig.StaticLabels
		}
	}

	return debugConfig
}

func GetAdvancedPluginConfig(specConfig *flowslatest.AdvancedPluginConfig) flowslatest.AdvancedPluginConfig {
	debugConfig := flowslatest.AdvancedPluginConfig{
		Env:      map[string]string{},
		Args:     []string{},
		Register: ptr.To(GetFieldDefaultBool(PluginAdvancedPath, "register")),
		Port:     ptr.To(GetFieldDefaultInt32(PluginAdvancedPath, "port")),
	}

	if specConfig != nil {
		if len(specConfig.Env) > 0 {
			debugConfig.Env = specConfig.Env
		}
		if len(specConfig.Args) > 0 {
			debugConfig.Args = specConfig.Args
		}
		if specConfig.Register != nil {
			debugConfig.Register = specConfig.Register
		}
		if specConfig.Port != nil && *specConfig.Port > 0 {
			debugConfig.Port = specConfig.Port
		}
	}

	return debugConfig
}
