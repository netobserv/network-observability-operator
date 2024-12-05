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

package logs

import (
	"context"
	"github.com/agoda-com/opentelemetry-logs-go/sdk/internal/env"
	"go.opentelemetry.io/otel"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Defaults for BatchLogRecordProcessorOptions.
const (
	DefaultMaxQueueSize       = 2048
	DefaultScheduleDelay      = 5000
	DefaultExportTimeout      = 30000
	DefaultMaxExportBatchSize = 512
)

// BatchLogRecordProcessorOption configures a BatchLogsProcessor.
type BatchLogRecordProcessorOption func(o *BatchLogRecordProcessorOptions)

// BatchLogRecordProcessorOptions is configuration settings for a
// BatchLogsProcessor.
type BatchLogRecordProcessorOptions struct {
	// MaxQueueSize is the maximum queue size to buffer logs for delayed processing. If the
	// queue gets full it drops the logs. Use BlockOnQueueFull to change this behavior.
	// The default value of MaxQueueSize is 2048.
	MaxQueueSize int

	// BatchTimeout is the maximum duration for constructing a batch. Processor
	// forcefully sends available logs when timeout is reached.
	// The default value of BatchTimeout is 5000 msec.
	BatchTimeout time.Duration

	// ExportTimeout specifies the maximum duration for exporting logs. If the timeout
	// is reached, the export will be cancelled.
	// The default value of ExportTimeout is 30000 msec.
	ExportTimeout time.Duration

	// MaxExportBatchSize is the maximum number of logs to process in a single batch.
	// If there are more than one batch worth of logs then it processes multiple batches
	// of logs one batch after the other without any delay.
	// The default value of MaxExportBatchSize is 512.
	MaxExportBatchSize int

	// BlockOnQueueFull blocks onEnd() and onStart() method if the queue is full
	// AND if BlockOnQueueFull is set to true.
	// Blocking option should be used carefully as it can severely affect the performance of an
	// application.
	BlockOnQueueFull bool
}

// WithMaxQueueSize returns a BatchLogRecordProcessorOption that configures the
// maximum queue size allowed for a BatchLogRecordProcessor.
func WithMaxQueueSize(size int) BatchLogRecordProcessorOption {
	return func(o *BatchLogRecordProcessorOptions) {
		o.MaxQueueSize = size
	}
}

// WithMaxExportBatchSize returns a BatchLogRecordProcessorOption that configures
// the maximum export batch size allowed for a BatchLogRecordProcessor.
func WithMaxExportBatchSize(size int) BatchLogRecordProcessorOption {
	return func(o *BatchLogRecordProcessorOptions) {
		o.MaxExportBatchSize = size
	}
}

// WithBatchTimeout returns a BatchLogRecordProcessorOption that configures the
// maximum delay allowed for a BatchLogRecordProcessor before it will export any
// held log (whether the queue is full or not).
func WithBatchTimeout(delay time.Duration) BatchLogRecordProcessorOption {
	return func(o *BatchLogRecordProcessorOptions) {
		o.BatchTimeout = delay
	}
}

// WithExportTimeout returns a BatchLogRecordProcessorOption that configures the
// amount of time a BatchLogRecordProcessor waits for an exporter to export before
// abandoning the export.
func WithExportTimeout(timeout time.Duration) BatchLogRecordProcessorOption {
	return func(o *BatchLogRecordProcessorOptions) {
		o.ExportTimeout = timeout
	}
}

// WithBlocking returns a BatchLogRecordProcessorOption that configures a
// BatchLogRecordProcessor to wait for enqueue operations to succeed instead of
// dropping data when the queue is full.
func WithBlocking() BatchLogRecordProcessorOption {
	return func(o *BatchLogRecordProcessorOptions) {
		o.BlockOnQueueFull = true
	}
}

// batchLogRecordProcessor is a LogRecordProcessor that batches asynchronously-received
// logs and sends them to a logs.Exporter when complete.
type batchLogRecordProcessor struct {
	e LogRecordExporter
	o BatchLogRecordProcessorOptions

	queue   chan ReadableLogRecord
	dropped uint32

	batch      []ReadableLogRecord
	batchMutex sync.Mutex
	timer      *time.Timer
	stopWait   sync.WaitGroup
	stopOnce   sync.Once
	stopCh     chan struct{}
	stopped    atomic.Bool
}

func (lrp *batchLogRecordProcessor) Shutdown(ctx context.Context) error {
	var err error
	lrp.stopOnce.Do(func() {
		lrp.stopped.Store(true)
		wait := make(chan struct{})
		go func() {
			close(lrp.stopCh)
			lrp.stopWait.Wait()
			if lrp.e != nil {
				if err := lrp.e.Shutdown(ctx); err != nil {
					otel.Handle(err)
				}
			}
			close(wait)
		}()
		// Wait until the wait group is done or the context is cancelled
		select {
		case <-wait:
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return err
}

var _ LogRecordProcessor = (*batchLogRecordProcessor)(nil)

// NewBatchLogRecordProcessor creates a new LogRecordProcessor that will send completed
// log batches to the exporter with the supplied options.
//
// If the exporter is nil, the logs processor will perform no action.
// see https://opentelemetry.io/docs/specs/otel/logs/sdk/#batching-processor
func NewBatchLogRecordProcessor(exporter LogRecordExporter, options ...BatchLogRecordProcessorOption) LogRecordProcessor {
	maxQueueSize := env.BatchLogsProcessorMaxQueueSize(DefaultMaxQueueSize)
	maxExportBatchSize := env.BatchLogsProcessorMaxExportBatchSize(DefaultMaxExportBatchSize)

	if maxExportBatchSize > maxQueueSize {
		if DefaultMaxExportBatchSize > maxQueueSize {
			maxExportBatchSize = maxQueueSize
		} else {
			maxExportBatchSize = DefaultMaxExportBatchSize
		}
	}

	o := BatchLogRecordProcessorOptions{
		BatchTimeout:       time.Duration(env.BatchLogsProcessorScheduleDelay(DefaultScheduleDelay)) * time.Millisecond,
		ExportTimeout:      time.Duration(env.BatchLogsProcessorExportTimeout(DefaultExportTimeout)) * time.Millisecond,
		MaxQueueSize:       maxQueueSize,
		MaxExportBatchSize: maxExportBatchSize,
	}
	for _, opt := range options {
		opt(&o)
	}
	blp := &batchLogRecordProcessor{
		e:      exporter,
		o:      o,
		batch:  make([]ReadableLogRecord, 0, o.MaxExportBatchSize),
		timer:  time.NewTimer(o.BatchTimeout),
		queue:  make(chan ReadableLogRecord, o.MaxQueueSize),
		stopCh: make(chan struct{}),
	}

	blp.stopWait.Add(1)
	go func() {
		defer blp.stopWait.Done()
		blp.processQueue()
		blp.drainQueue()
	}()

	return blp
}

func (lrp *batchLogRecordProcessor) OnEmit(rol ReadableLogRecord) {

	// Do not enqueue spans after Shutdown.
	if lrp.stopped.Load() {
		return
	}
	// Do not enqueue logs if we are just going to drop them.
	if lrp.e == nil {
		return
	}

	lrp.enqueue(rol)
}

type forceFlushLogs struct {
	ReadableLogRecord
	flushed chan struct{}
}

// processQueue removes logs from the `queue` channel until processor
// is shut down. It calls the exporter in batches of up to MaxExportBatchSize
// waiting up to BatchTimeout to form a batch.
func (lrp *batchLogRecordProcessor) processQueue() {
	defer lrp.timer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case <-lrp.stopCh:
			return
		case <-lrp.timer.C:
			if err := lrp.exportLogs(ctx); err != nil {
				otel.Handle(err)
			}
		case sd := <-lrp.queue:
			if ffs, ok := sd.(forceFlushLogs); ok {
				close(ffs.flushed)
				continue
			}
			lrp.batchMutex.Lock()
			lrp.batch = append(lrp.batch, sd)
			shouldExport := len(lrp.batch) >= lrp.o.MaxExportBatchSize
			lrp.batchMutex.Unlock()
			if shouldExport {
				if !lrp.timer.Stop() {
					<-lrp.timer.C
				}
				if err := lrp.exportLogs(ctx); err != nil {
					otel.Handle(err)
				}
			}
		}
	}
}

// drainQueue awaits the any caller that had added to bsp.stopWait
// to finish the enqueue, then exports the final batch.
func (lrp *batchLogRecordProcessor) drainQueue() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case sd := <-lrp.queue:
			if sd == nil {
				if err := lrp.exportLogs(ctx); err != nil {
					otel.Handle(err)
				}
				return
			}

			lrp.batchMutex.Lock()
			lrp.batch = append(lrp.batch, sd)
			shouldExport := len(lrp.batch) == lrp.o.MaxExportBatchSize
			lrp.batchMutex.Unlock()

			if shouldExport {
				if err := lrp.exportLogs(ctx); err != nil {
					otel.Handle(err)
				}
			}
		default:
			close(lrp.queue)
		}
	}
}

