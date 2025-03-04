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

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/resources"
)

type transformerReconciler struct {
	*reconcilers.Instance
	deployment       *appsv1.Deployment
	promService      *corev1.Service
	hpa              *ascv2.HorizontalPodAutoscaler
	serviceAccount   *corev1.ServiceAccount
	staticConfigMap  *corev1.ConfigMap
	dynamicConfigMap *corev1.ConfigMap
	roleBinding      *rbacv1.ClusterRoleBinding
	serviceMonitor   *monitoringv1.ServiceMonitor
	prometheusRule   *monitoringv1.PrometheusRule
}

func newTransformerReconciler(cmn *reconcilers.Instance) *transformerReconciler {
	name := name(ConfKafkaTransformer)
	rec := transformerReconciler{
		Instance:         cmn,
		deployment:       cmn.Managed.NewDeployment(name),
		promService:      cmn.Managed.NewService(promServiceName(ConfKafkaTransformer)),
		hpa:              cmn.Managed.NewHPA(name),
		serviceAccount:   cmn.Managed.NewServiceAccount(name),
		staticConfigMap:  cmn.Managed.NewConfigMap(staticConfigMapName(ConfKafkaTransformer)),
		dynamicConfigMap: cmn.Managed.NewConfigMap(dynamicConfigMapName(ConfKafkaTransformer)),
		roleBinding:      cmn.Managed.NewCRB(RoleBindingName(ConfKafkaTransformer)),
	}
	if cmn.ClusterInfo.HasSvcMonitor() {
		rec.serviceMonitor = cmn.Managed.NewServiceMonitor(serviceMonitorName(ConfKafkaTransformer))
	}
	if cmn.ClusterInfo.HasPromRule() {
		rec.prometheusRule = cmn.Managed.NewPrometheusRule(prometheusRuleName(ConfKafkaTransformer))
	}
	return &rec
}

func (r *transformerReconciler) context(ctx context.Context) context.Context {
	l := log.FromContext(ctx).WithName("transformer")
	return log.IntoContext(ctx, l)
}

// cleanupNamespace cleans up old namespace
func (r *transformerReconciler) cleanupNamespace(ctx context.Context) {
	r.Managed.CleanupPreviousNamespace(ctx)
}

func (r *transformerReconciler) getStatus() *status.Instance {
	return &r.Status
}

func (r *transformerReconciler) reconcile(ctx context.Context, desired *flowslatest.FlowCollector, flowMetrics *metricslatest.FlowMetricList, detectedSubnets []flowslatest.SubnetLabel) error {
	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if !helper.UseKafka(&desired.Spec) {
		r.Status.SetUnused("Transformer only used with Kafka")
		r.Managed.TryDeleteAll(ctx)
		return nil
	}

	r.Status.SetReady() // will be overidden if necessary, as error or pending

	builder, err := newTransfoBuilder(r.Instance, &desired.Spec, flowMetrics, detectedSubnets)
	if err != nil {
		return err
	}
	newSCM, configDigest, err := builder.staticConfigMap()
	if err != nil {
		return err
	}
	annotations := map[string]string{
		constants.PodConfigurationDigest: configDigest,
	}
	if !r.Managed.Exists(r.staticConfigMap) {
		if err := r.CreateOwned(ctx, newSCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newSCM.Data, r.staticConfigMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.staticConfigMap, newSCM); err != nil {
			return err
		}
	}

	if err := r.reconcileDynamicConfigMap(ctx, &builder); err != nil {
		return err
	}

	if err := r.reconcilePermissions(ctx, &builder); err != nil {
		return err
	}

	err = r.reconcilePrometheusService(ctx, &builder)
	if err != nil {
		return err
	}

	if helper.UseLoki(&desired.Spec) {
		// Watch for Loki certificate if necessary; we'll ignore in that case the returned digest, as we don't need to restart pods on cert rotation
		// because certificate is always reloaded from file
		if _, err = r.Watcher.ProcessCACert(ctx, r.Client, &r.Loki.TLS, r.Namespace); err != nil {
			return err
		}
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

func (r *transformerReconciler) reconcileDynamicConfigMap(ctx context.Context, builder *transfoBuilder) error {
	newDCM, err := builder.dynamicConfigMap()
	if err != nil {
		return err
	}
	if !r.Managed.Exists(r.dynamicConfigMap) {
		if err := r.CreateOwned(ctx, newDCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newDCM.Data, r.dynamicConfigMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.dynamicConfigMap, newDCM); err != nil {
			return err
		}
	}
	return nil
}

func (r *transformerReconciler) reconcileDeployment(ctx context.Context, desiredFLP *flowslatest.FlowCollectorFLP, builder *transfoBuilder, annotations map[string]string) error {
	report := helper.NewChangeReport("FLP Deployment")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileDeployment(
		ctx,
		r.Instance,
		r.deployment,
		builder.deployment(annotations),
		constants.FLPName,
		helper.PtrInt32(desiredFLP.KafkaConsumerReplicas),
		&desiredFLP.KafkaConsumerAutoscaler,
		&report,
	)
}

func (r *transformerReconciler) reconcileHPA(ctx context.Context, desiredFLP *flowslatest.FlowCollectorFLP, builder *transfoBuilder) error {
	report := helper.NewChangeReport("FLP autoscaler")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileHPA(
		ctx,
		r.Instance,
		r.hpa,
		builder.autoScaler(),
		&desiredFLP.KafkaConsumerAutoscaler,
		&report,
	)
}

func (r *transformerReconciler) reconcilePrometheusService(ctx context.Context, builder *transfoBuilder) error {
	report := helper.NewChangeReport("FLP prometheus service")
	defer report.LogIfNeeded(ctx)

	if err := r.ReconcileService(ctx, r.promService, builder.promService(), &report); err != nil {
		return err
	}
	if r.ClusterInfo.HasSvcMonitor() {
		serviceMonitor := builder.generic.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	if r.ClusterInfo.HasPromRule() {
		promRules := builder.generic.prometheusRule()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.prometheusRule, promRules, &report, helper.PrometheusRuleChanged); err != nil {
			return err
		}
	}
	return nil
}

func (r *transformerReconciler) reconcilePermissions(ctx context.Context, builder *transfoBuilder) error {
	if !r.Managed.Exists(r.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // We only configure name, update is not needed for now

	roles := []constants.ClusterRoleName{
		constants.FLPInformersRole,
	}
	if helper.UseLoki(builder.generic.desired) && builder.generic.desired.Loki.Mode == flowslatest.LokiModeLokiStack {
		roles = append(roles, constants.LokiWriterRole)
	}
	bindings, crBindings := resources.GetAllBindings(
		r.Namespace,
		builder.generic.name(),
		builder.generic.name(),
		[]constants.RoleName{constants.ConfigWatcherRole},
		roles,
	)
	if err := r.ReconcileClusterRoleBindings(ctx, crBindings); err != nil {
		return err
	}
	return r.ReconcileRoleBindings(ctx, bindings)
}
