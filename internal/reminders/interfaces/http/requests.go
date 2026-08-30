package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/reminders/application"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
)

type createReminderRequest struct {
	ActivityID         string `json:"activity_id"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	ScheduledAt        string `json:"scheduled_at"`
	Timezone           string `json:"timezone"`
	RecurrenceRule     string `json:"recurrence_rule"`
	RecurrenceInterval int    `json:"recurrence_interval"`
	RecurrenceUntil    string `json:"recurrence_until"`
	RecurrenceCount    int    `json:"recurrence_count"`
	MaxRetries         int    `json:"max_retries"`
}

func (r createReminderRequest) toDomain() (domain.ReminderCreate, error) {
	scheduledAt, err := parseRequiredTime(r.ScheduledAt)
	if err != nil {
		return domain.ReminderCreate{}, err
	}
	recurrenceUntil, err := parseOptionalTimeString(r.RecurrenceUntil)
	if err != nil {
		return domain.ReminderCreate{}, err
	}
	return domain.ReminderCreate{
		ActivityID:         r.ActivityID,
		Title:              r.Title,
		Description:        r.Description,
		ScheduledAt:        scheduledAt,
		Timezone:           r.Timezone,
		RecurrenceRule:     r.RecurrenceRule,
		RecurrenceInterval: r.RecurrenceInterval,
		RecurrenceUntil:    recurrenceUntil,
		RecurrenceCount:    r.RecurrenceCount,
		MaxRetries:         r.MaxRetries,
		Source:             domain.SourceUser,
		CreatedBy:          domain.CreatedByUser,
	}, nil
}

type updateReminderRequest struct {
	Title              *string `json:"title"`
	Description        *string `json:"description"`
	ScheduledAt        *string `json:"scheduled_at"`
	Timezone           *string `json:"timezone"`
	RecurrenceRule     *string `json:"recurrence_rule"`
	RecurrenceInterval *int    `json:"recurrence_interval"`
	RecurrenceUntil    *string `json:"recurrence_until"`
	RecurrenceCount    *int    `json:"recurrence_count"`
	MaxRetries         *int    `json:"max_retries"`
}

func (r updateReminderRequest) toDomain() (domain.ReminderUpdate, error) {
	scheduledAt, err := parseOptionalTime(r.ScheduledAt)
	if err != nil {
		return domain.ReminderUpdate{}, err
	}
	recurrenceUntil, err := parseOptionalTime(r.RecurrenceUntil)
	if err != nil {
		return domain.ReminderUpdate{}, err
	}
	return domain.ReminderUpdate{
		Title:              r.Title,
		Description:        r.Description,
		ScheduledAt:        scheduledAt,
		Timezone:           r.Timezone,
		RecurrenceRule:     r.RecurrenceRule,
		RecurrenceInterval: r.RecurrenceInterval,
		RecurrenceUntil:    recurrenceUntil,
		RecurrenceCount:    r.RecurrenceCount,
		MaxRetries:         r.MaxRetries,
	}, nil
}

func listFilterFromRequest(r *http.Request, userID string) (domain.ListFilter, error) {
	query := r.URL.Query()
	from, err := parseQueryTime(query.Get("from"))
	if err != nil {
		return domain.ListFilter{}, err
	}
	to, err := parseQueryTime(query.Get("to"))
	if err != nil {
		return domain.ListFilter{}, err
	}

	limit := application.DefaultListLimit
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			return domain.ListFilter{}, domain.ErrInvalidLimit
		}
		limit = parsed
	}

	return domain.ListFilter{
		UserID:     userID,
		ActivityID: strings.TrimSpace(query.Get("activity_id")),
		Status:     strings.TrimSpace(query.Get("status")),
		From:       from,
		To:         to,
		Limit:      limit,
		Cursor:     strings.TrimSpace(query.Get("cursor")),
	}, nil
}

func dueLimitFromRequest(r *http.Request) (int, error) {
	rawLimit := strings.TrimSpace(r.URL.Query().Get("limit"))
	if rawLimit == "" {
		return application.DefaultListLimit, nil
	}
	limit, err := strconv.Atoi(rawLimit)
	if err != nil {
		return 0, domain.ErrInvalidLimit
	}
	return limit, nil
}

func parseRequiredTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, domain.ErrInvalidScheduledAt
	}
	return time.Parse(time.RFC3339, value)
}

func parseOptionalTime(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := parseRequiredTime(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalTimeString(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := parseRequiredTime(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseQueryTime(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
