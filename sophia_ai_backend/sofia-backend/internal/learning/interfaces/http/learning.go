package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/learning/application"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

type beliefResponse struct {
	ID                 string    `json:"id"`
	Statement          string    `json:"statement"`
	Category           string    `json:"category"`
	Scope              string    `json:"scope"`
	ScopeKey           string    `json:"scope_key"`
	Confidence         float64   `json:"confidence"`
	DecayedConfidence  float64   `json:"decayed_confidence"`
	EvidenceCount      int       `json:"evidence_count"`
	ContradictionCount int       `json:"contradiction_count"`
	TrustTier          int       `json:"trust_tier"`
	PromptSlot         string    `json:"prompt_slot"`
	Status             string    `json:"status"`
	FirstObservedAt    time.Time `json:"first_observed_at"`
	LastReinforcedAt   time.Time `json:"last_reinforced_at"`
}

type summaryResponse struct {
	ID                  string              `json:"id"`
	Date                string              `json:"date"`
	GeneratedAt         time.Time           `json:"generated_at"`
	Observations        []string            `json:"observations"`
	DeltaVsPrevious     map[string][]string `json:"delta_vs_previous"`
	Stats               map[string]any      `json:"stats"`
	SynthesisCost       map[string]any      `json:"synthesis_cost"`
	PromptVersionBefore string              `json:"prompt_version_before"`
	PromptVersionAfter  string              `json:"prompt_version_after"`
}

func (h *Handler) Beliefs(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	filter, err := beliefFilter(r)
	if err != nil {
		httpjson.BadRequest(w, err.Error())
		return
	}
	page, err := h.service.ListBeliefs(r.Context(), userID, filter)
	if err != nil {
		httpjson.InternalServerError(w)
		return
	}
	items := make([]beliefResponse, 0, len(page.Beliefs))
	now := time.Now().UTC()
	for _, belief := range page.Beliefs {
		items = append(items, beliefToResponse(belief, now))
	}
	httpjson.WriteJSON(w, http.StatusOK, map[string]any{"beliefs": items, "next_cursor": page.NextCursor})
}

func (h *Handler) BeliefResource(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/learning/beliefs/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	var err error
	switch r.Method {
	case http.MethodDelete:
		_, err = h.service.RetireBelief(r.Context(), userID, id)
		if err == nil {
			w.WriteHeader(http.StatusNoContent)
		}
	case http.MethodPatch:
		var request struct {
			Statement string `json:"statement"`
		}
		if decodeErr := httpjson.DecodeJSON(w, r, &request, h.requestLimit); decodeErr != nil {
			httpjson.BadRequest(w, "invalid request body")
			return
		}
		var belief *domain.Belief
		belief, err = h.service.UpdateBeliefStatement(r.Context(), userID, id, request.Statement)
		if err == nil {
			httpjson.WriteJSON(w, http.StatusOK, beliefToResponse(belief, time.Now().UTC()))
		}
	default:
		httpjson.MethodNotAllowed(w)
		return
	}
	if err != nil {
		h.handleLearningError(w, err)
	}
}

func (h *Handler) PromptVersion(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	version, err := h.service.GetActivePromptVersion(r.Context(), userID)
	if err != nil {
		h.handleLearningError(w, err)
		return
	}
	if version == nil {
		h.handleLearningError(w, domain.ErrPromptVersionNotFound)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, map[string]any{"version": version.Version, "content": version.Content, "token_count": version.TokenCount, "created_at": version.CreatedAt, "belief_ids": version.BeliefIDs, "quality": map[string]any{"approval_rate": version.Quality.ApprovalRate, "correction_rate": version.Quality.CorrectionRate, "sample_size": version.Quality.SampleSize}})
}

func (h *Handler) Summaries(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	limit, err := parseLimit(r, domain.DefaultListLimit())
	if err != nil {
		httpjson.BadRequest(w, err.Error())
		return
	}
	page, err := h.service.ListDailySummariesPage(r.Context(), userID, limit, strings.TrimSpace(r.URL.Query().Get("cursor")))
	if err != nil {
		httpjson.InternalServerError(w)
		return
	}
	items := make([]summaryResponse, 0, len(page.Summaries))
	for _, summary := range page.Summaries {
		items = append(items, summaryToResponse(summary))
	}
	httpjson.WriteJSON(w, http.StatusOK, map[string]any{"summaries": items, "next_cursor": page.NextCursor})
}

