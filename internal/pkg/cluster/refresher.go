package cluster

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// startRefreshLoop starts a 10-minutes ticker to refresh cluster info;
// e.g. used to get updated warnings related to node count in case of cluster upscaling
func (c *Info) startRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	log := log.FromContext(ctx)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Info("Refreshing ClusterInfo")
				if err := c.refresh(ctx); err != nil {
					log.Error(err, "error while refreshing ClusterInfo")
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *Info) refresh(ctx context.Context) error {
	if err := c.fetchAvailableAPIs(ctx); err != nil {
		return err
	}
	if err := c.fetchClusterInfo(ctx); err != nil {
		return err
	}
	c.onRefresh()
	return nil
}
