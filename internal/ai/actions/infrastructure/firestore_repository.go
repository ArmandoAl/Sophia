package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const actionProposalsCollection = "ai_action_proposals"

type FirestoreAIActionProposalRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

func NewFirestoreAIActionProposalRepository(client *firestore.Client) *FirestoreAIActionProposalRepository {
	return &FirestoreAIActionProposalRepository{client: client, timeout: 5 * time.Second}
}

type firestoreProposal struct {
	ID                   string     `firestore:"id"`
	UserID               string     `firestore:"user_id"`
	ToolName             string     `firestore:"tool_name"`
	Status               string     `firestore:"status"`
	ProposedInput        string     `firestore:"proposed_input"`
	Reason               string     `firestore:"reason"`
	RiskLevel            string     `firestore:"risk_level"`
	RequiresConfirmation bool       `firestore:"requires_confirmation"`
	ExpiresAt            *time.Time `firestore:"expires_at,omitempty"`
	CreatedAt            time.Time  `firestore:"created_at"`
	UpdatedAt            time.Time  `firestore:"updated_at"`
	ConfirmedAt          *time.Time `firestore:"confirmed_at,omitempty"`
	RejectedAt           *time.Time `firestore:"rejected_at,omitempty"`
	ExecutedAt           *time.Time `firestore:"executed_at,omitempty"`
	ExecutionResult      string     `firestore:"execution_result"`
	ExecutionError       string     `firestore:"execution_error"`
}

func (r *FirestoreAIActionProposalRepository) Create(ctx context.Context, proposal *domain.AIActionProposal) error {
	return r.Update(ctx, proposal)
}

func (r *FirestoreAIActionProposalRepository) FindByID(ctx context.Context, userID, proposalID string) (*domain.AIActionProposal, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc, err := r.client.Collection(actionProposalsCollection).Doc(proposalID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	proposal, err := documentToProposal(doc)
	if err != nil {
		return nil, err
	}
	if proposal.UserID != userID {
		return nil, nil
	}
	return proposal, nil
}

func (r *FirestoreAIActionProposalRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.AIActionProposal, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(actionProposalsCollection).Where("user_id", "==", filter.UserID)
	if filter.Status != "" {
		query = query.Where("status", "==", filter.Status)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.AIActionProposal, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		proposal, err := documentToProposal(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, proposal)
	}
	return result, nil
}

func (r *FirestoreAIActionProposalRepository) Update(ctx context.Context, proposal *domain.AIActionProposal) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(actionProposalsCollection).Doc(proposal.ID).Set(ctx, proposalToDocument(proposal))
	return err
}

func proposalToDocument(p *domain.AIActionProposal) firestoreProposal {
	return firestoreProposal{ID: p.ID, UserID: p.UserID, ToolName: p.ToolName, Status: p.Status, ProposedInput: string(p.ProposedInput), Reason: p.Reason, RiskLevel: p.RiskLevel, RequiresConfirmation: p.RequiresConfirmation, ExpiresAt: p.ExpiresAt, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, ConfirmedAt: p.ConfirmedAt, RejectedAt: p.RejectedAt, ExecutedAt: p.ExecutedAt, ExecutionResult: string(p.ExecutionResult), ExecutionError: p.ExecutionError}
}

func documentToProposal(doc *firestore.DocumentSnapshot) (*domain.AIActionProposal, error) {
	var p firestoreProposal
	if err := doc.DataTo(&p); err != nil {
		return nil, err
	}
	return &domain.AIActionProposal{ID: p.ID, UserID: p.UserID, ToolName: p.ToolName, Status: p.Status, ProposedInput: []byte(p.ProposedInput), Reason: p.Reason, RiskLevel: p.RiskLevel, RequiresConfirmation: p.RequiresConfirmation, ExpiresAt: p.ExpiresAt, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, ConfirmedAt: p.ConfirmedAt, RejectedAt: p.RejectedAt, ExecutedAt: p.ExecutedAt, ExecutionResult: []byte(p.ExecutionResult), ExecutionError: p.ExecutionError}, nil
}
