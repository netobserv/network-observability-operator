package alerts

import (
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/stretchr/testify/assert"
)

func TestSumBy(t *testing.T) {
	pql := sumBy("rate(my_metric[1m])", flowslatest.GroupByNode, asSource, "")
	assert.Equal(t,
		`sum(label_replace(rate(my_metric[1m]), "node", "$1", "SrcK8S_HostName", "(.*)")) by (node)`,
		pql,
	)

	pql = sumBy("rate(my_metric[1m])", flowslatest.GroupByWorkload, asDest, "")
	assert.Equal(t,
		`sum(label_replace(label_replace(label_replace(rate(my_metric[1m]), "namespace", "$1", "DstK8S_Namespace", "(.*)"), "workload", "$1", "DstK8S_OwnerName", "(.*)"), "kind", "$1", "DstK8S_OwnerType", "(.*)")) by (namespace,workload,kind)`,
		pql,
	)

	pql = sumBy("rate(my_metric[1m])", "", "", "")
	assert.Equal(t, `sum(rate(my_metric[1m]))`, pql)
}

func TestPercentagePromQL(t *testing.T) {
	pql := percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "", "")
	assert.Equal(t, "100 * (sum(rate(my_metric[1m]))) / (sum(rate(my_total[1m]))) > 10", pql)

	pql = percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "20", "")
	assert.Equal(t, "100 * (sum(rate(my_metric[1m]))) / (sum(rate(my_total[1m]))) > 10 < 20", pql)

	pql = percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "20", "2")
	assert.Equal(t, "100 * (sum(rate(my_metric[1m]))) / (sum(rate(my_total[1m])) > 2) > 10 < 20", pql)
}
