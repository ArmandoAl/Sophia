package application

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	memoryapp "github.com/armandoalvarado/sofia-backend/internal/memory/application"
	"github.com/google/uuid"
)

const recompileListLimit = 1000

type BeliefListFilter struct {
	Scope, ScopeKey, Category, PromptSlot, Status, Cursor string
	TrustTier, Limit                                      int
}

type BeliefPage struct {
	Beliefs    []*domain.Belief
	NextCursor string
}

type SummaryPage struct {
	Summaries  []*domain.DailySummary
	NextCursor string
}

type Service struct {
	beliefs         domain.BeliefRepository
	prompts         domain.PromptVersionRepository
	summaries       domain.DailySummaryRepository
	contexts        domain.UserContextRepository
	embeddings      memoryapp.EmbeddingProvider
	dedupeThreshold float64
	audit           AuditRecorder
}

type AuditRecorder interface {
	RecordAuditLog(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) error
}

func NewService(beliefs domain.BeliefRepository, prompts domain.PromptVersionRepository, summaries domain.DailySummaryRepository) *Service {
	return &Service{beliefs: beliefs, prompts: prompts, summaries: summaries}
}

func (s *Service) SetEmbeddingProvider(provider memoryapp.EmbeddingProvider, dedupeThreshold float64) {
	s.embeddings = provider
	s.dedupeThreshold = dedupeThreshold
}

func (s *Service) SetAuditRecorder(audit AuditRecorder) {
	s.audit = audit
}

func (s *Service) SetContextRepository(contexts domain.UserContextRepository) {
	s.contexts = contexts
}

func (s *Service) UpsertBelief(ctx context.Context, userID, statement, category string, scopeValues ...string) (*domain.Belief, error) {
	scope, scopeKey := domain.ScopeGlobal, ""
	if len(scopeValues) > 0 {
		scope = scopeValues[0]
	}
	if len(scopeValues) > 1 {
		scopeKey = scopeValues[1]
	}
	return s.upsertBelief(ctx, userID, statement, category, scope, scopeKey, domain.TrustDecision, domain.InitialConfidence, "")
}

func (s *Service) UpsertBeliefWithTrust(ctx context.Context, userID, statement, category, scope, scopeKey string, trustTier int, confidence float64, batchID string) (*domain.Belief, error) {
	return s.upsertBelief(ctx, userID, statement, category, scope, scopeKey, trustTier, confidence, batchID)
}

func (s *Service) upsertBelief(ctx context.Context, userID, statement, category, scope, scopeKey string, trustTier int, confidence float64, batchID string) (*domain.Belief, error) {
	if confidence < 0 || confidence > 1 {
		return nil, domain.ErrInvalidConfidence
	}
	belief, err := domain.NewBelief(uuid.New().String(), userID, domain.BeliefCreate{
		Statement: statement,
		Category:  category,
		Scope:     scope,
		ScopeKey:  scopeKey,
		TrustTier: trustTier,
		BatchID:   batchID,
	})
	if err != nil {
		return nil, err
	}
	belief.SetInitialConfidence(confidence)
	if trustTier == domain.TrustInferred {
		belief.PromptSlot = domain.PromptSlotSituational
	}
	terms := belief.SearchTerms
	if len(terms) > 0 {
		candidates, err := s.beliefs.SearchByTerms(ctx, userID, terms, 20)
		if err != nil {
			return nil, err
		}
		for _, candidate := range candidates {
			if candidate.Category != category || candidate.EffectiveScope() != belief.Scope || candidate.ScopeKey != belief.ScopeKey {
				continue
			}
			if !domain.HasHighTermOverlap(candidate.SearchTerms, terms) {
				continue
			}
			candidate.Reinforce(trustTier)
			if err := s.beliefs.Update(ctx, candidate); err != nil {
				return nil, err
			}
			return candidate, nil
		}
	}
	embedding := s.embedStatement(ctx, userID, statement)
	if len(embedding) > 0 {
		candidates, err := s.beliefs.ListActiveByScope(ctx, userID, belief.Scope, belief.ScopeKey, recompileListLimit)
		if err != nil {
			return nil, err
		}
		var match *domain.Belief
		bestSimilarity := s.dedupeThreshold
		for _, candidate := range candidates {
			if candidate.Category != category {
				continue
			}
			similarity := cosineSimilarity(candidate.Embedding, embedding)
			if similarity >= bestSimilarity {
				match = candidate
				bestSimilarity = similarity
			}
		}
		if match != nil {
			match.Reinforce(trustTier)
			if err := s.beliefs.Update(ctx, match); err != nil {
				return nil, err
			}
			return match, nil
		}
	}
	belief.Embedding = append([]float32(nil), embedding...)
	if err := s.beliefs.Create(ctx, belief); err != nil {
		return nil, err
	}
	return belief, nil
}

