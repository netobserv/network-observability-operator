package cluster

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/coreos/go-semver/semver"
	configv1 "github.com/openshift/api/config/v1"
	osv1 "github.com/openshift/api/console/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	securityv1 "github.com/openshift/api/security/v1"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type NetworkType string

const (
	OpenShiftSDN  NetworkType = "OpenShiftSDN"
	OVNKubernetes NetworkType = "OVNKubernetes"
)

type Info struct {
	apisMap          map[string]bool
	apisMapLock      sync.RWMutex
	id               string
	openShiftVersion *semver.Version
	cni              NetworkType
	nbNodes          uint16
	ready            bool
	readinessLock    sync.RWMutex
	dcl              *discovery.DiscoveryClient
	cl               client.Client
	onRefresh        func()
}

var (
	consolePlugin  = "consoleplugins." + osv1.GroupVersion.String()
	cno            = "networks." + operatorv1.GroupVersion.String()
	svcMonitor     = "servicemonitors." + monv1.SchemeGroupVersion.String()
	promRule       = "prometheusrules." + monv1.SchemeGroupVersion.String()
	ocpSecurity    = "securitycontextconstraints." + securityv1.SchemeGroupVersion.String()
	endpointSlices = "endpointslices." + discoveryv1.SchemeGroupVersion.String()
)

func NewInfo(ctx context.Context, cl client.Client, dcl *discovery.DiscoveryClient, onRefresh func()) (*Info, func(ctx context.Context) error, error) {
	info := Info{cl: cl, dcl: dcl, onRefresh: onRefresh}
	if err := info.fetchAvailableAPIs(ctx); err != nil {
		return &info, nil, err
	}
	return &info, info.postCreate, nil
}

