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
	assert.Len(t, conds, 5)
	assertHasCondition(t, conds, "Ready", "SomeFailures", metav1.ConditionFalse)
	assertHasCondition(t, conds, "FlowCollectorLegacyReady", "CreatingDaemonSet", metav1.ConditionFalse)
	assertHasCondition(t, conds, "MonitoringReady", "AnError", metav1.ConditionFalse)
	assertHasCondition(t, conds, "FlowCollectorLegacyFailure", "NoFailure", metav1.ConditionFalse)
	assertHasCondition(t, conds, "MonitoringFailure", "AnError", metav1.ConditionTrue)

	sl.SetReady() // temporary until controllers are broken down
	sl.CheckDaemonSetProgress(&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Status: appsv1.DaemonSetStatus{
		DesiredNumberScheduled: 3,
		UpdatedNumberScheduled: 1,
	}})
	sm.SetUnknown()

	conds = s.getConditions()
	assert.Len(t, conds, 5)
	assertHasCondition(t, conds, "Ready", "Pending", metav1.ConditionFalse)
	assertHasCondition(t, conds, "FlowCollectorLegacyReady", "DaemonSetNotReady", metav1.ConditionFalse)
	assertHasCondition(t, conds, "MonitoringReady", "Ready", metav1.ConditionUnknown)
	assertHasCondition(t, conds, "FlowCollectorLegacyFailure", "NoFailure", metav1.ConditionFalse)
	assertHasCondition(t, conds, "MonitoringFailure", "NoFailure", metav1.ConditionFalse)

	sl.SetReady() // temporary until controllers are broken down
	sl.CheckDaemonSetProgress(&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Status: appsv1.DaemonSetStatus{
		DesiredNumberScheduled: 3,
		UpdatedNumberScheduled: 3,
	}})
	sm.SetUnused()

	conds = s.getConditions()
	assert.Len(t, conds, 5)
	assertHasCondition(t, conds, "Ready", "Ready", metav1.ConditionTrue)
	assertHasCondition(t, conds, "FlowCollectorLegacyReady", "Ready", metav1.ConditionTrue)
	assertHasCondition(t, conds, "MonitoringReady", "ComponentUnused", metav1.ConditionUnknown)
	assertHasCondition(t, conds, "FlowCollectorLegacyFailure", "NoFailure", metav1.ConditionFalse)
	assertHasCondition(t, conds, "MonitoringFailure", "NoFailure", metav1.ConditionFalse)

	sl.SetReady() // temporary until controllers are broken down
	sl.CheckDeploymentProgress(&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Status: appsv1.DeploymentStatus{
		UpdatedReplicas: 2,
		Replicas:        2,
	}})
	sm.SetReady()

	conds = s.getConditions()
	assert.Len(t, conds, 5)
	assertHasCondition(t, conds, "Ready", "Ready", metav1.ConditionTrue)
	assertHasCondition(t, conds, "FlowCollectorLegacyReady", "Ready", metav1.ConditionTrue)
	assertHasCondition(t, conds, "MonitoringReady", "Ready", metav1.ConditionTrue)
	assertHasCondition(t, conds, "FlowCollectorLegacyFailure", "NoFailure", metav1.ConditionFalse)
	assertHasCondition(t, conds, "MonitoringFailure", "NoFailure", metav1.ConditionFalse)
}

func assertHasCondition(t *testing.T, conditions []*metav1.Condition, searchType, reason string, value metav1.ConditionStatus) {
	for _, c := range conditions {
		if c.Type == searchType {
			assert.Equal(t, reason, c.Reason, conditions)
			assert.Equal(t, value, c.Status, conditions)
			return
		}
	}
	assert.Fail(t, "Condition type not found", searchType, conditions)
}
