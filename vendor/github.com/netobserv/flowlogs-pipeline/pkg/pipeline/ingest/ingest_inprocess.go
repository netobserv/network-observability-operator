package ingest

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
)

// InProcess ingester is meant to be imported and used from another program
// via pipeline.StartFLPInProcess
type InProcess struct {
	in chan config.GenericMap
}

func NewInProcess(in chan config.GenericMap) *InProcess {
	return &InProcess{in: in}
}

func (d *InProcess) Ingest(out chan<- config.GenericMap) {
	go func() {
		<-utils.ExitChannel()
		d.Close()
	}()
	for rec := range d.in {
		out <- rec
	}
}

func (d *InProcess) Close() {
}
