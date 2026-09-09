package domain

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
)

func TestConfirmDirectSetsApprovedDirectFeedback(t *testing.T) {
	proposal := newTestProposal(t)
	proposal.CreatedAt = time.Now().Add(-2 * time.Second)

	if err := proposal.Confirm(nil); err != nil {
		t.Fatal(err)
	}
	if proposal.Status != StatusConfirmed || proposal.Feedback != FeedbackApprovedDirect {
		t.Fatalf("unexpected confirm: status=%s feedback=%s", proposal.Status, proposal.Feedback)
	}
	if len(proposal.CorrectedInput) != 0 || len(proposal.CorrectionDelta) != 0 {
		t.Fatalf("direct confirm stored a correction: %+v", proposal)
	}
	if proposal.DecidedAt == nil || proposal.DecisionLatencyMS < 2000 {
		t.Fatalf("expected decision latency from created_at, got decided_at=%v latency=%d", proposal.DecidedAt, proposal.DecisionLatencyMS)
	}
}

func TestConfirmCorrectedComputesFirstLevelDelta(t *testing.T) {
	proposal := newTestProposal(t)
	corrected := json.RawMessage(`{"title":"Gym","timezone":"UTC","priority":"high"}`)

	if err := proposal.Confirm(corrected); err != nil {
		t.Fatal(err)
	}
	if proposal.Feedback != FeedbackApprovedCorrected {
		t.Fatalf("expected %s, got %s", FeedbackApprovedCorrected, proposal.Feedback)
	}
	if !jsonEqual(proposal.CorrectedInput, corrected) {
		t.Fatalf("corrected input not stored: %s", proposal.CorrectedInput)
	}
	if !slices.Equal(proposal.CorrectionDelta, []string{"priority", "title"}) {
		t.Fatalf("unexpected correction delta: %+v", proposal.CorrectionDelta)
	}
}

func TestRejectStoresReasonAndRejectedFeedback(t *testing.T) {
	proposal := newTestProposal(t)
	proposal.CreatedAt = time.Now().Add(-time.Second)

	if err := proposal.Reject("wrong_time"); err != nil {
		t.Fatal(err)
	}
	if proposal.Status != StatusRejected || proposal.Feedback != FeedbackRejected || proposal.RejectionReason != "wrong_time" {
		t.Fatalf("unexpected reject: %+v", proposal)
	}
	if proposal.DecidedAt == nil || proposal.DecisionLatencyMS < 1000 {
		t.Fatalf("expected decision metadata, got decided_at=%v latency=%d", proposal.DecidedAt, proposal.DecisionLatencyMS)
	}
}

func TestIsValidFeedback(t *testing.T) {
	if !IsValidFeedback(FeedbackApprovedDirect) || !IsValidFeedback(FeedbackIgnored) || IsValidFeedback("maybe") {
		t.Fatal("IsValidFeedback does not match expected constants")
	}
}

func newTestProposal(t *testing.T) *AIActionProposal {
	t.Helper()
	proposal, err := NewAIActionProposal("proposal-1", "user-1", ProposalCreate{
		ToolName:      "create_activity",
		ProposedInput: json.RawMessage(`{"title":"Run","timezone":"UTC"}`),
		RiskLevel:     RiskLow,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	return proposal
}
