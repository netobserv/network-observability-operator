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

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
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
	}
	if info.availableAPIs.HasPromRule() {
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

func (r *flpTransformerReconciler) reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	// Retrieve current owned objects
	err := r.nobjMngr.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Transformer only used with Kafka
	if !helper.UseKafka(&desired.Spec) {
		r.nobjMngr.TryDeleteAll(ctx)
		return nil
	}

	builder := newTransfoBuilder(r.nobjMngr.Namespace, r.image, &desired.Spec, r.useOpenShiftSCC, r.CertWatcher)
	newCM, configDigest, dbConfigMap, err := builder.configMap()
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
	if r.reconcilersCommonInfo.availableAPIs.HasConsoleConfig() {
		if err := r.reconcileDashboardConfig(ctx, dbConfigMap); err != nil {
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
	report := helper.NewChangeReport("FLP Deployment")
	defer report.LogIfNeeded(ctx)

	new := builder.deployment(configDigest)

	// Annotate pod with certificate reference so that it is reloaded if modified
	if err := r.CertWatcher.AnnotatePod(ctx, r.Client, &new.Spec.Template, lokiCerts, kafkaCerts); err != nil {
		return err
	}

	if !r.nobjMngr.Exists(r.owned.deployment) {
		if err := r.CreateOwned(ctx, new); err != nil {
			return err
		}
	} else if helper.DeploymentChanged(r.owned.deployment, new, constants.FLPName, helper.HPADisabled(&desiredFLP.KafkaConsumerAutoscaler), desiredFLP.KafkaConsumerReplicas, &report) {
		if err := r.UpdateOwned(ctx, r.owned.deployment, new); err != nil {
			return err
		}
	} else {
		// Deployment up to date, check if it's ready
		r.CheckDeploymentInProgress(r.owned.deployment)
	}

	// Delete or Create / Update Autoscaler according to HPA option
	if helper.HPADisabled(&desiredFLP.KafkaConsumerAutoscaler) {
		r.nobjMngr.TryDelete(ctx, r.owned.hpa)
	} else {
		newASC := builder.autoScaler()
		if !r.nobjMngr.Exists(r.owned.hpa) {
			if err := r.CreateOwned(ctx, newASC); err != nil {
				return err
			}
		} else if helper.AutoScalerChanged(r.owned.hpa, desiredFLP.KafkaConsumerAutoscaler, &report) {
			if err := r.UpdateOwned(ctx, r.owned.hpa, newASC); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *flpTransformerReconciler) reconcilePrometheusService(ctx context.Context, builder *transfoBuilder) error {
	report := helper.NewChangeReport("FLP prometheus service")
	defer report.LogIfNeeded(ctx)

	if !r.nobjMngr.Exists(r.owned.promService) {
		if err := r.CreateOwned(ctx, builder.newPromService()); err != nil {
			return err
		}
	} else {
		newSVC := builder.fromPromService(r.owned.promService)
		if helper.ServiceChanged(r.owned.promService, newSVC, &report) {
			if err := r.UpdateOwned(ctx, r.owned.promService, newSVC); err != nil {
				return err
			}
		}
	}
	if r.availableAPIs.HasSvcMonitor() {
		serviceMonitor := builder.generic.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.nobjMngr, &r.ClientHelper, r.owned.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	if r.availableAPIs.HasPromRule() {
		promRules := builder.generic.prometheusRule()
		if err := reconcilers.GenericReconcile(ctx, r.nobjMngr, &r.ClientHelper, r.owned.prometheusRule, promRules, &report, helper.PrometheusRuleChanged); err != nil {
			return err
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
