package cluster

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	configv1 "github.com/openshift/api/config/v1"
	osv1 "github.com/openshift/api/console/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	securityv1 "github.com/openshift/api/security/v1"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

func NewInfo(ctx context.Context, dcl *discovery.DiscoveryClient) (Info, error) {
	info := Info{}
	if err := info.fetchAvailableAPIs(ctx, dcl); err != nil {
		return info, err
	}
	return info, nil
}

func (c *Info) CheckClusterInfo(ctx context.Context, cl client.Client) error {
	if c.IsOpenShift() && !c.fetchedClusterVersion {
		if err := c.fetchOpenShiftClusterVersion(ctx, cl); err != nil {
			return err
		}
	}
	return nil
}

func (c *Info) fetchAvailableAPIs(ctx context.Context, client *discovery.DiscoveryClient) error {
	log := log.FromContext(ctx)
	c.apisMap = map[string]bool{
		consolePlugin: false,
		cno:           false,
		svcMonitor:    false,
		promRule:      false,
		ocpSecurity:   false,
	}
	_, resources, err := client.ServerGroupsAndResources()
	// We may receive partial data along with an error
	var discErr *discovery.ErrGroupDiscoveryFailed
	if err != nil && (!errors.As(err, &discErr) || len(resources) == 0) {
		return err
	}
	for apiName := range c.apisMap {
		if hasAPI(apiName, resources) {
			c.apisMap[apiName] = true
		} else {
			// Check if the wanted API is in error
			for gv, err := range discErr.Groups {
				if strings.Contains(apiName, gv.String()) {
					log.Error(err, "some API-related features are unavailable; you can check for stale APIs with 'kubectl get apiservice'", "GroupVersion", gv.String())
				}
			}
		}
	}

	return nil
}

func hasAPI(apiName string, resources []*metav1.APIResourceList) bool {
	for i := range resources {
		for j := range resources[i].APIResources {
			gvk := resources[i].APIResources[j].Name + "." + resources[i].GroupVersion
			if apiName == gvk {
				return true
			}
		}
	}
	return false
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
	openshiftVersion := *c.openShiftVersion
	// Ignore pre-release block for comparison
	openshiftVersion.PreRelease = ""
	return !openshiftVersion.LessThan(*version), nil
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
