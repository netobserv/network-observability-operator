package pbflow

import (
	"encoding/binary"
	"net"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/ebpf"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var protoLog = logrus.WithField("component", "pbflow")

// FlowsToPB is an auxiliary function to convert flow records, as returned by the eBPF agent,
// into protobuf-encoded messages ready to be sent to the collector via GRPC
func FlowsToPB(inputRecords []*model.Record, maxLen int) []*Records {
	entries := make([]*Record, 0, len(inputRecords))
	for _, record := range inputRecords {
		entries = append(entries, FlowToPB(record))
	}
	var records []*Records
	for len(entries) > 0 {
		end := len(entries)
		if end > maxLen {
			end = maxLen
		}
		records = append(records, &Records{Entries: entries[:end]})
		entries = entries[end:]
	}
	return records
}

// FlowToPB is an auxiliary function to convert a single flow record, as returned by the eBPF agent,
// into a protobuf-encoded message ready to be sent to the collector via kafka
func FlowToPB(fr *model.Record) *Record {
	var pbflowRecord = Record{
		EthProtocol: uint32(fr.Metrics.EthProtocol),
		Direction:   Direction(fr.Metrics.DirectionFirstSeen),
		DataLink: &DataLink{
			SrcMac: macToUint64(&fr.Metrics.SrcMac),
			DstMac: macToUint64(&fr.Metrics.DstMac),
		},
		Network: &Network{
			Dscp: uint32(fr.Metrics.Dscp),
		},
		Transport: &Transport{
			Protocol: uint32(fr.ID.TransportProtocol),
			SrcPort:  uint32(fr.ID.SrcPort),
			DstPort:  uint32(fr.ID.DstPort),
		},
		IcmpType: uint32(fr.ID.IcmpType),
		IcmpCode: uint32(fr.ID.IcmpCode),
		Bytes:    fr.Metrics.Bytes,
		TimeFlowStart: &timestamppb.Timestamp{
			Seconds: fr.TimeFlowStart.Unix(),
			Nanos:   int32(fr.TimeFlowStart.Nanosecond()),
		},
		TimeFlowEnd: &timestamppb.Timestamp{
			Seconds: fr.TimeFlowEnd.Unix(),
			Nanos:   int32(fr.TimeFlowEnd.Nanosecond()),
		},
		Packets:     uint64(fr.Metrics.Packets),
		AgentIp:     agentIP(fr.AgentIP),
		Flags:       uint32(fr.Metrics.Flags),
		TimeFlowRtt: durationpb.New(fr.TimeFlowRtt),
		Sampling:    fr.Metrics.Sampling,
	}
	if fr.Metrics.AdditionalMetrics != nil {
		pbflowRecord.PktDropBytes = fr.Metrics.AdditionalMetrics.PktDrops.Bytes
		pbflowRecord.PktDropPackets = uint64(fr.Metrics.AdditionalMetrics.PktDrops.Packets)
		pbflowRecord.PktDropLatestFlags = uint32(fr.Metrics.AdditionalMetrics.PktDrops.LatestFlags)
		pbflowRecord.PktDropLatestState = uint32(fr.Metrics.AdditionalMetrics.PktDrops.LatestState)
		pbflowRecord.PktDropLatestDropCause = fr.Metrics.AdditionalMetrics.PktDrops.LatestDropCause
		pbflowRecord.DnsId = uint32(fr.Metrics.AdditionalMetrics.DnsRecord.Id)
		pbflowRecord.DnsFlags = uint32(fr.Metrics.AdditionalMetrics.DnsRecord.Flags)
		pbflowRecord.DnsErrno = uint32(fr.Metrics.AdditionalMetrics.DnsRecord.Errno)
		if fr.Metrics.AdditionalMetrics.DnsRecord.Latency != 0 {
			pbflowRecord.DnsLatency = durationpb.New(fr.DNSLatency)
		}
		pbflowRecord.Xlat = &Xlat{
			SrcPort: uint32(fr.Metrics.AdditionalMetrics.TranslatedFlow.Sport),
			DstPort: uint32(fr.Metrics.AdditionalMetrics.TranslatedFlow.Dport),
			ZoneId:  uint32(fr.Metrics.AdditionalMetrics.TranslatedFlow.ZoneId),
		}
	}
	pbflowRecord.DupList = make([]*DupMapEntry, 0)
	for _, intf := range fr.Interfaces {
		pbflowRecord.DupList = append(pbflowRecord.DupList, &DupMapEntry{
			Interface: intf.Interface,
			Direction: Direction(intf.Direction),
			Udn:       intf.Udn,
		})
	}
	if fr.Metrics.EthProtocol == model.IPv6Type {
		pbflowRecord.Network.SrcAddr = &IP{IpFamily: &IP_Ipv6{Ipv6: fr.ID.SrcIp[:]}}
		pbflowRecord.Network.DstAddr = &IP{IpFamily: &IP_Ipv6{Ipv6: fr.ID.DstIp[:]}}
		if fr.Metrics.AdditionalMetrics != nil {
			pbflowRecord.Xlat.SrcAddr = &IP{IpFamily: &IP_Ipv6{Ipv6: fr.Metrics.AdditionalMetrics.TranslatedFlow.Saddr[:]}}
			pbflowRecord.Xlat.DstAddr = &IP{IpFamily: &IP_Ipv6{Ipv6: fr.Metrics.AdditionalMetrics.TranslatedFlow.Daddr[:]}}
		}
	} else {
		pbflowRecord.Network.SrcAddr = &IP{IpFamily: &IP_Ipv4{Ipv4: model.IntEncodeV4(fr.ID.SrcIp)}}
		pbflowRecord.Network.DstAddr = &IP{IpFamily: &IP_Ipv4{Ipv4: model.IntEncodeV4(fr.ID.DstIp)}}
		if fr.Metrics.AdditionalMetrics != nil {
			pbflowRecord.Xlat.SrcAddr = &IP{IpFamily: &IP_Ipv4{Ipv4: model.IntEncodeV4(fr.Metrics.AdditionalMetrics.TranslatedFlow.Saddr)}}
			pbflowRecord.Xlat.DstAddr = &IP{IpFamily: &IP_Ipv4{Ipv4: model.IntEncodeV4(fr.Metrics.AdditionalMetrics.TranslatedFlow.Daddr)}}
		}
	}

	if len(fr.NetworkMonitorEventsMD) != 0 {
		pbflowRecord.NetworkEventsMetadata = make([]*NetworkEvent, 0)
		for _, networkEvent := range fr.NetworkMonitorEventsMD {
			pbflowRecord.NetworkEventsMetadata = append(pbflowRecord.NetworkEventsMetadata, &NetworkEvent{
				Events: networkEvent,
			})
		}
	}
	return &pbflowRecord
}

func PBToFlow(pb *Record) *model.Record {
	if pb == nil {
		return nil
	}
	out := model.Record{
		ID: ebpf.BpfFlowId{
			TransportProtocol: uint8(pb.Transport.Protocol),
			SrcIp:             ipToIPAddr(pb.Network.GetSrcAddr()),
			DstIp:             ipToIPAddr(pb.Network.GetDstAddr()),
			SrcPort:           uint16(pb.Transport.SrcPort),
			DstPort:           uint16(pb.Transport.DstPort),
			IcmpType:          uint8(pb.IcmpType),
			IcmpCode:          uint8(pb.IcmpCode),
		},
		Metrics: model.BpfFlowContent{
			BpfFlowMetrics: &ebpf.BpfFlowMetrics{
				EthProtocol: uint16(pb.EthProtocol),
				SrcMac:      macToUint8(pb.DataLink.GetSrcMac()),
				DstMac:      macToUint8(pb.DataLink.GetDstMac()),
				Bytes:       pb.Bytes,
				Packets:     uint32(pb.Packets),
				Flags:       uint16(pb.Flags),
				Dscp:        uint8(pb.Network.Dscp),
				Sampling:    pb.Sampling,
			},
			AdditionalMetrics: &ebpf.BpfAdditionalMetrics{
				PktDrops: ebpf.BpfPktDropsT{
					Bytes:           pb.PktDropBytes,
					Packets:         uint32(pb.PktDropPackets),
					LatestFlags:     uint16(pb.PktDropLatestFlags),
					LatestState:     uint8(pb.PktDropLatestState),
					LatestDropCause: pb.PktDropLatestDropCause,
				},
				DnsRecord: ebpf.BpfDnsRecordT{
					Id:      uint16(pb.DnsId),
					Flags:   uint16(pb.DnsFlags),
					Errno:   uint8(pb.DnsErrno),
					Latency: uint64(pb.DnsLatency.AsDuration()),
				},
				TranslatedFlow: ebpf.BpfTranslatedFlowT{
					Saddr:  ipToIPAddr(pb.Xlat.GetSrcAddr()),
					Daddr:  ipToIPAddr(pb.Xlat.GetDstAddr()),
					Sport:  uint16(pb.Xlat.GetSrcPort()),
					Dport:  uint16(pb.Xlat.GetDstPort()),
					ZoneId: uint16(pb.Xlat.GetZoneId()),
				},
			},
		},
		TimeFlowStart: pb.TimeFlowStart.AsTime(),
		TimeFlowEnd:   pb.TimeFlowEnd.AsTime(),
		AgentIP:       pbIPToNetIP(pb.AgentIp),
		TimeFlowRtt:   pb.TimeFlowRtt.AsDuration(),
		DNSLatency:    pb.DnsLatency.AsDuration(),
	}

	if len(pb.GetDupList()) != 0 {
		for _, entry := range pb.GetDupList() {
			out.Interfaces = append(out.Interfaces, model.IntfDirUdn{
				Interface: entry.Interface,
				Direction: int(entry.Direction),
				Udn:       entry.Udn,
			})
		}
	}

	if len(pb.GetNetworkEventsMetadata()) != 0 {
		for _, e := range pb.GetNetworkEventsMetadata() {
			m := map[string]string{}
			for k, v := range e.Events {
				m[k] = v
			}
			out.NetworkMonitorEventsMD = append(out.NetworkMonitorEventsMD, m)
		}
		protoLog.Tracef("decoded Network events monitor metadata: %v", out.NetworkMonitorEventsMD)
	}

	return &out
}

// Mac bytes are encoded in the same order as in the array. This is, a Mac
// like 11:22:33:44:55:66 will be encoded as 0x112233445566
func macToUint64(m *[model.MacLen]uint8) uint64 {
	return uint64(m[5]) |
		(uint64(m[4]) << 8) |
		(uint64(m[3]) << 16) |
		(uint64(m[2]) << 24) |
		(uint64(m[1]) << 32) |
		(uint64(m[0]) << 40)
}

func agentIP(nip net.IP) *IP {
	if ip := nip.To4(); ip != nil {
		return &IP{IpFamily: &IP_Ipv4{Ipv4: binary.BigEndian.Uint32(ip)}}
	}
	// IPv6 address
	return &IP{IpFamily: &IP_Ipv6{Ipv6: nip}}
}

func pbIPToNetIP(ip *IP) net.IP {
	if ip.GetIpv6() != nil {
		return net.IP(ip.GetIpv6())
	}
	n := ip.GetIpv4()
	return net.IPv4(
		byte((n>>24)&0xFF),
		byte((n>>16)&0xFF),
		byte((n>>8)&0xFF),
		byte(n&0xFF))
}

func ipToIPAddr(ip *IP) model.IPAddr {
	return model.IPAddrFromNetIP(pbIPToNetIP(ip))
}

func macToUint8(mac uint64) [6]uint8 {
	return [6]uint8{
		uint8(mac >> 40),
		uint8(mac >> 32),
		uint8(mac >> 24),
		uint8(mac >> 16),
		uint8(mac >> 8),
		uint8(mac),
	}
}
