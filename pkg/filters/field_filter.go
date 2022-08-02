package filters

import "github.com/netobserv/flowlogs-pipeline/pkg/api"

var ovsGoflowUnusedFields = []string{
	"BiFlowDirection",
	"CustomBytes1",
	"CustomBytes2",
	"CustomBytes3",
	"CustomBytes4",
	"CustomBytes5",
	"CustomInteger1",
	"CustomInteger2",
	"CustomInteger3",
	"CustomInteger4",
	"CustomInteger5",
	"DstAS",
	"DstNet",
	"DstVlan",
	"EgressVrfID",
	"IPTTL",
	"IPTos",
	"IngressVrfID",
	"MPLS1Label",
	"MPLS1TTL",
	"MPLS2Label",
	"MPLS2TTL",
	"MPLS3Label",
	"MPLS3TTL",
	"MPLSCount",
	"MPLSLastLabel",
	"MPLSLastTTL",
	"NextHop",
	"NextHopAS",
	"SamplingRate",
	"SrcAS",
	"SrcNet",
	"SrcVlan",
	"TCPFlags",
	"VlanId",
}

func dropListToRules(list []string) []api.TransformFilterRule {
	rules := make([]api.TransformFilterRule, len(list))
	for i, field := range list {
		rules[i] = api.TransformFilterRule{
			Input: field,
			Type:  "remove_field",
		}
	}
	return rules
}

func GetOVSGoflowUnusedRules() []api.TransformFilterRule {
	return dropListToRules(ovsGoflowUnusedFields)
}
