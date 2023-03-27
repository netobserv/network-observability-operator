package helper

import (
	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
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

func HPADisabled(spec *flowslatest.FlowCollectorHPA) bool {
	return spec.Status == flowslatest.HPAStatusDisabled
}

func HPAEnabled(spec *flowslatest.FlowCollectorHPA) bool {
	return spec.Status == flowslatest.HPAStatusEnabled
}

func LokiNoAuthToken(spec *flowslatest.FlowCollectorLoki) bool {
	return spec.AuthToken == flowslatest.LokiAuthDisabled
}

func LokiUseHostToken(spec *flowslatest.FlowCollectorLoki) bool {
	return spec.AuthToken == flowslatest.LokiAuthUseHostToken
}

func LokiForwardUserToken(spec *flowslatest.FlowCollectorLoki) bool {
	return spec.AuthToken == flowslatest.LokiAuthForwardUserToken
}

func LokiStatusTLS(spec *flowslatest.FlowCollectorLoki) flowslatest.ClientTLS {
	if spec.StatusTLS != nil {
		return *spec.StatusTLS
	}
	return spec.TLS
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
