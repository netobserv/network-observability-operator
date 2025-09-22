package v1beta2

import (
	"strconv"

	"github.com/netobserv/network-observability-operator/internal/controller/constants"
)

func (spec *FlowCollectorSpec) GetNamespace() string {
	if spec.Namespace != "" {
		return spec.Namespace
	}
	return constants.DefaultOperatorNamespace
}

func (spec *FlowCollectorSpec) GetSampling() int {
	if spec.Agent.EBPF.Sampling == nil {
		return 50
	}
	return int(*spec.Agent.EBPF.Sampling)
}

func (spec *FlowCollectorSpec) UseKafka() bool {
	return spec.DeploymentModel == DeploymentModelKafka
}

func (spec *FlowCollectorSpec) HasKafkaExporter() bool {
	for _, ex := range spec.Exporters {
		if ex.Type == KafkaExporter {
			return true
		}
	}
	return false
}

func (spec *FlowCollectorHPA) HPAEnabled() bool {
	return spec != nil && spec.Status == HPAStatusEnabled
}

func (cfg *SASLConfig) UseSASL() bool {
	return cfg.Type == SASLPlain || cfg.Type == SASLScramSHA512
}

func (spec *FlowCollectorSpec) UseLoki() bool {
	// nil should fallback to default value, which is "true"
	return spec.Loki.Enable == nil || *spec.Loki.Enable
}

func (spec *FlowCollectorSpec) UsePrometheus() bool {
	// nil should fallback to default value, which is "true"
	return spec.Prometheus.Querier.Enable == nil || *spec.Prometheus.Querier.Enable
}

func (spec *FlowCollectorSpec) UseConsolePlugin() bool {
	return (spec.UseLoki() || spec.UsePrometheus()) &&
		// nil should fallback to default value, which is "true"
		(spec.ConsolePlugin.Enable == nil || *spec.ConsolePlugin.Enable)
}

func (spec *FlowCollectorSpec) UseTestConsolePlugin() bool {
	if spec.ConsolePlugin.Advanced != nil {
		env := spec.ConsolePlugin.Advanced.Env[constants.EnvTestConsole]
		// Use ParseBool to allow common variants ("true", "True", "1"...) and ignore non-bools
		b, err := strconv.ParseBool(env)
		return err == nil && b
	}
	return false
}

func (spec *FlowCollectorEBPF) IsAgentFeatureEnabled(feature AgentFeature) bool {
	for _, f := range spec.Features {
		if f == feature {
			return true
		}
	}
	return false
}

func (spec *FlowCollectorEBPF) IsPrivileged() bool {
	return spec.Privileged
}

func (spec *FlowCollectorEBPF) IsPktDropEnabled() bool {
	if (spec.IsPrivileged() || spec.IsEbpfManagerEnabled()) && spec.IsAgentFeatureEnabled(PacketDrop) {
		return true
	}
	return false
}

func (spec *FlowCollectorEBPF) IsDNSTrackingEnabled() bool {
	return spec.IsAgentFeatureEnabled(DNSTracking)
}

func (spec *FlowCollectorEBPF) IsFlowRTTEnabled() bool {
	return spec.IsAgentFeatureEnabled(FlowRTT)
}

func (spec *FlowCollectorEBPF) IsNetworkEventsEnabled() bool {
	return spec.IsAgentFeatureEnabled(NetworkEvents)
}

func (spec *FlowCollectorEBPF) IsPacketTranslationEnabled() bool {
	return spec.IsAgentFeatureEnabled(PacketTranslation)
}

func (spec *FlowCollectorEBPF) IsEbpfManagerEnabled() bool {
	return spec.IsAgentFeatureEnabled(EbpfManager)
}

func (spec *FlowCollectorEBPF) IsUDNMappingEnabled() bool {
	return spec.IsAgentFeatureEnabled(UDNMapping)
}

func (spec *FlowCollectorEBPF) IsIPSecEnabled() bool {
	return spec.IsAgentFeatureEnabled(IPSec)
}

func (spec *FlowCollectorEBPF) IsEBPFMetricsEnabled() bool {
	return spec.Metrics.Enable == nil || *spec.Metrics.Enable
}

func (spec *FlowCollectorEBPF) IsEBPFFlowFilterEnabled() bool {
	return spec.FlowFilter != nil && spec.FlowFilter.Enable != nil && *spec.FlowFilter.Enable
}

func (spec *FlowCollectorSpec) HasFiltersSampling() bool {
	if spec.Agent.EBPF.FlowFilter != nil {
		for i := range spec.Agent.EBPF.FlowFilter.Rules {
			if spec.Agent.EBPF.FlowFilter.Rules[i].Sampling != nil && *spec.Agent.EBPF.FlowFilter.Rules[i].Sampling > 1 {
				return true
			}
		}
	}
	for _, rule := range spec.Processor.Filters {
		if rule.Sampling > 1 {
			return true
		}
	}
	return false
}

func (spec *FlowCollectorFLP) HasConntrack() bool {
	return spec != nil && spec.LogTypes != nil && *spec.LogTypes != LogTypeFlows
}

func (spec *FlowCollectorFLP) IsMultiClusterEnabled() bool {
	return spec != nil && spec.MultiClusterDeployment != nil && *spec.MultiClusterDeployment
}

func (spec *FlowCollectorFLP) IsZoneEnabled() bool {
	return spec != nil && spec.AddZone != nil && *spec.AddZone
}

func (spec *FlowCollectorFLP) IsSubnetLabelsEnabled() bool {
	return spec.HasAutoDetectOpenShiftNetworks() || len(spec.SubnetLabels.CustomLabels) > 0
}

func (spec *FlowCollectorFLP) HasSecondaryIndexes() bool {
	return spec.Advanced != nil && len(spec.Advanced.SecondaryNetworks) > 0
}

func (spec *FlowCollectorFLP) HasAutoDetectOpenShiftNetworks() bool {
	return spec.SubnetLabels.OpenShiftAutoDetect == nil || *spec.SubnetLabels.OpenShiftAutoDetect
}

func (spec *FlowCollectorFLP) HasFLPDeduper() bool {
	return spec.Deduper != nil && spec.Deduper.Mode != "" && spec.Deduper.Mode != FLPDeduperDisabled
}

func (spec *FlowCollectorEBPF) GetMetricsPort() int32 {
	port := int32(constants.EBPFMetricPort)
	if spec.Metrics.Server.Port != nil {
		port = *spec.Metrics.Server.Port
	}
	return port
}

func (spec *FlowCollectorFLP) GetMetricsPort() int32 {
	port := int32(constants.FLPMetricsPort)
	if spec.Metrics.Server.Port != nil {
		port = *spec.Metrics.Server.Port
	}
	return port
}

func (spec *FlowCollectorSpec) HasExperimentalAlertsHealth() bool {
	if spec.Processor.Advanced != nil {
		env := spec.Processor.Advanced.Env["EXPERIMENTAL_ALERTS_HEALTH"]
		// Use ParseBool to allow common variants ("true", "True", "1"...) and ignore non-bools
		b, err := strconv.ParseBool(env)
		return err == nil && b
	}
	return false
}
