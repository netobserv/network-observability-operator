package reconcilers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/netobserv/network-observability-operator/pkg/conditions"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
}

// CreateOwned is an helper function that creates an object, sets owner reference and writes info & errors logs
func (c *ClientHelper) CreateOwned(ctx context.Context, obj client.Object) error {
	log := log.FromContext(ctx)
	c.changed = true
	err := c.SetControllerReference(obj)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to set controller reference of %s", obj.GetName()))
		return err
	}

	objectType := reflect.TypeOf(obj)
	kind := objectType.String()
	if objectType == reflect.TypeOf(&unstructured.Unstructured{}) {
		kind = obj.(*unstructured.Unstructured).GetKind()
	}
	log.Info(fmt.Sprintf("Creating %s Namespace %s Name %s", kind, obj.GetNamespace(), obj.GetName()))
	err = c.Create(ctx, obj)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to create %s Namespace %s Name %s", kind, obj.GetNamespace(), obj.GetName()))
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
		log.Error(err, fmt.Sprintf("Failed to set controller reference of %s", obj.GetName()))
		return err
	}
	kind := reflect.TypeOf(obj).String()
	log.Info(fmt.Sprintf("Updating %s Namespace %s Name %s", kind, obj.GetNamespace(), obj.GetName()))
	err = c.Update(ctx, obj)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to update %s Namespace %s Name %s", kind, obj.GetNamespace(), obj.GetName()))
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

// FindContainer searches in pod containers one that matches the provided name
func FindContainer(podSpec *corev1.PodSpec, name string) *corev1.Container {
	for i := range podSpec.Containers {
		if podSpec.Containers[i].Name == name {
			return &podSpec.Containers[i]
		}
	}
	return nil
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

// Apply a given manifest with an optional namespace to override
func (c *ClientHelper) Apply(ctx context.Context, u *unstructured.Unstructured) error {
	log := log.FromContext(ctx, "component", "ClientHelper", "function", "Apply")

	err := c.Get(ctx, types.NamespacedName{
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
	}, u)
	if errors.IsNotFound(err) {
		log.Info(u.GetName() + " in ns " + u.GetNamespace() + " is not found")
		return c.Create(ctx, u)
	}
	return err
}

func (c *ClientHelper) GetClientCondition() metav1.Condition {
	var condition metav1.Condition
	if c.DidChange() {
		condition = conditions.Updating()
	} else if c.IsInProgress() {
		condition = conditions.DeploymentInProgress()
	} else {
		// TODO: update this to get deployments statuses instead of considering our components as 'ready'
		condition = conditions.Ready()
	}
	return condition
}
