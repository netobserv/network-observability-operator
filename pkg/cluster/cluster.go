package cluster

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-semver/semver"
	configv1 "github.com/openshift/api/config/v1"
	osv1 "github.com/openshift/api/console/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	securityv1 "github.com/openshift/api/security/v1"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
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
	consolePlugin = "consoleplugins." + osv1.GroupVersion.String()
	cno           = "networks." + operatorv1.GroupVersion.String()
	svcMonitor    = "servicemonitors." + monv1.SchemeGroupVersion.String()
	promRule      = "prometheusrules." + monv1.SchemeGroupVersion.String()
	ocpSecurity   = "securitycontextconstraints." + securityv1.SchemeGroupVersion.String()
)

func NewInfo(dcl *discovery.DiscoveryClient) (Info, error) {
	info := Info{}
	if err := info.fetchAvailableAPIs(dcl); err != nil {
		return info, err
	}
	return info, nil
}

func (c *Info) CheckClusterInfo(ctx context.Context, cl client.Client) error {
	var errs []error
	if c.IsOpenShift() && !c.fetchedClusterVersion {
		if err := c.fetchOpenShiftClusterVersion(ctx, cl); err != nil {
			errs = append(errs, err)
		}
		if err := c.fetchOpenShiftClusterID(ctx, cl); err != nil {
			errs = append(errs, err)
		}
		if len(errs) != 0 {
			return kerrors.NewAggregate(errs)
		}
		c.fetchedClusterVersion = true
	}
	return nil
}

func (c *Info) fetchAvailableAPIs(client *discovery.DiscoveryClient) error {
	c.apisMap = map[string]bool{
		consolePlugin: false,
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
				gvk := resources[i].APIResources[j].Name + "." + resources[i].GroupVersion
				if apiName == gvk {
					c.apisMap[apiName] = true
					break out
				}
			}
		}
	}
	return nil
}

func (c *Info) fetchOpenShiftClusterVersion(ctx context.Context, cl client.Client) error {
	cno := &configv1.ClusterOperator{}
	err := cl.Get(ctx, types.NamespacedName{Name: "network"}, cno)
	if err != nil {
		return fmt.Errorf("error fetching OpenShift Cluster Network Operator: %w", err)
	}
	for _, v := range cno.Status.Versions {
		if v.Name == "operator" {
			ver, err := semver.NewVersion(v.Version)
			if err != nil {
				return fmt.Errorf("error parsing OpenShift Cluster Network Operator version: %w", err)
			}
			c.openShiftVersion = ver
			break
		}
	}
	return nil
}

func (c *Info) fetchOpenShiftClusterID(ctx context.Context, cl client.Client) error {
	key := client.ObjectKey{Name: "version"}
	version := &configv1.ClusterVersion{}
	if err := cl.Get(ctx, key, version); err != nil {
		return fmt.Errorf("could not fetch ClusterVersion: %w", err)
	}
	c.ID = string(version.Spec.ClusterID)
	return nil
}

// MockOpenShiftVersion shouldn't be used except for testing
func (c *Info) MockOpenShiftVersion(v string) {
	if c.apisMap == nil {
		c.apisMap = make(map[string]bool)
	}
	if v == "" {
		// No OpenShift
		c.apisMap[ocpSecurity] = false
		c.openShiftVersion = nil
	} else {
		c.apisMap[ocpSecurity] = true
		c.openShiftVersion = semver.New(v)
	}
}

func (c *Info) GetOpenShiftVersion() string {
	return c.openShiftVersion.String()
}

func (c *Info) OpenShiftVersionIsAtLeast(v string) (bool, error) {
	if c.openShiftVersion == nil {
		return false, errors.New("OpenShift version not defined, can't compare versions")
	}
	version, err := semver.NewVersion(v)
	if err != nil {
		return false, err
	}
	return !c.openShiftVersion.LessThan(*version), nil
}

// IsOpenShift assumes having openshift SCC API <=> being on openshift
func (c *Info) IsOpenShift() bool {
	return c.HasOCPSecurity()
}

// HasConsolePlugin returns true if "consoleplugins.console.openshift.io" API was found
func (c *Info) HasConsolePlugin() bool {
	return c.apisMap[consolePlugin]
}

// HasOCPSecurity returns true if "consoles.config.openshift.io" API was found
func (c *Info) HasOCPSecurity() bool {
	return c.apisMap[ocpSecurity]
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
