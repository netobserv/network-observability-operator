package tracer

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"

	cilium "github.com/cilium/ebpf"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/ebpf"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type FilterConfig struct {
	FilterDirection       string
	FilterIPCIDR          string
	FilterProtocol        string
	FilterSourcePort      intstr.IntOrString
	FilterDestinationPort intstr.IntOrString
	FilterPort            intstr.IntOrString
	FilterIcmpType        int
	FilterIcmpCode        int
	FilterPeerIP          string
	FilterAction          string
	FilterTCPFlags        string
	FilterDrops           bool
	FilterSample          uint32
}

type Filter struct {
	// eBPF objs to create/update eBPF maps
	objects *ebpf.BpfObjects
	config  []*FilterConfig
}

func NewFilter(objects *ebpf.BpfObjects, cfg []*FilterConfig) *Filter {
	return &Filter{
		objects: objects,
		config:  cfg,
	}
}

func (f *Filter) ProgramFilter() error {

	for _, config := range f.config {
		log.Infof("Flow filter config: %v", f.config)
		key, err := f.getFilterKey(config)
		if err != nil {
			return fmt.Errorf("failed to get filter key: %w", err)
		}

		val, err := f.getFilterValue(config)
		if err != nil {
			return fmt.Errorf("failed to get filter value: %w", err)
		}

		err = f.objects.FilterMap.Update(key, val, cilium.UpdateAny)
		if err != nil {
			return fmt.Errorf("failed to update filter map: %w", err)
		}

		log.Infof("Programmed filter with key: %v, value: %v", key, val)
	}
	return nil
}

func (f *Filter) getFilterKey(config *FilterConfig) (ebpf.BpfFilterKeyT, error) {
	key := ebpf.BpfFilterKeyT{}
	if config.FilterIPCIDR == "" {
		config.FilterIPCIDR = "0.0.0.0/0"
	}
	ip, ipNet, err := net.ParseCIDR(config.FilterIPCIDR)
	if err != nil {
		return key, fmt.Errorf("failed to parse FlowFilterIPCIDR: %w", err)
	}
	if ip.To4() != nil {
		copy(key.IpData[:], ip.To4())
	} else {
		copy(key.IpData[:], ip.To16())
	}
	pfLen, _ := ipNet.Mask.Size()
	key.PrefixLen = uint32(pfLen)

	return key, nil
}

// nolint:cyclop
func (f *Filter) getFilterValue(config *FilterConfig) (ebpf.BpfFilterValueT, error) {
	val := ebpf.BpfFilterValueT{}

	switch config.FilterDirection {
	case "Ingress":
		val.Direction = ebpf.BpfDirectionTINGRESS
	case "Egress":
		val.Direction = ebpf.BpfDirectionTEGRESS
	default:
		val.Direction = ebpf.BpfDirectionTMAX_DIRECTION
	}

	switch config.FilterAction {
	case "Reject":
		val.Action = ebpf.BpfFilterActionTREJECT
	case "Accept":
		val.Action = ebpf.BpfFilterActionTACCEPT
	default:
		val.Action = ebpf.BpfFilterActionTMAX_FILTER_ACTIONS
	}

	switch config.FilterProtocol {
	case "TCP":
		val.Protocol = syscall.IPPROTO_TCP
	case "UDP":
		val.Protocol = syscall.IPPROTO_UDP
	case "SCTP":
		val.Protocol = syscall.IPPROTO_SCTP
	case "ICMP":
		val.Protocol = syscall.IPPROTO_ICMP
	case "ICMPv6":
		val.Protocol = syscall.IPPROTO_ICMPV6
	}

	val.DstPortStart, val.DstPortEnd = getDstPortsRange(config)
	val.DstPort1, val.DstPort2 = getDstPorts(config)
	val.SrcPortStart, val.SrcPortEnd = getSrcPortsRange(config)
	val.SrcPort1, val.SrcPort2 = getSrcPorts(config)
	val.PortStart, val.PortEnd = getPortsRange(config)
	val.Port1, val.Port2 = getPorts(config)
	val.IcmpType = uint8(config.FilterIcmpType)
	val.IcmpCode = uint8(config.FilterIcmpCode)

	if config.FilterPeerIP != "" {
		ip := net.ParseIP(config.FilterPeerIP)
		if ip.To4() != nil {
			copy(val.Ip[:], ip.To4())
		} else {
			copy(val.Ip[:], ip.To16())
		}
	}

	switch config.FilterTCPFlags {
	case "SYN":
		val.TcpFlags = ebpf.BpfTcpFlagsTSYN_FLAG
	case "SYN-ACK":
		val.TcpFlags = ebpf.BpfTcpFlagsTSYN_ACK_FLAG
	case "ACK":
		val.TcpFlags = ebpf.BpfTcpFlagsTACK_FLAG
	case "FIN":
		val.TcpFlags = ebpf.BpfTcpFlagsTFIN_FLAG
	case "RST":
		val.TcpFlags = ebpf.BpfTcpFlagsTRST_FLAG
	case "PUSH":
		val.TcpFlags = ebpf.BpfTcpFlagsTPSH_FLAG
	case "URG":
		val.TcpFlags = ebpf.BpfTcpFlagsTURG_FLAG
	case "ECE":
		val.TcpFlags = ebpf.BpfTcpFlagsTECE_FLAG
	case "CWR":
		val.TcpFlags = ebpf.BpfTcpFlagsTCWR_FLAG
	case "FIN-ACK":
		val.TcpFlags = ebpf.BpfTcpFlagsTFIN_ACK_FLAG
	case "RST-ACK":
		val.TcpFlags = ebpf.BpfTcpFlagsTRST_ACK_FLAG
	}

	if config.FilterDrops {
		val.FilterDrops = 1
	}

	if config.FilterSample != 0 {
		val.Sample = config.FilterSample
	}
	return val, nil
}

