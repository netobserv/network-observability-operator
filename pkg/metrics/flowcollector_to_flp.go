package metrics

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
)

func ToFLP(fromCRD *flowslatest.MetricDefinition) (*api.PromMetricsItem, error) {
	m := &api.PromMetricsItem{
		Name:     fromCRD.Name,
		Type:     strings.ToLower(string(fromCRD.Type)),
		Filters:  []api.PromMetricsFilter{},
		Labels:   fromCRD.Labels,
		ValueKey: fromCRD.ValueField,
	}
	for _, f := range fromCRD.Filters {
		m.Filters = append(m.Filters, api.PromMetricsFilter{Key: f.Field, Value: f.Value})
	}
	for _, b := range fromCRD.Buckets {
		f, err := strconv.ParseFloat(b, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse metric buckets as floats: '%s'; error was: %w", b, err)
		}
		m.Buckets = append(m.Buckets, f)
	}
	return m, nil
}
