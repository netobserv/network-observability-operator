package flow

import (
	"net"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
)

type InterfaceNamer func(ifIndex int) string

// Decorate adds to the flows extra metadata fields that are not directly fetched by eBPF:
// - The interface name (corresponding to the interface index in the flow).
// - The IP address of the agent host.
func Decorate(agentIP net.IP, ifaceNamer InterfaceNamer) func(in <-chan []*model.Record, out chan<- []*model.Record) {
	return func(in <-chan []*model.Record, out chan<- []*model.Record) {
		for flows := range in {
			for _, flow := range flows {
				flow.Interface = ifaceNamer(int(flow.ID.IfIndex))
				flow.AgentIP = agentIP
			}
			out <- flows
		}
	}
}
