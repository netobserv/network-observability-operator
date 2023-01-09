package reconcilers

import (
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
)

type Common struct {
	helper.ClientHelper
	CertWatcher *watchers.CertificatesWatcher
}
