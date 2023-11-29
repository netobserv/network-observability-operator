package manager

import (
	"context"
	"fmt"

	"github.com/netobserv/network-observability-operator/pkg/discover"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/narrowcache"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces;services;serviceaccounts;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings;clusterroles;rolebindings;roles,verbs=get;list;create;delete;update;watch
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=get;create;delete;update;patch;list;watch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=consoles,verbs=get;update;list;update;watch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=hostnetwork,verbs=use
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=list;create;update;watch
//+kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,verbs=list;get;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors;prometheusrules,verbs=get;create;delete;update;patch;list;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch
//+kubebuilder:rbac:groups=loki.grafana.com,resources=network,resourceNames=logs,verbs=get;create
//+kubebuilder:rbac:urls="/metrics",verbs=get

type Registerer func(context.Context, *Manager) error

type Manager struct {
	manager.Manager
	discover.AvailableAPIs
	Client client.Client
	Status *status.Manager
	Config *Config
	vendor discover.Vendor
}

func NewManager(
	ctx context.Context,
	kcfg *rest.Config,
	opcfg *Config,
	opts *ctrl.Options,
	ctrls []Registerer,
) (*Manager, error) {

	log := log.FromContext(ctx)
	log.Info("Creating manager")

	narrowCache := narrowcache.NewConfig(kcfg, narrowcache.ConfigMaps, narrowcache.Secrets)
	opts.Client = client.Options{Cache: narrowCache.ControllerRuntimeClientCacheOptions()}

	internalManager, err := ctrl.NewManager(kcfg, *opts)
	if err != nil {
		return nil, err
	}
	client, err := narrowCache.CreateClient(internalManager.GetClient())
	if err != nil {
		return nil, fmt.Errorf("unable to create narrow cache client: %w", err)
	}

	log.Info("Discovering APIs")
	dc, err := discovery.NewDiscoveryClientForConfig(kcfg)
	if err != nil {
		return nil, fmt.Errorf("can't instantiate discovery client: %w", err)
	}
	permissions := discover.Permissions{Client: dc}
	vendor := permissions.Vendor(ctx)
	apis, err := discover.NewAvailableAPIs(dc)
	if err != nil {
		return nil, fmt.Errorf("can't discover available APIs: %w", err)
	}

	this := &Manager{
		Manager:       internalManager,
		AvailableAPIs: *apis,
		Status:        status.NewManager(),
		Client:        client,
		Config:        opcfg,
		vendor:        vendor,
	}

	log.Info("Building controllers")
	for _, f := range ctrls {
		if err := f(ctx, this); err != nil {
			return nil, fmt.Errorf("unable to create controller: %w", err)
		}
	}

	return this, nil
}

func (m *Manager) GetClient() client.Client {
	return m.Client
}

func (m *Manager) IsOpenShift() bool {
	return m.vendor == discover.VendorOpenShift
}
