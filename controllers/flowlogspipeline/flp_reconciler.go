package flowlogspipeline

import (
	"context"
	"fmt"
	"reflect"

	"github.com/netobserv/network-observability-operator/pkg/helper"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
)

// Type alias
type flpSpec = flowsv1alpha1.FlowCollectorFLP

// FLPReconciler reconciles the current flowlogs-pipeline state with the desired configuration
type FLPReconciler struct {
	reconcilers.ClientHelper
	nobjMngr *reconcilers.NamespacedObjectManager
	owned    ownedObjects
}

type ownedObjects struct {
	deployment     *appsv1.Deployment
	daemonSet      *appsv1.DaemonSet
	service        *corev1.Service
	hpa            *ascv2.HorizontalPodAutoscaler
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
}

func NewReconciler(cl reconcilers.ClientHelper, ns, prevNS string) FLPReconciler {
	owned := ownedObjects{
		deployment:     &appsv1.Deployment{},
		daemonSet:      &appsv1.DaemonSet{},
		service:        &corev1.Service{},
		hpa:            &ascv2.HorizontalPodAutoscaler{},
		serviceAccount: &corev1.ServiceAccount{},
		configMap:      &corev1.ConfigMap{},
	}
	nobjMngr := reconcilers.NewNamespacedObjectManager(cl, ns, prevNS)
	nobjMngr.AddManagedObject(constants.FLPName, owned.deployment)
	nobjMngr.AddManagedObject(constants.FLPName, owned.daemonSet)
	nobjMngr.AddManagedObject(constants.FLPName, owned.service)
	nobjMngr.AddManagedObject(constants.FLPName, owned.hpa)
	nobjMngr.AddManagedObject(constants.FLPName, owned.serviceAccount)
	nobjMngr.AddManagedObject(configMapName, owned.configMap)

	return FLPReconciler{ClientHelper: cl, nobjMngr: nobjMngr, owned: owned}
}

// InitStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *FLPReconciler) InitStaticResources(ctx context.Context) error {
	return r.reconcilePermissions(ctx)
}

// PrepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *FLPReconciler) PrepareNamespaceChange(ctx context.Context) error {
	// Switching namespace => delete everything in the previous namespace
	r.nobjMngr.CleanupNamespace(ctx)
	return r.reconcilePermissions(ctx)
}

func validateDesired(desired *flpSpec) error {
	if desired.Port == 4789 ||
		desired.Port == 6081 ||
		desired.Port == 500 ||
		desired.Port == 4500 {
		return fmt.Errorf("flowlogs-pipeline port value is not authorized")
	}
	return nil
}

// Reconcile is the reconciler entry point to reconcile the current flowlogs-pipeline state with the desired configuration
func (r *FLPReconciler) Reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	desiredFLP := &desired.Spec.FlowlogsPipeline
	err := validateDesired(desiredFLP)
	if err != nil {
		return err
	}
	desiredLoki := &desired.Spec.Loki
	portProtocol := corev1.ProtocolUDP
	if desired.Spec.EBPF != nil {
		portProtocol = corev1.ProtocolTCP
	}
	builder := newBuilder(r.nobjMngr.Namespace, portProtocol, desiredFLP, desiredLoki)
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
	} else if !reflect.DeepEqual(newCM.Data, r.owned.configMap.Data) {
		if err := r.UpdateOwned(ctx, r.owned.configMap, newCM); err != nil {
			return err
		}
	}

	switch desiredFLP.Kind {
	case constants.DeploymentKind:
		return r.reconcileAsDeployment(ctx, desiredFLP, &builder, configDigest)
	case constants.DaemonSetKind:
		return r.reconcileAsDaemonSet(ctx, desiredFLP, &builder, configDigest)
	default:
		return fmt.Errorf("could not reconcile collector, invalid kind: %s", desiredFLP.Kind)
	}
}

