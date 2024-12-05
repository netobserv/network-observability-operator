/*
 * Copyright (C) 2021 IBM, Inc.
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

package encode

import (
	"reflect"
	"strings"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/encode/metrics"
	promserver "github.com/netobserv/flowlogs-pipeline/pkg/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var plog = logrus.WithField("component", "encode.Prometheus")

const defaultExpiryTime = time.Duration(2 * time.Minute)

// nolint:revive
type EncodeProm struct {
	cfg          *api.PromEncode
	registerer   prometheus.Registerer
	metricCommon *MetricsCommonStruct
	updateChan   chan config.StageParam
	server       *promserver.PromServer
	regName      string
}

func (e *EncodeProm) Gatherer() prometheus.Gatherer {
	return e.server
}

// Encode encodes a metric before being stored; the heavy work is done by the MetricCommonEncode
func (e *EncodeProm) Encode(metricRecord config.GenericMap) {
	plog.Tracef("entering EncodeMetric. metricRecord = %v", metricRecord)
	e.metricCommon.MetricCommonEncode(e, metricRecord)
	e.checkConfUpdate()
}

func (e *EncodeProm) ProcessCounter(m interface{}, labels map[string]string, value float64) error {
	counter := m.(*prometheus.CounterVec)
	mm, err := counter.GetMetricWith(labels)
	if err != nil {
		return err
	}
	mm.Add(value)
	return nil
}

func (e *EncodeProm) ProcessGauge(m interface{}, labels map[string]string, value float64, _ string) error {
	gauge := m.(*prometheus.GaugeVec)
	mm, err := gauge.GetMetricWith(labels)
	if err != nil {
		return err
	}
	mm.Set(value)
	return nil
}

func (e *EncodeProm) ProcessHist(m interface{}, labels map[string]string, value float64) error {
	hist := m.(*prometheus.HistogramVec)
	mm, err := hist.GetMetricWith(labels)
	if err != nil {
		return err
	}
	mm.Observe(value)
	return nil
}

func (e *EncodeProm) ProcessAggHist(m interface{}, labels map[string]string, values []float64) error {
	hist := m.(*prometheus.HistogramVec)
	mm, err := hist.GetMetricWith(labels)
	if err != nil {
		return err
	}
	for _, v := range values {
		mm.Observe(v)
	}
	return nil
}

func (e *EncodeProm) GetChacheEntry(entryLabels map[string]string, m interface{}) interface{} {
	switch mv := m.(type) {
	case *prometheus.CounterVec:
		return func() { mv.Delete(entryLabels) }
	case *prometheus.GaugeVec:
		return func() { mv.Delete(entryLabels) }
	case *prometheus.HistogramVec:
		return func() { mv.Delete(entryLabels) }
	}
	return nil
}

// callback function from lru cleanup
func (e *EncodeProm) Cleanup(cleanupFunc interface{}) {
	cleanupFunc.(func())()
}

func (e *EncodeProm) addCounter(fullMetricName string, mInfo *metrics.Preprocessed) prometheus.Collector {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{Name: fullMetricName, Help: ""}, mInfo.TargetLabels())
	e.metricCommon.AddCounter(fullMetricName, counter, mInfo)
	return counter
}

func (e *EncodeProm) addGauge(fullMetricName string, mInfo *metrics.Preprocessed) prometheus.Collector {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: fullMetricName, Help: ""}, mInfo.TargetLabels())
	e.metricCommon.AddGauge(fullMetricName, gauge, mInfo)
	return gauge
}

func (e *EncodeProm) addHistogram(fullMetricName string, mInfo *metrics.Preprocessed) prometheus.Collector {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: fullMetricName, Help: ""}, mInfo.TargetLabels())
	e.metricCommon.AddHist(fullMetricName, histogram, mInfo)
	return histogram
}

func (e *EncodeProm) addAgghistogram(fullMetricName string, mInfo *metrics.Preprocessed) prometheus.Collector {
	agghistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: fullMetricName, Help: ""}, mInfo.TargetLabels())
	e.metricCommon.AddAggHist(fullMetricName, agghistogram, mInfo)
	return agghistogram
}

func (e *EncodeProm) unregisterMetric(c interface{}) {
	if c, ok := c.(prometheus.Collector); ok {
		e.registerer.Unregister(c)
	}
}

func (e *EncodeProm) cleanDeletedGeneric(newCfg api.PromEncode, metrics map[string]mInfoStruct) {
	for fullName, m := range metrics {
		if !strings.HasPrefix(fullName, newCfg.Prefix) {
			if c, ok := m.genericMetric.(prometheus.Collector); ok {
				e.registerer.Unregister(c)
			}
			e.unregisterMetric(m.genericMetric)
			delete(metrics, fullName)
		}
		metricName := strings.TrimPrefix(fullName, newCfg.Prefix)
		found := false
		for i := range newCfg.Metrics {
			if metricName == newCfg.Metrics[i].Name {
				found = true
				break
			}
		}
		if !found {
			e.unregisterMetric(m.genericMetric)
			delete(metrics, fullName)
		}
	}
}

func (e *EncodeProm) cleanDeletedMetrics(newCfg api.PromEncode) {
	e.cleanDeletedGeneric(newCfg, e.metricCommon.counters)
	e.cleanDeletedGeneric(newCfg, e.metricCommon.gauges)
	e.cleanDeletedGeneric(newCfg, e.metricCommon.histos)
	e.cleanDeletedGeneric(newCfg, e.metricCommon.aggHistos)
}

// returns true if a registry restart is needed
func (e *EncodeProm) checkMetricUpdate(prefix string, apiItem *api.MetricsItem, store map[string]mInfoStruct, createMetric func(string, *metrics.Preprocessed) prometheus.Collector) bool {
	fullMetricName := prefix + apiItem.Name
	plog.Debugf("Checking metric: %s", fullMetricName)
	mInfo := metrics.Preprocess(apiItem)
	if oldMetric, ok := store[fullMetricName]; ok {
		if !reflect.DeepEqual(mInfo.TargetLabels(), oldMetric.info.TargetLabels()) {
			plog.Debug("Changes detected in labels")
			return true
		}
		if !reflect.DeepEqual(mInfo.MetricsItem, oldMetric.info.MetricsItem) {
			plog.Debug("Changes detected: unregistering and replacing")
			e.unregisterMetric(oldMetric.genericMetric)
			c := createMetric(fullMetricName, mInfo)
			err := e.registerer.Register(c)
			if err != nil {
				plog.Errorf("error in prometheus.Register: %v", err)
			}
		} else {
			plog.Debug("No changes found")
		}
	} else {
		plog.Debug("New metric")
		c := createMetric(fullMetricName, mInfo)
		err := e.registerer.Register(c)
		if err != nil {
			plog.Errorf("error in prometheus.Register: %v", err)
		}
	}
	return false
}

func (e *EncodeProm) checkConfUpdate() {
	select {
	case stage := <-e.updateChan:
		cfg := api.PromEncode{}
		if stage.Encode != nil && stage.Encode.Prom != nil {
			cfg = *stage.Encode.Prom
		}
		plog.Infof("Received config update: %v", cfg)

		e.cleanDeletedMetrics(cfg)

		needNewRegistry := false
		for i := range cfg.Metrics {
			switch cfg.Metrics[i].Type {
			case api.MetricCounter:
				needNewRegistry = e.checkMetricUpdate(cfg.Prefix, &cfg.Metrics[i], e.metricCommon.counters, e.addCounter)
			case api.MetricGauge:
				needNewRegistry = e.checkMetricUpdate(cfg.Prefix, &cfg.Metrics[i], e.metricCommon.gauges, e.addGauge)
			case api.MetricHistogram:
				needNewRegistry = e.checkMetricUpdate(cfg.Prefix, &cfg.Metrics[i], e.metricCommon.histos, e.addHistogram)
			case api.MetricAggHistogram:
				needNewRegistry = e.checkMetricUpdate(cfg.Prefix, &cfg.Metrics[i], e.metricCommon.aggHistos, e.addAgghistogram)
			case "default":
				plog.Errorf("invalid metric type = %v, skipping", cfg.Metrics[i].Type)
				continue
			}
			if needNewRegistry {
				break
			}
		}
		e.cfg = &cfg
		if needNewRegistry {
			// cf https://pkg.go.dev/github.com/prometheus/client_golang@v1.19.0/prometheus#Registerer.Unregister
			plog.Info("Changes detected on labels: need registry reset.")
			e.resetRegistry()
			break
		}
	default:
		// Nothing to do
		return
	}
}

func (e *EncodeProm) resetRegistry() {
	e.metricCommon.cleanupInfoStructs()
	reg := prometheus.NewRegistry()
	e.registerer = reg
	for i := range e.cfg.Metrics {
		mCfg := &e.cfg.Metrics[i]
		fullMetricName := e.cfg.Prefix + mCfg.Name
		mInfo := metrics.Preprocess(mCfg)
		plog.Debugf("Create metric: %s, Labels: %v", fullMetricName, mInfo.TargetLabels())
		var m prometheus.Collector
		switch mCfg.Type {
		case api.MetricCounter:
			m = e.addCounter(fullMetricName, mInfo)
		case api.MetricGauge:
			m = e.addGauge(fullMetricName, mInfo)
		case api.MetricHistogram:
			m = e.addHistogram(fullMetricName, mInfo)
		case api.MetricAggHistogram:
			m = e.addAgghistogram(fullMetricName, mInfo)
		case "default":
			plog.Errorf("invalid metric type = %v, skipping", mCfg.Type)
			continue
		}
		if m != nil {
			err := e.registerer.Register(m)
			if err != nil {
				plog.Errorf("error in prometheus.Register: %v", err)
			}
		}
	}
	e.server.SetRegistry(e.regName, reg)
}

func NewEncodeProm(opMetrics *operational.Metrics, params config.StageParam) (Encoder, error) {
	cfg := api.PromEncode{}
	if params.Encode != nil && params.Encode.Prom != nil {
		cfg = *params.Encode.Prom
	}

	expiryTime := cfg.ExpiryTime
	if expiryTime.Duration == 0 {
		expiryTime.Duration = defaultExpiryTime
	}
	plog.Debugf("expiryTime = %v", expiryTime)

	registry := prometheus.NewRegistry()

	w := &EncodeProm{
		cfg:        &cfg,
		registerer: registry,
		updateChan: make(chan config.StageParam),
		server:     promserver.SharedServer,
		regName:    params.Name,
	}

	if cfg.PromConnectionInfo != nil {
		// Start new server
		w.server = promserver.StartServerAsync(cfg.PromConnectionInfo, params.Name, registry)
	}

	metricCommon := NewMetricsCommonStruct(opMetrics, cfg.MaxMetrics, params.Name, expiryTime, w.Cleanup)
	w.metricCommon = metricCommon

	// Init metrics
	w.resetRegistry()

	return w, nil
}

func (e *EncodeProm) Update(config config.StageParam) {
	e.updateChan <- config
}
