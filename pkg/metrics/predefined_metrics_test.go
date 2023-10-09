package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncludeExclude(t *testing.T) {
	assert := assert.New(t)

	// IgnoreTags set, Include list unset => resolving ignore tags
	res := GetEnabledNames([]string{"egress", "packets", "flows"}, nil)
	assert.Equal([]string{
		"node_ingress_bytes_total",
		"node_rtt",
		"node_drop_packets_total",
		"node_drop_bytes_total",
		"namespace_ingress_bytes_total",
		"namespace_rtt",
		"namespace_drop_packets_total",
		"namespace_drop_bytes_total",
		"workload_ingress_bytes_total",
		"workload_rtt",
		"workload_drop_packets_total",
		"workload_drop_bytes_total",
	}, res)

	// IgnoreTags set, Include list set => keep include list
	res = GetEnabledNames([]string{"egress", "packets"}, &[]string{"namespace_flows_total"})
	assert.Equal([]string{"namespace_flows_total"}, res)

	// IgnoreTags set as defaults, Include list unset => use default include list
	res = GetEnabledNames([]string{"egress", "packets", "nodes-flows", "namespaces-flows", "workloads-flows", "namespaces"}, nil)
	assert.Equal(DefaultIncludeList, res)

	// IgnoreTags set as defaults, Include list set => use include list
	res = GetEnabledNames([]string{"egress", "packets", "nodes-flows", "namespaces-flows", "workloads-flows", "namespaces"}, &[]string{"namespace_flows_total"})
	assert.Equal([]string{"namespace_flows_total"}, res)
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
