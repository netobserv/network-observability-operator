package flowlogspipeline

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/discover"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

// Type alias
type flpSpec = flowsv1alpha1.FlowCollectorFLP

// FLPReconciler reconciles the current flowlogs-pipeline state with the desired configuration
type FLPReconciler struct {
	reconcilers.ClientHelper
	nobjMngr          *reconcilers.NamespacedObjectManager
	owned             ownedObjects
	singleReconcilers []singleDeploymentReconciler
	useOpenShiftSCC   bool
}

type singleDeploymentReconciler struct {
	reconcilers.ClientHelper
	nobjMngr        *reconcilers.NamespacedObjectManager
	owned           ownedObjects
	confKind        string
	useOpenShiftSCC bool
}

type ownedObjects struct {
	deployment             *appsv1.Deployment
	daemonSet              *appsv1.DaemonSet
	service                *corev1.Service
	promService            *corev1.Service
	hpa                    *ascv2.HorizontalPodAutoscaler
	serviceAccount         *corev1.ServiceAccount
	configMap              *corev1.ConfigMap
	roleBindingIngester    *rbacv1.ClusterRoleBinding
	roleBindingTransformer *rbacv1.ClusterRoleBinding
}

func NewReconciler(ctx context.Context, cl reconcilers.ClientHelper, ns, prevNS string, permissionsVendor *discover.Permissions) FLPReconciler {
	owned := ownedObjects{
		service: &corev1.Service{},
	}
	nobjMngr := reconcilers.NewNamespacedObjectManager(cl, ns, prevNS)
	nobjMngr.AddManagedObject(constants.FLPName, owned.service)

	openshift := permissionsVendor.Vendor(ctx) == discover.VendorOpenShift

	flpReconciler := FLPReconciler{ClientHelper: cl, nobjMngr: nobjMngr, owned: owned, useOpenShiftSCC: openshift}
	flpReconciler.singleReconcilers = append(flpReconciler.singleReconcilers, newSingleReconciler(cl, ns, prevNS, ConfSingle, openshift))
	flpReconciler.singleReconcilers = append(flpReconciler.singleReconcilers, newSingleReconciler(cl, ns, prevNS, ConfKafkaIngester, openshift))
	flpReconciler.singleReconcilers = append(flpReconciler.singleReconcilers, newSingleReconciler(cl, ns, prevNS, ConfKafkaTransformer, openshift))
	return flpReconciler
}

func newSingleReconciler(cl reconcilers.ClientHelper, ns, prevNS, confKind string, useOpenShiftSCC bool) singleDeploymentReconciler {
	owned := ownedObjects{
		deployment:             &appsv1.Deployment{},
		daemonSet:              &appsv1.DaemonSet{},
		service:                &corev1.Service{},
		promService:            &corev1.Service{},
		hpa:                    &ascv2.HorizontalPodAutoscaler{},
		serviceAccount:         &corev1.ServiceAccount{},
		configMap:              &corev1.ConfigMap{},
		roleBindingIngester:    &rbacv1.ClusterRoleBinding{},
		roleBindingTransformer: &rbacv1.ClusterRoleBinding{},
	}
	nobjMngr := reconcilers.NewNamespacedObjectManager(cl, ns, prevNS)
	nobjMngr.AddManagedObject(constants.FLPName+FlpConfSuffix[confKind], owned.deployment)
	nobjMngr.AddManagedObject(constants.FLPName+FlpConfSuffix[confKind], owned.daemonSet)
	nobjMngr.AddManagedObject(constants.FLPName+FlpConfSuffix[confKind], owned.hpa)
	nobjMngr.AddManagedObject(constants.FLPName+FlpConfSuffix[confKind], owned.serviceAccount)
	nobjMngr.AddManagedObject(constants.FLPName+FlpConfSuffix[confKind]+PromServiceSuffix, owned.promService)
	nobjMngr.AddManagedObject(configMapName+FlpConfSuffix[confKind], owned.configMap)

	if confKind == ConfSingle || confKind == ConfKafkaIngester {
		nobjMngr.AddManagedObject(constants.FLPName+FlpConfSuffix[confKind]+FlpConfSuffix[ConfKafkaIngester]+"role", owned.roleBindingIngester)
	}
	if confKind == ConfSingle || confKind == ConfKafkaTransformer {
		nobjMngr.AddManagedObject(constants.FLPName+FlpConfSuffix[confKind]+FlpConfSuffix[ConfKafkaTransformer]+"role", owned.roleBindingIngester)
	}
	return singleDeploymentReconciler{ClientHelper: cl, nobjMngr: nobjMngr, owned: owned, confKind: confKind, useOpenShiftSCC: useOpenShiftSCC}
}

// InitStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *FLPReconciler) InitStaticResources(ctx context.Context) error {
	return r.reconcilePermissions(ctx)
}

// PrepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *FLPReconciler) PrepareNamespaceChange(ctx context.Context) error {
	// Switching namespace => delete everything in the previous namespace
	for i := 0; i < len(r.singleReconcilers); i++ {
		r.singleReconcilers[i].nobjMngr.CleanupPreviousNamespace(ctx)
	}
	r.nobjMngr.CleanupPreviousNamespace(ctx)
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

