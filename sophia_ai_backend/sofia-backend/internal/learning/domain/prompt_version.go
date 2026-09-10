package domain

import (
	"context"
	"strings"
	"time"
)

const PromptTokenCeiling = 1000

type PromptQuality struct {
	ApprovalRate   float64
	CorrectionRate float64
	SampleSize     int
}

type PromptVersion struct {
	ID                     string
	UserID                 string
	Version                int
	Content                string
	TokenCount             int
	BeliefIDs              []string
	GeneratedFromSummaryID string
	CreatedAt              time.Time
	Active                 bool
	Quality                PromptQuality
}

type PromptVersionCreate struct {
	Content                string
	BeliefIDs              []string
	GeneratedFromSummaryID string
	Quality                PromptQuality
}

type PromptVersionRepository interface {
	CreateActive(ctx context.Context, version *PromptVersion) error
	DeactivateActive(ctx context.Context, userID string) error
	GetActive(ctx context.Context, userID string) (*PromptVersion, error)
	ListVersions(ctx context.Context, userID string, limit int) ([]*PromptVersion, error)
}

func NewPromptVersion(id, userID string, input PromptVersionCreate) (*PromptVersion, error) {
	version := &PromptVersion{
		ID:                     strings.TrimSpace(id),
		UserID:                 strings.TrimSpace(userID),
		Content:                strings.TrimSpace(input.Content),
		BeliefIDs:              append([]string(nil), input.BeliefIDs...),
		GeneratedFromSummaryID: strings.TrimSpace(input.GeneratedFromSummaryID),
		CreatedAt:              time.Now().UTC(),
		Active:                 false,
		Quality:                input.Quality,
	}
	version.TokenCount = ApproximateTokens(version.Content)
	if err := version.Validate(); err != nil {
		return nil, err
	}
	return version, nil
}

func (p *PromptVersion) Validate() error {
	if strings.TrimSpace(p.Content) == "" {
		return ErrInvalidPromptContent
	}
	return nil
}
