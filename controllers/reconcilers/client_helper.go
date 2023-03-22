package reconcilers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClientHelper includes a kube client with some additional helper functions
type ClientHelper struct {
	client.Client
	SetControllerReference func(client.Object) error
	changed                bool
	deplInProgress         bool
	CertWatcher            *watchers.CertificatesWatcher
}

// CreateOwned is an helper function that creates an object, sets owner reference and writes info & errors logs
func (c *ClientHelper) CreateOwned(ctx context.Context, obj client.Object) error {
	log := log.FromContext(ctx)
	c.changed = true
	err := c.SetControllerReference(obj)
	if err != nil {
		log.Error(err, "Failed to set controller reference")
		return err
	}
	kind := reflect.TypeOf(obj).String()
	log.Info("Creating a new "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
	err = c.Create(ctx, obj)
	if err != nil {
		log.Error(err, "Failed to create new "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		return err
	}
	return nil
}

// UpdateOwned is an helper function that updates an object, sets owner reference and writes info & errors logs
func (c *ClientHelper) UpdateOwned(ctx context.Context, old, obj client.Object) error {
	log := log.FromContext(ctx)
	c.changed = true
	if old != nil {
		obj.SetResourceVersion(old.GetResourceVersion())
	}
	err := c.SetControllerReference(obj)
	if err != nil {
		log.Error(err, "Failed to set controller reference")
		return err
	}
	kind := reflect.TypeOf(obj).String()
	log.Info("Updating "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
	err = c.Update(ctx, obj)
	if err != nil {
		log.Error(err, "Failed to update "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		return err
	}
	return nil
}

func (c *ClientHelper) DidChange() bool {
	return c.changed
}

func (c *ClientHelper) IsInProgress() bool {
	return c.deplInProgress
}

func (c *ClientHelper) CheckDeploymentInProgress(d *appsv1.Deployment) {
	if d.Status.AvailableReplicas < d.Status.Replicas {
		c.deplInProgress = true
	}
}

func (c *ClientHelper) CheckDaemonSetInProgress(ds *appsv1.DaemonSet) {
	if ds.Status.NumberAvailable < ds.Status.DesiredNumberScheduled {
		c.deplInProgress = true
	}
}

func (c *ClientHelper) ReconcileClusterRoleBinding(ctx context.Context, desired *rbacv1.ClusterRoleBinding) error {
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

func (c *ClientHelper) ReconcileRoleBinding(ctx context.Context, desired *rbacv1.RoleBinding) error {
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

func (c *ClientHelper) ReconcileClusterRole(ctx context.Context, desired *rbacv1.ClusterRole) error {
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

func (c *ClientHelper) ReconcileRole(ctx context.Context, desired *rbacv1.Role) error {
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

func (c *ClientHelper) ReconcileConfigMap(ctx context.Context, desired *corev1.ConfigMap) error {
	actual := corev1.ConfigMap{}
	if err := c.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &actual); err != nil {
		if errors.IsNotFound(err) {
			return c.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile Configmap %s: %w", desired.Name, err)
	}

	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		reflect.DeepEqual(actual.Data, desired.Data) {
		// configmap already reconciled. Exiting
		return nil
	}

	return c.UpdateOwned(ctx, &actual, desired)
}
