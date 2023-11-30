package flp

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

// flpTransformerReconciler reconciles the current flowlogs-pipeline-transformer state with the desired configuration
type flpTransformerReconciler struct {
	*reconcilers.Instance
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

func newTransformerReconciler(cmn *reconcilers.Instance) *flpTransformerReconciler {
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
	cmn.Managed.AddManagedObject(name, owned.deployment)
	cmn.Managed.AddManagedObject(name, owned.hpa)
	cmn.Managed.AddManagedObject(name, owned.serviceAccount)
	cmn.Managed.AddManagedObject(promServiceName(ConfKafkaTransformer), owned.promService)
	cmn.Managed.AddManagedObject(RoleBindingName(ConfKafkaTransformer), owned.roleBinding)
	cmn.Managed.AddManagedObject(configMapName(ConfKafkaTransformer), owned.configMap)
	if cmn.AvailableAPIs.HasSvcMonitor() {
		cmn.Managed.AddManagedObject(serviceMonitorName(ConfKafkaTransformer), owned.serviceMonitor)
	}
	if cmn.AvailableAPIs.HasPromRule() {
		cmn.Managed.AddManagedObject(prometheusRuleName(ConfKafkaTransformer), owned.prometheusRule)
	}

	return &flpTransformerReconciler{
		Instance: cmn,
		owned:    owned,
	}
}

func (r *flpTransformerReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithValues(contextReconcilerName, "transformer")
	return log.IntoContext(ctx, l)
}

// cleanupNamespace cleans up old namespace
func (r *flpTransformerReconciler) cleanupNamespace(ctx context.Context) {
	r.Managed.CleanupPreviousNamespace(ctx)
}

func (r *flpTransformerReconciler) reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	// Transformer only used with Kafka
	if !helper.UseKafka(&desired.Spec) {
		r.Managed.TryDeleteAll(ctx)
		return nil
	}

	builder, err := newTransfoBuilder(r.Instance, &desired.Spec)
	if err != nil {
		return err
	}
	newCM, configDigest, err := builder.configMap()
	if err != nil {
		return err
	}
	annotations := map[string]string{
		constants.PodConfigurationDigest: configDigest,
	}
	if !r.Managed.Exists(r.owned.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newCM.Data, r.owned.configMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.owned.configMap, newCM); err != nil {
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

	// Watch for Loki certificate if necessary; we'll ignore in that case the returned digest, as we don't need to restart pods on cert rotation
	// because certificate is always reloaded from file
	if _, err = r.Watcher.ProcessCACert(ctx, r.Client, &r.Loki.TLS, r.Namespace); err != nil {
		return err
	}

	// Watch for Kafka certificate if necessary; need to restart pods in case of cert rotation
	if err = annotateKafkaCerts(ctx, r.Common, &desired.Spec.Kafka, "kafka", annotations); err != nil {
		return err
	}
	// Same for Kafka exporters
	if err = annotateKafkaExporterCerts(ctx, r.Common, desired.Spec.Exporters, annotations); err != nil {
		return err
	}
	// Watch for monitoring caCert
	if err = reconcileMonitoringCerts(ctx, r.Common, &desired.Spec.Processor.Metrics.Server.TLS, r.Namespace); err != nil {
		return err
	}

	if err = r.reconcileDeployment(ctx, &desired.Spec.Processor, &builder, annotations); err != nil {
		return err
	}

	return r.reconcileHPA(ctx, &desired.Spec.Processor, &builder)
}

func (r *flpTransformerReconciler) reconcileDeployment(ctx context.Context, desiredFLP *flpSpec, builder *transfoBuilder, annotations map[string]string) error {
	report := helper.NewChangeReport("FLP Deployment")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileDeployment(
		ctx,
		r.Instance,
		r.owned.deployment,
		builder.deployment(annotations),
		constants.FLPName,
		helper.PtrInt32(desiredFLP.KafkaConsumerReplicas),
		&desiredFLP.KafkaConsumerAutoscaler,
		&report,
	)
}

func (r *flpTransformerReconciler) reconcileHPA(ctx context.Context, desiredFLP *flpSpec, builder *transfoBuilder) error {
	report := helper.NewChangeReport("FLP autoscaler")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileHPA(
		ctx,
		r.Instance,
		r.owned.hpa,
		builder.autoScaler(),
		&desiredFLP.KafkaConsumerAutoscaler,
		&report,
	)
}

func (r *flpTransformerReconciler) reconcilePrometheusService(ctx context.Context, builder *transfoBuilder) error {
	report := helper.NewChangeReport("FLP prometheus service")
	defer report.LogIfNeeded(ctx)

	if err := r.ReconcileService(ctx, r.owned.promService, builder.promService(), &report); err != nil {
		return err
	}
	if r.AvailableAPIs.HasSvcMonitor() {
		serviceMonitor := builder.generic.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.owned.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	if r.AvailableAPIs.HasPromRule() {
		promRules := builder.generic.prometheusRule()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.owned.prometheusRule, promRules, &report, helper.PrometheusRuleChanged); err != nil {
			return err
		}
	}
	return nil
}

func (r *flpTransformerReconciler) reconcilePermissions(ctx context.Context, builder *transfoBuilder) error {
	if !r.Managed.Exists(r.owned.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now

	cr := BuildClusterRoleTransformer()
	if err := r.ReconcileClusterRole(ctx, cr); err != nil {
		return err
	}

	desired := builder.clusterRoleBinding()
	if err := r.ReconcileClusterRoleBinding(ctx, desired); err != nil {
		return err
	}

	return reconcileLokiRoles(ctx, r.Common, &builder.generic)
}
