package flp

import (
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

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

func (b *transfoBuilder) cacheDeployment(annotations map[string]string) *appsv1.Deployment {
	pod := b.generic.cachePodTemplate(annotations)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flpCacheName,
			Namespace: b.generic.info.Namespace,
			Labels:    b.generic.cacheLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: b.generic.cacheSelector,
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

// The operator needs to have at least the same permissions as flowlogs-pipeline in order to grant them
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=create;delete;patch;update;get;watch;list
//+kubebuilder:rbac:groups=core,resources=pods;services;nodes;configmaps,verbs=get;list;watch

func BuildClusterRoleTransformer() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name(ConfKafkaTransformer),
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Verbs:     []string{"list", "get", "watch"},
			Resources: []string{"pods", "services", "nodes", "configmaps"},
		}, {
			APIGroups: []string{"apps"},
			Verbs:     []string{"list", "get", "watch"},
			Resources: []string{"replicasets"},
		}, {
			APIGroups: []string{"autoscaling"},
			Verbs:     []string{"create", "delete", "patch", "update", "get", "watch", "list"},
			Resources: []string{"horizontalpodautoscalers"},
		}},
	}
}

func (b *transfoBuilder) serviceAccount() *corev1.ServiceAccount {
	return b.generic.serviceAccount()
}

func (b *transfoBuilder) clusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return b.generic.clusterRoleBinding(ConfKafkaTransformer, false)
}
