package metrics

import (
	"regexp"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/utils/filters"
)

type Preprocessed struct {
	*api.MetricsItem
	filters         []preprocessedFilter
	MappedLabels    []MappedLabel
	FlattenedLabels []MappedLabel
}

type MappedLabel struct {
	Source string
	Target string
}

type preprocessedFilter struct {
	predicate filters.Predicate
	useFlat   bool
}

func (p *Preprocessed) TargetLabels() []string {
	var targetLabels []string
	for _, l := range p.FlattenedLabels {
		targetLabels = append(targetLabels, l.Target)
	}
	for _, l := range p.MappedLabels {
		targetLabels = append(targetLabels, l.Target)
	}
	return targetLabels
}

func filterToPredicate(filter api.MetricsFilter) filters.Predicate {
	switch filter.Type {
	case api.MetricFilterEqual:
		return filters.Equal(filter.Key, filter.Value, true)
	case api.MetricFilterNotEqual:
		return filters.NotEqual(filter.Key, filter.Value, true)
	case api.MetricFilterPresence:
		return filters.Presence(filter.Key)
	case api.MetricFilterAbsence:
		return filters.Absence(filter.Key)
	case api.MetricFilterRegex:
		r, _ := regexp.Compile(filter.Value)
		return filters.Regex(filter.Key, r)
	case api.MetricFilterNotRegex:
		r, _ := regexp.Compile(filter.Value)
		return filters.NotRegex(filter.Key, r)
	}
	// Default = Exact
	return filters.Equal(filter.Key, filter.Value, true)
}

func Preprocess(def *api.MetricsItem) *Preprocessed {
	mi := Preprocessed{
		MetricsItem: def,
	}
	for _, l := range def.Labels {
		ml := MappedLabel{Source: l, Target: l}
		if as := def.Remap[l]; as != "" {
			ml.Target = as
		}
		if mi.isFlattened(l) {
			mi.FlattenedLabels = append(mi.FlattenedLabels, ml)
		} else {
			mi.MappedLabels = append(mi.MappedLabels, ml)
		}
	}
	for _, f := range def.Filters {
		mi.filters = append(mi.filters, preprocessedFilter{
			predicate: filterToPredicate(f),
			useFlat:   mi.isFlattened(f.Key),
		})
	}
	return &mi
}

func (p *Preprocessed) isFlattened(fieldPath string) bool {
	for _, flat := range p.Flatten {
		if fieldPath == flat || strings.HasPrefix(fieldPath, flat+">") {
			return true
		}
	}
	return false
}
