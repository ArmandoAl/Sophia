package domain

import (
	"context"
	"strings"
	"time"
)

const (
	DeliveryProviderFCM  = "fcm"
	DeliveryProviderNoop = "noop"

	DeliveryStatusDelivered = "delivered"
	DeliveryStatusFailed    = "failed"
	DeliveryStatusSkipped   = "skipped"
)

type DeliveryAttempt struct {
	ID         string
	UserID     string
	ReminderID string
	Provider   string
	Status     string
	DryRun     bool
	ErrorType  string
	CreatedAt  time.Time
}

type DeliveryAttemptRepository interface {
	Create(ctx context.Context, attempt *DeliveryAttempt) error
}

func NewDeliveryAttempt(id, userID, reminderID, provider, status string, dryRun bool, errorType string) *DeliveryAttempt {
	return &DeliveryAttempt{
		ID:         strings.TrimSpace(id),
		UserID:     strings.TrimSpace(userID),
		ReminderID: strings.TrimSpace(reminderID),
		Provider:   strings.TrimSpace(provider),
		Status:     strings.TrimSpace(status),
		DryRun:     dryRun,
		ErrorType:  strings.TrimSpace(errorType),
		CreatedAt:  time.Now(),
	}
}
