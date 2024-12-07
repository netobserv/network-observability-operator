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
	"context"
	"encoding/binary"
	"fmt"
	"net"

	ms "github.com/mitchellh/mapstructure"
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	pUtils "github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	"github.com/netsampler/goflow2/decoders/netflow/templates"
	_ "github.com/netsampler/goflow2/decoders/netflow/templates/memory" // required for goflow in-memory templates
	goflowFormat "github.com/netsampler/goflow2/format"
	goflowCommonFormat "github.com/netsampler/goflow2/format/common"
	_ "github.com/netsampler/goflow2/format/protobuf" // required for goflow protobuf
	goflowpb "github.com/netsampler/goflow2/pb"
	"github.com/netsampler/goflow2/utils"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	channelSize = 1000
)

type ingestCollector struct {
	hostname   string
	port       int
	portLegacy int
	in         chan map[string]interface{}
	exitChan   <-chan struct{}
	metrics    *metrics
}

// TransportWrapper is an implementation of the goflow2 transport interface
type TransportWrapper struct {
	c chan map[string]interface{}
}

func NewWrapper(c chan map[string]interface{}) *TransportWrapper {
	tw := TransportWrapper{c: c}
	return &tw
}

func RenderMessage(message *goflowpb.FlowMessage) (map[string]interface{}, error) {
	outputMap := make(map[string]interface{})
	err := ms.Decode(message, &outputMap)
	if err != nil {
		return nil, err
	}
	outputMap["DstAddr"] = goflowCommonFormat.RenderIP(message.DstAddr)
	outputMap["SrcAddr"] = goflowCommonFormat.RenderIP(message.SrcAddr)
	outputMap["DstMac"] = renderMac(message.DstMac)
	outputMap["SrcMac"] = renderMac(message.SrcMac)
	return outputMap, nil
}

func renderMac(macValue uint64) string {
	mac := make([]byte, 8)
	binary.BigEndian.PutUint64(mac, macValue)
	return net.HardwareAddr(mac[2:]).String()
}

func (w *TransportWrapper) Send(_, data []byte) error {
	message := goflowpb.FlowMessage{}
	err := proto.Unmarshal(data, &message)
	if err != nil {
		// temporary fix
		// A PR was submitted to log this error from goflow2:
		// https://github.com/netsampler/goflow2/pull/86
		log.Error(err)
		return err
	}
	renderedMsg, err := RenderMessage(&message)
	if err == nil {
		w.c <- renderedMsg
	}
	return err
}

// Ingest ingests entries from a network collector using goflow2 library (https://github.com/netsampler/goflow2)
func (c *ingestCollector) Ingest(out chan<- config.GenericMap) {
	ctx := context.Background()
	c.metrics.createOutQueueLen(out)

	// initialize background listeners (a.k.a.netflow+legacy collector)
	c.initCollectorListener(ctx)

	// forever process log lines received by collector
	c.processLogLines(out)
}

func (c *ingestCollector) initCollectorListener(ctx context.Context) {
	transporter := NewWrapper(c.in)
	formatter, err := goflowFormat.FindFormat(ctx, "pb")
	if err != nil {
		log.Fatal(err)
	}

	if c.port > 0 {
		// cf https://github.com/netsampler/goflow2/pull/49
		tpl, err := templates.FindTemplateSystem(ctx, "memory")
		if err != nil {
			log.Fatalf("goflow2 error: could not find memory template system: %v", err)
		}
		defer tpl.Close(ctx)

		go func() {
			sNF := utils.NewStateNetFlow()
			sNF.Format = formatter
			sNF.Transport = transporter
			sNF.Logger = log.StandardLogger()
			sNF.TemplateSystem = tpl

			log.Infof("listening for netflow on host %s, port = %d", c.hostname, c.port)
			err = sNF.FlowRoutine(1, c.hostname, c.port, false)
			log.Fatal(err)
		}()
	}

	if c.portLegacy > 0 {
		go func() {
			sLegacyNF := utils.NewStateNFLegacy()
			sLegacyNF.Format = formatter
			sLegacyNF.Transport = transporter
			sLegacyNF.Logger = log.StandardLogger()

			log.Infof("listening for legacy netflow on host %s, port = %d", c.hostname, c.portLegacy)
			err = sLegacyNF.FlowRoutine(1, c.hostname, c.portLegacy, false)
			log.Fatal(err)
		}()
	}
}

func (c *ingestCollector) processLogLines(out chan<- config.GenericMap) {
	for {
		select {
		case <-c.exitChan:
			log.Debugf("exiting ingestCollector because of signal")
			return
		case record := <-c.in:
			out <- record
		}
	}
}

// NewIngestCollector create a new ingester
func NewIngestCollector(opMetrics *operational.Metrics, params config.StageParam) (Ingester, error) {
	jsonIngestCollector := api.IngestCollector{}
	if params.Ingest != nil && params.Ingest.Collector != nil {
		jsonIngestCollector = *params.Ingest.Collector
	}
	if jsonIngestCollector.HostName == "" {
		return nil, fmt.Errorf("ingest hostname not specified")
	}
	if jsonIngestCollector.Port == 0 && jsonIngestCollector.PortLegacy == 0 {
		return nil, fmt.Errorf("no ingest port specified")
	}

	log.Infof("hostname = %s", jsonIngestCollector.HostName)
	log.Infof("port = %d", jsonIngestCollector.Port)
	log.Infof("portLegacy = %d", jsonIngestCollector.PortLegacy)

	in := make(chan map[string]interface{}, channelSize)
	metrics := newMetrics(opMetrics, params.Name, params.Ingest.Type, func() int { return len(in) })

	return &ingestCollector{
		hostname:   jsonIngestCollector.HostName,
		port:       jsonIngestCollector.Port,
		portLegacy: jsonIngestCollector.PortLegacy,
		exitChan:   pUtils.ExitChannel(),
		in:         in,
		metrics:    metrics,
	}, nil
}
