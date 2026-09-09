package domain

import (
	"math"
	"testing"
	"time"
)

func TestReinforceRaisesConfidenceAsymptotically(t *testing.T) {
	belief := mustBelief(t, "Prefiere agendar reuniones antes de las 11am", CategorySchedule)
	start := belief.Confidence
	belief.Reinforce()
	if belief.EvidenceCount != 2 {
		t.Fatalf("EvidenceCount = %d, want 2", belief.EvidenceCount)
	}
	if belief.Confidence <= start || belief.Confidence >= 1 {
		t.Fatalf("confidence should rise toward 1, start=%v got=%v", start, belief.Confidence)
	}
	firstGain := belief.Confidence - start
	mid := belief.Confidence
	belief.Reinforce()
	secondGain := belief.Confidence - mid
	if secondGain >= firstGain {
		t.Fatalf("gains should shrink as confidence approaches 1: first=%v second=%v", firstGain, secondGain)
	}
	if belief.LastReinforcedAt.Before(belief.FirstObservedAt) {
		t.Fatal("LastReinforcedAt should update on reinforce")
	}
}

func TestContradictLowersConfidenceFasterThanReinforce(t *testing.T) {
	belief := mustBelief(t, "Prefiere reuniones por la mañana", CategorySchedule)
	belief.Confidence = 0.5
	start := belief.Confidence
	reinforced := *belief
	reinforced.Reinforce()
	gain := reinforced.Confidence - start

	belief.Contradict()
	drop := start - belief.Confidence
	if drop <= gain {
		t.Fatalf("contradiction should outweigh one reinforce: drop=%v gain=%v", drop, gain)
	}
	if math.Abs(drop-gain*ContradictionWeight) > 1e-9 {
		t.Fatalf("contradiction should apply 2.5x reinforce weight: drop=%v gain=%v", drop, gain)
	}
	if belief.ContradictionCount != 1 || belief.LastContradictedAt == nil {
		t.Fatalf("contradiction metadata missing: %+v", belief)
	}
}

func TestDecayedConfidenceHalvesAfterHalfLife(t *testing.T) {
	belief := mustBelief(t, "Prefiere trabajo profundo por la mañana", CategoryWorkStyle)
	belief.Confidence = 0.8
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	belief.LastReinforcedAt = now.Add(-90 * 24 * time.Hour)
	got := belief.DecayedConfidence(now)
	want := 0.4
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("DecayedConfidence = %v, want ~%v", got, want)
	}
}

func TestSupersedeChainsWithoutDeleting(t *testing.T) {
	belief := mustBelief(t, "Prefiere reuniones por la mañana", CategorySchedule)
	if err := belief.Supersede("belief-2", "more specific schedule"); err != nil {
		t.Fatal(err)
	}
	if belief.Status != StatusSuperseded || belief.SupersededBy != "belief-2" {
		t.Fatalf("supersede did not chain: %+v", belief)
	}
	if err := belief.Supersede("belief-3", "again"); err == nil {
		t.Fatal("expected error superseding a non-active belief")
	}
}

func mustBelief(t *testing.T, statement, category string) *Belief {
	t.Helper()
	belief, err := NewBelief("b1", "user-1", BeliefCreate{Statement: statement, Category: category})
	if err != nil {
		t.Fatal(err)
	}
	return belief
}
