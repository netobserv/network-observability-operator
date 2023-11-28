package reconcilers

import (
	"context"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/discover"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type Common struct {
	helper.Client
	Status            status.Instance
	Watcher           *watchers.Watcher
	Namespace         string
	PreviousNamespace string
	UseOpenShiftSCC   bool
	AvailableAPIs     *discover.AvailableAPIs
	Loki              *helper.LokiConfig
	ClusterID         string
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
	Image   string
}

func (c *Common) NewInstance(image string) *Instance {
	managed := NewNamespacedObjectManager(c)
	return &Instance{
		Common:  c,
		Managed: managed,
		Image:   image,
	}
}

func (c *Common) ReconcileClusterRoleBinding(ctx context.Context, desired *rbacv1.ClusterRoleBinding) error {
	return ReconcileClusterRoleBinding(ctx, &c.Client, desired)
}

func (c *Common) ReconcileRoleBinding(ctx context.Context, desired *rbacv1.RoleBinding) error {
	return ReconcileRoleBinding(ctx, &c.Client, desired)
}

func (c *Common) ReconcileClusterRole(ctx context.Context, desired *rbacv1.ClusterRole) error {
	return ReconcileClusterRole(ctx, &c.Client, desired)
}

func (c *Common) ReconcileRole(ctx context.Context, desired *rbacv1.Role) error {
	return ReconcileRole(ctx, &c.Client, desired)
}

func (c *Common) ReconcileConfigMap(ctx context.Context, desired *corev1.ConfigMap, delete bool) error {
	return ReconcileConfigMap(ctx, &c.Client, desired, delete)
}

func (i *Instance) ReconcileService(ctx context.Context, old, new *corev1.Service, report *helper.ChangeReport) error {
	return ReconcileService(ctx, i, old, new, report)
}
