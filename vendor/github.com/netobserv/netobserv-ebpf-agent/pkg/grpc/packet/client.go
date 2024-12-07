// Package pktgrpc provides the basic interfaces to build a gRPC+Protobuf packet client & server
package pktgrpc

import (
	"github.com/netobserv/netobserv-ebpf-agent/pkg/pbpacket"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ClientConnection wraps a gRPC+protobuf connection
type ClientConnection struct {
	client pbpacket.CollectorClient
	conn   *grpc.ClientConn
}

func ConnectClient(hostIP string, hostPort int) (*ClientConnection, error) {
	// TODO: allow configuring some options (keepalive, backoff...)
	socket := utils.GetSocket(hostIP, hostPort)
	conn, err := grpc.NewClient(socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &ClientConnection{
		client: pbpacket.NewCollectorClient(conn),
		conn:   conn,
	}, nil
}

func (cp *ClientConnection) Client() pbpacket.CollectorClient {
	return cp.client
}

func (cp *ClientConnection) Close() error {
	return cp.conn.Close()
}
