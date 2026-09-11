package application

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
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

type EntityContext struct {
	Entity      *domain.UserContext
	Beliefs     []*domain.Belief
	UserBeliefs []*domain.Belief
	Episodes    []*domain.Episode
	TokenCount  int
}

type SummaryPage struct {
	Summaries  []*domain.DailySummary
	NextCursor string
}

type Service struct {
	beliefs            domain.BeliefRepository
	prompts            domain.PromptVersionRepository
	summaries          domain.DailySummaryRepository
	contexts           domain.UserContextRepository
	candidates         domain.EntityCandidateRepository
	entityResolver     *EntityResolver
	episodes           domain.EpisodeRepository
	episodeModel       runtimedomain.ModelClient
	episodeMinSalience float64
	episodeDailyLimit  int
	promotionThreshold int
	embeddings         memoryapp.EmbeddingProvider
	dedupeThreshold    float64
	audit              AuditRecorder
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
	s.entityResolver = NewEntityResolver(contexts, s.candidates, 0)
}

func (s *Service) SetEntityCandidateRepository(candidates domain.EntityCandidateRepository, promotionThreshold int) {
	s.candidates = candidates
	if promotionThreshold <= 0 {
		promotionThreshold = domain.DefaultEntityPromotionThreshold
	}
	s.promotionThreshold = promotionThreshold
	if s.contexts != nil {
		s.entityResolver = NewEntityResolver(s.contexts, candidates, 0)
	}
}

func (s *Service) ResolveKnownEntities(ctx context.Context, userID, message string) ([]domain.EntityMatch, error) {
	if s.entityResolver == nil {
		return nil, nil
	}
	return s.entityResolver.ResolveKnown(ctx, userID, message)
}

func (s *Service) UpsertBelief(ctx context.Context, userID, statement, category string, scopeValues ...string) (*domain.Belief, error) {
	scope, scopeKey := domain.ScopeGlobal, ""
	if len(scopeValues) > 0 {
		scope = scopeValues[0]
	}
	if len(scopeValues) > 1 {
		scopeKey = scopeValues[1]
	}
	return s.upsertBelief(ctx, userID, domain.BeliefCreate{Statement: statement, Category: category, Scope: scope, ScopeKey: scopeKey, TrustTier: domain.TrustDecision}, domain.InitialConfidence)
}

func (s *Service) UpsertBeliefWithTrust(ctx context.Context, userID, statement, category, scope, scopeKey string, trustTier int, confidence float64, batchID string) (*domain.Belief, error) {
	return s.upsertBelief(ctx, userID, domain.BeliefCreate{Statement: statement, Category: category, Scope: scope, ScopeKey: scopeKey, TrustTier: trustTier, BatchID: batchID}, confidence)
}

func (s *Service) UpsertEntityBelief(ctx context.Context, userID, entityID, statement, category, factKind string, validUntil, followUpAt *time.Time, sensitive ...bool) (*domain.Belief, error) {
	if s.contexts == nil {
		return nil, domain.ErrUserContextNotFound
	}
	entity, err := s.contexts.FindByID(ctx, userID, entityID)
	if err != nil || entity == nil || !entity.CanLoadContext() || (entity.Kind != domain.ScopePerson && entity.Kind != domain.ContextKindGroup) {
		return nil, domain.ErrUserContextNotFound
	}
	return s.upsertBelief(ctx, userID, domain.BeliefCreate{
		Statement: statement, Category: category, SubjectType: domain.SubjectEntity, SubjectID: entityID,
		FactKind: factKind, ValidUntil: validUntil, FollowUpAt: followUpAt, Sensitive: len(sensitive) > 0 && sensitive[0], TrustTier: domain.TrustDecision,
	}, domain.InitialConfidence)
}

