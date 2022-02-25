package goflowkube

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	ascv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
)

// Type alias
type goflowKubeSpec = flowsv1alpha1.FlowCollectorGoflowKube
type lokiSpec = flowsv1alpha1.FlowCollectorLoki

// GFKReconciler reconciles the current goflow-kube state with the desired configuration
type GFKReconciler struct {
	reconcilers.ClientHelper
	nobjMngr *reconcilers.NamespacedObjectManager
	owned    ownedObjects
}

type ownedObjects struct {
	deployment     *appsv1.Deployment
	daemonSet      *appsv1.DaemonSet
	service        *corev1.Service
	hpa            *ascv1.HorizontalPodAutoscaler
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
}

func NewReconciler(cl reconcilers.ClientHelper, ns, prevNS string) GFKReconciler {
	owned := ownedObjects{
		deployment:     &appsv1.Deployment{},
		daemonSet:      &appsv1.DaemonSet{},
		service:        &corev1.Service{},
		hpa:            &ascv1.HorizontalPodAutoscaler{},
		serviceAccount: &corev1.ServiceAccount{},
		configMap:      &corev1.ConfigMap{},
	}
	nobjMngr := reconcilers.NewNamespacedObjectManager(cl, ns, prevNS)
	nobjMngr.AddManagedObject(constants.GoflowKubeName, owned.deployment)
	nobjMngr.AddManagedObject(constants.GoflowKubeName, owned.daemonSet)
	nobjMngr.AddManagedObject(constants.GoflowKubeName, owned.service)
	nobjMngr.AddManagedObject(constants.GoflowKubeName, owned.hpa)
	nobjMngr.AddManagedObject(constants.GoflowKubeName, owned.serviceAccount)
	nobjMngr.AddManagedObject(configMapName, owned.configMap)

	return GFKReconciler{ClientHelper: cl, nobjMngr: nobjMngr, owned: owned}
}

// InitStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *GFKReconciler) InitStaticResources(ctx context.Context) error {
	return r.createPermissions(ctx, true)
}

// PrepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *GFKReconciler) PrepareNamespaceChange(ctx context.Context) error {
	// Switching namespace => delete everything in the previous namespace
	r.nobjMngr.CleanupNamespace(ctx)
	return r.createPermissions(ctx, false)
}

func validateDesired(desiredGoflowKube *goflowKubeSpec) error {
	if desiredGoflowKube.Port == 4789 ||
		desiredGoflowKube.Port == 6081 ||
		desiredGoflowKube.Port == 500 ||
		desiredGoflowKube.Port == 4500 {
		return fmt.Errorf("goflowkube port value is not authorized")
	}
	return nil
}

// Reconcile is the reconciler entry point to reconcile the current goflow-kube state with the desired configuration
func (r *GFKReconciler) Reconcile(ctx context.Context, desiredGoflowKube *goflowKubeSpec, desiredLoki *lokiSpec) error {
	err := validateDesired(desiredGoflowKube)
	if err != nil {
		return err
	}

	builder := newBuilder(r.nobjMngr.Namespace, desiredGoflowKube, desiredLoki)
	// Retrieve current owned objects
	err = r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}
	newCM, configDigest := builder.configMap()
	if !r.nobjMngr.Exists(r.owned.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return err
		}
	} else if !reflect.DeepEqual(newCM, r.owned.configMap.Data) {
		if err := r.UpdateOwned(ctx, r.owned.configMap, newCM); err != nil {
			return err
		}
	}

	switch desiredGoflowKube.Kind {
	case constants.DeploymentKind:
		return r.reconcileAsDeployment(ctx, desiredGoflowKube, &builder, configDigest)
	case constants.DaemonSetKind:
		return r.reconcileAsDaemonSet(ctx, desiredGoflowKube, &builder, configDigest)
	default:
		return fmt.Errorf("could not reconcile collector, invalid kind: %s", desiredGoflowKube.Kind)
	}
}

