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
