package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/activities/application"
	"github.com/armandoalvarado/sofia-backend/internal/activities/domain"
)

type createActivityRequest struct {
	Type           string   `json:"type"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"`
	StartAt        *string  `json:"start_at"`
	DueAt          *string  `json:"due_at"`
	Timezone       string   `json:"timezone"`
	Tags           []string `json:"tags"`
	HasReminder    bool     `json:"has_reminder"`
	ReminderAt     *string  `json:"reminder_at"`
	RecurrenceRule string   `json:"recurrence_rule"`
	AIManageable   bool     `json:"ai_manageable"`
	AINotes        string   `json:"ai_notes"`
}

func (r createActivityRequest) toDomain() (domain.ActivityCreate, error) {
	startAt, err := parseOptionalTime(r.StartAt)
	if err != nil {
		return domain.ActivityCreate{}, err
	}
	dueAt, err := parseOptionalTime(r.DueAt)
	if err != nil {
		return domain.ActivityCreate{}, err
	}
	reminderAt, err := parseOptionalTime(r.ReminderAt)
	if err != nil {
		return domain.ActivityCreate{}, err
	}

	return domain.ActivityCreate{
		Type:           r.Type,
		Title:          r.Title,
		Description:    r.Description,
		Priority:       r.Priority,
		StartAt:        startAt,
		DueAt:          dueAt,
		Timezone:       r.Timezone,
		Tags:           r.Tags,
		HasReminder:    r.HasReminder,
		ReminderAt:     reminderAt,
		RecurrenceRule: r.RecurrenceRule,
		AIManageable:   r.AIManageable,
		AINotes:        r.AINotes,
	}, nil
}

type updateActivityRequest struct {
	Type           *string   `json:"type"`
	Title          *string   `json:"title"`
	Description    *string   `json:"description"`
	Status         *string   `json:"status"`
	Priority       *string   `json:"priority"`
	StartAt        *string   `json:"start_at"`
	DueAt          *string   `json:"due_at"`
	Timezone       *string   `json:"timezone"`
	Tags           *[]string `json:"tags"`
	HasReminder    *bool     `json:"has_reminder"`
	ReminderAt     *string   `json:"reminder_at"`
	RecurrenceRule *string   `json:"recurrence_rule"`
	AIManageable   *bool     `json:"ai_manageable"`
	AINotes        *string   `json:"ai_notes"`
}

func (r *updateActivityRequest) toDomain() (domain.ActivityUpdate, error) {
	startAt, err := parseOptionalTimePointer(r.StartAt)
	if err != nil {
		return domain.ActivityUpdate{}, err
	}
	dueAt, err := parseOptionalTimePointer(r.DueAt)
	if err != nil {
		return domain.ActivityUpdate{}, err
	}
	reminderAt, err := parseOptionalTimePointer(r.ReminderAt)
	if err != nil {
		return domain.ActivityUpdate{}, err
	}

	return domain.ActivityUpdate{
		Type:           r.Type,
		Title:          r.Title,
		Description:    r.Description,
		Status:         r.Status,
		Priority:       r.Priority,
		StartAt:        startAt,
		DueAt:          dueAt,
		Timezone:       r.Timezone,
		Tags:           r.Tags,
		HasReminder:    r.HasReminder,
		ReminderAt:     reminderAt,
		RecurrenceRule: r.RecurrenceRule,
		AIManageable:   r.AIManageable,
		AINotes:        r.AINotes,
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
		UserID:   userID,
		Status:   strings.TrimSpace(query.Get("status")),
		Type:     strings.TrimSpace(query.Get("type")),
		Priority: strings.TrimSpace(query.Get("priority")),
		From:     from,
		To:       to,
		Tag:      strings.ToLower(strings.TrimSpace(query.Get("tag"))),
		Limit:    limit,
		Cursor:   strings.TrimSpace(query.Get("cursor")),
	}, nil
}

func parseOptionalTime(value *string) (*time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalTimePointer(value *string) (**time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := parseOptionalTime(value)
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
