package flp

import (
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
)

type transfoBuilder struct {
	generic builder
}

func newTransfoBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec, flowMetrics *metricslatest.FlowMetricList, detectedSubnets []flowslatest.SubnetLabel) (transfoBuilder, error) {
	gen, err := NewBuilder(info, desired, flowMetrics, detectedSubnets, ConfKafkaTransformer)
	return transfoBuilder{
		generic: gen,
	}, err
}

func (b *transfoBuilder) deployment(annotations map[string]string) *appsv1.Deployment {
	pod := b.generic.podTemplate(false /*no listen*/, false /*no host network*/, annotations)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.generic.name(),
			Namespace: b.generic.info.Namespace,
			Labels:    b.generic.labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: b.generic.desired.Processor.KafkaConsumerReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.generic.selector,
			},
			Template: pod,
		},
	}
}

func (b *transfoBuilder) staticConfigMap() (*corev1.ConfigMap, string, error) {
	pipeline := b.generic.NewKafkaPipeline()
	err := pipeline.AddProcessorStages()
	if err != nil {
		return nil, "", err
	}
	return b.generic.StaticConfigMap()
}

func (b *transfoBuilder) dynamicConfigMap() (*corev1.ConfigMap, error) {
	pipeline := b.generic.NewKafkaPipeline()
	err := pipeline.AddProcessorStages()
	if err != nil {
		return nil, err
	}
	return b.generic.DynamicConfigMap()
}

func (b *transfoBuilder) promService() *corev1.Service {
	return b.generic.promService()
}

func (b *transfoBuilder) autoScaler() *ascv2.HorizontalPodAutoscaler {
	return &ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.generic.name(),
			Namespace: b.generic.info.Namespace,
			Labels:    b.generic.labels,
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       b.generic.name(),
			},
			MinReplicas: b.generic.desired.Processor.KafkaConsumerAutoscaler.MinReplicas,
			MaxReplicas: b.generic.desired.Processor.KafkaConsumerAutoscaler.MaxReplicas,
			Metrics:     b.generic.desired.Processor.KafkaConsumerAutoscaler.Metrics,
		},
	}
}

func (b *transfoBuilder) serviceAccount() *corev1.ServiceAccount {
	return b.generic.serviceAccount()
}
