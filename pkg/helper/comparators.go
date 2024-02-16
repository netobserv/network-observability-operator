package helper

import (
	"fmt"
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
)

type ReconcileAction int

const (
	ActionNone = iota
	ActionCreate
	ActionUpdate
)

func DaemonSetChanged(current, desired *appsv1.DaemonSet) ReconcileAction {
	if desired == nil {
		return ActionNone
	}
	if current == nil {
		return ActionCreate
	}
	cSpec, dSpec := current.Spec, desired.Spec
	eq := equality.Semantic.DeepDerivative
	if !IsSubSet(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) ||
		!eq(dSpec.Selector, cSpec.Selector) ||
		!eq(dSpec.Template, cSpec.Template) ||
		assignationChanged(&cSpec.Template, &dSpec.Template, nil) {

		return ActionUpdate
	}

	// Env vars aren't covered by DeepDerivative when they are removed: deep-compare them
	dConts := dSpec.Template.Spec.Containers
	cConts := cSpec.Template.Spec.Containers
	if len(dConts) > 0 && len(cConts) > 0 && !equality.Semantic.DeepEqual(dConts[0].Env, cConts[0].Env) {
		return ActionUpdate
	}

	return ActionNone
}

func DeploymentChanged(old, new *appsv1.Deployment, contName string, checkReplicas bool, desiredReplicas int32, report *ChangeReport) bool {
	return report.Check("Pod changed", PodChanged(&old.Spec.Template, &new.Spec.Template, contName, report)) ||
		report.Check("Replicas changed", (checkReplicas && *old.Spec.Replicas != desiredReplicas))
}

func PodChanged(old, new *corev1.PodTemplateSpec, containerName string, report *ChangeReport) bool {
	if annotationsChanged(old, new, report) || volumesChanged(old, new, report) || assignationChanged(old, new, report) {
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

func assignationChanged(old, new *corev1.PodTemplateSpec, report *ChangeReport) bool {
	if !equality.Semantic.DeepDerivative(old.Spec.NodeSelector, new.Spec.NodeSelector) {
		if report != nil {
			report.Add("NodeSelector changed")
		}
		return true
	}
	if !equality.Semantic.DeepDerivative(old.Spec.Affinity, new.Spec.Affinity) {
		if report != nil {
			report.Add("Affinity changed")
		}
		return true
	}
	if !equality.Semantic.DeepDerivative(old.Spec.PriorityClassName, new.Spec.PriorityClassName) {
		if report != nil {
			report.Add("PriorityClassName changed")
		}
		return true
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
	return report.Check("Service annotations changed", !equality.Semantic.DeepDerivative(new.Annotations, old.Annotations)) ||
		report.Check("Service labels changed", !equality.Semantic.DeepDerivative(new.Labels, old.Labels)) ||
		report.Check("Service spec changed", !equality.Semantic.DeepDerivative(new.Spec, old.Spec))
}

func ServiceMonitorChanged(old, new *monitoringv1.ServiceMonitor, report *ChangeReport) bool {
	return report.Check("ServiceMonitor spec changed", !equality.Semantic.DeepDerivative(new.Spec, old.Spec)) ||
		report.Check("ServiceMonitor labels changed", !IsSubSet(old.Labels, new.Labels))
}

func PrometheusRuleChanged(old, new *monitoringv1.PrometheusRule, report *ChangeReport) bool {
	// Note: DeepDerivative misses changes in Spec.Groups.Rules (covered by test "Expecting PrometheusRule to exist and be updated")
	return report.Check("PrometheusRule spec changed", !equality.Semantic.DeepEqual(new.Spec, old.Spec)) ||
		report.Check("PrometheusRule labels changed", !IsSubSet(old.Labels, new.Labels))
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
