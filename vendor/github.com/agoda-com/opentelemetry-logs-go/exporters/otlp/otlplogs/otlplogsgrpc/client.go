/*
Copyright Agoda Services Co.,Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package otlplogsgrpc

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/internal"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/internal/otlpconfig"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/internal/retry"

	"go.opentelemetry.io/otel"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcClient struct {
	endpoint      string
	dialOpts      []grpc.DialOption
	metadata      metadata.MD
	exportTimeout time.Duration
	requestFunc   retry.RequestFunc

	// stopCtx is used as a parent context for all exports. Therefore, when it
	// is canceled with the stopFunc all exports are canceled.
	stopCtx context.Context
	// stopFunc cancels stopCtx, stopping any active exports.
	stopFunc context.CancelFunc

	// ourConn keeps track of where conn was created: true if created here on
	// Start, or false if passed with an option. This is important on Shutdown
	// as the conn should only be closed if created here on start. Otherwise,
	// it is up to the processes that passed the conn to close it.
	ourConn bool
	conn    *grpc.ClientConn
	tscMu   sync.RWMutex
	tsc     collogspb.LogsServiceClient
}

func NewClient(opts ...Option) *grpcClient {
	cfg := otlpconfig.NewGRPCConfig(asGRPCOptions(opts)...)

	ctx, cancel := context.WithCancel(context.Background())

	c := &grpcClient{
		endpoint:      cfg.Logs.Endpoint,
		exportTimeout: cfg.Logs.Timeout,
		requestFunc:   cfg.RetryConfig.RequestFunc(retryable),
		dialOpts:      cfg.DialOptions,
		stopCtx:       ctx,
		stopFunc:      cancel,
		conn:          cfg.GRPCConn,
	}

	if len(cfg.Logs.Headers) > 0 {
		c.metadata = metadata.New(cfg.Logs.Headers)
	}

	return c
}

// Start establishes a gRPC connection to the collector.
func (c *grpcClient) Start(ctx context.Context) error {
	if c.conn == nil {
		// If the caller did not provide a ClientConn when the grpcClient was
		// created, create one using the configuration they did provide.
		conn, err := grpc.DialContext(ctx, c.endpoint, c.dialOpts...)
		if err != nil {
			return err
		}
		// Keep track that we own the lifecycle of this conn and need to close
		// it on Shutdown.
		c.ourConn = true
		c.conn = conn
	}

	// The otlplogs.Client interface states this method is called just once,
	// so no need to check if already started.
	c.tscMu.Lock()
	c.tsc = collogspb.NewLogsServiceClient(c.conn)
	c.tscMu.Unlock()

	return nil
}

var errAlreadyStopped = errors.New("the grpcClient is already stopped")

// Stop shuts down the grpcClient.
//
// Any active connections to a remote endpoint are closed if they were created
// by the grpcClient. Any gRPC connection passed during creation using
// WithGRPCConn will not be closed. It is the caller's responsibility to
// handle cleanup of that resource.
//
// This method synchronizes with the UploadLogs method of the grpcClient. It
// will wait for any active calls to that method to complete unimpeded, or it
// will cancel any active calls if ctx expires. If ctx expires, the context
// error will be forwarded as the returned error. All grpcClient held resources
// will still be released in this situation.
//
// If the grpcClient has already stopped, an error will be returned describing
// this.
func (c *grpcClient) Stop(ctx context.Context) error {
	// Make sure to return context error if the context is done when calling this method.
	err := ctx.Err()

	// Acquire the c.tscMu lock within the ctx lifetime.
	acquired := make(chan struct{})
	go func() {
		c.tscMu.Lock()
		close(acquired)
	}()

	select {
	case <-ctx.Done():
		// The Stop timeout is reached. Kill any remaining exports to force
		// the clear of the lock and save the timeout error to return and
		// signal the shutdown timed out before cleanly stopping.
		c.stopFunc()
		err = ctx.Err()

		// To ensure the grpcClient is not left in a dirty state c.tsc needs to be
		// set to nil. To avoid the race condition when doing this, ensure
		// that all the exports are killed (initiated by c.stopFunc).
		<-acquired
	case <-acquired:
	}
	// Hold the tscMu lock for the rest of the function to ensure no new
	// exports are started.
	defer c.tscMu.Unlock()

	// The otlplogs.Client interface states this method is called only
	// once, but there is no guarantee it is called after Start. Ensure the
	// grpcClient is started before doing anything and let the called know if they
	// made a mistake.
	if c.tsc == nil {
		return errAlreadyStopped
	}

	// Clear c.tsc to signal the grpcClient is stopped.
	c.tsc = nil

	if c.ourConn {
		closeErr := c.conn.Close()
		// A context timeout error takes precedence over this error.
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}
	return err
}

var errShutdown = errors.New("the grpcClient is shutdown")

// UploadLogs sends log records.
//
// Retryable errors from the server will be handled according to any
// RetryConfig the grpcClient was created with.
func (c *grpcClient) UploadLogs(ctx context.Context, protoLogs []*logspb.ResourceLogs) error {
	// Hold a read lock to ensure a shut down initiated after this starts does
	// not abandon the export. This read lock acquire has less priority than a
	// write lock acquire (i.e. Stop), meaning if the grpcClient is shutting down
	// this will come after the shut down.
	c.tscMu.RLock()
	defer c.tscMu.RUnlock()

	if c.tsc == nil {
		return errShutdown
	}

	ctx, cancel := c.exportContext(ctx)
	defer cancel()

	return c.requestFunc(ctx, func(iCtx context.Context) error {
		resp, err := c.tsc.Export(iCtx, &collogspb.ExportLogsServiceRequest{
			ResourceLogs: protoLogs,
		})
		if resp != nil && resp.PartialSuccess != nil {
			msg := resp.PartialSuccess.GetErrorMessage()
			n := resp.PartialSuccess.GetRejectedLogRecords()
			if n != 0 || msg != "" {
				err := internal.LogRecordPartialSuccessError(n, msg)
				otel.Handle(err)
			}
		}
		// nil is converted to OK.
		if status.Code(err) == codes.OK {
			// Success.
			return nil
		}
		return err
	})
}

// exportContext returns a copy of parent with an appropriate deadline and
// cancellation function.
//
// It is the callers responsibility to cancel the returned context once its
// use is complete, via the parent or directly with the returned CancelFunc, to
// ensure all resources are correctly released.
func (c *grpcClient) exportContext(parent context.Context) (context.Context, context.CancelFunc) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if c.exportTimeout > 0 {
		ctx, cancel = context.WithTimeout(parent, c.exportTimeout)
	} else {
		ctx, cancel = context.WithCancel(parent)
	}

	if c.metadata.Len() > 0 {
		ctx = metadata.NewOutgoingContext(ctx, c.metadata)
	}

	// Unify the grpcClient stopCtx with the parent.
	go func() {
		select {
		case <-ctx.Done():
		case <-c.stopCtx.Done():
			// Cancel the export as the shutdown has timed out.
			cancel()
		}
	}()

	return ctx, cancel
}

// retryable returns if err identifies a request that can be retried and a
// duration to wait for if an explicit throttle time is included in err.
func retryable(err error) (bool, time.Duration) {
	s := status.Convert(err)
	switch s.Code() {
	case codes.Canceled,
		codes.DeadlineExceeded,
		codes.ResourceExhausted,
		codes.Aborted,
		codes.OutOfRange,
		codes.Unavailable,
		codes.DataLoss:
		return true, throttleDelay(s)
	}

	// Not a retry-able error.
	return false, 0
}

// throttleDelay returns a duration to wait for if an explicit throttle time
// is included in the response status.
func throttleDelay(s *status.Status) time.Duration {
	for _, detail := range s.Details() {
		if t, ok := detail.(*errdetails.RetryInfo); ok {
			return t.RetryDelay.AsDuration()
		}
	}
	return 0
}

// MarshalLog is the marshaling function used by the logging system to represent this Client.
func (c *grpcClient) MarshalLog() interface{} {
	return struct {
		Type     string
		Endpoint string
	}{
		Type:     "otlphttpgrpc",
		Endpoint: c.endpoint,
	}
}
