package domain

import (
	"context"
	"encoding/json"
	"errors"
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
)

var (
	ErrProposalNotFound     = errors.New("action proposal not found")
	ErrInvalidToolName      = errors.New("tool_name is required")
	ErrInvalidStatus        = errors.New("invalid action proposal status")
	ErrInvalidRiskLevel     = errors.New("invalid risk_level")
	ErrInvalidProposedInput = errors.New("proposed_input is required")
	ErrInvalidTransition    = errors.New("invalid action proposal transition")
	ErrToolDisabled         = errors.New("tool is disabled")
	ErrConfirmationRequired = errors.New("action proposal requires confirmation")
	ErrAutonomyNotAllowed   = errors.New("autonomy level does not allow this action")
)

type AIActionProposal struct {
	ID                   string          `json:"id"`
	UserID               string          `json:"user_id"`
	ToolName             string          `json:"tool_name"`
	Status               string          `json:"status"`
	ProposedInput        json.RawMessage `json:"proposed_input"`
	Reason               string          `json:"reason"`
	RiskLevel            string          `json:"risk_level"`
	RequiresConfirmation bool            `json:"requires_confirmation"`
	ExpiresAt            *time.Time      `json:"expires_at,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	ConfirmedAt          *time.Time      `json:"confirmed_at,omitempty"`
	RejectedAt           *time.Time      `json:"rejected_at,omitempty"`
	ExecutedAt           *time.Time      `json:"executed_at,omitempty"`
	ExecutionResult      json.RawMessage `json:"execution_result,omitempty"`
	ExecutionError       string          `json:"execution_error,omitempty"`
}

type ProposalCreate struct {
	ToolName      string
	ProposedInput json.RawMessage
	Reason        string
	RiskLevel     string
	ExpiresAt     *time.Time
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
	Update(ctx context.Context, proposal *AIActionProposal) error
}

func NewAIActionProposal(id, userID string, input ProposalCreate, requiresConfirmation bool) (*AIActionProposal, error) {
	now := time.Now()
	proposal := &AIActionProposal{
		ID:                   id,
		UserID:               userID,
		ToolName:             strings.TrimSpace(input.ToolName),
		Status:               StatusProposed,
		ProposedInput:        cloneJSON(input.ProposedInput),
		Reason:               strings.TrimSpace(input.Reason),
		RiskLevel:            defaultString(input.RiskLevel, RiskMedium),
		RequiresConfirmation: requiresConfirmation,
		ExpiresAt:            input.ExpiresAt,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := proposal.Validate(); err != nil {
		return nil, err
	}
	return proposal, nil
}

func (p *AIActionProposal) Confirm() error {
	if p.Status != StatusProposed {
		return ErrInvalidTransition
	}
	now := time.Now()
	p.Status = StatusConfirmed
	p.ConfirmedAt = &now
	p.UpdatedAt = now
	return nil
}

func (p *AIActionProposal) Reject() error {
	if p.Status != StatusProposed && p.Status != StatusConfirmed {
		return ErrInvalidTransition
	}
	now := time.Now()
	p.Status = StatusRejected
	p.RejectedAt = &now
	p.UpdatedAt = now
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
