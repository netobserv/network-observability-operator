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

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/kafka"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/prometheus/client_golang/prometheus"
	kafkago "github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
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

	kafkaWriter, err := kafka.NewWriter(&config)
	if err != nil {
		return nil, err
	}

	return &encodeKafka{
		kafkaParams:    config,
		kafkaWriter:    kafkaWriter,
		recordsWritten: opMetrics.CreateRecordsWrittenCounter(params.Name),
	}, nil
}
