package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	"github.com/google/uuid"
)

const recompileListLimit = 1000

type Service struct {
	beliefs   domain.BeliefRepository
	prompts   domain.PromptVersionRepository
	summaries domain.DailySummaryRepository
}

func NewService(beliefs domain.BeliefRepository, prompts domain.PromptVersionRepository, summaries domain.DailySummaryRepository) *Service {
	return &Service{beliefs: beliefs, prompts: prompts, summaries: summaries}
}

func (s *Service) UpsertBelief(ctx context.Context, userID, statement, category string) (*domain.Belief, error) {
	terms := memorydomain.ExtractTerms(statement, 40)
	if len(terms) > 0 {
		candidates, err := s.beliefs.SearchByTerms(ctx, userID, terms, 20)
		if err != nil {
			return nil, err
		}
		for _, candidate := range candidates {
			if candidate.Category != category {
				continue
			}
			if !domain.HasHighTermOverlap(candidate.SearchTerms, terms) {
				continue
			}
			candidate.Reinforce()
			if err := s.beliefs.Update(ctx, candidate); err != nil {
				return nil, err
			}
			return candidate, nil
		}
	}
	belief, err := domain.NewBelief(uuid.New().String(), userID, domain.BeliefCreate{
		Statement: statement,
		Category:  category,
	})
	if err != nil {
		return nil, err
	}
	if err := s.beliefs.Create(ctx, belief); err != nil {
		return nil, err
	}
	return belief, nil
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

func (s *Service) ApplyNovelBelief(ctx context.Context, userID, statement, category string, confidence float64) (*domain.Belief, error) {
	belief, err := s.UpsertBelief(ctx, userID, statement, category)
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

func (s *Service) FindDailySummary(ctx context.Context, userID, date string) (*domain.DailySummary, error) {
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
		return nil, nil
	}
	content := strings.Join(coreLines, "\n")
	return s.CreateActivePromptVersion(ctx, userID, domain.PromptVersionCreate{
		Content:                content,
		BeliefIDs:              coreIDs,
		GeneratedFromSummaryID: strings.TrimSpace(summaryID),
	})
}
