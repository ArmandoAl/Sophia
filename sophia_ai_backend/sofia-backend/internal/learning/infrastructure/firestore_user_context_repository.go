package infrastructure

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	userContextsCollection     = "user_contexts"
	entityCandidatesCollection = "entity_candidates"
)

var _ domain.UserContextRepository = (*FirestoreUserContextRepository)(nil)
var _ domain.EntityCandidateRepository = (*FirestoreEntityCandidateRepository)(nil)

type FirestoreUserContextRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type FirestoreEntityCandidateRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreUserContext struct {
	ID           string    `firestore:"id"`
	UserID       string    `firestore:"user_id"`
	Kind         string    `firestore:"kind"`
	Slug         string    `firestore:"slug"`
	Label        string    `firestore:"label"`
	Aliases      []string  `firestore:"aliases"`
	Relationship string    `firestore:"relationship"`
	CreatedFrom  string    `firestore:"created_from"`
	Status       string    `firestore:"status"`
	MergedInto   string    `firestore:"merged_into"`
	ThreadsMuted bool      `firestore:"threads_muted"`
	Active       bool      `firestore:"active"`
	CreatedAt    time.Time `firestore:"created_at"`
	UpdatedAt    time.Time `firestore:"updated_at"`
}

type firestoreEntityCandidate struct {
	UserID         string    `firestore:"user_id"`
	Name           string    `firestore:"name"`
	NormalizedName string    `firestore:"normalized_name"`
	MentionCount   int       `firestore:"mention_count"`
	FirstSeenAt    time.Time `firestore:"first_seen_at"`
	LastSeenAt     time.Time `firestore:"last_seen_at"`
	SampleContexts []string  `firestore:"sample_contexts"`
}

func NewFirestoreUserContextRepository(client *firestore.Client) *FirestoreUserContextRepository {
	return &FirestoreUserContextRepository{client: client, timeout: 5 * time.Second}
}

func NewFirestoreEntityCandidateRepository(client *firestore.Client) *FirestoreEntityCandidateRepository {
	return &FirestoreEntityCandidateRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreEntityCandidateRepository) UpsertMention(ctx context.Context, userID, name, normalizedName, sampleContext string, now time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	ref := r.client.Collection(entityCandidatesCollection).Doc(entityCandidateID(userID, normalizedName))
	return r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		stored := firestoreEntityCandidate{UserID: userID, Name: name, NormalizedName: normalizedName, FirstSeenAt: now}
		doc, err := tx.Get(ref)
		if err == nil {
			if err := doc.DataTo(&stored); err != nil {
				return err
			}
		} else if status.Code(err) != codes.NotFound {
			return err
		}
		stored.MentionCount++
		stored.LastSeenAt = now
		if sampleContext != "" && len(stored.SampleContexts) < 3 && !containsString(stored.SampleContexts, sampleContext) {
			stored.SampleContexts = append(stored.SampleContexts, sampleContext)
		}
		return tx.Set(ref, stored)
	})
}

func (r *FirestoreEntityCandidateRepository) List(ctx context.Context, userID string) ([]*domain.EntityCandidate, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	iter := r.client.Collection(entityCandidatesCollection).Where("user_id", "==", userID).Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.EntityCandidate, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var stored firestoreEntityCandidate
		if err := doc.DataTo(&stored); err != nil {
			return nil, err
		}
		result = append(result, &domain.EntityCandidate{
			UserID: stored.UserID, Name: stored.Name, NormalizedName: stored.NormalizedName,
			MentionCount: stored.MentionCount, FirstSeenAt: stored.FirstSeenAt, LastSeenAt: stored.LastSeenAt,
			SampleContexts: cloneStrings(stored.SampleContexts),
		})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].LastSeenAt.After(result[j].LastSeenAt) })
	return result, nil
}

func (r *FirestoreEntityCandidateRepository) Delete(ctx context.Context, userID, normalizedName string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(entityCandidatesCollection).Doc(entityCandidateID(userID, normalizedName)).Delete(ctx)
	return err
}

func entityCandidateID(userID, normalizedName string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(userID+"\x00"+normalizedName)))
}

