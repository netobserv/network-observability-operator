package helper

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dashboard struct {
	Rows []struct {
		Panels []struct {
			Targets []struct {
				Expr         string `json:"expr"`
				LegendFormat string `json:"legendFormat"`
			} `json:"targets"`
			Title string `json:"title"`
		} `json:"panels"`
		Title string `json:"title"`
	} `json:"rows"`
	Title string `json:"title"`
}

func TestCreateFlowMetricsDashboard_All(t *testing.T) {
	assert := assert.New(t)

	js, err := CreateFlowMetricsDashboard("netobserv", []string{})
	assert.NoError(err)

	var d dashboard
	err = json.Unmarshal([]byte(js), &d)
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

	js, err := CreateFlowMetricsDashboard("netobserv", []string{
		metricTagNamespaces,
		metricTagWorkloads,
		metricTagEgress,
		metricTagPackets,
		metricTagPktsDropBytes,
		metricTagPktsDropPackets})
	assert.NoError(err)

	var d dashboard
	err = json.Unmarshal([]byte(js), &d)
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

func TestCreateFlowMetricsDashboard_RemoveByMetricName(t *testing.T) {
	assert := assert.New(t)

	js, err := CreateFlowMetricsDashboard("netobserv", []string{
		metricTagNamespaces,
		metricTagWorkloads,
		"netobserv_node_egress_packets_total",
		"netobserv_node_ingress_packets_total",
		"netobserv_node_egress_bytes_total",
		metricTagPktsDropBytes,
		metricTagPktsDropPackets,
	})
	assert.NoError(err)

	var d dashboard
	err = json.Unmarshal([]byte(js), &d)
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

func TestCreateFlowMetricsDashboard_DefaultIgnoreTags(t *testing.T) {
	assert := assert.New(t)

	js, err := CreateFlowMetricsDashboard("netobserv", []string{"egress", "packets", "namespaces", metricTagPktsDropBytes, metricTagPktsDropPackets})
	assert.NoError(err)

	var d dashboard
	err = json.Unmarshal([]byte(js), &d)
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
