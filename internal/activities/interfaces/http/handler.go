package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/armandoalvarado/sofia-backend/internal/activities/application"
	"github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
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
		var req createActivityRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		input, err := req.toDomain()
		if err != nil {
			httpjson.BadRequest(w, "invalid date")
			return
		}
		activity, err := h.service.CreateActivity(r.Context(), userID, input)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, activityResponseFromDomain(activity))
	case http.MethodGet:
		filter, err := listFilterFromRequest(r, userID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		activities, err := h.service.ListActivities(r.Context(), filter)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, listResponseFromDomain(activities))
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

	activityID, action, ok := parseActivityPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if action != "" {
		h.action(w, r, userID, activityID, action)
		return
	}

	switch r.Method {
	case http.MethodGet:
		activity, err := h.service.GetActivity(r.Context(), userID, activityID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		if activity == nil {
			h.handleError(w, domain.ErrActivityNotFound)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, activityResponseFromDomain(activity))
	case http.MethodPatch:
		var req updateActivityRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		update, err := req.toDomain()
		if err != nil {
			httpjson.BadRequest(w, "invalid date")
			return
		}
		activity, err := h.service.UpdateActivity(r.Context(), userID, activityID, update)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, activityResponseFromDomain(activity))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) action(w http.ResponseWriter, r *http.Request, userID, activityID, action string) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}

	var (
		activity *domain.Activity
		err      error
	)

	switch action {
	case "complete":
		activity, err = h.service.CompleteActivity(r.Context(), userID, activityID)
	case "cancel":
		activity, err = h.service.CancelActivity(r.Context(), userID, activityID)
	case "archive":
		activity, err = h.service.ArchiveActivity(r.Context(), userID, activityID)
	default:
		http.NotFound(w, r)
		return
	}

	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, activityResponseFromDomain(activity))
}

func authenticatedUserID(r *http.Request) (string, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func parseActivityPath(path string) (activityID string, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/activities/")
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
	case errors.Is(err, domain.ErrActivityNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, remindersdomain.ErrRemindersDisabled):
		httpjson.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidTitle),
		errors.Is(err, domain.ErrInvalidType),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrInvalidPriority),
		errors.Is(err, domain.ErrInvalidTimezone),
		errors.Is(err, domain.ErrInvalidSource),
		errors.Is(err, domain.ErrInvalidCreatedBy),
		errors.Is(err, domain.ErrInvalidReminder),
		errors.Is(err, domain.ErrInvalidLimit),
		errors.Is(err, domain.ErrInvalidDateRange):
		httpjson.BadRequest(w, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
