package controllers

import (
	"github.com/netobserv/network-observability-operator/controllers/flp"
	"github.com/netobserv/network-observability-operator/controllers/monitoring"
	"github.com/netobserv/network-observability-operator/pkg/manager"
)

var Registerers = []manager.Registerer{Start, flp.Start, monitoring.Start}
