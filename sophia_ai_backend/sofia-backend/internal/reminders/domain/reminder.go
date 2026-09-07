package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	StatusPending   = "pending"
	StatusDue       = "due"
	StatusDelivered = "delivered"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"
	StatusArchived  = "archived"

	SourceUser        = "user"
	SourceActivity    = "activity"
	SourceAISuggested = "ai_suggested"
	SourceAICreated   = "ai_created"

	CreatedByUser  = "user"
	CreatedBySofia = "sofia"

	DefaultMaxRetries = 3

	RecurrenceNone    = "none"
	RecurrenceDaily   = "daily"
	RecurrenceWeekly  = "weekly"
	RecurrenceMonthly = "monthly"
)

var (
	ErrReminderNotFound   = errors.New("reminder not found")
	ErrInvalidTitle       = errors.New("title is required")
	ErrInvalidScheduledAt = errors.New("scheduled_at is required")
	ErrInvalidTimezone    = errors.New("timezone is required")
	ErrInvalidStatus      = errors.New("invalid reminder status")
	ErrInvalidSource      = errors.New("invalid reminder source")
	ErrInvalidCreatedBy   = errors.New("invalid created_by")
	ErrInvalidRecurrence  = errors.New("invalid recurrence_rule")
	ErrInvalidInterval    = errors.New("recurrence_interval must be positive")
	ErrInvalidCount       = errors.New("recurrence_count cannot be negative")
	ErrInvalidLimit       = errors.New("limit must be positive")
	ErrInvalidDateRange   = errors.New("from must be before to")
	ErrRemindersDisabled  = errors.New("reminders are disabled for this user")
	ErrActivityNotAllowed = errors.New("activity does not belong to the authenticated user")
	ErrReminderClaimed    = errors.New("reminder is already claimed")
)

