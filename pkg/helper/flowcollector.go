package helper

import (
	"strconv"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	netobservManagedLabel = "netobserv-managed"
)

func GetSampling(spec *flowslatest.FlowCollectorSpec) int {
	if spec.Agent.EBPF.Sampling == nil {
		return 50
	}
	return int(*spec.Agent.EBPF.Sampling)
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

func HPAEnabled(spec *flowslatest.FlowCollectorHPA) bool {
	return spec != nil && spec.Status == flowslatest.HPAStatusEnabled
}

func GetRecordTypes(processor *flowslatest.FlowCollectorFLP) []api.ConnTrackOutputRecordTypeEnum {
	if processor.LogTypes != nil {
		switch *processor.LogTypes {
		case flowslatest.LogTypeFlows:
			return []api.ConnTrackOutputRecordTypeEnum{api.ConnTrackFlowLog}
		case flowslatest.LogTypeConversations:
			return []api.ConnTrackOutputRecordTypeEnum{
				api.ConnTrackNewConnection,
				api.ConnTrackHeartbeat,
				api.ConnTrackEndConnection,
			}
		case flowslatest.LogTypeEndedConversations:
			return []api.ConnTrackOutputRecordTypeEnum{api.ConnTrackEndConnection}
		case flowslatest.LogTypeAll:
			return []api.ConnTrackOutputRecordTypeEnum{
				api.ConnTrackFlowLog,
				api.ConnTrackNewConnection,
				api.ConnTrackHeartbeat,
				api.ConnTrackEndConnection,
			}
		}
	}
	return []api.ConnTrackOutputRecordTypeEnum{api.ConnTrackFlowLog}
}

func UseSASL(cfg *flowslatest.SASLConfig) bool {
	return cfg.Type == flowslatest.SASLPlain || cfg.Type == flowslatest.SASLScramSHA512
}

func UseLoki(spec *flowslatest.FlowCollectorSpec) bool {
	// nil should fallback to default value, which is "true"
	return spec.Loki.Enable == nil || *spec.Loki.Enable
}

func UsePrometheus(spec *flowslatest.FlowCollectorSpec) bool {
	// nil should fallback to default value, which is "true"
	return spec.Prometheus.Querier.Enable == nil || *spec.Prometheus.Querier.Enable
}

func UseConsolePlugin(spec *flowslatest.FlowCollectorSpec) bool {
	return (UseLoki(spec) || UsePrometheus(spec)) &&
		// nil should fallback to default value, which is "true"
		(spec.ConsolePlugin.Enable == nil || *spec.ConsolePlugin.Enable)
}

func UseTestConsolePlugin(spec *flowslatest.FlowCollectorSpec) bool {
	if spec.ConsolePlugin.Advanced != nil {
		env := spec.ConsolePlugin.Advanced.Env[constants.EnvTestConsole]
		// Use ParseBool to allow common variants ("true", "True", "1"...) and ignore non-bools
		b, err := strconv.ParseBool(env)
		return err == nil && b
	}
	return false
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

func IsNetworkEventsEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return IsAgentFeatureEnabled(spec, flowslatest.NetworkEvents)
}

func IsPacketTranslationEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return IsAgentFeatureEnabled(spec, flowslatest.PacketTranslation)
}
func IsEbpfManagerEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return IsAgentFeatureEnabled(spec, flowslatest.EbpfManager)
}

func IsUDNMappingEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return IsAgentFeatureEnabled(spec, flowslatest.UDNMapping)
}

func IsConntrack(spec *flowslatest.FlowCollectorFLP) bool {
	return spec != nil && spec.LogTypes != nil && *spec.LogTypes != flowslatest.LogTypeFlows
}

func IsMultiClusterEnabled(spec *flowslatest.FlowCollectorFLP) bool {
	return spec != nil && spec.MultiClusterDeployment != nil && *spec.MultiClusterDeployment
}

func IsZoneEnabled(spec *flowslatest.FlowCollectorFLP) bool {
	return spec != nil && spec.AddZone != nil && *spec.AddZone
}

func IsEBPFMetricsEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return spec.Metrics.Enable == nil || *spec.Metrics.Enable
}

func IsSubnetLabelsEnabled(spec *flowslatest.FlowCollectorFLP) bool {
	return AutoDetectOpenShiftNetworks(spec) || len(spec.SubnetLabels.CustomLabels) > 0
}

func IsEBFPFlowFilterEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return spec.FlowFilter != nil && spec.FlowFilter.Enable != nil && *spec.FlowFilter.Enable
}

func GetEBPFMetricsPort(spec *flowslatest.FlowCollectorEBPF) int32 {
	port := int32(constants.EBPFMetricPort)
	if spec.Metrics.Server.Port != nil {
		port = *spec.Metrics.Server.Port
	}
	return port
}