func (r *FirestoreUserContextRepository) Create(ctx context.Context, value *domain.UserContext) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(userContextsCollection).Doc(value.ID).Create(ctx, userContextToDocument(value))
	if status.Code(err) == codes.AlreadyExists {
		return domain.ErrUserContextExists
	}
	return err
}

func (r *FirestoreUserContextRepository) List(ctx context.Context, userID string) ([]*domain.UserContext, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	iter := r.client.Collection(userContextsCollection).Where("user_id", "==", userID).Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.UserContext, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		value, err := documentToUserContext(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].UpdatedAt.After(result[j].UpdatedAt) })
	return result, nil
}

func (r *FirestoreUserContextRepository) FindBySlug(ctx context.Context, userID, kind, slug string) (*domain.UserContext, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	iter := r.client.Collection(userContextsCollection).
		Where("user_id", "==", userID).
		Where("kind", "==", kind).
		Where("slug", "==", slug).
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
	return documentToUserContext(doc)
}

func (r *FirestoreUserContextRepository) FindByID(ctx context.Context, userID, contextID string) (*domain.UserContext, error) {
	return r.findByID(ctx, userID, contextID)
}

func (r *FirestoreUserContextRepository) Update(ctx context.Context, userID, contextID string, update domain.UserContextUpdate) (*domain.UserContext, error) {
	value, err := r.findByID(ctx, userID, contextID)
	if err != nil {
		return nil, err
	}
	if err := value.Apply(update); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	if _, err := r.client.Collection(userContextsCollection).Doc(contextID).Set(ctx, userContextToDocument(value)); err != nil {
		return nil, err
	}
	return value, nil
}

func (r *FirestoreUserContextRepository) Archive(ctx context.Context, userID, contextID string) (*domain.UserContext, error) {
	value, err := r.findByID(ctx, userID, contextID)
	if err != nil {
		return nil, err
	}
	value.Archive()
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	if _, err := r.client.Collection(userContextsCollection).Doc(contextID).Set(ctx, userContextToDocument(value)); err != nil {
		return nil, err
	}
	return value, nil
}

func (r *FirestoreUserContextRepository) Merge(ctx context.Context, userID, sourceID string, target *domain.UserContext) (*domain.UserContext, error) {
	value, err := r.findByID(ctx, userID, sourceID)
	if err != nil {
		return nil, err
	}
	if err := value.Merge(target); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	if _, err := r.client.Collection(userContextsCollection).Doc(sourceID).Set(ctx, userContextToDocument(value)); err != nil {
		return nil, err
	}
	return value, nil
}

func (r *FirestoreUserContextRepository) findByID(ctx context.Context, userID, contextID string) (*domain.UserContext, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc, err := r.client.Collection(userContextsCollection).Doc(contextID).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, domain.ErrUserContextNotFound
	}
	if err != nil {
		return nil, err
	}
	value, err := documentToUserContext(doc)
	if err != nil {
		return nil, err
	}
	if value.UserID != userID {
		return nil, domain.ErrUserContextNotFound
	}
	return value, nil
}

func userContextToDocument(value *domain.UserContext) firestoreUserContext {
	return firestoreUserContext{
		ID: value.ID, UserID: value.UserID, Kind: value.Kind, Slug: value.Slug, Label: value.Label,
		Aliases: cloneStrings(value.Aliases), Relationship: value.Relationship, CreatedFrom: value.EffectiveCreatedFrom(),
		Status: value.EffectiveStatus(), MergedInto: value.MergedInto, ThreadsMuted: value.ThreadsMuted, Active: value.Active, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
}

func documentToUserContext(doc *firestore.DocumentSnapshot) (*domain.UserContext, error) {
	var stored firestoreUserContext
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return &domain.UserContext{
		ID: stored.ID, UserID: stored.UserID, Kind: stored.Kind, Slug: stored.Slug, Label: stored.Label,
		Aliases: cloneStrings(stored.Aliases), Relationship: stored.Relationship, CreatedFrom: stored.CreatedFrom,
		Status: stored.Status, MergedInto: stored.MergedInto, ThreadsMuted: stored.ThreadsMuted, Active: stored.Active, CreatedAt: stored.CreatedAt, UpdatedAt: stored.UpdatedAt,
	}, nil
}
