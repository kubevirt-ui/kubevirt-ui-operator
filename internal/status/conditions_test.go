package status

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetCondition_transitionTimeOnlyChangesWithStatus(t *testing.T) {
	var conds []metav1.Condition
	SetCondition(&conds, "Ready", metav1.ConditionFalse, "A", "first")
	if len(conds) != 1 || conds[0].LastTransitionTime.IsZero() {
		t.Fatalf("expected one condition with a transition time, got %+v", conds)
	}

	old := metav1.NewTime(time.Now().Add(-time.Hour))
	conds[0].LastTransitionTime = old

	SetCondition(&conds, "Ready", metav1.ConditionFalse, "B", "second")
	if conds[0].Reason != "B" || conds[0].Message != "second" {
		t.Fatalf("reason/message not updated: %+v", conds[0])
	}
	if !conds[0].LastTransitionTime.Equal(&old) {
		t.Fatalf("LastTransitionTime changed without a status change")
	}

	SetCondition(&conds, "Ready", metav1.ConditionTrue, "B", "second")
	if conds[0].LastTransitionTime.Equal(&old) {
		t.Fatalf("LastTransitionTime should change when status changes")
	}
}
