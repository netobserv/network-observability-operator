package flp

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
)

func getConfiguredMetrics(cm *corev1.ConfigMap) (api.MetricsItems, error) {
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

func defaultBuilderWithMetrics(metrics *metricslatest.FlowMetricList) (*Builder, error) {
	cfg := getConfig()
	loki := helper.NewLokiConfig(&cfg.Loki, "any")
	info := reconcilers.Common{Namespace: "namespace", Loki: &loki}
	return newInProcessBuilder(info.NewInstance(image, status.Instance{}), constants.FLPName, &cfg, metrics)
}

func metric(metrics api.MetricsItems, name string) *api.MetricsItem {
	for i := range metrics {
		if metrics[i].Name == name {
			return &metrics[i]
		}
	}
	return nil
}

func TestFlowMetricToFLP(t *testing.T) {
	assert := assert.New(t)

	b, err := defaultBuilderWithMetrics(&metricslatest.FlowMetricList{
		Items: []metricslatest.FlowMetric{
			{Spec: metricslatest.FlowMetricSpec{
				MetricName: "m_1",
				Type:       metricslatest.CounterMetric,
				ValueField: "val",
				Labels:     []string{"by_field"},
				Filters:    []metricslatest.MetricFilter{{Field: "f", Value: "v", MatchType: metricslatest.MatchEqual}},
			}},
			{Spec: metricslatest.FlowMetricSpec{
				MetricName:        "m_2",
				Type:              metricslatest.HistogramMetric,
				Labels:            []string{"by_field"},
				IncludeDuplicates: true,
				Direction:         metricslatest.Egress,
				Filters: []metricslatest.MetricFilter{
					{Field: "f", Value: "v", MatchType: metricslatest.MatchRegex},
					{Field: "f2", MatchType: metricslatest.MatchAbsence},
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
	assert.Equal(api.MetricsItem{
		Name: "m_1",
		Type: "counter",
		Filters: []api.MetricsFilter{
			{Key: "f", Value: "v", Type: api.PromFilterEqual},
			{Key: "Duplicate", Value: "true", Type: api.PromFilterNotEqual},
		},
		ValueKey: "val",
		Labels:   []string{"by_field"},
	}, *m1)
	m2 := metric(items, "m_2")
	assert.Equal(api.MetricsItem{
		Name: "m_2",
		Type: "histogram",
		Filters: []api.MetricsFilter{
			{Key: "f", Value: "v", Type: api.PromFilterRegex},
			{Key: "f2", Type: api.PromFilterAbsence},
			{Key: "FlowDirection", Value: "1|2", Type: api.PromFilterRegex},
		},
		Labels:  []string{"by_field"},
		Buckets: []float64{1, 5, 10, 50, 100},
	}, *m2)
}
