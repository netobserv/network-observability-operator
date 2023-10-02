package helper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
)

func MetricsDefinitionToFLP(fromCRD flowslatest.MetricDefinition) (*api.PromMetricsItem, error) {
	m := &api.PromMetricsItem{
		Name:     fromCRD.Name,
		Type:     strings.ToLower(string(fromCRD.Type)),
		Filters:  []api.PromMetricsFilter{},
		Labels:   fromCRD.Labels,
		ValueKey: fromCRD.ValueField,
	}
	for k, v := range fromCRD.Filters {
		m.Filters = append(m.Filters, api.PromMetricsFilter{Key: k, Value: v})
	}
	for _, b := range fromCRD.Buckets {
		if f, err := strconv.ParseFloat(b, 64); err != nil {
			return nil, fmt.Errorf("could not parse metric buckets as floats: '%s'; error was: %w", b, err)
		} else {
			m.Buckets = append(m.Buckets, f)
		}
	}
	return m, nil
}
