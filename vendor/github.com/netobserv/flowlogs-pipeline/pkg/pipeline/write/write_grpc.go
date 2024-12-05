package write

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/write/grpc"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/write/grpc/genericmap"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
)

type writeGRPC struct {
	hostIP     string
	hostPort   int
	clientConn *grpc.ClientConnection
}

// Write writes a flow before being stored
func (t *writeGRPC) Write(v config.GenericMap) {
	logrus.Tracef("entering writeGRPC Write %s", v)
	value, _ := json.Marshal(v)
	if _, err := t.clientConn.Client().Send(context.TODO(), &genericmap.Flow{
		GenericMap: &anypb.Any{
			Value: value,
		},
	}); err != nil {
		logrus.Errorf("writeGRPC send error: %v", err)
	}
}

// NewWriteGRPC create a new write
func NewWriteGRPC(params config.StageParam) (Writer, error) {
	logrus.Debugf("entering NewWriteGRPC")

	writeGRPC := &writeGRPC{}
	if params.Write != nil && params.Write.GRPC != nil {
		if err := params.Write.GRPC.Validate(); err != nil {
			return nil, fmt.Errorf("the provided config is not valid: %w", err)
		}
		writeGRPC.hostIP = params.Write.GRPC.TargetHost
		writeGRPC.hostPort = params.Write.GRPC.TargetPort
	} else {
		return nil, fmt.Errorf("write.grpc param is mandatory: %v", params.Write)
	}
	logrus.Debugf("NewWriteGRPC ConnectClient %s:%d...", writeGRPC.hostIP, writeGRPC.hostPort)
	clientConn, err := grpc.ConnectClient(writeGRPC.hostIP, writeGRPC.hostPort)
	if err != nil {
		return nil, err
	}
	writeGRPC.clientConn = clientConn
	return writeGRPC, nil
}
