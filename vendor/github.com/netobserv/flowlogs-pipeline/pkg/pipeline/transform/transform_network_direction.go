package transform

import (
	"fmt"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
)

const (
	ingress = 0
	egress  = 1
	inner   = 2
)

func validateReinterpretDirectionConfig(info *api.NetworkTransformDirectionInfo) error {
	if info.FlowDirectionField == "" {
		return fmt.Errorf("invalid config for transform.Network rule %s: missing FlowDirectionField", api.NetworkReinterpretDirection)
	}
	if info.ReporterIPField == "" {
		return fmt.Errorf("invalid config for transform.Network rule %s: missing ReporterIPField", api.NetworkReinterpretDirection)
	}
	if info.SrcHostField == "" {
		return fmt.Errorf("invalid config for transform.Network rule %s: missing SrcHostField", api.NetworkReinterpretDirection)
	}
	if info.DstHostField == "" {
		return fmt.Errorf("invalid config for transform.Network rule %s: missing DstHostField", api.NetworkReinterpretDirection)
	}
	return nil
}

func reinterpretDirection(output config.GenericMap, info *api.NetworkTransformDirectionInfo) {
	if fd, ok := output[info.FlowDirectionField]; ok && len(info.IfDirectionField) > 0 {
		output[info.IfDirectionField] = fd
	}
	var srcNode, dstNode, reporter string
	if gen, ok := output[info.ReporterIPField]; ok {
		if str, ok := gen.(string); ok {
			reporter = str
		}
	}
	if len(reporter) == 0 {
		return
	}
	if gen, ok := output[info.SrcHostField]; ok {
		if str, ok := gen.(string); ok {
			srcNode = str
		}
	}
	if gen, ok := output[info.DstHostField]; ok {
		if str, ok := gen.(string); ok {
			dstNode = str
		}
	}
	if srcNode != dstNode {
		if srcNode == reporter {
			output[info.FlowDirectionField] = egress
		} else if dstNode == reporter {
			output[info.FlowDirectionField] = ingress
		}
	} else if srcNode != "" {
		output[info.FlowDirectionField] = inner
	}
}
