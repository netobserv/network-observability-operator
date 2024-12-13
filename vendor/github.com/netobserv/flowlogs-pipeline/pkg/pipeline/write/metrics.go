package write

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	*operational.Metrics
	recordsWritten prometheus.Counter
}

func newMetrics(opMetrics *operational.Metrics, stage string) *metrics {
	return &metrics{
		Metrics:        opMetrics,
		recordsWritten: opMetrics.CreateRecordsWrittenCounter(stage),
	}
}
