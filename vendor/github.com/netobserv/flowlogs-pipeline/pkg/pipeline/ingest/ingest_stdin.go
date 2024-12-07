/*
 * Copyright (C) 2023 IBM, Inc.
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
	"bufio"
	"os"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/decode"
	pUtils "github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	"github.com/sirupsen/logrus"
)

const (
	stdinChannelSize = 1000
)

var slog = logrus.WithField("component", "ingest.Stdin")

type ingestStdin struct {
	in       chan string
	eof      chan struct{}
	exitChan <-chan struct{}
	metrics  *metrics
	decoder  decode.Decoder
}

// Ingest ingests entries from stdin
func (s *ingestStdin) Ingest(out chan<- config.GenericMap) {
	slog.Debugf("entering ingestStdin.Ingest")
	s.metrics.createOutQueueLen(out)

	go s.getStdinInput()

	// process log lines received by stdin
	s.processLogLines(out)
}

func (s *ingestStdin) getStdinInput() {
	scanner := bufio.NewScanner(os.Stdin)
	// Loop to read lines from stdin until an error or EOF is encountered
	for scanner.Scan() {
		s.in <- scanner.Text()
	}

	// Check for errors
	if err := scanner.Err(); err != nil {
		slog.WithError(err).Errorf("Error reading standard input")
	}
	close(s.eof)
}

func (s *ingestStdin) processLogLines(out chan<- config.GenericMap) {
	for {
		select {
		case <-s.exitChan:
			slog.Debugf("exiting ingestStdin because of signal")
			return
		case <-s.eof:
			slog.Debugf("exiting ingestStdin because of EOF")
			return
		case line := <-s.in:
			s.processRecord(out, line)
		}
	}
}

func (s *ingestStdin) processRecord(out chan<- config.GenericMap, line string) {
	slog.Debugf("Decoding %s", line)
	decoded, err := s.decoder.Decode([]byte(line))
	if err != nil {
		slog.WithError(err).Warnf("ignoring line %v", line)
		s.metrics.error("Ignoring line")
		return
	}
	s.metrics.flowsProcessed.Inc()
	out <- decoded
}

// NewIngestStdin create a new ingester
func NewIngestStdin(opMetrics *operational.Metrics, params config.StageParam) (Ingester, error) {
	slog.Debugf("Entering NewIngestStdin")

	in := make(chan string, stdinChannelSize)
	eof := make(chan struct{})
	metrics := newMetrics(opMetrics, params.Name, params.Ingest.Type, func() int { return len(in) })
	decoderParams := api.Decoder{Type: api.DecoderJSON}
	decoder, err := decode.GetDecoder(decoderParams)
	if err != nil {
		return nil, err
	}

	return &ingestStdin{
		exitChan: pUtils.ExitChannel(),
		in:       in,
		eof:      eof,
		metrics:  metrics,
		decoder:  decoder,
	}, nil
}
