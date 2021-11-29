package goflowkube

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	ascv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
)

// Reconciler reconciles the current goflow-kube state with the desired configuration
type Reconciler struct {
	client.Client
	SetControllerReference func(client.Object) error
	Namespace              string
}

// Reconcile is the reconciler entry point to reconcile the current goflow-kube state with the desired configuration
func (r *Reconciler) Reconcile(ctx context.Context,
	desiredGoflowKube *flowsv1alpha1.FlowCollectorGoflowKube,
	desiredLoki *flowsv1alpha1.FlowCollectorLoki) error {

	// Check if goflow-kube already exists, as a deployment or as a daemon set
	nsname := types.NamespacedName{Name: constants.GoflowKubeName, Namespace: r.Namespace}
	oldDepl, err := r.getObj(ctx, nsname, &appsv1.Deployment{}, constants.DeploymentKind)
	if err != nil {
		return err
	}
	oldDS, err := r.getObj(ctx, nsname, &appsv1.DaemonSet{}, constants.DaemonSetKind)
	if err != nil {
		return err
	}
	// If none of them already exist, it must be the first setup. Thus, setup permissions.
	if oldDepl == nil && oldDS == nil {
		r.setupPermissions(ctx, r.Namespace)
	}
	oldSVC, err := r.getObj(ctx, nsname, &corev1.Service{}, constants.ServiceKind)
	if err != nil {
		return err
	}
	oldASC, err := r.getObj(ctx, nsname, &ascv1.HorizontalPodAutoscaler{}, constants.AutoscalerKind)
	if err != nil {
		return err
	}
	oldCM, err := r.getObj(ctx, types.NamespacedName{Name: configMapName, Namespace: r.Namespace}, &corev1.ConfigMap{}, constants.ConfigMapKind)
	if err != nil {
		return err
	}
	newCM := buildConfigMap(desiredGoflowKube, desiredLoki, r.Namespace)
	if oldCM == nil || !reflect.DeepEqual(newCM, oldCM.(*corev1.ConfigMap).Data) {
		r.createOrUpdate(ctx, oldCM, newCM, constants.ConfigMapKind)
	}

	switch desiredGoflowKube.Kind {
	case constants.DeploymentKind:
		// Kind changed: delete DaemonSet and create Deployment+Service
		if oldDS != nil {
			r.delete(ctx, oldDS, constants.DaemonSetKind)
		}
		if oldDepl == nil || deploymentNeedsUpdate(oldDepl.(*appsv1.Deployment), desiredGoflowKube, r.Namespace) {
			newDepl := buildDeployment(desiredGoflowKube, r.Namespace)
			r.createOrUpdate(ctx, oldDepl, newDepl, constants.DeploymentKind)
		}
		if oldSVC == nil || serviceNeedsUpdate(oldSVC.(*corev1.Service), desiredGoflowKube, r.Namespace) {
			newSVC := buildService(desiredGoflowKube, r.Namespace)
			r.createOrUpdate(ctx, oldSVC, newSVC, constants.DeploymentKind)
		}

		//Delete or Create / Update Autoscaler according to HPA option
		if oldASC != nil && desiredGoflowKube.HPA == nil {
			r.delete(ctx, oldASC, constants.AutoscalerKind)
		} else if desiredGoflowKube.HPA != nil {
			if oldASC == nil || autoScalerNeedsUpdate(oldASC.(*ascv1.HorizontalPodAutoscaler), desiredGoflowKube, r.Namespace) {
				newASC := buildAutoScaler(desiredGoflowKube, r.Namespace)
				r.createOrUpdate(ctx, oldASC, newASC, constants.AutoscalerKind)
			}
		}
	case constants.DaemonSetKind:
		// Kind changed: delete Deployment / Service / HPA and create DaemonSet
		if oldDepl != nil {
			r.delete(ctx, oldDepl, constants.DeploymentKind)
		}
		if oldSVC != nil {
			r.delete(ctx, oldSVC, constants.ServiceKind)
		}
		if oldASC != nil {
			r.delete(ctx, oldASC, constants.AutoscalerKind)
		}
		if oldDS != nil && !daemonSetNeedsUpdate(oldDS.(*appsv1.DaemonSet), desiredGoflowKube, r.Namespace) {
			return nil
		}
		newDS := buildDaemonSet(desiredGoflowKube, r.Namespace)
		r.createOrUpdate(ctx, oldDS, newDS, constants.DaemonSetKind)
	default:
		return fmt.Errorf("Could not reconcile collector, invalid kind: %s", desiredGoflowKube.Kind)
	}
	return nil
}

