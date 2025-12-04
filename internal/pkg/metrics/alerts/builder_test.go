package alerts

import (
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/stretchr/testify/assert"
)

func TestBuildLabelFilter(t *testing.T) {
	// Test GroupByNode with source side
	rb := &ruleBuilder{
		healthRule: &flowslatest.HealthRuleVariant{
			GroupBy: flowslatest.GroupByNode,
		},
		side: asSource,
	}
	filter := rb.buildLabelFilter("")
	assert.Equal(t, `{SrcK8S_HostName!=""}`, filter)

	// Test GroupByNode with destination side
	rb.side = asDest
	filter = rb.buildLabelFilter("")
	assert.Equal(t, `{DstK8S_HostName!=""}`, filter)

	// Test GroupByNamespace
	rb.healthRule.GroupBy = flowslatest.GroupByNamespace
	rb.side = asSource
	filter = rb.buildLabelFilter("")
	assert.Equal(t, `{SrcK8S_Namespace!=""}`, filter)

	// Test GroupByWorkload
	rb.healthRule.GroupBy = flowslatest.GroupByWorkload
	rb.side = asDest
	filter = rb.buildLabelFilter("")
	assert.Equal(t, `{DstK8S_Namespace!="",DstK8S_OwnerName!="",DstK8S_OwnerType!=""}`, filter)

	// Test with additional filter
	rb.healthRule.GroupBy = flowslatest.GroupByNamespace
	rb.side = asSource
	filter = rb.buildLabelFilter(`DnsFlagsResponseCode!="NoError"`)
	assert.Equal(t, `{SrcK8S_Namespace!="",DnsFlagsResponseCode!="NoError"}`, filter)

	// Test with action filter (netpol)
	rb.healthRule.GroupBy = flowslatest.GroupByWorkload
	rb.side = asDest
	filter = rb.buildLabelFilter(`action="drop"`)
	assert.Equal(t, `{DstK8S_Namespace!="",DstK8S_OwnerName!="",DstK8S_OwnerType!="",action="drop"}`, filter)

	// Test no grouping (global)
	rb.healthRule.GroupBy = ""
	rb.side = ""
	filter = rb.buildLabelFilter("")
	assert.Equal(t, "", filter)

	// Test no grouping with additional filter
	filter = rb.buildLabelFilter(`DnsFlagsResponseCode!="NoError"`)
	assert.Equal(t, `{DnsFlagsResponseCode!="NoError"}`, filter)
}

