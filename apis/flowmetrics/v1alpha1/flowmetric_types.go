/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MetricType string
type FilterMatchType string
type FlowDirection string

const (
	CounterMetric   MetricType = "Counter"
	HistogramMetric MetricType = "Histogram"
	// Note: we don't expose gauge on purpose to avoid configuration mistake related to gauge limitation.
	// 99% of times, "counter" or "histogram" should be the ones to use. We can eventually revisit later.
	MatchEqual    FilterMatchType = "Equal"
	MatchNotEqual FilterMatchType = "NotEqual"
	MatchPresence FilterMatchType = "Presence"
	MatchAbsence  FilterMatchType = "Absence"
	MatchRegex    FilterMatchType = "MatchRegex"
	MatchNotRegex FilterMatchType = "NotMatchRegex"
	Egress        FlowDirection   = "Egress"
	Ingress       FlowDirection   = "Ingress"
	AnyDirection  FlowDirection   = "Any"
)

type MetricFilter struct {
	// Name of the field to filter on
	// +required
	Field string `json:"field"`

	// Value to filter on. When `matchType` is `Equal` or `NotEqual`, you can use field injection with `$(SomeField)` to refer to any other field of the flow.
	// +optional
	Value string `json:"value"`

	// Type of matching to apply
	// +kubebuilder:validation:Enum:="Equal";"NotEqual";"Presence";"Absence";"MatchRegex";"NotMatchRegex"
	// +kubebuilder:default:="Equal"
	MatchType FilterMatchType `json:"matchType"`
}

// FlowMetricSpec defines the desired state of FlowMetric
// The provided API allows you to customize these metrics according to your needs.<br>
// When adding new metrics or modifying existing labels, you must carefully monitor the memory
// usage of Prometheus workloads as this could potentially have a high impact. Cf https://rhobs-handbook.netlify.app/products/openshiftmonitoring/telemetry.md/#what-is-the-cardinality-of-a-metric<br>
// To check the cardinality of all NetObserv metrics, run as `promql`: `count({__name__=~"netobserv.*"}) by (__name__)`.
type FlowMetricSpec struct {
	// Name of the metric. In Prometheus, it is automatically prefixed with "netobserv_".
	// +required
	MetricName string `json:"metricName"`

	// Metric type: "Counter" or "Histogram".
	// Use "Counter" for any value that increases over time and on which you can compute a rate, such as Bytes or Packets.
	// Use "Histogram" for any value that must be sampled independently, such as latencies.
	// +kubebuilder:validation:Enum:="Counter";"Histogram"
	// +required
	Type MetricType `json:"type"`

	// `valueField` is the flow field that must be used as a value for this metric. This field must hold numeric values.
	// Leave empty to count flows rather than a specific value per flow.
	// Refer to the documentation for the list of available fields: https://docs.openshift.com/container-platform/latest/observability/network_observability/json-flows-format-reference.html.
	// +optional
	ValueField string `json:"valueField,omitempty"`

	// `filters` is a list of fields and values used to restrict which flows are taken into account.
	// Refer to the documentation for the list of available fields: https://docs.openshift.com/container-platform/latest/observability/network_observability/json-flows-format-reference.html.
	// +optional
	Filters []MetricFilter `json:"filters"`

	// `labels` is a list of fields that should be used as Prometheus labels, also known as dimensions.
	// From choosing labels results the level of granularity of this metric, and the available aggregations at query time.
	// It must be done carefully as it impacts the metric cardinality (cf https://rhobs-handbook.netlify.app/products/openshiftmonitoring/telemetry.md/#what-is-the-cardinality-of-a-metric).
	// In general, avoid setting very high cardinality labels such as IP or MAC addresses.
	// "SrcK8S_OwnerName" or "DstK8S_OwnerName" should be preferred over "SrcK8S_Name" or "DstK8S_Name" as much as possible.
	// Refer to the documentation for the list of available fields: https://docs.openshift.com/container-platform/latest/observability/network_observability/json-flows-format-reference.html.
	// +optional
	Labels []string `json:"labels"`

	// `flatten` is a list of array-type fields that must be flattened, such as Interfaces or NetworkEvents. Flattened fields generate one metric per item in that field.
	// For instance, when flattening `Interfaces` on a bytes counter, a flow having Interfaces [br-ex, ens5] increases one counter for `br-ex` and another for `ens5`.
	// +optional
	Flatten []string `json:"flatten"`

	// Set the `remap` property to use different names for the generated metric labels than the flow fields. Use the origin flow fields as keys, and the desired label names as values.
	// +optional
	Remap map[string]string `json:"remap"`

	// Filter for ingress, egress or any direction flows.
	// When set to `Ingress`, it is equivalent to adding the regular expression filter on `FlowDirection`: `0|2`.
	// When set to `Egress`, it is equivalent to adding the regular expression filter on `FlowDirection`: `1|2`.
	// +kubebuilder:validation:Enum:="Any";"Egress";"Ingress"
	// +kubebuilder:default:="Any"
	// +optional
	Direction FlowDirection `json:"direction,omitempty"`

	// A list of buckets to use when `type` is "Histogram". The list must be parsable as floats. When not set, Prometheus default buckets are used.
	// +optional
	Buckets []string `json:"buckets,omitempty"`

	// When nonzero, scale factor (divider) of the value. Metric value = Flow value / Divider.
	// +optional
	Divider string `json:"divider"`

	// Charts configuration, for the OpenShift Console in the administrator view, Dashboards menu.
	// +optional
	Charts []Chart `json:"charts,omitempty"`
}

