package conditions

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TypeReady   = "Ready"
	TypePending = "Pending"
	TypeFailed  = "Failed"
)

type ErrorCondition struct {
	metav1.Condition
	Error error
}

func Updating() metav1.Condition {
	return metav1.Condition{
		Type:   TypePending,
		Reason: "Updating",
	}
}

func DeploymentInProgress() metav1.Condition {
	return metav1.Condition{
		Type:   TypePending,
		Reason: "DeploymentInProgress",
	}
}

func Ready() metav1.Condition {
	return metav1.Condition{
		Type:   TypeReady,
		Reason: "Ready",
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

func MissingLokiSecret(message string) *metav1.Condition {
	return &metav1.Condition{
		Type:    "LokiStack" + TypePending,
		Reason:  "MissingLokiSecret",
		Message: message,
	}
}

func WaitingDependentOperator(prefix string, message string) *metav1.Condition {
	return &metav1.Condition{
		Type:    prefix + TypePending,
		Reason:  "DependentOperatorInstanceMissing",
		Message: message,
	}
}

func ReconcileDependentOperatorsFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeFailed,
			Reason:  "ReconcileDependentOperatorsFailed",
			Message: "Failed to reconcile dependent operators: " + err.Error(),
		},
		Error: err,
	}
}

func ReconcileSecretsFailed(err error) *ErrorCondition {
	return &ErrorCondition{
		Condition: metav1.Condition{
			Type:    TypeFailed,
			Reason:  "ReconcileSecretsFailed",
			Message: "Failed to reconcile secrets: " + err.Error(),
		},
		Error: err,
	}
}

func clearPreviousStatuses(conditions *[]metav1.Condition) {
	for _, cond := range *conditions {
		cond.Status = metav1.ConditionFalse
		meta.SetStatusCondition(conditions, cond)
	}
}

func SetNewConditions(oldConditions *[]metav1.Condition, conditions *[]metav1.Condition) {
	clearPreviousStatuses(oldConditions)

	for _, cond := range *conditions {
		cond.Status = metav1.ConditionTrue
		meta.SetStatusCondition(oldConditions, cond)
	}
}

func DependenciesReady(conditions []metav1.Condition) bool {
	ready := true
	for _, c := range conditions {
		ready = ready && (c.Type == "LokiStackReady" || c.Type == "KafkaReady")
	}
	return ready
}