func (r *GFKReconciler) reconcileAsDeployment(ctx context.Context, desiredGoflowKube *goflowKubeSpec, builder *builder, configDigest string) error {
	// Kind changed: delete DaemonSet and create Deployment+Service
	ns := r.nobjMngr.Namespace
	r.nobjMngr.TryDelete(ctx, r.owned.daemonSet)

	newDepl := builder.deployment(configDigest)
	if !r.nobjMngr.Exists(r.owned.deployment) {
		if err := r.CreateOwned(ctx, newDepl); err != nil {
			return err
		}
	} else if deploymentNeedsUpdate(r.owned.deployment, desiredGoflowKube, ns, configDigest) {
		if err := r.UpdateOwned(ctx, r.owned.deployment, newDepl); err != nil {
			return err
		}
	}
	if !r.nobjMngr.Exists(r.owned.service) {
		newSVC := builder.service(nil)
		if err := r.CreateOwned(ctx, newSVC); err != nil {
			return err
		}
	} else if serviceNeedsUpdate(r.owned.service, desiredGoflowKube, ns) {
		newSVC := builder.service(r.owned.service)
		if err := r.UpdateOwned(ctx, r.owned.service, newSVC); err != nil {
			return err
		}
	}

	// Delete or Create / Update Autoscaler according to HPA option
	if desiredGoflowKube.HPA == nil {
		r.nobjMngr.TryDelete(ctx, r.owned.hpa)
	} else if desiredGoflowKube.HPA != nil {
		newASC := builder.autoScaler()
		if !r.nobjMngr.Exists(r.owned.hpa) {
			if err := r.CreateOwned(ctx, newASC); err != nil {
				return err
			}
		} else if autoScalerNeedsUpdate(r.owned.hpa, desiredGoflowKube, ns) {
			if err := r.UpdateOwned(ctx, r.owned.hpa, newASC); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *GFKReconciler) reconcileAsDaemonSet(ctx context.Context, desiredGoflowKube *goflowKubeSpec, builder *builder, configDigest string) error {
	// Kind changed: delete Deployment / Service / HPA and create DaemonSet
	ns := r.nobjMngr.Namespace
	r.nobjMngr.TryDelete(ctx, r.owned.deployment)
	r.nobjMngr.TryDelete(ctx, r.owned.service)
	r.nobjMngr.TryDelete(ctx, r.owned.hpa)
	newDS := builder.daemonSet(configDigest)
	if !r.nobjMngr.Exists(r.owned.daemonSet) {
		if err := r.CreateOwned(ctx, newDS); err != nil {
			return err
		}
	} else if daemonSetNeedsUpdate(r.owned.daemonSet, desiredGoflowKube, ns, configDigest) {
		if err := r.UpdateOwned(ctx, r.owned.daemonSet, newDS); err != nil {
			return err
		}
	}
	return nil
}

func (r *GFKReconciler) createPermissions(ctx context.Context, firstInstall bool) error {
	// Cluster role is only installed once
	if firstInstall {
		if err := r.CreateOwned(ctx, buildClusterRole()); err != nil {
			return err
		}
	}
	// Service account has to be re-created when namespace changes (it is namespace-scoped)
	if err := r.CreateOwned(ctx, buildServiceAccount(r.nobjMngr.Namespace)); err != nil {
		return err
	}
	// Cluster role binding has to be updated when namespace changes (it is not namespace-scoped)
	if firstInstall {
		if err := r.CreateOwned(ctx, buildClusterRoleBinding(r.nobjMngr.Namespace)); err != nil {
			return err
		}
	} else {
		if err := r.UpdateOwned(ctx, nil, buildClusterRoleBinding(r.nobjMngr.Namespace)); err != nil {
			return err
		}
	}
	return nil
}

func daemonSetNeedsUpdate(ds *appsv1.DaemonSet, desired *goflowKubeSpec, ns, configDigest string) bool {
	if ds.Namespace != ns {
		return true
	}
	return containerNeedsUpdate(&ds.Spec.Template.Spec, desired) ||
		configChanged(&ds.Spec.Template, configDigest)
}

func deploymentNeedsUpdate(depl *appsv1.Deployment, desired *goflowKubeSpec, ns, configDigest string) bool {
	if depl.Namespace != ns {
		return true
	}
	return containerNeedsUpdate(&depl.Spec.Template.Spec, desired) ||
		configChanged(&depl.Spec.Template, configDigest) ||
		(desired.HPA == nil && *depl.Spec.Replicas != desired.Replicas)
}

func configChanged(tmpl *corev1.PodTemplateSpec, configDigest string) bool {
	return tmpl.Annotations == nil || tmpl.Annotations[PodConfigurationDigest] != configDigest
}

func serviceNeedsUpdate(svc *corev1.Service, desired *goflowKubeSpec, ns string) bool {
	if svc.Namespace != ns {
		return true
	}
	for _, port := range svc.Spec.Ports {
		if port.Port == desired.Port && port.Protocol == corev1.ProtocolUDP {
			return false
		}
	}
	return true
}

func containerNeedsUpdate(podSpec *corev1.PodSpec, desired *goflowKubeSpec) bool {
	container := reconcilers.FindContainer(podSpec, constants.GoflowKubeName)
	if container == nil {
		return true
	}
	if desired.Image != container.Image || desired.ImagePullPolicy != string(container.ImagePullPolicy) {
		return true
	}
	if !reflect.DeepEqual(desired.Resources, container.Resources) {
		return true
	}
	if len(container.Command) != 3 || container.Command[2] != buildMainCommand(desired) {
		return true
	}
	return false
}

func autoScalerNeedsUpdate(asc *ascv1.HorizontalPodAutoscaler, desired *goflowKubeSpec, ns string) bool {
	if asc.Namespace != ns {
		return true
	}
	differentPointerValues := func(a, b *int32) bool {
		return (a == nil && b != nil) || (a != nil && b == nil) || (a != nil && *a != *b)
	}
	if asc.Spec.MaxReplicas != desired.HPA.MaxReplicas ||
		differentPointerValues(asc.Spec.MinReplicas, desired.HPA.MinReplicas) ||
		differentPointerValues(asc.Spec.TargetCPUUtilizationPercentage, desired.HPA.TargetCPUUtilizationPercentage) {
		return true
	}
	return false
}
