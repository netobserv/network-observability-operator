package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/write/grpc/genericmap"
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
// set of *genericmap.Flow by the provided channel.
func StartCollector(
	port int, recordForwarder chan<- *genericmap.Flow, options ...CollectorOption,
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
	genericmap.RegisterCollectorServer(grpcServer, &collectorAPI{
		recordForwarder: recordForwarder,
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
	genericmap.UnimplementedCollectorServer
	recordForwarder chan<- *genericmap.Flow
}

var okReply = &genericmap.CollectorReply{}

func (c *collectorAPI) Send(_ context.Context, records *genericmap.Flow) (*genericmap.CollectorReply, error) {
	c.recordForwarder <- records
	return okReply, nil
}
