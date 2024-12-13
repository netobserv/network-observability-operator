package metrics

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type MetricDefinition struct {
	Name   string
	Help   string
	Type   metricType
	Labels []string
}

type PromTLS struct {
	CertPath string
	KeyPath  string
}

type PromConnectionInfo struct {
	Address string
	Port    int
	TLS     *PromTLS
}

type Settings struct {
	PromConnectionInfo
	Prefix string
}

type metricType string

const (
	TypeCounter   metricType = "counter"
	TypeGauge     metricType = "gauge"
	TypeHistogram metricType = "histogram"
)

var allMetrics = []MetricDefinition{}

func defineMetric(name, help string, t metricType, labels ...string) MetricDefinition {
	def := MetricDefinition{
		Name:   name,
		Help:   help,
		Type:   t,
		Labels: labels,
	}
	allMetrics = append(allMetrics, def)
	return def
}

var (
	evictionsTotal = defineMetric(
		"evictions_total",
		"Number of eviction events",
		TypeCounter,
		"source",
		"reason",
	)
	evictedFlowsTotal = defineMetric(
		"evicted_flows_total",
		"Number of evicted flows",
		TypeCounter,
		"source",
		"reason",
	)
	evictedPktTotal = defineMetric(
		"evicted_packets_total",
		"Number of evicted packets",
		TypeCounter,
		"source",
		"reason",
	)
	lookupAndDeleteMapDurationSeconds = defineMetric(
		"lookup_and_delete_map_duration_seconds",
		"Lookup and delete map duration in seconds",
		TypeHistogram,
	)
	droppedFlows = defineMetric(
		"dropped_flows_total",
		"Number of dropped flows",
		TypeCounter,
		"source",
		"reason",
	)
	filterFlows = defineMetric(
		"filtered_flows_total",
		"Number of filtered flows",
		TypeCounter,
		"source",
		"reason",
	)
	networkEvents = defineMetric(
		"network_events_total",
		"Number of Network Events flows",
		TypeCounter,
		"source",
		"reason",
	)
	bufferSize = defineMetric(
		"buffer_size",
		"Buffer size",
		TypeGauge,
		"name",
	)
	exportedBatchCounterTotal = defineMetric(
		"exported_batch_total",
		"Exported batches",
		TypeCounter,
		"exporter",
	)
	samplingRate = defineMetric(
		"sampling_rate",
		"Sampling rate",
		TypeGauge,
	)
	errorsCounter = defineMetric(
		"errors_total",
		"errors counter",
		TypeCounter,
		"component",
		"error",
	)
)

func (def *MetricDefinition) mapLabels(labels []string) prometheus.Labels {
	if len(labels) != len(def.Labels) {
		logrus.Errorf("Could not map labels, length differ in def %s [%v / %v]", def.Name, def.Labels, labels)
	}
	labelsMap := prometheus.Labels{}
	for i, label := range labels {
		labelsMap[def.Labels[i]] = label
	}
	return labelsMap
}

func verifyMetricType(def *MetricDefinition, t metricType) {
	if def.Type != t {
		logrus.Panicf("operational metric %q is of type %q but is being registered as %q", def.Name, def.Type, t)
	}
}

type Metrics struct {
	Settings *Settings

	// Shared metrics:
	EvictionCounter       *EvictionCounter
	EvictedFlowsCounter   *EvictionCounter
	EvictedPacketsCounter *EvictionCounter
	DroppedFlowsCounter   *EvictionCounter
	FilteredFlowsCounter  *EvictionCounter
	NetworkEventsCounter  *EvictionCounter
	BufferSizeGauge       *BufferSizeGauge
	Errors                *ErrorCounter
}

func NewMetrics(settings *Settings) *Metrics {
	m := &Metrics{
		Settings: settings,
	}
	m.EvictionCounter = &EvictionCounter{vec: m.NewCounterVec(&evictionsTotal)}
	m.EvictedFlowsCounter = &EvictionCounter{vec: m.NewCounterVec(&evictedFlowsTotal)}
	m.EvictedPacketsCounter = &EvictionCounter{vec: m.NewCounterVec(&evictedPktTotal)}
	m.DroppedFlowsCounter = &EvictionCounter{vec: m.NewCounterVec(&droppedFlows)}
	m.FilteredFlowsCounter = &EvictionCounter{vec: m.NewCounterVec(&filterFlows)}
	m.NetworkEventsCounter = &EvictionCounter{vec: m.NewCounterVec(&networkEvents)}
	m.BufferSizeGauge = &BufferSizeGauge{vec: m.NewGaugeVec(&bufferSize)}
	m.Errors = &ErrorCounter{vec: m.NewCounterVec(&errorsCounter)}
	return m
}

