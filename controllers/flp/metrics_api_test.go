package flp

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
)

func getConfiguredMetrics(cm *corev1.ConfigMap) (api.PromMetricsItems, error) {
	var cfs config.ConfigFileStruct
	err := json.Unmarshal([]byte(cm.Data[configFile]), &cfs)
	if err != nil {
		return nil, err
	}
	for _, stage := range cfs.Parameters {
		if stage.Encode != nil && stage.Encode.Type == "prom" {
			return stage.Encode.Prom.Metrics, nil
		}
	}
	return nil, errors.New("prom encode stage not found")
}

func defaultBuilderWithMetrics(metrics *v1alpha1.FlowMetricList) (monolithBuilder, error) {
	cfg := getConfig()
	loki := helper.NewLokiConfig(&cfg.Loki, "any")
	info := reconcilers.Common{Namespace: "namespace", Loki: &loki}
	return newMonolithBuilder(info.NewInstance(image, status.Instance{}), &cfg, metrics)
}

func metric(metrics api.PromMetricsItems, name string) *api.PromMetricsItem {
	for i := range metrics {
		if metrics[i].Name == name {
			return &metrics[i]
		}
	}
	return nil
}

func TestFlowMetricToFLP(t *testing.T) {
	assert := assert.New(t)

	b, err := defaultBuilderWithMetrics(&v1alpha1.FlowMetricList{
		Items: []v1alpha1.FlowMetric{
			{Spec: v1alpha1.FlowMetricSpec{
				MetricName: "m_1",
				Type:       v1alpha1.CounterMetric,
				ValueField: "val",
				Labels:     []string{"by_field"},
				Filters:    []v1alpha1.MetricFilter{{Field: "f", Value: "v", MatchType: v1alpha1.MatchExact}},
			}},
			{Spec: v1alpha1.FlowMetricSpec{
				MetricName:        "m_2",
				Type:              v1alpha1.HistogramMetric,
				Labels:            []string{"by_field"},
				IncludeDuplicates: true,
				Direction:         v1alpha1.Egress,
				Filters: []v1alpha1.MetricFilter{
					{Field: "f", Value: "v", MatchType: v1alpha1.MatchRegex},
					{Field: "f2", MatchType: v1alpha1.MatchAbsence},
				},
				Buckets: []string{"1", "5", "10", "50", "100"},
			}},
		},
	})
	assert.NoError(err)
	cm, _, err := b.configMap()
	assert.NoError(err)
	items, err := getConfiguredMetrics(cm)
	assert.NoError(err)
	names := getSortedMetricsNames(items)
	assert.Equal([]string{
		"m_1",
		"m_2",
		"namespace_flows_total",
		"node_ingress_bytes_total",
		"workload_ingress_bytes_total",
	}, names)

	m1 := metric(items, "m_1")
	assert.Equal(api.PromMetricsItem{
		Name:   "m_1",
		Type:   "counter",
		Filter: api.PromMetricsFilter{Key: "", Value: "", Type: ""},
		Filters: []api.PromMetricsFilter{
			{Key: "f", Value: "v", Type: api.PromFilterExact},
			{Key: "Duplicate", Value: "false", Type: api.PromFilterExact},
		},
		ValueKey: "val",
		Labels:   []string{"by_field"},
	}, *m1)
	m2 := metric(items, "m_2")
	assert.Equal(api.PromMetricsItem{
		Name:   "m_2",
		Type:   "histogram",
		Filter: api.PromMetricsFilter{Key: "", Value: "", Type: ""},
		Filters: []api.PromMetricsFilter{
			{Key: "f", Value: "v", Type: api.PromFilterRegex},
			{Key: "f2", Type: api.PromFilterAbsence},
			{Key: "FlowDirection", Value: "1|2", Type: api.PromFilterRegex},
		},
		Labels:  []string{"by_field"},
		Buckets: []float64{1, 5, 10, 50, 100},
	}, *m2)
}
