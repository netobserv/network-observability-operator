package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Status string

const (
	StatusUnknown    Status = "Unknown"
	StatusInProgress Status = "InProgress"
	StatusReady      Status = "Ready"
	StatusFailure    Status = "Failure"
)

type ComponentStatus struct {
	name    ComponentName
	status  Status
	reason  string
	message string
}

func (s *ComponentStatus) toCondition() metav1.Condition {
	c := metav1.Condition{
		Type:    string(s.name) + "Ready",
		Reason:  "Ready",
		Message: s.message,
	}
	if s.reason != "" {
		c.Reason = s.reason
	}
	switch s.status {
	case StatusUnknown:
		c.Status = metav1.ConditionUnknown
	case StatusFailure, StatusInProgress:
		c.Status = metav1.ConditionFalse
	case StatusReady:
		c.Status = metav1.ConditionTrue
	}
	return c
}