// register will register against the default registry. May panic or not depending on settings
func (m *Metrics) register(c prometheus.Collector, name string) {
	err := prometheus.DefaultRegisterer.Register(c)
	if err != nil {
		if errors.As(err, &prometheus.AlreadyRegisteredError{}) {
			logrus.Warningf("metrics registration error [%s]: %v", name, err)
		} else {
			logrus.Panicf("metrics registration error [%s]: %v", name, err)
		}
	}
}

func (m *Metrics) NewCounter(def *MetricDefinition, labels ...string) prometheus.Counter {
	verifyMetricType(def, TypeCounter)
	fullName := m.Settings.Prefix + def.Name
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        fullName,
		Help:        def.Help,
		ConstLabels: def.mapLabels(labels),
	})
	m.register(c, fullName)
	return c
}

func (m *Metrics) NewCounterVec(def *MetricDefinition) *prometheus.CounterVec {
	verifyMetricType(def, TypeCounter)
	fullName := m.Settings.Prefix + def.Name
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: fullName,
		Help: def.Help,
	}, def.Labels)
	m.register(c, fullName)
	return c
}

func (m *Metrics) NewGauge(def *MetricDefinition, labels ...string) prometheus.Gauge {
	verifyMetricType(def, TypeGauge)
	fullName := m.Settings.Prefix + def.Name
	c := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        fullName,
		Help:        def.Help,
		ConstLabels: def.mapLabels(labels),
	})
	m.register(c, fullName)
	return c
}

func (m *Metrics) NewGaugeVec(def *MetricDefinition) *prometheus.GaugeVec {
	verifyMetricType(def, TypeGauge)
	fullName := m.Settings.Prefix + def.Name
	g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: fullName,
		Help: def.Help,
	}, def.Labels)
	m.register(g, fullName)
	return g
}

func (m *Metrics) NewHistogram(def *MetricDefinition, buckets []float64, labels ...string) prometheus.Histogram {
	verifyMetricType(def, TypeHistogram)
	fullName := m.Settings.Prefix + def.Name
	c := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        fullName,
		Help:        def.Help,
		Buckets:     buckets,
		ConstLabels: def.mapLabels(labels),
	})
	m.register(c, fullName)
	return c
}

// EvictionCounter provides syntactic sugar hidding prom's counter for eviction purpose
type EvictionCounter struct {
	vec *prometheus.CounterVec
}

func (c *EvictionCounter) WithSourceAndReason(source, reason string) prometheus.Counter {
	return c.vec.WithLabelValues(source, reason)
}

func (c *EvictionCounter) WithSource(source string) prometheus.Counter {
	return c.vec.WithLabelValues(source, "")
}

func (m *Metrics) CreateTimeSpendInLookupAndDelete() prometheus.Histogram {
	return m.NewHistogram(&lookupAndDeleteMapDurationSeconds, []float64{.001, .01, .1, 1, 10, 100, 1000, 10000})
}

// BufferSizeGauge provides syntactic sugar hidding prom's gauge tailored for buffer size
type BufferSizeGauge struct {
	vec *prometheus.GaugeVec
}

func (g *BufferSizeGauge) WithBufferName(bufferName string) prometheus.Gauge {
	return g.vec.WithLabelValues(bufferName)
}

func (m *Metrics) CreateBatchCounter(exporter string) prometheus.Counter {
	return m.NewCounter(&exportedBatchCounterTotal, exporter)
}

func (m *Metrics) CreateSamplingRate() prometheus.Gauge {
	return m.NewGauge(&samplingRate)
}

type ErrorCounter struct {
	vec *prometheus.CounterVec
}

func (c *ErrorCounter) WithErrorName(component, errName string) prometheus.Counter {
	return c.vec.WithLabelValues(component, errName)
}
