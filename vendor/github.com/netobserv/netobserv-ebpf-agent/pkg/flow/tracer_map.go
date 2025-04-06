package flow

import (
	"context"
	"maps"
	"runtime"
	"sync"
	"time"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/ebpf"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/metrics"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"

	"github.com/gavv/monotime"
	"github.com/netobserv/gopipes/pkg/node"
	ovnobserv "github.com/ovn-org/ovn-kubernetes/go-controller/observability-lib/sampledecoder"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var mtlog = logrus.WithField("component", "flow.MapTracer")

// MapTracer accesses a mapped source of flows (the eBPF PerCPU HashMap), deserializes it into
// a flow model.Record structure, and performs the accumulation of each perCPU-record into a single flow
type MapTracer struct {
	mapFetcher               mapFetcher
	evictionTimeout          time.Duration
	staleEntriesEvictTimeout time.Duration
	// manages the access to the eviction routines, avoiding two evictions happening at the same time
	evictionCond               *sync.Cond
	metrics                    *metrics.Metrics
	timeSpentinLookupAndDelete prometheus.Histogram
	s                          *ovnobserv.SampleDecoder
	udnEnabled                 bool
}

type mapFetcher interface {
	LookupAndDeleteMap(*metrics.Metrics) map[ebpf.BpfFlowId]model.BpfFlowContent
	DeleteMapsStaleEntries(timeOut time.Duration)
}

func NewMapTracer(fetcher mapFetcher, evictionTimeout, staleEntriesEvictTimeout time.Duration, m *metrics.Metrics,
	s *ovnobserv.SampleDecoder, udnEnabled bool) *MapTracer {
	return &MapTracer{
		mapFetcher:                 fetcher,
		evictionTimeout:            evictionTimeout,
		evictionCond:               sync.NewCond(&sync.Mutex{}),
		staleEntriesEvictTimeout:   staleEntriesEvictTimeout,
		metrics:                    m,
		timeSpentinLookupAndDelete: m.CreateTimeSpendInLookupAndDelete(),
		s:                          s,
		udnEnabled:                 udnEnabled,
	}
}

// Flush forces reading (and removing) all the flows from the source eBPF map
// and sending the entries to the next stage in the pipeline
func (m *MapTracer) Flush() {
	m.evictionCond.Broadcast()
}

func (m *MapTracer) TraceLoop(ctx context.Context, forceGC bool) node.StartFunc[[]*model.Record] {
	return func(out chan<- []*model.Record) {
		evictionTicker := time.NewTicker(m.evictionTimeout)
		go m.evictionSynchronization(ctx, forceGC, out)
		for {
			select {
			case <-ctx.Done():
				evictionTicker.Stop()
				mtlog.Debug("exiting trace loop due to context cancellation")
				return
			case <-evictionTicker.C:
				mtlog.Debug("triggering flow eviction on timer")
				m.Flush()
			}
		}
	}
}

// evictionSynchronization loop just waits for the evictionCond to happen
// and triggers the actual eviction. It makes sure that only one eviction
// is being triggered at the same time
func (m *MapTracer) evictionSynchronization(ctx context.Context, forceGC bool, out chan<- []*model.Record) {
	// flow eviction loop. It just keeps waiting for eviction until someone triggers the
	// evictionCond.Broadcast signal
	for {
		// make sure we only evict once at a time, even if there are multiple eviction signals
		m.evictionCond.L.Lock()
		m.evictionCond.Wait()
		select {
		case <-ctx.Done():
			mtlog.Debug("context canceled. Stopping goroutine before evicting flows")
			return
		default:
			mtlog.Debug("evictionSynchronization signal received")
			m.evictFlows(ctx, forceGC, out)
		}
		m.evictionCond.L.Unlock()

	}
}

func (m *MapTracer) evictFlows(ctx context.Context, forceGC bool, forwardFlows chan<- []*model.Record) {
	// it's important that this monotonic timer reports same or approximate values as kernel-side bpf_ktime_get_ns()
	monotonicTimeNow := monotime.Now()
	currentTime := time.Now()

	var forwardingFlows []*model.Record
	flows := m.mapFetcher.LookupAndDeleteMap(m.metrics)
	elapsed := time.Since(currentTime)
	udnCache := make(map[string]string)
	if m.s != nil && m.udnEnabled {
		udnsMap, err := m.s.GetInterfaceUDNs()
		if err != nil {
			mtlog.Errorf("failed to get udns to interfaces map : %v", err)
		} else {
			maps.Copy(udnCache, udnsMap)
			mtlog.Tracef("GetInterfaceUDNS map: %v", udnCache)
		}
	}
	for flowKey, flowMetrics := range flows {
		forwardingFlows = append(forwardingFlows, model.NewRecord(
			flowKey,
			&flowMetrics,
			currentTime,
			uint64(monotonicTimeNow),
			m.s,
			udnCache,
		))
	}
	m.mapFetcher.DeleteMapsStaleEntries(m.staleEntriesEvictTimeout)
	select {
	case <-ctx.Done():
		mtlog.Debug("skipping flow eviction as agent is being stopped")
	default:
		forwardFlows <- forwardingFlows
	}

	if forceGC {
		runtime.GC()
	}
	m.metrics.EvictionCounter.WithSource("hashmap").Inc()
	m.metrics.EvictedFlowsCounter.WithSource("hashmap").Add(float64(len(forwardingFlows)))
	m.timeSpentinLookupAndDelete.Observe(elapsed.Seconds())
	mtlog.Debugf("%d flows evicted", len(forwardingFlows))
}
