package consoleplugin

import (
	"context"
	"reflect"

	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
)

const pluginName = "network-observability-plugin"

// Type alias
type pluginSpec = flowsv1alpha1.FlowCollectorConsolePlugin

// CPReconciler reconciles the current console plugin state with the desired configuration
type CPReconciler struct {
	reconcilers.ClientHelper
	nobjMngr *reconcilers.NamespacedObjectManager
	owned    ownedObjects
}

type ownedObjects struct {
	deployment     *appsv1.Deployment
	service        *corev1.Service
	serviceAccount *corev1.ServiceAccount
}

func NewReconciler(cl reconcilers.ClientHelper, ns, prevNS string) CPReconciler {
	owned := ownedObjects{
		deployment:     &appsv1.Deployment{},
		service:        &corev1.Service{},
		serviceAccount: &corev1.ServiceAccount{},
	}
	nobjMngr := reconcilers.NewNamespacedObjectManager(cl, ns, prevNS)
	nobjMngr.AddManagedObject(pluginName, owned.deployment)
	nobjMngr.AddManagedObject(pluginName, owned.service)
	nobjMngr.AddManagedObject(pluginName, owned.serviceAccount)

	return CPReconciler{ClientHelper: cl, nobjMngr: nobjMngr, owned: owned}
}

// InitStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *CPReconciler) InitStaticResources(ctx context.Context) error {
	return r.CreateOwned(ctx, buildServiceAccount(r.nobjMngr.Namespace))
}

// PrepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *CPReconciler) PrepareNamespaceChange(ctx context.Context) error {
	// Switching namespace => delete everything in the previous namespace
	r.nobjMngr.CleanupNamespace(ctx)
	return r.CreateOwned(ctx, buildServiceAccount(r.nobjMngr.Namespace))
}

// Reconcile is the reconciler entry point to reconcile the current plugin state with the desired configuration
func (r *CPReconciler) Reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollectorSpec) error {
	ns := r.nobjMngr.Namespace
	// Retrieve current owned objects
	err := r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Console plugin is cluster-scope (it's not deployed in our namespace) however it must still be updated if our namespace changes
	oldPlg := osv1alpha1.ConsolePlugin{}
	pluginExists := true
	err = r.Get(ctx, types.NamespacedName{Name: pluginName}, &oldPlg)
	if err != nil {
		if errors.IsNotFound(err) {
			pluginExists = false
		} else {
			return err
		}
	}

	// Create object builder
	builder := newBuilder(ns, &desired.ConsolePlugin, &desired.Loki)

	// Check if objects need update
	consolePlugin := builder.consolePlugin()
	if !pluginExists {
		if err := r.CreateOwned(ctx, consolePlugin); err != nil {
			return err
		}
	} else if pluginNeedsUpdate(&oldPlg, &desired.ConsolePlugin, ns) {
		if err := r.UpdateOwned(ctx, &oldPlg, consolePlugin); err != nil {
			return err
		}
	}

	newDepl := builder.deployment()
	if !r.nobjMngr.Exists(r.owned.deployment) {
		if err := r.CreateOwned(ctx, newDepl); err != nil {
			return err
		}
	} else if deploymentNeedsUpdate(r.owned.deployment, desired, ns) {
		if err := r.UpdateOwned(ctx, r.owned.deployment, newDepl); err != nil {
			return err
		}
	}

	if !r.nobjMngr.Exists(r.owned.service) {
		newSVC := builder.service(nil)
		if err := r.CreateOwned(ctx, newSVC); err != nil {
			return err
		}
	} else if serviceNeedsUpdate(r.owned.service, &desired.ConsolePlugin, ns) {
		newSVC := builder.service(r.owned.service)
		if err := r.UpdateOwned(ctx, r.owned.service, newSVC); err != nil {
			return err
		}
	}
	return nil
}

func pluginNeedsUpdate(plg *osv1alpha1.ConsolePlugin, desired *pluginSpec, ns string) bool {
	return plg.Spec.Service.Namespace != ns ||
		plg.Spec.Service.Port != desired.Port
}

func deploymentNeedsUpdate(depl *appsv1.Deployment, desired *flowsv1alpha1.FlowCollectorSpec, ns string) bool {
	if depl.Namespace != ns {
		return true
	}
	return containerNeedsUpdate(&depl.Spec.Template.Spec, &desired.ConsolePlugin) ||
		hasLokiURLChanged(depl, &desired.Loki) ||
		*depl.Spec.Replicas != desired.ConsolePlugin.Replicas
}

func hasLokiURLChanged(depl *appsv1.Deployment, loki *flowsv1alpha1.FlowCollectorLoki) bool {
	return depl.Annotations[lokiURLAnnotation] != querierURL(loki)
}

func querierURL(loki *flowsv1alpha1.FlowCollectorLoki) string {
	if loki.QuerierURL != "" {
		return loki.QuerierURL
	}
	return loki.URL
}

func serviceNeedsUpdate(svc *corev1.Service, desired *flowsv1alpha1.FlowCollectorConsolePlugin, ns string) bool {
	if svc.Namespace != ns {
		return true
	}
	for _, port := range svc.Spec.Ports {
		if port.Port == desired.Port && port.Protocol == "TCP" {
			return false
		}
	}
	return true
}

func containerNeedsUpdate(podSpec *corev1.PodSpec, desired *pluginSpec) bool {
	container := reconcilers.FindContainer(podSpec, pluginName)
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
