package flp

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
)

type monolithBuilder struct {
	generic builder
}

func newMonolithBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec, flowMetrics *metricslatest.FlowMetricList, detectedSubnets []flowslatest.SubnetLabel) (monolithBuilder, error) {
	gen, err := NewBuilder(info, desired, flowMetrics, detectedSubnets, ConfMonolith)
	return monolithBuilder{
		generic: gen,
	}, err
}

func (b *monolithBuilder) daemonSet(annotations map[string]string) *appsv1.DaemonSet {
	pod := b.generic.podTemplate(true /*listens*/, !b.generic.info.ClusterInfo.IsOpenShift(), annotations)
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.generic.name(),
			Namespace: b.generic.info.Namespace,
			Labels:    b.generic.labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: b.generic.selector,
			},
			Template: pod,
		},
	}
}

func (b *monolithBuilder) staticConfigMap() (*corev1.ConfigMap, string, error) {
	pipeline := b.generic.NewGRPCPipeline()
	err := pipeline.AddProcessorStages()
	if err != nil {
		return nil, "", err
	}
	return b.generic.StaticConfigMap()
}

func (b *monolithBuilder) dynamicConfigMap() (*corev1.ConfigMap, error) {
	pipeline := b.generic.NewGRPCPipeline()
	err := pipeline.AddProcessorStages()
	if err != nil {
		return nil, err
	}
	return b.generic.DynamicConfigMap()
}

func (b *monolithBuilder) promService() *corev1.Service {
	return b.generic.promService()
}

func (b *monolithBuilder) serviceAccount() *corev1.ServiceAccount {
	return b.generic.serviceAccount()
}
