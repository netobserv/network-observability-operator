package reconcilers

import (
	"context"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/cluster"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type Common struct {
	helper.Client
	Watcher           *watchers.Watcher
	Namespace         string
	PreviousNamespace string
	ClusterInfo       *cluster.Info
	Loki              *helper.LokiConfig
	IsDownstream      bool
}

func (c *Common) PrivilegedNamespace() string {
	return c.Namespace + constants.EBPFPrivilegedNSSuffix
}

func (c *Common) PreviousPrivilegedNamespace() string {
	return c.PreviousNamespace + constants.EBPFPrivilegedNSSuffix
}

type Instance struct {
	*Common
	Managed *NamespacedObjectManager
	Images  []string
	Status  status.Instance
}

func (c *Common) NewInstance(images []string, st status.Instance) *Instance {
	managed := NewNamespacedObjectManager(c)
	return &Instance{
		Common:  c,
		Managed: managed,
		Images:  images,
		Status:  st,
	}
}

func (c *Common) ReconcileClusterRoleBinding(ctx context.Context, desired *rbacv1.ClusterRoleBinding) error {
	return ReconcileClusterRoleBinding(ctx, &c.Client, desired)
}

func (c *Common) ReconcileRoleBinding(ctx context.Context, desired *rbacv1.RoleBinding) error {
	return ReconcileRoleBinding(ctx, &c.Client, desired)
}

func (c *Common) ReconcileClusterRoleBindings(ctx context.Context, desired []*rbacv1.ClusterRoleBinding) error {
	for _, d := range desired {
		if err := c.ReconcileClusterRoleBinding(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

func (c *Common) ReconcileRoleBindings(ctx context.Context, desired []*rbacv1.RoleBinding) error {
	for _, d := range desired {
		if err := c.ReconcileRoleBinding(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

func (c *Common) ReconcileConfigMap(ctx context.Context, old, new *corev1.ConfigMap) error {
	return ReconcileConfigMap(ctx, &c.Client, old, new)
}

func (i *Instance) ReconcileService(ctx context.Context, old, new *corev1.Service, report *helper.ChangeReport) error {
	return ReconcileService(ctx, i, old, new, report)
}
