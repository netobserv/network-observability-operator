package packets

import (
	"encoding/base64"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

func PacketToMap(pr *model.PacketRecord) config.GenericMap {
	out := config.GenericMap{}

	if pr == nil {
		return out
	}

	packet := gopacket.NewPacket(pr.Stream, layers.LayerTypeEthernet, gopacket.Lazy)
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth, _ := ethLayer.(*layers.Ethernet)
		out["SrcMac"] = eth.SrcMAC.String()
		out["DstMac"] = eth.DstMAC.String()
	}

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		out["SrcPort"] = tcp.SrcPort.String()
		out["DstPort"] = tcp.DstPort.String()
	} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		out["SrcPort"] = udp.SrcPort.String()
		out["DstPort"] = udp.DstPort.String()
	} else if sctpLayer := packet.Layer(layers.LayerTypeSCTP); sctpLayer != nil {
		sctp, _ := sctpLayer.(*layers.SCTP)
		out["SrcPort"] = sctp.SrcPort.String()
		out["DstPort"] = sctp.DstPort.String()
	}

	if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
		ipv4, _ := ipv4Layer.(*layers.IPv4)
		out["SrcAddr"] = ipv4.SrcIP.String()
		out["DstAddr"] = ipv4.DstIP.String()
		out["Proto"] = ipv4.Protocol
	} else if ipv6Layer := packet.Layer(layers.LayerTypeIPv6); ipv6Layer != nil {
		ipv6, _ := ipv6Layer.(*layers.IPv6)
		out["SrcAddr"] = ipv6.SrcIP.String()
		out["DstAddr"] = ipv6.DstIP.String()
		out["Proto"] = ipv6.NextHeader
	}

	if icmpv4Layer := packet.Layer(layers.LayerTypeICMPv4); icmpv4Layer != nil {
		icmpv4, _ := icmpv4Layer.(*layers.ICMPv4)
		out["IcmpType"] = icmpv4.TypeCode.Type()
		out["IcmpCode"] = icmpv4.TypeCode.Code()
	} else if icmpv6Layer := packet.Layer(layers.LayerTypeICMPv6); icmpv6Layer != nil {
		icmpv6, _ := icmpv6Layer.(*layers.ICMPv6)
		out["IcmpType"] = icmpv6.TypeCode.Type()
		out["IcmpCode"] = icmpv6.TypeCode.Code()
	}

	if dnsLayer := packet.Layer(layers.LayerTypeDNS); dnsLayer != nil {
		dns, _ := dnsLayer.(*layers.DNS)
		out["DnsId"] = dns.ID
		out["DnsFlagsResponseCode"] = dns.ResponseCode.String()
		//TODO: add DNS questions / answers / authorities
	}

	out["Bytes"] = len(pr.Stream)
	// Data is base64 encoded to avoid marshal / unmarshal issues
	out["Data"] = base64.StdEncoding.EncodeToString(packet.Data())
	out["Time"] = pr.Time.Unix()

	return out
}
