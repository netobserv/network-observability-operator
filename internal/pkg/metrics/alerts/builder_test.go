package alerts

import (
	"net/http"
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
			expected: "netobserv:health:dns_errors:rate2m",
		},
		{
			name:     "DNSErrors by Namespace",
			template: flowslatest.HealthRuleDNSErrors,
			groupBy:  flowslatest.GroupByNamespace,
			side:     asDest,
			expected: "netobserv:health:dns_errors:namespace:dst:rate2m",
		},
		{
			name:     "DNSErrors by Node",
			template: flowslatest.HealthRuleDNSErrors,
			groupBy:  flowslatest.GroupByNode,
			side:     asDest,
			expected: "netobserv:health:dns_errors:node:dst:rate2m",
		},
		{
			name:     "DNSErrors by Workload",
			template: flowslatest.HealthRuleDNSErrors,
			groupBy:  flowslatest.GroupByWorkload,
			side:     asDest,
			expected: "netobserv:health:dns_errors:workload:dst:rate2m",
		},

		// Packet Drops By Kernel
		{
			name:     "PacketDropsByKernel no grouping",
			template: flowslatest.HealthRulePacketDropsByKernel,
			groupBy:  "",
			side:     asSource,
			expected: "netobserv:health:packet_drops_by_kernel:rate2m",
		},
		{
			name:     "PacketDropsByKernel by Namespace src",
			template: flowslatest.HealthRulePacketDropsByKernel,
			groupBy:  flowslatest.GroupByNamespace,
			side:     asSource,
			expected: "netobserv:health:packet_drops_by_kernel:namespace:src:rate2m",
		},
		{
			name:     "PacketDropsByKernel by Namespace dst",
			template: flowslatest.HealthRulePacketDropsByKernel,
			groupBy:  flowslatest.GroupByNamespace,
			side:     asDest,
			expected: "netobserv:health:packet_drops_by_kernel:namespace:dst:rate2m",
		},

		// IPsec Errors
		{
			name:     "IPsecErrors no grouping",
			template: flowslatest.HealthRuleIPsecErrors,
			groupBy:  "",
			side:     asSource,
			expected: "netobserv:health:ipsec_errors:rate2m",
		},
		{
			name:     "IPsecErrors by Node",
			template: flowslatest.HealthRuleIPsecErrors,
			groupBy:  flowslatest.GroupByNode,
			side:     asSource,
			expected: "netobserv:health:ipsec_errors:node:src:rate2m",
		},

		// Netpol Denied
		{
			name:     "NetpolDenied no grouping",
			template: flowslatest.HealthRuleNetpolDenied,
			groupBy:  "",
			side:     asSource,
			expected: "netobserv:health:netpol_denied:rate2m",
		},
		{
			name:     "NetpolDenied by Workload",
			template: flowslatest.HealthRuleNetpolDenied,
			groupBy:  flowslatest.GroupByWorkload,
			side:     asDest,
			expected: "netobserv:health:netpol_denied:workload:dst:rate2m",
		},

		// Latency High Trend
		{
			name:     "LatencyHighTrend no grouping",
			template: flowslatest.HealthRuleLatencyHighTrend,
			groupBy:  "",
			side:     asSource,
			expected: "netobserv:health:latency_high_trend:rate2m",
		},
		{
			name:     "LatencyHighTrend by Node",
			template: flowslatest.HealthRuleLatencyHighTrend,
			groupBy:  flowslatest.GroupByNode,
			side:     asDest,
			expected: "netobserv:health:latency_high_trend:node:dst:rate2m",
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

func TestHealthAnnotationMetadata(t *testing.T) {
	tests := []struct {
		name                    string
		groupBy                 flowslatest.HealthRuleGroupBy
		expectedNodeLabels      []string
		expectedNamespaceLabels []string
		expectedOwnerLabels     []string
	}{
		{
			name:                    "Global (no grouping)",
			groupBy:                 "",
			expectedNodeLabels:      nil,
			expectedNamespaceLabels: nil,
			expectedOwnerLabels:     nil,
		},
		{
			name:                    "GroupBy Node",
			groupBy:                 flowslatest.GroupByNode,
			expectedNodeLabels:      []string{"node"},
			expectedNamespaceLabels: nil,
			expectedOwnerLabels:     nil,
		},
		{
			name:                    "GroupBy Namespace",
			groupBy:                 flowslatest.GroupByNamespace,
			expectedNodeLabels:      nil,
			expectedNamespaceLabels: []string{"namespace"},
			expectedOwnerLabels:     nil,
		},
		{
			name:                    "GroupBy Workload",
			groupBy:                 flowslatest.GroupByWorkload,
			expectedNodeLabels:      nil,
			expectedNamespaceLabels: nil,
			expectedOwnerLabels:     []string{"workload"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := &ruleBuilder{
				template: flowslatest.HealthRulePacketDropsByKernel,
				healthRule: &flowslatest.HealthRuleVariant{
					GroupBy: tt.groupBy,
					Thresholds: flowslatest.HealthRuleThresholds{
						Critical: "10",
					},
				},
				threshold: "10",
				side:      asSource,
			}

			// Build the health annotation
			annotBytes, err := rb.buildHealthAnnotation(nil)
			assert.NoError(t, err)
			annotStr := string(annotBytes)

			// Check for nodeLabels
			if tt.expectedNodeLabels != nil {
				assert.Contains(t, annotStr, `"nodeLabels":["node"]`,
					"Expected nodeLabels in annotation for %s", tt.name)
			} else {
				assert.NotContains(t, annotStr, `"nodeLabels"`,
					"Did not expect nodeLabels in annotation for %s", tt.name)
			}

			// Check for namespaceLabels
			if tt.expectedNamespaceLabels != nil {
				assert.Contains(t, annotStr, `"namespaceLabels":["namespace"]`,
					"Expected namespaceLabels in annotation for %s", tt.name)
			} else {
				assert.NotContains(t, annotStr, `"namespaceLabels"`,
					"Did not expect namespaceLabels in annotation for %s", tt.name)
			}

			// Check for ownerLabels
			if tt.expectedOwnerLabels != nil {
				assert.Contains(t, annotStr, `"ownerLabels":["workload"]`,
					"Expected ownerLabels in annotation for %s", tt.name)
			} else {
				assert.NotContains(t, annotStr, `"ownerLabels"`,
					"Did not expect ownerLabels in annotation for %s", tt.name)
			}
		})
	}
}

func TestBuildRunbookURL(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		// Health Rule templates
		{
			template: "PacketDropsByKernel",
			expected: runbookURLBase + "/PacketDropsByKernel.md",
		},
		{
			template: "PacketDropsByDevice",
			expected: runbookURLBase + "/PacketDropsByDevice.md",
		},
		{
			template: "IPsecErrors",
			expected: runbookURLBase + "/IPsecErrors.md",
		},
		{
			template: "NetpolDenied",
			expected: runbookURLBase + "/NetpolDenied.md",
		},
		{
			template: "LatencyHighTrend",
			expected: runbookURLBase + "/LatencyHighTrend.md",
		},
		{
			template: "DNSErrors",
			expected: runbookURLBase + "/DNSErrors.md",
		},
		{
			template: "DNSNxDomain",
			expected: runbookURLBase + "/DNSNxDomain.md",
		},
		{
			template: "ExternalEgressHighTrend",
			expected: runbookURLBase + "/ExternalEgressHighTrend.md",
		},
		{
			template: "ExternalIngressHighTrend",
			expected: runbookURLBase + "/ExternalIngressHighTrend.md",
		},
		// Operator alert templates
		{
			template: "NetObservNoFlows",
			expected: runbookURLBase + "/NetObservNoFlows.md",
		},
		{
			template: "NetObservLokiError",
			expected: runbookURLBase + "/NetObservLokiError.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			url := buildRunbookURL(tt.template)
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestRunbookURLsExist(t *testing.T) {
	// Test all templates to ensure their runbook URLs exist
	templates := []string{
		// Health Rule templates
		"PacketDropsByKernel",
		"PacketDropsByDevice",
		"IPsecErrors",
		"NetpolDenied",
		"LatencyHighTrend",
		"DNSErrors",
		"DNSNxDomain",
		"ExternalEgressHighTrend",
		"ExternalIngressHighTrend",
		// Operator alert templates
		"NetObservNoFlows",
		"NetObservLokiError",
	}

	for _, template := range templates {
		t.Run(template, func(t *testing.T) {
			url := buildRunbookURL(template)
			resp, err := http.Get(url)
			assert.NoError(t, err, "Failed to fetch runbook URL: %s", url)
			if err == nil {
				defer resp.Body.Close()
				assert.NotEqual(t, http.StatusNotFound, resp.StatusCode,
					"Runbook URL returns 404: %s", url)
			}
		})
	}
}
