package reconcilers

import (
	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
)

func URL(loki *flowsv1alpha1.FlowCollectorLoki) string {
	// force loki url to loki distributor from operator if requested
	if loki.InstanceSpec != nil && loki.InstanceSpec.Enable {
		return "https://loki-distributor-http:3100"
	}
	return loki.URL
}

func QuerierURL(loki *flowsv1alpha1.FlowCollectorLoki) string {
	// force loki url to loki query front end from operator if requested
	if loki.InstanceSpec != nil && loki.InstanceSpec.Enable {
		return "https://loki-query-frontend-http:3100"
	} else if loki.QuerierURL != "" {
		return loki.QuerierURL
	}
	return loki.URL
}
