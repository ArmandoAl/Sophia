package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	TypeTask      = "task"
	TypeHabit     = "habit"
	TypeRoutine   = "routine"
	TypeGoal      = "goal"
	TypeEvent     = "event"
	TypeChecklist = "checklist"

	StatusPending   = "pending"
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusCanceled  = "canceled"
	StatusArchived  = "archived"

	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"

	SourceUser        = "user"
	SourceAISuggested = "ai_suggested"
	SourceAICreated   = "ai_created"
	SourceImported    = "imported"

	CreatedByUser  = "user"
	CreatedBySofia = "sofia"
)

var (
	ErrActivityNotFound = errors.New("activity not found")
	ErrInvalidTitle     = errors.New("title is required")
	ErrInvalidType      = errors.New("invalid activity type")
	ErrInvalidStatus    = errors.New("invalid activity status")
	ErrInvalidPriority  = errors.New("invalid activity priority")
	ErrInvalidTimezone  = errors.New("timezone is required")
	ErrInvalidSource    = errors.New("invalid activity source")
	ErrInvalidCreatedBy = errors.New("invalid created_by")
	ErrInvalidReminder  = errors.New("reminder_at requires has_reminder=true")
	ErrInvalidLimit     = errors.New("limit must be positive")
	ErrInvalidDateRange = errors.New("from must be before to")
)

type Activity struct {
	ID             string
	UserID         string
	Type           string
	Title          string
	Description    string
	Status         string
	Priority       string
	StartAt        *time.Time
	DueAt          *time.Time
	CompletedAt    *time.Time
	CanceledAt     *time.Time
	ArchivedAt     *time.Time
	Timezone       string
	Tags           []string
	Source         string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	HasReminder    bool
	ReminderAt     *time.Time
	RecurrenceRule string
	AIManageable   bool
	AINotes        string
}

type ActivityCreate struct {
	Type           string
	Title          string
	Description    string
	Priority       string
	StartAt        *time.Time
	DueAt          *time.Time
	Timezone       string
	Tags           []string
	HasReminder    bool
	ReminderAt     *time.Time
	RecurrenceRule string
	AIManageable   bool
	AINotes        string
}

type ActivityUpdate struct {
	Type           *string
	Title          *string
	Description    *string
	Status         *string
	Priority       *string
	StartAt        **time.Time
	DueAt          **time.Time
	Timezone       *string
	Tags           *[]string
	HasReminder    *bool
	ReminderAt     **time.Time
	RecurrenceRule *string
	AIManageable   *bool
	AINotes        *string
}

type ListFilter struct {
	UserID   string
	Status   string
	Type     string
	Priority string
	From     *time.Time
	To       *time.Time
	Tag      string
	Limit    int
	Cursor   string
}

type ActivityRepository interface {
	Create(ctx context.Context, activity *Activity) error
	FindByID(ctx context.Context, userID, activityID string) (*Activity, error)
	List(ctx context.Context, filter ListFilter) ([]*Activity, error)
	Update(ctx context.Context, activity *Activity) error
	Archive(ctx context.Context, userID, activityID string) (*Activity, error)
	Complete(ctx context.Context, userID, activityID string) (*Activity, error)
	Cancel(ctx context.Context, userID, activityID string) (*Activity, error)
}

func NewActivity(id, userID string, input ActivityCreate) (*Activity, error) {
	now := time.Now()
	activity := &Activity{
		ID:             id,
		UserID:         userID,
		Type:           defaultString(input.Type, TypeTask),
		Title:          strings.TrimSpace(input.Title),
		Description:    strings.TrimSpace(input.Description),
		Status:         StatusPending,
		Priority:       defaultString(input.Priority, PriorityMedium),
		StartAt:        input.StartAt,
		DueAt:          input.DueAt,
		Timezone:       strings.TrimSpace(input.Timezone),
		Tags:           normalizeTags(input.Tags),
		Source:         SourceUser,
		CreatedBy:      CreatedByUser,
		CreatedAt:      now,
		UpdatedAt:      now,
		HasReminder:    input.HasReminder,
		ReminderAt:     input.ReminderAt,
		RecurrenceRule: strings.TrimSpace(input.RecurrenceRule),
		AIManageable:   input.AIManageable,
		AINotes:        strings.TrimSpace(input.AINotes),
	}
	if err := activity.Validate(); err != nil {
		return nil, err
	}
	return activity, nil
}

