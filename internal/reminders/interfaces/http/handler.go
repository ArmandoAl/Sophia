package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/application"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
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
		var req createReminderRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		input, err := req.toDomain()
		if err != nil {
			httpjson.BadRequest(w, "invalid scheduled_at")
			return
		}
		reminder, err := h.service.CreateReminder(r.Context(), userID, input)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, reminderResponseFromDomain(reminder))
	case http.MethodGet:
		filter, err := listFilterFromRequest(r, userID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		reminders, err := h.service.ListReminders(r.Context(), filter)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, listResponseFromDomain(reminders))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) Due(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	limit, err := dueLimitFromRequest(r)
	if err != nil {
		h.handleError(w, err)
		return
	}
	reminders, err := h.service.ListDueReminders(r.Context(), userID, time.Now(), limit)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, listResponseFromDomain(reminders))
}

func (h *Handler) Resource(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}

	reminderID, action, ok := parseReminderPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if action != "" {
		h.action(w, r, userID, reminderID, action)
		return
	}

	switch r.Method {
	case http.MethodGet:
		reminder, err := h.service.GetReminder(r.Context(), userID, reminderID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, reminderResponseFromDomain(reminder))
	case http.MethodPatch:
		var req updateReminderRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		update, err := req.toDomain()
		if err != nil {
			httpjson.BadRequest(w, "invalid scheduled_at")
			return
		}
		reminder, err := h.service.UpdateReminder(r.Context(), userID, reminderID, update)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, reminderResponseFromDomain(reminder))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) action(w http.ResponseWriter, r *http.Request, userID, reminderID, action string) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}

	var (
		reminder *domain.Reminder
		err      error
	)

	switch action {
	case "cancel":
		reminder, err = h.service.CancelReminder(r.Context(), userID, reminderID)
	case "archive":
		reminder, err = h.service.ArchiveReminder(r.Context(), userID, reminderID)
	default:
		http.NotFound(w, r)
		return
	}

	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, reminderResponseFromDomain(reminder))
}

func authenticatedUserID(r *http.Request) (string, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func parseReminderPath(path string) (reminderID string, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/reminders/")
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
	case errors.Is(err, domain.ErrReminderNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrRemindersDisabled):
		httpjson.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrActivityNotAllowed):
		httpjson.WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrInvalidTitle),
		errors.Is(err, domain.ErrInvalidScheduledAt),
		errors.Is(err, domain.ErrInvalidTimezone),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrInvalidSource),
		errors.Is(err, domain.ErrInvalidCreatedBy),
		errors.Is(err, domain.ErrInvalidLimit),
		errors.Is(err, domain.ErrInvalidDateRange):
		httpjson.BadRequest(w, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
