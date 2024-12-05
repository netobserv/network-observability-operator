/*
 * Copyright (C) 2022 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package operational

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type MetricDefinition struct {
	Name   string
	Help   string
	Type   metricType
	Labels []string
}

type metricType string

const TypeCounter metricType = "counter"
const TypeGauge metricType = "gauge"
const TypeHistogram metricType = "histogram"
const TypeSummary metricType = "summary"

var allMetrics = []MetricDefinition{}

func DefineMetric(name, help string, t metricType, labels ...string) MetricDefinition {
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
	recordsWritten = DefineMetric(
		"records_written",
		"Number of output records written",
		TypeCounter,
		"stage",
	)
	stageInQueueSize = DefineMetric(
		"stage_in_queue_size",
		"Pipeline stage input queue size (number of elements in queue)",
		TypeGauge,
		"stage",
	)
	stageOutQueueSize = DefineMetric(
		"stage_out_queue_size",
		"Pipeline stage output queue size (number of elements in queue)",
		TypeGauge,
		"stage",
	)
	stageDuration = DefineMetric(
		"stage_duration_ms",
		"Pipeline stage duration in milliseconds",
		TypeHistogram,
		"stage",
	)
	indexerHit = DefineMetric(
		"secondary_network_indexer_hit",
		"Counter of hits per secondary network index for Kubernetes enrichment",
		TypeCounter,
		"kind",
		"namespace",
		"network",
		"warning",
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
	settings           *config.MetricsSettings
	stageDurationHisto *prometheus.HistogramVec
}

func NewMetrics(settings *config.MetricsSettings) *Metrics {
	return &Metrics{settings: settings}
}

// register will register against the default registry. May panic or not depending on settings
func (o *Metrics) register(c prometheus.Collector, name string) {
	err := prometheus.DefaultRegisterer.Register(c)
	if err != nil {
		var castErr prometheus.AlreadyRegisteredError
		if errors.As(err, &castErr) {
			logrus.Warningf("metrics registration error [%s]: %v", name, err)
		} else if o.settings.NoPanic {
			logrus.Errorf("metrics registration error [%s]: %v", name, err)
		} else {
			logrus.Panicf("metrics registration error [%s]: %v", name, err)
		}
	}
}

func (o *Metrics) NewCounter(def *MetricDefinition, labels ...string) prometheus.Counter {
	verifyMetricType(def, TypeCounter)
	fullName := o.settings.Prefix + def.Name
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        fullName,
		Help:        def.Help,
		ConstLabels: def.mapLabels(labels),
	})
	o.register(c, fullName)
	return c
}

func (o *Metrics) NewCounterVec(def *MetricDefinition) *prometheus.CounterVec {
	verifyMetricType(def, TypeCounter)
	fullName := o.settings.Prefix + def.Name
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: fullName,
		Help: def.Help,
	}, def.Labels)
	o.register(c, fullName)
	return c
}

func (o *Metrics) NewGauge(def *MetricDefinition, labels ...string) prometheus.Gauge {
	verifyMetricType(def, TypeGauge)
	fullName := o.settings.Prefix + def.Name
	c := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        fullName,
		Help:        def.Help,
		ConstLabels: def.mapLabels(labels),
	})
	o.register(c, fullName)
	return c
}

func (o *Metrics) NewGaugeVec(def *MetricDefinition) *prometheus.GaugeVec {
	verifyMetricType(def, TypeGauge)
	fullName := o.settings.Prefix + def.Name
	c := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: fullName,
		Help: def.Help,
	}, def.Labels)
	o.register(c, fullName)
	return c
}

func (o *Metrics) NewGaugeFunc(def *MetricDefinition, f func() float64, labels ...string) {
	verifyMetricType(def, TypeGauge)
	fullName := o.settings.Prefix + def.Name
	c := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        fullName,
		Help:        def.Help,
		ConstLabels: def.mapLabels(labels),
	}, f)
	o.register(c, fullName)
}

func (o *Metrics) NewHistogram(def *MetricDefinition, buckets []float64, labels ...string) prometheus.Histogram {
	verifyMetricType(def, TypeHistogram)
	fullName := o.settings.Prefix + def.Name
	c := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        fullName,
		Help:        def.Help,
		Buckets:     buckets,
		ConstLabels: def.mapLabels(labels),
	})
	o.register(c, fullName)
	return c
}

func (o *Metrics) NewHistogramVec(def *MetricDefinition, buckets []float64) *prometheus.HistogramVec {
	verifyMetricType(def, TypeHistogram)
	fullName := o.settings.Prefix + def.Name
	c := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    fullName,
		Help:    def.Help,
		Buckets: buckets,
	}, def.Labels)
	o.register(c, fullName)
	return c
}

func (o *Metrics) NewSummary(def *MetricDefinition, labels ...string) prometheus.Summary {
	verifyMetricType(def, TypeSummary)
	fullName := o.settings.Prefix + def.Name
	c := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:        fullName,
		Help:        def.Help,
		ConstLabels: def.mapLabels(labels),
		// arbitrary objectives for now
		Objectives: map[float64]float64{
			0.5:  0.02,
			0.95: 0.01,
		},
	})
	o.register(c, fullName)
	return c
}

func (o *Metrics) CreateRecordsWrittenCounter(stage string) prometheus.Counter {
	return o.NewCounter(&recordsWritten, stage)
}

func (o *Metrics) CreateInQueueSizeGauge(stage string, f func() int) {
	o.NewGaugeFunc(&stageInQueueSize, func() float64 { return float64(f()) }, stage)
}

func (o *Metrics) CreateOutQueueSizeGauge(stage string, f func() int) {
	o.NewGaugeFunc(&stageOutQueueSize, func() float64 { return float64(f()) }, stage)
}

func (o *Metrics) GetOrCreateStageDurationHisto() *prometheus.HistogramVec {
	if o.stageDurationHisto == nil {
		o.stageDurationHisto = o.NewHistogramVec(&stageDuration, []float64{.001, .01, .1, 1, 10, 100, 1000, 10000})
	}
	return o.stageDurationHisto
}

func (o *Metrics) CreateIndexerHitCounter() *prometheus.CounterVec {
	return o.NewCounterVec(&indexerHit)
}

func GetDocumentation() string {
	doc := ""
	sort.Slice(allMetrics, func(i, j int) bool {
		return allMetrics[i].Name < allMetrics[j].Name
	})
	for _, opts := range allMetrics {
		doc += fmt.Sprintf(
			`
### %s
| **Name** | %s | 
|:---|:---|
| **Description** | %s | 
| **Type** | %s | 
| **Labels** | %s | 

`,
			opts.Name,
			opts.Name,
			opts.Help,
			opts.Type,
			strings.Join(opts.Labels, ", "),
		)
	}

	return doc
}
