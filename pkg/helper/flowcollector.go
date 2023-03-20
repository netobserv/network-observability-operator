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

func GetRecordTypes(processor *flowslatest.FlowCollectorFLP) []string {
	outputRecordTypes := []string{constants.FlowLogRecordType}
	if processor.OutputRecordTypes != nil {
		switch *processor.OutputRecordTypes {
		case flowslatest.OutputRecordFlows:
			outputRecordTypes = []string{
				constants.FlowLogRecordType,
			}
		case flowslatest.OutputRecordConnections:
			outputRecordTypes = []string{
				constants.NewConnectionRecordType,
				constants.HeartbeatRecordType,
				constants.EndConnectionRecordType,
			}
		case flowslatest.OutputRecordEndedConnections:
			outputRecordTypes = []string{
				constants.EndConnectionRecordType,
			}
		case flowslatest.OutputRecordAll:
			outputRecordTypes = []string{
				constants.FlowLogRecordType,
				constants.NewConnectionRecordType,
				constants.HeartbeatRecordType,
				constants.EndConnectionRecordType,
			}
		}
	}
	return outputRecordTypes
}
