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
		Type:    "Waiting" + string(s.name),
		Message: s.message,
	}
	switch s.status {
	case StatusUnknown:
		c.Status = metav1.ConditionUnknown
		c.Reason = "Unused"
	case StatusFailure, StatusInProgress:
		c.Status = metav1.ConditionTrue
		c.Reason = "NotReady"
	case StatusReady:
		c.Status = metav1.ConditionFalse
		c.Reason = "Ready"
	}
	if s.reason != "" {
		c.Reason = s.reason
	}
	return c
}
