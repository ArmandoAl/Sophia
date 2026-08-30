package http

import (
	"errors"
	"net/http"

	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
	"github.com/armandoalvarado/sofia-backend/internal/users/application"
	"github.com/armandoalvarado/sofia-backend/internal/users/domain"
)

type Handler struct {
	service      *application.Service
	requestLimit int64
}

func NewHandler(service *application.Service, requestLimit int64) *Handler {
	return &Handler{service: service, requestLimit: requestLimit}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}

	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	me, err := h.service.GetMe(userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, meResponseFromDomain(me))
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httpjson.MethodNotAllowed(w)
		return
	}

	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	var req updateProfileRequest
	if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
		httpjson.BadRequest(w, "invalid request body")
		return
	}

	profile, err := h.service.UpdateProfile(userID, req.toDomain())
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, profileResponseFromDomain(profile))
}

func (h *Handler) AISettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		settings, err := h.service.GetAISettings(userID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, aiSettingsResponseFromDomain(settings))
	case http.MethodPatch:
		var req updateAISettingsRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		settings, err := h.service.UpdateAISettings(userID, req.toDomain())
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, aiSettingsResponseFromDomain(settings))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}

	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	profile, err := h.service.CompleteOnboarding(userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, profileResponseFromDomain(profile))
}

func authenticatedUserID(r *http.Request) (string, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidTimezone),
		errors.Is(err, domain.ErrInvalidLocale),
		errors.Is(err, domain.ErrInvalidAssistantName),
		errors.Is(err, domain.ErrInvalidPreferredLanguage),
		errors.Is(err, domain.ErrInvalidProactivityLevel),
		errors.Is(err, domain.ErrInvalidAutonomyLevel):
		httpjson.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrProfileNotFound),
		errors.Is(err, domain.ErrAISettingsNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
