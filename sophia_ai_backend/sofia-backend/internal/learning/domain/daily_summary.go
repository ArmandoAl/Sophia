package domain

import (
	"context"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

type DailyDelta struct {
	Reinforced   []string
	Contradicted []string
	Novel        []string
}

type DailyStats struct {
	ProposalsTotal        int
	ApprovedDirect        int
	ApprovedCorrected     int
	Rejected              int
	Ignored               int
	MedianDecisionSeconds float64
}

type SynthesisCost struct {
	InputTokens  int
	OutputTokens int
	Model        string
	LatencyMS    int64
}

type DailySummary struct {
	ID                  string
	UserID              string
	Date                string
	GeneratedAt         time.Time
	Observations        []string
	DeltaVsPrevious     DailyDelta
	Stats               DailyStats
	SynthesisCost       SynthesisCost
	PromptVersionBefore string
	PromptVersionAfter  string
}

type DailySummaryCreate struct {
	Date                string
	Observations        []string
	DeltaVsPrevious     DailyDelta
	Stats               DailyStats
	SynthesisCost       SynthesisCost
	PromptVersionBefore string
	PromptVersionAfter  string
}

type DailySummaryRepository interface {
	Create(ctx context.Context, summary *DailySummary) error
	FindByDate(ctx context.Context, userID, date string) (*DailySummary, error)
	ListRecent(ctx context.Context, userID string, limit int) ([]*DailySummary, error)
}

func DailySummaryID(userID, date string) string {
	return strings.TrimSpace(userID) + "_" + strings.TrimSpace(date)
}

func NewDailySummary(userID string, input DailySummaryCreate) (*DailySummary, error) {
	date := strings.TrimSpace(input.Date)
	if _, err := time.Parse(DateLayout, date); err != nil {
		return nil, ErrInvalidDate
	}
	summary := &DailySummary{
		ID:                  DailySummaryID(userID, date),
		UserID:              strings.TrimSpace(userID),
		Date:                date,
		GeneratedAt:         time.Now().UTC(),
		Observations:        append([]string(nil), input.Observations...),
		DeltaVsPrevious:     cloneDailyDelta(input.DeltaVsPrevious),
		Stats:               input.Stats,
		SynthesisCost:       input.SynthesisCost,
		PromptVersionBefore: strings.TrimSpace(input.PromptVersionBefore),
		PromptVersionAfter:  strings.TrimSpace(input.PromptVersionAfter),
	}
	return summary, nil
}

func cloneDailyDelta(delta DailyDelta) DailyDelta {
	return DailyDelta{
		Reinforced:   append([]string(nil), delta.Reinforced...),
		Contradicted: append([]string(nil), delta.Contradicted...),
		Novel:        append([]string(nil), delta.Novel...),
	}
}
