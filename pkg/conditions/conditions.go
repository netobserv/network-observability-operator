package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const TypeReady = "Ready"

type ErrorCondition struct {
	metav1.Condition
	Error error
}

func Updating() *metav1.Condition {
	return &metav1.Condition{
		Type:   TypeReady,
		Status: metav1.ConditionFalse,
		Reason: "Updating",
	}
}

func DeploymentInProgress() *metav1.Condition {
	return &metav1.Condition{
		Type:   TypeReady,
		Status: metav1.ConditionFalse,
		Reason: "DeploymentInProgress",
	}
}

func Ready() *metav1.Condition {
	return &metav1.Condition{
		Type:   TypeReady,
		Status: metav1.ConditionTrue,
		Reason: "Ready",
	}
}

func CannotCreateNamespace(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  "CannotCreateNamespace",
			Message: "Cannot create namespace: " + err.Error(),
		},
		Error: err,
	}
}

func NamespaceChangeFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  "NamespaceChangeFailed",
			Message: "Failed to handle namespace change: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileFLPFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  "ReconcileFLPFailed",
			Message: "Failed to reconcile flowlogs-pipeline: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileCNOFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  "ReconcileCNOFailed",
			Message: "Failed to reconcile ovs-flows-config ConfigMap: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileOVNKFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  "ReconcileOVNKFailed",
			Message: "Failed to reconcile ovn-kubernetes DaemonSet: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileAgentFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  "ReconcileAgentFailed",
			Message: "Failed to reconcile eBPF Netobserv Agent: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileConsolePluginFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  "ReconcileConsolePluginFailed",
			Message: "Failed to reconcile Console plugin: " + err.Error(),
		},
		Error: err,
	}
}
