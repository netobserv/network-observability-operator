package model

import "github.com/netobserv/netobserv-ebpf-agent/pkg/ebpf"

type BpfFlowContent struct {
	*ebpf.BpfFlowMetrics
	AdditionalMetrics *ebpf.BpfAdditionalMetrics
}

type BpfFlowContents []BpfFlowContent

func (a *BpfFlowContents) Accumulate() BpfFlowContent {
	res := BpfFlowContent{}
	for _, p := range *a {
		res.AccumulateBase(p.BpfFlowMetrics)
		res.AccumulateAdditional(p.AdditionalMetrics)
	}
	return res
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

func (p *BpfFlowContent) AccumulateAdditional(other *ebpf.BpfAdditionalMetrics) {
	if other == nil {
		return
	}
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
}

func allZerosMac(s [6]uint8) bool {
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}