func (r *FLPReconciler) reconcileAsDeployment(ctx context.Context, desiredFLP *flpSpec, builder *builder, configDigest string) error {
	// Kind may have changed: try delete DaemonSet and create Deployment+Service
	ns := r.nobjMngr.Namespace
	r.nobjMngr.TryDelete(ctx, r.owned.daemonSet)

	if !r.nobjMngr.Exists(r.owned.deployment) {
		if err := r.CreateOwned(ctx, builder.deployment(configDigest)); err != nil {
			return err
		}
	} else if deploymentNeedsUpdate(r.owned.deployment, desiredFLP, configDigest) {
		if err := r.UpdateOwned(ctx, r.owned.deployment, builder.deployment(configDigest)); err != nil {
			return err
		}
	}
	if !r.nobjMngr.Exists(r.owned.service) {
		newSVC := builder.service(nil)
		if err := r.CreateOwned(ctx, newSVC); err != nil {
			return err
		}
	} else if serviceNeedsUpdate(r.owned.service, desiredFLP) {
		newSVC := builder.service(r.owned.service)
		if err := r.UpdateOwned(ctx, r.owned.service, newSVC); err != nil {
			return err
		}
	}

	// Delete or Create / Update Autoscaler according to HPA option
	if desiredFLP.HPA == nil {
		r.nobjMngr.TryDelete(ctx, r.owned.hpa)
	} else if desiredFLP.HPA != nil {
		newASC := builder.autoScaler()
		if !r.nobjMngr.Exists(r.owned.hpa) {
			if err := r.CreateOwned(ctx, newASC); err != nil {
				return err
			}
		} else if autoScalerNeedsUpdate(r.owned.hpa, desiredFLP, ns) {
			if err := r.UpdateOwned(ctx, r.owned.hpa, newASC); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *FLPReconciler) reconcileAsDaemonSet(ctx context.Context, desiredFLP *flpSpec, builder *builder, configDigest string) error {
	// Kind may have changed: try delete Deployment / Service / HPA and create DaemonSet
	r.nobjMngr.TryDelete(ctx, r.owned.deployment)
	r.nobjMngr.TryDelete(ctx, r.owned.service)
	r.nobjMngr.TryDelete(ctx, r.owned.hpa)
	if !r.nobjMngr.Exists(r.owned.daemonSet) {
		return r.CreateOwned(ctx, builder.daemonSet(configDigest))
	} else if daemonSetNeedsUpdate(r.owned.daemonSet, desiredFLP, configDigest) {
		return r.UpdateOwned(ctx, r.owned.daemonSet, builder.daemonSet(configDigest))
	}
	return nil
}

func (r *FLPReconciler) reconcilePermissions(ctx context.Context) error {
	// Cluster role is only installed once
	if err := r.reconcileClusterRole(ctx); err != nil {
		return err
	}
	// Service account has to be re-created when namespace changes (it is namespace-scoped)
	if err := r.CreateOwned(ctx, buildServiceAccount(r.nobjMngr.Namespace)); err != nil {
		return err
	}
	// Cluster role binding has to be updated when namespace changes (it is not namespace-scoped)
	return r.reconcileClusterRoleBinding(ctx)
}

func (r *FLPReconciler) reconcileClusterRole(ctx context.Context) error {
	desired := buildClusterRole()
	actual := v1.ClusterRole{}
	if err := r.Client.Get(ctx,
		types.NamespacedName{Name: constants.FLPName},
		&actual,
	); err != nil {
		if errors.IsNotFound(err) {
			return r.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile %s ClusterRole: %w", constants.FLPName, err)
	}

	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		reflect.DeepEqual(actual.Rules, desired.Rules) {
		// cluster role already reconciled. Exiting
		return nil
	}

	return r.UpdateOwned(ctx, &actual, desired)
}

func (r *FLPReconciler) reconcileClusterRoleBinding(ctx context.Context) error {
	desired := buildClusterRoleBinding(r.nobjMngr.Namespace)
	actual := v1.ClusterRoleBinding{}
	if err := r.Client.Get(ctx,
		types.NamespacedName{Name: constants.FLPName},
		&actual,
	); err != nil {
		if errors.IsNotFound(err) {
			return r.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile %s ClusterRoleBinding: %w", constants.FLPName, err)
	}
	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		actual.RoleRef == desired.RoleRef &&
		reflect.DeepEqual(actual.Subjects, desired.Subjects) {
		// cluster role binding already reconciled. Exiting
		return nil
	}
	return r.UpdateOwned(ctx, &actual, desired)
}

func daemonSetNeedsUpdate(ds *appsv1.DaemonSet, desired *flpSpec, configDigest string) bool {
	return containerNeedsUpdate(&ds.Spec.Template.Spec, desired, true) ||
		configChanged(&ds.Spec.Template, configDigest)
}

func deploymentNeedsUpdate(depl *appsv1.Deployment, desired *flpSpec, configDigest string) bool {
	return containerNeedsUpdate(&depl.Spec.Template.Spec, desired, false) ||
		configChanged(&depl.Spec.Template, configDigest) ||
		(desired.HPA == nil && *depl.Spec.Replicas != desired.Replicas)
}

func configChanged(tmpl *corev1.PodTemplateSpec, configDigest string) bool {
	return tmpl.Annotations == nil || tmpl.Annotations[PodConfigurationDigest] != configDigest
}

func serviceNeedsUpdate(svc *corev1.Service, desired *flpSpec) bool {
	for _, port := range svc.Spec.Ports {
		if port.Port == desired.Port && port.Protocol == corev1.ProtocolUDP {
			return false
		}
	}
	return true
}

func containerNeedsUpdate(podSpec *corev1.PodSpec, desired *flpSpec, expectHostPort bool) bool {
	// Note, we don't check for changed port / host port here, because that would change also the configmap,
	//	which also triggers pod update anyway
	container := reconcilers.FindContainer(podSpec, constants.FLPName)
	return container == nil ||
		desired.Image != container.Image ||
		desired.ImagePullPolicy != string(container.ImagePullPolicy) ||
		probesNeedUpdate(container, desired.EnableKubeProbes) ||
		!reflect.DeepEqual(desired.Resources, container.Resources)
}

func probesNeedUpdate(container *corev1.Container, enabled bool) bool {
	if enabled {
		return container.LivenessProbe == nil || container.StartupProbe == nil
	}
	return container.LivenessProbe != nil || container.StartupProbe != nil
}

func autoScalerNeedsUpdate(asc *ascv2.HorizontalPodAutoscaler, desired *flpSpec, ns string) bool {
	if asc.Namespace != ns {
		return true
	}
	differentPointerValues := func(a, b *int32) bool {
		return (a == nil && b != nil) || (a != nil && b == nil) || (a != nil && *a != *b)
	}
	if asc.Spec.MaxReplicas != desired.HPA.MaxReplicas ||
		differentPointerValues(asc.Spec.MinReplicas, desired.HPA.MinReplicas) {
		return true
	}
	if !reflect.DeepEqual(asc.Spec.Metrics, desired.HPA.Metrics) {
		return true
	}
	return false
}
