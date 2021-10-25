package flowcollector

import (
	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
)

type Configurer interface {
	Set(ipfix flowsv1alpha1.FlowCollectorIPFIX) error
	Delete() error
}

