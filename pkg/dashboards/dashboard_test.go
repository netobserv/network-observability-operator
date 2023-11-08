package dashboards

import (
	"testing"

	"github.com/netobserv/network-observability-operator/pkg/metrics"
	"github.com/netobserv/network-observability-operator/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestCreateFlowMetricsDashboard_All(t *testing.T) {
	assert := assert.New(t)

	js, err := CreateFlowMetricsDashboard("netobserv", metrics.GetAllNames())
	assert.NoError(err)

	d, err := test.DashboardFromBytes([]byte(js))
	assert.NoError(err)

	assert.Equal("NetObserv", d.Title)
	assert.Len(d.Rows, 18)

	// First row
	row := 0
	assert.Equal("Top byte rates sent per source and destination nodes", d.Rows[row].Title)
	assert.Len(d.Rows[row].Panels, 1)
	assert.Equal("", d.Rows[row].Panels[0].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 1)
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr, "label_replace(label_replace(topk(10,sum(rate(netobserv_node_egress_bytes_total[1m])) by (SrcK8S_HostName, DstK8S_HostName))")

	// 8th row
	row = 7
	assert.Equal("Top byte rates received per source and destination namespaces", d.Rows[row].Title)
	assert.Len(d.Rows[row].Panels, 2)
	assert.Equal("Applications", d.Rows[row].Panels[0].Title)
	assert.Equal("Infrastructure", d.Rows[row].Panels[1].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 1)
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr,
		`label_replace(label_replace(topk(10,sum(rate(netobserv_namespace_ingress_bytes_total{SrcK8S_Namespace!~"|netobserv|openshift.*"}[1m]) or rate(netobserv_namespace_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*",DstK8S_Namespace!~"|netobserv|openshift.*"}[1m])) by (SrcK8S_Namespace, DstK8S_Namespace))`,
	)
	assert.Contains(d.Rows[row].Panels[1].Targets[0].Expr,
		`label_replace(label_replace(topk(10,sum(rate(netobserv_namespace_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*"}[1m]) or rate(netobserv_namespace_ingress_bytes_total{SrcK8S_Namespace!~"netobserv|openshift.*",DstK8S_Namespace=~"netobserv|openshift.*"}[1m])) by (SrcK8S_Namespace, DstK8S_Namespace))`,
	)

	// 16th row
	row = 15
	assert.Equal("Top packet rates received per source and destination workloads", d.Rows[row].Title)
	assert.Len(d.Rows[row].Panels, 2)
	assert.Equal("Applications", d.Rows[row].Panels[0].Title)
	assert.Equal("Infrastructure", d.Rows[row].Panels[1].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 1)
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr,
		`label_replace(label_replace(topk(10,sum(rate(netobserv_workload_ingress_packets_total{SrcK8S_Namespace!~"|netobserv|openshift.*"}[1m]) or rate(netobserv_workload_ingress_packets_total{SrcK8S_Namespace=~"netobserv|openshift.*",DstK8S_Namespace!~"|netobserv|openshift.*"}[1m])) by (SrcK8S_Namespace, SrcK8S_OwnerName, DstK8S_Namespace, DstK8S_OwnerName))`,
	)
	assert.Contains(d.Rows[row].Panels[1].Targets[0].Expr,
		`label_replace(label_replace(topk(10,sum(rate(netobserv_workload_ingress_packets_total{SrcK8S_Namespace=~"netobserv|openshift.*"}[1m]) or rate(netobserv_workload_ingress_packets_total{SrcK8S_Namespace!~"netobserv|openshift.*",DstK8S_Namespace=~"netobserv|openshift.*"}[1m])) by (SrcK8S_Namespace, SrcK8S_OwnerName, DstK8S_Namespace, DstK8S_OwnerName))`,
	)
}

func TestCreateFlowMetricsDashboard_OnlyNodeIngressBytes(t *testing.T) {
	assert := assert.New(t)

	js, err := CreateFlowMetricsDashboard("netobserv", []string{"node_ingress_bytes_total"})
	assert.NoError(err)

	d, err := test.DashboardFromBytes([]byte(js))
	assert.NoError(err)

	assert.Equal("NetObserv", d.Title)
	assert.Len(d.Rows, 1)

	// First row
	row := 0
	assert.Equal("Top byte rates received per source and destination nodes", d.Rows[row].Title)
	assert.Len(d.Rows[row].Panels, 1)
	assert.Equal("", d.Rows[row].Panels[0].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 1)
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr, "label_replace(label_replace(topk(10,sum(rate(netobserv_node_ingress_bytes_total[1m])) by (SrcK8S_HostName, DstK8S_HostName))")
}

