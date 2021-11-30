package goflowkube

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	ascv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type ownedObjects struct {
	deployment *appsv1.Deployment
	daemonSet  *appsv1.DaemonSet
	service    *corev1.Service
	hpa        *ascv1.HorizontalPodAutoscaler
	configMap  *corev1.ConfigMap
}

// Reconcile is the reconciler entry point to reconcile the current goflow-kube state with the desired configuration
func (r *Reconciler) Reconcile(ctx context.Context,
	desiredGoflowKube *flowsv1alpha1.FlowCollectorGoflowKube,
	desiredLoki *flowsv1alpha1.FlowCollectorLoki, previousNamespace string) error {

	old := &ownedObjects{}
	var err error
	if previousNamespace == "" {
		r.createPermissions(ctx)
	} else if previousNamespace != r.Namespace {
		// Switching namespace => delete everything in the previous namespace
		r.updatePermissions(ctx, previousNamespace)
		r.deleteOwnedObjects(ctx, previousNamespace)
	} else {
		// Retrieve current owned objects
		old, err = r.getOwnedObjects(ctx)
		if err != nil {
			return err
		}
	}

	newCM, configDigest := buildConfigMap(desiredGoflowKube, desiredLoki, r.Namespace)
	if old.configMap == nil || !reflect.DeepEqual(newCM, old.configMap.Data) {
		r.createOrUpdate(ctx, old.configMap, newCM, constants.ConfigMapKind)
	}

	switch desiredGoflowKube.Kind {
	case constants.DeploymentKind:
		r.reconcileAsDeployment(ctx, old, desiredGoflowKube, configDigest)
	case constants.DaemonSetKind:
		r.reconcileAsDaemonSet(ctx, old, desiredGoflowKube, configDigest)
	default:
		return fmt.Errorf("could not reconcile collector, invalid kind: %s", desiredGoflowKube.Kind)
	}
	return nil
}

// getOwnedObjects retrieves current / old objects before the reconciliation run
func (r *Reconciler) getOwnedObjects(ctx context.Context) (*ownedObjects, error) {
	nsname := types.NamespacedName{Name: constants.GoflowKubeName, Namespace: r.Namespace}
	objs := ownedObjects{}
	if oldDepl, err := r.getObj(ctx, nsname, &appsv1.Deployment{}, constants.DeploymentKind); err != nil {
		return nil, err
	} else if oldDepl != nil {
		objs.deployment = oldDepl.(*appsv1.Deployment)
	}
	if oldDS, err := r.getObj(ctx, nsname, &appsv1.DaemonSet{}, constants.DaemonSetKind); err != nil {
		return nil, err
	} else if oldDS != nil {
		objs.daemonSet = oldDS.(*appsv1.DaemonSet)
	}
	if oldSVC, err := r.getObj(ctx, nsname, &corev1.Service{}, constants.ServiceKind); err != nil {
		return nil, err
	} else if oldSVC != nil {
		objs.service = oldSVC.(*corev1.Service)
	}
	if oldASC, err := r.getObj(ctx, nsname, &ascv1.HorizontalPodAutoscaler{}, constants.AutoscalerKind); err != nil {
		return nil, err
	} else if oldASC != nil {
		objs.hpa = oldASC.(*ascv1.HorizontalPodAutoscaler)
	}
	if oldCM, err := r.getObj(ctx, types.NamespacedName{Name: configMapName, Namespace: r.Namespace}, &corev1.ConfigMap{}, constants.ConfigMapKind); err != nil {
		return nil, err
	} else if oldCM != nil {
		objs.configMap = oldCM.(*corev1.ConfigMap)
	}
	return &objs, nil
}

func (r *Reconciler) deleteOwnedObjects(ctx context.Context, ns string) {
	meta := metav1.ObjectMeta{
		Name:      constants.GoflowKubeName,
		Namespace: ns,
	}
	r.delete(ctx, &appsv1.Deployment{ObjectMeta: meta}, constants.DeploymentKind)
	r.delete(ctx, &appsv1.DaemonSet{ObjectMeta: meta}, constants.DaemonSetKind)
	r.delete(ctx, &corev1.Service{ObjectMeta: meta}, constants.ServiceKind)
	r.delete(ctx, &ascv1.HorizontalPodAutoscaler{ObjectMeta: meta}, constants.AutoscalerKind)
	r.delete(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
		Name:      configMapName,
		Namespace: ns,
	}}, constants.ConfigMapKind)
}

