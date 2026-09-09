package application

import (
	"context"
	"testing"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/ingestion/domain"
	ingestioninfra "github.com/armandoalvarado/sofia-backend/internal/ingestion/infrastructure"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	learninginfra "github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

func TestEnqueueIsIdempotentAndRetireIsBatchScoped(t *testing.T) {
	ctx := context.Background()
	batches := ingestioninfra.NewInMemoryBatchRepository()
	beliefs := learninginfra.NewInMemoryBeliefRepository()
	learning := learningapp.NewService(beliefs, learninginfra.NewInMemoryPromptVersionRepository(), learninginfra.NewInMemoryDailySummaryRepository())
	service := NewService(batches, learning)
	conversation := domain.RawConversation{Source: domain.SourceManual, ExternalID: "external-1", Participants: []string{"A"}, Messages: []domain.RawMessage{{Author: "A", Text: "text", SentAt: time.Now()}}}
	first, err := service.Enqueue(ctx, "user-1", []domain.RawConversation{conversation})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Enqueue(ctx, "user-1", []domain.RawConversation{conversation})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("duplicate external id created batches %q and %q", first.ID, second.ID)
	}
	imported, err := learning.UpsertBeliefWithTrust(ctx, "user-1", "likes mornings", learningdomain.CategoryPersonal, learningdomain.ScopeGlobal, "", learningdomain.TrustInferred, 0.5, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	other, err := learning.UpsertBeliefWithTrust(ctx, "user-1", "likes evenings", learningdomain.CategoryPersonal, learningdomain.ScopeGlobal, "", learningdomain.TrustInferred, 0.5, "other-batch")
	if err != nil {
		t.Fatal(err)
	}
	count, err := service.RetireBatch(ctx, "user-1", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("retired %d beliefs, want 1", count)
	}
	gotImported, _ := beliefs.FindByID(ctx, "user-1", imported.ID)
	gotOther, _ := beliefs.FindByID(ctx, "user-1", other.ID)
	if gotImported.Status != learningdomain.StatusRetired || gotOther.Status != learningdomain.StatusActive {
		t.Fatalf("retirement crossed batch boundary: imported=%s other=%s", gotImported.Status, gotOther.Status)
	}
}
