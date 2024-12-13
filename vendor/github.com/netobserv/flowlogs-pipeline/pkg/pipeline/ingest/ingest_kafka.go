/*
 * Copyright (C) 2022 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	 http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package ingest

import (
	"errors"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/decode"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var klog = logrus.WithField("component", "ingest.Kafka")

type kafkaReadMessage interface {
	ReadMessage(ctx context.Context) (kafkago.Message, error)
	Config() kafkago.ReaderConfig
	Stats() kafkago.ReaderStats
}

type ingestKafka struct {
	kafkaReader      kafkaReadMessage
	decoder          decode.Decoder
	in               chan []byte
	exitChan         <-chan struct{}
	batchReadTimeout int64
	batchMaxLength   int
	metrics          *metrics
	canLogMessages   bool
}

const defaultBatchReadTimeout = int64(1000)
const defaultKafkaBatchMaxLength = 500
const defaultKafkaCommitInterval = 500

const kafkaStatsPeriod = 15 * time.Second

// Ingest ingests entries from kafka topic
func (k *ingestKafka) Ingest(out chan<- config.GenericMap) {
	klog.Debugf("entering ingestKafka.Ingest")
	k.metrics.createOutQueueLen(out)

	// initialize background listener
	k.kafkaListener()

	// forever process log lines received by collector
	k.processLogLines(out)
}

// background thread to read kafka messages; place received items into ingestKafka input channel
func (k *ingestKafka) kafkaListener() {
	klog.Debugf("entering kafkaListener")

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		go k.reportStats()
	}

	go func() {
		for {
			if k.isStopped() {
				klog.Info("gracefully exiting")
				return
			}
			klog.Trace("fetching messages from Kafka")
			// block until a message arrives
			kafkaMessage, err := k.kafkaReader.ReadMessage(context.Background())
			if err != nil {
				klog.Errorln(err)
				k.metrics.error("Cannot read message")
				continue
			}
			if k.canLogMessages && logrus.IsLevelEnabled(logrus.TraceLevel) {
				klog.Tracef("string(kafkaMessage) = %s\n", string(kafkaMessage.Value))
			}
			k.metrics.flowsProcessed.Inc()
			messageLen := len(kafkaMessage.Value)
			k.metrics.batchSizeBytes.Observe(float64(messageLen) + float64(len(kafkaMessage.Key)))
			if messageLen > 0 {
				// process message
				k.in <- kafkaMessage.Value
			}
		}
	}()
}

func (k *ingestKafka) isStopped() bool {
	select {
	case <-k.exitChan:
		return true
	default:
		return false
	}
}

func (k *ingestKafka) processRecordDelay(record config.GenericMap) {
	timeFlowEndInterface, ok := record["TimeFlowEndMs"]
	if !ok {
		// "trace" level used to minimize performance impact
		klog.Tracef("TimeFlowEndMs missing in record %v", record)
		k.metrics.error("TimeFlowEndMs missing")
		return
	}
	timeFlowEnd, ok := timeFlowEndInterface.(int64)
	if !ok {
		// "trace" level used to minimize performance impact
		klog.Tracef("Cannot parse TimeFlowEndMs of record %v", record)
		k.metrics.error("Cannot parse TimeFlowEndMs")
		return
	}
	delay := time.Since(time.UnixMilli(timeFlowEnd)).Seconds()
	k.metrics.latency.Observe(delay)
}

func (k *ingestKafka) processRecord(record []byte, out chan<- config.GenericMap) {
	// Decode batch
	decoded, err := k.decoder.Decode(record)
	if err != nil {
		klog.WithError(err).Warnf("ignoring flow")
		return
	}
	k.processRecordDelay(decoded)

	// Send batch
	out <- decoded
}

// read items from ingestKafka input channel, pool them, and send down the pipeline
func (k *ingestKafka) processLogLines(out chan<- config.GenericMap) {
	for {
		select {
		case <-k.exitChan:
			klog.Debugf("exiting ingestKafka because of signal")
			return
		case record := <-k.in:
			k.processRecord(record, out)
		}
	}
}

// reportStats periodically reports kafka stats
func (k *ingestKafka) reportStats() {
	ticker := time.NewTicker(kafkaStatsPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-k.exitChan:
			klog.Debug("gracefully exiting stats reporter")
		case <-ticker.C:
			klog.Debugf("reader stats: %#v", k.kafkaReader.Stats())
		}
	}
}

// NewIngestKafka create a new ingester
// nolint:cyclop
func NewIngestKafka(opMetrics *operational.Metrics, params config.StageParam) (Ingester, error) {
	klog.Debugf("entering NewIngestKafka")
	jsonIngestKafka := api.IngestKafka{}
	var ingestType string
	if params.Ingest != nil {
		ingestType = params.Ingest.Type
		if params.Ingest.Kafka != nil {
			jsonIngestKafka = *params.Ingest.Kafka
		}
	}

	// connect to the kafka server
	startOffsetString := jsonIngestKafka.StartOffset
	var startOffset int64
	switch startOffsetString {
	case "FirstOffset", "":
		startOffset = kafkago.FirstOffset
	case "LastOffset":
		startOffset = kafkago.LastOffset
	default:
		startOffset = kafkago.FirstOffset
		klog.Errorf("illegal value for StartOffset: %s\n", startOffsetString)
	}
	klog.Debugf("startOffset = %v", startOffset)
	groupBalancers := make([]kafkago.GroupBalancer, 0)
	for _, gb := range jsonIngestKafka.GroupBalancers {
		switch gb {
		case "range":
			groupBalancers = append(groupBalancers, &kafkago.RangeGroupBalancer{})
		case "roundRobin":
			groupBalancers = append(groupBalancers, &kafkago.RoundRobinGroupBalancer{})
		case "rackAffinity":
			groupBalancers = append(groupBalancers, &kafkago.RackAffinityGroupBalancer{})
		default:
			klog.Warningf("groupbalancers parameter missing")
			groupBalancers = append(groupBalancers, &kafkago.RoundRobinGroupBalancer{})
		}
	}

	batchReadTimeout := defaultBatchReadTimeout
	if jsonIngestKafka.BatchReadTimeout != 0 {
		batchReadTimeout = jsonIngestKafka.BatchReadTimeout
	}
	klog.Infof("batchReadTimeout = %d", batchReadTimeout)

	commitInterval := int64(defaultKafkaCommitInterval)
	if jsonIngestKafka.CommitInterval != 0 {
		commitInterval = jsonIngestKafka.CommitInterval
	}
	klog.Infof("commitInterval = %d", jsonIngestKafka.CommitInterval)

	dialer := &kafkago.Dialer{
		Timeout:   kafkago.DefaultDialer.Timeout,
		DualStack: kafkago.DefaultDialer.DualStack,
	}
	if jsonIngestKafka.TLS != nil {
		klog.Infof("Using TLS configuration: %v", jsonIngestKafka.TLS)
		tlsConfig, err := jsonIngestKafka.TLS.Build()
		if err != nil {
			return nil, err
		}
		dialer.TLS = tlsConfig
	}

	if jsonIngestKafka.SASL != nil {
		m, err := utils.SetupSASLMechanism(jsonIngestKafka.SASL)
		if err != nil {
			return nil, err
		}
		dialer.SASLMechanism = m
	}

	readerConfig := kafkago.ReaderConfig{
		Brokers:        jsonIngestKafka.Brokers,
		Topic:          jsonIngestKafka.Topic,
		GroupID:        jsonIngestKafka.GroupID,
		GroupBalancers: groupBalancers,
		StartOffset:    startOffset,
		CommitInterval: time.Duration(commitInterval) * time.Millisecond,
		Dialer:         dialer,
	}

	if jsonIngestKafka.PullQueueCapacity > 0 {
		readerConfig.QueueCapacity = jsonIngestKafka.PullQueueCapacity
	}

	if jsonIngestKafka.PullMaxBytes > 0 {
		readerConfig.MaxBytes = jsonIngestKafka.PullMaxBytes
	}

	klog.Debugf("reader config: %#v", readerConfig)

	kafkaReader := kafkago.NewReader(readerConfig)
	if kafkaReader == nil {
		errMsg := "NewIngestKafka: failed to create kafka-go reader"
		klog.Errorf("%s", errMsg)
		return nil, errors.New(errMsg)
	}

	decoder, err := decode.GetDecoder(jsonIngestKafka.Decoder)
	if err != nil {
		return nil, err
	}

	bml := defaultKafkaBatchMaxLength
	if jsonIngestKafka.BatchMaxLen != 0 {
		bml = jsonIngestKafka.BatchMaxLen
	}

	in := make(chan []byte, 2*bml)
	metrics := newMetrics(opMetrics, params.Name, ingestType, func() int { return len(in) })

	return &ingestKafka{
		kafkaReader:      kafkaReader,
		decoder:          decoder,
		exitChan:         utils.ExitChannel(),
		in:               in,
		batchMaxLength:   bml,
		batchReadTimeout: batchReadTimeout,
		metrics:          metrics,
		canLogMessages:   jsonIngestKafka.Decoder.Type == api.DecoderJSON,
	}, nil
}
