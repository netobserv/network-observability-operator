package flp

import (
	"fmt"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	sliceslatest "github.com/netobserv/network-observability-operator/api/flowcollectorslice/v1alpha1"
)

func slicesToFilters(fc *flowslatest.FlowCollectorSpec, fcSlices []sliceslatest.FlowCollectorSlice) []api.TransformFilterRule {
	if !fc.IsSliceEnabled() {
		return nil
	}
	if fc.Processor.SlicesConfig.CollectionMode == flowslatest.CollectionAlwaysCollect {
		return nil
	}
	var rules []api.TransformFilterRule
	for _, ns := range fc.Processor.SlicesConfig.NamespacesAllowList {
		var query string
		if len(ns) >= 2 && strings.HasPrefix(ns, "/") && strings.HasSuffix(ns, "/") {
			// Regex
			pattern := strings.TrimPrefix(strings.TrimSuffix(ns, "/"), "/")
			query = fmt.Sprintf(`SrcK8S_Namespace=~"%s" or DstK8S_Namespace=~"%s"`, pattern, pattern)
		} else {
			query = fmt.Sprintf(`SrcK8S_Namespace="%s" or DstK8S_Namespace="%s"`, ns, ns)
		}
		rules = append(rules, api.TransformFilterRule{
			Type:           api.KeepEntryQuery,
			KeepEntryQuery: query,
		})
	}
	for i := range fcSlices {
		query := fmt.Sprintf(`SrcK8S_Namespace="%s" or DstK8S_Namespace="%s"`, fcSlices[i].Namespace, fcSlices[i].Namespace)
		rules = append(rules, api.TransformFilterRule{
			Type:              api.KeepEntryQuery,
			KeepEntryQuery:    query,
			KeepEntrySampling: uint16(fcSlices[i].Spec.Sampling),
		})
	}
	return rules
}

func slicesToFCSubnetLabels(fcSlices []sliceslatest.FlowCollectorSlice) []flowslatest.SubnetLabel {
	var fcLabels []flowslatest.SubnetLabel
	for i := range fcSlices {
		for _, sl := range fcSlices[i].Spec.SubnetLabels {
			fcLabels = append(fcLabels, flowslatest.SubnetLabel{
				Name:  sl.Name,
				CIDRs: sl.CIDRs,
			})
		}
	}
	return fcLabels
}
