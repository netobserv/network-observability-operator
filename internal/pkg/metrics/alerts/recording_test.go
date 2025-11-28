package alerts

import (
	"context"
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/stretchr/testify/assert"
)

func TestBuildRecordingRules(t *testing.T) {
	ctx := context.Background()

	// Create health rules with recording-rule mode
	healthRules := []flowslatest.HealthRule{
		{
			Template: flowslatest.AlertDNSErrors,
			Mode:     flowslatest.HealthRuleModeRecordingRule,
			Variants: []flowslatest.HealthRuleVariant{
				{
					GroupBy: flowslatest.GroupByNamespace,
				},
			},
		},
	}

	// Create a FlowCollectorSpec with recording rules
	spec := &flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				HealthRules: &healthRules,
			},
			Advanced: &flowslatest.AdvancedProcessorConfig{
				Env: map[string]string{
					"EXPERIMENTAL_ALERTS_HEALTH": "true",
				},
			},
		},
		Agent: flowslatest.FlowCollectorAgent{
			Type: flowslatest.AgentEBPF,
			EBPF: flowslatest.FlowCollectorEBPF{
				Features: []flowslatest.AgentFeature{
					flowslatest.DNSTracking,
				},
			},
		},
	}

	rules := BuildRules(ctx, spec)

	// Should have some recording rules
	assert.NotEmpty(t, rules, "should generate recording rules")

	// Verify that recording rules have the correct structure
	hasRecording := false
	for _, rule := range rules {
		if rule.Record != "" {
			hasRecording = true
			assert.NotEmpty(t, rule.Expr.StrVal, "recording rule should have an expression")
			assert.Contains(t, rule.Labels, "netobserv", "recording rule should have netobserv label")
			assert.Equal(t, "health", rule.Labels["netobserv"])
		}
	}

	assert.True(t, hasRecording, "should have at least one recording rule")
}

