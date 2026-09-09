package http

import (
	"errors"
	"net/http"
	"strings"

	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/ingestion/application"
	"github.com/armandoalvarado/sofia-backend/internal/ingestion/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

const maxRequestBytes = int64(900 << 10)

type Handler struct {
	service      *application.Service
	requestLimit int64
}

func NewHandler(service *application.Service, requestLimit int64) *Handler {
	return &Handler{service: service, requestLimit: min(requestLimit, maxRequestBytes)}
}

func (h *Handler) Conversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}
	var request struct {
		Conversations []domain.RawConversation `json:"conversations"`
	}
	if err := httpjson.DecodeJSON(w, r, &request, h.requestLimit); err != nil {
		httpjson.BadRequest(w, "invalid request body")
		return
	}
	batch, err := h.service.Enqueue(r.Context(), userID, request.Conversations)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusAccepted, batch)
}

func (h *Handler) Batches(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	if r.Method != http.MethodDelete {
		httpjson.MethodNotAllowed(w)
		return
	}
	batchID := strings.TrimPrefix(r.URL.Path, "/ingestion/batches/")
	if batchID == "" || strings.Contains(batchID, "/") {
		http.NotFound(w, r)
		return
	}
	count, err := h.service.RetireBatch(r.Context(), userID, batchID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, map[string]any{"batch_id": batchID, "retired_beliefs": count})
}

func authenticatedUserID(r *http.Request) (string, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	return claims.UserID, ok && claims.UserID != ""
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidBatch):
		httpjson.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrBatchNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
