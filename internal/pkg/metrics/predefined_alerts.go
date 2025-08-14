package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type alertConversion struct {
	summary     func(string) string
	description func(string, string) string
}

var (
	conversions = map[flowslatest.FLPAlertGroupName]alertConversion{
		flowslatest.AlertTooManyDrops: {
			summary: func(groupDir string) string {
				return fmt.Sprintf("Too many packets are dropped %s", groupDir)
			},
			description: func(fromTo, threshold string) string {
				return fmt.Sprintf("NetObserv is detecting more than %s%% of dropped packets%s.", threshold, fromTo)
			},
		},
	}
)

func BuildAlertRules(ctx context.Context, fc *flowslatest.FlowCollectorSpec) []monitoringv1.Rule {
	log := log.FromContext(ctx)
	rules := []monitoringv1.Rule{}

	groups := fc.GetFLPAlerts()
	for _, group := range groups {
		if ok, _ := group.IsAllowed(fc); !ok {
			continue
		}
		for _, alert := range group.Alerts {
			if r, err := convertToRules(group.Name, &alert); err != nil {
				log.Error(err, "unable to configure an alert")
			} else if len(r) > 0 {
				rules = append(rules, r...)
			}
		}
	}

	if !slices.Contains(fc.Processor.Metrics.DisableAlerts, flowslatest.AlertNoFlows) {
		r := alertNoFlows()
		rules = append(rules, *r)
	}
	if !slices.Contains(fc.Processor.Metrics.DisableAlerts, flowslatest.AlertLokiError) {
		r := alertLokiError()
		rules = append(rules, *r)
	}

	return rules
}

func convertToRules(groupName flowslatest.FLPAlertGroupName, alert *flowslatest.FLPAlert) ([]monitoringv1.Rule, error) {
	var rules []monitoringv1.Rule
	var upperThreshold string
	// Create up to 3 rules, one per severity, with non-overlapping thresholds
	thresholds := []struct {
		s string
		t string
	}{
		{s: "critical", t: alert.Thresholds.Critical},
		{s: "warning", t: alert.Thresholds.Warning},
		{s: "info", t: alert.Thresholds.Info},
	}
	for _, st := range thresholds {
		if st.t != "" {
			if r, err := convertToRule(groupName, alert, st.s, st.t, upperThreshold); err != nil {
				return nil, err
			} else if r != nil {
				rules = append(rules, *r)
			}
			upperThreshold = st.t
		}
	}
	return rules, nil
}

func convertToRule(groupName flowslatest.FLPAlertGroupName, alert *flowslatest.FLPAlert, severity, threshold, upperThreshold string) (*monitoringv1.Rule, error) {
	conv, found := conversions[groupName]
	if !found {
		return nil, fmt.Errorf("unknown alert group name: %s", groupName)
	}

	suffix, labels, text := getLabelsAndTexts(alert)

	d := monitoringv1.Duration("5m")
	additionalDescription := fmt.Sprintf("You can turn off this alert by adding '%s' to spec.processor.metrics.disableAlerts in FlowCollector, or reconfigure it via spec.processor.metrics.alertGroups.", groupName)

	metrics, totalMetrics := flowslatest.GetElligibleMetricsForAlert(groupName, alert)
	var strLabels string
	if len(labels) > 0 {
		strLabels = fmt.Sprintf(" by (%s)", strings.Join(labels, ","))
	}
	for i := 0; i < len(metrics); i++ {
		metrics[i] = "rate(netobserv_" + metrics[i] + "[2m])"
	}
	for i := 0; i < len(totalMetrics); i++ {
		totalMetrics[i] = "rate(netobserv_" + totalMetrics[i] + "[2m])"
	}

	var lowVolumeThreshold, upPctThreshold string
	if alert.LowVolumeThreshold != "" {
		lowVolumeThreshold = " > " + alert.LowVolumeThreshold
	}
	if upperThreshold != "" {
		upPctThreshold = " < " + upperThreshold
	}

	promql := fmt.Sprintf(
		"100 * sum (%s)%s / (sum(%s)%s%s) > %s%s",
		strings.Join(metrics, " OR "),
		strLabels,
		strings.Join(totalMetrics, " OR "),
		strLabels,
		lowVolumeThreshold,
		threshold,
		upPctThreshold,
	)

	bAnnot, err := buildHealthAnnotation(groupName, alert, threshold)
	if err != nil {
		return nil, err
	}

	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("NetObserv%s%s%s", groupName, strings.ToUpper(string(severity[0])), suffix),
		Annotations: map[string]string{
			"description":                 conv.description(text, threshold) + " " + additionalDescription,
			"summary":                     conv.summary(getSummaryComplement(alert)),
			"netobserv_io_network_health": string(bAnnot),
		},
		Expr: intstr.FromString(promql),
		For:  &d,
		Labels: map[string]string{
			"severity":  strings.ToLower(severity),
			"app":       "netobserv", // means that the rule is created by netobserv
			"netobserv": "true",      // means that the rule should be fetched by netobserv console plugin for health
		},
	}, nil
}

