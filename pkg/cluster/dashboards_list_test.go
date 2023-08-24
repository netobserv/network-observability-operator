package cluster

import (
	"context"
	"testing"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/stretchr/testify/assert"
)

func getDashboardsForVersion(version string) []string {
	info := Info{}
	info.SetOpenShiftVersion(version)
	d := NewDashboards()
	d.CheckClusterDashboards(context.Background(), &info)
	return d.GetList()
}

func TestDashboardsPerOCPVersion(t *testing.T) {
	// Check previous versions
	assert.Equal(t, []string{
		constants.KubernetesNetworkDashboard,
	}, getDashboardsForVersion("4.12.5"))

	// 4.15 introduces new dashboards; check exact version
	assert.Equal(t, []string{
		constants.KubernetesNetworkDashboard,
		constants.IngressDashboardCMName,
		constants.NetStatsDashboardCMName,
		constants.OVNDashboardCMName,
	}, getDashboardsForVersion("4.15.0"))

	// Check future versions
	assert.Equal(t, []string{
		constants.KubernetesNetworkDashboard,
		constants.IngressDashboardCMName,
		constants.NetStatsDashboardCMName,
		constants.OVNDashboardCMName,
	}, getDashboardsForVersion("4.15.5"))
}
