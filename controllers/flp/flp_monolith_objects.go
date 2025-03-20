package flp

import (
	appsv1 "k8s.io/api/apps/v1"
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
	monoName           = constants.FLPName
	monoShortName      = constants.FLPShortName
	monoConfigMap      = monoName + "-config"
	monoDynConfigMap   = monoName + "-config-dynamic"
	monoPromService    = monoName + "-prom"
	monoServiceMonitor = monoName + "-monitor"
	monoPromRule       = monoName + "-alert"
)

type monolithBuilder struct {
	info            *reconcilers.Instance
	desired         *flowslatest.FlowCollectorSpec
	flowMetrics     *metricslatest.FlowMetricList
	detectedSubnets []flowslatest.SubnetLabel
	version         string
	promTLS         *flowslatest.CertificateReference
	volumes         volumes.Builder
}

func newMonolithBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec, flowMetrics *metricslatest.FlowMetricList, detectedSubnets []flowslatest.SubnetLabel) (monolithBuilder, error) {
	version := helper.ExtractVersion(info.Images[constants.ControllerBaseImageIndex])
	promTLS, err := getPromTLS(desired, monoPromService)
	if err != nil {
		return monolithBuilder{}, err
	}
	return monolithBuilder{
		info:            info,
		desired:         desired,
		flowMetrics:     flowMetrics,
		detectedSubnets: detectedSubnets,
		version:         helper.MaxLabelLength(version),
		promTLS:         promTLS,
	}, nil
}

func (b *monolithBuilder) appLabel() map[string]string {
	return map[string]string{
		"app": monoName,
	}
}

func (b *monolithBuilder) appVersionLabels() map[string]string {
	return map[string]string{
		"app":     monoName,
		"version": b.version,
	}
}

func (b *monolithBuilder) daemonSet(annotations map[string]string) *appsv1.DaemonSet {
	pod := podTemplate(
		monoName,
		b.version,
		b.info.Images[constants.ControllerBaseImageIndex],
		monoConfigMap,
		b.desired,
		&b.volumes,
		true, /*listens*/
		!b.info.ClusterInfo.IsOpenShift(),
		annotations,
	)
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      monoName,
			Namespace: b.info.Namespace,
			Labels:    b.appVersionLabels(),
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: b.appLabel(),
			},
			Template: pod,
		},
	}
}

func (b *monolithBuilder) configMaps() (*corev1.ConfigMap, string, *corev1.ConfigMap, error) {
	kafkaStage := newGRPCPipeline(b.desired)
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
	data, err := getStaticJSONConfig(b.desired, &b.volumes, b.promTLS, &pipeline, monoDynConfigMap)
	if err != nil {
		return nil, "", nil, err
	}
	staticCM, digest, err := configMap(monoConfigMap, b.info.Namespace, data, monoName)
	if err != nil {
		return nil, "", nil, err
	}

	// Get dynamic CM (hot reload)
	data, err = getDynamicJSONConfig(&pipeline)
	if err != nil {
		return nil, "", nil, err
	}
	dynamicCM, _, err := configMap(monoDynConfigMap, b.info.Namespace, data, monoName)
	if err != nil {
		return nil, "", nil, err
	}

	return staticCM, digest, dynamicCM, err
}

func (b *monolithBuilder) promService() *corev1.Service {
	return promService(
		b.desired,
		monoPromService,
		b.info.Namespace,
		monoName,
	)
}

func (b *monolithBuilder) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      monoName,
			Namespace: b.info.Namespace,
			Labels:    b.appLabel(),
		},
	}
}

func (b *monolithBuilder) serviceMonitor() *monitoringv1.ServiceMonitor {
	return serviceMonitor(
		b.desired,
		monoServiceMonitor,
		monoPromService,
		b.info.Namespace,
		monoName,
		b.version,
		b.info.IsDownstream,
	)
}

func (b *monolithBuilder) prometheusRule() *monitoringv1.PrometheusRule {
	return prometheusRule(
		b.desired,
		monoPromRule,
		b.info.Namespace,
		monoName,
		b.version,
	)
}
