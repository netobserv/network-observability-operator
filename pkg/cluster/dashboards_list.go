package cluster

import (
	"context"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var orderedDashboards = []string{
	constants.KubernetesNetworkDashboard,
	constants.FlowDashboardCMName,
	constants.HealthDashboardCMName,
	constants.IngressDashboardCMName,
	constants.NetStatsDashboardCMName,
	constants.OVNDashboardCMName,
}

type Dashboards struct {
	availableCMNames map[string]interface{}
}

func NewDashboards() Dashboards {
	return Dashboards{
		availableCMNames: map[string]interface{}{
			constants.KubernetesNetworkDashboard: nil,
		},
	}
}

func (d *Dashboards) CheckClusterDashboards(ctx context.Context, clusterInfo *Info) {
	if ok, err := clusterInfo.OpenShiftVersionIsAtLeast("4.15.0"); err != nil {
		// Log error but do not fail: it's likely a bug in code, if the openshift version cannot be found
		log.FromContext(ctx).Error(err, "Could not get available dashboards for this cluster version. Is it OpenShift?")
	} else if ok {
		d.SetAvailable(constants.IngressDashboardCMName)
		d.SetAvailable(constants.NetStatsDashboardCMName)
		d.SetAvailable(constants.OVNDashboardCMName)
	}
}

func (d *Dashboards) SetAvailable(name string) {
	d.availableCMNames[name] = nil
}

func (d *Dashboards) GetList() []string {
	list := []string{}
	for _, name := range orderedDashboards {
		if _, ok := d.availableCMNames[name]; ok {
			list = append(list, name)
		}
	}
	return list
}
