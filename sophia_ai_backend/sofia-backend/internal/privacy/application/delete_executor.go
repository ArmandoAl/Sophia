package application

import (
	"context"
	"fmt"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/database"
	"github.com/armandoalvarado/sofia-backend/internal/privacy/domain"
	"github.com/google/uuid"
)

var DefaultDeleteCollections = []string{
	"user_profiles",
	"ai_settings",
	"activities",
	"reminders",
	"mood_entries",
	"activity_outcomes",
	"reflections",
	"memories",
	"user_beliefs",
	"user_contexts",
	"entity_candidates",
	"episodes",
	"prompt_versions",
	"daily_summaries",
	"ai_action_proposals",
	database.V1ConversationsCollection,
	database.V1ConversationMessagesCollection,
}

type UserDataDeleter interface {
	DeleteUserDocuments(ctx context.Context, collection, userID string) (int, error)
}

type DeleteExecutor struct {
	requests    domain.DeleteRequestRepository
	auditLogs   domain.AuditLogRepository
	deleter     UserDataDeleter
	collections []string
	now         func() time.Time
}

type DeleteExecutorOptions struct {
	Collections []string
	Now         func() time.Time
}

type DeleteExecutionReport struct {
	GeneratedAt time.Time                    `json:"generated_at"`
	Processed   int                          `json:"processed"`
	Completed   int                          `json:"completed"`
	Failed      int                          `json:"failed"`
	Requests    []DeleteRequestProcessResult `json:"requests"`
}

type DeleteRequestProcessResult struct {
	RequestID           string         `json:"request_id"`
	UserID              string         `json:"user_id"`
	Status              string         `json:"status"`
	DeletedByCollection map[string]int `json:"deleted_by_collection"`
	Error               string         `json:"error,omitempty"`
}

func NewDeleteExecutor(requests domain.DeleteRequestRepository, auditLogs domain.AuditLogRepository, deleter UserDataDeleter, options DeleteExecutorOptions) *DeleteExecutor {
	collections := options.Collections
	if len(collections) == 0 {
		collections = DefaultDeleteCollections
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &DeleteExecutor{requests: requests, auditLogs: auditLogs, deleter: deleter, collections: collections, now: now}
}

func (e *DeleteExecutor) ProcessApproved(ctx context.Context, limit int) (*DeleteExecutionReport, error) {
	if limit <= 0 {
		limit = 20
	}
	requests, err := e.requests.ListByStatus(ctx, domain.DeleteRequestStatusApproved, limit)
	if err != nil {
		return nil, err
	}

	report := &DeleteExecutionReport{GeneratedAt: e.now(), Requests: []DeleteRequestProcessResult{}}
	for _, request := range requests {
		report.Processed++
		result := e.processOne(ctx, request)
		report.Requests = append(report.Requests, result)
		if result.Status == domain.DeleteRequestStatusCompleted {
			report.Completed++
		}
		if result.Status == domain.DeleteRequestStatusFailed {
			report.Failed++
		}
	}
	return report, nil
}

func (e *DeleteExecutor) processOne(ctx context.Context, request *domain.DeleteRequest) DeleteRequestProcessResult {
	result := DeleteRequestProcessResult{
		RequestID:           request.ID,
		UserID:              request.UserID,
		Status:              request.Status,
		DeletedByCollection: map[string]int{},
	}
	if request.Status != domain.DeleteRequestStatusApproved {
		return result
	}

	now := e.now()
	request.Status = domain.DeleteRequestStatusProcessing
	request.UpdatedAt = now
	if err := e.requests.Update(ctx, request); err != nil {
		result.Status = domain.DeleteRequestStatusFailed
		result.Error = err.Error()
		return result
	}

	for _, collection := range e.collections {
		count, err := e.deleter.DeleteUserDocuments(ctx, collection, request.UserID)
		if err != nil {
			e.markFailed(ctx, request, &result, fmt.Sprintf("delete %s: %v", collection, err))
			return result
		}
		result.DeletedByCollection[collection] = count
	}

	request.Status = domain.DeleteRequestStatusCompleted
	request.UpdatedAt = e.now()
	if err := e.requests.Update(ctx, request); err != nil {
		result.Status = domain.DeleteRequestStatusFailed
		result.Error = err.Error()
		return result
	}
	result.Status = request.Status
	_ = e.createAudit(ctx, request.UserID, domain.AuditActionDeleteProcessed, request.ID, map[string]any{
		"delete_request_id": request.ID,
		"collections":       result.DeletedByCollection,
	})
	return result
}

func (e *DeleteExecutor) markFailed(ctx context.Context, request *domain.DeleteRequest, result *DeleteRequestProcessResult, reason string) {
	request.Status = domain.DeleteRequestStatusFailed
	request.UpdatedAt = e.now()
	_ = e.requests.Update(ctx, request)
	result.Status = request.Status
	result.Error = reason
	_ = e.createAudit(ctx, request.UserID, domain.AuditActionDeleteFailed, request.ID, map[string]any{
		"delete_request_id": request.ID,
		"reason":            reason,
	})
}

func (e *DeleteExecutor) createAudit(ctx context.Context, userID, action, requestID string, metadata map[string]any) error {
	if e.auditLogs == nil {
		return nil
	}
	return e.auditLogs.Create(ctx, &domain.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       action,
		ResourceType: "delete_request",
		ResourceID:   requestID,
		CreatedAt:    e.now(),
		Metadata:     redactMetadata(metadata),
	})
}
