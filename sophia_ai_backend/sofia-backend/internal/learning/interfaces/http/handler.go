package http

import (
	"errors"
	"net/http"
	"strings"

	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/learning/application"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
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
		var request createContextRequest
		if err := httpjson.DecodeJSON(w, r, &request, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		value, err := h.service.CreateUserContext(r.Context(), userID, request.toDomain())
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, contextResponseFromDomain(value))
	case http.MethodGet:
		kind, slug := strings.TrimSpace(r.URL.Query().Get("kind")), strings.TrimSpace(r.URL.Query().Get("slug"))
		if kind != "" || slug != "" {
			if kind == "" || slug == "" {
				httpjson.BadRequest(w, "kind and slug are required together")
				return
			}
			value, err := h.service.FindUserContextBySlug(r.Context(), userID, kind, slug)
			if err != nil {
				h.handleError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, contextResponseFromDomain(value))
			return
		}
		values, err := h.service.ListUserContexts(r.Context(), userID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, contextListResponseFromDomain(values))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) Resource(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	contextID, action, ok := parseContextPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if action == "archive" && r.Method == http.MethodPost {
		value, err := h.service.ArchiveUserContext(r.Context(), userID, contextID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, contextResponseFromDomain(value))
		return
	}
	if action != "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPatch {
		httpjson.MethodNotAllowed(w)
		return
	}
	var request updateContextRequest
	if err := httpjson.DecodeJSON(w, r, &request, h.requestLimit); err != nil {
		httpjson.BadRequest(w, "invalid request body")
		return
	}
	value, err := h.service.UpdateUserContext(r.Context(), userID, contextID, request.toDomain())
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, contextResponseFromDomain(value))
}

func authenticatedUserID(r *http.Request) (string, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func parseContextPath(path string) (contextID, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/contexts/")
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
	case errors.Is(err, domain.ErrUserContextNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrUserContextExists):
		httpjson.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidContextKind), errors.Is(err, domain.ErrInvalidContextSlug), errors.Is(err, domain.ErrInvalidContextLabel):
		httpjson.BadRequest(w, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
