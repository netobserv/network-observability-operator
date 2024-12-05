package flow

import (
	"time"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/metrics"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
	"github.com/sirupsen/logrus"
)

const initialLogPeriod = time.Minute
const maxLogPeriod = time.Hour

var cllog = logrus.WithField("component", "capacity.Limiter")

// CapacityLimiter forwards the flows between two nodes but checks the status of the destination
// node's buffered channel. If it is already full, it drops the incoming flow and periodically will
// log a message about the number of lost flows.
type CapacityLimiter struct {
	droppedFlows int
	metrics      *metrics.Metrics
}

func NewCapacityLimiter(m *metrics.Metrics) *CapacityLimiter {
	return &CapacityLimiter{metrics: m}
}

func (c *CapacityLimiter) Limit(in <-chan []*model.Record, out chan<- []*model.Record) {
	go c.logDroppedFlows()
	for i := range in {
		if len(out) < cap(out) || cap(out) == 0 {
			out <- i
		} else {
			c.metrics.DroppedFlowsCounter.WithSourceAndReason("limiter", "full").Add(float64(len(i)))
			c.droppedFlows += len(i)
		}
	}
}

func (c *CapacityLimiter) logDroppedFlows() {
	logPeriod := initialLogPeriod
	debugging := logrus.IsLevelEnabled(logrus.DebugLevel)
	for {
		time.Sleep(logPeriod)

		// a race condition might happen in this counter but it's not important as it's just for
		// logging purposes
		df := c.droppedFlows
		if df > 0 {
			c.droppedFlows = 0
			cllog.Warnf("%d flows were dropped during the last %s because the agent is forwarding "+
				"more flows than the remote ingestor is able to process. You might "+
				"want to increase the CACHE_MAX_FLOWS and CACHE_ACTIVE_TIMEOUT property",
				df, logPeriod)

			// if not debug logs, backoff to avoid flooding the log with warning messages
			if !debugging && logPeriod < maxLogPeriod {
				logPeriod *= 2
				if logPeriod > maxLogPeriod {
					logPeriod = maxLogPeriod
				}
			}
		} else {
			logPeriod = initialLogPeriod
		}
	}
}
