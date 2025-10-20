package alerts

import (
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/stretchr/testify/assert"
)

func TestSumBy(t *testing.T) {
	// Test GroupByNode with source side - should filter by SrcK8S_HostName
	pql := sumBy("rate(my_metric[1m])", flowslatest.GroupByNode, asSource, "")
	assert.Equal(t,
		`sum(label_replace(rate(my_metric{SrcK8S_HostName!=""}[1m]), "node", "$1", "SrcK8S_HostName", "(.*)")) by (node)`,
		pql,
	)
	assert.Contains(t, pql, `SrcK8S_HostName!=""`, "should filter out metrics without SrcK8S_HostName label")

	// Test GroupByWorkload with destination side - should filter by all three K8s labels
	pql = sumBy("rate(my_metric[1m])", flowslatest.GroupByWorkload, asDest, "")
	assert.Equal(t,
		`sum(label_replace(label_replace(label_replace(rate(my_metric{DstK8S_Namespace!="",DstK8S_OwnerName!="",DstK8S_OwnerType!=""}[1m]), "namespace", "$1", "DstK8S_Namespace", "(.*)"), "workload", "$1", "DstK8S_OwnerName", "(.*)"), "kind", "$1", "DstK8S_OwnerType", "(.*)")) by (namespace,workload,kind)`,
		pql,
	)
	assert.Contains(t, pql, `DstK8S_Namespace!=""`, "should filter out metrics without DstK8S_Namespace label")
	assert.Contains(t, pql, `DstK8S_OwnerName!=""`, "should filter out metrics without DstK8S_OwnerName label")
	assert.Contains(t, pql, `DstK8S_OwnerType!=""`, "should filter out metrics without DstK8S_OwnerType label")

	// Test GroupByNamespace - should filter by K8S_Namespace
	pql = sumBy("rate(my_metric[1m])", flowslatest.GroupByNamespace, asDest, "")
	assert.Equal(t,
		`sum(label_replace(rate(my_metric{DstK8S_Namespace!=""}[1m]), "namespace", "$1", "DstK8S_Namespace", "(.*)")) by (namespace)`,
		pql,
	)
	assert.Contains(t, pql, `DstK8S_Namespace!=""`, "should filter out metrics with empty or missing namespace")

	// Test no grouping - should NOT add any filters
	pql = sumBy("rate(my_metric[1m])", "", "", "")
	assert.Equal(t, `sum(rate(my_metric[1m]))`, pql)
	assert.NotContains(t, pql, "K8S_", "should not add K8s label filters when not grouping")

	// Test with existing label selector (like DNS errors) - should merge filters
	pql = sumBy(`rate(my_metric{DnsFlagsResponseCode!="NoError"}[1m])`, flowslatest.GroupByNamespace, asDest, "")
	assert.Equal(t,
		`sum(label_replace(rate(my_metric{DstK8S_Namespace!="",DnsFlagsResponseCode!="NoError"}[1m]), "namespace", "$1", "DstK8S_Namespace", "(.*)")) by (namespace)`,
		pql,
	)
	assert.Contains(t, pql, `DstK8S_Namespace!="",DnsFlagsResponseCode!="NoError"`, "should merge new filter with existing labels")
}

func TestPercentagePromQL(t *testing.T) {
	pql := percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "", "")
	assert.Equal(t, "100 * (sum(rate(my_metric[1m]))) / (sum(rate(my_total[1m]))) > 10", pql)

	pql = percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "20", "")
	assert.Equal(t, "100 * (sum(rate(my_metric[1m]))) / (sum(rate(my_total[1m]))) > 10 < 20", pql)

	pql = percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "20", "2")
	assert.Equal(t, "100 * (sum(rate(my_metric[1m]))) / (sum(rate(my_total[1m])) > 2) > 10 < 20", pql)
}
