package helper

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/netobserv-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/netobserv-operator/internal/controller/constants"
)

// Client includes a kube client with some additional helper functions
type Client struct {
	client.Client
	SetOwnerReference func(client.Object) error
	IsOwned           func(client.Object) bool
}

func UnmanagedClient(cl client.Client) Client {
	return Client{
		Client:            cl,
		SetOwnerReference: func(_ client.Object) error { return nil },
		IsOwned:           func(_ client.Object) bool { return false },
	}
}

func NewControllerClientHelper(ctx context.Context, ns string, c client.Client) (*Client, error) {
	dpl, err := getControllerDeployment(ctx, ns, c)
	if err != nil {
		return nil, err
	}
	// Hack: use a controller proxy for cluster-scope resources owner references.
	// Because the controller deployment is namespace-scoped, it cannot be defined as the owner of a cluster-scope resource.
	// So we'll use something else instead, that is part of the installed bundle, and that will allow cascading deletion down to statically created
	// cluster scope resources during uninstallation. Any bundled ClusterRole resource, for instance, can be used as such proxy.
	clusterProxy, err := getControllerClusterScopeProxy(ctx, c)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: c,
		SetOwnerReference: func(obj client.Object) error {
			if obj.GetNamespace() == "" {
				return controllerutil.SetControllerReference(clusterProxy, obj, c.Scheme(), controllerutil.WithBlockOwnerDeletion(false))
			}
			return controllerutil.SetControllerReference(dpl, obj, c.Scheme(), controllerutil.WithBlockOwnerDeletion(false))
		},
		IsOwned: isOwnedByController,
	}, nil
}

func NewFlowCollectorClientHelper(ctx context.Context, c client.Client) (*Client, *flowslatest.FlowCollector, error) {
	fc, err := getFlowCollector(ctx, c)
	if err != nil || fc == nil {
		return nil, fc, err
	}
	return &Client{
		Client: c,
		SetOwnerReference: func(obj client.Object) error {
			return controllerutil.SetControllerReference(fc, obj, c.Scheme())
		},
		IsOwned: IsOwned,
	}, fc, nil
}

// CreateOwned is an helper function that creates an object, sets owner reference and writes info & errors logs
func (c *Client) CreateOwned(ctx context.Context, obj client.Object) error {
	log := log.FromContext(ctx)
	err := c.SetOwnerReference(obj)
	if err != nil {
		log.Error(err, "Failed to set controller reference")
		return err
	}
	AddManagedLabel(obj)
	kind := reflect.TypeOf(obj).String()
	log.Info("CREATING a new "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
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
	if old != nil {
		obj.SetResourceVersion(old.GetResourceVersion())
	}
	err := c.SetOwnerReference(obj)
	if err != nil {
		log.Error(err, "Failed to set controller reference")
		return err
	}
	AddManagedLabel(obj)
	kind := reflect.TypeOf(obj).String()
	log.Info("UPDATING "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
	err = c.Update(ctx, obj)
	if err != nil {
		log.Error(err, "Failed to update "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		return err
	}
	err = c.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err != nil {
		log.Error(err, "Failed to get updated resource "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		return err
	}
	if obj.GetResourceVersion() == old.GetResourceVersion() {
		log.Info(kind+" not updated", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
	}
	return nil
}

// UpdateIfOwned is an helper function that updates an object if currently owned and managed by the operator
func (c *Client) UpdateIfOwned(ctx context.Context, old, obj client.Object) error {
	log := log.FromContext(ctx)

	if old != nil && !c.IsOwned(old) {
		kind := reflect.TypeOf(obj).String()
		log.Info("SKIP "+kind+" update since not owned", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		return nil
	}
	return c.UpdateOwned(ctx, old, obj)
}

// DeleteIfOwned is an helper function that deletes an object only if it's currently owned and managed by the operator
func (c *Client) DeleteIfOwned(ctx context.Context, obj client.Object) error {
	log := log.FromContext(ctx)
	kind := reflect.TypeOf(obj).String()

	if obj == nil {
		return nil
	}

	if !c.IsOwned(obj) {
		log.Info("SKIP "+kind+" deletion since not owned", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		return nil
	}

	log.Info("DELETING "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
	err := c.Delete(ctx, obj)
	if err != nil {
		log.Error(err, "Failed to delete "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		return err
	}
	return nil
}

func getFlowCollector(ctx context.Context, c client.Client) (*flowslatest.FlowCollector, error) {
	log := log.FromContext(ctx)
	desired := &flowslatest.FlowCollector{}
	if err := c.Get(ctx, constants.FlowCollectorName, desired); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("FlowCollector resource not found. Ignoring since object must be deleted")
			return nil, nil
		}
		// Error reading the object - requeue the request.
		return nil, err
	}
	return desired, nil
}

func getControllerDeployment(ctx context.Context, ns string, c client.Client) (*appsv1.Deployment, error) {
	dpl := &appsv1.Deployment{}
	if err := c.Get(ctx, types.NamespacedName{
		Name:      constants.ControllerName,
		Namespace: ns,
	}, dpl); err != nil {
		return nil, err
	}
	return dpl, nil
}

const (
	proxyRoleName = string(constants.FLPInformersRole)
)

func getControllerClusterScopeProxy(ctx context.Context, c client.Client) (client.Object, error) {
	// Pick any bundled ClusterRole as a proxy; do not use a CSV-defined role because their name is not guaranteed to be stable across cluster/OLM versions
	cr := &rbacv1.ClusterRole{}
	if err := c.Get(ctx, types.NamespacedName{Name: proxyRoleName}, cr); err != nil {
		return nil, err
	}
	return cr, nil
}

func isOwnedByController(obj client.Object) bool {
	// ownership is forced if netobserv-managed label is explicitly set to true
	if IsManaged(obj) {
		return true
	}
	// else we check for owner references
	refs := obj.GetOwnerReferences()
	return len(refs) > 0 && (refs[0].Name == constants.ControllerName || refs[0].Name == proxyRoleName)
}
