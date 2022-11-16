package flowlogspipeline

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/discover"
)

// Type alias
type flpSpec = flowsv1alpha1.FlowCollectorFLP

// FLPReconciler reconciles the current flowlogs-pipeline state with the desired configuration
type FLPReconciler struct {
	reconcilers []singleReconciler
}

const contextReconcilerName = "FLP kind"

type singleReconciler interface {
	context(ctx context.Context) context.Context
	initStaticResources(ctx context.Context) error
	prepareNamespaceChange(ctx context.Context) error
	reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error
}

func NewReconciler(ctx context.Context, cl reconcilers.ClientHelper, ns, prevNS, image string, permissionsVendor *discover.Permissions, availableAPIs *discover.AvailableAPIs) FLPReconciler {
	return FLPReconciler{
		reconcilers: []singleReconciler{
			newMonolithReconciler(ctx, cl, ns, prevNS, image, permissionsVendor, availableAPIs),
			newTransformerReconciler(ctx, cl, ns, prevNS, image, permissionsVendor, availableAPIs),
			newIngesterReconciler(ctx, cl, ns, prevNS, image, permissionsVendor, availableAPIs),
		},
	}
}

// InitStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *FLPReconciler) InitStaticResources(ctx context.Context) error {
	for _, sr := range r.reconcilers {
		if err := sr.initStaticResources(sr.context(ctx)); err != nil {
			return err
		}
	}
	return nil
}

// PrepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *FLPReconciler) PrepareNamespaceChange(ctx context.Context) error {
	for _, sr := range r.reconcilers {
		if err := sr.prepareNamespaceChange(sr.context(ctx)); err != nil {
			return err
		}
	}
	return nil
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

func (r *FLPReconciler) Reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	if err := validateDesired(&desired.Spec.Processor); err != nil {
		return err
	}
	for _, sr := range r.reconcilers {
		if err := sr.reconcile(sr.context(ctx), desired); err != nil {
			return err
		}
	}
	return nil
}

func daemonSetNeedsUpdate(ds *appsv1.DaemonSet, desired *flpSpec, image, configDigest string) bool {
	return containerNeedsUpdate(&ds.Spec.Template.Spec, desired, image) ||
		configChanged(&ds.Spec.Template, configDigest)
}

func configChanged(tmpl *corev1.PodTemplateSpec, configDigest string) bool {
	return tmpl.Annotations == nil || tmpl.Annotations[PodConfigurationDigest] != configDigest
}

func serviceNeedsUpdate(actual *corev1.Service, desired *corev1.Service) bool {
	return !equality.Semantic.DeepDerivative(desired.ObjectMeta, actual.ObjectMeta) ||
		!equality.Semantic.DeepDerivative(desired.Spec, actual.Spec)
}

func containerNeedsUpdate(podSpec *corev1.PodSpec, desired *flpSpec, image string) bool {
	// Note, we don't check for changed port / host port here, because that would change also the configmap,
	//	which also triggers pod update anyway
	container := reconcilers.FindContainer(podSpec, constants.FLPName)
	return container == nil ||
		image != container.Image ||
		desired.ImagePullPolicy != string(container.ImagePullPolicy) ||
		probesNeedUpdate(container, desired.EnableKubeProbes) ||
		!equality.Semantic.DeepDerivative(desired.Resources, container.Resources)
}

func probesNeedUpdate(container *corev1.Container, enabled bool) bool {
	if enabled {
		return container.LivenessProbe == nil || container.StartupProbe == nil
	}
	return container.LivenessProbe != nil || container.StartupProbe != nil
}
