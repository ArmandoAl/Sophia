package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	StatusProposed  = "proposed"
	StatusConfirmed = "confirmed"
	StatusRejected  = "rejected"
	StatusExecuted  = "executed"
	StatusFailed    = "failed"
	StatusExpired   = "expired"

	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"

	FeedbackApprovedDirect    = "approved_direct"
	FeedbackApprovedCorrected = "approved_corrected"
	FeedbackRejected          = "rejected"
	FeedbackIgnored           = "ignored"
	FeedbackExpired           = "expired"

	AutonomyModeProposed       = "proposed"
	AutonomyModeAutoExecuted   = "auto_executed"
	PredictionModelHeuristicV1 = "heuristic_v1"
)

var (
	ErrProposalNotFound     = errors.New("action proposal not found")
	ErrInvalidToolName      = errors.New("tool_name is required")
	ErrInvalidStatus        = errors.New("invalid action proposal status")
	ErrInvalidRiskLevel     = errors.New("invalid risk_level")
	ErrInvalidFeedback      = errors.New("invalid action proposal feedback")
	ErrInvalidProposedInput = errors.New("proposed_input is required")
	ErrInvalidTransition    = errors.New("invalid action proposal transition")
	ErrToolDisabled         = errors.New("tool is disabled")
	ErrConfirmationRequired = errors.New("action proposal requires confirmation")
	ErrAutonomyNotAllowed   = errors.New("autonomy level does not allow this action")
)

