package conditions

import (
	"sort"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TypePending    = "Pending"
	TypeFailed     = "Failed"
	MessagePending = "Some FlowCollector components pending on dependencies"
)

type ErrorCondition struct {
	metav1.Condition
	Error error
}

func Updating() *metav1.Condition {
	return &metav1.Condition{
		Type:    TypePending,
		Reason:  "Updating",
		Message: MessagePending,
	}
}

func DeploymentInProgress() *metav1.Condition {
	return &metav1.Condition{
		Type:    TypePending,
		Reason:  "DeploymentInProgress",
		Message: MessagePending,
	}
}

func Ready() *metav1.Condition {
	return &metav1.Condition{
		Type:    "Ready",
		Reason:  "Ready",
		Message: "All components ready",
	}
}

func CannotCreateNamespace(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeFailed,
			Reason:  "CannotCreateNamespace",
			Message: "Cannot create namespace: " + err.Error(),
		},
		Error: err,
	}
}

func NamespaceChangeFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeFailed,
			Reason:  "NamespaceChangeFailed",
			Message: "Failed to handle namespace change: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileFLPFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeFailed,
			Reason:  "ReconcileFLPFailed",
			Message: "Failed to reconcile flowlogs-pipeline: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileCNOFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeFailed,
			Reason:  "ReconcileCNOFailed",
			Message: "Failed to reconcile ovs-flows-config ConfigMap: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileOVNKFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeFailed,
			Reason:  "ReconcileOVNKFailed",
			Message: "Failed to reconcile ovn-kubernetes DaemonSet: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileAgentFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeFailed,
			Reason:  "ReconcileAgentFailed",
			Message: "Failed to reconcile eBPF Netobserv Agent: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileConsolePluginFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeFailed,
			Reason:  "ReconcileConsolePluginFailed",
			Message: "Failed to reconcile Console plugin: " + err.Error(),
		},
		Error: err,
	}
}

// set previous conditions to false as FlowCollector manage only one status at a time
func clearPreviousConditions(fc *flowslatest.FlowCollector) {
	for _, existingCondition := range fc.Status.Conditions {
		existingCondition.Status = metav1.ConditionFalse
		meta.SetStatusCondition(&fc.Status.Conditions, existingCondition)
	}
}

// sort conditions by date, latest first
func sortConditions(fc *flowslatest.FlowCollector) {
	sort.SliceStable(fc.Status.Conditions, func(i, j int) bool {
		return !fc.Status.Conditions[i].LastTransitionTime.Before(&fc.Status.Conditions[j].LastTransitionTime)
	})
}

// add a single condition to true, keeping the others to status false
func AddUniqueCondition(cond *metav1.Condition, fc *flowslatest.FlowCollector) {
	clearPreviousConditions(fc)
	cond.Status = metav1.ConditionTrue
	meta.SetStatusCondition(&fc.Status.Conditions, *cond)
	sortConditions(fc)
}
