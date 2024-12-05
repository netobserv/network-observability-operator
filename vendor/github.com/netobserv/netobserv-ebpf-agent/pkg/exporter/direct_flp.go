package exporter

import (
	"fmt"

	flpconfig "github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/decode"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/decode/packets"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
	"gopkg.in/yaml.v2"
)

// DirectFLP flow exporter
type DirectFLP struct {
	fwd chan flpconfig.GenericMap
}

func StartDirectFLP(jsonConfig string, bufLen int) (*DirectFLP, error) {
	var cfg flpconfig.ConfigFileStruct
	// Note that, despite jsonConfig being json, we use yaml unmarshaler because the json one
	// is screwed up for HTTPClientConfig in github.com/prometheus/common/config (used for Loki)
	// This is ok as YAML is a superset of JSON.
	// E.g. try unmarshaling `{"clientConfig":{"proxy_url":null}}` as a api.WriteLoki
	// See also https://github.com/prometheus/prometheus/issues/11816
	if err := yaml.Unmarshal([]byte(jsonConfig), &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	fwd := make(chan flpconfig.GenericMap, bufLen)
	err := pipeline.StartFLPInProcess(&cfg, fwd)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pipeline %w", err)
	}

	return &DirectFLP{fwd: fwd}, nil
}

// ExportFlows accepts slices of *model.Record by its input channel, converts them
// to *pbflow.Records instances, and submits them to the collector.
func (d *DirectFLP) ExportFlows(input <-chan []*model.Record) {
	for inputRecords := range input {
		for _, rec := range inputRecords {
			d.fwd <- decode.RecordToMap(rec)
		}
	}
}

// ExportPackets accepts slices of *model.PacketRecord by its input channel, converts them
// to *pbflow.Records instances, and submits them to the collector.
func (d *DirectFLP) ExportPackets(input <-chan []*model.PacketRecord) {
	for inputPackets := range input {
		for _, packet := range inputPackets {
			if len(packet.Stream) != 0 {
				d.fwd <- packets.PacketToMap(packet)
			}
		}
	}
}

func (d *DirectFLP) Close() {
	close(d.fwd)
}
