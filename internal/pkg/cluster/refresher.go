package cluster

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// Normal refresh interval when everything is working
	normalRefreshInterval = 10 * time.Minute
	// Initial retry interval after an error
	initialRetryInterval = 30 * time.Second
	// Maximum retry interval (caps exponential backoff)
	maxRetryInterval = 10 * time.Minute
)

// startRefreshLoop starts a periodic refresh of cluster info with exponential backoff on errors
// Normal operation: refreshes every 10 minutes
// On error: retries with exponential backoff (30s, 1m, 2m, 4m, 8m, 10m max)
func (c *Info) startRefreshLoop(ctx context.Context) {
	log := log.FromContext(ctx)

	// Start with normal interval
	currentInterval := normalRefreshInterval
	timer := time.NewTimer(currentInterval)

	go func() {
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				log.Info("Refreshing ClusterInfo", "interval", currentInterval)
				err := c.refresh(ctx)

				if err != nil {
					log.Error(err, "error while refreshing ClusterInfo")
					// Exponential backoff on error
					if currentInterval == normalRefreshInterval {
						// First error: start with initial retry interval
						currentInterval = initialRetryInterval
					} else {
						// Subsequent errors: double the interval (exponential backoff)
						currentInterval *= 2
						if currentInterval > maxRetryInterval {
							currentInterval = maxRetryInterval
						}
					}
					log.Info("Scheduling retry with backoff", "nextRetry", currentInterval)
				} else if currentInterval != normalRefreshInterval {
					// Success: reset to normal interval
					log.Info("ClusterInfo refresh successful, resetting to normal interval")
					currentInterval = normalRefreshInterval
				}

				// Schedule next refresh
				timer.Reset(currentInterval)

			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *Info) refresh(ctx context.Context) error {
	// During refresh, allow critical API failures to give the API server time to recover
	// The operator will continue with existing cluster info and retry on next refresh cycle
	if err := c.fetchAvailableAPIsInternal(ctx, true); err != nil {
		return err
	}
	if err := c.fetchClusterInfo(ctx); err != nil {
		return err
	}
	c.onRefresh()
	return nil
}
