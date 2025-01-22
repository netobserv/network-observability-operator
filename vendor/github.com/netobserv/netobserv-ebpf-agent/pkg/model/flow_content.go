package model

import (
	"github.com/netobserv/netobserv-ebpf-agent/pkg/ebpf"
)

type BpfFlowContent struct {
	*ebpf.BpfFlowMetrics
	AdditionalMetrics *ebpf.BpfAdditionalMetrics
}

func NewBpfFlowContent(metrics ebpf.BpfFlowMetrics) BpfFlowContent {
	return BpfFlowContent{BpfFlowMetrics: &metrics}
}

func (p *BpfFlowContent) AccumulateBase(other *ebpf.BpfFlowMetrics) {
	p.BpfFlowMetrics = AccumulateBase(p.BpfFlowMetrics, other)
}

func AccumulateBase(p *ebpf.BpfFlowMetrics, other *ebpf.BpfFlowMetrics) *ebpf.BpfFlowMetrics {
	if other == nil {
		return p
	}
	if p == nil {
		return other
	}
	// time == 0 if the value has not been yet set
	if p.StartMonoTimeTs == 0 || (p.StartMonoTimeTs > other.StartMonoTimeTs && other.StartMonoTimeTs != 0) {
		p.StartMonoTimeTs = other.StartMonoTimeTs
	}
	if p.EndMonoTimeTs == 0 || p.EndMonoTimeTs < other.EndMonoTimeTs {
		p.EndMonoTimeTs = other.EndMonoTimeTs
	}
	p.Bytes += other.Bytes
	p.Packets += other.Packets
	p.Flags |= other.Flags
	if other.EthProtocol != 0 {
		p.EthProtocol = other.EthProtocol
	}
	if allZerosMac(p.SrcMac) {
		p.SrcMac = other.SrcMac
	}
	if allZerosMac(p.DstMac) {
		p.DstMac = other.DstMac
	}
	if other.Dscp != 0 {
		p.Dscp = other.Dscp
	}
	if other.Sampling != 0 {
		p.Sampling = other.Sampling
	}
	return p
}

func (p *BpfFlowContent) buildBaseFromAdditional(add *ebpf.BpfAdditionalMetrics) {
	if add == nil {
		return
	}
	// Accumulate time into base metrics if unset
	if p.BpfFlowMetrics.StartMonoTimeTs == 0 || (p.BpfFlowMetrics.StartMonoTimeTs > add.StartMonoTimeTs && add.StartMonoTimeTs != 0) {
		p.BpfFlowMetrics.StartMonoTimeTs = add.StartMonoTimeTs
	}
	if p.BpfFlowMetrics.EndMonoTimeTs == 0 || p.BpfFlowMetrics.EndMonoTimeTs < add.EndMonoTimeTs {
		p.BpfFlowMetrics.EndMonoTimeTs = add.EndMonoTimeTs
	}
	if p.BpfFlowMetrics.EthProtocol == 0 {
		p.BpfFlowMetrics.EthProtocol = add.EthProtocol
	}
	if p.BpfFlowMetrics.Flags == 0 && add.PktDrops.LatestFlags != 0 {
		p.BpfFlowMetrics.Flags = add.PktDrops.LatestFlags
	}
}

func (p *BpfFlowContent) AccumulateAdditional(other *ebpf.BpfAdditionalMetrics) {
	if other == nil {
		return
	}
	p.buildBaseFromAdditional(other)
	if p.AdditionalMetrics == nil {
		p.AdditionalMetrics = other
		return
	}
	// DNS
	p.AdditionalMetrics.DnsRecord.Flags |= other.DnsRecord.Flags
	if other.DnsRecord.Id != 0 {
		p.AdditionalMetrics.DnsRecord.Id = other.DnsRecord.Id
	}
	if p.AdditionalMetrics.DnsRecord.Errno != other.DnsRecord.Errno {
		p.AdditionalMetrics.DnsRecord.Errno = other.DnsRecord.Errno
	}
	if p.AdditionalMetrics.DnsRecord.Latency < other.DnsRecord.Latency {
		p.AdditionalMetrics.DnsRecord.Latency = other.DnsRecord.Latency
	}
	// Drop statistics
	p.AdditionalMetrics.PktDrops.Bytes += other.PktDrops.Bytes
	p.AdditionalMetrics.PktDrops.Packets += other.PktDrops.Packets
	p.AdditionalMetrics.PktDrops.LatestFlags |= other.PktDrops.LatestFlags
	if other.PktDrops.LatestDropCause != 0 {
		p.AdditionalMetrics.PktDrops.LatestDropCause = other.PktDrops.LatestDropCause
	}
	if other.PktDrops.LatestState != 0 {
		p.AdditionalMetrics.PktDrops.LatestState = other.PktDrops.LatestState
	}
	// RTT
	if p.AdditionalMetrics.FlowRtt < other.FlowRtt {
		p.AdditionalMetrics.FlowRtt = other.FlowRtt
	}
	// Network events
	for _, md := range other.NetworkEvents {
		if !AllZerosMetaData(md) && !networkEventsMDExist(p.AdditionalMetrics.NetworkEvents, md) {
			copy(p.AdditionalMetrics.NetworkEvents[p.AdditionalMetrics.NetworkEventsIdx][:], md[:])
			p.AdditionalMetrics.NetworkEventsIdx = (p.AdditionalMetrics.NetworkEventsIdx + 1) % MaxNetworkEvents
		}
	}
	// Packet Translations
	if !AllZeroIP(IP(other.TranslatedFlow.Saddr)) && !AllZeroIP(IP(other.TranslatedFlow.Daddr)) {
		p.AdditionalMetrics.TranslatedFlow = other.TranslatedFlow
	}
	// Accumulate interfaces + directions
	accumulateInterfaces(&p.AdditionalMetrics.NbObservedIntf, &p.AdditionalMetrics.ObservedIntf, other.NbObservedIntf, other.ObservedIntf)
}

func accumulateInterfaces(dstSize *uint8, dstIntf *[MaxObservedInterfaces]ebpf.BpfObservedIntfT, srcSize uint8, srcIntf [MaxObservedInterfaces]ebpf.BpfObservedIntfT) {
	iObs := uint8(0)
outer:
	for *dstSize < uint8(len(dstIntf)) && iObs < srcSize {
		for u := uint8(0); u < *dstSize; u++ {
			if dstIntf[u].Direction == srcIntf[iObs].Direction &&
				dstIntf[u].IfIndex == srcIntf[iObs].IfIndex {
				// Ignore if already exists
				iObs++
				continue outer
			}
		}
		dstIntf[*dstSize] = srcIntf[iObs]
		*dstSize++
		iObs++
	}
}

func allZerosMac(s [6]uint8) bool {
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}
