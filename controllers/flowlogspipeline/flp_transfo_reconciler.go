package flowlogspipeline

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

// flpTransformerReconciler reconciles the current flowlogs-pipeline-transformer state with the desired configuration
type flpTransformerReconciler struct {
	singleReconciler
	reconcilersCommonInfo
	owned transfoOwnedObjects
}

type transfoOwnedObjects struct {
	deployment     *appsv1.Deployment
	promService    *corev1.Service
	hpa            *ascv2.HorizontalPodAutoscaler
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	roleBinding    *rbacv1.ClusterRoleBinding
	serviceMonitor *monitoringv1.ServiceMonitor
	prometheusRule *monitoringv1.PrometheusRule
}

func newTransformerReconciler(info *reconcilersCommonInfo) *flpTransformerReconciler {
	name := name(ConfKafkaTransformer)
	owned := transfoOwnedObjects{
		deployment:     &appsv1.Deployment{},
		promService:    &corev1.Service{},
		hpa:            &ascv2.HorizontalPodAutoscaler{},
		serviceAccount: &corev1.ServiceAccount{},
		configMap:      &corev1.ConfigMap{},
		roleBinding:    &rbacv1.ClusterRoleBinding{},
		serviceMonitor: &monitoringv1.ServiceMonitor{},
		prometheusRule: &monitoringv1.PrometheusRule{},
	}
	info.nobjMngr.AddManagedObject(name, owned.deployment)
	info.nobjMngr.AddManagedObject(name, owned.hpa)
	info.nobjMngr.AddManagedObject(name, owned.serviceAccount)
	info.nobjMngr.AddManagedObject(promServiceName(ConfKafkaTransformer), owned.promService)
	info.nobjMngr.AddManagedObject(RoleBindingName(ConfKafkaTransformer), owned.roleBinding)
	info.nobjMngr.AddManagedObject(configMapName(ConfKafkaTransformer), owned.configMap)
	if info.availableAPIs.HasSvcMonitor() {
		info.nobjMngr.AddManagedObject(serviceMonitorName(ConfKafkaTransformer), owned.serviceMonitor)
		info.nobjMngr.AddManagedObject(prometheusRuleName(ConfKafkaTransformer), owned.prometheusRule)
	}

	return &flpTransformerReconciler{
		reconcilersCommonInfo: *info,
		owned:                 owned,
	}
}

func (r *flpTransformerReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithValues(contextReconcilerName, "transformer")
	return log.IntoContext(ctx, l)
}

// initStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *flpTransformerReconciler) initStaticResources(ctx context.Context) error {
	cr := buildClusterRoleTransformer()
	return r.ReconcileClusterRole(ctx, cr)
}

// prepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *flpTransformerReconciler) prepareNamespaceChange(ctx context.Context) error {
	// Switching namespace => delete everything in the previous namespace
	r.nobjMngr.CleanupPreviousNamespace(ctx)
	cr := buildClusterRoleTransformer()
	return r.ReconcileClusterRole(ctx, cr)
}

func (r *flpTransformerReconciler) reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollector) error {
	// Retrieve current owned objects
	err := r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Transformer only used with Kafka
	if !desired.Spec.UseKafka() {
		r.nobjMngr.TryDeleteAll(ctx)
		return nil
	}

	builder := newTransfoBuilder(r.nobjMngr.Namespace, r.image, &desired.Spec, r.useOpenShiftSCC, r.CertWatcher)
	newCM, configDigest, err := builder.configMap()
	if err != nil {
		return err
	}
	if !r.nobjMngr.Exists(r.owned.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newCM.Data, r.owned.configMap.Data) {
		if err := r.UpdateOwned(ctx, r.owned.configMap, newCM); err != nil {
			return err
		}
	}

	if err := r.reconcilePermissions(ctx, &builder); err != nil {
		return err
	}

	err = r.reconcilePrometheusService(ctx, &builder)
	if err != nil {
		return err
	}

	return r.reconcileDeployment(ctx, &desired.Spec.Processor, &builder, configDigest)
}

