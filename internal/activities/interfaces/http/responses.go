package http

import (
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/activities/domain"
)

type activityResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Type           string     `json:"type"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	StartAt        *time.Time `json:"start_at,omitempty"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CanceledAt     *time.Time `json:"canceled_at,omitempty"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty"`
	Timezone       string     `json:"timezone"`
	Tags           []string   `json:"tags"`
	Source         string     `json:"source"`
	CreatedBy      string     `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	HasReminder    bool       `json:"has_reminder"`
	ReminderAt     *time.Time `json:"reminder_at,omitempty"`
	RecurrenceRule string     `json:"recurrence_rule,omitempty"`
	AIManageable   bool       `json:"ai_manageable"`
	AINotes        string     `json:"ai_notes,omitempty"`
}

type listActivitiesResponse struct {
	Activities []activityResponse `json:"activities"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

func activityResponseFromDomain(activity *domain.Activity) activityResponse {
	return activityResponse{
		ID:             activity.ID,
		UserID:         activity.UserID,
		Type:           activity.Type,
		Title:          activity.Title,
		Description:    activity.Description,
		Status:         activity.Status,
		Priority:       activity.Priority,
		StartAt:        activity.StartAt,
		DueAt:          activity.DueAt,
		CompletedAt:    activity.CompletedAt,
		CanceledAt:     activity.CanceledAt,
		ArchivedAt:     activity.ArchivedAt,
		Timezone:       activity.Timezone,
		Tags:           append([]string(nil), activity.Tags...),
		Source:         activity.Source,
		CreatedBy:      activity.CreatedBy,
		CreatedAt:      activity.CreatedAt,
		UpdatedAt:      activity.UpdatedAt,
		HasReminder:    activity.HasReminder,
		ReminderAt:     activity.ReminderAt,
		RecurrenceRule: activity.RecurrenceRule,
		AIManageable:   activity.AIManageable,
		AINotes:        activity.AINotes,
	}
}

func listResponseFromDomain(activities []*domain.Activity) listActivitiesResponse {
	response := listActivitiesResponse{Activities: make([]activityResponse, 0, len(activities))}
	for _, activity := range activities {
		response.Activities = append(response.Activities, activityResponseFromDomain(activity))
	}
	if len(activities) > 0 {
		response.NextCursor = activities[len(activities)-1].ID
	}
	return response
}
