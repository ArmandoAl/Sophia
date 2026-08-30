package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
	"github.com/armandoalvarado/sofia-backend/internal/tools/application"
	"github.com/armandoalvarado/sofia-backend/internal/tools/domain"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Collection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	tools, err := h.service.ListTools(r.Context())
	if err != nil {
		httpjson.InternalServerError(w)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, map[string]any{"tools": tools})
}

func (h *Handler) Resource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/tools/")
	if name == "" || name == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	tool, err := h.service.GetTool(r.Context(), name)
	if err != nil {
		if errors.Is(err, domain.ErrToolNotFound) {
			httpjson.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpjson.InternalServerError(w)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, tool)
}
