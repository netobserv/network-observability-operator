package manager

import (
	"context"
	"fmt"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/pkg/cluster"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/migrator"
	"github.com/netobserv/network-observability-operator/pkg/narrowcache"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//+kubebuilder:rbac:groups=core,resources=namespaces;services;serviceaccounts;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods;nodes;endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings;rolebindings,verbs=get;list;create;delete;update;watch
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=get;create;delete;update;patch;list;watch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=consoles,verbs=get;list;update;watch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors;flowmetrics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/status;flowmetrics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=hostnetwork,verbs=use
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=list;create;update;watch
//+kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,verbs=list;get;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors;prometheusrules,verbs=get;create;delete;update;patch;list;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions;networks,verbs=get;list;watch
//+kubebuilder:rbac:groups=loki.grafana.com,resources=network,resourceNames=logs,verbs=create
//+kubebuilder:rbac:groups=metrics.k8s.io,resources=pods,verbs=create
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bpfman.io,resources=bpfapplications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions/status,verbs=update;patch
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=create;delete;patch;update;get;watch;list

type Registerer func(context.Context, *Manager) error

type Manager struct {
	manager.Manager
	ClusterInfo *cluster.Info
	Client      client.Client
	Status      *status.Manager
	Config      *Config
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
	info, err := cluster.NewInfo(dc)
	if err != nil {
		return nil, fmt.Errorf("can't collect cluster info: %w", err)
	}
	flowslatest.CurrentClusterInfo = &info

	this := &Manager{
		Manager:     internalManager,
		ClusterInfo: &info,
		Status:      status.NewManager(),
		Client:      client,
		Config:      opcfg,
	}

	log.Info("Building controllers")
	for _, f := range ctrls {
		if err := f(ctx, this); err != nil {
			return nil, fmt.Errorf("unable to create controller: %w", err)
		}
	}

	// On every startup, make sure stored CRs are up to date with the defined storage version.
	// This is simply going to run dummy patches to make the API server keep etcd consistent.
	mig := migrator.New(kcfg, []string{
		"flowcollectors.flows.netobserv.io",
		"flowmetrics.flows.netobserv.io",
	})
	if err := internalManager.Add(mig); err != nil {
		return nil, fmt.Errorf("unable to register migrator: %w", err)
	}

	if err := internalManager.Add(manager.RunnableFunc(func(ctx context.Context) error {
		return info.CheckClusterInfo(ctx, internalManager.GetClient())
	})); err != nil {
		return nil, fmt.Errorf("can't collect more cluster info: %w", err)
	}

	return this, nil
}

func (m *Manager) GetClient() client.Client {
	return m.Client
}