type Reminder struct {
	ID                    string
	UserID                string
	ActivityID            string
	Title                 string
	Description           string
	Status                string
	ScheduledAt           time.Time
	Timezone              string
	RecurrenceRule        string
	RecurrenceInterval    int
	RecurrenceUntil       *time.Time
	RecurrenceCount       int
	RecurrenceOccurrences int
	NextRunAt             time.Time
	LastRunAt             *time.Time
	DeliveredAt           *time.Time
	FailedAt              *time.Time
	FailureReason         string
	RetryCount            int
	MaxRetries            int
	ProcessingBy          string
	ProcessingUntil       *time.Time
	LastAttemptAt         *time.Time
	Source                string
	CreatedBy             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ReminderCreate struct {
	ActivityID         string
	Title              string
	Description        string
	ScheduledAt        time.Time
	Timezone           string
	RecurrenceRule     string
	RecurrenceInterval int
	RecurrenceUntil    *time.Time
	RecurrenceCount    int
	Source             string
	CreatedBy          string
	MaxRetries         int
}

type ReminderUpdate struct {
	Title              *string
	Description        *string
	ScheduledAt        *time.Time
	Timezone           *string
	RecurrenceRule     *string
	RecurrenceInterval *int
	RecurrenceUntil    *time.Time
	RecurrenceCount    *int
	MaxRetries         *int
}

type ListFilter struct {
	UserID     string
	ActivityID string
	Status     string
	From       *time.Time
	To         *time.Time
	Limit      int
	Cursor     string
}

type ReminderRepository interface {
	Create(ctx context.Context, reminder *Reminder) error
	FindByID(ctx context.Context, userID, reminderID string) (*Reminder, error)
	List(ctx context.Context, filter ListFilter) ([]*Reminder, error)
	Update(ctx context.Context, reminder *Reminder) error
	Cancel(ctx context.Context, userID, reminderID string) (*Reminder, error)
	Archive(ctx context.Context, userID, reminderID string) (*Reminder, error)
	ListDue(ctx context.Context, userID string, now time.Time, limit int) ([]*Reminder, error)
	ClaimDue(ctx context.Context, userID, reminderID, workerID string, now, leaseUntil time.Time) (*Reminder, error)
	MarkDelivered(ctx context.Context, userID, reminderID string) (*Reminder, error)
	MarkFailed(ctx context.Context, userID, reminderID, reason string) (*Reminder, error)
}

func NewReminder(id, userID string, input ReminderCreate) (*Reminder, error) {
	now := time.Now()
	maxRetries := input.MaxRetries
	if maxRetries == 0 {
		maxRetries = DefaultMaxRetries
	}

	reminder := &Reminder{
		ID:                 id,
		UserID:             userID,
		ActivityID:         strings.TrimSpace(input.ActivityID),
		Title:              strings.TrimSpace(input.Title),
		Description:        strings.TrimSpace(input.Description),
		Status:             StatusPending,
		ScheduledAt:        input.ScheduledAt,
		Timezone:           strings.TrimSpace(input.Timezone),
		RecurrenceRule:     normalizeRecurrence(input.RecurrenceRule),
		RecurrenceInterval: input.RecurrenceInterval,
		RecurrenceUntil:    cloneTime(input.RecurrenceUntil),
		RecurrenceCount:    input.RecurrenceCount,
		NextRunAt:          input.ScheduledAt,
		MaxRetries:         maxRetries,
		Source:             defaultString(input.Source, SourceUser),
		CreatedBy:          defaultString(input.CreatedBy, CreatedByUser),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := reminder.Validate(); err != nil {
		return nil, err
	}
	return reminder, nil
}

func (r *Reminder) Apply(update ReminderUpdate) error {
	if update.Title != nil {
		r.Title = strings.TrimSpace(*update.Title)
	}
	if update.Description != nil {
		r.Description = strings.TrimSpace(*update.Description)
	}
	if update.ScheduledAt != nil {
		r.ScheduledAt = *update.ScheduledAt
		r.NextRunAt = *update.ScheduledAt
	}
	if update.Timezone != nil {
		r.Timezone = strings.TrimSpace(*update.Timezone)
	}
	if update.RecurrenceRule != nil {
		r.RecurrenceRule = normalizeRecurrence(*update.RecurrenceRule)
	}
	if update.RecurrenceInterval != nil {
		r.RecurrenceInterval = *update.RecurrenceInterval
	}
	if update.RecurrenceUntil != nil {
		r.RecurrenceUntil = cloneTime(update.RecurrenceUntil)
	}
	if update.RecurrenceCount != nil {
		r.RecurrenceCount = *update.RecurrenceCount
	}
	if update.MaxRetries != nil {
		r.MaxRetries = *update.MaxRetries
	}
	r.UpdatedAt = time.Now()
	return r.Validate()
}

func (r *Reminder) MarkCanceled() error {
	now := time.Now()
	r.Status = StatusCanceled
	r.UpdatedAt = now
	return r.Validate()
}

func (r *Reminder) MarkArchived() error {
	now := time.Now()
	r.Status = StatusArchived
	r.UpdatedAt = now
	return r.Validate()
}

func (r *Reminder) MarkDelivered() error {
	now := time.Now()
	r.LastRunAt = &now
	r.UpdatedAt = now
	r.ClearProcessing()
	if r.HasRecurrence() {
		r.RecurrenceOccurrences++
		next := r.nextOccurrence()
		if r.recurrenceFinished(next) {
			r.Status = StatusDelivered
			r.DeliveredAt = &now
			return r.Validate()
		}
		r.Status = StatusPending
		r.NextRunAt = next
		r.DeliveredAt = nil
		r.FailedAt = nil
		r.FailureReason = ""
		return r.Validate()
	}
	r.Status = StatusDelivered
	r.DeliveredAt = &now
	return r.Validate()
}

func (r *Reminder) MarkFailed(reason string) error {
	now := time.Now()
	r.Status = StatusFailed
	r.LastRunAt = &now
	r.FailedAt = &now
	r.FailureReason = strings.TrimSpace(reason)
	r.RetryCount++
	r.UpdatedAt = now
	r.ClearProcessing()
	return r.Validate()
}

func (r *Reminder) IsDue(now time.Time) bool {
	return r.Status == StatusPending && (r.NextRunAt.Equal(now) || r.NextRunAt.Before(now))
}

func (r *Reminder) IsClaimable(now time.Time) bool {
	return r.IsDue(now) && (r.ProcessingUntil == nil || !r.ProcessingUntil.After(now))
}

func (r *Reminder) Claim(workerID string, now, leaseUntil time.Time) error {
	if !r.IsClaimable(now) {
		return ErrReminderClaimed
	}
	r.ProcessingBy = strings.TrimSpace(workerID)
	r.ProcessingUntil = cloneTime(&leaseUntil)
	r.LastAttemptAt = cloneTime(&now)
	r.UpdatedAt = now
	return r.Validate()
}

func (r *Reminder) ClearProcessing() {
	r.ProcessingBy = ""
	r.ProcessingUntil = nil
}

func (r *Reminder) HasRecurrence() bool {
	return normalizeRecurrence(r.RecurrenceRule) != RecurrenceNone
}

func (r *Reminder) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return ErrInvalidTitle
	}
	if r.ScheduledAt.IsZero() {
		return ErrInvalidScheduledAt
	}
	if strings.TrimSpace(r.Timezone) == "" {
		return ErrInvalidTimezone
	}
	if !IsValidStatus(r.Status) {
		return ErrInvalidStatus
	}
	if !IsValidSource(r.Source) {
		return ErrInvalidSource
	}
	if !IsValidCreatedBy(r.CreatedBy) {
		return ErrInvalidCreatedBy
	}
	r.RecurrenceRule = normalizeRecurrence(r.RecurrenceRule)
	if !IsValidRecurrence(r.RecurrenceRule) {
		return ErrInvalidRecurrence
	}
	if r.HasRecurrence() {
		if r.RecurrenceInterval == 0 {
			r.RecurrenceInterval = 1
		}
		if r.RecurrenceInterval <= 0 {
			return ErrInvalidInterval
		}
	} else {
		r.RecurrenceInterval = 0
	}
	if r.RecurrenceCount < 0 {
		return ErrInvalidCount
	}
	if r.NextRunAt.IsZero() {
		r.NextRunAt = r.ScheduledAt
	}
	return nil
}