type AIActionProposal struct {
	ID                     string          `json:"id"`
	UserID                 string          `json:"user_id"`
	ConversationID         string          `json:"conversation_id,omitempty"`
	ToolName               string          `json:"tool_name"`
	Status                 string          `json:"status"`
	ProposedInput          json.RawMessage `json:"proposed_input"`
	Reason                 string          `json:"reason"`
	RiskLevel              string          `json:"risk_level"`
	RequiresConfirmation   bool            `json:"requires_confirmation"`
	ExpiresAt              *time.Time      `json:"expires_at,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	ConfirmedAt            *time.Time      `json:"confirmed_at,omitempty"`
	RejectedAt             *time.Time      `json:"rejected_at,omitempty"`
	ExecutedAt             *time.Time      `json:"executed_at,omitempty"`
	ExecutionResult        json.RawMessage `json:"execution_result,omitempty"`
	ExecutionError         string          `json:"execution_error,omitempty"`
	Feedback               string          `json:"feedback,omitempty"`
	CorrectedInput         json.RawMessage `json:"corrected_input,omitempty"`
	CorrectionDelta        []string        `json:"correction_delta,omitempty"`
	RejectionReason        string          `json:"rejection_reason,omitempty"`
	DecidedAt              *time.Time      `json:"decided_at,omitempty"`
	DecisionLatencyMS      int64           `json:"decision_latency_ms,omitempty"`
	PredictedApproval      float64         `json:"predicted_approval,omitempty"`
	PredictionBasis        []string        `json:"prediction_basis,omitempty"`
	PredictionModelVersion string          `json:"prediction_model_version,omitempty"`
	AutonomyModeUsed       string          `json:"autonomy_mode_used,omitempty"`
}

type ProposalCreate struct {
	ToolName       string
	ProposedInput  json.RawMessage
	Reason         string
	RiskLevel      string
	ExpiresAt      *time.Time
	ConversationID string
}

type ListFilter struct {
	UserID string
	Status string
	Limit  int
}

type AIActionProposalRepository interface {
	Create(ctx context.Context, proposal *AIActionProposal) error
	FindByID(ctx context.Context, userID, proposalID string) (*AIActionProposal, error)
	List(ctx context.Context, filter ListFilter) ([]*AIActionProposal, error)
	ListByDateRange(ctx context.Context, userID string, from, to time.Time) ([]*AIActionProposal, error)
	Update(ctx context.Context, proposal *AIActionProposal) error
}

func NewAIActionProposal(id, userID string, input ProposalCreate, requiresConfirmation bool) (*AIActionProposal, error) {
	now := time.Now()
	proposal := &AIActionProposal{
		ID:                     id,
		UserID:                 userID,
		ConversationID:         strings.TrimSpace(input.ConversationID),
		ToolName:               strings.TrimSpace(input.ToolName),
		Status:                 StatusProposed,
		ProposedInput:          cloneJSON(input.ProposedInput),
		Reason:                 strings.TrimSpace(input.Reason),
		RiskLevel:              defaultString(input.RiskLevel, RiskMedium),
		RequiresConfirmation:   requiresConfirmation,
		ExpiresAt:              input.ExpiresAt,
		CreatedAt:              now,
		UpdatedAt:              now,
		PredictionModelVersion: PredictionModelHeuristicV1,
		AutonomyModeUsed:       AutonomyModeProposed,
	}
	if err := proposal.Validate(); err != nil {
		return nil, err
	}
	return proposal, nil
}

func (p *AIActionProposal) Confirm(correctedInput json.RawMessage) error {
	if p.Status != StatusProposed {
		return ErrInvalidTransition
	}
	now := time.Now()
	p.Status = StatusConfirmed
	p.ConfirmedAt = &now
	p.UpdatedAt = now
	p.markDecision(now)
	if isEmptyJSON(correctedInput) || jsonEqual(correctedInput, p.ProposedInput) {
		p.Feedback = FeedbackApprovedDirect
		return nil
	}
	p.Feedback = FeedbackApprovedCorrected
	p.CorrectedInput = cloneJSON(correctedInput)
	p.CorrectionDelta = firstLevelDelta(p.ProposedInput, correctedInput)
	return nil
}

func (p *AIActionProposal) Reject(reason string) error {
	if p.Status != StatusProposed && p.Status != StatusConfirmed {
		return ErrInvalidTransition
	}
	now := time.Now()
	p.Status = StatusRejected
	p.RejectedAt = &now
	p.UpdatedAt = now
	p.Feedback = FeedbackRejected
	p.RejectionReason = strings.TrimSpace(reason)
	p.markDecision(now)
	return nil
}

func (p *AIActionProposal) MarkExecuted(result json.RawMessage) error {
	now := time.Now()
	p.Status = StatusExecuted
	p.ExecutedAt = &now
	p.ExecutionResult = cloneJSON(result)
	p.ExecutionError = ""
	p.UpdatedAt = now
	return p.Validate()
}

func (p *AIActionProposal) MarkFailed(message string) error {
	now := time.Now()
	p.Status = StatusFailed
	p.ExecutedAt = &now
	p.ExecutionError = strings.TrimSpace(message)
	p.UpdatedAt = now
	return p.Validate()
}

func (p *AIActionProposal) ExecutionInput() json.RawMessage {
	if !isEmptyJSON(p.CorrectedInput) {
		return p.CorrectedInput
	}
	return p.ProposedInput
}

func (p *AIActionProposal) markDecision(now time.Time) {
	p.DecidedAt = &now
	latency := now.Sub(p.CreatedAt).Milliseconds()
	if latency < 0 {
		latency = 0
	}
	p.DecisionLatencyMS = latency
}

func (p *AIActionProposal) Validate() error {
	if strings.TrimSpace(p.ToolName) == "" {
		return ErrInvalidToolName
	}
	if !json.Valid(p.ProposedInput) || len(p.ProposedInput) == 0 {
		return ErrInvalidProposedInput
	}
	if !IsValidStatus(p.Status) {
		return ErrInvalidStatus
	}
	if !IsValidRiskLevel(p.RiskLevel) {
		return ErrInvalidRiskLevel
	}
	if p.Feedback != "" && !IsValidFeedback(p.Feedback) {
		return ErrInvalidFeedback
	}
	return nil
}

func IsValidStatus(value string) bool {
	switch value {
	case StatusProposed, StatusConfirmed, StatusRejected, StatusExecuted, StatusFailed, StatusExpired:
		return true
	default:
		return false
	}
}

func IsValidRiskLevel(value string) bool {
	switch value {
	case RiskLow, RiskMedium, RiskHigh:
		return true
	default:
		return false
	}
}

func IsValidFeedback(value string) bool {
	switch value {
	case FeedbackApprovedDirect, FeedbackApprovedCorrected, FeedbackRejected, FeedbackIgnored, FeedbackExpired:
		return true
	default:
		return false
	}
}

func cloneJSON(value json.RawMessage) json.RawMessage {
	if value == nil {
		return nil
	}
	return append(json.RawMessage(nil), value...)
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func isEmptyJSON(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func jsonEqual(a, b json.RawMessage) bool {
	if isEmptyJSON(a) && isEmptyJSON(b) {
		return true
	}
	var left, right any
	if err := json.Unmarshal(a, &left); err != nil {
		return bytes.Equal(bytes.TrimSpace(a), bytes.TrimSpace(b))
	}
	if err := json.Unmarshal(b, &right); err != nil {
		return bytes.Equal(bytes.TrimSpace(a), bytes.TrimSpace(b))
	}
	return reflect.DeepEqual(left, right)
}

func firstLevelDelta(original, corrected json.RawMessage) []string {
	origKeys := firstLevelValues(original)
	corrKeys := firstLevelValues(corrected)
	seen := make(map[string]struct{}, len(origKeys)+len(corrKeys))
	for key := range origKeys {
		seen[key] = struct{}{}
	}
	for key := range corrKeys {
		seen[key] = struct{}{}
	}
	delta := make([]string, 0, len(seen))
	for key := range seen {
		left, leftOK := origKeys[key]
		right, rightOK := corrKeys[key]
		if !leftOK || !rightOK || !jsonEqual(left, right) {
			delta = append(delta, key)
		}
	}
	sort.Strings(delta)
	return delta
}

func firstLevelValues(raw json.RawMessage) map[string]json.RawMessage {
	var obj map[string]json.RawMessage
	if json.Unmarshal(raw, &obj) != nil || obj == nil {
		return map[string]json.RawMessage{}
	}
	return obj
}
