package reconcilers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/discover"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Common struct {
	helper.Client
	Watcher           *watchers.Watcher
	Namespace         string
	PreviousNamespace string
	UseOpenShiftSCC   bool
	AvailableAPIs     *discover.AvailableAPIs
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
	actual := rbacv1.ClusterRoleBinding{}
	if err := c.Get(ctx, types.NamespacedName{Name: desired.ObjectMeta.Name}, &actual); err != nil {
		if errors.IsNotFound(err) {
			return c.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile ClusterRoleBinding %s: %w", desired.Name, err)
	}
	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		actual.RoleRef == desired.RoleRef &&
		reflect.DeepEqual(actual.Subjects, desired.Subjects) {
		if actual.RoleRef != desired.RoleRef {
			//Roleref cannot be updated deleting and creating a new rolebinding
			log := log.FromContext(ctx)
			log.Info("Deleting old ClusterRoleBinding", "Namespace", actual.GetNamespace(), "Name", actual.GetName())
			err := c.Delete(ctx, &actual)
			if err != nil {
				log.Error(err, "error deleting old ClusterRoleBinding", "Namespace", actual.GetNamespace(), "Name", actual.GetName())
			}
			return c.CreateOwned(ctx, desired)
		}
		// cluster role binding already reconciled. Exiting
		return nil
	}
	return c.UpdateOwned(ctx, &actual, desired)
}

func (c *Common) ReconcileRoleBinding(ctx context.Context, desired *rbacv1.RoleBinding) error {
	actual := rbacv1.RoleBinding{}
	if err := c.Get(ctx, types.NamespacedName{Name: desired.ObjectMeta.Name}, &actual); err != nil {
		if errors.IsNotFound(err) {
			return c.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile RoleBinding %s: %w", desired.Name, err)
	}
	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		actual.RoleRef == desired.RoleRef &&
		reflect.DeepEqual(actual.Subjects, desired.Subjects) {
		if actual.RoleRef != desired.RoleRef {
			//Roleref cannot be updated deleting and creating a new rolebinding
			log := log.FromContext(ctx)
			log.Info("Deleting old RoleBinding", "Namespace", actual.GetNamespace(), "Name", actual.GetName())
			err := c.Delete(ctx, &actual)
			if err != nil {
				log.Error(err, "error deleting old RoleBinding", "Namespace", actual.GetNamespace(), "Name", actual.GetName())
			}
			return c.CreateOwned(ctx, desired)
		}
		// role binding already reconciled. Exiting
		return nil
	}
	return c.UpdateOwned(ctx, &actual, desired)
}

func (c *Common) ReconcileClusterRole(ctx context.Context, desired *rbacv1.ClusterRole) error {
	actual := rbacv1.ClusterRole{}
	if err := c.Get(ctx, types.NamespacedName{Name: desired.Name}, &actual); err != nil {
		if errors.IsNotFound(err) {
			return c.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile ClusterRole %s: %w", desired.Name, err)
	}

	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		reflect.DeepEqual(actual.Rules, desired.Rules) {
		// cluster role already reconciled. Exiting
		return nil
	}

	return c.UpdateOwned(ctx, &actual, desired)
}

func (c *Common) ReconcileRole(ctx context.Context, desired *rbacv1.Role) error {
	actual := rbacv1.Role{}
	if err := c.Get(ctx, types.NamespacedName{Name: desired.Name}, &actual); err != nil {
		if errors.IsNotFound(err) {
			return c.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile Role %s: %w", desired.Name, err)
	}

	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		reflect.DeepEqual(actual.Rules, desired.Rules) {
		// role already reconciled. Exiting
		return nil
	}

	return c.UpdateOwned(ctx, &actual, desired)
}

func (c *Common) ReconcileConfigMap(ctx context.Context, desired *corev1.ConfigMap, delete bool) error {
	actual := corev1.ConfigMap{}
	if err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &actual); err != nil {
		if errors.IsNotFound(err) {
			if delete {
				return nil
			}
			return c.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile Configmap %s: %w", desired.Name, err)
	}

	if delete {
		return c.Delete(ctx, desired)
	}

	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		reflect.DeepEqual(actual.Data, desired.Data) {
		// configmap already reconciled. Exiting
		return nil
	}

	return c.UpdateOwned(ctx, &actual, desired)
}

func (i *Instance) ReconcileService(ctx context.Context, old, new *corev1.Service, report *helper.ChangeReport) error {
	if !i.Managed.Exists(old) {
		if err := i.CreateOwned(ctx, new); err != nil {
			return err
		}
	} else if helper.ServiceChanged(old, new, report) {
		// In case we're updating an existing service, we need to build from the old one to keep immutable fields such as clusterIP
		newSVC := old.DeepCopy()
		newSVC.Spec.Ports = new.Spec.Ports
		newSVC.ObjectMeta.Annotations = new.ObjectMeta.Annotations
		if err := i.UpdateOwned(ctx, old, newSVC); err != nil {
			return err
		}
	}
	return nil
}

func GenericReconcile[K client.Object](ctx context.Context, m *NamespacedObjectManager, cl *helper.Client, old, new K, report *helper.ChangeReport, changeFunc func(old, new K, report *helper.ChangeReport) bool) error {
	if !m.Exists(old) {
		return cl.CreateOwned(ctx, new)
	}
	if changeFunc(old, new, report) {
		return cl.UpdateOwned(ctx, old, new)
	}
	return nil
}
