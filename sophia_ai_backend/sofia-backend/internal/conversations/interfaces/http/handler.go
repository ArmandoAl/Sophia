package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/conversations/application"
	"github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
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
		var req createConversationRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		conversation, err := h.service.CreateConversation(r.Context(), userID, req.toDomain())
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, conversationResponse{Conversation: conversation})
	case http.MethodGet:
		filter, err := listFilterFromRequest(r, userID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		conversations, err := h.service.ListConversations(r.Context(), filter)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, listConversationsFromDomain(conversations, filter.Limit))
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
	conversationID, action, ok := parseConversationPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	switch action {
	case "":
		h.conversation(w, r, userID, conversationID)
	case "messages":
		h.messages(w, r, userID, conversationID)
	case "archive":
		h.archive(w, r, userID, conversationID)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) conversation(w http.ResponseWriter, r *http.Request, userID, conversationID string) {
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	conversation, err := h.service.GetConversation(r.Context(), userID, conversationID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, conversationResponse{Conversation: conversation})
}

func (h *Handler) messages(w http.ResponseWriter, r *http.Request, userID, conversationID string) {
	switch r.Method {
	case http.MethodGet:
		limit, cursor, err := limitCursorFromRequest(r)
		if err != nil {
			h.handleError(w, err)
			return
		}
		messages, err := h.service.ListMessages(r.Context(), userID, conversationID, limit, cursor)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, listMessagesFromDomain(messages, limit))
	case http.MethodPost:
		var req sendMessageRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		result, err := h.service.SendMessage(r.Context(), userID, conversationID, req.Content, req.ActiveContext)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, sendMessageFromResult(result))
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) archive(w http.ResponseWriter, r *http.Request, userID, conversationID string) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}
	conversation, err := h.service.ArchiveConversation(r.Context(), userID, conversationID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, conversationResponse{Conversation: conversation})
}

func authenticatedUserID(r *http.Request) (string, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func parseConversationPath(path string) (conversationID string, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/conversations/")
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

func listFilterFromRequest(r *http.Request, userID string) (domain.ListFilter, error) {
	limit, cursor, err := limitCursorFromRequest(r)
	if err != nil {
		return domain.ListFilter{}, err
	}
	return domain.ListFilter{
		UserID: userID,
		Status: strings.TrimSpace(r.URL.Query().Get("status")),
		Limit:  limit,
		Cursor: cursor,
	}, nil
}

func limitCursorFromRequest(r *http.Request) (int, string, error) {
	limit := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return 0, "", domain.ErrInvalidLimit
		}
		limit = parsed
	}
	return limit, strings.TrimSpace(r.URL.Query().Get("cursor")), nil
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrConversationNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidTitle),
		errors.Is(err, domain.ErrInvalidContent),
		errors.Is(err, domain.ErrInvalidRole),
		errors.Is(err, domain.ErrInvalidLimit),
		errors.Is(err, runtimedomain.ErrInvalidMessage):
		httpjson.BadRequest(w, err.Error())
	case errors.Is(err, runtimedomain.ErrUnsafeActionProposal):
		httpjson.WriteError(w, http.StatusConflict, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
