package model

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"reflect"
	"time"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/ebpf"

	ovnmodel "github.com/ovn-org/ovn-kubernetes/go-controller/observability-lib/model"
	ovnobserv "github.com/ovn-org/ovn-kubernetes/go-controller/observability-lib/sampledecoder"
)

// Values according to field 61 in https://www.iana.org/assignments/ipfix/ipfix.xhtml
const (
	DirectionIngress = 0
	DirectionEgress  = 1
	MacLen           = 6
	// IPv4Type / IPv6Type value as defined in IEEE 802: https://www.iana.org/assignments/ieee-802-numbers/ieee-802-numbers.xhtml
	IPv6Type                 = 0x86DD
	NetworkEventsMaxEventsMD = 8
	MaxNetworkEvents         = 4
	MaxObservedInterfaces    = 4
)

type HumanBytes uint64
type MacAddr [MacLen]uint8
type Direction uint8

// IPAddr encodes v4 and v6 IPs with a fixed length.
// IPv4 addresses are encoded as IPv6 addresses with prefix ::ffff/96
// as described in https://datatracker.ietf.org/doc/html/rfc4038#section-4.2
// (same behavior as Go's net.IP type)
type IPAddr [net.IPv6len]uint8

type InterfaceNamer func(ifIndex int) string

var (
	agentIP        net.IP
	interfaceNamer InterfaceNamer = func(ifIndex int) string { return fmt.Sprintf("[namer unset] %d", ifIndex) }
)

func SetGlobals(ip net.IP, ifaceNamer InterfaceNamer) {
	agentIP = ip
	interfaceNamer = ifaceNamer
}

// record structure as parsed from eBPF
type RawRecord ebpf.BpfFlowRecordT

// Record contains accumulated metrics from a flow
type Record struct {
	ID      ebpf.BpfFlowId
	Metrics BpfFlowContent

	// TODO: redundant field from RecordMetrics. Reorganize structs
	TimeFlowStart time.Time
	TimeFlowEnd   time.Time
	DNSLatency    time.Duration
	Interfaces    []IntfDirUdn
	// AgentIP provides information about the source of the flow (the Agent that traced it)
	AgentIP net.IP
	// Calculated RTT which is set when record is created by calling NewRecord
	TimeFlowRtt            time.Duration
	NetworkMonitorEventsMD []map[string]string
}

var udnsCache map[string]string

func NewRecord(
	key ebpf.BpfFlowId,
	metrics *BpfFlowContent,
	currentTime time.Time,
	monotonicCurrentTime uint64,
	s *ovnobserv.SampleDecoder,
) *Record {
	udnsCache = make(map[string]string)
	startDelta := time.Duration(monotonicCurrentTime - metrics.StartMonoTimeTs)
	endDelta := time.Duration(monotonicCurrentTime - metrics.EndMonoTimeTs)

	var record = Record{
		ID:            key,
		Metrics:       *metrics,
		TimeFlowStart: currentTime.Add(-startDelta),
		TimeFlowEnd:   currentTime.Add(-endDelta),
		AgentIP:       agentIP,
		Interfaces: []IntfDirUdn{NewIntfDirUdn(
			interfaceNamer(int(metrics.IfIndexFirstSeen)),
			int(metrics.DirectionFirstSeen),
			s)},
	}
	if metrics.AdditionalMetrics != nil {
		for i := uint8(0); i < record.Metrics.AdditionalMetrics.NbObservedIntf; i++ {
			record.Interfaces = append(record.Interfaces, NewIntfDirUdn(
				interfaceNamer(int(metrics.AdditionalMetrics.ObservedIntf[i].IfIndex)),
				int(metrics.AdditionalMetrics.ObservedIntf[i].Direction),
				s,
			))
		}
		if metrics.AdditionalMetrics.FlowRtt != 0 {
			record.TimeFlowRtt = time.Duration(metrics.AdditionalMetrics.FlowRtt)
		}
		if metrics.AdditionalMetrics.DnsRecord.Latency != 0 {
			record.DNSLatency = time.Duration(metrics.AdditionalMetrics.DnsRecord.Latency)
		}
	}
	if s != nil && metrics.AdditionalMetrics != nil {
		seen := make(map[string]bool)
		record.NetworkMonitorEventsMD = make([]map[string]string, 0)
		for _, metadata := range metrics.AdditionalMetrics.NetworkEvents {
			if !AllZerosMetaData(metadata) {
				var cm map[string]string
				if md, err := s.DecodeCookie8Bytes(metadata); err == nil {
					acl, ok := md.(*ovnmodel.ACLEvent)
					mdStr := md.String()
					if !seen[mdStr] {
						if ok {
							cm = map[string]string{
								"Action":    acl.Action,
								"Type":      acl.Actor,
								"Feature":   "acl",
								"Name":      acl.Name,
								"Namespace": acl.Namespace,
								"Direction": acl.Direction,
							}
						} else {
							cm = map[string]string{
								"Message": mdStr,
							}
						}
						record.NetworkMonitorEventsMD = append(record.NetworkMonitorEventsMD, cm)
						seen[mdStr] = true
					}
				}
			}
		}
	}
	return &record
}

