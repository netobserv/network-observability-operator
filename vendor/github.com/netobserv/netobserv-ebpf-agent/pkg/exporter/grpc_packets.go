package exporter

import (
	"context"
	"time"

	grpc "github.com/netobserv/netobserv-ebpf-agent/pkg/grpc/packet"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/pbpacket"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/utils/packets"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
)

type GRPCPacketProto struct {
	hostIP     string
	hostPort   int
	clientConn *grpc.ClientConnection
}

var gplog = logrus.WithField("component", "packet/GRPCPackets")

// WritePacket writes the given packet data out to gRPC.
func writeGRPCPacket(time time.Time, data []byte, conn *grpc.ClientConnection) error {
	bytes, err := packets.GetPacketBytesWithHeader(time, data)
	if err != nil {
		return err
	}
	_, err = conn.Client().Send(context.TODO(), &pbpacket.Packet{
		Pcap: &anypb.Any{
			Value: bytes,
		},
	})
	return err
}

func StartGRPCPacketSend(hostIP string, hostPort int) (*GRPCPacketProto, error) {
	clientConn, err := grpc.ConnectClient(hostIP, hostPort)
	if err != nil {
		return nil, err
	}
	return &GRPCPacketProto{
		hostIP:     hostIP,
		hostPort:   hostPort,
		clientConn: clientConn,
	}, nil
}

func (p *GRPCPacketProto) ExportGRPCPackets(in <-chan []*model.PacketRecord) {
	for packetRecord := range in {
		var errs []error
		for _, packet := range packetRecord {
			if len(packet.Stream) != 0 {
				if err := writeGRPCPacket(packet.Time, packet.Stream, p.clientConn); err != nil {
					errs = append(errs, err)
				}
			}
		}
		if len(errs) != 0 {
			gplog.Errorf("%d errors while sending packets:\n%s", len(errs), errs)
		}
	}
	if err := p.clientConn.Close(); err != nil {
		gplog.WithError(err).Warn("couldn't close packet export client")
	}
}
