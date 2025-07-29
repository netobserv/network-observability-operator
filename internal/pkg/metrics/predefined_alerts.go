package metrics

import (
	"context"
	"fmt"
	"slices"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type alertConversion struct {
	summary     string
	description func(string, string) string
}

const (
	healthAnnotation = "netobserv_io_health"
)

var (
	conversions = map[flowslatest.FLPAlertGroupName]alertConversion{
		flowslatest.AlertTooManyDrops: {
			summary: "Too many packets are dropped",
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
		for i, alert := range group.Alerts {
			if rule, err := convertToRule(group.Name, i, &alert); err != nil {
				log.Error(err, "unable to configure an alert")
			} else if rule != nil {
				rules = append(rules, *rule)
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

func convertToRule(groupName flowslatest.FLPAlertGroupName, idx int, alert *flowslatest.FLPAlert) (*monitoringv1.Rule, error) {
	conv, found := conversions[groupName]
	if !found {
		return nil, fmt.Errorf("unknown alert group name: %s", groupName)
	}

	labels, text := getLabelsAndTexts(alert)

	d := monitoringv1.Duration("5m")
	additionalDescription := fmt.Sprintf("You can turn off this alert by adding '%s' to spec.processor.metrics.disableAlerts in FlowCollector, or configure it via spec.processor.metrics.alertGroups.", groupName)

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
	promql := fmt.Sprintf(
		"100 * sum (%s)%s / sum(%s)%s > %s",
		strings.Join(metrics, " OR "),
		strLabels,
		strings.Join(totalMetrics, " OR "),
		strLabels,
		alert.Threshold,
	)

	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("NetObserv%s_%d", groupName, idx),
		Annotations: map[string]string{
			"description":    conv.description(text, alert.Threshold) + " " + additionalDescription,
			"summary":        conv.summary,
			healthAnnotation: "{}",
		},
		Expr: intstr.FromString(promql),
		For:  &d,
		Labels: map[string]string{
			"severity": strings.ToLower(string(alert.Severity)),
			"app":      "netobserv",
		},
	}, nil
}

func getLabelsAndTexts(alert *flowslatest.FLPAlert) ([]string, string) {
	var labelRoots []string
	var textFunc func(string) string
	switch alert.Grouping {
	case flowslatest.GroupingPerNode:
		labelRoots = []string{"K8S_HostName"}
		textFunc = func(dir string) string { return fmt.Sprintf("node={{ $labels.%sK8S_HostName }}", dir) }
	case flowslatest.GroupingPerNamespace:
		labelRoots = []string{"K8S_Namespace"}
		textFunc = func(dir string) string { return fmt.Sprintf("namespace={{ $labels.%sK8S_Namespace }}", dir) }
	case flowslatest.GroupingPerWorkload:
		labelRoots = []string{"K8S_Namespace", "K8S_OwnerName", "K8S_OwnerType"}
		textFunc = func(dir string) string {
			return fmt.Sprintf("workload={{ $labels.%sK8S_OwnerName }} ({{ $labels.%sK8S_OwnerType }})", dir, dir)
		}
	}
	var labels []string
	var strFrom, strTo string
	if alert.GroupingDirection == flowslatest.GroupingBySource || alert.GroupingDirection == flowslatest.GroupingBySourceAndDestination {
		for _, lblRoot := range labelRoots {
			labels = append(labels, "Src"+lblRoot)
		}
		if textFunc != nil {
			strFrom = fmt.Sprintf(" from [%s]", textFunc("Src"))
		}
	}
	if alert.GroupingDirection == flowslatest.GroupingByDestination || alert.GroupingDirection == flowslatest.GroupingBySourceAndDestination {
		for _, lblRoot := range labelRoots {
			labels = append(labels, "Dst"+lblRoot)
		}
		if textFunc != nil {
			strTo = fmt.Sprintf(" to [%s]", textFunc("Dst"))
		}
	}

	return labels, strFrom + strTo
}

func alertNoFlows() *monitoringv1.Rule {
	d := monitoringv1.Duration("10m")

	// Not receiving flows
	return &monitoringv1.Rule{
		Alert: string(flowslatest.AlertNoFlows),
		Annotations: map[string]string{
			"description":    "NetObserv flowlogs-pipeline is not receiving any flow, this is either a connection issue with the agent, or an agent issue",
			"summary":        "NetObserv flowlogs-pipeline is not receiving any flow",
			healthAnnotation: "{}",
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
			"description":    "NetObserv flowlogs-pipeline is dropping flows because of Loki errors, Loki may be down or having issues ingesting every flows. Please check Loki and flowlogs-pipeline logs.",
			"summary":        "NetObserv flowlogs-pipeline is dropping flows because of Loki errors",
			healthAnnotation: "{}",
		},
		Expr: intstr.FromString("sum(rate(netobserv_loki_dropped_entries_total[1m])) > 0"),
		For:  &d,
		Labels: map[string]string{
			"severity": "warning",
			"app":      "netobserv",
		},
	}
}
