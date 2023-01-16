package helper

import (
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/netobserv/network-observability-operator/controllers/constants"
)

func PodChanged(old, new *corev1.PodTemplateSpec, containerName string) bool {
	if annotationsChanged(old, new) || volumesChanged(old, new) {
		return true
	}
	// Find containers
	oldContainer := FindContainer(&old.Spec, containerName)
	if oldContainer == nil {
		return true
	}
	newContainer := FindContainer(&new.Spec, containerName)
	if newContainer == nil {
		return true
	}
	return containerChanged(oldContainer, newContainer)
}

func annotationsChanged(old, new *corev1.PodTemplateSpec) bool {
	if old.Annotations == nil && new.Annotations == nil {
		return false
	}
	if old.Annotations == nil {
		return true
	}
	// Check domain annotations (config digest, certificate stamp...)
	for k, v := range old.Annotations {
		if strings.HasPrefix(k, constants.AnnotationDomain) {
			if new.Annotations[k] != v {
				return true
			}
		}
	}
	return false
}

func volumesChanged(old, new *corev1.PodTemplateSpec) bool {
	return !equality.Semantic.DeepDerivative(new.Spec.Volumes, old.Spec.Volumes)
}

func containerChanged(old, new *corev1.Container) bool {
	return new.Image != old.Image ||
		new.ImagePullPolicy != old.ImagePullPolicy ||
		!equality.Semantic.DeepDerivative(new.Args, old.Args) ||
		!equality.Semantic.DeepDerivative(new.Resources, old.Resources) ||
		!equality.Semantic.DeepEqual(old.LivenessProbe, new.LivenessProbe) ||
		!equality.Semantic.DeepEqual(old.StartupProbe, new.StartupProbe)
}

func ServiceChanged(old, new *corev1.Service) bool {
	return !equality.Semantic.DeepDerivative(new.ObjectMeta, old.ObjectMeta) ||
		!equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}

func ServiceMonitorChanged(old, new *monitoringv1.ServiceMonitor) bool {
	return !equality.Semantic.DeepDerivative(new.ObjectMeta, old.ObjectMeta) ||
		!equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}

func PrometheusRuleChanged(old, new *monitoringv1.PrometheusRule) bool {
	return !equality.Semantic.DeepDerivative(new.ObjectMeta, old.ObjectMeta) ||
		!equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}

// FindContainer searches in pod containers one that matches the provided name
func FindContainer(podSpec *corev1.PodSpec, name string) *corev1.Container {
	for i := range podSpec.Containers {
		if podSpec.Containers[i].Name == name {
			return &podSpec.Containers[i]
		}
	}
	return nil
}
