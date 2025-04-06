package tracer

import (
	"github.com/netobserv/netobserv-ebpf-agent/pkg/ebpf"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/metrics"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
)

// This file contains legacy implementations kept for old kernels

func (m *FlowFetcher) legacyLookupAndDeleteMap(met *metrics.Metrics) map[ebpf.BpfFlowId]model.BpfFlowContent {
	flowMap := m.objects.AggregatedFlows

	iterator := flowMap.Iterate()
	var flows = make(map[ebpf.BpfFlowId]model.BpfFlowContent, m.cacheMaxSize)
	var id ebpf.BpfFlowId
	var baseMetrics ebpf.BpfFlowMetrics
	count := 0

	// Deleting while iterating is really bad for performance (like, really!) as it causes seeing multiple times the same key
	// This is solved in >=4.20 kernels with LookupAndDelete
	for iterator.Next(&id, &baseMetrics) {
		count++
		if err := flowMap.Delete(id); err != nil {
			log.WithError(err).WithField("flowId", id).Warnf("couldn't delete flow entry")
			met.Errors.WithErrorName("flow-fetcher-legacy", "CannotDeleteFlows", metrics.HighSeverity).Inc()
		}
		flows[id] = model.NewBpfFlowContent(baseMetrics)
	}
	met.BufferSizeGauge.WithBufferName("hashmap-legacy-total").Set(float64(count))
	met.BufferSizeGauge.WithBufferName("hashmap-legacy-unique").Set(float64(len(flows)))

	m.ReadGlobalCounter(met)
	return flows
}

func (p *PacketFetcher) legacyLookupAndDeleteMap(met *metrics.Metrics) map[int][]*byte {
	packetMap := p.objects.PacketRecord
	iterator := packetMap.Iterate()
	packets := make(map[int][]*byte, p.cacheMaxSize)

	var id int
	var packet []*byte
	for iterator.Next(&id, &packet) {
		if err := packetMap.Delete(id); err != nil {
			log.WithError(err).WithField("packetID ", id).Warnf("couldn't delete  entry")
			met.Errors.WithErrorName("pkt-fetcher-legacy", "CannotDeleteEntry", metrics.HighSeverity).Inc()
		}
		packets[id] = append(packets[id], packet...)
	}
	return packets
}
