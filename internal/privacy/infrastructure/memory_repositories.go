package infrastructure

import (
	"context"
	"sort"
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/privacy/domain"
)

type InMemoryAuditLogRepository struct {
	mu   sync.Mutex
	logs map[string]*domain.AuditLog
}

type InMemoryDeleteRequestRepository struct {
	mu       sync.Mutex
	requests map[string]*domain.DeleteRequest
}

func NewInMemoryAuditLogRepository() *InMemoryAuditLogRepository {
	return &InMemoryAuditLogRepository{logs: map[string]*domain.AuditLog{}}
}

func NewInMemoryDeleteRequestRepository() *InMemoryDeleteRequestRepository {
	return &InMemoryDeleteRequestRepository{requests: map[string]*domain.DeleteRequest{}}
}

func (r *InMemoryAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID] = cloneAuditLog(log)
	return nil
}

func (r *InMemoryAuditLogRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*domain.AuditLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := []*domain.AuditLog{}
	for _, log := range r.logs {
		if log.UserID == userID {
			result = append(result, cloneAuditLog(log))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryDeleteRequestRepository) Create(ctx context.Context, request *domain.DeleteRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[request.ID] = cloneDeleteRequest(request)
	return nil
}

func (r *InMemoryDeleteRequestRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*domain.DeleteRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := []*domain.DeleteRequest{}
	for _, request := range r.requests {
		if request.UserID == userID {
			result = append(result, cloneDeleteRequest(request))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryDeleteRequestRepository) ListByStatus(ctx context.Context, status string, limit int) ([]*domain.DeleteRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := []*domain.DeleteRequest{}
	for _, request := range r.requests {
		if request.Status == status {
			result = append(result, cloneDeleteRequest(request))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryDeleteRequestRepository) Update(ctx context.Context, request *domain.DeleteRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[request.ID] = cloneDeleteRequest(request)
	return nil
}

func cloneAuditLog(log *domain.AuditLog) *domain.AuditLog {
	if log == nil {
		return nil
	}
	cp := *log
	cp.Metadata = map[string]any{}
	for key, value := range log.Metadata {
		cp.Metadata[key] = value
	}
	return &cp
}

func cloneDeleteRequest(request *domain.DeleteRequest) *domain.DeleteRequest {
	if request == nil {
		return nil
	}
	cp := *request
	return &cp
}
