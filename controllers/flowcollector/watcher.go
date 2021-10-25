// Package flow provides the functionality to watch for FlowCollector configuration changes and
// accordingly update the ovs-flows-config configmap
package flowcollector

import (
	"fmt"

	"github.com/netobserv/network-observability-operator/api/v1alpha1"

	"github.com/sirupsen/logrus"
)

var wlog = logrus.WithField("component", "flowcollector.Watcher")

// EventType related to a FlowCollector change
type EventType int

const (
	Created EventType = iota
	Deleted
	Modified
)

func (e EventType) String() string {
	switch e {
	case Created:
		return "Created"
	case Deleted:
		return "Deleted"
	case Modified:
		return "Modified"
	}
	return fmt.Sprintf("Invalid (%d)", int(e))
}

type Event struct {
	Type   EventType
	Object *v1alpha1.FlowCollector
}

type Watcher struct {
	Events     <-chan Event
	Configurer Configurer
}

// Start the watching process as a background goroutine
func (w *Watcher) Start() {
	go func() {
		for event := range w.Events {
			w.handle(event)
		}
	}()
}

func (w *Watcher) handle(event Event) {
	switch event.Type {
	case Created, Modified:
		if event.Object == nil {
			wlog.WithField("eventType", event.Type).
				Warn("received incomplete flow collector. Ignoring")
		}
		if err := w.Configurer.Set(event.Object.Spec.IPFIX); err != nil {
			wlog.WithFields(logrus.Fields{
				"eventType":     event.Type,
				"ipfixConfig":   event.Object.Spec.IPFIX,
				logrus.ErrorKey: err,
			}).Error("can't update configuration for Cluster Network Operator")
		}
	case Deleted:
		if err := w.Configurer.Delete(); err != nil {
			wlog.WithFields(logrus.Fields{
				"eventType":     event.Type,
				"ipfixConfig":   event.Object.Spec.IPFIX,
				logrus.ErrorKey: err,
			}).Error("can't delete configuration for Cluster Network Operator")
		}
	}
}
