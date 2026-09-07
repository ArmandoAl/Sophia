package http

import (
	"net/http"

	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
	"github.com/armandoalvarado/sofia-backend/internal/privacy/application"
)

type Handler struct {
	service      *application.Service
	requestLimit int64
}

type deleteRequestBody struct {
	Reason string `json:"reason"`
}

func NewHandler(service *application.Service, requestLimit int64) *Handler {
	return &Handler{service: service, requestLimit: requestLimit}
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	export, err := h.service.ExportUserData(r.Context(), userID)
	if err != nil {
		httpjson.InternalServerError(w)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, export)
}

func (h *Handler) DeleteRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	var body deleteRequestBody
	if r.Body != nil && r.ContentLength != 0 {
		if err := httpjson.DecodeJSON(w, r, &body, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
	}
	request, err := h.service.CreateDeleteRequest(r.Context(), userID, body.Reason)
	if err != nil {
		httpjson.InternalServerError(w)
		return
	}
	httpjson.WriteJSON(w, http.StatusAccepted, request)
}

func authenticatedUserID(r *http.Request) (string, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}
