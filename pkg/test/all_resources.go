package test

import (
	"github.com/netobserv/network-observability-operator/controllers/constants"
)

const (
	TestNamespace = "test-ns"
)

var (
	// All managed resources
	AgentDS            = DaemonSet(constants.EBPFAgentName)
	AgentSA            = ServiceAccount(constants.EBPFServiceAccount)
	AgentMetricsSvc    = Service(constants.EBPFAgentMetricsSvcName)
	AgentFLPMetricsSvc = Service("netobserv-ebpf-agent-prom")
	AgentSM            = ServiceMonitor(constants.EBPFAgentMetricsSvcMonitoringName)
	AgentFLPSM         = ServiceMonitor(constants.EBPFAgentName + "-monitor")
	AgentFLPRule       = PrometheusRule(constants.EBPFAgentName + "-alert")
	AgentFLPCRB        = ClusterRoleBinding(constants.EBPFAgentName)
	AgentNS            = Namespace(TestNamespace + "-privileged")
	FLPDepl            = Deployment(constants.FLPName)
	FLPCM              = ConfigMap(constants.FLPName + "-config")
	FLPSA              = ServiceAccount(constants.FLPName)
	FLPCRB             = ClusterRoleBinding(constants.FLPName)
	FLPHPA             = HPA(constants.FLPName)
	FLPMetricsSvc      = Service(constants.FLPName + "-prom")
	FLPSM              = ServiceMonitor(constants.FLPName + "-monitor")
	FLPRule            = PrometheusRule(constants.FLPName + "-alert")
	PluginDepl         = Deployment(constants.PluginName)
	PluginCM           = ConfigMap("console-plugin-config")
	PluginSvc          = Service(constants.PluginName)
	PluginSA           = ServiceAccount(constants.PluginName)
	PluginCRB          = ClusterRoleBinding(constants.PluginName)
	PluginSM           = ServiceMonitor(constants.PluginName)
	LokiWriterCR       = ClusterRole(constants.LokiCRWriter)
	LokiReaderCR       = ClusterRole(constants.LokiCRReader)
	LokiWriterCRB      = ClusterRoleBinding(constants.LokiCRBWriter)

	// Old resources
	FLPDS = DaemonSet(constants.FLPName)

	ClusterResources = []ResourceRef{
		FLPCRB, AgentFLPCRB, AgentNS, PluginCRB,
		LokiWriterCR, LokiReaderCR, LokiWriterCRB,
	}
	FLPResources = []ResourceRef{
		FLPDepl, FLPDS, FLPCM, FLPSA, FLPHPA, FLPMetricsSvc, FLPSM, FLPRule,
	}
	PluginResources = []ResourceRef{
		PluginDepl, PluginCM, PluginSvc, PluginSA, PluginSM,
	}
	AgentResources = []ResourceRef{
		AgentDS, AgentSA, AgentMetricsSvc, AgentFLPMetricsSvc, AgentSM, AgentFLPSM, AgentFLPRule,
	}
)

func GetClusterResourcesIn(used []ResourceRef) []ResourceRef {
	return filter(used, ClusterResources)
}

func GetAgentResourcesIn(used []ResourceRef) []ResourceRef {
	return filter(used, AgentResources)
}

func GetFLPResourcesIn(used []ResourceRef) []ResourceRef {
	return filter(used, FLPResources)
}

func GetPluginResourcesIn(used []ResourceRef) []ResourceRef {
	return filter(used, PluginResources)
}

func GetAgentResourcesNotIn(used []ResourceRef) []ResourceRef {
	return getComplement(used, AgentResources)
}

func GetFLPResourcesNotIn(used []ResourceRef) []ResourceRef {
	return getComplement(used, FLPResources)
}

func GetPluginResourcesNotIn(used []ResourceRef) []ResourceRef {
	return getComplement(used, PluginResources)
}

func filter(used []ResourceRef, among []ResourceRef) []ResourceRef {
	var ret []ResourceRef
	for _, r := range among {
		if hasResource(r, used) {
			ret = append(ret, r)
		}
	}
	return ret
}

func getComplement(used []ResourceRef, among []ResourceRef) []ResourceRef {
	var compl []ResourceRef
	for _, r := range among {
		if !hasResource(r, used) {
			compl = append(compl, r)
		}
	}
	return compl
}

func hasResource(toCheck ResourceRef, list []ResourceRef) bool {
	for _, rUsed := range list {
		if rUsed.kind == toCheck.kind && rUsed.name == toCheck.name {
			return true
		}
	}
	return false
}
