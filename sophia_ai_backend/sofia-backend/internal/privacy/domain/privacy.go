package domain

import (
	"context"
	"time"
)

const (
	DeleteRequestStatusPending    = "pending"
	DeleteRequestStatusApproved   = "approved"
	DeleteRequestStatusProcessing = "processing"
	DeleteRequestStatusCompleted  = "completed"
	DeleteRequestStatusFailed     = "failed"
	DeleteRequestStatusRejected   = "rejected"

	AuditActionExportRequested = "user_data_export_requested"
	AuditActionDeleteRequested = "user_delete_requested"
	AuditActionDeleteProcessed = "user_delete_processed"
	AuditActionDeleteFailed    = "user_delete_failed"
)

type AuditLog struct {
	ID           string         `json:"id"`
	UserID       string         `json:"user_id"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	CreatedAt    time.Time      `json:"created_at"`
	Metadata     map[string]any `json:"metadata"`
}

type DeleteRequest struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	ListByUser(ctx context.Context, userID string, limit int) ([]*AuditLog, error)
}

type DeleteRequestRepository interface {
	Create(ctx context.Context, request *DeleteRequest) error
	ListByUser(ctx context.Context, userID string, limit int) ([]*DeleteRequest, error)
	ListByStatus(ctx context.Context, status string, limit int) ([]*DeleteRequest, error)
	Update(ctx context.Context, request *DeleteRequest) error
}