func TestRecordingRuleNames(t *testing.T) {
	tests := []struct {
		name     string
		template flowslatest.HealthRuleTemplate
		groupBy  flowslatest.HealthRuleGroupBy
		side     srcOrDst
		expected string
	}{
		// DNS Errors
		{
			name:     "DNSErrors no grouping",
			template: flowslatest.HealthRuleDNSErrors,
			groupBy:  "",
			side:     asDest,
			expected: "netobserv:health:dns_errors:rate5m",
		},
		{
			name:     "DNSErrors by Namespace",
			template: flowslatest.HealthRuleDNSErrors,
			groupBy:  flowslatest.GroupByNamespace,
			side:     asDest,
			expected: "netobserv:health:dns_errors:namespace:dst:rate5m",
		},
		{
			name:     "DNSErrors by Node",
			template: flowslatest.HealthRuleDNSErrors,
			groupBy:  flowslatest.GroupByNode,
			side:     asDest,
			expected: "netobserv:health:dns_errors:node:dst:rate5m",
		},
		{
			name:     "DNSErrors by Workload",
			template: flowslatest.HealthRuleDNSErrors,
			groupBy:  flowslatest.GroupByWorkload,
			side:     asDest,
			expected: "netobserv:health:dns_errors:workload:dst:rate5m",
		},

		// Packet Drops By Kernel
		{
			name:     "PacketDropsByKernel no grouping",
			template: flowslatest.HealthRulePacketDropsByKernel,
			groupBy:  "",
			side:     asSource,
			expected: "netobserv:health:packet_drops_by_kernel:rate5m",
		},
		{
			name:     "PacketDropsByKernel by Namespace src",
			template: flowslatest.HealthRulePacketDropsByKernel,
			groupBy:  flowslatest.GroupByNamespace,
			side:     asSource,
			expected: "netobserv:health:packet_drops_by_kernel:namespace:src:rate5m",
		},
		{
			name:     "PacketDropsByKernel by Namespace dst",
			template: flowslatest.HealthRulePacketDropsByKernel,
			groupBy:  flowslatest.GroupByNamespace,
			side:     asDest,
			expected: "netobserv:health:packet_drops_by_kernel:namespace:dst:rate5m",
		},

		// IPsec Errors
		{
			name:     "IPsecErrors no grouping",
			template: flowslatest.HealthRuleIPsecErrors,
			groupBy:  "",
			side:     asSource,
			expected: "netobserv:health:ipsec_errors:rate5m",
		},
		{
			name:     "IPsecErrors by Node",
			template: flowslatest.HealthRuleIPsecErrors,
			groupBy:  flowslatest.GroupByNode,
			side:     asSource,
			expected: "netobserv:health:ipsec_errors:node:src:rate5m",
		},

		// Netpol Denied
		{
			name:     "NetpolDenied no grouping",
			template: flowslatest.HealthRuleNetpolDenied,
			groupBy:  "",
			side:     asSource,
			expected: "netobserv:health:netpol_denied:rate5m",
		},
		{
			name:     "NetpolDenied by Workload",
			template: flowslatest.HealthRuleNetpolDenied,
			groupBy:  flowslatest.GroupByWorkload,
			side:     asDest,
			expected: "netobserv:health:netpol_denied:workload:dst:rate5m",
		},

		// Cross AZ
		{
			name:     "CrossAZ no grouping",
			template: flowslatest.HealthRuleCrossAZ,
			groupBy:  "",
			side:     asSource,
			expected: "netobserv:health:cross_az:rate5m",
		},
		{
			name:     "CrossAZ by Namespace",
			template: flowslatest.HealthRuleCrossAZ,
			groupBy:  flowslatest.GroupByNamespace,
			side:     asSource,
			expected: "netobserv:health:cross_az:namespace:src:rate5m",
		},

		// Latency High Trend
		{
			name:     "LatencyHighTrend no grouping",
			template: flowslatest.HealthRuleLatencyHighTrend,
			groupBy:  "",
			side:     asSource,
			expected: "netobserv:health:latency_high_trend:rate5m",
		},
		{
			name:     "LatencyHighTrend by Node",
			template: flowslatest.HealthRuleLatencyHighTrend,
			groupBy:  flowslatest.GroupByNode,
			side:     asDest,
			expected: "netobserv:health:latency_high_trend:node:dst:rate5m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := &ruleBuilder{
				template: tt.template,
				healthRule: &flowslatest.HealthRuleVariant{
					GroupBy: tt.groupBy,
				},
				mode: flowslatest.ModeRecording,
				side: tt.side,
			}

			name := rb.buildRecordingRuleName()
			assert.Equal(t, tt.expected, name)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"DNSErrors", "dns_errors"},
		{"PacketDropsByKernel", "packet_drops_by_kernel"},
		{"PacketDropsByDevice", "packet_drops_by_device"},
		{"IPsecErrors", "ipsec_errors"},
		{"NetpolDenied", "netpol_denied"},
		{"LatencyHighTrend", "latency_high_trend"},
		{"ExternalEgressHighTrend", "external_egress_high_trend"},
		{"ExternalIngressHighTrend", "external_ingress_high_trend"},
		{"CrossAZ", "cross_az"},
		{"LokiError", "loki_error"},
		{"NoFlows", "no_flows"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
