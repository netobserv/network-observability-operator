package metrics

import "github.com/netobserv/flowlogs-pipeline/pkg/config"

func (p *Preprocessed) ApplyFilters(flow config.GenericMap, flatParts []config.GenericMap) (bool, []config.GenericMap) {
	filteredParts := flatParts
	for _, filter := range p.filters {
		if filter.useFlat {
			filteredParts = filter.filterFlatParts(filteredParts)
			if len(filteredParts) == 0 {
				return false, nil
			}
		} else if !filter.predicate(flow) {
			return false, nil
		}
	}
	return true, filteredParts
}

func (pf *preprocessedFilter) filterFlatParts(flatParts []config.GenericMap) []config.GenericMap {
	var filteredParts []config.GenericMap
	for _, part := range flatParts {
		if pf.predicate(part) {
			filteredParts = append(filteredParts, part)
		}
	}
	return filteredParts
}