func (r *FLPReconciler) GetServiceName(kafka *flowsv1alpha1.FlowCollectorKafka) string {
	return constants.FLPName
}

func (r *FLPReconciler) Reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	for i := 0; i < len(r.singleReconcilers); i++ {
		err := r.singleReconcilers[i].Reconcile(ctx, desired)
		if err != nil {
			return err
		}
	}
	return nil
}

// Check if a configKind should be deployed
func checkDeployNeeded(fc *flowsv1alpha1.FlowCollectorSpec, confKind string) (bool, error) {
	switch confKind {
	case ConfSingle:
		return !fc.Kafka.Enable, nil
	case ConfKafkaTransformer:
		return fc.Kafka.Enable, nil
	case ConfKafkaIngester:
		// disabled if ebpf-agent is enabled, as it sends the flows directly to the transformer
		return fc.Kafka.Enable && fc.Agent == flowsv1alpha1.AgentIPFIX, nil
	default:
		return false, fmt.Errorf("unknown flowlogs-pipelines config kind")
	}
}

// Reconcile is the reconciler entry point to reconcile the current flowlogs-pipeline state with the desired configuration
func (r *singleDeploymentReconciler) Reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	desiredFLP := &desired.Spec.FlowlogsPipeline
	desiredLoki := &desired.Spec.Loki
	desiredKafka := &desired.Spec.Kafka
	err := validateDesired(desiredFLP)
	if err != nil {
		return err
	}

	shouldDeploy, err := checkDeployNeeded(&desired.Spec, r.confKind)
	if err != nil {
		return err
	}
	if !shouldDeploy {
		r.nobjMngr.CleanupCurrentNamespace(ctx)
		return nil
	}

	// Retrieve current owned objects
	err = r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}

	builder := newBuilder(r.nobjMngr.Namespace, desired.Spec.Agent, desiredFLP, desiredLoki, desiredKafka, r.confKind, r.useOpenShiftSCC)
	newCM, configDigest, err := builder.configMap()
	if err != nil {
		return err
	}
	if !r.nobjMngr.Exists(r.owned.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return err
		}
	} else if !reflect.DeepEqual(newCM.Data, r.owned.configMap.Data) {
		if err := r.UpdateOwned(ctx, r.owned.configMap, newCM); err != nil {
			return err
		}
	}

	if err := r.reconcileServiceAccount(ctx, &builder); err != nil {
		return err
	}

	err = r.reconcilePrometheusService(ctx, &builder)
	if err != nil {
		return err
	}

	if r.confKind == ConfKafkaTransformer {
		return r.reconcileAsDeployment(ctx, desiredFLP, &builder, configDigest)
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

func (r *singleDeploymentReconciler) reconcileAsDeployment(ctx context.Context, desiredFLP *flpSpec, builder *builder, configDigest string) error {
	// Kind may have changed: try delete DaemonSet and create Deployment+Service
	ns := r.nobjMngr.Namespace
	r.nobjMngr.TryDelete(ctx, r.owned.daemonSet)

	if !r.nobjMngr.Exists(r.owned.deployment) {
		if err := r.CreateOwned(ctx, builder.deployment(configDigest)); err != nil {
			return err
		}
	} else if deploymentNeedsUpdate(r.owned.deployment, desiredFLP, configDigest, constants.FLPName+FlpConfSuffix[r.confKind]) {
		if err := r.UpdateOwned(ctx, r.owned.deployment, builder.deployment(configDigest)); err != nil {
			return err
		}
	} else {
		// Deployment up to date, check if it's ready
		r.CheckDeploymentInProgress(r.owned.deployment)
	}
	if r.confKind != ConfKafkaTransformer {
		if err := r.reconcileService(ctx, desiredFLP, builder); err != nil {
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

func (r *singleDeploymentReconciler) reconcileService(ctx context.Context, desiredFLP *flpSpec, builder *builder) error {
	actual := corev1.Service{}
	if err := r.Client.Get(ctx,
		types.NamespacedName{Name: constants.FLPName, Namespace: r.nobjMngr.Namespace},
		&actual,
	); err != nil {
		if errors.IsNotFound(err) {
			return r.CreateOwned(ctx, builder.service(nil))
		}
		return fmt.Errorf("can't reconcile %s Serviceg: %w", constants.FLPName, err)
	}
	newSVC := builder.service(&actual)
	if serviceNeedsUpdate(&actual, newSVC) {
		if err := r.UpdateOwned(ctx, &actual, newSVC); err != nil {
			return err
		}
	}
	return nil
}

func (r *singleDeploymentReconciler) reconcilePrometheusService(ctx context.Context, builder *builder) error {
	if !r.nobjMngr.Exists(r.owned.promService) {
		return r.CreateOwned(ctx, builder.promService(nil))
	}
	newSVC := builder.promService(r.owned.promService)
	if serviceNeedsUpdate(r.owned.promService, newSVC) {
		if err := r.UpdateOwned(ctx, r.owned.promService, newSVC); err != nil {
			return err
		}
	}
	return nil
}

func (r *singleDeploymentReconciler) reconcileAsDaemonSet(ctx context.Context, desiredFLP *flpSpec, builder *builder, configDigest string) error {
	// Kind may have changed: try delete Deployment / Service / HPA and create DaemonSet
	r.nobjMngr.TryDelete(ctx, r.owned.deployment)
	r.nobjMngr.TryDelete(ctx, r.owned.hpa)
	if err := r.Client.Delete(ctx, builder.service(nil)); !errors.IsNotFound(err) {
		return err
	}
	if !r.nobjMngr.Exists(r.owned.daemonSet) {
		return r.CreateOwned(ctx, builder.daemonSet(configDigest))
	} else if daemonSetNeedsUpdate(r.owned.daemonSet, desiredFLP, configDigest, constants.FLPName+FlpConfSuffix[r.confKind]) {
		return r.UpdateOwned(ctx, r.owned.daemonSet, builder.daemonSet(configDigest))
	} else {
		// DaemonSet up to date, check if it's ready
		r.CheckDaemonSetInProgress(r.owned.daemonSet)
	}
	return nil
}

func (r *singleDeploymentReconciler) reconcileServiceAccount(ctx context.Context, builder *builder) error {
	if !r.nobjMngr.Exists(r.owned.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now
	if r.confKind == ConfKafkaIngester || r.confKind == ConfSingle {
		if err := r.reconcileClusterRoleBinding(ctx, builder, ConfKafkaIngester); err != nil {
			return err
		}
	}
	if r.confKind == ConfKafkaTransformer || r.confKind == ConfSingle {
		if err := r.reconcileClusterRoleBinding(ctx, builder, ConfKafkaTransformer); err != nil {
			return err
		}
	}
	return nil
}

func (r *singleDeploymentReconciler) reconcileClusterRoleBinding(ctx context.Context, builder *builder, roleKind string) error {
	desired := builder.clusterRoleBinding(roleKind)
	actual := rbacv1.ClusterRoleBinding{}
	if err := r.Client.Get(ctx,
		types.NamespacedName{Name: desired.ObjectMeta.Name},
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
		if actual.RoleRef != desired.RoleRef {
			//Roleref cannot be updated deleting and creating a new rolebinding
			r.nobjMngr.TryDelete(ctx, &actual)
			return r.CreateOwned(ctx, desired)
		}
		// cluster role binding already reconciled. Exiting
		return nil
	}
	return r.UpdateOwned(ctx, &actual, desired)
}

func (r *FLPReconciler) reconcilePermissions(ctx context.Context) error {
	// Cluster role is only installed once
	if err := r.reconcileClusterRole(ctx, buildClusterRoleIngester(r.useOpenShiftSCC), constants.FLPName+FlpConfSuffix[ConfKafkaIngester]); err != nil {
		return err
	}
	if err := r.reconcileClusterRole(ctx, buildClusterRoleTransformer(r.useOpenShiftSCC), constants.FLPName+FlpConfSuffix[ConfKafkaTransformer]); err != nil {
		return err
	}
	return nil
}

func (r *FLPReconciler) reconcileClusterRole(ctx context.Context, desired *rbacv1.ClusterRole, name string) error {
	actual := rbacv1.ClusterRole{}
	if err := r.Client.Get(ctx,
		types.NamespacedName{Name: name},
		&actual,
	); err != nil {
		if errors.IsNotFound(err) {
			return r.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile %s ClusterRole: %w", name, err)
	}

	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		reflect.DeepEqual(actual.Rules, desired.Rules) {
		// cluster role already reconciled. Exiting
		return nil
	}

	return r.UpdateOwned(ctx, &actual, desired)
}

func daemonSetNeedsUpdate(ds *appsv1.DaemonSet, desired *flpSpec, configDigest string, name string) bool {
	return containerNeedsUpdate(&ds.Spec.Template.Spec, desired, true, name) ||
		configChanged(&ds.Spec.Template, configDigest)
}

func deploymentNeedsUpdate(depl *appsv1.Deployment, desired *flpSpec, configDigest string, name string) bool {
	return containerNeedsUpdate(&depl.Spec.Template.Spec, desired, false, name) ||
		configChanged(&depl.Spec.Template, configDigest) ||
		(desired.HPA == nil && *depl.Spec.Replicas != desired.Replicas)
}

func configChanged(tmpl *corev1.PodTemplateSpec, configDigest string) bool {
	return tmpl.Annotations == nil || tmpl.Annotations[PodConfigurationDigest] != configDigest
}

func serviceNeedsUpdate(actual *corev1.Service, desired *corev1.Service) bool {
	return !reflect.DeepEqual(actual.ObjectMeta, desired.ObjectMeta) ||
		!reflect.DeepEqual(actual.Spec, desired.Spec)
}

func containerNeedsUpdate(podSpec *corev1.PodSpec, desired *flpSpec, expectHostPort bool, name string) bool {
	// Note, we don't check for changed port / host port here, because that would change also the configmap,
	//	which also triggers pod update anyway
	container := reconcilers.FindContainer(podSpec, name)
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
