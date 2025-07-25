package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStatusWorkflow(t *testing.T) {
	s := NewManager()
	sl := s.ForComponent(FlowCollectorLegacy)
	sm := s.ForComponent(Monitoring)

	sl.SetReady() // temporary until controllers are broken down
	sl.SetCreatingDaemonSet(&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "test"}})
	sm.SetFailure("AnError", "bad one")

	conds := s.getConditions()
	assert.Len(t, conds, 4)
	assertHasCondition(t, conds, "Ready", "Failure", metav1.ConditionFalse)
	assertHasCondition(t, conds, "WaitingFlowCollectorLegacy", "CreatingDaemonSet", metav1.ConditionTrue)
	assertHasCondition(t, conds, "WaitingMonitoring", "AnError", metav1.ConditionTrue)

	sl.SetReady() // temporary until controllers are broken down
	sl.CheckDaemonSetProgress(&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Status: appsv1.DaemonSetStatus{
		DesiredNumberScheduled: 3,
		UpdatedNumberScheduled: 1,
	}})
	sm.SetUnknown()

	conds = s.getConditions()
	assert.Len(t, conds, 4)
	assertHasCondition(t, conds, "Ready", "Pending", metav1.ConditionFalse)
	assertHasCondition(t, conds, "WaitingFlowCollectorLegacy", "DaemonSetNotReady", metav1.ConditionTrue)
	assertHasCondition(t, conds, "WaitingMonitoring", "Unused", metav1.ConditionUnknown)

	sl.SetReady() // temporary until controllers are broken down
	sl.CheckDaemonSetProgress(&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Status: appsv1.DaemonSetStatus{
		DesiredNumberScheduled: 3,
		UpdatedNumberScheduled: 3,
	}})
	sm.SetUnused("message")

	conds = s.getConditions()
	assert.Len(t, conds, 4)
	assertHasCondition(t, conds, "Ready", "Ready", metav1.ConditionTrue)
	assertHasCondition(t, conds, "WaitingFlowCollectorLegacy", "Ready", metav1.ConditionFalse)
	assertHasCondition(t, conds, "WaitingMonitoring", "ComponentUnused", metav1.ConditionUnknown)

	sl.SetReady() // temporary until controllers are broken down
	sl.CheckDeploymentProgress(&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Status: appsv1.DeploymentStatus{
		UpdatedReplicas: 2,
		Replicas:        2,
	}})
	sm.SetReady()

	conds = s.getConditions()
	assert.Len(t, conds, 4)
	assertHasCondition(t, conds, "Ready", "Ready", metav1.ConditionTrue)
	assertHasCondition(t, conds, "WaitingFlowCollectorLegacy", "Ready", metav1.ConditionFalse)
	assertHasCondition(t, conds, "WaitingMonitoring", "Ready", metav1.ConditionFalse)
}

func assertHasCondition(t *testing.T, conditions []metav1.Condition, searchType, reason string, value metav1.ConditionStatus) {
	for _, c := range conditions {
		if c.Type == searchType {
			assert.Equal(t, reason, c.Reason, conditions)
			assert.Equal(t, value, c.Status, conditions)
			return
		}
	}
	assert.Fail(t, "Condition type not found", searchType, conditions)
}
