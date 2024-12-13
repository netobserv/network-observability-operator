package pipeline

import (
	"context"
	"fmt"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/ingest"
	"github.com/netobserv/flowlogs-pipeline/pkg/prometheus"
)

// StartFLPInProcess is an entry point to start the whole FLP / pipeline processing from imported code
func StartFLPInProcess(cfg *config.ConfigFileStruct, in chan config.GenericMap) error {
	promServer := prometheus.InitializePrometheus(&cfg.MetricsSettings)

	// Create new flows pipeline
	ingester := ingest.NewInProcess(in)
	flp, err := newPipelineFromIngester(cfg, ingester)
	if err != nil {
		return fmt.Errorf("failed to initialize pipeline %w", err)
	}

	// Starts the flows pipeline; blocking call
	go func() {
		flp.Run()
		if promServer != nil {
			_ = promServer.Shutdown(context.Background())
		}
	}()

	return nil
}