func (h *Handler) SummaryResource(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		httpjson.Unauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		httpjson.MethodNotAllowed(w)
		return
	}
	date := strings.TrimPrefix(r.URL.Path, "/learning/summaries/")
	if date == "" || strings.Contains(date, "/") {
		http.NotFound(w, r)
		return
	}
	summary, err := h.service.FindDailySummary(r.Context(), userID, date)
	if err != nil {
		h.handleLearningError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, summaryToResponse(summary))
}

func beliefFilter(r *http.Request) (application.BeliefListFilter, error) {
	query := r.URL.Query()
	filter := application.BeliefListFilter{Scope: query.Get("scope"), ScopeKey: query.Get("scope_key"), Category: query.Get("category"), PromptSlot: query.Get("prompt_slot"), Status: query.Get("status"), Cursor: query.Get("cursor")}
	if filter.Status == "" {
		filter.Status = domain.StatusActive
	}
	if filter.Scope != "" && !domain.IsValidScope(filter.Scope) || filter.Category != "" && !domain.IsValidCategory(filter.Category) || filter.PromptSlot != "" && !domain.IsValidPromptSlot(filter.PromptSlot) || !domain.IsValidStatus(filter.Status) {
		return filter, errors.New("invalid belief filter")
	}
	if raw := query.Get("trust_tier"); raw != "" {
		tier, err := strconv.Atoi(raw)
		if err != nil || tier == 0 || !domain.IsValidTrustTier(tier) {
			return filter, errors.New("invalid trust_tier")
		}
		filter.TrustTier = tier
	}
	limit, err := parseLimit(r, domain.DefaultListLimit())
	filter.Limit = limit
	return filter, err
}

func parseLimit(r *http.Request, fallback int) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 || value > 100 {
		return 0, errors.New("limit must be between 1 and 100")
	}
	return value, nil
}

func beliefToResponse(belief *domain.Belief, now time.Time) beliefResponse {
	return beliefResponse{ID: belief.ID, Statement: belief.Statement, Category: belief.Category, Scope: belief.EffectiveScope(), ScopeKey: belief.ScopeKey, Confidence: belief.Confidence, DecayedConfidence: belief.DecayedConfidence(now), EvidenceCount: belief.EvidenceCount, ContradictionCount: belief.ContradictionCount, TrustTier: belief.EffectiveTrustTier(), PromptSlot: belief.PromptSlot, Status: belief.Status, FirstObservedAt: belief.FirstObservedAt, LastReinforcedAt: belief.LastReinforcedAt}
}

func summaryToResponse(summary *domain.DailySummary) summaryResponse {
	return summaryResponse{
		ID: summary.ID, Date: summary.Date, GeneratedAt: summary.GeneratedAt, Observations: append([]string(nil), summary.Observations...),
		DeltaVsPrevious:     map[string][]string{"reinforced": summary.DeltaVsPrevious.Reinforced, "contradicted": summary.DeltaVsPrevious.Contradicted, "novel": summary.DeltaVsPrevious.Novel},
		Stats:               map[string]any{"proposals_total": summary.Stats.ProposalsTotal, "approved_direct": summary.Stats.ApprovedDirect, "approved_corrected": summary.Stats.ApprovedCorrected, "rejected": summary.Stats.Rejected, "ignored": summary.Stats.Ignored, "median_decision_seconds": summary.Stats.MedianDecisionSeconds},
		SynthesisCost:       map[string]any{"input_tokens": summary.SynthesisCost.InputTokens, "output_tokens": summary.SynthesisCost.OutputTokens, "model": summary.SynthesisCost.Model, "latency_ms": summary.SynthesisCost.LatencyMS},
		PromptVersionBefore: summary.PromptVersionBefore, PromptVersionAfter: summary.PromptVersionAfter,
	}
}

func (h *Handler) handleLearningError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrBeliefNotFound), errors.Is(err, domain.ErrPromptVersionNotFound), errors.Is(err, domain.ErrDailySummaryNotFound):
		httpjson.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidStatement), errors.Is(err, domain.ErrInvalidDate):
		httpjson.BadRequest(w, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
