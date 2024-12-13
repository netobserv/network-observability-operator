/*
 * Copyright (C) 2021 IBM, Inc.
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
	"fmt"
	"os"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/decode"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	log "github.com/sirupsen/logrus"
)

type ingestFile struct {
	params       config.Ingest
	decoder      decode.Decoder
	exitChan     <-chan struct{}
	PrevRecords  []config.GenericMap
	TotalRecords int
}

const (
	delaySeconds = 10
	chunkLines   = 100
)

// Ingest ingests entries from a file and resends the same data every delaySeconds seconds
func (ingestF *ingestFile) Ingest(out chan<- config.GenericMap) {
	var filename string
	if ingestF.params.File != nil {
		filename = ingestF.params.File.Filename
	}
	var lines [][]byte
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		log.Debugf("%s", text)
		lines = append(lines, []byte(text))
	}

	log.Debugf("Ingesting %d log lines from %s", len(lines), filename)
	switch ingestF.params.Type {
	case "file":
		ingestF.sendAllLines(lines, out)
	case "file_loop":
		// loop forever
		ticker := time.NewTicker(time.Duration(delaySeconds) * time.Second)
		for {
			select {
			case <-ingestF.exitChan:
				log.Debugf("exiting ingestFile because of signal")
				return
			case <-ticker.C:
				ingestF.sendAllLines(lines, out)
			}
		}
	case "file_chunks":
		// sends the lines in chunks. Useful for testing parallelization
		ingestF.TotalRecords = len(lines)
		for len(lines) > 0 {
			if len(lines) > chunkLines {
				ingestF.sendAllLines(lines[:chunkLines], out)
				lines = lines[chunkLines:]
			} else {
				ingestF.sendAllLines(lines, out)
				lines = nil
			}
		}
	}
}

func (ingestF *ingestFile) sendAllLines(lines [][]byte, out chan<- config.GenericMap) {
	log.Debugf("ingestFile sending %d lines", len(lines))
	ingestF.TotalRecords = len(lines)
	for _, line := range lines {
		decoded, err := ingestF.decoder.Decode(line)
		if err != nil {
			log.WithError(err).Warnf("ignoring line")
			continue
		}
		out <- decoded
	}
}

// NewIngestFile create a new ingester
func NewIngestFile(params config.StageParam) (Ingester, error) {
	log.Debugf("entering NewIngestFile")
	if params.Ingest == nil || params.Ingest.File == nil || params.Ingest.File.Filename == "" {
		return nil, fmt.Errorf("ingest filename not specified")
	}

	log.Debugf("input file name = %s", params.Ingest.File.Filename)
	decoder, err := decode.GetDecoder(params.Ingest.File.Decoder)
	if err != nil {
		return nil, err
	}

	return &ingestFile{
		params:   *params.Ingest,
		exitChan: utils.ExitChannel(),
		decoder:  decoder,
	}, nil
}
