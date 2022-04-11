package discover

import (
	"context"

	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Vendor enumerates different Kubernetes distributions
type Vendor int

const (
	VendorUnknown = iota
	VendorVanilla
	VendorOpenShift
)

// Permissions discovers the actual security and permissions provider
type Permissions struct {
	vendor Vendor
	Client client.Client
}

// Vendor that provides the current permissions implementation
func (c *Permissions) Vendor(ctx context.Context) Vendor {
	if c.vendor != VendorUnknown {
		return c.vendor
	}
	services := apiregv1.APIServiceList{}
	rlog := log.FromContext(ctx)
	if err := c.Client.List(ctx, &services); err != nil {
		rlog.Error(err, "fetching vendor: couldn't retrieve API services")
		return VendorUnknown
	}
	for i := range services.Items {
		if services.Items[i].Name == "v1.security.openshift.io" {
			rlog.Info("fetching vendor: found OpenShift")
			c.vendor = VendorOpenShift
			return c.vendor
		}
	}
	rlog.Info("fetching vendor: any of our registered vendors matched. " +
		"Assuming vanilla Kubernetes")
	c.vendor = VendorVanilla
	return c.vendor
}
