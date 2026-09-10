package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/learning/application"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

func TestBeliefPatchRejectsUnknownField(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewInMemoryBeliefRepository()
	service := application.NewService(repo, infrastructure.NewInMemoryPromptVersionRepository(), infrastructure.NewInMemoryDailySummaryRepository())
	belief, err := service.UpsertBelief(ctx, "user-1", "texto", domain.CategoryPersonal)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(service, 1<<20)
	request := httptest.NewRequest(http.MethodPatch, "/learning/beliefs/"+belief.ID, strings.NewReader(`{"statement":"nuevo","confidence":1}`))
	request = request.WithContext(authctx.WithClaims(request.Context(), authctx.Claims{UserID: "user-1"}))
	response := httptest.NewRecorder()
	handler.BeliefResource(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}
