package dashboards

import (
	"fmt"
	"strings"
)

type metricScope struct {
	metricPart                string
	titlePart                 string
	labels                    []string
	legendPart                string
	labelsReplacementTemplate string
	splitAppInfra             bool
}

var (
	srcDstNodeScope = metricScope{
		metricPart: "node",
		titlePart:  "per node",
		labels:     []string{"SrcK8S_HostName", "DstK8S_HostName"},
		legendPart: "{{SrcK8S_HostName}} -> {{DstK8S_HostName}}",
		labelsReplacementTemplate: `label_replace(
			label_replace(
				%s,
				"SrcK8S_HostName", "(not namespaced)", "SrcK8S_HostName", "()"
			),
			"DstK8S_HostName", "(not namespaced)", "DstK8S_HostName", "()"
		)`,
		splitAppInfra: false,
	}
	srcDstNamespaceScope = metricScope{
		metricPart: "namespace",
		titlePart:  "per namespace",
		labels:     []string{"SrcK8S_Namespace", "DstK8S_Namespace"},
		legendPart: "{{SrcK8S_Namespace}} -> {{DstK8S_Namespace}}",
		labelsReplacementTemplate: `label_replace(
			label_replace(
				%s,
				"SrcK8S_Namespace", "(not namespaced)", "SrcK8S_Namespace", "()"
			),
			"DstK8S_Namespace", "(not namespaced)", "DstK8S_Namespace", "()"
		)`,
		splitAppInfra: true,
	}
	srcDstWorkloadScope = metricScope{
		metricPart: "workload",
		titlePart:  "per workload",
		labels:     []string{"SrcK8S_Namespace", "SrcK8S_OwnerName", "DstK8S_Namespace", "DstK8S_OwnerName"},
		legendPart: "{{SrcK8S_OwnerName}} ({{SrcK8S_Namespace}}) -> {{DstK8S_OwnerName}} ({{DstK8S_Namespace}})",
		labelsReplacementTemplate: `label_replace(
			label_replace(
				%s,
				"SrcK8S_Namespace", "non pods", "SrcK8S_Namespace", "()"
			),
			"DstK8S_Namespace", "non pods", "DstK8S_Namespace", "()"
		)`,
		splitAppInfra: true,
	}
)

func (s *metricScope) joinLabels() string {
	return strings.Join(s.labels, ",")
}

func (s *metricScope) labelReplace(q string) string {
	return fmt.Sprintf(s.labelsReplacementTemplate, q)
}
