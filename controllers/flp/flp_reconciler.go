package flp

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type flpObjects struct {
	*reconcilers.Instance
	deployment     *appsv1.Deployment
	promService    *corev1.Service
	hpa            *ascv2.HorizontalPodAutoscaler
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	roleBinding    *rbacv1.ClusterRoleBinding
	serviceMonitor *monitoringv1.ServiceMonitor
	prometheusRule *monitoringv1.PrometheusRule
}

func newFLPObjects(cmn *reconcilers.Instance) *flpObjects {
	rec := flpObjects{
		Instance:       cmn,
		deployment:     cmn.Managed.NewDeployment(constants.FLPName),
		promService:    cmn.Managed.NewService(promServiceName(constants.FLPName)),
		hpa:            cmn.Managed.NewHPA(constants.FLPName),
		serviceAccount: cmn.Managed.NewServiceAccount(constants.FLPName),
		configMap:      cmn.Managed.NewConfigMap(configMapName(constants.FLPName)),
		roleBinding:    cmn.Managed.NewCRB(constants.FLPName),
	}
	if cmn.AvailableAPIs.HasSvcMonitor() {
		rec.serviceMonitor = cmn.Managed.NewServiceMonitor(serviceMonitorName(constants.FLPName))
	}
	if cmn.AvailableAPIs.HasPromRule() {
		rec.prometheusRule = cmn.Managed.NewPrometheusRule(prometheusRuleName(constants.FLPName))
	}
	return &rec
}

// cleanupNamespace cleans up old namespace
func (r *flpObjects) cleanupNamespace(ctx context.Context) {
	r.Managed.CleanupPreviousNamespace(ctx)
}

func (r *flpObjects) reconcile(ctx context.Context, desired *flowslatest.FlowCollector, flowMetrics *metricslatest.FlowMetricList) error {
	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if !helper.UseKafka(&desired.Spec) {
		r.Status.SetUnused("FLP only used with Kafka")
		r.Managed.TryDeleteAll(ctx)
		return nil
	}

	r.Status.SetReady() // will be overidden if necessary, as error or pending

	builder, err := newKafkaConsumerBuilder(r.Instance, &desired.Spec, flowMetrics)
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
	if !r.Managed.Exists(r.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return err
		}
	} else if !equality.Semantic.DeepDerivative(newCM.Data, r.configMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.configMap, newCM); err != nil {
			return err
		}
	}

	if err := reconcileRBAC(ctx, r.Common, constants.FLPName, constants.FLPName, r.Common.Namespace, builder.desired.Loki.Mode); err != nil {
		return err
	}

	if err := reconcilePrometheusService(ctx, r.Instance, r.promService, r.serviceMonitor, r.prometheusRule, builder); err != nil {
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

	if err = r.reconcileDeployment(ctx, &desired.Spec.Processor, builder, annotations); err != nil {
		return err
	}

	return r.reconcileHPA(ctx, &desired.Spec.Processor, builder)
}

func (r *flpObjects) reconcileDeployment(ctx context.Context, desiredFLP *flowslatest.FlowCollectorFLP, builder *Builder, annotations map[string]string) error {
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

func (r *flpObjects) reconcileHPA(ctx context.Context, desiredFLP *flowslatest.FlowCollectorFLP, builder *Builder) error {
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