func (r *flpTransformerReconciler) reconcileDeployment(ctx context.Context, desiredFLP *flpSpec, builder *transfoBuilder, configDigest string) error {
	ns := r.nobjMngr.Namespace
	new := builder.deployment(configDigest)

	// Annotate pod with certificate reference so that it is reloaded if modified
	if err := r.CertWatcher.AnnotatePod(ctx, r.Client, &new.Spec.Template, lokiCerts, kafkaCerts); err != nil {
		return err
	}

	if !r.nobjMngr.Exists(r.owned.deployment) {
		if err := r.CreateOwned(ctx, new); err != nil {
			return err
		}
	} else if r.deploymentNeedsUpdate(r.owned.deployment, new, desiredFLP) {
		if err := r.UpdateOwned(ctx, r.owned.deployment, new); err != nil {
			return err
		}
	} else {
		// Deployment up to date, check if it's ready
		r.CheckDeploymentInProgress(r.owned.deployment)
	}

	// Delete or Create / Update Autoscaler according to HPA option
	if desiredFLP.KafkaConsumerAutoscaler.Disabled() {
		r.nobjMngr.TryDelete(ctx, r.owned.hpa)
	} else {
		newASC := builder.autoScaler()
		if !r.nobjMngr.Exists(r.owned.hpa) {
			if err := r.CreateOwned(ctx, newASC); err != nil {
				return err
			}
		} else if autoScalerNeedsUpdate(r.owned.hpa, desiredFLP.KafkaConsumerAutoscaler, ns) {
			if err := r.UpdateOwned(ctx, r.owned.hpa, newASC); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *flpTransformerReconciler) reconcilePrometheusService(ctx context.Context, builder *transfoBuilder) error {
	if !r.nobjMngr.Exists(r.owned.promService) {
		if err := r.CreateOwned(ctx, builder.newPromService()); err != nil {
			return err
		}
		if r.availableAPIs.HasSvcMonitor() {
			if err := r.CreateOwned(ctx, builder.generic.serviceMonitor()); err != nil {
				return err
			}
			if err := r.CreateOwned(ctx, builder.generic.prometheusRule()); err != nil {
				return err
			}
		}
		return nil
	}
	newSVC := builder.fromPromService(r.owned.promService)
	if helper.ServiceChanged(r.owned.promService, newSVC) {
		if err := r.UpdateOwned(ctx, r.owned.promService, newSVC); err != nil {
			return err
		}
	}
	if r.availableAPIs.HasSvcMonitor() {
		newMonitorSvc := builder.generic.serviceMonitor()
		if helper.ServiceMonitorChanged(r.owned.serviceMonitor, newMonitorSvc) {
			if err := r.UpdateOwned(ctx, r.owned.serviceMonitor, newMonitorSvc); err != nil {
				return err
			}
		}
		newPromRules := builder.generic.prometheusRule()
		if helper.PrometheusRuleChanged(r.owned.prometheusRule, newPromRules) {
			if err := r.UpdateOwned(ctx, r.owned.prometheusRule, newPromRules); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *flpTransformerReconciler) reconcilePermissions(ctx context.Context, builder *transfoBuilder) error {
	if !r.nobjMngr.Exists(r.owned.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now

	desired := builder.clusterRoleBinding()
	if err := r.ReconcileClusterRoleBinding(ctx, desired); err != nil {
		return err
	}
	return nil
}

func (r *flpTransformerReconciler) deploymentNeedsUpdate(old, new *appsv1.Deployment, desired *flpSpec) bool {
	return helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.FLPName) ||
		(desired.KafkaConsumerAutoscaler.Disabled() && *old.Spec.Replicas != desired.KafkaConsumerReplicas)
}

func autoScalerNeedsUpdate(asc *ascv2.HorizontalPodAutoscaler, desired flowsv1alpha1.FlowCollectorHPA, ns string) bool {
	if asc.Namespace != ns {
		return true
	}
	differentPointerValues := func(a, b *int32) bool {
		return (a == nil && b != nil) || (a != nil && b == nil) || (a != nil && *a != *b)
	}
	if asc.Spec.MaxReplicas != desired.MaxReplicas ||
		differentPointerValues(asc.Spec.MinReplicas, desired.MinReplicas) {
		return true
	}
	if !equality.Semantic.DeepDerivative(desired.Metrics, asc.Spec.Metrics) {
		return true
	}
	return false
}
