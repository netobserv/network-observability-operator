package controllers

import (
	"github.com/netobserv/network-observability-operator/internal/controller/flp"
	"github.com/netobserv/network-observability-operator/internal/controller/monitoring"
	"github.com/netobserv/network-observability-operator/internal/controller/networkpolicy"
	"github.com/netobserv/network-observability-operator/internal/pkg/manager"
)

var Registerers = []manager.Registerer{Start, flp.Start, monitoring.Start, networkpolicy.Start}
