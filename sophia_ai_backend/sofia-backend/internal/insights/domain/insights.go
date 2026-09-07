package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	MoodVeryLow = "very_low"
	MoodLow     = "low"
	MoodNeutral = "neutral"
	MoodGood    = "good"
	MoodGreat   = "great"

	ResultCompleted          = "completed"
	ResultPartiallyCompleted = "partially_completed"
	ResultSkipped            = "skipped"
	ResultFailed             = "failed"
	ResultRescheduled        = "rescheduled"
)

var (
	ErrMoodEntryNotFound       = errors.New("mood entry not found")
	ErrActivityOutcomeNotFound = errors.New("activity outcome not found")
	ErrReflectionNotFound      = errors.New("reflection not found")
	ErrInvalidMood             = errors.New("invalid mood")
	ErrInvalidResult           = errors.New("invalid result")
	ErrInvalidLevel            = errors.New("level must be between 1 and 5")
	ErrInvalidOccurredAt       = errors.New("occurred_at is required")
	ErrInvalidTimezone         = errors.New("timezone is required")
	ErrInvalidContent          = errors.New("content is required")
	ErrInvalidActivityID       = errors.New("activity_id is required")
	ErrInvalidLimit            = errors.New("limit must be positive")
	ErrInvalidDateRange        = errors.New("from must be before to")
	ErrActivityNotAllowed      = errors.New("activity does not belong to the authenticated user")
)

type MoodEntry struct {
	ID          string
	UserID      string
	ActivityID  string
	Mood        string
	EnergyLevel int
	StressLevel int
	FocusLevel  int
	Note        string
	OccurredAt  time.Time
	Timezone    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ActivityOutcome struct {
	ID                string
	UserID            string
	ActivityID        string
	Result            string
	CompletionQuality int
	SatisfactionLevel int
	DifficultyLevel   int
	Note              string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Reflection struct {
	ID         string
	UserID     string
	ActivityID string
	Title      string
	Content    string
	Sentiment  string
	Tags       []string
	OccurredAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type MoodEntryCreate struct {
	ActivityID  string
	Mood        string
	EnergyLevel int
	StressLevel int
	FocusLevel  int
	Note        string
	OccurredAt  time.Time
	Timezone    string
}

type ActivityOutcomeCreate struct {
	ActivityID        string
	Result            string
	CompletionQuality int
	SatisfactionLevel int
	DifficultyLevel   int
	Note              string
}

type ReflectionCreate struct {
	ActivityID string
	Title      string
	Content    string
	Sentiment  string
	Tags       []string
	OccurredAt time.Time
}

type ListFilter struct {
	UserID     string
	ActivityID string
	From       *time.Time
	To         *time.Time
	Limit      int
}

type Summary struct {
	AverageEnergy    float64
	AverageStress    float64
	AverageFocus     float64
	TotalOutcomes    int
	TotalReflections int
}

type MoodEntryRepository interface {
	Create(ctx context.Context, entry *MoodEntry) error
	List(ctx context.Context, filter ListFilter) ([]*MoodEntry, error)
}

type ActivityOutcomeRepository interface {
	Create(ctx context.Context, outcome *ActivityOutcome) error
	List(ctx context.Context, filter ListFilter) ([]*ActivityOutcome, error)
}

type ReflectionRepository interface {
	Create(ctx context.Context, reflection *Reflection) error
	List(ctx context.Context, filter ListFilter) ([]*Reflection, error)
}

func NewMoodEntry(id, userID string, input MoodEntryCreate) (*MoodEntry, error) {
	now := time.Now()
	entry := &MoodEntry{
		ID:          id,
		UserID:      userID,
		ActivityID:  strings.TrimSpace(input.ActivityID),
		Mood:        strings.TrimSpace(input.Mood),
		EnergyLevel: input.EnergyLevel,
		StressLevel: input.StressLevel,
		FocusLevel:  input.FocusLevel,
		Note:        strings.TrimSpace(input.Note),
		OccurredAt:  input.OccurredAt,
		Timezone:    strings.TrimSpace(input.Timezone),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := entry.Validate(); err != nil {
		return nil, err
	}
	return entry, nil
}

func NewActivityOutcome(id, userID string, input ActivityOutcomeCreate) (*ActivityOutcome, error) {
	now := time.Now()
	outcome := &ActivityOutcome{
		ID:                id,
		UserID:            userID,
		ActivityID:        strings.TrimSpace(input.ActivityID),
		Result:            strings.TrimSpace(input.Result),
		CompletionQuality: input.CompletionQuality,
		SatisfactionLevel: input.SatisfactionLevel,
		DifficultyLevel:   input.DifficultyLevel,
		Note:              strings.TrimSpace(input.Note),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := outcome.Validate(); err != nil {
		return nil, err
	}
	return outcome, nil
}

func NewReflection(id, userID string, input ReflectionCreate) (*Reflection, error) {
	now := time.Now()
	reflection := &Reflection{
		ID:         id,
		UserID:     userID,
		ActivityID: strings.TrimSpace(input.ActivityID),
		Title:      strings.TrimSpace(input.Title),
		Content:    strings.TrimSpace(input.Content),
		Sentiment:  strings.TrimSpace(input.Sentiment),
		Tags:       normalizeTags(input.Tags),
		OccurredAt: input.OccurredAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := reflection.Validate(); err != nil {
		return nil, err
	}
	return reflection, nil
}

func (m *MoodEntry) Validate() error {
	if !IsValidMood(m.Mood) {
		return ErrInvalidMood
	}
	if !validLevel(m.EnergyLevel) || !validLevel(m.StressLevel) || !validLevel(m.FocusLevel) {
		return ErrInvalidLevel
	}
	if m.OccurredAt.IsZero() {
		return ErrInvalidOccurredAt
	}
	if strings.TrimSpace(m.Timezone) == "" {
		return ErrInvalidTimezone
	}
	return nil
}

func (o *ActivityOutcome) Validate() error {
	if strings.TrimSpace(o.ActivityID) == "" {
		return ErrInvalidActivityID
	}
	if !IsValidResult(o.Result) {
		return ErrInvalidResult
	}
	if !validLevel(o.CompletionQuality) || !validLevel(o.SatisfactionLevel) || !validLevel(o.DifficultyLevel) {
		return ErrInvalidLevel
	}
	return nil
}

func (r *Reflection) Validate() error {
	if strings.TrimSpace(r.Content) == "" {
		return ErrInvalidContent
	}
	if r.OccurredAt.IsZero() {
		return ErrInvalidOccurredAt
	}
	return nil
}

func ValidateFilter(filter ListFilter) error {
	if filter.Limit < 0 {
		return ErrInvalidLimit
	}
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return ErrInvalidDateRange
	}
	return nil
}

func IsValidMood(value string) bool {
	switch value {
	case MoodVeryLow, MoodLow, MoodNeutral, MoodGood, MoodGreat:
		return true
	default:
		return false
	}
}

func IsValidResult(value string) bool {
	switch value {
	case ResultCompleted, ResultPartiallyCompleted, ResultSkipped, ResultFailed, ResultRescheduled:
		return true
	default:
		return false
	}
}

func validLevel(value int) bool {
	return value >= 1 && value <= 5
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
