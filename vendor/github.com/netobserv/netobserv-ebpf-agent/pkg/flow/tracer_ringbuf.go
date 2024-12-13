package flow

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/metrics"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"

	"github.com/cilium/ebpf/ringbuf"
	"github.com/netobserv/gopipes/pkg/node"
	"github.com/sirupsen/logrus"
)

var rtlog = logrus.WithField("component", "flow.RingBufTracer")

// RingBufTracer receives single-packet flows via ringbuffer (usually, these that couldn't be
// added in the eBPF kernel space due to the map being full or busy) and submits them to the
// userspace Aggregator map
type RingBufTracer struct {
	mapFlusher mapFlusher
	ringBuffer ringBufReader
	stats      stats
	metrics    *metrics.Metrics
}

type ringBufReader interface {
	ReadRingBuf() (ringbuf.Record, error)
}

// stats supports atomic logging of ringBuffer metrics
type stats struct {
	loggingTimeout time.Duration
	isForwarding   int32
	forwardedFlows int32
	mapFullErrs    int32
}

type mapFlusher interface {
	Flush()
}

func NewRingBufTracer(reader ringBufReader, flusher mapFlusher, logTimeout time.Duration, m *metrics.Metrics) *RingBufTracer {
	return &RingBufTracer{
		mapFlusher: flusher,
		ringBuffer: reader,
		stats:      stats{loggingTimeout: logTimeout},
		metrics:    m,
	}
}

func (m *RingBufTracer) TraceLoop(ctx context.Context) node.StartFunc[*model.RawRecord] {
	return func(out chan<- *model.RawRecord) {
		debugging := logrus.IsLevelEnabled(logrus.DebugLevel)
		for {
			select {
			case <-ctx.Done():
				rtlog.Debug("exiting trace loop due to context cancellation")
				return
			default:
				if err := m.listenAndForwardRingBuffer(debugging, out); err != nil {
					if errors.Is(err, ringbuf.ErrClosed) {
						rtlog.Debug("Received signal, exiting..")
						return
					}
					rtlog.WithError(err).Warn("ignoring flow event")
					continue
				}
			}
		}
	}
}

func (m *RingBufTracer) listenAndForwardRingBuffer(debugging bool, forwardCh chan<- *model.RawRecord) error {
	event, err := m.ringBuffer.ReadRingBuf()
	if err != nil {
		m.metrics.Errors.WithErrorName("ringbuffer", "CannotReadRingbuffer").Inc()
		return fmt.Errorf("reading from ring buffer: %w", err)
	}
	// Parses the ringbuf event entry into an Event structure.
	readFlow, err := model.ReadFrom(bytes.NewBuffer(event.RawSample))
	if err != nil {
		m.metrics.Errors.WithErrorName("ringbuffer", "CannotParseRingbuffer").Inc()
		return fmt.Errorf("parsing data received from the ring buffer: %w", err)
	}
	mapFullError := readFlow.Metrics.Errno == uint8(syscall.E2BIG)
	if debugging {
		m.stats.logRingBufferFlows(mapFullError)
	}
	errno := syscall.Errno(readFlow.Metrics.Errno)
	// In ringbuffer, a "flow" is a 1-packet flow, it hasn't gone through aggregation yet. So we use the packet counter metric.
	m.metrics.EvictedPacketsCounter.WithSourceAndReason("ringbuffer", errno.Error()).Inc()
	// Will need to send it to accounter anyway to account regardless of complete/ongoing flow
	forwardCh <- readFlow
	return nil
}

// logRingBufferFlows avoids flooding logs on long series of evicted flows by grouping how
// many flows are forwarded
func (m *stats) logRingBufferFlows(mapFullErr bool) {
	atomic.AddInt32(&m.forwardedFlows, 1)
	if mapFullErr {
		atomic.AddInt32(&m.mapFullErrs, 1)
	}
	if atomic.CompareAndSwapInt32(&m.isForwarding, 0, 1) {
		go func() {
			time.Sleep(m.loggingTimeout)
			mfe := atomic.LoadInt32(&m.mapFullErrs)
			l := rtlog.WithFields(logrus.Fields{
				"flows":       atomic.LoadInt32(&m.forwardedFlows),
				"mapFullErrs": mfe,
			})
			if mfe == 0 {
				l.Debug("received flows via ringbuffer")
			} else {
				l.Debug("received flows via ringbuffer. You might want to increase the CACHE_MAX_FLOWS value")
			}
			atomic.StoreInt32(&m.forwardedFlows, 0)
			atomic.StoreInt32(&m.isForwarding, 0)
			atomic.StoreInt32(&m.mapFullErrs, 0)
		}()
	}
}