func GetFlowCollectorMetricsPort(spec *flowslatest.FlowCollectorSpec) int32 {
	port := int32(constants.FLPMetricsPort)
	if spec.Processor.Metrics.Server.Port != nil {
		port = *spec.Processor.Metrics.Server.Port
	}
	return port
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

func AddOwnedLabel(obj client.Object) {
	// set netobserv-managed label to true so users can easily switch to false if they want to skip ownership
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[netobservManagedLabel] = "true"
	obj.SetLabels(labels)
}

func SkipOwnership(obj client.Object) bool {
	// ownership is ignored if netobserv-managed label is explicitly set to false
	labels := obj.GetLabels()
	return labels != nil && labels[netobservManagedLabel] == "false"
}

func IsOwned(obj client.Object) bool {
	if SkipOwnership(obj) {
		return false
	}
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
	cfg := flowslatest.AdvancedAgentConfig{
		Env: map[string]string{},
		Scheduling: &flowslatest.SchedulingConfig{
			NodeSelector:      map[string]string{},
			Tolerations:       []corev1.Toleration{{Operator: corev1.TolerationOpExists}},
			Affinity:          nil,
			PriorityClassName: "",
		},
	}

	if specConfig != nil {
		if len(specConfig.Env) > 0 {
			cfg.Env = specConfig.Env
		}
		if specConfig.Scheduling != nil {
			if len(specConfig.Scheduling.NodeSelector) > 0 {
				cfg.Scheduling.NodeSelector = specConfig.Scheduling.NodeSelector
			}
			if len(specConfig.Scheduling.Tolerations) > 0 {
				cfg.Scheduling.Tolerations = specConfig.Scheduling.Tolerations
			}
			if specConfig.Scheduling.Affinity != nil {
				cfg.Scheduling.Affinity = specConfig.Scheduling.Affinity
			}
			if len(specConfig.Scheduling.PriorityClassName) > 0 {
				cfg.Scheduling.PriorityClassName = specConfig.Scheduling.PriorityClassName
			}
		}
	}

	return cfg
}

func GetAdvancedProcessorConfig(specConfig *flowslatest.AdvancedProcessorConfig) flowslatest.AdvancedProcessorConfig {
	cfg := flowslatest.AdvancedProcessorConfig{
		Env:                            map[string]string{},
		Port:                           ptr.To(GetFieldDefaultInt32(ProcessorAdvancedPath, "port")),
		HealthPort:                     ptr.To(GetFieldDefaultInt32(ProcessorAdvancedPath, "healthPort")),
		ProfilePort:                    ptr.To(GetFieldDefaultInt32(ProcessorAdvancedPath, "profilePort")),
		EnableKubeProbes:               ptr.To(GetFieldDefaultBool(ProcessorAdvancedPath, "enableKubeProbes")),
		DropUnusedFields:               ptr.To(GetFieldDefaultBool(ProcessorAdvancedPath, "dropUnusedFields")),
		ConversationHeartbeatInterval:  ptr.To(GetFieldDefaultDuration(ProcessorAdvancedPath, "conversationHeartbeatInterval")),
		ConversationEndTimeout:         ptr.To(GetFieldDefaultDuration(ProcessorAdvancedPath, "conversationEndTimeout")),
		ConversationTerminatingTimeout: ptr.To(GetFieldDefaultDuration(ProcessorAdvancedPath, "conversationTerminatingTimeout")),
		Scheduling: &flowslatest.SchedulingConfig{
			NodeSelector:      map[string]string{},
			Tolerations:       []corev1.Toleration{{Operator: corev1.TolerationOpExists}},
			Affinity:          nil,
			PriorityClassName: "",
		},
	}

	if specConfig != nil {
		if len(specConfig.Env) > 0 {
			cfg.Env = specConfig.Env
		}
		if specConfig.Port != nil && *specConfig.Port > 0 {
			cfg.Port = specConfig.Port
		}
		if specConfig.HealthPort != nil && *specConfig.HealthPort > 0 {
			cfg.HealthPort = specConfig.HealthPort
		}
		if specConfig.ProfilePort != nil && *specConfig.ProfilePort > 0 {
			cfg.ProfilePort = specConfig.ProfilePort
		}
		if specConfig.EnableKubeProbes != nil {
			cfg.EnableKubeProbes = specConfig.EnableKubeProbes
		}
		if specConfig.DropUnusedFields != nil {
			cfg.DropUnusedFields = specConfig.DropUnusedFields
		}
		if specConfig.ConversationHeartbeatInterval != nil {
			cfg.ConversationHeartbeatInterval = specConfig.ConversationHeartbeatInterval
		}
		if specConfig.ConversationEndTimeout != nil {
			cfg.ConversationEndTimeout = specConfig.ConversationEndTimeout
		}
		if specConfig.ConversationTerminatingTimeout != nil {
			cfg.ConversationTerminatingTimeout = specConfig.ConversationTerminatingTimeout
		}
		if specConfig.Scheduling != nil {
			if len(specConfig.Scheduling.NodeSelector) > 0 {
				cfg.Scheduling.NodeSelector = specConfig.Scheduling.NodeSelector
			}
			if len(specConfig.Scheduling.Tolerations) > 0 {
				cfg.Scheduling.Tolerations = specConfig.Scheduling.Tolerations
			}
			if specConfig.Scheduling.Affinity != nil {
				cfg.Scheduling.Affinity = specConfig.Scheduling.Affinity
			}
			if len(specConfig.Scheduling.PriorityClassName) > 0 {
				cfg.Scheduling.PriorityClassName = specConfig.Scheduling.PriorityClassName
			}
		}
	}

	return cfg
}

func GetAdvancedLokiConfig(specConfig *flowslatest.AdvancedLokiConfig) flowslatest.AdvancedLokiConfig {
	cfg := flowslatest.AdvancedLokiConfig{
		WriteMinBackoff: ptr.To(GetFieldDefaultDuration(LokiAdvancedPath, "writeMinBackoff")),
		WriteMaxBackoff: ptr.To(GetFieldDefaultDuration(LokiAdvancedPath, "writeMaxBackoff")),
		WriteMaxRetries: ptr.To(GetFieldDefaultInt32(LokiAdvancedPath, "writeMaxRetries")),
		StaticLabels:    GetFieldDefaultMapString(LokiAdvancedPath, "staticLabels"),
	}

	if specConfig != nil {
		if specConfig.WriteMinBackoff != nil {
			cfg.WriteMinBackoff = specConfig.WriteMinBackoff
		}
		if specConfig.WriteMaxBackoff != nil {
			cfg.WriteMaxBackoff = specConfig.WriteMaxBackoff
		}
		if specConfig.WriteMaxRetries != nil {
			cfg.WriteMaxRetries = specConfig.WriteMaxRetries
		}
		if specConfig.StaticLabels != nil {
			cfg.StaticLabels = specConfig.StaticLabels
		}
	}

	return cfg
}

func GetAdvancedPluginConfig(specConfig *flowslatest.AdvancedPluginConfig) flowslatest.AdvancedPluginConfig {
	cfg := flowslatest.AdvancedPluginConfig{
		Env:      map[string]string{},
		Args:     []string{},
		Register: ptr.To(GetFieldDefaultBool(PluginAdvancedPath, "register")),
		Port:     ptr.To(GetFieldDefaultInt32(PluginAdvancedPath, "port")),
		Scheduling: &flowslatest.SchedulingConfig{
			NodeSelector:      map[string]string{},
			Tolerations:       []corev1.Toleration{{Operator: corev1.TolerationOpExists}},
			Affinity:          nil,
			PriorityClassName: "",
		},
	}

	if specConfig != nil {
		if len(specConfig.Env) > 0 {
			cfg.Env = specConfig.Env
		}
		if len(specConfig.Args) > 0 {
			cfg.Args = specConfig.Args
		}
		if specConfig.Register != nil {
			cfg.Register = specConfig.Register
		}
		if specConfig.Port != nil && *specConfig.Port > 0 {
			cfg.Port = specConfig.Port
		}
		if specConfig.Scheduling != nil {
			if len(specConfig.Scheduling.NodeSelector) > 0 {
				cfg.Scheduling.NodeSelector = specConfig.Scheduling.NodeSelector
			}
			if len(specConfig.Scheduling.Tolerations) > 0 {
				cfg.Scheduling.Tolerations = specConfig.Scheduling.Tolerations
			}
			if specConfig.Scheduling.Affinity != nil {
				cfg.Scheduling.Affinity = specConfig.Scheduling.Affinity
			}
			if len(specConfig.Scheduling.PriorityClassName) > 0 {
				cfg.Scheduling.PriorityClassName = specConfig.Scheduling.PriorityClassName
			}
		}
	}

	return cfg
}

func AutoDetectOpenShiftNetworks(spec *flowslatest.FlowCollectorFLP) bool {
	return spec.SubnetLabels.OpenShiftAutoDetect == nil || *spec.SubnetLabels.OpenShiftAutoDetect
}

func HasFLPDeduper(spec *flowslatest.FlowCollectorSpec) bool {
	return spec.Processor.Deduper != nil && spec.Processor.Deduper.Mode != "" && spec.Processor.Deduper.Mode != flowslatest.FLPDeduperDisabled
}