type IntfDirUdn struct {
	Interface string
	Direction int
	Udn       string
}

func NewIntfDirUdn(intf string, dir int, s *ovnobserv.SampleDecoder) IntfDirUdn {
	var udn string
	if s == nil {
		return IntfDirUdn{Interface: intf, Direction: dir, Udn: ""}
	}

	// Load UDN cache if empty
	if len(udnsCache) == 0 {
		m, err := s.GetInterfaceUDNs()
		if err != nil {
			return IntfDirUdn{Interface: intf, Direction: dir, Udn: ""}
		}
		udnsCache = m
	}

	// Look up the interface in the cache
	if v, ok := udnsCache[intf]; ok {
		if v != "" {
			udn = v
		} else {
			udn = "default"
		}
	}

	return IntfDirUdn{Interface: intf, Direction: dir, Udn: udn}
}

func networkEventsMDExist(events [MaxNetworkEvents][NetworkEventsMaxEventsMD]uint8, md [NetworkEventsMaxEventsMD]uint8) bool {
	for _, e := range events {
		if reflect.DeepEqual(e, md) {
			return true
		}
	}
	return false
}

// IP returns the net.IP equivalent object
func IP(ia IPAddr) net.IP {
	return ia[:]
}

// IntEncodeV4 encodes an IPv4 address as an integer (in network encoding, big endian).
// It assumes that the passed IP is already IPv4. Otherwise, it would just encode the
// last 4 bytes of an IPv6 address
func IntEncodeV4(ia [net.IPv6len]uint8) uint32 {
	return binary.BigEndian.Uint32(ia[net.IPv6len-net.IPv4len : net.IPv6len])
}

// IPAddrFromNetIP returns IPAddr from net.IP
func IPAddrFromNetIP(netIP net.IP) IPAddr {
	var arr [net.IPv6len]uint8
	copy(arr[:], (netIP)[0:net.IPv6len])
	return arr
}

func (ia *IPAddr) MarshalJSON() ([]byte, error) {
	return []byte(`"` + IP(*ia).String() + `"`), nil
}

func (m *MacAddr) String() string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", m[0], m[1], m[2], m[3], m[4], m[5])
}

func (m *MacAddr) MarshalJSON() ([]byte, error) {
	return []byte("\"" + m.String() + "\""), nil
}

// ReadFrom reads a Record from a binary source, in LittleEndian order
func ReadFrom(reader io.Reader) (*RawRecord, error) {
	var fr RawRecord
	err := binary.Read(reader, binary.LittleEndian, &fr)
	return &fr, err
}

func AllZerosMetaData(s [NetworkEventsMaxEventsMD]uint8) bool {
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}

func AllZeroIP(ip net.IP) bool {
	if ip.Equal(net.IPv4zero) || ip.Equal(net.IPv6zero) {
		return true
	}
	return false
}