func TestCreateFlowMetricsDashboard_DefaultList(t *testing.T) {
	assert := assert.New(t)

	js, err := CreateFlowMetricsDashboard("netobserv", metrics.DefaultIncludeList)
	assert.NoError(err)

	d, err := test.DashboardFromBytes([]byte(js))
	assert.NoError(err)

	assert.Equal("NetObserv", d.Title)
	assert.Len(d.Rows, 3)

	// First row
	row := 0
	assert.Equal("Top byte rates received per source and destination nodes", d.Rows[row].Title)
	assert.Len(d.Rows[row].Panels, 1)
	assert.Equal("", d.Rows[row].Panels[0].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 1)
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr, "label_replace(label_replace(topk(10,sum(rate(netobserv_node_ingress_bytes_total[1m])) by (SrcK8S_HostName, DstK8S_HostName))")

	// 2nd row
	row = 1
	assert.Equal("Top byte rates received per source and destination namespaces", d.Rows[row].Title)
	assert.Len(d.Rows[row].Panels, 2)
	assert.Equal("Applications", d.Rows[row].Panels[0].Title)
	assert.Equal("Infrastructure", d.Rows[row].Panels[1].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 1)
	// Make sure netobserv_namespace_ingress_bytes_total was replaced with netobserv_workload_ingress_bytes_total
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr,
		`label_replace(label_replace(topk(10,sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace!~"|netobserv|openshift.*"}[1m]) or rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*",DstK8S_Namespace!~"|netobserv|openshift.*"}[1m])) by (SrcK8S_Namespace, DstK8S_Namespace))`,
	)
	assert.Contains(d.Rows[row].Panels[1].Targets[0].Expr,
		`label_replace(label_replace(topk(10,sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*"}[1m]) or rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace!~"netobserv|openshift.*",DstK8S_Namespace=~"netobserv|openshift.*"}[1m])) by (SrcK8S_Namespace, DstK8S_Namespace))`,
	)

	// 3rd row
	row = 2
	assert.Equal("Top byte rates received per source and destination workloads", d.Rows[row].Title)
	assert.Len(d.Rows[row].Panels, 2)
	assert.Equal("Applications", d.Rows[row].Panels[0].Title)
	assert.Equal("Infrastructure", d.Rows[row].Panels[1].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 1)
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr,
		`label_replace(label_replace(topk(10,sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace!~"|netobserv|openshift.*"}[1m]) or rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*",DstK8S_Namespace!~"|netobserv|openshift.*"}[1m])) by (SrcK8S_Namespace, SrcK8S_OwnerName, DstK8S_Namespace, DstK8S_OwnerName))`,
	)
	assert.Contains(d.Rows[row].Panels[1].Targets[0].Expr,
		`label_replace(label_replace(topk(10,sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*"}[1m]) or rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace!~"netobserv|openshift.*",DstK8S_Namespace=~"netobserv|openshift.*"}[1m])) by (SrcK8S_Namespace, SrcK8S_OwnerName, DstK8S_Namespace, DstK8S_OwnerName))`,
	)
}

func TestCreateHealthDashboard_Default(t *testing.T) {
	assert := assert.New(t)

	js, err := CreateHealthDashboard("netobserv", metrics.DefaultIncludeList)
	assert.NoError(err)

	d, err := test.DashboardFromBytes([]byte(js))
	assert.NoError(err)

	assert.Equal("NetObserv / Health", d.Title)
	assert.Equal([]string{
		"Flows",
		"Flows Overhead",
		"Top flow rates per source and destination namespaces",
		"Agents",
		"Processor",
		"Operator",
	}, d.Titles())

	// First row
	row := 0
	assert.Len(d.Rows[row].Panels, 1)
	assert.Equal("Rates", d.Rows[row].Panels[0].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 3)
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr, "netobserv_ingest_flows_processed")

	// 3rd row
	row = 2
	assert.Len(d.Rows[row].Panels, 2)
	assert.Equal("Applications", d.Rows[row].Panels[0].Title)
	assert.Equal("Infrastructure", d.Rows[row].Panels[1].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 1)
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr, "netobserv_namespace_flows_total")
}

func TestCreateHealthDashboard_NoFlowMetric(t *testing.T) {
	assert := assert.New(t)

	js, err := CreateHealthDashboard("netobserv", []string{})
	assert.NoError(err)

	d, err := test.DashboardFromBytes([]byte(js))
	assert.NoError(err)

	assert.Equal("NetObserv / Health", d.Title)
	assert.Equal([]string{
		"Flows",
		"Agents",
		"Processor",
		"Operator",
	}, d.Titles())
}
