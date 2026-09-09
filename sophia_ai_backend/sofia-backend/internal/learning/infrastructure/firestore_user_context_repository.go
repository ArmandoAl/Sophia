package infrastructure

import (
	"context"
	"errors"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const userContextsCollection = "user_contexts"

var _ domain.UserContextRepository = (*FirestoreUserContextRepository)(nil)

type FirestoreUserContextRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreUserContext struct {
	ID        string    `firestore:"id"`
	UserID    string    `firestore:"user_id"`
	Kind      string    `firestore:"kind"`
	Slug      string    `firestore:"slug"`
	Label     string    `firestore:"label"`
	Aliases   []string  `firestore:"aliases"`
	Active    bool      `firestore:"active"`
	CreatedAt time.Time `firestore:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at"`
}

func NewFirestoreUserContextRepository(client *firestore.Client) *FirestoreUserContextRepository {
	return &FirestoreUserContextRepository{client: client, timeout: 5 * time.Second}
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
		Aliases: cloneStrings(value.Aliases), Active: value.Active, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
}

func documentToUserContext(doc *firestore.DocumentSnapshot) (*domain.UserContext, error) {
	var stored firestoreUserContext
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return &domain.UserContext{
		ID: stored.ID, UserID: stored.UserID, Kind: stored.Kind, Slug: stored.Slug, Label: stored.Label,
		Aliases: cloneStrings(stored.Aliases), Active: stored.Active, CreatedAt: stored.CreatedAt, UpdatedAt: stored.UpdatedAt,
	}, nil
}