func (r *Reconciler) setupPermissions(ctx context.Context, ns string) {
	log := log.FromContext(ctx)
	log.Info("Setup permissions for " + constants.GoflowKubeName)
	rbacObjects := buildRBAC(ns)
	for _, rbacObj := range rbacObjects {
		err := r.SetControllerReference(rbacObj)
		if err != nil {
			log.Error(err, "Failed to set controller reference")
		}
		err = r.Create(ctx, rbacObj)
		if err != nil {
			log.Error(err, "Failed to setup permissions for "+constants.GoflowKubeName)
		}
	}
}

func (r *Reconciler) getObj(ctx context.Context, nsname types.NamespacedName, obj client.Object, kind string) (client.Object, error) {
	err := r.Get(ctx, nsname, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			log.FromContext(ctx).Error(err, "Failed to get "+constants.GoflowKubeName+" "+kind)
			return nil, err
		}
	}
	return obj, nil
}

func (r *Reconciler) createOrUpdate(ctx context.Context, old, new client.Object, kind string) {
	log := log.FromContext(ctx)
	err := r.SetControllerReference(new)
	if err != nil {
		log.Error(err, "Failed to set controller reference")
	}
	if old == nil {
		log.Info("Creating a new "+kind, "Namespace", new.GetNamespace(), "Name", new.GetName())
		err := r.Create(ctx, new)
		if err != nil {
			log.Error(err, "Failed to create new "+kind, "Namespace", new.GetNamespace(), "Name", new.GetName())
			return
		}
	} else {
		log.Info("Updating "+kind, "Namespace", new.GetNamespace(), "Name", new.GetName())
		err := r.Update(ctx, new)
		if err != nil {
			log.Error(err, "Failed to update "+kind, "Namespace", new.GetNamespace(), "Name", new.GetName())
			return
		}
	}
}

func (r *Reconciler) delete(ctx context.Context, old client.Object, kind string) {
	log := log.FromContext(ctx)
	log.Info("Deleting old "+kind, "Namespace", old.GetNamespace(), "Name", old.GetName())
	err := r.Delete(ctx, old)
	if err != nil {
		log.Error(err, "Failed to delete old "+kind, "Namespace", old.GetNamespace(), "Name", old.GetName())
	}
}

func daemonSetNeedsUpdate(ds *appsv1.DaemonSet, desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) bool {
	if ds.Namespace != ns {
		return true
	}
	return containerNeedsUpdate(&ds.Spec.Template.Spec, desired)
}

func deploymentNeedsUpdate(depl *appsv1.Deployment, desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) bool {
	if depl.Namespace != ns {
		return true
	}
	return containerNeedsUpdate(&depl.Spec.Template.Spec, desired) ||
		*depl.Spec.Replicas != desired.Replicas
}

func serviceNeedsUpdate(svc *corev1.Service, desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) bool {
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

func containerNeedsUpdate(podSpec *corev1.PodSpec, desired *flowsv1alpha1.FlowCollectorGoflowKube) bool {
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
	if len(container.Command) != 3 || container.Command[2] != buildMainCommand(desired) {
		return true
	}
	return false
}

func autoScalerNeedsUpdate(asc *ascv1.HorizontalPodAutoscaler, desired *flowsv1alpha1.FlowCollectorGoflowKube, ns string) bool {
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

func findContainer(podSpec *corev1.PodSpec) *corev1.Container {
	for _, ctnr := range podSpec.Containers {
		if ctnr.Name == constants.GoflowKubeName {
			return &ctnr
		}
	}
	return nil
}
