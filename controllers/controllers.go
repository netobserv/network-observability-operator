package controllers

import (
	"github.com/netobserv/network-observability-operator/controllers/monitoring"
	"github.com/netobserv/network-observability-operator/pkg/manager"
)

var Registerers = []manager.Registerer{Start, monitoring.Start}
