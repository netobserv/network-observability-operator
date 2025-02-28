package metrics

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
)

func (p *Preprocessed) GenerateFlatParts(flow config.GenericMap) []config.GenericMap {
	if len(p.MetricsItem.Flatten) == 0 {
		return nil
	}
	// Want to generate sub-flows from {A=foo, B=[{B1=x, B2=y},{B1=z}], C=[foo,bar]}
	// => {B>B1=x, B>B2=y, C=foo}, {B>B1=z, C=foo}, {B>B1=x, B>B2=y, C=bar}, {B>B1=z, C=bar}
	var partsPerLabel [][]config.GenericMap
	for _, fl := range p.MetricsItem.Flatten {
		if anyVal, ok := flow[fl]; ok {
			// Intermediate step to get:
			// [{B>B1=x, B>B2=y}, {B>B1=z}], [C=foo, C=bar]
			var partsForLabel []config.GenericMap
			switch v := anyVal.(type) {
			case []any:
				prefix := fl + ">"
				for _, vv := range v {
					switch vvv := vv.(type) {
					case map[string]string:
						partsForLabel = append(partsForLabel, flattenNested(prefix, vvv))
					default:
						partsForLabel = append(partsForLabel, config.GenericMap{fl: vv})
					}
				}
			case []map[string]string:
				prefix := fl + ">"
				for _, vv := range v {
					partsForLabel = append(partsForLabel, flattenNested(prefix, vv))
				}
			case []string:
				for _, vv := range v {
					partsForLabel = append(partsForLabel, config.GenericMap{fl: vv})
				}
			}
			if len(partsForLabel) > 0 {
				partsPerLabel = append(partsPerLabel, partsForLabel)
			}
		}
	}
	return distribute(partsPerLabel)
}

func distribute(allUnflat [][]config.GenericMap) []config.GenericMap {
	// turn
	// [{B>B1=x, B>B2=y}, {B>B1=z}], [{C=foo}, {C=bar}]
	// into
	// [{B>B1=x, B>B2=y, C=foo}, {B>B1=z, C=foo}, {B>B1=x, B>B2=y, C=bar}, {B>B1=z, C=bar}]
	totalCard := 1
	for _, part := range allUnflat {
		if len(part) > 1 {
			totalCard *= len(part)
		}
	}
	ret := make([]config.GenericMap, totalCard)
	indexes := make([]int, len(allUnflat))
	for c := range ret {
		ret[c] = config.GenericMap{}
		incIndex := false
		for i, part := range allUnflat {
			index := indexes[i]
			for k, v := range part[index] {
				ret[c][k] = v
			}
			if !incIndex {
				if index+1 == len(part) {
					indexes[i] = 0
				} else {
					indexes[i] = index + 1
					incIndex = true
				}
			}
		}
	}
	return ret
}

func flattenNested(prefix string, nested map[string]string) config.GenericMap {
	subFlow := config.GenericMap{}
	for k, v := range nested {
		subFlow[prefix+k] = v
	}
	return subFlow
}
