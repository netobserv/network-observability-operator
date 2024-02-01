package loki

import (
	"strconv"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/ebpf"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

func GetLokiLabels(desired *flowslatest.FlowCollectorSpec) []string {
	indexFields := constants.LokiIndexFields

	if desired.Processor.LogTypes != nil && *desired.Processor.LogTypes != flowslatest.LogTypeFlows {
		indexFields = append(indexFields, constants.LokiConnectionIndexFields...)
	}

	if helper.IsMultiClusterEnabled(&desired.Processor) {
		indexFields = append(indexFields, constants.ClusterNameLabelName)
	}

	if helper.IsZoneEnabled(&desired.Processor) {
		indexFields = append(indexFields, constants.LokiZoneIndexFields...)
	}

	if helper.UseEBPF(desired) {
		dedupJustMark, _ := strconv.ParseBool(ebpf.DedupeJustMarkDefault)
		if desired.Agent.EBPF.Advanced != nil {
			if v, ok := desired.Agent.EBPF.Advanced.Env[ebpf.EnvDedupeJustMark]; ok {
				dedupJustMark, _ = strconv.ParseBool(v)
			}
		}
		if dedupJustMark {
			indexFields = append(indexFields, constants.LokiDeduperMarkIndexFields...)
		}
	}

	return indexFields
}