// exportLogs is a subroutine of processing and draining the queue.
func (lrp *batchLogRecordProcessor) exportLogs(ctx context.Context) error {
	lrp.timer.Reset(lrp.o.BatchTimeout)

	lrp.batchMutex.Lock()
	defer lrp.batchMutex.Unlock()

	if lrp.o.ExportTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, lrp.o.ExportTimeout)
		defer cancel()
	}

	if l := len(lrp.batch); l > 0 {
		//global.Debug("exporting logs", "count", len(lrp.batch), "total_dropped", atomic.LoadUint32(&lrp.dropped))
		err := lrp.e.Export(ctx, lrp.batch)

		// A new batch is always created after exporting, even if the batch failed to be exported.
		//
		// It is up to the exporter to implement any type of retry logic if a batch is failing
		// to be exported, since it is specific to the protocol and backend being sent to.
		lrp.batch = lrp.batch[:0]

		if err != nil {
			return err
		}
	}
	return nil
}

func (lrp *batchLogRecordProcessor) enqueue(sd ReadableLogRecord) {
	ctx := context.TODO()
	if lrp.o.BlockOnQueueFull {
		lrp.enqueueBlockOnQueueFull(ctx, sd)
	} else {
		lrp.enqueueDrop(ctx, sd)
	}
}

