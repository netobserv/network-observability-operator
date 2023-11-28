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

func (s *ComponentStatus) readyCondition() metav1.Condition {
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

func (s *ComponentStatus) failureCondition() metav1.Condition {
	c := metav1.Condition{
		Type: string(s.name) + "Failure",
	}
	switch s.status {
	case StatusFailure:
		c.Status = metav1.ConditionTrue
		c.Reason = s.reason
		c.Message = s.message
	case StatusReady, StatusInProgress, StatusUnknown:
		c.Status = metav1.ConditionFalse
		c.Reason = "NoFailure"
	}
	return c
}

func (s *ComponentStatus) toConditions() []*metav1.Condition {
	r := s.readyCondition()
	f := s.failureCondition()
	return []*metav1.Condition{&r, &f}
}
