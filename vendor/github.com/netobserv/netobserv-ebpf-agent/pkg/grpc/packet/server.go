package pktgrpc

import (
	"context"
	"fmt"
	"net"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/pbpacket"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// CollectorServer wraps a Flow Collector connection & session
type CollectorServer struct {
	grpcServer *grpc.Server
}

type collectorOptions struct {
	grpcServerOptions []grpc.ServerOption
}

// CollectorOption allows overriding the default configuration of the CollectorServer instance.
// Use them in the StartCollector function.
type CollectorOption func(options *collectorOptions)

func WithGRPCServerOptions(options ...grpc.ServerOption) CollectorOption {
	return func(copt *collectorOptions) {
		copt.grpcServerOptions = options
	}
}

// StartCollector listens in background for gRPC+Protobuf flows in the given port, and forwards each
// set of *pbpacket.Packet by the provided channel.
func StartCollector(
	port int, pktForwarder chan<- *pbpacket.Packet, options ...CollectorOption,
) (*CollectorServer, error) {
	copts := collectorOptions{}
	for _, opt := range options {
		opt(&copts)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer(copts.grpcServerOptions...)
	pbpacket.RegisterCollectorServer(grpcServer, &collectorAPI{
		pktForwarder: pktForwarder,
	})
	reflection.Register(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			panic("error connecting to server: " + err.Error())
		}
	}()
	return &CollectorServer{
		grpcServer: grpcServer,
	}, nil
}

func (c *CollectorServer) Close() error {
	c.grpcServer.Stop()
	return nil
}

type collectorAPI struct {
	pbpacket.UnimplementedCollectorServer
	pktForwarder chan<- *pbpacket.Packet
}

var okReply = &pbpacket.CollectorReply{}

func (c *collectorAPI) Send(_ context.Context, pkts *pbpacket.Packet) (*pbpacket.CollectorReply, error) {
	c.pktForwarder <- pkts
	return okReply, nil
}
