package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	ingestiondomain "github.com/armandoalvarado/sofia-backend/internal/ingestion/domain"
	ingestioninfra "github.com/armandoalvarado/sofia-backend/internal/ingestion/infrastructure"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	learninginfra "github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

type extractModel struct{ calls int }

func (m *extractModel) Generate(ctx context.Context, request runtimedomain.ModelRequest) (runtimedomain.ModelResponse, error) {
	if request.Task != runtimedomain.TaskExtract {
		return runtimedomain.ModelResponse{}, fmt.Errorf("unexpected task %q", request.Task)
	}
	m.calls++
	statements := []string{"adora madrugar", "colecciona vinilos", "cocina los domingos"}
	return runtimedomain.ModelResponse{AssistantMessage: fmt.Sprintf(`{"beliefs":[{"statement":%q,"category":"personal","scope":"global","scope_key":"","confidence":0.9}]}`, statements[m.calls-1]), Usage: runtimedomain.Usage{InputTokens: 10, OutputTokens: 5, Model: "extract-test"}}, nil
}

func TestRunOnceProcesses120MessagesInThreeWindows(t *testing.T) {
	ctx := context.Background()
	batches := ingestioninfra.NewInMemoryBatchRepository()
	beliefs := learninginfra.NewInMemoryBeliefRepository()
	learning := learningapp.NewService(beliefs, learninginfra.NewInMemoryPromptVersionRepository(), learninginfra.NewInMemoryDailySummaryRepository())
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	messages := make([]ingestiondomain.RawMessage, 120)
	for i := range messages {
		messages[i] = ingestiondomain.RawMessage{Author: "A", Text: fmt.Sprintf("message %d", i), SentAt: now}
	}
	batch := &ingestiondomain.Batch{ID: "batch-1", UserID: "user-1", Status: ingestiondomain.StatusPending, Conversations: []ingestiondomain.RawConversation{{Source: ingestiondomain.SourceWhatsApp, ExternalID: "chat-1", Participants: []string{"A"}, Messages: messages}}, CreatedAt: now, UpdatedAt: now}
	if err := batches.Create(ctx, batch); err != nil {
		t.Fatal(err)
	}
	model := &extractModel{}
	runner := New(batches, learning, model, Options{MaxTokens: 1000, Now: func() time.Time { return now }})
	processed, err := runner.RunOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if model.calls != 3 || processed.WindowsProcessed != 3 || processed.Status != ingestiondomain.StatusDone {
		t.Fatalf("unexpected processing result: calls=%d batch=%+v", model.calls, processed)
	}
	values, err := beliefs.ListActive(ctx, "user-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 3 {
		t.Fatalf("got %d beliefs, want 3", len(values))
	}
	for _, belief := range values {
		if belief.TrustTier != learningdomain.TrustInferred || belief.Confidence > 0.5 || belief.PromptSlot != learningdomain.PromptSlotSituational || belief.BatchID != batch.ID {
			t.Fatalf("unexpected inferred belief: %+v", belief)
		}
	}
}
