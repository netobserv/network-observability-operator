package metrics

import (
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/stretchr/testify/assert"
)

func TestIncludeExclude(t *testing.T) {
	assert := assert.New(t)

	// IgnoreTags set, Include list unset => resolving ignore tags
	res := GetAsIncludeList([]string{"egress", "packets", "flows"}, nil)
	assert.Equal([]flowslatest.FLPMetric{
		"node_ingress_bytes_total",
		"node_rtt_seconds",
		"node_drop_bytes_total",
		"node_dns_latency_seconds",
		"namespace_ingress_bytes_total",
		"namespace_rtt_seconds",
		"namespace_drop_bytes_total",
		"namespace_dns_latency_seconds",
		"workload_ingress_bytes_total",
		"workload_rtt_seconds",
		"workload_drop_bytes_total",
		"workload_dns_latency_seconds",
	}, *res)

	// IgnoreTags set, Include list set => keep include list
	res = GetAsIncludeList([]string{"egress", "packets"}, &[]flowslatest.FLPMetric{"namespace_flows_total"})
	assert.Equal([]flowslatest.FLPMetric{"namespace_flows_total"}, *res)

	// IgnoreTags set as defaults, Include list unset => use default include list
	res = GetAsIncludeList([]string{"egress", "packets", "nodes-flows", "namespaces-flows", "workloads-flows", "namespaces"}, nil)
	assert.Nil(res)

	// IgnoreTags set as defaults, Include list set => use include list
	res = GetAsIncludeList([]string{"egress", "packets", "nodes-flows", "namespaces-flows", "workloads-flows", "namespaces"}, &[]flowslatest.FLPMetric{"namespace_flows_total"})
	assert.Equal([]flowslatest.FLPMetric{"namespace_flows_total"}, *res)
}

func TestGetDefinitions(t *testing.T) {
	assert := assert.New(t)

	res := GetDefinitions([]string{"namespace_flows_total", "node_ingress_bytes_total", "workload_egress_packets_total"})
	assert.Len(res, 3)
	assert.Equal("node_ingress_bytes_total", res[0].Name)
	assert.Equal("Bytes", res[0].ValueKey)
	assert.Equal([]string{"SrcK8S_HostName", "DstK8S_HostName"}, res[0].Labels)
	assert.Equal("namespace_flows_total", res[1].Name)
	assert.Empty(res[1].ValueKey)
	assert.Equal([]string{"SrcK8S_Namespace", "DstK8S_Namespace"}, res[1].Labels)
	assert.Equal("workload_egress_packets_total", res[2].Name)
	assert.Equal("Packets", res[2].ValueKey)
	assert.Equal([]string{"SrcK8S_Namespace", "DstK8S_Namespace", "SrcK8S_OwnerName", "DstK8S_OwnerName", "SrcK8S_OwnerType", "DstK8S_OwnerType"}, res[2].Labels)
}
