package reconcilers

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClientHelper includes a kube client with some additional helper functions
type ClientHelper struct {
	client.Client
	SetControllerReference func(client.Object) error
}

// CreateOwned is an helper function that creates an object, sets owner reference and writes info & errors logs
func (c *ClientHelper) CreateOwned(ctx context.Context, obj client.Object) error {
	log := log.FromContext(ctx)
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

// FindContainer searches in pod containers one that matches the provided name
func FindContainer(podSpec *corev1.PodSpec, name string) *corev1.Container {
	for i := range podSpec.Containers {
		if podSpec.Containers[i].Name == name {
			return &podSpec.Containers[i]
		}
	}
	return nil
}