func getLabelsAndTexts(alert *flowslatest.FLPAlert) (string, []string, string) {
	var nameSuffix string
	var labelRoots []string
	var detailedTextFunc func(string) string
	switch alert.Grouping {
	case flowslatest.GroupingPerNode:
		labelRoots = []string{"K8S_HostName"}
		detailedTextFunc = func(dir string) string { return fmt.Sprintf("node={{ $labels.%sK8S_HostName }}", dir) }
		nameSuffix = "Node"
	case flowslatest.GroupingPerNamespace:
		labelRoots = []string{"K8S_Namespace"}
		detailedTextFunc = func(dir string) string { return fmt.Sprintf("namespace={{ $labels.%sK8S_Namespace }}", dir) }
		nameSuffix = "Namespace"
	case flowslatest.GroupingPerWorkload:
		labelRoots = []string{"K8S_Namespace", "K8S_OwnerName", "K8S_OwnerType"}
		detailedTextFunc = func(dir string) string {
			return fmt.Sprintf("workload={{ $labels.%sK8S_OwnerName }} ({{ $labels.%sK8S_OwnerType }})", dir, dir)
		}
		nameSuffix = "Workload"
	}
	var labels []string
	var strFrom, strTo string
	if alert.GroupingDirection == flowslatest.GroupingBySource || alert.GroupingDirection == flowslatest.GroupingBySourceAndDestination {
		for _, lblRoot := range labelRoots {
			labels = append(labels, "Src"+lblRoot)
		}
		if detailedTextFunc != nil {
			strFrom = fmt.Sprintf(" from [%s]", detailedTextFunc("Src"))
		}
		if nameSuffix != "" && alert.GroupingDirection != flowslatest.GroupingBySourceAndDestination {
			nameSuffix = "Src" + nameSuffix
		}
	}
	if alert.GroupingDirection == flowslatest.GroupingByDestination || alert.GroupingDirection == flowslatest.GroupingBySourceAndDestination {
		for _, lblRoot := range labelRoots {
			labels = append(labels, "Dst"+lblRoot)
		}
		if detailedTextFunc != nil {
			strTo = fmt.Sprintf(" to [%s]", detailedTextFunc("Dst"))
		}
		if nameSuffix != "" && alert.GroupingDirection != flowslatest.GroupingBySourceAndDestination {
			nameSuffix = "Dst" + nameSuffix
		}
	}

	return nameSuffix, labels, strFrom + strTo
}

func buildHealthAnnotation(groupName flowslatest.FLPAlertGroupName, alert *flowslatest.FLPAlert, threshold string) ([]byte, error) {
	// The health annotation contains json-encoded information used in console plugin display
	annotation := map[string]any{
		"threshold": threshold,
		"unit":      "%",
	}
	switch alert.Grouping {
	case flowslatest.GroupingPerNode:
		annotation["nodeLabels"] = []string{"SrcK8S_HostName", "DstK8S_HostName"}
	case flowslatest.GroupingPerNamespace, flowslatest.GroupingPerWorkload:
		annotation["namespaceLabels"] = []string{"SrcK8S_Namespace", "DstK8S_Namespace"}
	}
	bAnnot, err := json.Marshal(annotation)
	if err != nil {
		return nil, fmt.Errorf("cannot encode alert annotation [name=%s]: %w", groupName, err)
	}
	return bAnnot, nil
}

func getSummaryComplement(alert *flowslatest.FLPAlert) string {
	summaryGroupDir := ""
	switch alert.Grouping {
	case flowslatest.GroupingPerNode:
		summaryGroupDir = "node"
	case flowslatest.GroupingPerNamespace:
		summaryGroupDir = "namespace"
	case flowslatest.GroupingPerWorkload:
		summaryGroupDir = "workload"
	}
	if summaryGroupDir != "" {
		switch alert.GroupingDirection {
		case flowslatest.GroupingBySource:
			summaryGroupDir = "from " + summaryGroupDir
		case flowslatest.GroupingByDestination:
			summaryGroupDir = "to " + summaryGroupDir
		case flowslatest.GroupingBySourceAndDestination:
			summaryGroupDir = "between " + summaryGroupDir + "s"
		}
	}
	return summaryGroupDir
}

func alertNoFlows() *monitoringv1.Rule {
	d := monitoringv1.Duration("10m")

	// Not receiving flows
	return &monitoringv1.Rule{
		Alert: string(flowslatest.AlertNoFlows),
		Annotations: map[string]string{
			"description": "NetObserv flowlogs-pipeline is not receiving any flow, this is either a connection issue with the agent, or an agent issue",
			"summary":     "NetObserv flowlogs-pipeline is not receiving any flow",
		},
		Expr: intstr.FromString("sum(rate(netobserv_ingest_flows_processed[1m])) == 0"),
		For:  &d,
		Labels: map[string]string{
			"severity": "warning",
			"app":      "netobserv",
		},
	}
}

func alertLokiError() *monitoringv1.Rule {
	d := monitoringv1.Duration("10m")

	return &monitoringv1.Rule{
		Alert: string(flowslatest.AlertLokiError),
		Annotations: map[string]string{
			"description": "NetObserv flowlogs-pipeline is dropping flows because of Loki errors, Loki may be down or having issues ingesting every flows. Please check Loki and flowlogs-pipeline logs.",
			"summary":     "NetObserv flowlogs-pipeline is dropping flows because of Loki errors",
		},
		Expr: intstr.FromString("sum(rate(netobserv_loki_dropped_entries_total[1m])) > 0"),
		For:  &d,
		Labels: map[string]string{
			"severity": "warning",
			"app":      "netobserv",
		},
	}
}
