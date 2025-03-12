package flp

import (
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	transfoName           = constants.FLPName + "-transformer"
	transfoShortName      = constants.FLPShortName + "transfo"
	transfoConfigMap      = transfoName + "-config"
	transfoDynConfigMap   = transfoName + "-config-dynamic"
	transfoPromService    = transfoName + "-prom"
	transfoServiceMonitor = transfoName + "-monitor"
	transfoPromRule       = transfoName + "-alert"
)

type transfoBuilder struct {
	info            *reconcilers.Instance
	desired         *flowslatest.FlowCollectorSpec
	flowMetrics     *metricslatest.FlowMetricList
	detectedSubnets []flowslatest.SubnetLabel
	version         string
	promTLS         *flowslatest.CertificateReference
	volumes         volumes.Builder
}

func newTransfoBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec, flowMetrics *metricslatest.FlowMetricList, detectedSubnets []flowslatest.SubnetLabel) (transfoBuilder, error) {
	version := helper.ExtractVersion(info.Images[constants.ControllerBaseImageIndex])
	promTLS, err := getPromTLS(desired, transfoPromService)
	if err != nil {
		return transfoBuilder{}, err
	}
	return transfoBuilder{
		info:            info,
		desired:         desired,
		flowMetrics:     flowMetrics,
		detectedSubnets: detectedSubnets,
		version:         helper.MaxLabelLength(version),
		promTLS:         promTLS,
	}, nil
}

func (b *transfoBuilder) appLabel() map[string]string {
	return map[string]string{
		"app": transfoName,
	}
}

func (b *transfoBuilder) appVersionLabels() map[string]string {
	return map[string]string{
		"app":     transfoName,
		"version": b.version,
	}
}

func (b *transfoBuilder) deployment(annotations map[string]string) *appsv1.Deployment {
	pod := podTemplate(
		transfoName,
		b.version,
		b.info.Images[constants.ControllerBaseImageIndex],
		transfoConfigMap,
		b.desired,
		&b.volumes,
		false, /*no listen*/
		false, /*no host network*/
		annotations,
	)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      transfoName,
			Namespace: b.info.Namespace,
			Labels:    b.appVersionLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: b.desired.Processor.KafkaConsumerReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.appLabel(),
			},
			Template: pod,
		},
	}
}

func (b *transfoBuilder) configMaps() (*corev1.ConfigMap, string, *corev1.ConfigMap, error) {
	kafkaStage := newKafkaPipeline(b.desired, &b.volumes)
	pipeline := newPipelineBuilder(
		b.desired,
		b.flowMetrics,
		b.detectedSubnets,
		b.info.Loki,
		b.info.ClusterInfo.ID,
		&b.volumes,
		&kafkaStage,
	)
	err := pipeline.AddProcessorStages()
	if err != nil {
		return nil, "", nil, err
	}

	// Get static CM
	data, err := getStaticJSONConfig(b.desired, &b.volumes, b.promTLS, &pipeline, transfoDynConfigMap)
	if err != nil {
		return nil, "", nil, err
	}
	staticCM, digest, err := configMap(transfoConfigMap, b.info.Namespace, data, transfoName)
	if err != nil {
		return nil, "", nil, err
	}

	// Get dynamic CM (hot reload)
	data, err = getDynamicJSONConfig(&pipeline)
	if err != nil {
		return nil, "", nil, err
	}
	dynamicCM, _, err := configMap(transfoDynConfigMap, b.info.Namespace, data, transfoName)
	if err != nil {
		return nil, "", nil, err
	}

	return staticCM, digest, dynamicCM, err
}

func (b *transfoBuilder) promService() *corev1.Service {
	return promService(
		b.desired,
		transfoPromService,
		b.info.Namespace,
		transfoName,
	)
}

func (b *transfoBuilder) autoScaler() *ascv2.HorizontalPodAutoscaler {
	return &ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      transfoName,
			Namespace: b.info.Namespace,
			Labels:    b.appLabel(),
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       transfoName,
			},
			MinReplicas: b.desired.Processor.KafkaConsumerAutoscaler.MinReplicas,
			MaxReplicas: b.desired.Processor.KafkaConsumerAutoscaler.MaxReplicas,
			Metrics:     b.desired.Processor.KafkaConsumerAutoscaler.Metrics,
		},
	}
}

func (b *transfoBuilder) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      transfoName,
			Namespace: b.info.Namespace,
			Labels:    b.appLabel(),
		},
	}
}

func (b *transfoBuilder) serviceMonitor() *monitoringv1.ServiceMonitor {
	return serviceMonitor(
		b.desired,
		transfoServiceMonitor,
		transfoPromService,
		b.info.Namespace,
		transfoName,
		b.version,
		b.info.IsDownstream,
	)
}

func (b *transfoBuilder) prometheusRule() *monitoringv1.PrometheusRule {
	return prometheusRule(
		b.desired,
		transfoPromRule,
		b.info.Namespace,
		transfoName,
		b.version,
	)
}
