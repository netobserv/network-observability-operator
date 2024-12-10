package utils

import (
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/sirupsen/logrus"
)

func Batcher(
	closeCh <-chan struct{},
	maxBatchLength int,
	batchTimeout time.Duration,
	inCh <-chan config.GenericMap,
	action func([]config.GenericMap),
) {
	log := logrus.WithField("component", "utils.Batcher")
	invokeTicker := time.NewTicker(batchTimeout)
	var entries []config.GenericMap
	log.Debug("starting")
	for {
		select {
		case <-closeCh:
			log.Debug("exiting due to closeCh")
			return
		case <-invokeTicker.C:
			if len(entries) == 0 {
				continue
			}
			log.Debugf("ticker signal: invoking action with %d entries", len(entries))
			es := entries
			entries = nil
			action(es)
		case gm := <-inCh:
			entries = append(entries, gm)
			if len(entries) >= maxBatchLength {
				log.Debugf("batch complete: invoking action with %d entries", len(entries))
				es := entries
				entries = nil
				action(es)
			}
		}
	}
}