func (r *Reconciler) reconcileAsDeployment(ctx context.Context, old *ownedObjects,
	desiredGoflowKube *flowsv1alpha1.FlowCollectorGoflowKube, configDigest string) {
	// Kind changed: delete DaemonSet and create Deployment+Service
	if old.daemonSet != nil {
		r.delete(ctx, old.daemonSet, constants.DaemonSetKind)
	}
	if old.deployment == nil ||
		deploymentNeedsUpdate(old.deployment, desiredGoflowKube, r.Namespace, configDigest) {
		newDepl := buildDeployment(desiredGoflowKube, r.Namespace, configDigest)
		r.createOrUpdate(ctx, old.deployment, newDepl, constants.DeploymentKind)
	}
	if old.service == nil || serviceNeedsUpdate(old.service, desiredGoflowKube, r.Namespace) {
		newSVC := buildService(desiredGoflowKube, r.Namespace)
		r.createOrUpdate(ctx, old.service, newSVC, constants.DeploymentKind)
	}

	// Delete or Create / Update Autoscaler according to HPA option
	if old.hpa != nil && desiredGoflowKube.HPA == nil {
		r.delete(ctx, old.hpa, constants.AutoscalerKind)
	} else if desiredGoflowKube.HPA != nil {
		if old.hpa == nil || autoScalerNeedsUpdate(old.hpa, desiredGoflowKube, r.Namespace) {
			newASC := buildAutoScaler(desiredGoflowKube, r.Namespace)
			r.createOrUpdate(ctx, old.hpa, newASC, constants.AutoscalerKind)
		}
	}
}

func (r *Reconciler) reconcileAsDaemonSet(ctx context.Context, old *ownedObjects,
	desiredGoflowKube *flowsv1alpha1.FlowCollectorGoflowKube, configDigest string) {
	// Kind changed: delete Deployment / Service / HPA and create DaemonSet
	if old.deployment != nil {
		r.delete(ctx, old.deployment, constants.DeploymentKind)
	}
	if old.service != nil {
		r.delete(ctx, old.service, constants.ServiceKind)
	}
	if old.hpa != nil {
		r.delete(ctx, old.hpa, constants.AutoscalerKind)
	}
	if old.daemonSet == nil ||
		daemonSetNeedsUpdate(old.daemonSet, desiredGoflowKube, r.Namespace, configDigest) {
		newDS := buildDaemonSet(desiredGoflowKube, r.Namespace, configDigest)
		r.createOrUpdate(ctx, old.daemonSet, newDS, constants.DaemonSetKind)
	}
}

func (r *Reconciler) createPermissions(ctx context.Context) {
	r.createOrUpdate(ctx, nil, buildClusterRole(), "ClusterRole")
	r.createOrUpdate(ctx, nil, buildServiceAccount(r.Namespace), "ServiceAccount")
	r.createOrUpdate(ctx, nil, buildClusterRoleBinding(r.Namespace), "ClusterRoleBinding")
}

func (r *Reconciler) updatePermissions(ctx context.Context, previousNamespace string) {
	// Replace service account (delete+create), update binding
	oldSA := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.GoflowKubeName,
			Namespace: previousNamespace,
		},
	}
	crbID := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.GoflowKubeName,
		},
	}
	r.delete(ctx, &oldSA, "ServiceAccount")
	r.createOrUpdate(ctx, nil, buildServiceAccount(r.Namespace), "ServiceAccount")
	r.createOrUpdate(ctx, &crbID, buildClusterRoleBinding(r.Namespace), "ClusterRoleBinding")
}

func (r *Reconciler) getObj(ctx context.Context, nsname types.NamespacedName, obj client.Object, kind string) (client.Object, error) {
	err := r.Get(ctx, nsname, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		log.FromContext(ctx).Error(err, "Failed to get "+constants.GoflowKubeName+" "+kind)
		return nil, err
	}
	return obj, nil
}

func (r *Reconciler) createOrUpdate(ctx context.Context, old, new client.Object, kind string) {
	log := log.FromContext(ctx)
	err := r.SetControllerReference(new)
	if err != nil {
		log.Error(err, "Failed to set controller reference")
	}
	// "old" may be nil but have a type assigned so "old == nil" won't work
	if old == nil || reflect.ValueOf(old).IsNil() {
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

func daemonSetNeedsUpdate(ds *appsv1.DaemonSet, desired *flowsv1alpha1.FlowCollectorGoflowKube,
	ns, configDigest string) bool {
	if ds.Namespace != ns {
		return true
	}
	return containerNeedsUpdate(&ds.Spec.Template.Spec, desired) ||
		configChanged(&ds.Spec.Template, configDigest)
}

func deploymentNeedsUpdate(depl *appsv1.Deployment, desired *flowsv1alpha1.FlowCollectorGoflowKube,
	ns, configDigest string) bool {
	if depl.Namespace != ns {
		return true
	}
	return containerNeedsUpdate(&depl.Spec.Template.Spec, desired) ||
		configChanged(&depl.Spec.Template, configDigest) ||
		*depl.Spec.Replicas != desired.Replicas
}

func configChanged(tmpl *corev1.PodTemplateSpec, configDigest string) bool {
	return tmpl.Annotations == nil || tmpl.Annotations[PodConfigurationDigest] != configDigest
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
	for i := range podSpec.Containers {
		if podSpec.Containers[i].Name == constants.GoflowKubeName {
			return &podSpec.Containers[i]
		}
	}
	return nil
}