type Unit string
type ChartType string

const (
	UnitBytes           Unit      = "bytes"
	UnitSeconds         Unit      = "seconds"
	UnitBPS             Unit      = "Bps"
	UnitPPS             Unit      = "pps"
	UnitPercent         Unit      = "percent"
	ChartTypeSingleStat ChartType = "SingleStat"
	ChartTypeLine       ChartType = "Line"
	ChartTypeStackArea  ChartType = "StackArea"
)

// Configures charts / dashboard generation associated to a metric
type Chart struct {
	// Name of the containing dashboard. If this name does not refer to an existing dashboard, a new dashboard is created.
	// +kubebuilder:default:="Main"
	DashboardName string `json:"dashboardName"`

	// Name of the containing dashboard section. If this name does not refer to an existing section, a new section is created.
	// If `sectionName` is omitted or empty, the chart is placed in the global top section.
	// +optional
	SectionName string `json:"sectionName,omitempty"`

	// Title of the chart.
	// +required
	Title string `json:"title"`

	// Unit of this chart. Only a few units are currently supported. Leave empty to use generic number.
	// +kubebuilder:validation:Enum:="bytes";"seconds";"Bps";"pps";"percent";""
	// +optional
	Unit Unit `json:"unit,omitempty"`

	// Type of the chart.
	// +kubebuilder:validation:Enum:="SingleStat";"Line";"StackArea"
	// +required
	Type ChartType `json:"type"`

	// List of queries to be displayed on this chart. If `type` is `SingleStat` and multiple queries are provided,
	// this chart is automatically expanded in several panels (one per query).
	// +required
	Queries []Query `json:"queries"`
}

// Configures PromQL queries
type Query struct {
	// The `promQL` query to be run against Prometheus. If the chart `type` is `SingleStat`, this query should only return
	// a single timeseries. For other types, a top 7 is displayed.
	// You can use `$METRIC` to refer to the metric defined in this resource. For example: `sum(rate($METRIC[2m]))`.
	// To learn more about `promQL`, refer to the Prometheus documentation: https://prometheus.io/docs/prometheus/latest/querying/basics/
	// +required
	PromQL string `json:"promQL"`

	// The query legend that applies to each timeseries represented in this chart. When multiple timeseries are displayed, you should set a legend
	// that distinguishes each of them. It can be done with the following format: `{{ Label }}`. For example, if the `promQL` groups timeseries per
	// label such as: `sum(rate($METRIC[2m])) by (Label1, Label2)`, you may write as the legend: `Label1={{ Label1 }}, Label2={{ Label2 }}`.
	// +required
	Legend string `json:"legend"`

	// Top N series to display per timestamp. Does not apply to `SingleStat` chart type.
	// +kubebuilder:default:=7
	// +kubebuilder:validation:Minimum=1
	// +required
	Top int `json:"top"`
}

// FlowMetricStatus defines the observed state of FlowMetric
type FlowMetricStatus struct {
	// `conditions` represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Metric Name",type="string",JSONPath=`.spec.metricName`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Cardinality",type="string",JSONPath=`.status.conditions[?(@.type=="CardinalityOK")].reason`
// FlowMetric is the API allowing to create custom metrics from the collected flow logs.
type FlowMetric struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlowMetricSpec   `json:"spec,omitempty"`
	Status FlowMetricStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlowMetricList contains a list of FlowMetric
type FlowMetricList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlowMetric `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlowMetric{}, &FlowMetricList{})
}