func (s *Service) upsertBelief(ctx context.Context, userID string, input domain.BeliefCreate, confidence float64) (*domain.Belief, error) {
	if confidence < 0 || confidence > 1 {
		return nil, domain.ErrInvalidConfidence
	}
	belief, err := domain.NewBelief(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	belief.SetInitialConfidence(confidence)
	if input.TrustTier == domain.TrustInferred {
		belief.PromptSlot = domain.PromptSlotSituational
	}
	terms := belief.SearchTerms
	if len(terms) > 0 {
		var candidates []*domain.Belief
		if belief.EffectiveSubjectType() == domain.SubjectEntity {
			candidates, err = s.beliefs.ListActiveBySubject(ctx, userID, domain.SubjectEntity, belief.SubjectID, recompileListLimit)
			candidates = domain.RankBeliefsByTermMatches(candidates, terms, 20)
		} else {
			candidates, err = s.beliefs.SearchByTerms(ctx, userID, terms, 20)
		}
		if err != nil {
			return nil, err
		}
		for _, candidate := range candidates {
			if candidate.Category != belief.Category || candidate.EffectiveScope() != belief.Scope || candidate.ScopeKey != belief.ScopeKey || !sameDedupeSubject(candidate, belief) {
				continue
			}
			if !domain.HasHighTermOverlap(candidate.SearchTerms, terms) {
				continue
			}
			reinforceMatch(candidate, belief, input.TrustTier)
			if err := s.beliefs.Update(ctx, candidate); err != nil {
				return nil, err
			}
			return candidate, nil
		}
	}
	embedding := s.embedStatement(ctx, userID, belief.Statement)
	if len(embedding) > 0 {
		var candidates []*domain.Belief
		if belief.EffectiveSubjectType() == domain.SubjectEntity {
			candidates, err = s.beliefs.ListActiveBySubject(ctx, userID, domain.SubjectEntity, belief.SubjectID, recompileListLimit)
		} else {
			candidates, err = s.beliefs.ListActiveByScope(ctx, userID, belief.Scope, belief.ScopeKey, recompileListLimit)
		}
		if err != nil {
			return nil, err
		}
		var match *domain.Belief
		bestSimilarity := s.dedupeThreshold
		for _, candidate := range candidates {
			if candidate.Category != belief.Category || !sameDedupeSubject(candidate, belief) {
				continue
			}
			similarity := cosineSimilarity(candidate.Embedding, embedding)
			if similarity >= bestSimilarity {
				match = candidate
				bestSimilarity = similarity
			}
		}
		if match != nil {
			reinforceMatch(match, belief, input.TrustTier)
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

func sameDedupeSubject(left, right *domain.Belief) bool {
	return left.EffectiveSubjectType() == right.EffectiveSubjectType() && left.SubjectID == right.SubjectID && left.FactKind == right.FactKind
}

func reinforceMatch(existing, incoming *domain.Belief, trustTier int) {
	existing.Reinforce(trustTier)
	previousFollowUp := existing.FollowUpAt
	existing.ValidUntil = incoming.ValidUntil
	existing.FollowUpAt = incoming.FollowUpAt
	if incoming.FollowUpAt != nil && (previousFollowUp == nil || !previousFollowUp.Equal(*incoming.FollowUpAt)) {
		existing.FollowedUpAt = nil
	}
	existing.Sensitive = existing.Sensitive || incoming.Sensitive
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
		if belief.EffectiveSubjectType() == domain.SubjectUser && belief.EffectiveScope() == domain.ScopeGlobal && belief.ScopeKey == "" && belief.EffectiveTrustTier() != domain.TrustInferred {
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
	s.invalidateEntityCatalog(userID)
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
	if !ok || (kind != domain.ScopePerson && kind != domain.ScopeMode && kind != domain.ContextKindGroup) || slug == "" || strings.Contains(slug, ":") {
		return nil, nil
	}
	value, err := s.contexts.FindBySlug(ctx, userID, kind, slug)
	if err != nil || value == nil || !value.CanLoadContext() {
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

func (s *Service) FindUserContextByID(ctx context.Context, userID, entityID string) (*domain.UserContext, error) {
	if s.contexts == nil {
		return nil, domain.ErrUserContextNotFound
	}
	value, err := s.contexts.FindByID(ctx, userID, entityID)
	if err != nil || value == nil || !value.CanLoadContext() {
		return nil, err
	}
	return value, nil
}

func (s *Service) UpdateUserContext(ctx context.Context, userID, contextID string, update domain.UserContextUpdate) (*domain.UserContext, error) {
	if update.Kind != nil {
		value, err := s.contexts.FindByID(ctx, userID, contextID)
		if err != nil {
			return nil, err
		}
		kind := strings.ToLower(strings.TrimSpace(*update.Kind))
		if existing, err := s.contexts.FindBySlug(ctx, userID, kind, value.Slug); err != nil {
			return nil, err
		} else if existing != nil && existing.ID != contextID {
			return nil, domain.ErrUserContextExists
		}
	}
	value, err := s.contexts.Update(ctx, userID, contextID, update)
	if err == nil {
		s.invalidateEntityCatalog(userID)
	}
	return value, err
}

func (s *Service) ArchiveUserContext(ctx context.Context, userID, contextID string) (*domain.UserContext, error) {
	value, err := s.contexts.Archive(ctx, userID, contextID)
	if err == nil {
		s.invalidateEntityCatalog(userID)
	}
	return value, err
}

func (s *Service) MergeUserContexts(ctx context.Context, userID, sourceID, targetID string) (*domain.UserContext, error) {
	target, err := s.contexts.FindByID(ctx, userID, targetID)
	if err != nil {
		return nil, err
	}
	source, err := s.contexts.FindByID(ctx, userID, sourceID)
	if err != nil {
		return nil, err
	}
	if err := source.Merge(target); err != nil {
		return nil, err
	}
	if err := s.beliefs.ReassignEntityFacts(ctx, userID, sourceID, targetID); err != nil {
		return nil, err
	}
	if s.episodes != nil {
		if err := s.episodes.ReassignEntity(ctx, userID, sourceID, targetID); err != nil {
			return nil, err
		}
	}
	merged, err := s.contexts.Merge(ctx, userID, sourceID, target)
	if err == nil {
		s.invalidateEntityCatalog(userID)
	}
	return merged, err
}

func (s *Service) ArchiveExpiredStates(ctx context.Context, now time.Time) (int, error) {
	return s.beliefs.ArchiveExpiredStates(ctx, now.UTC())
}

func (s *Service) GetEntityContext(ctx context.Context, userID, entityID string, tokenBudget int) (EntityContext, error) {
	if s.contexts == nil {
		return EntityContext{}, domain.ErrUserContextNotFound
	}
	entity, err := s.contexts.FindByID(ctx, userID, entityID)
	if err != nil {
		return EntityContext{}, err
	}
	if entity == nil || !entity.CanLoadContext() || !domain.IsValidContextKind(entity.Kind) {
		return EntityContext{}, domain.ErrUserContextNotFound
	}
	if tokenBudget <= 0 {
		tokenBudget = domain.DefaultEntityContextTokenBudget
	}
	facts, err := s.beliefs.ListActiveBySubject(ctx, userID, domain.SubjectEntity, entityID, recompileListLimit)
	if err != nil {
		return EntityContext{}, err
	}
	scope := entity.Kind
	if scope == domain.ContextKindGroup {
		scope = domain.ScopePerson
	}
	userBeliefs, err := s.beliefs.ListActiveByScope(ctx, userID, scope, scope+":"+entity.Slug, recompileListLimit)
	if err != nil {
		return EntityContext{}, err
	}
	now := time.Now().UTC()
	states, relationships, traits := make([]*domain.Belief, 0), make([]*domain.Belief, 0), make([]*domain.Belief, 0)
	for _, belief := range facts {
		switch belief.FactKind {
		case domain.FactKindState:
			if belief.ValidUntil != nil && belief.ValidUntil.After(now) {
				states = append(states, belief)
			} else {
				belief.Status = domain.StatusArchived
				if err := s.beliefs.Update(ctx, belief); err != nil {
					return EntityContext{}, err
				}
			}
		case domain.FactKindRelationship:
			relationships = append(relationships, belief)
		case domain.FactKindTrait:
			traits = append(traits, belief)
		}
	}
	sort.SliceStable(states, func(i, j int) bool { return states[i].LastReinforcedAt.After(states[j].LastReinforcedAt) })
	sortByBeliefValue(relationships, now)
	sortByBeliefValue(traits, now)
	filteredUserBeliefs := userBeliefs[:0]
	for _, belief := range userBeliefs {
		if belief.EffectiveSubjectType() == domain.SubjectUser {
			filteredUserBeliefs = append(filteredUserBeliefs, belief)
		}
	}
	sortByBeliefValue(filteredUserBeliefs, now)

	result := EntityContext{Entity: entity, Beliefs: make([]*domain.Belief, 0, len(facts)+len(filteredUserBeliefs))}
	content := ""
	for _, belief := range append(append(append(states, relationships...), traits...), filteredUserBeliefs...) {
		candidate := belief.Statement
		if content != "" {
			candidate = content + "\n" + belief.Statement
		}
		tokens := domain.ApproximateTokens(candidate)
		if tokens > tokenBudget {
			break
		}
		result.Beliefs = append(result.Beliefs, belief)
		result.TokenCount = tokens
		content = candidate
	}
	return result, nil
}

func (s *Service) GetEntityDetails(ctx context.Context, userID, entityID string) (EntityContext, error) {
	entity, err := s.contexts.FindByID(ctx, userID, entityID)
	if err != nil {
		return EntityContext{}, err
	}
	if entity == nil {
		return EntityContext{}, domain.ErrUserContextNotFound
	}
	beliefs, err := s.beliefs.List(ctx, userID, recompileListLimit)
	if err != nil {
		return EntityContext{}, err
	}
	facts := make([]*domain.Belief, 0)
	userBeliefs := make([]*domain.Belief, 0)
	scope := entity.Kind
	if scope == domain.ContextKindGroup {
		scope = domain.ScopePerson
	}
	scopeKey := scope + ":" + entity.Slug
	for _, belief := range beliefs {
		if belief.EffectiveSubjectType() == domain.SubjectEntity && belief.SubjectID == entityID {
			facts = append(facts, belief)
		} else if belief.EffectiveSubjectType() == domain.SubjectUser && belief.EffectiveScope() == scope && belief.ScopeKey == scopeKey {
			userBeliefs = append(userBeliefs, belief)
		}
	}
	sort.SliceStable(facts, func(i, j int) bool { return facts[i].LastReinforcedAt.After(facts[j].LastReinforcedAt) })
	sort.SliceStable(userBeliefs, func(i, j int) bool { return userBeliefs[i].LastReinforcedAt.After(userBeliefs[j].LastReinforcedAt) })
	episodes, err := s.SearchEpisodes(ctx, userID, entityID, nil, domain.DefaultEpisodeSearchLimit)
	if err != nil {
		return EntityContext{}, err
	}
	return EntityContext{Entity: entity, Beliefs: facts, UserBeliefs: userBeliefs, Episodes: episodes}, nil
}

func (s *Service) ResolveEntities(ctx context.Context, userID, message string) ([]domain.EntityMatch, error) {
	if s.entityResolver == nil {
		return []domain.EntityMatch{}, nil
	}
	return s.entityResolver.Resolve(ctx, userID, message)
}

func (s *Service) ListEntities(ctx context.Context, userID, kind, status string) ([]*domain.UserContext, error) {
	kind, status = strings.ToLower(strings.TrimSpace(kind)), strings.ToLower(strings.TrimSpace(status))
	if kind != "" && !domain.IsValidContextKind(kind) {
		return nil, domain.ErrInvalidContextKind
	}
	if status != "" && !domain.IsValidContextStatus(status) {
		return nil, domain.ErrInvalidContextStatus
	}
	values, err := s.ListUserContexts(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.UserContext, 0, len(values))
	for _, value := range values {
		if (kind == "" || value.Kind == kind) && (status == "" || value.EffectiveStatus() == status) {
			result = append(result, value)
		}
	}
	return result, nil
}

func (s *Service) PromoteEntityCandidates(ctx context.Context, userID string, now time.Time) (int, error) {
	if s.candidates == nil {
		return 0, nil
	}
	threshold := s.promotionThreshold
	if threshold <= 0 {
		threshold = domain.DefaultEntityPromotionThreshold
	}
	values, err := s.candidates.List(ctx, userID)
	if err != nil {
		return 0, err
	}
	promoted := 0
	for _, candidate := range values {
		if candidate.MentionCount >= threshold {
			slug := strings.Join(strings.Fields(candidate.NormalizedName), "-")
			if existing, err := s.contexts.FindBySlug(ctx, userID, domain.ScopePerson, slug); err != nil {
				return promoted, err
			} else if existing == nil {
				if _, err := s.CreateUserContext(ctx, userID, domain.UserContextCreate{
					Kind: domain.ScopePerson, Slug: slug, Label: candidate.Name, Aliases: []string{candidate.Name},
					CreatedFrom: domain.CreatedFromInferred, Status: domain.ContextStatusPendingReview,
				}); err != nil {
					return promoted, err
				}
				promoted++
			}
			if err := s.candidates.Delete(ctx, userID, candidate.NormalizedName); err != nil {
				return promoted, err
			}
			continue
		}
		if !candidate.FirstSeenAt.IsZero() && now.Sub(candidate.FirstSeenAt) >= domain.EntityCandidateRetention {
			if err := s.candidates.Delete(ctx, userID, candidate.NormalizedName); err != nil {
				return promoted, err
			}
		}
	}
	return promoted, nil
}

func (s *Service) invalidateEntityCatalog(userID string) {
	if s.entityResolver != nil {
		s.entityResolver.Invalidate(userID)
	}
}

func sortByBeliefValue(beliefs []*domain.Belief, now time.Time) {
	sort.SliceStable(beliefs, func(i, j int) bool {
		return beliefs[i].PromptValue(now) > beliefs[j].PromptValue(now)
	})
}

func (s *Service) ContextBeliefStatements(ctx context.Context, userID, scopeKey string) ([]string, error) {
	contextValue, err := s.FindUserContextByScopeKey(ctx, userID, scopeKey)
	if err != nil || contextValue == nil {
		return nil, err
	}
	entityContext, err := s.GetEntityContext(ctx, userID, contextValue.ID, 0)
	if err != nil {
		return nil, err
	}
	statements := make([]string, len(entityContext.Beliefs))
	for i, belief := range entityContext.Beliefs {
		statements[i] = belief.Statement
	}
	return statements, nil
}
