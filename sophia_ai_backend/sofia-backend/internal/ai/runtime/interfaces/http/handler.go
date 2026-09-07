package http

import (
	"errors"
	"net/http"

	"github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	authcontext "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

type Handler struct {
	runtime      domain.RuntimeService
	requestLimit int64
}

type messageRequest struct {
	Message string `json:"message"`
	DryRun  bool   `json:"dry_run"`
}

func NewHandler(runtime domain.RuntimeService, requestLimit int64) *Handler {
	return &Handler{runtime: runtime, requestLimit: requestLimit}
}

func (h *Handler) Message(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}
	claims, ok := authcontext.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		httpjson.Unauthorized(w)
		return
	}

	var request messageRequest
	if err := httpjson.DecodeJSON(w, r, &request, h.requestLimit); err != nil {
		httpjson.BadRequest(w, err.Error())
		return
	}

	response, err := h.runtime.HandleMessage(r.Context(), domain.RuntimeRequest{
		UserID:  claims.UserID,
		Message: request.Message,
		DryRun:  request.DryRun,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidMessage):
			httpjson.BadRequest(w, err.Error())
		case errors.Is(err, domain.ErrUnsafeActionProposal):
			httpjson.WriteError(w, http.StatusConflict, err.Error())
		default:
			httpjson.InternalServerError(w)
		}
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, response)
}
