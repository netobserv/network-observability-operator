package manager

import (
	"context"
	"fmt"

	flowslatest "github.com/netobserv/netobserv-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/netobserv-operator/internal/controller/constants"
	"github.com/netobserv/netobserv-operator/internal/pkg/cluster"
	"github.com/netobserv/netobserv-operator/internal/pkg/manager/status"
	"github.com/netobserv/netobserv-operator/internal/pkg/migrator"
	"github.com/netobserv/netobserv-operator/internal/pkg/narrowcache"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Registerer func(context.Context, *Manager) (PostCreateHook, error)
type PostCreateHook = func(ctx context.Context) error

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

	narrowCache := narrowcache.NewConfig(kcfg,
		narrowcache.ConfigMaps,
		narrowcache.ClusterRoles,
		narrowcache.ClusterRoleBindings,
		narrowcache.Daemonsets,
		narrowcache.Deployments,
		narrowcache.HorizontalPodAutoscalers,
		narrowcache.Namespaces,
		narrowcache.NetworkPolicies,
		narrowcache.Roles,
		narrowcache.RoleBindings,
		narrowcache.Secrets,
		narrowcache.Services,
		narrowcache.ServiceAccounts,
		narrowcache.Endpoints,
		narrowcache.EndpointSlices,
	)
	opts.Client = client.Options{Cache: narrowCache.ControllerRuntimeClientCacheOptions()}
	opts.Cache = cache.Options{
		ByObject: map[client.Object]cache.ByObject{
			&corev1.Pod{}: {
				Label: labels.SelectorFromSet(map[string]string{"part-of": constants.OperatorName}),
			},
		},
	}

	internalManager, err := ctrl.NewManager(kcfg, *opts)
	if err != nil {
		return nil, err
	}
	client, err := narrowCache.CreateClient(internalManager.GetClient())
	if err != nil {
		return nil, fmt.Errorf("unable to create narrow cache client: %w", err)
	}

	statusMgr := status.NewManager()
	statusMgr.SetEventRecorder(internalManager.GetEventRecorderFor("flowcollector-controller")) //nolint:staticcheck

	log.Info("Discovering APIs")
	dc, err := discovery.NewDiscoveryClientForConfig(kcfg)
	if err != nil {
		return nil, fmt.Errorf("can't instantiate discovery client: %w", err)
	}
	info, postCreate, err := cluster.NewInfo(ctx, kcfg, dc, func() { statusMgr.Sync(ctx, client) })
	if err != nil {
		return nil, fmt.Errorf("can't collect cluster info: %w", err)
	}
	flowslatest.CurrentClusterInfo = info

	this := &Manager{
		Manager:     internalManager,
		ClusterInfo: info,
		Status:      statusMgr,
		Client:      client,
		Config:      opcfg,
	}

	log.Info("Building controllers")
	for _, f := range ctrls {
		if hook, err := f(ctx, this); err != nil {
			return nil, fmt.Errorf("unable to create controller: %w", err)
		} else if hook != nil {
			if err := internalManager.Add(manager.RunnableFunc(hook)); err != nil {
				return nil, fmt.Errorf("unable to register controller post-create hook: %w", err)
			}
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
		return postCreate(ctx)
	})); err != nil {
		return nil, fmt.Errorf("can't collect more cluster info: %w", err)
	}

	return this, nil
}

func (m *Manager) GetClient() client.Client {
	return m.Client
}
