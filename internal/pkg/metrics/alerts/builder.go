package alerts

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

type srcOrDst string

const (
	asSource srcOrDst = "Src"
	asDest   srcOrDst = "Dst"
)

type ruleBuilder struct {
	template        flowslatest.AlertTemplate
	alert           *flowslatest.AlertVariant
	enabledMetrics  []string
	side            srcOrDst
	severity        string
	threshold       string
	upperThreshold  string
	upperValueRange string
	trafficLink     *trafficLink
	extraLinks      []link
	duration        monitoringv1.Duration
}

type trafficLink struct {
	ExtraFilter       string `json:"extraFilter"`
	BackAndForth      bool   `json:"backAndForth"`
	FilterDestination bool   `json:"filterDestination"`
}

type link struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func BuildRules(ctx context.Context, fc *flowslatest.FlowCollectorSpec) []monitoringv1.Rule {
	log := log.FromContext(ctx)
	rules := []monitoringv1.Rule{}

	alerts := fc.GetFLPAlerts()
	metrics := fc.GetIncludeList()
	for _, alert := range alerts {
		if ok, _ := alert.IsAllowed(fc); !ok {
			continue
		}
		for _, variant := range alert.Variants {
			if r, err := convertToRules(alert.Template, &variant, metrics); err != nil {
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

func convertToRules(template flowslatest.AlertTemplate, alert *flowslatest.AlertVariant, enabledMetrics []string) ([]monitoringv1.Rule, error) {
	var rules []monitoringv1.Rule
	var upperThreshold string
	sides := []srcOrDst{asSource, asDest}
	if alert.GroupBy == "" {
		// No side for global group
		sides = []srcOrDst{""}
	}
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
			for _, side := range sides {
				rb := ruleBuilder{
					template:       template,
					alert:          alert,
					enabledMetrics: enabledMetrics,
					side:           side,
					severity:       st.s,
					threshold:      st.t,
					upperThreshold: upperThreshold,
					duration:       monitoringv1.Duration("5m"),
				}
				if r, err := rb.convertToRule(); err != nil {
					return nil, err
				} else if r != nil {
					rules = append(rules, *r)
				}
			}
			upperThreshold = st.t
		}
	}
	return rules, nil
}

func (rb *ruleBuilder) convertToRule() (*monitoringv1.Rule, error) {
	switch rb.template {
	case flowslatest.AlertPacketDropsByDevice:
		return rb.deviceDrops()
	case flowslatest.AlertPacketDropsByKernel:
		return rb.kernelDrops()
	case flowslatest.AlertIPsecErrors:
		return rb.ipsecErrors()
	case flowslatest.AlertDNSErrors:
		return rb.dnsErrors()
	case flowslatest.AlertDNSNxDomain:
		return rb.dnsNxDomainErrors()
	case flowslatest.AlertNetpolDenied:
		return rb.netpolDenied()
	case flowslatest.AlertLatencyHighTrend:
		return rb.latencyTrend()
	case flowslatest.AlertCrossAZ, flowslatest.AlertExternalEgressHighTrend, flowslatest.AlertExternalIngressHighTrend:
		return nil, nil // TODO
	case flowslatest.AlertLokiError, flowslatest.AlertNoFlows:
		// error
	}
	return nil, fmt.Errorf("unknown alert template: %s", rb.template)
}

func (rb *ruleBuilder) additionalDescription() string {
	return fmt.Sprintf("You can turn off this alert by adding '%s' to spec.processor.metrics.disableAlerts in FlowCollector, or reconfigure it via spec.processor.metrics.alerts.", rb.template)
}

func (rb *ruleBuilder) createRule(promQL, summary, description string) (*monitoringv1.Rule, error) {
	bAnnot, err := rb.buildHealthAnnotation(nil)
	if err != nil {
		return nil, err
	}

	var gr string
	if rb.alert.GroupBy != "" {
		gr = "Per" + string(rb.side) + string(rb.alert.GroupBy)
	}
	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("%s_%s%s", rb.template, gr, strings.ToUpper(rb.severity[:1])+rb.severity[1:]),
		Annotations: map[string]string{
			"description":                 description,
			"summary":                     summary,
			"netobserv_io_network_health": string(bAnnot),
		},
		Expr:   intstr.FromString(promQL),
		For:    &rb.duration,
		Labels: buildLabels(rb.severity, true),
	}, nil
}

func (rb *ruleBuilder) getAlertLegend() string {
	var sideText string
	switch rb.side {
	case asSource:
		sideText = "source "
	case asDest:
		sideText = "dest. "
	}
	switch rb.alert.GroupBy {
	case flowslatest.GroupByNode:
		return " [" + sideText + "node={{ $labels.node }}]"
	case flowslatest.GroupByNamespace:
		return " [" + sideText + "namespace={{ $labels.namespace }}]"
	case flowslatest.GroupByWorkload:
		return " [" + sideText + "workload={{ $labels.workload }} ({{ $labels.kind }})]"
	}
	return ""
}

func (rb *ruleBuilder) buildHealthAnnotation(override map[string]any) ([]byte, error) {
	// The health annotation contains json-encoded information used in console plugin display
	annotation := map[string]any{
		"threshold": rb.threshold,
		"unit":      "%",
	}
	if rb.upperValueRange != "" {
		annotation["upperBound"] = rb.upperValueRange
	}
	if rb.trafficLink != nil {
		annotation["trafficLink"] = rb.trafficLink
	}
	if len(rb.extraLinks) > 0 {
		annotation["links"] = rb.extraLinks
	}
	switch rb.alert.GroupBy {
	case flowslatest.GroupByNode:
		annotation["nodeLabels"] = []string{"node"}
	case flowslatest.GroupByNamespace, flowslatest.GroupByWorkload:
		annotation["namespaceLabels"] = []string{"namespace"}
	}
	for k, v := range override {
		annotation[k] = v
	}
	bAnnot, err := json.Marshal(annotation)
	if err != nil {
		return nil, fmt.Errorf("cannot encode alert annotation [template=%s]: %w", rb.template, err)
	}
	return bAnnot, nil
}

func buildLabels(severity string, forHealth bool) map[string]string {
	m := map[string]string{
		"severity": strings.ToLower(severity),
		"app":      "netobserv", // means that the rule is created by netobserv
	}
	if forHealth {
		m["netobserv"] = "true" // means that the rule should be fetched by netobserv console plugin for health
	}
	return m
}

func (rb *ruleBuilder) buildLabelFilter(additionalFilter string) string {
	var filters []string

	// Build label matchers to filter out metrics where K8s labels don't exist or are empty
	// This prevents alerts from firing with empty namespace/workload/node labels
	switch rb.alert.GroupBy {
	case flowslatest.GroupByNode:
		filters = append(filters, string(rb.side)+`K8S_HostName!=""`)
	case flowslatest.GroupByNamespace:
		filters = append(filters, string(rb.side)+`K8S_Namespace!=""`)
	case flowslatest.GroupByWorkload:
		filters = append(filters, string(rb.side)+`K8S_Namespace!=""`)
		filters = append(filters, string(rb.side)+`K8S_OwnerName!=""`)
		filters = append(filters, string(rb.side)+`K8S_OwnerType!=""`)
	}

	// Add additional business logic filters
	if additionalFilter != "" {
		filters = append(filters, additionalFilter)
	}

	if len(filters) == 0 {
		return ""
	}
	return "{" + strings.Join(filters, ",") + "}"
}

func (rb *ruleBuilder) getMetricsForAlert() (string, string) {
	var reqMetric1, reqMetric2 string
	reqMetrics1, reqMetrics2 := flowslatest.GetElligibleMetricsForAlert(rb.template, rb.alert)
	if len(reqMetrics1) > 0 {
		reqMetric1 = flowslatest.GetFirstRequiredMetrics(reqMetrics1, rb.enabledMetrics)
	}
	if len(reqMetrics2) > 0 {
		reqMetric2 = flowslatest.GetFirstRequiredMetrics(reqMetrics2, rb.enabledMetrics)
	}
	return reqMetric1, reqMetric2
}
