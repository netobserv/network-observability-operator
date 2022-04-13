package discover

import (
	"context"

	securityv1 "github.com/openshift/api/security/v1"
	"k8s.io/client-go/discovery"
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
	Client *discovery.DiscoveryClient
}

// Vendor that provides the current permissions implementation
func (c *Permissions) Vendor(ctx context.Context) Vendor {
	if c.vendor != VendorUnknown {
		return c.vendor
	}
	rlog := log.FromContext(ctx)
	groupsList, err := c.Client.ServerGroups()
	if err != nil {
		rlog.Error(err, "fetching vendor: couldn't retrieve API services")
		return VendorUnknown
	}
	for i := range groupsList.Groups {
		if groupsList.Groups[i].Name == securityv1.GroupName {
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
