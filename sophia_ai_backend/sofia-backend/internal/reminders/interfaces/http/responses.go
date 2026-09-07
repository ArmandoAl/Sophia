package http

import (
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
)

type reminderResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	ActivityID     string     `json:"activity_id,omitempty"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	ScheduledAt    time.Time  `json:"scheduled_at"`
	Timezone       string     `json:"timezone"`
	RecurrenceRule string     `json:"recurrence_rule,omitempty"`
	NextRunAt      time.Time  `json:"next_run_at"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	FailedAt       *time.Time `json:"failed_at,omitempty"`
	FailureReason  string     `json:"failure_reason,omitempty"`
	RetryCount     int        `json:"retry_count"`
	MaxRetries     int        `json:"max_retries"`
	Source         string     `json:"source"`
	CreatedBy      string     `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type remindersListResponse struct {
	Reminders  []reminderResponse `json:"reminders"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

func reminderResponseFromDomain(reminder *domain.Reminder) reminderResponse {
	return reminderResponse{
		ID:             reminder.ID,
		UserID:         reminder.UserID,
		ActivityID:     reminder.ActivityID,
		Title:          reminder.Title,
		Description:    reminder.Description,
		Status:         reminder.Status,
		ScheduledAt:    reminder.ScheduledAt,
		Timezone:       reminder.Timezone,
		RecurrenceRule: reminder.RecurrenceRule,
		NextRunAt:      reminder.NextRunAt,
		LastRunAt:      reminder.LastRunAt,
		DeliveredAt:    reminder.DeliveredAt,
		FailedAt:       reminder.FailedAt,
		FailureReason:  reminder.FailureReason,
		RetryCount:     reminder.RetryCount,
		MaxRetries:     reminder.MaxRetries,
		Source:         reminder.Source,
		CreatedBy:      reminder.CreatedBy,
		CreatedAt:      reminder.CreatedAt,
		UpdatedAt:      reminder.UpdatedAt,
	}
}

func listResponseFromDomain(reminders []*domain.Reminder) remindersListResponse {
	response := remindersListResponse{Reminders: make([]reminderResponse, 0, len(reminders))}
	for _, reminder := range reminders {
		response.Reminders = append(response.Reminders, reminderResponseFromDomain(reminder))
	}
	if len(reminders) > 0 {
		response.NextCursor = reminders[len(reminders)-1].ID
	}
	return response
}
