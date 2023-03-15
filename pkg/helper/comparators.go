package helper

import (
	"fmt"
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
)

func DeploymentChanged(old, new *appsv1.Deployment, contName string, checkReplicas bool, desiredReplicas int32, report *ChangeReport) bool {
	return report.Check("Pod changed", PodChanged(&old.Spec.Template, &new.Spec.Template, contName, report)) ||
		report.Check("Replicas changed", (checkReplicas && *old.Spec.Replicas != desiredReplicas))
}

func PodChanged(old, new *corev1.PodTemplateSpec, containerName string, report *ChangeReport) bool {
	if annotationsChanged(old, new, report) || volumesChanged(old, new, report) {
		return true
	}
	// Find containers
	oldContainer := FindContainer(&old.Spec, containerName)
	if oldContainer == nil {
		report.Add("Old container not found")
		return true
	}
	newContainer := FindContainer(&new.Spec, containerName)
	if newContainer == nil {
		report.Add("New container not found")
		return true
	}
	return report.Check("Container changed", containerChanged(oldContainer, newContainer, report))
}

func annotationsChanged(old, new *corev1.PodTemplateSpec, report *ChangeReport) bool {
	if old.Annotations == nil && new.Annotations == nil {
		return false
	}
	if old.Annotations == nil {
		report.Add("New annotations, previously none")
		return true
	}
	// Check domain annotations (config digest, certificate stamp...)
	for k, v := range old.Annotations {
		if strings.HasPrefix(k, constants.AnnotationDomain) {
			if new.Annotations[k] != v {
				report.Add(fmt.Sprintf("Annotation changed: '%s: %s'", k, v))
				return true
			}
		}
	}
	return false
}

func volumesChanged(old, new *corev1.PodTemplateSpec, report *ChangeReport) bool {
	return report.Check("Volumes changed", !equality.Semantic.DeepDerivative(new.Spec.Volumes, old.Spec.Volumes))
}

func containerChanged(old, new *corev1.Container, report *ChangeReport) bool {
	return report.Check("Image changed", new.Image != old.Image) ||
		report.Check("Pull policy changed", new.ImagePullPolicy != old.ImagePullPolicy) ||
		report.Check("Args changed", !equality.Semantic.DeepDerivative(new.Args, old.Args)) ||
		report.Check("Resources req/limit changed", !equality.Semantic.DeepDerivative(new.Resources, old.Resources)) ||
		report.Check("Liveness probe changed", probeChanged(new.LivenessProbe, old.LivenessProbe)) ||
		report.Check("Startup probe changed", probeChanged(new.StartupProbe, old.StartupProbe))
}

func probeChanged(old, new *corev1.Probe) bool {
	return (old == nil && new != nil) || (old != nil && new == nil)
}

func ServiceChanged(old, new *corev1.Service, report *ChangeReport) bool {
	return report.Check("Service meta changed", !equality.Semantic.DeepDerivative(new.ObjectMeta, old.ObjectMeta)) ||
		report.Check("Service spec changed", !equality.Semantic.DeepDerivative(new.Spec, old.Spec))
}

func ServiceMonitorChanged(old, new *monitoringv1.ServiceMonitor, report *ChangeReport) bool {
	return report.Check("ServiceMonitor spec changed", !equality.Semantic.DeepDerivative(new.Spec, old.Spec))
}

func PrometheusRuleChanged(old, new *monitoringv1.PrometheusRule, report *ChangeReport) bool {
	return report.Check("PrometheusRule spec changed", !equality.Semantic.DeepDerivative(new.Spec, old.Spec))
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

func AutoScalerChanged(asc *ascv2.HorizontalPodAutoscaler, desired flowslatest.FlowCollectorHPA, report *ChangeReport) bool {
	differentPointerValues := func(a, b *int32) bool {
		return (a == nil && b != nil) || (a != nil && b == nil) || (a != nil && *a != *b)
	}
	if report.Check("Max replicas changed", asc.Spec.MaxReplicas != desired.MaxReplicas) ||
		report.Check("Min replicas changed", differentPointerValues(asc.Spec.MinReplicas, desired.MinReplicas)) {
		return true
	}
	if report.Check("Metrics changed", !equality.Semantic.DeepDerivative(desired.Metrics, asc.Spec.Metrics)) {
		return true
	}
	return false
}
