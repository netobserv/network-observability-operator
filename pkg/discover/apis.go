package discover

import (
	"strings"

	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	operatorv1 "github.com/openshift/api/operator/v1"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	"k8s.io/client-go/discovery"
)

var (
	console    = "consoleplugins." + osv1alpha1.GroupName
	cno        = "networks." + operatorv1.GroupName
	svcMonitor = "servicemonitors." + monitoring.GroupName
)

// AvailableAPIs discovers the available APIs in the running cluster
type AvailableAPIs struct {
	apisMap map[string]bool
}

func NewAvailableAPIs(client *discovery.DiscoveryClient) (*AvailableAPIs, error) {
	apiMap := map[string]bool{
		console:    false,
		cno:        false,
		svcMonitor: false,
	}
	_, resources, err := client.ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}
	for apiName := range apiMap {
	out:
		for i := range resources {
			for j := range resources[i].APIResources {
				fullName := resources[i].APIResources[j].Name + "." + resources[i].GroupVersion
				if strings.HasPrefix(fullName, apiName) {
					apiMap[apiName] = true
					break out
				}
			}
		}
	}
	return &AvailableAPIs{apisMap: apiMap}, nil
}

// HasConsole returns true if "consoleplugins.console.openshift.io" API was found
func (c *AvailableAPIs) HasConsole() bool {
	return c.apisMap[console]
}

// HasCNO returns true if "networks.operator.openshift.io" API was found
func (c *AvailableAPIs) HasCNO() bool {
	return c.apisMap[cno]
}

// HasSvcMonitor returns true if "servicemonitors.monitoring.coreos.com" API was found
func (c *AvailableAPIs) HasSvcMonitor() bool {
	return c.apisMap[svcMonitor]
}
