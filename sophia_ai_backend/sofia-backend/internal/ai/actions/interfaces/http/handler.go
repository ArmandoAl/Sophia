package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/ai/actions/application"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

type Handler struct {
	service      *application.Service
	requestLimit int64
}

type createProposalRequest struct {
	ToolName      string          `json:"tool_name"`
	ProposedInput json.RawMessage `json:"proposed_input"`
	Reason        string          `json:"reason"`
	RiskLevel     string          `json:"risk_level"`
	ExpiresAt     string          `json:"expires_at"`
}

type confirmProposalRequest struct {
	CorrectedInput    json.RawMessage `json:"corrected_input"`
	DecisionLatencyMS *int64          `json:"decision_latency_ms"`
}

type rejectProposalRequest struct {
	RejectionReason   string `json:"rejection_reason"`
	DecisionLatencyMS *int64 `json:"decision_latency_ms"`
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
		var req createProposalRequest
		if err := httpjson.DecodeJSON(w, r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		input, err := req.toDomain()
		if err != nil {
			httpjson.BadRequest(w, "invalid expires_at")
			return
		}
		proposal, err := h.service.CreateActionProposal(r.Context(), userID, input)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, proposal)
	case http.MethodGet:
		limit := application.DefaultListLimit
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				httpjson.BadRequest(w, "invalid limit")
				return
			}
			limit = parsed
		}
		proposals, err := h.service.ListActionProposals(r.Context(), actionsdomain.ListFilter{UserID: userID, Status: strings.TrimSpace(r.URL.Query().Get("status")), Limit: limit})
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]any{"action_proposals": proposals})
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
	id, action, ok := parseProposalPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if action == "explanation" {
		if r.Method != http.MethodGet {
			httpjson.MethodNotAllowed(w)
			return
		}
		explanation, err := h.service.ExplainActionProposal(r.Context(), userID, id)
		if err != nil {
			h.handleError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, explanation)
		return
	}
	if action != "" {
		h.action(w, r, userID, id, action)
		return
	}
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	proposal, err := h.service.GetActionProposal(r.Context(), userID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, proposal)
}

func (h *Handler) action(w http.ResponseWriter, r *http.Request, userID, id, action string) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}
	var (
		proposal *actionsdomain.AIActionProposal
		err      error
	)
	switch action {
	case "confirm":
		var req confirmProposalRequest
		if err = decodeOptionalJSON(r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		proposal, err = h.service.ConfirmActionProposal(r.Context(), userID, id, req.CorrectedInput, req.DecisionLatencyMS)
	case "reject":
		var req rejectProposalRequest
		if err = decodeOptionalJSON(r, &req, h.requestLimit); err != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		proposal, err = h.service.RejectActionProposal(r.Context(), userID, id, req.RejectionReason, req.DecisionLatencyMS)
	case "execute":
		proposal, err = h.service.ExecuteConfirmedActionProposal(r.Context(), userID, id)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, proposal)
}

func (r createProposalRequest) toDomain() (actionsdomain.ProposalCreate, error) {
	var expiresAt *time.Time
	if strings.TrimSpace(r.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(r.ExpiresAt))
		if err != nil {
			return actionsdomain.ProposalCreate{}, err
		}
		expiresAt = &parsed
	}
	return actionsdomain.ProposalCreate{ToolName: r.ToolName, ProposedInput: r.ProposedInput, Reason: r.Reason, RiskLevel: r.RiskLevel, ExpiresAt: expiresAt}, nil
}

func decodeOptionalJSON(r *http.Request, dst any, limitBytes int64) error {
	if r.Body == nil {
		return nil
	}
	r.Body = http.MaxBytesReader(nil, r.Body, limitBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}

func authenticatedUserID(r *http.Request) (string, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func parseProposalPath(path string) (id string, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/ai/action-proposals/")
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
	case errors.Is(err, actionsdomain.ErrProposalNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, actionsdomain.ErrToolDisabled), errors.Is(err, actionsdomain.ErrAutonomyNotAllowed):
		httpjson.WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, actionsdomain.ErrConfirmationRequired), errors.Is(err, actionsdomain.ErrInvalidTransition):
		httpjson.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, actionsdomain.ErrInvalidToolName), errors.Is(err, actionsdomain.ErrInvalidRiskLevel), errors.Is(err, actionsdomain.ErrInvalidProposedInput):
		httpjson.BadRequest(w, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
