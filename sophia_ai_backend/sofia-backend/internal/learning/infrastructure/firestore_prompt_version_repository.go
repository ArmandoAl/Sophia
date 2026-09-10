package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"google.golang.org/api/iterator"
)

var _ domain.PromptVersionRepository = (*FirestorePromptVersionRepository)(nil)

const promptVersionsCollection = "prompt_versions"

type FirestorePromptVersionRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestorePromptVersion struct {
	ID                     string                 `firestore:"id"`
	UserID                 string                 `firestore:"user_id"`
	Version                int                    `firestore:"version"`
	Content                string                 `firestore:"content"`
	TokenCount             int                    `firestore:"token_count"`
	BeliefIDs              []string               `firestore:"belief_ids"`
	GeneratedFromSummaryID string                 `firestore:"generated_from_summary_id"`
	CreatedAt              time.Time              `firestore:"created_at"`
	Active                 bool                   `firestore:"active"`
	Quality                firestorePromptQuality `firestore:"quality"`
}

type firestorePromptQuality struct {
	ApprovalRate   float64 `firestore:"approval_rate"`
	CorrectionRate float64 `firestore:"correction_rate"`
	SampleSize     int     `firestore:"sample_size"`
}

func NewFirestorePromptVersionRepository(client *firestore.Client) *FirestorePromptVersionRepository {
	return &FirestorePromptVersionRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestorePromptVersionRepository) CreateActive(ctx context.Context, version *domain.PromptVersion) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	col := r.client.Collection(promptVersionsCollection)
	return r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		activeIter := tx.Documents(col.Where("user_id", "==", version.UserID).Where("active", "==", true))
		defer activeIter.Stop()
		for {
			doc, err := activeIter.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				return err
			}
			var stored firestorePromptVersion
			if err := doc.DataTo(&stored); err != nil {
				return err
			}
			stored.Active = false
			if err := tx.Set(doc.Ref, stored); err != nil {
				return err
			}
		}

		next := 1
		latestIter := tx.Documents(col.Where("user_id", "==", version.UserID).OrderBy("version", firestore.Desc).Limit(1))
		defer latestIter.Stop()
		latest, err := latestIter.Next()
		if err == nil {
			var stored firestorePromptVersion
			if err := latest.DataTo(&stored); err != nil {
				return err
			}
			next = stored.Version + 1
		} else if !errors.Is(err, iterator.Done) {
			return err
		}

		version.Version = next
		version.Active = true
		version.TokenCount = domain.ApproximateTokens(version.Content)
		if version.CreatedAt.IsZero() {
			version.CreatedAt = time.Now().UTC()
		}
		return tx.Set(col.Doc(version.ID), promptVersionToDocument(version))
	})
}

func (r *FirestorePromptVersionRepository) DeactivateActive(ctx context.Context, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	col := r.client.Collection(promptVersionsCollection)
	return r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		iter := tx.Documents(col.Where("user_id", "==", userID).Where("active", "==", true))
		defer iter.Stop()
		for {
			doc, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				return nil
			}
			if err != nil {
				return err
			}
			if err := tx.Update(doc.Ref, []firestore.Update{{Path: "active", Value: false}}); err != nil {
				return err
			}
		}
	})
}

func (r *FirestorePromptVersionRepository) GetActive(ctx context.Context, userID string) (*domain.PromptVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	iter := r.client.Collection(promptVersionsCollection).
		Where("user_id", "==", userID).
		Where("active", "==", true).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if errors.Is(err, iterator.Done) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return documentToPromptVersion(doc)
}

func (r *FirestorePromptVersionRepository) ListVersions(ctx context.Context, userID string, limit int) ([]*domain.PromptVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(promptVersionsCollection).
		Where("user_id", "==", userID).
		OrderBy("version", firestore.Desc)
	if limit > 0 {
		query = query.Limit(limit)
	}
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.PromptVersion, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		version, err := documentToPromptVersion(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, version)
	}
	return result, nil
}

func promptVersionToDocument(version *domain.PromptVersion) firestorePromptVersion {
	return firestorePromptVersion{
		ID:                     version.ID,
		UserID:                 version.UserID,
		Version:                version.Version,
		Content:                version.Content,
		TokenCount:             version.TokenCount,
		BeliefIDs:              cloneStrings(version.BeliefIDs),
		GeneratedFromSummaryID: version.GeneratedFromSummaryID,
		CreatedAt:              version.CreatedAt,
		Active:                 version.Active,
		Quality: firestorePromptQuality{
			ApprovalRate:   version.Quality.ApprovalRate,
			CorrectionRate: version.Quality.CorrectionRate,
			SampleSize:     version.Quality.SampleSize,
		},
	}
}

func documentToPromptVersion(doc *firestore.DocumentSnapshot) (*domain.PromptVersion, error) {
	var stored firestorePromptVersion
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return stored.toDomain(), nil
}

func (m firestorePromptVersion) toDomain() *domain.PromptVersion {
	return &domain.PromptVersion{
		ID:                     m.ID,
		UserID:                 m.UserID,
		Version:                m.Version,
		Content:                m.Content,
		TokenCount:             m.TokenCount,
		BeliefIDs:              cloneStrings(m.BeliefIDs),
		GeneratedFromSummaryID: m.GeneratedFromSummaryID,
		CreatedAt:              m.CreatedAt,
		Active:                 m.Active,
		Quality: domain.PromptQuality{
			ApprovalRate:   m.Quality.ApprovalRate,
			CorrectionRate: m.Quality.CorrectionRate,
			SampleSize:     m.Quality.SampleSize,
		},
	}
}
