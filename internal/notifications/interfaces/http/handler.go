package http

import (
	"errors"
	"net/http"
	"strings"

	authcontext "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/notifications/application"
	"github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

type Handler struct {
	service      *application.Service
	requestLimit int64
}

type registerDeviceTokenRequest struct {
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

type deviceTokenResponse struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	Platform      string `json:"platform"`
	TokenRedacted string `json:"token_redacted"`
	Enabled       bool   `json:"enabled"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	LastSeenAt    string `json:"last_seen_at"`
}

func NewHandler(service *application.Service, requestLimit int64) *Handler {
	return &Handler{service: service, requestLimit: requestLimit}
}

func (h *Handler) Collection(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPost:
		var request registerDeviceTokenRequest
		if err := httpjson.DecodeJSON(w, r, &request, h.requestLimit); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		token, err := h.service.RegisterDeviceToken(r.Context(), userID, domain.DeviceTokenCreate{Platform: request.Platform, Token: request.Token})
		if err != nil {
			writeDomainError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, toResponse(token))
	case http.MethodGet:
		tokens, err := h.service.ListDeviceTokens(r.Context(), userID)
		if err != nil {
			httpjson.InternalServerError(w)
			return
		}
		response := make([]deviceTokenResponse, 0, len(tokens))
		for _, token := range tokens {
			response = append(response, toResponse(token))
		}
		httpjson.WriteJSON(w, http.StatusOK, response)
	default:
		httpjson.MethodNotAllowed(w)
	}
}

func (h *Handler) Resource(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodDelete {
		httpjson.MethodNotAllowed(w)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/notifications/device-tokens/")
	if strings.TrimSpace(id) == "" {
		httpjson.BadRequest(w, "device token id is required")
		return
	}
	if err := h.service.DeleteDeviceToken(r.Context(), userID, id); err != nil {
		writeDomainError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func authenticatedUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	claims, ok := authcontext.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		httpjson.Unauthorized(w)
		return "", false
	}
	return claims.UserID, true
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidPlatform), errors.Is(err, domain.ErrInvalidToken):
		httpjson.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrDeviceTokenNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}

func toResponse(token *domain.DeviceToken) deviceTokenResponse {
	return deviceTokenResponse{
		ID:            token.ID,
		UserID:        token.UserID,
		Platform:      token.Platform,
		TokenRedacted: RedactToken(token.Token),
		Enabled:       token.Enabled,
		CreatedAt:     token.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     token.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastSeenAt:    token.LastSeenAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func RedactToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 8 {
		return "[REDACTED]"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
