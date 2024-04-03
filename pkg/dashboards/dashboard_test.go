package dashboards

import (
	"testing"

	"github.com/netobserv/network-observability-operator/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestCreateFlowMetricsDashboard_All(t *testing.T) {
	assert := assert.New(t)

	defs := metrics.GetDefinitions(metrics.GetAllNames())
	js := CreateFlowMetricsDashboards(defs)

	d, err := FromBytes([]byte(js))
	assert.NoError(err)

	assert.Equal("NetObserv", d.Title)

	assert.Equal([]string{"Traffic rates", "TCP latencies", "Byte and packet drops", "DNS"}, d.Titles())

	assert.Len(d.Rows[0].Panels, 32)

	p := d.FindPanel("Top egress traffic per node")
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal("topk(7, sum(rate(netobserv_node_egress_bytes_total{}[2m])) by (SrcK8S_HostName,DstK8S_HostName))", p.Targets[0].Expr)

	p = d.FindPanel("Top P50 DNS latency per node (ms)")
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal("topk(7, histogram_quantile(0.5, sum(rate(netobserv_node_dns_latency_seconds_bucket{}[2m])) by (le,SrcK8S_HostName,DstK8S_HostName))*1000 > 0)", p.Targets[0].Expr)

	p = d.FindPanel("Top P99 DNS latency per node (ms)")
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal("topk(7, histogram_quantile(0.99, sum(rate(netobserv_node_dns_latency_seconds_bucket{}[2m])) by (le,SrcK8S_HostName,DstK8S_HostName))*1000 > 0)", p.Targets[0].Expr)

	p = d.FindPanel("Top ingress traffic per app namespace")
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal(
		`topk(7, (sum(rate(netobserv_namespace_ingress_bytes_total{K8S_FlowLayer="app",SrcK8S_Namespace!=""}[2m])) by (SrcK8S_Namespace,DstK8S_Namespace))`+
			` or (sum(rate(netobserv_namespace_ingress_bytes_total{K8S_FlowLayer="app",DstK8S_Namespace!=""}[2m])) by (SrcK8S_Namespace,DstK8S_Namespace)))`,
		p.Targets[0].Expr,
	)
	p = d.FindPanel("Top ingress traffic per infra namespace")
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal(
		`topk(7, (sum(rate(netobserv_namespace_ingress_bytes_total{K8S_FlowLayer="infra",SrcK8S_Namespace!=""}[2m])) by (SrcK8S_Namespace,DstK8S_Namespace))`+
			` or (sum(rate(netobserv_namespace_ingress_bytes_total{K8S_FlowLayer="infra",DstK8S_Namespace!=""}[2m])) by (SrcK8S_Namespace,DstK8S_Namespace)))`,
		p.Targets[0].Expr,
	)

	p = d.FindPanel("Top P50 sRTT per infra namespace (ms)")
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal(
		`topk(7, (histogram_quantile(0.5, sum(rate(netobserv_namespace_rtt_seconds_bucket{K8S_FlowLayer="infra",SrcK8S_Namespace!=""}[2m])) by (le,SrcK8S_Namespace,DstK8S_Namespace))*1000 > 0)`+
			` or (histogram_quantile(0.5, sum(rate(netobserv_namespace_rtt_seconds_bucket{K8S_FlowLayer="infra",DstK8S_Namespace!=""}[2m])) by (le,SrcK8S_Namespace,DstK8S_Namespace))*1000 > 0))`,
		p.Targets[0].Expr,
	)

	p = d.FindPanel("Top ingress traffic per app workload")
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal(
		`topk(7, sum(rate(netobserv_workload_ingress_packets_total{K8S_FlowLayer="app"}[2m])) by (SrcK8S_Namespace,SrcK8S_OwnerName,DstK8S_Namespace,DstK8S_OwnerName))`,
		p.Targets[0].Expr,
	)
	p = d.FindPanel("Top ingress traffic per infra workload")
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal(
		`topk(7, sum(rate(netobserv_workload_ingress_packets_total{K8S_FlowLayer="infra"}[2m])) by (SrcK8S_Namespace,SrcK8S_OwnerName,DstK8S_Namespace,DstK8S_OwnerName))`,
		p.Targets[0].Expr,
	)
}

