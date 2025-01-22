package exporter

import (
	"context"

	grpc "github.com/netobserv/netobserv-ebpf-agent/pkg/grpc/flow"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/metrics"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/pbflow"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var glog = logrus.WithField("component", "exporter/GRPCProto")

const componentGRPC = "grpc"

// GRPCProto flow exporter. Its ExportFlows method accepts slices of *model.Record
// by its input channel, converts them to *pbflow.Records instances, and submits
// them to the collector.
type GRPCProto struct {
	hostIP     string
	hostPort   int
	clientConn *grpc.ClientConnection
	// maxFlowsPerMessage limits the maximum number of flows per GRPC message.
	// If a message contains more flows than this number, the GRPC message will be split into
	// multiple messages.
	maxFlowsPerMessage int
	metrics            *metrics.Metrics
	batchCounter       prometheus.Counter
}

func StartGRPCProto(hostIP string, hostPort int, maxFlowsPerMessage int, m *metrics.Metrics) (*GRPCProto, error) {
	clientConn, err := grpc.ConnectClient(hostIP, hostPort)
	if err != nil {
		return nil, err
	}
	return &GRPCProto{
		hostIP:             hostIP,
		hostPort:           hostPort,
		clientConn:         clientConn,
		maxFlowsPerMessage: maxFlowsPerMessage,
		metrics:            m,
		batchCounter:       m.CreateBatchCounter(componentGRPC),
	}, nil
}

// ExportFlows accepts slices of *model.Record by its input channel, converts them
// to *pbflow.Records instances, and submits them to the collector.
func (g *GRPCProto) ExportFlows(input <-chan []*model.Record) {
	socket := utils.GetSocket(g.hostIP, g.hostPort)
	log := glog.WithField("collector", socket)
	for inputRecords := range input {
		g.metrics.EvictionCounter.WithSource(componentGRPC).Inc()
		for _, pbRecords := range pbflow.FlowsToPB(inputRecords, g.maxFlowsPerMessage) {
			log.Debugf("sending %d records", len(pbRecords.Entries))
			if _, err := g.clientConn.Client().Send(context.TODO(), pbRecords); err != nil {
				g.metrics.Errors.WithErrorName(componentGRPC, "CannotWriteMessage").Inc()
				log.WithError(err).Error("couldn't send flow records to collector")
			}
			g.batchCounter.Inc()
			g.metrics.EvictedFlowsCounter.WithSource(componentGRPC).Add(float64(len(pbRecords.Entries)))
		}
	}
	if err := g.clientConn.Close(); err != nil {
		log.WithError(err).Warn("couldn't close flow export client")
		g.metrics.Errors.WithErrorName(componentGRPC, "CannotCloseClient").Inc()
	}
}
