package http

import (
	"context"
	"encoding/json"
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

func TestEntityEndpointsAreUserScopedAndArchiveInsteadOfDelete(t *testing.T) {
	service := application.NewService(infrastructure.NewInMemoryBeliefRepository(), infrastructure.NewInMemoryPromptVersionRepository(), infrastructure.NewInMemoryDailySummaryRepository())
	service.SetContextRepository(infrastructure.NewInMemoryUserContextRepository())
	handler := NewHandler(service, 1<<20)

	request := httptest.NewRequest(http.MethodPost, "/entities", strings.NewReader(`{"kind":"person","slug":"diana","label":"Diana"}`))
	request = request.WithContext(authctx.WithClaims(request.Context(), authctx.Claims{UserID: "user-1"}))
	response := httptest.NewRecorder()
	handler.Entities(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", response.Code, response.Body.String())
	}
	var created contextResponse
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	request = httptest.NewRequest(http.MethodGet, "/entities?kind=person&status=active", nil)
	request = request.WithContext(authctx.WithClaims(request.Context(), authctx.Claims{UserID: "user-2"}))
	response = httptest.NewRecorder()
	handler.Entities(response, request)
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), created.ID) {
		t.Fatalf("entity leaked across users: status=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/entities/"+created.ID, nil)
	request = request.WithContext(authctx.WithClaims(request.Context(), authctx.Claims{UserID: "user-1"}))
	response = httptest.NewRecorder()
	handler.EntityResource(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", response.Code, response.Body.String())
	}
	values, err := service.ListEntities(context.Background(), "user-1", "", domain.ContextStatusArchived)
	if err != nil || len(values) != 1 || values[0].ID != created.ID {
		t.Fatalf("DELETE did not archive entity: %+v err=%v", values, err)
	}
	request = httptest.NewRequest(http.MethodGet, "/entities/"+created.ID, nil)
	request = request.WithContext(authctx.WithClaims(request.Context(), authctx.Claims{UserID: "user-1"}))
	response = httptest.NewRecorder()
	handler.EntityResource(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("archived entity was deleted: status=%d body=%s", response.Code, response.Body.String())
	}
}
