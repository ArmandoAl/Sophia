package http

import (
	"errors"
	"net/http"
	"strings"

	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/memory/application"
	"github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

type Handler struct {
	service      *application.Service
	requestLimit int64
}

func NewHandler(service *application.Service, requestLimit int64) *Handler {
	return &Handler{service: service, requestLimit: requestLimit}
}

func (h *Handler) Collection(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req createMemoryRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		memory, err := h.service.CreateMemory(r.Context(), userID, req.toDomain())
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, memoryResponseFromDomain(memory))
	case http.MethodGet:
		filter, err := listFilterFromRequest(r, userID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		memories, err := h.service.ListMemories(r.Context(), filter)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, memoriesListResponseFromDomain(memories))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}

	var req searchMemoryRequest
	if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
		httpjson.BadRequest(w, "invalid request body")
		return
	}
	memories, err := h.service.SearchMemory(r.Context(), req.toDomain(userID))
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, memoriesListResponseFromDomain(memories))
}

func (h *Handler) Resource(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	memoryID, action, ok := parseMemoryPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if action != "" {
		h.action(w, r, userID, memoryID, action)
		return
	}

	switch r.Method {
	case http.MethodGet:
		memory, err := h.service.GetMemory(r.Context(), userID, memoryID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, memoryResponseFromDomain(memory))
	case http.MethodPatch:
		var req updateMemoryRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		memory, err := h.service.UpdateMemory(r.Context(), userID, memoryID, req.toDomain())
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, memoryResponseFromDomain(memory))
	case http.MethodDelete:
		memory, err := h.service.DeleteMemory(r.Context(), userID, memoryID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, memoryResponseFromDomain(memory))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) action(w http.ResponseWriter, r *http.Request, userID, memoryID, action string) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}
	switch action {
	case "archive":
		memory, err := h.service.ArchiveMemory(r.Context(), userID, memoryID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, memoryResponseFromDomain(memory))
	default:
		http.NotFound(w, r)
	}
}

func authenticatedUserID(r *http.Request) (string, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func parseMemoryPath(path string) (memoryID string, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/memory/")
	if rest == "" || rest == path {
		return "", "", false
	}

	parts := strings.Split(rest, "/")
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", true
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], true
	}
	return "", "", false
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrMemoryNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrMemoryDisabled):
		httpjson.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidType),
		errors.Is(err, domain.ErrInvalidTitle),
		errors.Is(err, domain.ErrInvalidContent),
		errors.Is(err, domain.ErrInvalidSource),
		errors.Is(err, domain.ErrInvalidVisibility),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrInvalidConfidence),
		errors.Is(err, domain.ErrInvalidImportance),
		errors.Is(err, domain.ErrInvalidLimit),
		errors.Is(err, domain.ErrInvalidSearchQuery):
		httpjson.BadRequest(w, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