func getSrcPortsRange(config *FilterConfig) (uint16, uint16) {
	if config.FilterSourcePort.Type == intstr.Int {
		return uint16(config.FilterSourcePort.IntVal), 0
	}
	start, end, err := getPortsFromString(config.FilterSourcePort.String(), "-")
	if err != nil {
		return 0, 0
	}
	return start, end
}

func getSrcPorts(config *FilterConfig) (uint16, uint16) {
	port1, port2, err := getPortsFromString(config.FilterSourcePort.String(), ",")
	if err != nil {
		return 0, 0
	}
	return port1, port2
}

func getDstPortsRange(config *FilterConfig) (uint16, uint16) {
	if config.FilterDestinationPort.Type == intstr.Int {
		return uint16(config.FilterDestinationPort.IntVal), 0
	}
	start, end, err := getPortsFromString(config.FilterDestinationPort.String(), "-")
	if err != nil {
		return 0, 0
	}
	return start, end
}

func getDstPorts(config *FilterConfig) (uint16, uint16) {
	port1, port2, err := getPortsFromString(config.FilterDestinationPort.String(), ",")
	if err != nil {
		return 0, 0
	}
	return port1, port2
}

func getPortsRange(config *FilterConfig) (uint16, uint16) {
	if config.FilterPort.Type == intstr.Int {
		return uint16(config.FilterPort.IntVal), 0
	}
	start, end, err := getPortsFromString(config.FilterPort.String(), "-")
	if err != nil {
		return 0, 0
	}
	return start, end
}

func getPorts(config *FilterConfig) (uint16, uint16) {
	port1, port2, err := getPortsFromString(config.FilterPort.String(), ",")
	if err != nil {
		return 0, 0
	}
	return port1, port2
}

func getPortsFromString(s, sep string) (uint16, uint16, error) {
	ps := strings.SplitN(s, sep, 2)
	if len(ps) != 2 {
		return 0, 0, fmt.Errorf("invalid ports range. Expected two integers separated by %s but found %s", sep, s)
	}
	startPort, err := strconv.ParseUint(ps[0], 10, 16)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start port number %w", err)
	}
	endPort, err := strconv.ParseUint(ps[1], 10, 16)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end port number %w", err)
	}
	if sep == "-" && startPort > endPort {
		return 0, 0, fmt.Errorf("invalid port range. Start port is greater than end port")
	}
	if startPort == endPort {
		return 0, 0, fmt.Errorf("invalid port range. Start and end port are equal. Remove the %s and enter a single port", sep)
	}
	if startPort == 0 {
		return 0, 0, fmt.Errorf("invalid start port 0")
	}
	return uint16(startPort), uint16(endPort), nil
}

func ConvertFilterPortsToInstr(intPort int32, rangePorts, ports string) intstr.IntOrString {
	if rangePorts != "" {
		return intstr.FromString(rangePorts)
	}
	if ports != "" {
		return intstr.FromString(ports)
	}
	return intstr.FromInt32(intPort)
}