func (s *Service) RetireBeliefsByBatch(ctx context.Context, userID, batchID string) (int, error) {
	return s.beliefs.RetireByBatchID(ctx, userID, batchID)
}

func (s *Service) embedStatement(ctx context.Context, userID, statement string) []float32 {
	if s.embeddings == nil {
		s.recordEmbeddingDegradation(ctx, userID, "disabled")
		return nil
	}
	embedding, err := s.embeddings.EmbedText(ctx, statement)
	if err != nil {
		s.recordEmbeddingDegradation(ctx, userID, "provider_error")
		return nil
	}
	if len(embedding) == 0 {
		s.recordEmbeddingDegradation(ctx, userID, "empty_vector")
		return nil
	}
	return embedding
}

func (s *Service) recordEmbeddingDegradation(ctx context.Context, userID, reason string) {
	if s.audit != nil {
		_ = s.audit.RecordAuditLog(ctx, userID, "belief_embedding_degraded", "belief", "", map[string]any{"reason": reason})
	}
}

func cosineSimilarity(left, right []float32) float64 {
	if len(left) == 0 || len(left) != len(right) {
		return -1
	}
	var dot, leftNorm, rightNorm float64
	for i := range left {
		l, r := float64(left[i]), float64(right[i])
		dot += l * r
		leftNorm += l * l
		rightNorm += r * r
	}
	if leftNorm == 0 || rightNorm == 0 {
		return -1
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

func (s *Service) GetBelief(ctx context.Context, userID, beliefID string) (*domain.Belief, error) {
	belief, err := s.beliefs.FindByID(ctx, userID, beliefID)
	if err != nil {
		return nil, err
	}
	if belief == nil {
		return nil, domain.ErrBeliefNotFound
	}
	return belief, nil
}

func (s *Service) ListActiveBeliefs(ctx context.Context, userID string, limit int) ([]*domain.Belief, error) {
	if limit <= 0 {
		limit = domain.DefaultListLimit()
	}
	return s.beliefs.ListActive(ctx, userID, limit)
}

func (s *Service) ListBeliefs(ctx context.Context, userID string, filter BeliefListFilter) (BeliefPage, error) {
	beliefs, err := s.beliefs.List(ctx, userID, 0)
	if err != nil {
		return BeliefPage{}, err
	}
	filtered := beliefs[:0]
	for _, belief := range beliefs {
		if filter.Scope != "" && belief.EffectiveScope() != filter.Scope || filter.ScopeKey != "" && belief.ScopeKey != filter.ScopeKey || filter.Category != "" && belief.Category != filter.Category || filter.TrustTier != 0 && belief.EffectiveTrustTier() != filter.TrustTier || filter.PromptSlot != "" && belief.PromptSlot != filter.PromptSlot || filter.Status != "" && belief.Status != filter.Status {
			continue
		}
		filtered = append(filtered, belief)
	}
	now := time.Now().UTC()
	sort.SliceStable(filtered, func(i, j int) bool {
		left, right := filtered[i].PromptValue(now), filtered[j].PromptValue(now)
		if left == right {
			return filtered[i].ID < filtered[j].ID
		}
		return left > right
	})
	start := 0
	if filter.Cursor != "" {
		for i, belief := range filtered {
			if belief.ID == filter.Cursor {
				start = i + 1
				break
			}
		}
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = domain.DefaultListLimit()
	}
	end := min(start+limit, len(filtered))
	page := BeliefPage{Beliefs: filtered[start:end]}
	if end < len(filtered) && end > start {
		page.NextCursor = filtered[end-1].ID
	}
	return page, nil
}

func (s *Service) UpdateBeliefStatement(ctx context.Context, userID, beliefID, statement string) (*domain.Belief, error) {
	belief, err := s.GetBelief(ctx, userID, beliefID)
	if err != nil {
		return nil, err
	}
	if err := belief.UpdateStatement(statement); err != nil {
		return nil, err
	}
	belief.Embedding = s.embedStatement(ctx, userID, belief.Statement)
	if err := s.beliefs.Update(ctx, belief); err != nil {
		return nil, err
	}
	s.recordBeliefAudit(ctx, userID, "belief_edited", belief.ID)
	return belief, nil
}

func (s *Service) RetireBelief(ctx context.Context, userID, beliefID string) (*domain.Belief, error) {
	belief, err := s.GetBelief(ctx, userID, beliefID)
	if err != nil {
		return nil, err
	}
	belief.Status = domain.StatusRetired
	belief.PromptSlot = domain.PromptSlotSituational
	if err := s.beliefs.Update(ctx, belief); err != nil {
		return nil, err
	}
	s.recordBeliefAudit(ctx, userID, "belief_retired", belief.ID)
	return belief, nil
}

func (s *Service) recordBeliefAudit(ctx context.Context, userID, action, beliefID string) {
	if s.audit != nil {
		_ = s.audit.RecordAuditLog(ctx, userID, action, "belief", beliefID, nil)
	}
}

func (s *Service) ReinforceBelief(ctx context.Context, userID, beliefID string) (*domain.Belief, error) {
	belief, err := s.GetBelief(ctx, userID, beliefID)
	if err != nil {
		return nil, err
	}
	belief.Reinforce()
	if err := s.beliefs.Update(ctx, belief); err != nil {
		return nil, err
	}
	return belief, nil
}

func (s *Service) ApplyNovelBelief(ctx context.Context, userID, statement, category string, confidence float64, scopeValues ...string) (*domain.Belief, error) {
	belief, err := s.UpsertBelief(ctx, userID, statement, category, scopeValues...)
	if err != nil {
		return nil, err
	}
	if belief.EvidenceCount == 1 && confidence >= 0 && confidence <= 1 {
		belief.Confidence = confidence
		if err := s.beliefs.Update(ctx, belief); err != nil {
			return nil, err
		}
	}
	return belief, nil
}

func (s *Service) ContradictBelief(ctx context.Context, userID, beliefID string) (*domain.Belief, error) {
	belief, err := s.GetBelief(ctx, userID, beliefID)
	if err != nil {
		return nil, err
	}
	belief.Contradict()
	if err := s.beliefs.Update(ctx, belief); err != nil {
		return nil, err
	}
	return belief, nil
}

func (s *Service) SupersedeBelief(ctx context.Context, userID, beliefID, newID, reason string) (*domain.Belief, error) {
	belief, err := s.GetBelief(ctx, userID, beliefID)
	if err != nil {
		return nil, err
	}
	if err := belief.Supersede(newID, reason); err != nil {
		return nil, err
	}
	if err := s.beliefs.Update(ctx, belief); err != nil {
		return nil, err
	}
	return belief, nil
}

func (s *Service) SetPromptSlot(ctx context.Context, userID, beliefID, slot string) (*domain.Belief, error) {
	return s.beliefs.SetPromptSlot(ctx, userID, beliefID, slot)
}

func (s *Service) CreateActivePromptVersion(ctx context.Context, userID string, input domain.PromptVersionCreate) (*domain.PromptVersion, error) {
	version, err := domain.NewPromptVersion(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.prompts.CreateActive(ctx, version); err != nil {
		return nil, err
	}
	return version, nil
}

func (s *Service) GetActivePromptVersion(ctx context.Context, userID string) (*domain.PromptVersion, error) {
	return s.prompts.GetActive(ctx, userID)
}

func (s *Service) ActivePromptContent(ctx context.Context, userID string) (string, error) {
	version, err := s.GetActivePromptVersion(ctx, userID)
	if err != nil || version == nil {
		return "", err
	}
	return version.Content, nil
}

func (s *Service) ListPromptVersions(ctx context.Context, userID string, limit int) ([]*domain.PromptVersion, error) {
	if limit <= 0 {
		limit = domain.DefaultListLimit()
	}
	return s.prompts.ListVersions(ctx, userID, limit)
}

func (s *Service) CreateDailySummary(ctx context.Context, userID string, input domain.DailySummaryCreate) (*domain.DailySummary, error) {
	summary, err := domain.NewDailySummary(userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.summaries.Create(ctx, summary); err != nil {
		return nil, err
	}
	return summary, nil
}

func (s *Service) UpdateDailySummary(ctx context.Context, summary *domain.DailySummary) error {
	return s.summaries.Update(ctx, summary)
}

func (s *Service) FindDailySummary(ctx context.Context, userID, date string) (*domain.DailySummary, error) {
	date = strings.TrimSpace(date)
	if _, err := time.Parse(domain.DateLayout, date); err != nil {
		return nil, domain.ErrInvalidDate
	}
	summary, err := s.summaries.FindByDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return nil, domain.ErrDailySummaryNotFound
	}
	return summary, nil
}

func (s *Service) ListRecentDailySummaries(ctx context.Context, userID string, limit int) ([]*domain.DailySummary, error) {
	if limit <= 0 {
		limit = domain.DefaultListLimit()
	}
	return s.summaries.ListRecent(ctx, userID, limit)
}

func (s *Service) ListDailySummariesPage(ctx context.Context, userID string, limit int, cursor string) (SummaryPage, error) {
	values, err := s.summaries.ListRecent(ctx, userID, recompileListLimit)
	if err != nil {
		return SummaryPage{}, err
	}
	start := 0
	if cursor != "" {
		for i, value := range values {
			if value.Date == cursor || value.ID == cursor {
				start = i + 1
				break
			}
		}
	}
	if limit <= 0 {
		limit = domain.DefaultListLimit()
	}
	end := min(start+limit, len(values))
	page := SummaryPage{Summaries: values[start:end]}
	if end < len(values) && end > 0 {
		page.NextCursor = values[end-1].Date
	}
	return page, nil
}

func (s *Service) SearchBeliefs(ctx context.Context, userID string, terms []string, limit int) ([]*domain.Belief, error) {
	return s.beliefs.SearchByTerms(ctx, userID, terms, limit)
}

func (s *Service) RecompilePromptBase(ctx context.Context, userID string) (*domain.PromptVersion, error) {
	return s.RecompilePromptBaseFromSummary(ctx, userID, "")
}

func (s *Service) RecompilePromptBaseFromSummary(ctx context.Context, userID, summaryID string) (*domain.PromptVersion, error) {
	const promptTokenCeiling = domain.PromptTokenCeiling

	beliefs, err := s.ListActiveBeliefs(ctx, userID, recompileListLimit)
	if err != nil {
		return nil, err
	}
	globals := beliefs[:0]
	for _, belief := range beliefs {
		// Los chats inferidos sirven al contexto situacional, pero no gobiernan el prompt base hasta que una decisión los confirme.
		if belief.EffectiveScope() == domain.ScopeGlobal && belief.ScopeKey == "" && belief.EffectiveTrustTier() != domain.TrustInferred {
			globals = append(globals, belief)
		} else if belief.EffectiveTrustTier() == domain.TrustInferred && belief.PromptSlot != domain.PromptSlotSituational {
			if _, err := s.SetPromptSlot(ctx, userID, belief.ID, domain.PromptSlotSituational); err != nil {
				return nil, err
			}
		}
	}
	beliefs = globals
	now := time.Now().UTC()
	sort.SliceStable(beliefs, func(i, j int) bool {
		left := beliefs[i].PromptValue(now)
		right := beliefs[j].PromptValue(now)
		if left == right {
			return beliefs[i].LastReinforcedAt.After(beliefs[j].LastReinforcedAt)
		}
		return left > right
	})

	coreLines := make([]string, 0, len(beliefs))
	coreIDs := make([]string, 0, len(beliefs))
	for _, belief := range beliefs {
		candidate := belief.Statement
		if len(coreLines) > 0 {
			candidate = strings.Join(append(append([]string{}, coreLines...), belief.Statement), "\n")
		}
		slot := domain.PromptSlotSituational
		if domain.ApproximateTokens(candidate) <= promptTokenCeiling {
			slot = domain.PromptSlotCore
			coreLines = append(coreLines, belief.Statement)
			coreIDs = append(coreIDs, belief.ID)
		}
		if _, err := s.SetPromptSlot(ctx, userID, belief.ID, slot); err != nil {
			return nil, err
		}
	}
	if len(coreLines) == 0 {
		return nil, s.prompts.DeactivateActive(ctx, userID)
	}
	content := strings.Join(coreLines, "\n")
	return s.CreateActivePromptVersion(ctx, userID, domain.PromptVersionCreate{
		Content:                content,
		BeliefIDs:              coreIDs,
		GeneratedFromSummaryID: strings.TrimSpace(summaryID),
	})
}

func (s *Service) CreateUserContext(ctx context.Context, userID string, input domain.UserContextCreate) (*domain.UserContext, error) {
	value, err := domain.NewUserContext(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	if existing, err := s.contexts.FindBySlug(ctx, userID, value.Kind, value.Slug); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, domain.ErrUserContextExists
	}
	if err := s.contexts.Create(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Service) ListUserContexts(ctx context.Context, userID string) ([]*domain.UserContext, error) {
	if s.contexts == nil {
		return []*domain.UserContext{}, nil
	}
	return s.contexts.List(ctx, userID)
}

func (s *Service) FindUserContextByScopeKey(ctx context.Context, userID, scopeKey string) (*domain.UserContext, error) {
	if s.contexts == nil {
		return nil, nil
	}
	kind, slug, ok := strings.Cut(strings.ToLower(strings.TrimSpace(scopeKey)), ":")
	if !ok || (kind != domain.ScopePerson && kind != domain.ScopeMode) || slug == "" || strings.Contains(slug, ":") {
		return nil, nil
	}
	value, err := s.contexts.FindBySlug(ctx, userID, kind, slug)
	if err != nil || value == nil || !value.Active {
		return nil, err
	}
	return value, nil
}

func (s *Service) FindUserContextBySlug(ctx context.Context, userID, kind, slug string) (*domain.UserContext, error) {
	value, err := s.contexts.FindBySlug(ctx, userID, strings.ToLower(strings.TrimSpace(kind)), strings.ToLower(strings.TrimSpace(slug)))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, domain.ErrUserContextNotFound
	}
	return value, nil
}

func (s *Service) UpdateUserContext(ctx context.Context, userID, contextID string, update domain.UserContextUpdate) (*domain.UserContext, error) {
	return s.contexts.Update(ctx, userID, contextID, update)
}

func (s *Service) ArchiveUserContext(ctx context.Context, userID, contextID string) (*domain.UserContext, error) {
	return s.contexts.Archive(ctx, userID, contextID)
}

func (s *Service) ContextBeliefStatements(ctx context.Context, userID, scopeKey string) ([]string, error) {
	contextValue, err := s.FindUserContextByScopeKey(ctx, userID, scopeKey)
	if err != nil || contextValue == nil {
		return nil, err
	}
	beliefs, err := s.beliefs.ListActiveByScope(ctx, userID, contextValue.Kind, contextValue.ScopeKey(), recompileListLimit)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	sort.SliceStable(beliefs, func(i, j int) bool {
		left, right := beliefs[i].PromptValue(now), beliefs[j].PromptValue(now)
		if left == right {
			return beliefs[i].LastReinforcedAt.After(beliefs[j].LastReinforcedAt)
		}
		return left > right
	})
	statements := make([]string, 0, len(beliefs))
	for _, belief := range beliefs {
		candidate := strings.Join(append(append([]string{}, statements...), belief.Statement), "\n")
		if domain.ApproximateTokens(candidate) <= domain.ContextFragmentTokenCeiling {
			statements = append(statements, belief.Statement)
		}
	}
	return statements, nil
}
