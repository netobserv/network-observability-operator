package helper

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Client includes a kube client with some additional helper functions
type Client struct {
	client.Client
	SetControllerReference func(client.Object) error
	SetChanged             func(bool)
	SetInProgress          func(bool)
}

func UnmanagedClient(cl client.Client) Client {
	return Client{
		Client:                 cl,
		SetControllerReference: func(o client.Object) error { return nil },
		SetChanged:             func(b bool) {},
		SetInProgress:          func(b bool) {},
	}
}

// CreateOwned is an helper function that creates an object, sets owner reference and writes info & errors logs
func (c *Client) CreateOwned(ctx context.Context, obj client.Object) error {
	log := log.FromContext(ctx)
	c.SetChanged(true)
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
func (c *Client) UpdateOwned(ctx context.Context, old, obj client.Object) error {
	log := log.FromContext(ctx)
	c.SetChanged(true)
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

func (c *Client) CheckDeploymentInProgress(d *appsv1.Deployment) {
	if d.Status.AvailableReplicas < d.Status.Replicas {
		c.SetInProgress(true)
	}
}

func (c *Client) CheckDaemonSetInProgress(ds *appsv1.DaemonSet) {
	if ds.Status.NumberAvailable < ds.Status.DesiredNumberScheduled {
		c.SetInProgress(true)
	}
}
