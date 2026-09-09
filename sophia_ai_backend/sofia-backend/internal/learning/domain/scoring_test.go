package domain

import (
	"math"
	"testing"
)

func TestPredictApprovalColdStartStaysUncertainAndNeverAutoExecutes(t *testing.T) {
	p, basis := PredictApproval(ToolHistory{}, nil)
	if math.Abs(p-0.5) > 0.01 {
		t.Fatalf("cold start probability = %v, want ~0.5", p)
	}
	if len(basis) != 0 {
		t.Fatalf("cold start should not cite beliefs, got %v", basis)
	}
	if ShouldAutoExecute(p, ToolHistory{}, true, DefaultAutonomyThreshold) {
		t.Fatal("cold start must never auto-execute")
	}
}

func TestPredictApprovalStrongRecentHistoryAutoExecutes(t *testing.T) {
	h := ToolHistory{Total: 40, ApprovedDirect: 36, Rejected: 4, DaysSinceLastSample: 1}
	p, _ := PredictApproval(h, nil)
	if p <= DefaultAutonomyThreshold {
		t.Fatalf("strong recent history probability = %v, want > %v", p, DefaultAutonomyThreshold)
	}
	if !ShouldAutoExecute(p, h, true, DefaultAutonomyThreshold) {
		t.Fatalf("strong recent reversible history should auto-execute, p=%v", p)
	}
}

func TestShouldAutoExecuteRejectsIrreversibleEvenAtHighProbability(t *testing.T) {
	h := ToolHistory{Total: 40, ApprovedDirect: 36, Rejected: 4, DaysSinceLastSample: 1}
	if ShouldAutoExecute(0.99, h, false, DefaultAutonomyThreshold) {
		t.Fatal("irreversible actions must never auto-execute")
	}
}

func TestPredictApprovalStrongContradictionDropsBelowThreshold(t *testing.T) {
	h := ToolHistory{Total: 40, ApprovedDirect: 36, Rejected: 4, DaysSinceLastSample: 1}
	baseline, _ := PredictApproval(h, nil)
	p, used := PredictApproval(h, []ScoredBelief{{ID: "belief-contradict", Confidence: 1, Contradicts: true}})
	if p >= DefaultAutonomyThreshold {
		t.Fatalf("contradictory belief should drop below threshold, p=%v", p)
	}
	if p >= baseline {
		t.Fatalf("contradiction should lower probability: baseline=%v got=%v", baseline, p)
	}
	if len(used) != 1 || used[0] != "belief-contradict" {
		t.Fatalf("expected contradicted belief in audit trail, got %v", used)
	}
	if ShouldAutoExecute(p, h, true, DefaultAutonomyThreshold) {
		t.Fatal("contradicted belief should block auto-execute")
	}
}

func TestPredictApprovalStaleHistoryReturnsToUncertainty(t *testing.T) {
	h := ToolHistory{Total: 40, ApprovedDirect: 36, Rejected: 4, DaysSinceLastSample: 8 * 30}
	p, _ := PredictApproval(h, nil)
	if math.Abs(p-0.5) > 0.05 {
		t.Fatalf("8 months idle should return to uncertainty, p=%v", p)
	}
	if ShouldAutoExecute(p, h, true, DefaultAutonomyThreshold) {
		t.Fatal("stale history must not auto-execute")
	}
}
