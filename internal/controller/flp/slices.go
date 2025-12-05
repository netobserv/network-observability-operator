package flp

import (
	"fmt"
	"net"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	sliceslatest "github.com/netobserv/network-observability-operator/api/flowcollectorslice/v1alpha1"
	"github.com/netobserv/network-observability-operator/internal/controller/flp/slicesstatus"
)

func slicesToFilters(fc *flowslatest.FlowCollectorSpec, fcSlices []sliceslatest.FlowCollectorSlice) []api.TransformFilterRule {
	if !fc.IsSliceEnabled() {
		return nil
	}
	if fc.Processor.SlicesConfig.CollectionMode == flowslatest.CollectionAlwaysCollect {
		return nil
	}
	processed := make(map[string]any)
	var rules []api.TransformFilterRule
	// First, process admin config
	for _, ns := range fc.Processor.SlicesConfig.NamespacesAllowList {
		if len(ns) >= 2 && strings.HasPrefix(ns, "/") && strings.HasSuffix(ns, "/") {
			// Strings enclosed between '/' are considered as regexes
			pattern := strings.TrimPrefix(strings.TrimSuffix(ns, "/"), "/")
			rules = append(rules, api.TransformFilterRule{
				Type:           api.KeepEntryQuery,
				KeepEntryQuery: fmt.Sprintf(`SrcK8S_Namespace=~"%s" or DstK8S_Namespace=~"%s"`, pattern, pattern),
			})
		} else if _, found := processed[ns]; !found {
			rules = append(rules, api.TransformFilterRule{
				Type:           api.KeepEntryQuery,
				KeepEntryQuery: fmt.Sprintf(`SrcK8S_Namespace="%s" or DstK8S_Namespace="%s"`, ns, ns),
			})
			processed[ns] = nil
		}
	}
	// Then add slices config
	for i := range fcSlices {
		if _, found := processed[fcSlices[i].Namespace]; !found {
			q := fmt.Sprintf(`SrcK8S_Namespace="%s" or DstK8S_Namespace="%s"`, fcSlices[i].Namespace, fcSlices[i].Namespace)
			rules = append(rules, api.TransformFilterRule{
				Type:              api.KeepEntryQuery,
				KeepEntryQuery:    q,
				KeepEntrySampling: uint16(fcSlices[i].Spec.Sampling),
			})
			processed[fcSlices[i].Namespace] = nil
			fcSlices[i].Status.FilterApplied = q
		} else {
			fcSlices[i].Status.FilterApplied = "(skipped, not needed)"
		}
	}
	return rules
}

func slicesToFCSubnetLabels(fcSlices []sliceslatest.FlowCollectorSlice, configuredCIDRs []*net.IPNet) []flowslatest.SubnetLabel {
	var fcLabels []flowslatest.SubnetLabel
	for i := range fcSlices {
		var hasError bool
		var countConfigured int
		for _, sl := range fcSlices[i].Spec.SubnetLabels {
			var cidrs []string
			for _, strCIDR := range sl.CIDRs {
				// Check for parse error
				if _, cidr, err := net.ParseCIDR(strCIDR); err != nil {
					hasError = true
					slicesstatus.SetFailure(&fcSlices[i], fmt.Sprintf("Wrong CIDR for subnet label '%s': %v", sl.Name, err))
				} else {
					var skip bool
					// Check for overlap with higher priority CIDRs
					for _, other := range configuredCIDRs {
						if other.Contains(cidr.IP) {
							thisMaskSize, _ := cidr.Mask.Size()
							otherMaskSize, _ := other.Mask.Size()
							if thisMaskSize >= otherMaskSize {
								// E.g: admin defined 10.100.0.0/16 and slice defined 10.100.10.0/24
								// => fully included, warn and skip adding for FLP
								slicesstatus.AddSubnetWarning(&fcSlices[i], fmt.Sprintf("CIDR for '%s' (%v) is fully included in admin config (%v) and will be ignored", sl.Name, cidr, other))
								skip = true
							} else {
								// E.g: admin defined 10.100.0.0/17 and slice defined 10.100.0.0/16
								// => slice includes admin config, warn but add to FLP
								slicesstatus.AddSubnetWarning(&fcSlices[i], fmt.Sprintf("CIDR for '%s' (%v) overlaps with admin config (%v)", sl.Name, cidr, other))
							}
						} else if cidr.Contains(other.IP) {
							// E.g: admin defined 10.100.10.0/24 and slice defined 10.100.0.0/16
							// => slice includes admin config, warn but add to FLP
							slicesstatus.AddSubnetWarning(&fcSlices[i], fmt.Sprintf("CIDR for '%s' (%v) overlaps with admin config (%v)", sl.Name, cidr, other))
						}
					}
					if !skip {
						cidrs = append(cidrs, strCIDR)
					}
				}
			}
			if len(cidrs) > 0 {
				fcLabels = append(fcLabels, flowslatest.SubnetLabel{
					Name:  sl.Name,
					CIDRs: cidrs,
				})
				countConfigured++
			}
		}
		fcSlices[i].Status.SubnetLabelsConfigured = countConfigured
		if !hasError {
			slicesstatus.SetReady(&fcSlices[i])
		}
	}
	return fcLabels
}
