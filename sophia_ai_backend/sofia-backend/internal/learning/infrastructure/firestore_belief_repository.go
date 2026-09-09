package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ domain.BeliefRepository = (*FirestoreBeliefRepository)(nil)

const userBeliefsCollection = "user_beliefs"

type FirestoreBeliefRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreBelief struct {
	ID                 string     `firestore:"id"`
	UserID             string     `firestore:"user_id"`
	Statement          string     `firestore:"statement"`
	Category           string     `firestore:"category"`
	Confidence         float64    `firestore:"confidence"`
	EvidenceCount      int        `firestore:"evidence_count"`
	ContradictionCount int        `firestore:"contradiction_count"`
	FirstObservedAt    time.Time  `firestore:"first_observed_at"`
	LastReinforcedAt   time.Time  `firestore:"last_reinforced_at"`
	LastContradictedAt *time.Time `firestore:"last_contradicted_at,omitempty"`
	Status             string     `firestore:"status"`
	SupersededBy       string     `firestore:"superseded_by"`
	SupersedeReason    string     `firestore:"supersede_reason"`
	PromptSlot         string     `firestore:"prompt_slot"`
	TokenCost          int        `firestore:"token_cost"`
	SearchTerms        []string   `firestore:"search_terms"`
}

func NewFirestoreBeliefRepository(client *firestore.Client) *FirestoreBeliefRepository {
	return &FirestoreBeliefRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreBeliefRepository) Create(ctx context.Context, belief *domain.Belief) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(userBeliefsCollection).Doc(belief.ID).Set(ctx, beliefToDocument(belief))
	return err
}

func (r *FirestoreBeliefRepository) Update(ctx context.Context, belief *domain.Belief) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(userBeliefsCollection).Doc(belief.ID).Set(ctx, beliefToDocument(belief))
	return err
}

func (r *FirestoreBeliefRepository) FindByID(ctx context.Context, userID, beliefID string) (*domain.Belief, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc, err := r.client.Collection(userBeliefsCollection).Doc(beliefID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	belief, err := documentToBelief(doc)
	if err != nil {
		return nil, err
	}
	if belief.UserID != userID {
		return nil, nil
	}
	return belief, nil
}

func (r *FirestoreBeliefRepository) ListActive(ctx context.Context, userID string, limit int) ([]*domain.Belief, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(userBeliefsCollection).
		Where("user_id", "==", userID).
		Where("status", "==", domain.StatusActive).
		OrderBy("confidence", firestore.Desc)
	if limit > 0 {
		query = query.Limit(limit)
	}
	return r.collect(ctx, query)
}

func (r *FirestoreBeliefRepository) SearchByTerms(ctx context.Context, userID string, terms []string, limit int) ([]*domain.Belief, error) {
	terms = memorydomain.CapSearchTerms(terms)
	if len(terms) == 0 {
		return []*domain.Belief{}, nil
	}
	if limit <= 0 {
		limit = domain.DefaultSearchLimit()
	}
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(userBeliefsCollection).
		Where("user_id", "==", userID).
		Where("status", "==", domain.StatusActive).
		Where("search_terms", "array-contains-any", terms).
		Limit(limit * 4)
	beliefs, err := r.collect(ctx, query)
	if err != nil {
		return nil, err
	}
	return domain.RankBeliefsByTermMatches(beliefs, terms, limit), nil
}

func (r *FirestoreBeliefRepository) SetPromptSlot(ctx context.Context, userID, beliefID, slot string) (*domain.Belief, error) {
	if !domain.IsValidPromptSlot(slot) {
		return nil, domain.ErrInvalidPromptSlot
	}
	belief, err := r.FindByID(ctx, userID, beliefID)
	if err != nil {
		return nil, err
	}
	if belief == nil {
		return nil, domain.ErrBeliefNotFound
	}
	belief.PromptSlot = slot
	if err := belief.Validate(); err != nil {
		return nil, err
	}
	if err := r.Update(ctx, belief); err != nil {
		return nil, err
	}
	return belief, nil
}

func (r *FirestoreBeliefRepository) collect(ctx context.Context, query firestore.Query) ([]*domain.Belief, error) {
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.Belief, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		belief, err := documentToBelief(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, belief)
	}
	return result, nil
}

func beliefToDocument(belief *domain.Belief) firestoreBelief {
	return firestoreBelief{
		ID:                 belief.ID,
		UserID:             belief.UserID,
		Statement:          belief.Statement,
		Category:           belief.Category,
		Confidence:         belief.Confidence,
		EvidenceCount:      belief.EvidenceCount,
		ContradictionCount: belief.ContradictionCount,
		FirstObservedAt:    belief.FirstObservedAt,
		LastReinforcedAt:   belief.LastReinforcedAt,
		LastContradictedAt: belief.LastContradictedAt,
		Status:             belief.Status,
		SupersededBy:       belief.SupersededBy,
		SupersedeReason:    belief.SupersedeReason,
		PromptSlot:         belief.PromptSlot,
		TokenCost:          belief.TokenCost,
		SearchTerms:        cloneStrings(belief.SearchTerms),
	}
}

func documentToBelief(doc *firestore.DocumentSnapshot) (*domain.Belief, error) {
	var stored firestoreBelief
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return stored.toDomain(), nil
}

func (m firestoreBelief) toDomain() *domain.Belief {
	return &domain.Belief{
		ID:                 m.ID,
		UserID:             m.UserID,
		Statement:          m.Statement,
		Category:           m.Category,
		Confidence:         m.Confidence,
		EvidenceCount:      m.EvidenceCount,
		ContradictionCount: m.ContradictionCount,
		FirstObservedAt:    m.FirstObservedAt,
		LastReinforcedAt:   m.LastReinforcedAt,
		LastContradictedAt: m.LastContradictedAt,
		Status:             m.Status,
		SupersededBy:       m.SupersededBy,
		SupersedeReason:    m.SupersedeReason,
		PromptSlot:         m.PromptSlot,
		TokenCost:          m.TokenCost,
		SearchTerms:        cloneStrings(m.SearchTerms),
	}
}