func (r *Reminder) nextOccurrence() time.Time {
	interval := r.RecurrenceInterval
	if interval <= 0 {
		interval = 1
	}
	base := r.NextRunAt
	if base.IsZero() {
		base = r.ScheduledAt
	}
	switch normalizeRecurrence(r.RecurrenceRule) {
	case RecurrenceDaily:
		return base.AddDate(0, 0, interval)
	case RecurrenceWeekly:
		return base.AddDate(0, 0, 7*interval)
	case RecurrenceMonthly:
		return base.AddDate(0, interval, 0)
	default:
		return base
	}
}

func (r *Reminder) recurrenceFinished(next time.Time) bool {
	if r.RecurrenceCount > 0 && r.RecurrenceOccurrences >= r.RecurrenceCount {
		return true
	}
	if r.RecurrenceUntil != nil && next.After(*r.RecurrenceUntil) {
		return true
	}
	return false
}

func ValidateFilter(filter ListFilter) error {
	if filter.Status != "" && !IsValidStatus(filter.Status) {
		return ErrInvalidStatus
	}
	if filter.Limit < 0 {
		return ErrInvalidLimit
	}
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return ErrInvalidDateRange
	}
	return nil
}

func IsValidStatus(value string) bool {
	switch value {
	case StatusPending, StatusDue, StatusDelivered, StatusFailed, StatusCanceled, StatusArchived:
		return true
	default:
		return false
	}
}

func IsValidSource(value string) bool {
	switch value {
	case SourceUser, SourceActivity, SourceAISuggested, SourceAICreated:
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

func IsValidRecurrence(value string) bool {
	switch normalizeRecurrence(value) {
	case RecurrenceNone, RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly:
		return true
	default:
		return false
	}
}

func normalizeRecurrence(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return RecurrenceNone
	}
	return value
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
