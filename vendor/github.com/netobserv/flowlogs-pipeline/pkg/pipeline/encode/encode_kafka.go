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

package encode

import (
	"encoding/json"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	"github.com/prometheus/client_golang/prometheus"
	kafkago "github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	defaultReadTimeoutSeconds  = int64(10)
	defaultWriteTimeoutSeconds = int64(10)
)

type kafkaWriteMessage interface {
	WriteMessages(ctx context.Context, msgs ...kafkago.Message) error
}

type encodeKafka struct {
	kafkaParams    api.EncodeKafka
	kafkaWriter    kafkaWriteMessage
	recordsWritten prometheus.Counter
}

// Encode writes entries to kafka topic
func (r *encodeKafka) Encode(entry config.GenericMap) {
	var entryByteArray []byte
	var err error
	entryByteArray, err = json.Marshal(entry)
	if err != nil {
		log.Errorf("encodeKafka error: %v", err)
		return
	}
	msg := kafkago.Message{
		Value: entryByteArray,
	}
	err = r.kafkaWriter.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Errorf("encodeKafka error: %v", err)
	} else {
		r.recordsWritten.Inc()
	}
}

func (r *encodeKafka) Update(_ config.StageParam) {
	log.Warn("Encode Kafka, update not supported")
}

// NewEncodeKafka create a new writer to kafka
func NewEncodeKafka(opMetrics *operational.Metrics, params config.StageParam) (Encoder, error) {
	log.Debugf("entering NewEncodeKafka")
	config := api.EncodeKafka{}
	if params.Encode != nil && params.Encode.Kafka != nil {
		config = *params.Encode.Kafka
	}

	var balancer kafkago.Balancer
	switch config.Balancer {
	case api.KafkaRoundRobin:
		balancer = &kafkago.RoundRobin{}
	case api.KafkaLeastBytes:
		balancer = &kafkago.LeastBytes{}
	case api.KafkaHash:
		balancer = &kafkago.Hash{}
	case api.KafkaCrc32:
		balancer = &kafkago.CRC32Balancer{}
	case api.KafkaMurmur2:
		balancer = &kafkago.Murmur2Balancer{}
	default:
		balancer = nil
	}

	readTimeoutSecs := defaultReadTimeoutSeconds
	if config.ReadTimeout != 0 {
		readTimeoutSecs = config.ReadTimeout
	}

	writeTimeoutSecs := defaultWriteTimeoutSeconds
	if config.WriteTimeout != 0 {
		writeTimeoutSecs = config.WriteTimeout
	}

	transport := kafkago.Transport{}
	if config.TLS != nil {
		log.Infof("Using TLS configuration: %v", config.TLS)
		tlsConfig, err := config.TLS.Build()
		if err != nil {
			return nil, err
		}
		transport.TLS = tlsConfig
	}

	if config.SASL != nil {
		m, err := utils.SetupSASLMechanism(config.SASL)
		if err != nil {
			return nil, err
		}
		transport.SASL = m
	}

	// connect to the kafka server
	kafkaWriter := kafkago.Writer{
		Addr:         kafkago.TCP(config.Address),
		Topic:        config.Topic,
		Balancer:     balancer,
		ReadTimeout:  time.Duration(readTimeoutSecs) * time.Second,
		WriteTimeout: time.Duration(writeTimeoutSecs) * time.Second,
		BatchSize:    config.BatchSize,
		BatchBytes:   config.BatchBytes,
		// Temporary fix may be we should implement a batching systems
		// https://github.com/segmentio/kafka-go/issues/326#issuecomment-519375403
		BatchTimeout: time.Nanosecond,
		Transport:    &transport,
	}

	return &encodeKafka{
		kafkaParams:    config,
		kafkaWriter:    &kafkaWriter,
		recordsWritten: opMetrics.CreateRecordsWrittenCounter(params.Name),
	}, nil
}
