package application

import (
	"context"
	"testing"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/privacy/domain"
	"github.com/armandoalvarado/sofia-backend/internal/privacy/infrastructure"
)

type fakeUserDataDeleter struct {
	docs map[string]map[string]int
}

func (d *fakeUserDataDeleter) DeleteUserDocuments(ctx context.Context, collection, userID string) (int, error) {
	if d.docs[collection] == nil {
		return 0, nil
	}
	count := d.docs[collection][userID]
	delete(d.docs[collection], userID)
	return count, nil
}

func TestDeleteExecutorDoesNotProcessPendingRequests(t *testing.T) {
	requestRepo := infrastructure.NewInMemoryDeleteRequestRepository()
	auditRepo := infrastructure.NewInMemoryAuditLogRepository()
	deleter := &fakeUserDataDeleter{docs: map[string]map[string]int{"activities": {"user-1": 2}}}
	mustCreateDeleteRequest(t, requestRepo, "request-1", "user-1", domain.DeleteRequestStatusPending)

	report, err := NewDeleteExecutor(requestRepo, auditRepo, deleter, DeleteExecutorOptions{Collections: []string{"activities"}}).ProcessApproved(context.Background(), 10)
	if err != nil {
		t.Fatalf("ProcessApproved returned error: %v", err)
	}

	if report.Processed != 0 || deleter.docs["activities"]["user-1"] != 2 {
		t.Fatalf("expected pending request to be ignored, report=%+v docs=%+v", report, deleter.docs)
	}
}

func TestDeleteExecutorProcessesApprovedRequest(t *testing.T) {
	requestRepo := infrastructure.NewInMemoryDeleteRequestRepository()
	auditRepo := infrastructure.NewInMemoryAuditLogRepository()
	deleter := &fakeUserDataDeleter{docs: map[string]map[string]int{
		"activities": {"user-1": 2},
		"memories":   {"user-1": 1},
	}}
	mustCreateDeleteRequest(t, requestRepo, "request-1", "user-1", domain.DeleteRequestStatusApproved)

	report, err := NewDeleteExecutor(requestRepo, auditRepo, deleter, DeleteExecutorOptions{Collections: []string{"activities", "memories"}}).ProcessApproved(context.Background(), 10)
	if err != nil {
		t.Fatalf("ProcessApproved returned error: %v", err)
	}

	if report.Completed != 1 || report.Requests[0].DeletedByCollection["activities"] != 2 || report.Requests[0].DeletedByCollection["memories"] != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
	requests, _ := requestRepo.ListByUser(context.Background(), "user-1", 10)
	if requests[0].Status != domain.DeleteRequestStatusCompleted {
		t.Fatalf("expected completed request, got %q", requests[0].Status)
	}
}

func TestDeleteExecutorCreatesAuditLog(t *testing.T) {
	requestRepo := infrastructure.NewInMemoryDeleteRequestRepository()
	auditRepo := infrastructure.NewInMemoryAuditLogRepository()
	deleter := &fakeUserDataDeleter{docs: map[string]map[string]int{"activities": {"user-1": 1}}}
	mustCreateDeleteRequest(t, requestRepo, "request-1", "user-1", domain.DeleteRequestStatusApproved)

	if _, err := NewDeleteExecutor(requestRepo, auditRepo, deleter, DeleteExecutorOptions{Collections: []string{"activities"}}).ProcessApproved(context.Background(), 10); err != nil {
		t.Fatalf("ProcessApproved returned error: %v", err)
	}

	logs, _ := auditRepo.ListByUser(context.Background(), "user-1", 10)
	if len(logs) != 1 || logs[0].Action != domain.AuditActionDeleteProcessed {
		t.Fatalf("expected delete processed audit log, got %+v", logs)
	}
}

func TestDeleteExecutorDoesNotDeleteOtherUsersData(t *testing.T) {
	requestRepo := infrastructure.NewInMemoryDeleteRequestRepository()
	auditRepo := infrastructure.NewInMemoryAuditLogRepository()
	deleter := &fakeUserDataDeleter{docs: map[string]map[string]int{
		"activities": {"user-1": 2, "user-2": 3},
	}}
	mustCreateDeleteRequest(t, requestRepo, "request-1", "user-1", domain.DeleteRequestStatusApproved)

	if _, err := NewDeleteExecutor(requestRepo, auditRepo, deleter, DeleteExecutorOptions{Collections: []string{"activities"}}).ProcessApproved(context.Background(), 10); err != nil {
		t.Fatalf("ProcessApproved returned error: %v", err)
	}

	if _, ok := deleter.docs["activities"]["user-1"]; ok {
		t.Fatal("expected user-1 data to be deleted")
	}
	if deleter.docs["activities"]["user-2"] != 3 {
		t.Fatalf("expected user-2 data to remain, got %+v", deleter.docs)
	}
}

func TestDefaultDeleteCollectionsIncludeV1AndExcludeLegacy(t *testing.T) {
	for _, collection := range DefaultDeleteCollections {
		if collection == "conversations" || collection == "conversation_messages" {
			t.Fatalf("legacy conversation collection must not be deleted: %q", collection)
		}
	}
	want := map[string]bool{"v1_conversations": true, "v1_conversation_messages": true}
	for _, collection := range DefaultDeleteCollections {
		delete(want, collection)
	}
	if len(want) != 0 {
		t.Fatalf("missing V1 conversation delete collections: %v", want)
	}
}

func mustCreateDeleteRequest(t *testing.T, repo *infrastructure.InMemoryDeleteRequestRepository, id, userID, status string) {
	t.Helper()
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	if err := repo.Create(context.Background(), &domain.DeleteRequest{
		ID:        id,
		UserID:    userID,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}
