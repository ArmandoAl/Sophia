package server

import (
	"net/http"

	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

type HealthHandler struct {
	Environment     string
	FirestoreStatus string
}

type HealthResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Firestore   string `json:"firestore"`
}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}

	status := "ok"
	if h.FirestoreStatus == "error" {
		status = "degraded"
	}

	httpjson.WriteJSON(w, http.StatusOK, HealthResponse{
		Status:      status,
		Environment: h.Environment,
		Firestore:   h.FirestoreStatus,
	})
}
