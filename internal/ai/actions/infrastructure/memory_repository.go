package infrastructure

import (
	"context"
	"sort"
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
)

type InMemoryAIActionProposalRepository struct {
	mu        sync.Mutex
	proposals map[string]*domain.AIActionProposal
}

func NewInMemoryAIActionProposalRepository() *InMemoryAIActionProposalRepository {
	return &InMemoryAIActionProposalRepository{proposals: make(map[string]*domain.AIActionProposal)}
}

func (r *InMemoryAIActionProposalRepository) Create(ctx context.Context, proposal *domain.AIActionProposal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.proposals[proposal.ID] = cloneProposal(proposal)
	return nil
}

func (r *InMemoryAIActionProposalRepository) FindByID(ctx context.Context, userID, proposalID string) (*domain.AIActionProposal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	proposal, ok := r.proposals[proposalID]
	if !ok || proposal.UserID != userID {
		return nil, nil
	}
	return cloneProposal(proposal), nil
}

func (r *InMemoryAIActionProposalRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.AIActionProposal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.AIActionProposal, 0)
	for _, proposal := range r.proposals {
		if proposal.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && proposal.Status != filter.Status {
			continue
		}
		result = append(result, cloneProposal(proposal))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *InMemoryAIActionProposalRepository) Update(ctx context.Context, proposal *domain.AIActionProposal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.proposals[proposal.ID] = cloneProposal(proposal)
	return nil
}

func cloneProposal(proposal *domain.AIActionProposal) *domain.AIActionProposal {
	if proposal == nil {
		return nil
	}
	cp := *proposal
	cp.ProposedInput = append([]byte(nil), proposal.ProposedInput...)
	cp.ExecutionResult = append([]byte(nil), proposal.ExecutionResult...)
	return &cp
}