// ForceFlush exports all ended logs that have not yet been exported.
func (lrp *batchLogRecordProcessor) ForceFlush(ctx context.Context) error {

	// Interrupt if context is already canceled.
	if err := ctx.Err(); err != nil {
		return err
	}
	// Do nothing after Shutdown.
	// Do not enqueue spans after Shutdown.
	if lrp.stopped.Load() {
		return nil
	}

	var err error
	if lrp.e != nil {
		flushCh := make(chan struct{})
		if lrp.enqueueBlockOnQueueFull(ctx, forceFlushLogs{flushed: flushCh}) {
			select {
			case <-flushCh:
				// Processed any items in queue prior to ForceFlush being called
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		wait := make(chan error)
		go func() {
			wait <- lrp.exportLogs(ctx)
			close(wait)
		}()
		// Wait until the export is finished or the context is cancelled/timed out
		select {
		case err = <-wait:
		case <-ctx.Done():
			err = ctx.Err()
		}
	}
	return err
}

func recoverSendOnClosedChan() {
	x := recover()
	switch err := x.(type) {
	case nil:
		return
	case runtime.Error:
		if err.Error() == "send on closed channel" {
			return
		}
	}
	panic(x)
}

func (lrp *batchLogRecordProcessor) enqueueBlockOnQueueFull(ctx context.Context, sd ReadableLogRecord) bool {

	// This ensures the bsp.queue<- below does not panic as the
	// processor shuts down.
	defer recoverSendOnClosedChan()

	select {
	case <-lrp.stopCh:
		return false
	default:
	}

	select {
	case lrp.queue <- sd:
		return true
	case <-ctx.Done():
		return false
	}
}

func (lrp *batchLogRecordProcessor) enqueueDrop(ctx context.Context, ld ReadableLogRecord) bool {

	// This ensures the bsp.queue<- below does not panic as the
	// processor shuts down.
	defer recoverSendOnClosedChan()

	select {
	case <-lrp.stopCh:
		return false
	default:
	}

	select {
	case lrp.queue <- ld:
		return true
	default:
		atomic.AddUint32(&lrp.dropped, 1)
	}
	return false
}

// MarshalLog is the marshaling function used by the logging system to represent this exporter.
func (lrp *batchLogRecordProcessor) MarshalLog() interface{} {
	return struct {
		Type              string
		LogRecordExporter LogRecordExporter
		Config            BatchLogRecordProcessorOptions
	}{
		Type:              "BatchLogRecordProcessor",
		LogRecordExporter: lrp.e,
		Config:            lrp.o,
	}
}
