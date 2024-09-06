package discover

import (
	osv1 "github.com/openshift/api/console/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/client-go/discovery"
)

var (
	consolePlugin = "consoleplugins." + osv1.GroupVersion.String()
	cno           = "networks." + operatorv1.GroupVersion.String()
	svcMonitor    = "servicemonitors." + monv1.SchemeGroupVersion.String()
	promRule      = "prometheusrules." + monv1.SchemeGroupVersion.String()
)

// AvailableAPIs discovers the available APIs in the running cluster
type AvailableAPIs struct {
	apisMap map[string]bool
}

func NewAvailableAPIs(client *discovery.DiscoveryClient) (*AvailableAPIs, error) {
	apiMap := map[string]bool{
		consolePlugin: false,
		cno:           false,
		svcMonitor:    false,
		promRule:      false,
	}
	_, resources, err := client.ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}
	for apiName := range apiMap {
	out:
		for i := range resources {
			for j := range resources[i].APIResources {
				gvk := resources[i].APIResources[j].Name + "." + resources[i].GroupVersion
				if apiName == gvk {
					apiMap[apiName] = true
					break out
				}
			}
		}
	}
	return &AvailableAPIs{apisMap: apiMap}, nil
}

// HasConsolePlugin returns true if "consoleplugins.console.openshift.io" API was found
func (c *AvailableAPIs) HasConsolePlugin() bool {
	return c.apisMap[consolePlugin]
}

// HasCNO returns true if "networks.operator.openshift.io" API was found
func (c *AvailableAPIs) HasCNO() bool {
	return c.apisMap[cno]
}

// HasSvcMonitor returns true if "servicemonitors.monitoring.coreos.com" API was found
func (c *AvailableAPIs) HasSvcMonitor() bool {
	return c.apisMap[svcMonitor]
}

// HasPromRule returns true if "prometheusrules.monitoring.coreos.com" API was found
func (c *AvailableAPIs) HasPromRule() bool {
	return c.apisMap[promRule]
}