func TestBuildRules_MixedModes(t *testing.T) {
	ctx := context.Background()

	// Create health rules with mixed modes
	healthRules := []flowslatest.HealthRule{
		{
			Template: flowslatest.AlertDNSErrors,
			Mode:     flowslatest.HealthRuleModeAlert,
			Variants: []flowslatest.HealthRuleVariant{
				{
					Thresholds: flowslatest.AlertThresholds{
						Warning: "5",
					},
					GroupBy: flowslatest.GroupByNamespace,
				},
			},
		},
		{
			Template: flowslatest.AlertPacketDropsByKernel,
			Mode:     flowslatest.HealthRuleModeRecordingRule,
			Variants: []flowslatest.HealthRuleVariant{
				{
					GroupBy: flowslatest.GroupByNode,
				},
			},
		},
	}

	// Disable default templates that we don't want (but not the ones we're overriding)
	disabledTemplates := []flowslatest.AlertTemplate{
		flowslatest.AlertPacketDropsByDevice,
		flowslatest.AlertIPsecErrors,
		flowslatest.AlertNetpolDenied,
		flowslatest.AlertLatencyHighTrend,
		flowslatest.AlertExternalEgressHighTrend,
		flowslatest.AlertExternalIngressHighTrend,
		flowslatest.AlertCrossAZ,
	}

	// Include necessary metrics
	includeList := []flowslatest.FLPMetric{
		"namespace_drop_packets_total",
		"namespace_ingress_packets_total",
		"namespace_dns_latency_seconds",
		"node_drop_packets_total",
		"node_ingress_packets_total",
	}

	spec := &flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				HealthRules:   &healthRules,
				DisableAlerts: disabledTemplates,
				IncludeList:   &includeList,
			},
			Advanced: &flowslatest.AdvancedProcessorConfig{
				Env: map[string]string{
					"EXPERIMENTAL_ALERTS_HEALTH": "true",
				},
			},
		},
		Agent: flowslatest.FlowCollectorAgent{
			Type: flowslatest.AgentEBPF,
			EBPF: flowslatest.FlowCollectorEBPF{
				Privileged: true,
				Features: []flowslatest.AgentFeature{
					flowslatest.DNSTracking,
					flowslatest.PacketDrop,
				},
			},
		},
	}

	rules := BuildRules(ctx, spec)

	// Should have both alerts and recording rules
	hasAlerts := false
	hasRecording := false

	for _, rule := range rules {
		if rule.Alert != "" {
			hasAlerts = true
		}
		if rule.Record != "" {
			hasRecording = true
		}
	}

	assert.True(t, hasAlerts, "should have alert rules for DNS errors")
	assert.True(t, hasRecording, "should have recording rules for packet drops")
}

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PacketDropsByKernel", "packet_drops_by_kernel"},
		{"DNSErrors", "d_n_s_errors"},
		{"NetpolDenied", "netpol_denied"},
		{"LatencyHighTrend", "latency_high_trend"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := camelToSnake(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRecordingRuleName(t *testing.T) {
	rb := ruleBuilder{
		template: flowslatest.AlertPacketDropsByKernel,
		alert: &flowslatest.HealthRuleVariant{
			GroupBy: flowslatest.GroupByNamespace,
		},
		side: asSource,
	}

	name := rb.buildRecordingRuleName()
	assert.Contains(t, name, "netobserv:health:")
	assert.Contains(t, name, "namespace")
	assert.Contains(t, name, "src")
	assert.Contains(t, name, ":rate5m")
}

func TestRecordingRuleLabels(t *testing.T) {
	rb := ruleBuilder{
		template: flowslatest.AlertDNSErrors,
		alert: &flowslatest.HealthRuleVariant{
			GroupBy: flowslatest.GroupByWorkload,
		},
		side: asDest,
	}

	labels := rb.buildRecordingRuleLabels()
	assert.Equal(t, "health", labels["netobserv"])
	assert.Equal(t, "DNSErrors", labels["health_template"])
	assert.Equal(t, "Workload", labels["health_groupby"])
	assert.Equal(t, "Dst", labels["health_side"])
}

func TestBuildRules_SystemRecordingRules(t *testing.T) {
	ctx := context.Background()

	// Create health rules with NoFlows and LokiError as recording rules
	healthRules := []flowslatest.HealthRule{
		{
			Template: flowslatest.AlertNoFlows,
			Mode:     flowslatest.HealthRuleModeRecordingRule,
			Variants: []flowslatest.HealthRuleVariant{{}},
		},
		{
			Template: flowslatest.AlertLokiError,
			Mode:     flowslatest.HealthRuleModeRecordingRule,
			Variants: []flowslatest.HealthRuleVariant{{}},
		},
	}

	spec := &flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				HealthRules: &healthRules,
			},
			Advanced: &flowslatest.AdvancedProcessorConfig{
				Env: map[string]string{
					"EXPERIMENTAL_ALERTS_HEALTH": "true",
				},
			},
		},
	}

	rules := BuildRules(ctx, spec)

	// Should have recording rules for NoFlows and LokiError
	var noFlowsRecording, lokiErrorRecording bool
	for _, rule := range rules {
		if rule.Record == "netobserv:health:no_flows:rate1m" {
			noFlowsRecording = true
			assert.Equal(t, "health", rule.Labels["netobserv"])
			assert.Equal(t, "NetObservNoFlows", rule.Labels["health_template"])
			assert.Contains(t, rule.Expr.StrVal, "netobserv_ingest_flows_processed")
		}
		if rule.Record == "netobserv:health:loki_errors:rate1m" {
			lokiErrorRecording = true
			assert.Equal(t, "health", rule.Labels["netobserv"])
			assert.Equal(t, "NetObservLokiError", rule.Labels["health_template"])
			assert.Contains(t, rule.Expr.StrVal, "netobserv_loki_dropped_entries_total")
		}
		// Should not have alert versions
		assert.NotEqual(t, "NetObservNoFlows", rule.Alert)
		assert.NotEqual(t, "NetObservLokiError", rule.Alert)
	}

	assert.True(t, noFlowsRecording, "should have NoFlows recording rule")
	assert.True(t, lokiErrorRecording, "should have LokiError recording rule")
}
