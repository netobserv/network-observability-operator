package cluster

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	configv1 "github.com/openshift/api/config/v1"
	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	operatorv1 "github.com/openshift/api/operator/v1"
	securityv1 "github.com/openshift/api/security/v1"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Info struct {
	ID                    string
	openShiftVersion      *semver.Version
	apisMap               map[string]bool
	fetchedClusterVersion bool
}

var (
	consolePlugin = "consoleplugins." + osv1alpha1.GroupName
	consoleConfig = "consoles." + configv1.GroupName
	cno           = "networks." + operatorv1.GroupName
	svcMonitor    = "servicemonitors." + monitoring.GroupName
	promRule      = "prometheusrules." + monitoring.GroupName
	ocpSecurity   = "securitycontextconstraints." + securityv1.GroupName
)

func NewInfo(dcl *discovery.DiscoveryClient) (Info, error) {
	info := Info{}
	if err := info.fetchAvailableAPIs(dcl); err != nil {
		return info, err
	}
	return info, nil
}

func (c *Info) CheckClusterInfo(ctx context.Context, cl client.Client) error {
	if c.HasOCPSecurity() && !c.fetchedClusterVersion {
		// Assumes having openshift security <=> being on openshift
		return c.fetchOpenShiftClusterVersion(ctx, cl)
	}
	return nil
}

func (c *Info) fetchAvailableAPIs(client *discovery.DiscoveryClient) error {
	c.apisMap = map[string]bool{
		consolePlugin: false,
		consoleConfig: false,
		cno:           false,
		svcMonitor:    false,
		promRule:      false,
		ocpSecurity:   false,
	}
	_, resources, err := client.ServerGroupsAndResources()
	if err != nil {
		return err
	}
	for apiName := range c.apisMap {
	out:
		for i := range resources {
			for j := range resources[i].APIResources {
				fullName := resources[i].APIResources[j].Name + "." + resources[i].GroupVersion
				if strings.HasPrefix(fullName, apiName) {
					c.apisMap[apiName] = true
					break out
				}
			}
		}
	}
	return nil
}

func (c *Info) fetchOpenShiftClusterVersion(ctx context.Context, cl client.Client) error {
	key := client.ObjectKey{Name: "version"}
	cversion := &configv1.ClusterVersion{}
	if err := cl.Get(ctx, key, cversion); err != nil {
		return fmt.Errorf("could not fetch ClusterVersion: %w", err)
	}
	c.ID = string(cversion.Spec.ClusterID)
	// Get version; use the same method as via `oc get clusterversion`, where printed column uses jsonPath:
	// .status.history[?(@.state=="Completed")].version
	for _, history := range cversion.Status.History {
		if history.State == "Completed" {
			c.openShiftVersion = semver.New(history.Version)
			break
		}
	}
	c.fetchedClusterVersion = true
	return nil
}

// SetOpenShiftVersion shouldn't be used except for testing
func (c *Info) SetOpenShiftVersion(v string) {
	c.openShiftVersion = semver.New(v)
}

func (c *Info) OpenShiftVersionIsAtLeast(v string) (bool, error) {
	if c.openShiftVersion == nil {
		return false, errors.New("OpenShift version not defined, can't compare versions")
	}
	version := semver.New(v)
	return !c.openShiftVersion.LessThan(*version), nil
}

// HasOCPSecurity returns true if "securitycontextconstraints.security.openshift.io" API was found
func (c *Info) HasOCPSecurity() bool {
	return c.apisMap[ocpSecurity]
}

// HasConsolePlugin returns true if "consoleplugins.console.openshift.io" API was found
func (c *Info) HasConsolePlugin() bool {
	return c.apisMap[consolePlugin]
}

// HasConsoleConfig returns true if "consoles.config.openshift.io" API was found
func (c *Info) HasConsoleConfig() bool {
	return c.apisMap[consoleConfig]
}

// HasCNO returns true if "networks.operator.openshift.io" API was found
func (c *Info) HasCNO() bool {
	return c.apisMap[cno]
}

// HasSvcMonitor returns true if "servicemonitors.monitoring.coreos.com" API was found
func (c *Info) HasSvcMonitor() bool {
	return c.apisMap[svcMonitor]
}

// HasPromRule returns true if "prometheusrules.monitoring.coreos.com" API was found
func (c *Info) HasPromRule() bool {
	return c.apisMap[promRule]
}