func (c *Info) fetchAvailableAPIs(ctx context.Context) error {
	log := log.FromContext(ctx)
	_, resources, err := c.dcl.ServerGroupsAndResources()
	// We may receive partial data along with an error
	var discErr *discovery.ErrGroupDiscoveryFailed
	if err != nil && (!errors.As(err, &discErr) || len(resources) == 0) {
		return err
	}
	apisMap := map[string]bool{
		consolePlugin:  false,
		cno:            false,
		svcMonitor:     false,
		promRule:       false,
		ocpSecurity:    false,
		endpointSlices: false,
	}
	for apiName := range apisMap {
		if hasAPI(apiName, resources) {
			apisMap[apiName] = true
		} else if discErr != nil {
			// Check if the wanted API is in error
			for gv, err := range discErr.Groups {
				if strings.Contains(apiName, gv.String()) {
					log.Error(err, "some API-related features are unavailable; you can check for stale APIs with 'kubectl get apiservice'", "GroupVersion", gv.String())
				}
			}
		}
	}

	c.apisMapLock.Lock()
	defer c.apisMapLock.Unlock()
	c.apisMap = apisMap

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

func (c *Info) postCreate(ctx context.Context) error {
	if err := c.fetchClusterInfo(ctx); err != nil {
		return err
	}
	c.startRefreshLoop(ctx)
	return nil
}

func (c *Info) fetchClusterInfo(ctx context.Context) error {
	var id string
	var openShiftVersion *semver.Version
	var cni NetworkType
	var nbNodes uint16
	if c.IsOpenShift() {
		// Fetch cluster ID, version and CNI
		key := client.ObjectKey{Name: "version"}
		cversion := &configv1.ClusterVersion{}
		if err := c.cl.Get(ctx, key, cversion); err != nil {
			return fmt.Errorf("could not fetch ClusterVersion: %w", err)
		}
		id = string(cversion.Spec.ClusterID)
		// Get version; use the same method as via `oc get clusterversion`, where printed column uses jsonPath:
		// .status.history[?(@.state=="Completed")].version
		for _, history := range cversion.Status.History {
			if history.State == "Completed" {
				openShiftVersion = semver.New(history.Version)
				break
			}
		}
		network := &configv1.Network{}
		err := c.cl.Get(ctx, client.ObjectKey{Name: "cluster"}, network)
		if err != nil {
			return fmt.Errorf("could not fetch Network resource: %w", err)
		}
		cni = NetworkType(network.Spec.NetworkType)
	}

	l := v1.NodeList{}
	if err := c.cl.List(ctx, &l); err != nil {
		return fmt.Errorf("could not retrieve number of nodes: %w", err)
	}
	nbNodes = uint16(len(l.Items))
	c.setInfo(id, openShiftVersion, cni, nbNodes)

	return nil
}

func (c *Info) setInfo(id string, openShiftVersion *semver.Version, cni NetworkType, nbNodes uint16) {
	c.readinessLock.Lock()
	defer c.readinessLock.Unlock()
	c.id = id
	c.openShiftVersion = openShiftVersion
	c.cni = cni
	c.nbNodes = nbNodes
	c.ready = true
}

// Mock shouldn't be used except for testing
func (c *Info) Mock(v string, cni NetworkType) {
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
	c.cni = cni
	c.ready = true
}

func (c *Info) GetID() string {
	c.readinessLock.RLock()
	defer c.readinessLock.RUnlock()
	return c.id
}

func (c *Info) GetOpenShiftVersion() (string, error) {
	c.readinessLock.RLock()
	defer c.readinessLock.RUnlock()
	if !c.ready {
		return "", errors.New("cluster info not collected")
	}
	if c.openShiftVersion == nil {
		return "", errors.New("unknown OpenShift version")
	}
	return c.openShiftVersion.String(), nil
}

func (c *Info) GetCNI() (NetworkType, error) {
	c.readinessLock.RLock()
	defer c.readinessLock.RUnlock()
	if !c.ready {
		return "", errors.New("cluster info not collected")
	}
	return c.cni, nil
}

func (c *Info) GetNbNodes() (uint16, error) {
	c.readinessLock.RLock()
	defer c.readinessLock.RUnlock()
	if !c.ready {
		return 0, errors.New("cluster info not collected")
	}
	return c.nbNodes, nil
}

func (c *Info) IsOpenShiftVersionLessThan(v string) (bool, string, error) {
	c.readinessLock.RLock()
	defer c.readinessLock.RUnlock()
	if !c.ready {
		return false, "", errors.New("cluster info not collected")
	}
	if c.openShiftVersion == nil {
		return false, "", errors.New("unknown OpenShift version, cannot compare versions")
	}
	version, err := semver.NewVersion(v)
	if err != nil {
		return false, "", err
	}
	openshiftVersion := *c.openShiftVersion
	// Ignore pre-release block for comparison
	openshiftVersion.PreRelease = ""
	return openshiftVersion.LessThan(*version), c.openShiftVersion.String(), nil
}

func (c *Info) IsOpenShiftVersionAtLeast(v string) (bool, string, error) {
	b, v, err := c.IsOpenShiftVersionLessThan(v)
	return !b, v, err
}

// IsOpenShift assumes having openshift SCC API <=> being on openshift
func (c *Info) IsOpenShift() bool {
	return c.HasOCPSecurity()
}

// HasConsolePlugin returns true if "consoleplugins.console.openshift.io" API was found
func (c *Info) HasConsolePlugin() bool {
	c.apisMapLock.RLock()
	defer c.apisMapLock.RUnlock()
	return c.apisMap[consolePlugin]
}

// HasOCPSecurity returns true if "consoles.config.openshift.io" API was found
func (c *Info) HasOCPSecurity() bool {
	c.apisMapLock.RLock()
	defer c.apisMapLock.RUnlock()
	return c.apisMap[ocpSecurity]
}

// HasCNO returns true if "networks.operator.openshift.io" API was found
func (c *Info) HasCNO() bool {
	c.apisMapLock.RLock()
	defer c.apisMapLock.RUnlock()
	return c.apisMap[cno]
}

// HasSvcMonitor returns true if "servicemonitors.monitoring.coreos.com" API was found
func (c *Info) HasSvcMonitor() bool {
	c.apisMapLock.RLock()
	defer c.apisMapLock.RUnlock()
	return c.apisMap[svcMonitor]
}

// HasPromRule returns true if "prometheusrules.monitoring.coreos.com" API was found
func (c *Info) HasPromRule() bool {
	c.apisMapLock.RLock()
	defer c.apisMapLock.RUnlock()
	return c.apisMap[promRule]
}

// HasEndpointSlices returns true if "endpointslices.discovery.k8s.io" API was found
func (c *Info) HasEndpointSlices() bool {
	c.apisMapLock.RLock()
	defer c.apisMapLock.RUnlock()
	return c.apisMap[endpointSlices]
}
