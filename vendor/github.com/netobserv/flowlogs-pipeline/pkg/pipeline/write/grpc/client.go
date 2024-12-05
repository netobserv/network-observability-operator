package grpc

import (
	"flag"
	"log"

	pb "github.com/netobserv/flowlogs-pipeline/pkg/pipeline/write/grpc/genericmap"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ClientConnection wraps a gRPC+protobuf connection
type ClientConnection struct {
	client pb.CollectorClient
	conn   *grpc.ClientConn
}

func ConnectClient(hostIP string, hostPort int) (*ClientConnection, error) {
	flag.Parse()
	// Set up a connection to the server.
	socket := utils.GetSocket(hostIP, hostPort)
	conn, err := grpc.NewClient(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &ClientConnection{
		client: pb.NewCollectorClient(conn),
		conn:   conn,
	}, nil
}

func (cp *ClientConnection) Client() pb.CollectorClient {
	return cp.client
}

func (cp *ClientConnection) Close() error {
	return cp.conn.Close()
}