func (a *Activity) Apply(update ActivityUpdate) error {
	if update.Type != nil {
		a.Type = strings.TrimSpace(*update.Type)
	}
	if update.Title != nil {
		a.Title = strings.TrimSpace(*update.Title)
	}
	if update.Description != nil {
		a.Description = strings.TrimSpace(*update.Description)
	}
	if update.Status != nil {
		a.Status = strings.TrimSpace(*update.Status)
	}
	if update.Priority != nil {
		a.Priority = strings.TrimSpace(*update.Priority)
	}
	if update.StartAt != nil {
		a.StartAt = *update.StartAt
	}
	if update.DueAt != nil {
		a.DueAt = *update.DueAt
	}
	if update.Timezone != nil {
		a.Timezone = strings.TrimSpace(*update.Timezone)
	}
	if update.Tags != nil {
		a.Tags = normalizeTags(*update.Tags)
	}
	if update.HasReminder != nil {
		a.HasReminder = *update.HasReminder
	}
	if update.ReminderAt != nil {
		a.ReminderAt = *update.ReminderAt
	}
	if update.RecurrenceRule != nil {
		a.RecurrenceRule = strings.TrimSpace(*update.RecurrenceRule)
	}
	if update.AIManageable != nil {
		a.AIManageable = *update.AIManageable
	}
	if update.AINotes != nil {
		a.AINotes = strings.TrimSpace(*update.AINotes)
	}

	a.UpdatedAt = time.Now()
	return a.Validate()
}

func (a *Activity) MarkCompleted() error {
	now := time.Now()
	a.Status = StatusCompleted
	a.CompletedAt = &now
	a.CanceledAt = nil
	a.UpdatedAt = now
	return a.Validate()
}

func (a *Activity) MarkCanceled() error {
	now := time.Now()
	a.Status = StatusCanceled
	a.CanceledAt = &now
	a.UpdatedAt = now
	return a.Validate()
}

func (a *Activity) MarkArchived() error {
	now := time.Now()
	a.Status = StatusArchived
	a.ArchivedAt = &now
	a.UpdatedAt = now
	return a.Validate()
}

func (a *Activity) Validate() error {
	if strings.TrimSpace(a.Title) == "" {
		return ErrInvalidTitle
	}
	if !IsValidType(a.Type) {
		return ErrInvalidType
	}
	if !IsValidStatus(a.Status) {
		return ErrInvalidStatus
	}
	if !IsValidPriority(a.Priority) {
		return ErrInvalidPriority
	}
	if strings.TrimSpace(a.Timezone) == "" {
		return ErrInvalidTimezone
	}
	if !IsValidSource(a.Source) {
		return ErrInvalidSource
	}
	if !IsValidCreatedBy(a.CreatedBy) {
		return ErrInvalidCreatedBy
	}
	if a.ReminderAt != nil && !a.HasReminder {
		return ErrInvalidReminder
	}
	return nil
}

func ValidateFilter(filter ListFilter) error {
	if filter.Status != "" && !IsValidStatus(filter.Status) {
		return ErrInvalidStatus
	}
	if filter.Type != "" && !IsValidType(filter.Type) {
		return ErrInvalidType
	}
	if filter.Priority != "" && !IsValidPriority(filter.Priority) {
		return ErrInvalidPriority
	}
	if filter.Limit < 0 {
		return ErrInvalidLimit
	}
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return ErrInvalidDateRange
	}
	return nil
}

func IsValidType(value string) bool {
	switch value {
	case TypeTask, TypeHabit, TypeRoutine, TypeGoal, TypeEvent, TypeChecklist:
		return true
	default:
		return false
	}
}

func IsValidStatus(value string) bool {
	switch value {
	case StatusPending, StatusActive, StatusCompleted, StatusCanceled, StatusArchived:
		return true
	default:
		return false
	}
}

func IsValidPriority(value string) bool {
	switch value {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

func IsValidSource(value string) bool {
	switch value {
	case SourceUser, SourceAISuggested, SourceAICreated, SourceImported:
		return true
	default:
		return false
	}
}

func IsValidCreatedBy(value string) bool {
	switch value {
	case CreatedByUser, CreatedBySofia:
		return true
	default:
		return false
	}
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

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
