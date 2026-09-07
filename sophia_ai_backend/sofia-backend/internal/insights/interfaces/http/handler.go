package http

import (
	"errors"
	"net/http"
	"strings"

	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/insights/application"
	"github.com/armandoalvarado/sofia-backend/internal/insights/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

type Handler struct {
	service      *application.Service
	requestLimit int64
}

func NewHandler(service *application.Service, requestLimit int64) *Handler {
	return &Handler{service: service, requestLimit: requestLimit}
}

func (h *Handler) Mood(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req createMoodEntryRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		input, err := req.toDomain()
		if err != nil {
			httpjson.BadRequest(w, "invalid occurred_at")
			return
		}
		entry, err := h.service.CreateMoodEntry(r.Context(), userID, input)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, moodEntryResponseFromDomain(entry))
	case http.MethodGet:
		filter, err := listFilterFromRequest(r, userID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		entries, err := h.service.ListMoodEntries(r.Context(), filter)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, moodEntriesListResponseFromDomain(entries))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) Outcomes(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req createActivityOutcomeRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		outcome, err := h.service.CreateActivityOutcome(r.Context(), userID, req.toDomain())
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, activityOutcomeResponseFromDomain(outcome))
	case http.MethodGet:
		filter, err := listFilterFromRequest(r, userID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		outcomes, err := h.service.ListActivityOutcomes(r.Context(), filter)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, activityOutcomesListResponseFromDomain(outcomes))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) Reflections(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req createReflectionRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		input, err := req.toDomain()
		if err != nil {
			httpjson.BadRequest(w, "invalid occurred_at")
			return
		}
		reflection, err := h.service.CreateReflection(r.Context(), userID, input)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, reflectionResponseFromDomain(reflection))
	case http.MethodGet:
		filter, err := listFilterFromRequest(r, userID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		reflections, err := h.service.ListReflections(r.Context(), filter)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, reflectionsListResponseFromDomain(reflections))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}

	filter, err := listFilterFromRequest(r, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	summary, err := h.service.GetInsightsSummary(r.Context(), filter)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, summaryResponseFromDomain(summary))
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
	case errors.Is(err, domain.ErrActivityNotAllowed):
		httpjson.WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrInvalidMood),
		errors.Is(err, domain.ErrInvalidResult),
		errors.Is(err, domain.ErrInvalidLevel),
		errors.Is(err, domain.ErrInvalidOccurredAt),
		errors.Is(err, domain.ErrInvalidTimezone),
		errors.Is(err, domain.ErrInvalidContent),
		errors.Is(err, domain.ErrInvalidActivityID),
		errors.Is(err, domain.ErrInvalidLimit),
		errors.Is(err, domain.ErrInvalidDateRange):
		httpjson.BadRequest(w, err.Error())
	default:
		if strings.TrimSpace(err.Error()) == "" {
			httpjson.InternalServerError(w)
			return
		}
		httpjson.InternalServerError(w)
	}
}
