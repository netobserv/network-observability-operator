package alerts

import (
	"context"
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/stretchr/testify/assert"
)

func TestBuildRecordingRules(t *testing.T) {
	ctx := context.Background()

	// Create a FlowCollectorSpec with recording rules mode
	spec := &flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				HealthMode: string(flowslatest.HealthModeRecordingRules),
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

	rules := BuildRecordingRules(ctx, spec)

	// Should have some recording rules
	assert.NotEmpty(t, rules, "should generate recording rules")

	// Verify we have the basic recording rules (NoFlows, LokiError)
	foundNoFlows := false
	foundLokiError := false

	for _, rule := range rules {
		if rule.Record == "netobserv:health:no_flows:rate1m" {
			foundNoFlows = true
			assert.Contains(t, rule.Labels, "health_template")
			assert.Equal(t, "NetObservNoFlows", rule.Labels["health_template"])
		}
		if rule.Record == "netobserv:health:loki_errors:rate1m" {
			foundLokiError = true
			assert.Contains(t, rule.Labels, "health_template")
			assert.Equal(t, "NetObservLokiError", rule.Labels["health_template"])
		}

		// Verify that recording rules have the correct structure
		if rule.Record != "" {
			assert.NotEmpty(t, rule.Expr.StrVal, "recording rule should have an expression")
			assert.Contains(t, rule.Labels, "netobserv", "recording rule should have netobserv label")
			assert.Equal(t, "health", rule.Labels["netobserv"])
		}
	}

	assert.True(t, foundNoFlows, "should have no_flows recording rule")
	assert.True(t, foundLokiError, "should have loki_errors recording rule")
}

func TestBuildRules_Dispatcher(t *testing.T) {
	ctx := context.Background()

	// Test with alerts mode (default)
	specAlerts := &flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				HealthMode: string(flowslatest.HealthModeAlerts),
			},
		},
	}

	alertRules := BuildRules(ctx, specAlerts)

	// Should have alerts (they have Alert field set)
	hasAlerts := false
	for _, rule := range alertRules {
		if rule.Alert != "" {
			hasAlerts = true
			break
		}
	}
	assert.True(t, hasAlerts, "alerts mode should generate alert rules")

	// Test with recording rules mode
	specRecording := &flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				HealthMode: string(flowslatest.HealthModeRecordingRules),
			},
		},
	}

	recordingRules := BuildRules(ctx, specRecording)

	// Should have recording rules (they have Record field set)
	hasRecording := false
	for _, rule := range recordingRules {
		if rule.Record != "" {
			hasRecording = true
			break
		}
	}
	assert.True(t, hasRecording, "recording mode should generate recording rules")

	// Verify no alerts in recording mode
	for _, rule := range recordingRules {
		assert.Empty(t, rule.Alert, "recording mode should not generate alert rules")
	}
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
		alert: &flowslatest.AlertVariant{
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
		alert: &flowslatest.AlertVariant{
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