func TestCreateFlowMetricsDashboard_OnlyNodeIngressBytes(t *testing.T) {
	assert := assert.New(t)

	defs := metrics.GetDefinitions([]string{"node_ingress_bytes_total"})
	js := CreateFlowMetricsDashboards(defs)

	d, err := FromBytes([]byte(js))
	assert.NoError(err)

	assert.Equal("NetObserv", d.Title)
	assert.Len(d.Rows, 1)

	row := d.FindRow("Byte rate received per node")
	assert.NotNil(row)
	assert.Len(row.Panels, 1)
	assert.Equal("", row.Panels[0].Title)
	assert.Len(row.Panels[0].Targets, 1)
	assert.Contains(row.Panels[0].Targets[0].Expr, "label_replace(label_replace(topk(7,sum(rate(netobserv_node_ingress_bytes_total[2m])) by (SrcK8S_HostName,DstK8S_HostName))")
}

func TestCreateFlowMetricsDashboard_DefaultList(t *testing.T) {
	assert := assert.New(t)

	defs := metrics.GetDefinitions(metrics.DefaultIncludeList)
	js := CreateFlowMetricsDashboards(defs)

	d, err := FromBytes([]byte(js))
	assert.NoError(err)

	assert.Equal("NetObserv", d.Title)
	assert.Len(d.Rows, 7)

	row := d.FindRow("Byte rate received per node")
	assert.NotNil(row)
	assert.Len(row.Panels, 1)
	assert.Equal("", row.Panels[0].Title)
	assert.Len(row.Panels[0].Targets, 1)
	assert.Contains(row.Panels[0].Targets[0].Expr, "label_replace(label_replace(topk(7,sum(rate(netobserv_node_ingress_bytes_total[2m])) by (SrcK8S_HostName,DstK8S_HostName))")

	row = d.FindRow("Byte rate received per namespace")
	assert.NotNil(row)
	assert.Len(row.Panels, 2)
	assert.Equal("Applications", row.Panels[0].Title)
	assert.Equal("Infrastructure", row.Panels[1].Title)
	assert.Len(row.Panels[0].Targets, 1)
	// Make sure netobserv_namespace_ingress_bytes_total was replaced with netobserv_workload_ingress_bytes_total
	assert.Contains(row.Panels[0].Targets[0].Expr,
		`label_replace(label_replace(topk(7,sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace!~"|netobserv|openshift.*"}[2m]) or rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*",DstK8S_Namespace!~"|netobserv|openshift.*"}[2m])) by (SrcK8S_Namespace,DstK8S_Namespace))`,
	)
	assert.Contains(row.Panels[1].Targets[0].Expr,
		`label_replace(label_replace(topk(7,sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*"}[2m]) or rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace!~"netobserv|openshift.*",DstK8S_Namespace=~"netobserv|openshift.*"}[2m])) by (SrcK8S_Namespace,DstK8S_Namespace))`,
	)

	row = d.FindRow("Byte rate received per workload")
	assert.NotNil(row)
	assert.Len(row.Panels, 2)
	assert.Equal("Applications", row.Panels[0].Title)
	assert.Equal("Infrastructure", row.Panels[1].Title)
	assert.Len(row.Panels[0].Targets, 1)
	assert.Contains(row.Panels[0].Targets[0].Expr,
		`label_replace(label_replace(topk(7,sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace!~"|netobserv|openshift.*"}[2m]) or rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*",DstK8S_Namespace!~"|netobserv|openshift.*"}[2m])) by (SrcK8S_Namespace,SrcK8S_OwnerName,DstK8S_Namespace,DstK8S_OwnerName))`,
	)
	assert.Contains(row.Panels[1].Targets[0].Expr,
		`label_replace(label_replace(topk(7,sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace=~"netobserv|openshift.*"}[2m]) or rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace!~"netobserv|openshift.*",DstK8S_Namespace=~"netobserv|openshift.*"}[2m])) by (SrcK8S_Namespace,SrcK8S_OwnerName,DstK8S_Namespace,DstK8S_OwnerName))`,
	)
}

func TestCreateHealthDashboard_Default(t *testing.T) {
	assert := assert.New(t)

	js, err := CreateHealthDashboard("netobserv")
	assert.NoError(err)

	d, err := FromBytes([]byte(js))
	assert.NoError(err)

	assert.Equal("NetObserv / Health", d.Title)
	assert.Equal([]string{"", "Flowlogs-pipeline statistics", "eBPF agent statistics", "Operator statistics", "Resource usage"}, d.Titles())

	// First row
	row := 0
	assert.Len(d.Rows[row].Panels, 4)
	assert.Equal("Flows per second", d.Rows[row].Panels[0].Title)
	assert.Len(d.Rows[row].Panels[0].Targets, 1)
	assert.Contains(d.Rows[row].Panels[0].Targets[0].Expr, "netobserv_ingest_flows_processed")
}
