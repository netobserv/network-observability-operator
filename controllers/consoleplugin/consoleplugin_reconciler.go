package consoleplugin

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
)

// Reconciler reconciles the current goflow-kube state with the desired configuration
type Reconciler struct {
	client.Client
	SetControllerReference func(client.Object) error
	OperatorNamespace      string
}

const pluginName = "network-observability-plugin"

// Reconcile is the reconciler entry point to reconcile the current plugin state with the desired configuration
func (r *Reconciler) Reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollectorConsolePlugin) error {
	// Check if goflow-kube already exists, as a deployment or as a daemon set
	nsname := types.NamespacedName{Name: pluginName, Namespace: r.OperatorNamespace}

	// Get existing objects
	oldDepl, err := r.getObj(ctx, nsname, &appsv1.Deployment{})
	if err != nil {
		return err
	}
	oldSVC, err := r.getObj(ctx, nsname, &corev1.Service{})
	if err != nil {
		return err
	}

	// First deployment, creating permissions and container plugin
	if oldDepl == nil && oldSVC == nil {
		rbac := buildRBAC(r.OperatorNamespace)
		for _, rbacObject := range rbac {
			r.createOrUpdate(ctx, nil, rbacObject)
		}
		consolePlugin := buildConsolePlugin(desired, r.OperatorNamespace)
		r.createOrUpdate(ctx, nil, consolePlugin)
	}

	// Check if objects need update
	if oldDepl == nil || deploymentNeedsUpdate(oldDepl.(*appsv1.Deployment), desired) {
		newDepl := buildDeployment(desired, r.OperatorNamespace)
		r.createOrUpdate(ctx, oldDepl, newDepl)
	}
	if oldSVC == nil || serviceNeedsUpdate(oldSVC.(*corev1.Service), desired) {
		newSVC := buildService(desired, r.OperatorNamespace)
		r.createOrUpdate(ctx, oldSVC, newSVC)
	}
	return nil
}

func deploymentNeedsUpdate(depl *appsv1.Deployment, desired *flowsv1alpha1.FlowCollectorConsolePlugin) bool {
	return containerNeedsUpdate(&depl.Spec.Template.Spec, desired) ||
		*depl.Spec.Replicas != desired.Replicas
}

func serviceNeedsUpdate(svc *corev1.Service, desired *flowsv1alpha1.FlowCollectorConsolePlugin) bool {
	for _, port := range svc.Spec.Ports {
		if port.Port == desired.Port && port.Protocol == "TCP" {
			return false
		}
	}
	return true
}

func containerNeedsUpdate(podSpec *corev1.PodSpec, desired *flowsv1alpha1.FlowCollectorConsolePlugin) bool {
	container := findContainer(podSpec)
	if container == nil {
		return true
	}
	if desired.Image != container.Image || desired.ImagePullPolicy != string(container.ImagePullPolicy) {
		return true
	}
	if !reflect.DeepEqual(desired.Resources, container.Resources) {
		return true
	}
	return false
}

func findContainer(podSpec *corev1.PodSpec) *corev1.Container {
	for _, ctnr := range podSpec.Containers {
		if ctnr.Name == pluginName {
			return &ctnr
		}
	}
	return nil
}

func (r *Reconciler) getIfExist(ctx context.Context, nsname types.NamespacedName, obj client.Object) (client.Object, error) {
	err := r.Get(ctx, nsname, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			log.FromContext(ctx).Error(err, "Failed to get object")
			return nil, err
		}
	}
	return obj, nil
}

func (r *Reconciler) createOrUpdate(ctx context.Context, old, new client.Object) {
	log := log.FromContext(ctx)
	err := r.SetControllerReference(new)
	if err != nil {
		log.Error(err, "Failed to set controller reference")
	}
	if old == nil {
		log.Info("Creating a new object", "Namespace", new.GetNamespace(), "Name", new.GetName())
		err := r.Create(ctx, new)
		if err != nil {
			log.Error(err, "Failed to create new object", "Namespace", new.GetNamespace(), "Name", new.GetName())
			return
		}
	} else {
		log.Info("Updating object", "Namespace", new.GetNamespace(), "Name", new.GetName())
		err := r.Update(ctx, new)
		if err != nil {
			log.Error(err, "Failed to update object", "Namespace", new.GetNamespace(), "Name", new.GetName())
			return
		}
	}
}

func (r *Reconciler) getObj(ctx context.Context, nsname types.NamespacedName, obj client.Object) (client.Object, error) {
	err := r.Get(ctx, nsname, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			log.FromContext(ctx).Error(err, "Failed to get object", obj.GetName())
			return nil, err
		}
	}
	return obj, nil
}
