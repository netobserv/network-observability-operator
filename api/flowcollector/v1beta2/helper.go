package v1beta2

import (
	"strconv"
	"time"

	"github.com/netobserv/netobserv-operator/internal/controller/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (spec *FlowCollectorSpec) GetNamespace() string {
	if spec.Namespace != "" {
		return spec.Namespace
	}
	return constants.DefaultOperatorNamespace
}

func (spec *FlowCollectorSpec) OnHold() bool {
	return spec.Execution.Mode == OnHold
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

func (spec *FlowCollectorSpec) HasS3Exporter() bool {
	return spec.GetFirstS3Exporter() != nil
}

// GetFirstS3Exporter returns the first exporters entry with type S3, or nil.
func (spec *FlowCollectorSpec) GetFirstS3Exporter() *FlowCollectorExporter {
	for _, ex := range spec.Exporters {
		if ex != nil && ex.Type == S3Exporter {
			return ex
		}
	}
	return nil
}

func (cfg *SASLConfig) UseSASL() bool {
	return cfg.Type == SASLPlain || cfg.Type == SASLScramSHA512
}

func (spec *FlowCollectorSpec) UseLoki() bool {
	// nil should fallback to default value, which is "true"
	return spec.Loki.Enable == nil || *spec.Loki.Enable
}

func (spec *FlowCollectorSpec) UseLokiDev() bool {
	return spec.UseLoki() && spec.Loki.Mode == LokiModeMonolithic && spec.Loki.Monolithic.InstallDemoLoki != nil && *spec.Loki.Monolithic.InstallDemoLoki
}

func (spec *FlowCollectorSpec) UsePrometheus() bool {
	// nil should fallback to default value, which is "true"
	return spec.Prometheus.Querier.Enable == nil || *spec.Prometheus.Querier.Enable
}

// UseFlowBuffer reports whether the in-memory flow buffer should be active.
// Explicit spec.processor.flowBuffer.enable overrides the Loki-based default
// (on when Loki is off, off when Loki is on).
func (spec *FlowCollectorSpec) UseFlowBuffer() bool {
	if spec.Processor.FlowBuffer != nil && spec.Processor.FlowBuffer.Enable != nil {
		return *spec.Processor.FlowBuffer.Enable
	}
	return !spec.UseLoki()
}

// UseS3 reports whether the Console plugin should query S3 Parquet.
// Explicit spec.consolePlugin.s3.enable overrides the default
// (on when an S3 exporter exists and Loki is off).
func (spec *FlowCollectorSpec) UseS3() bool {
	if spec.ConsolePlugin.S3 != nil && spec.ConsolePlugin.S3.Enable != nil {
		return *spec.ConsolePlugin.S3.Enable
	}
	return spec.HasS3Exporter() && !spec.UseLoki()
}

func (spec *FlowCollectorSpec) UseWebConsole() bool {
	return (spec.UseLoki() || spec.UsePrometheus() || spec.UseFlowBuffer() || spec.UseS3()) &&
		// nil should fallback to default value, which is "true"
		(spec.ConsolePlugin.Enable == nil || *spec.ConsolePlugin.Enable)
}

// GetFLPServiceName returns the Kubernetes Service name used to reach FLP (ingest and/or flow-buffer query).
func (spec *FlowCollectorSpec) GetFLPServiceName() string {
	if spec.UseKafka() {
		return constants.FLPTransfoName
	}
	return constants.FLPName
}

// GetFlowBufferMaxEntries returns the configured max entries, or the default.
func (spec *FlowCollectorSpec) GetFlowBufferMaxEntries() int32 {
	if spec.Processor.FlowBuffer != nil && spec.Processor.FlowBuffer.MaxEntries != nil {
		return *spec.Processor.FlowBuffer.MaxEntries
	}
	return 50000
}

// GetFlowBufferQueryTimeout returns the configured query timeout, or the default.
func (spec *FlowCollectorSpec) GetFlowBufferQueryTimeout() metav1.Duration {
	if spec.Processor.FlowBuffer != nil && spec.Processor.FlowBuffer.QueryTimeout != nil {
		return *spec.Processor.FlowBuffer.QueryTimeout
	}
	return metav1.Duration{Duration: 2 * time.Second}
}

func (spec *FlowCollectorSpec) UseStandaloneConsole(hasPluginAPI bool) bool {
	// defaults to true if there's no plugin API, false otherwise
	return (spec.ConsolePlugin.Standalone != nil && *spec.ConsolePlugin.Standalone ||
		spec.ConsolePlugin.Standalone == nil && !hasPluginAPI)
}

// NeedsConsolePluginDeployment is true when the console plugin Deployment (and related objects)
// should be reconciled. When false, the console reconciler only tears down or marks unused and
// does not need a resolved plugin image. hasPluginAPI is whether the ConsolePlugin CRD exists.
func (spec *FlowCollectorSpec) NeedsConsolePluginDeployment(hasPluginAPI bool) bool {
	return spec.UseWebConsole() &&
		(hasPluginAPI || spec.UseStandaloneConsole(hasPluginAPI)) &&
		!spec.OnHold()
}

func (spec *FlowCollectorSpec) UseHostNetwork() bool {
	return spec.DeploymentModel == DeploymentModelDirect
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

func (spec *FlowCollectorEBPF) IsTLSTrackingEnabled() bool {
	return spec.IsAgentFeatureEnabled(TLSTracking)
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

func (spec *FlowCollectorSpec) GetSecondaryIndexes() []SecondaryNetwork {
	if spec.Processor.Advanced != nil && len(spec.Processor.Advanced.SecondaryNetworks) > 0 {
		return spec.Processor.Advanced.SecondaryNetworks
	}
	if spec.Agent.EBPF.Privileged {
		// Turn-on auto-detection in FLP by interface+MAC or interface+IP
		return []SecondaryNetwork{
			{Index: []SecondaryNetworkIndex{SecondaryNetworkIndexByInterface, SecondaryNetworkIndexByIP}},
			{Index: []SecondaryNetworkIndex{SecondaryNetworkIndexByInterface, SecondaryNetworkIndexByMAC}},
		}
	}
	return nil
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

func (spec *FlowCollectorSpec) DeployNetworkPolicy(trueByDefault bool) bool {
	if trueByDefault {
		return spec.NetworkPolicy.Enable == nil || *spec.NetworkPolicy.Enable
	}
	return spec.NetworkPolicy.Enable != nil && *spec.NetworkPolicy.Enable
}

func (spec *FlowCollectorFLP) GetFLPReplicas() int32 {
	if spec.ConsumerReplicas != nil {
		return *spec.ConsumerReplicas
	} else if spec.KafkaConsumerReplicas != nil {
		return *spec.KafkaConsumerReplicas
	}
	return 3
}

func (spec *FlowCollectorHPA) IsHPAEnabled() bool {
	return spec != nil && spec.Status == HPAStatusEnabled
}

func (spec *FlowCollectorFLP) IsUnmanagedFLPReplicas() bool {
	if spec.UnmanagedReplicas {
		return true
	}
	return spec.KafkaConsumerAutoscaler.IsHPAEnabled()
}

func (spec *FlowCollectorFLP) IsInformerCacheProxyEnabled() bool {
	return spec.InformerCacheProxy != nil && spec.InformerCacheProxy.Enabled != nil && *spec.InformerCacheProxy.Enabled
}

// DefaultK8sCachePort is the default gRPC port where FLP processors listen for k8scache updates
const DefaultK8sCachePort int32 = 9402

// GetK8sCachePort returns the gRPC port where FLP processors listen for k8scache updates.
// If spec.processor.informerCacheProxy.advanced.processorPort is configured, it returns that value.
// Otherwise, it returns the default port (9402).
func (spec *FlowCollectorFLP) GetK8sCachePort() int32 {
	if spec.InformerCacheProxy != nil &&
		spec.InformerCacheProxy.Advanced != nil &&
		spec.InformerCacheProxy.Advanced.ProcessorPort != nil {
		return *spec.InformerCacheProxy.Advanced.ProcessorPort
	}
	return DefaultK8sCachePort
}

func (spec *FlowCollectorInformerCacheProxy) GetTLSType() TLSConfigType {
	if spec == nil || spec.TLS == nil {
		return TLSAuto
	}
	return spec.TLS.Type
}

func (spec *FlowCollectorInformerCacheProxy) UsesOpenShiftServiceCA(isOpenShift bool) bool {
	if !isOpenShift {
		return false
	}
	tlsType := spec.GetTLSType()
	return tlsType == TLSAuto
}

func (spec *FlowCollectorConsolePlugin) IsUnmanagedConsolePluginReplicas() bool {
	if spec.UnmanagedReplicas {
		return true
	}
	return spec.Autoscaler.IsHPAEnabled()
}

func (spec *FlowCollectorSpec) IsSliceEnabled() bool {
	return spec.Processor.SlicesConfig != nil && spec.Processor.SlicesConfig.Enable
}

func IsEnvEnabled(vars map[string]string, key string) bool {
	env := vars[key]
	// Use ParseBool to allow common variants ("true", "True", "1"...) and ignore non-bools
	b, err := strconv.ParseBool(env)
	return err == nil && b
}
