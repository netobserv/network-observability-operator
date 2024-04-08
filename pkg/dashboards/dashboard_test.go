package dashboards

import (
	"testing"

	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestCreateFlowMetricsDashboard_All(t *testing.T) {
	assert := assert.New(t)

	defs := metrics.GetDefinitions(metrics.GetAllNames())
	js := CreateFlowMetricsDashboards(defs)

	d, err := FromBytes([]byte(js["Main"]))
	assert.NoError(err)

	assert.Equal("NetObserv / Main", d.Title)

	assert.Equal([]string{"", "Traffic rates", "TCP latencies", "Byte and packet drops", "DNS"}, d.Titles())

	assert.Len(d.Rows[0].Panels, 16)
	assert.Len(d.Rows[1].Panels, 20)

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

	p = d.FindNthPanel("Top ingress traffic per app workload", 2) // pps variant
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal(
		`topk(7, (sum(rate(netobserv_workload_ingress_packets_total{K8S_FlowLayer="app",SrcK8S_Namespace!=""}[2m])) by (SrcK8S_Namespace,SrcK8S_OwnerName,DstK8S_Namespace,DstK8S_OwnerName))`+
			` or (sum(rate(netobserv_workload_ingress_packets_total{K8S_FlowLayer="app",DstK8S_Namespace!=""}[2m])) by (SrcK8S_Namespace,SrcK8S_OwnerName,DstK8S_Namespace,DstK8S_OwnerName)))`,
		p.Targets[0].Expr,
	)
	p = d.FindNthPanel("Top ingress traffic per infra workload", 2) // pps variant
	assert.NotNil(p)
	assert.Len(p.Targets, 1)
	assert.Equal(
		`topk(7, (sum(rate(netobserv_workload_ingress_packets_total{K8S_FlowLayer="infra",SrcK8S_Namespace!=""}[2m])) by (SrcK8S_Namespace,SrcK8S_OwnerName,DstK8S_Namespace,DstK8S_OwnerName))`+
			` or (sum(rate(netobserv_workload_ingress_packets_total{K8S_FlowLayer="infra",DstK8S_Namespace!=""}[2m])) by (SrcK8S_Namespace,SrcK8S_OwnerName,DstK8S_Namespace,DstK8S_OwnerName)))`,
		p.Targets[0].Expr,
	)
}

func TestCreateFlowMetricsDashboard_OnlyNodeIngressBytes(t *testing.T) {
	assert := assert.New(t)

	defs := metrics.GetDefinitions([]string{"node_ingress_bytes_total"})
	js := CreateFlowMetricsDashboards(defs)

	d, err := FromBytes([]byte(js["Main"]))
	assert.NoError(err)

	assert.Equal("NetObserv / Main", d.Title)
	assert.Equal([]string{"", "Traffic rates"}, d.Titles())

	topRow := d.FindRow("")
	assert.Equal([]string{"Total ingress traffic"}, topRow.Titles())

	trafficRow := d.FindRow("Traffic rates")
	assert.Equal([]string{"Top ingress traffic per node"}, trafficRow.Titles())
}

func TestCreateFlowMetricsDashboard_DefaultList(t *testing.T) {
	assert := assert.New(t)

	defs := metrics.GetDefinitions(metrics.DefaultIncludeList)
	js := CreateFlowMetricsDashboards(defs)

	d, err := FromBytes([]byte(js["Main"]))
	assert.NoError(err)

	assert.Equal("NetObserv / Main", d.Title)
	assert.Equal([]string{"", "Traffic rates", "TCP latencies", "Byte and packet drops", "DNS"}, d.Titles())

	topRow := d.FindRow("")
	assert.Equal([]string{
		"Total ingress traffic",
		"TCP latency, p99",
		"Drops",
		"DNS latency, p99",
		"DNS error rate",
		"Infra ingress traffic",
		"Apps ingress traffic",
	}, topRow.Titles())

	trafficRow := d.FindRow("Traffic rates")
	assert.Equal([]string{
		"Top ingress traffic per node",
		"Top ingress traffic per infra namespace",
		"Top ingress traffic per app namespace",
		"Top ingress traffic per infra workload",
		"Top ingress traffic per app workload",
	}, trafficRow.Titles())

	rttRow := d.FindRow("TCP latencies")
	assert.Equal([]string{
		"Top P50 sRTT per infra namespace (ms)",
		"Top P50 sRTT per app namespace (ms)",
		"Top P99 sRTT per infra namespace (ms)",
		"Top P99 sRTT per app namespace (ms)",
	}, rttRow.Titles())

	dropsRow := d.FindRow("Byte and packet drops")
	assert.Equal([]string{
		"Top drops per infra namespace",
		"Top drops per app namespace",
	}, dropsRow.Titles())

	dnsRow := d.FindRow("DNS")
	assert.Equal([]string{
		"Top P50 DNS latency per infra namespace (ms)",
		"Top P50 DNS latency per app namespace (ms)",
		"Top P99 DNS latency per infra namespace (ms)",
		"Top P99 DNS latency per app namespace (ms)",
		"DNS error rate per infra namespace",
		"DNS error rate per app namespace",
	}, dnsRow.Titles())
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

func TestCreateCustomDashboard(t *testing.T) {
	assert := assert.New(t)

	js := CreateFlowMetricsDashboards([]metricslatest.FlowMetric{
		{
			Spec: metricslatest.FlowMetricSpec{
				MetricName: "my_metric",
				Charts: []metricslatest.Chart{
					{
						DashboardName: "Main",
						SectionName:   "My section",
						Title:         "My chart",
						Unit:          metricslatest.UnitBPS,
						Type:          metricslatest.ChartTypeSingleStat,
						Queries: []metricslatest.Query{
							{
								PromQL: `sum(rate($METRIC{label="foo"}[5m]))`,
								Legend: "",
							},
						},
					},
					{
						DashboardName: "Main",
						SectionName:   "My next section",
						Title:         "My next chart",
						Unit:          metricslatest.UnitBPS,
						Type:          metricslatest.ChartTypeLine,
						Queries: []metricslatest.Query{
							{
								PromQL: `sum(rate($METRIC{label="foo"}[5m])) by (lbl1,lbl2)`,
								Legend: "{{lbl1}}: {{lbl2}}",
							},
						},
					},
				},
			},
		},
		{
			Spec: metricslatest.FlowMetricSpec{
				MetricName: "my_metric",
				Charts: []metricslatest.Chart{
					{
						DashboardName: "My other dashboard",
						SectionName:   "Other section",
						Title:         "Other chart",
						Unit:          metricslatest.UnitBPS,
						Type:          metricslatest.ChartTypeLine,
						Queries: []metricslatest.Query{
							{
								PromQL: `sum(rate($METRIC{label="foo"}[5m])) by (lbl1,lbl2)`,
								Legend: "{{lbl1}}: {{lbl2}}",
							},
						},
					},
				},
			},
		},
	})

	d, err := FromBytes([]byte(js["Main"]))
	assert.NoError(err)

	assert.Equal("NetObserv / Main", d.Title)
	assert.Equal([]string{"My section", "My next section"}, d.Titles())

	r1 := d.FindRow("My section")
	assert.Equal([]string{"My chart"}, r1.Titles())
	assert.Equal(Panel{
		Title:  "My chart",
		Type:   "singlestat",
		Span:   3,
		Format: "Bps",
		Targets: []Target{
			{
				Expr:         "sum(rate(netobserv_my_metric{label=\"foo\"}[5m]))",
				LegendFormat: "",
			},
		},
	}, r1.Panels[0])

	r2 := d.FindRow("My next section")
	assert.Equal([]string{"My next chart"}, r2.Titles())
	assert.Equal(Panel{
		Title:  "My next chart",
		Type:   "graph",
		Span:   4,
		Format: "Bps",
		Targets: []Target{
			{
				Expr:         "topk(7, sum(rate(netobserv_my_metric{label=\"foo\"}[5m])) by (lbl1,lbl2))",
				LegendFormat: "{{lbl1}}: {{lbl2}}",
			},
		},
	}, r2.Panels[0])

	d, err = FromBytes([]byte(js["My other dashboard"]))
	assert.NoError(err)

	assert.Equal("NetObserv / My other dashboard", d.Title)
	assert.Equal([]string{"Other section"}, d.Titles())

	r1 = d.FindRow("Other section")
	assert.Equal([]string{"Other chart"}, r1.Titles())
	assert.Equal(Panel{
		Title:  "Other chart",
		Type:   "graph",
		Span:   4,
		Format: "Bps",
		Targets: []Target{
			{
				Expr:         "topk(7, sum(rate(netobserv_my_metric{label=\"foo\"}[5m])) by (lbl1,lbl2))",
				LegendFormat: "{{lbl1}}: {{lbl2}}",
			},
		},
	}, r1.Panels[0])
}
