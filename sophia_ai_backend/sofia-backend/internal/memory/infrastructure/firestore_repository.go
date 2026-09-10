package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ domain.MemoryRepository = (*FirestoreMemoryRepository)(nil)

const memoriesCollection = "memories"

type FirestoreMemoryRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreMemory struct {
	ID             string     `firestore:"id"`
	UserID         string     `firestore:"user_id"`
	Type           string     `firestore:"type"`
	Title          string     `firestore:"title"`
	Content        string     `firestore:"content"`
	Summary        string     `firestore:"summary"`
	Source         string     `firestore:"source"`
	SourceID       string     `firestore:"source_id"`
	Confidence     string     `firestore:"confidence"`
	Importance     string     `firestore:"importance"`
	Tags           []string   `firestore:"tags"`
	Visibility     string     `firestore:"visibility"`
	Status         string     `firestore:"status"`
	SearchTerms    []string   `firestore:"search_terms"`
	CreatedAt      time.Time  `firestore:"created_at"`
	UpdatedAt      time.Time  `firestore:"updated_at"`
	LastAccessedAt *time.Time `firestore:"last_accessed_at,omitempty"`
}

func NewFirestoreMemoryRepository(client *firestore.Client) *FirestoreMemoryRepository {
	return &FirestoreMemoryRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreMemoryRepository) Create(ctx context.Context, memory *domain.Memory) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.client.Collection(memoriesCollection).Doc(memory.ID).Set(ctx, memoryToDocument(memory))
	return err
}

func (r *FirestoreMemoryRepository) FindByID(ctx context.Context, userID, memoryID string) (*domain.Memory, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	doc, err := r.client.Collection(memoriesCollection).Doc(memoryID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	memory, err := documentToMemory(doc)
	if err != nil {
		return nil, err
	}
	if memory.UserID != userID || memory.Status == domain.StatusDeleted {
		return nil, nil
	}
	return memory, nil
}

func (r *FirestoreMemoryRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Memory, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(memoriesCollection).Where("user_id", "==", filter.UserID)
	if filter.Status != "" {
		query = query.Where("status", "==", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type", "==", filter.Type)
	}
	if filter.Importance != "" {
		query = query.Where("importance", "==", filter.Importance)
	}
	query = query.OrderBy("created_at", firestore.Desc).OrderBy("id", firestore.Desc)
	if filter.Cursor != "" {
		cursorDoc, err := r.client.Collection(memoriesCollection).Doc(filter.Cursor).Get(ctx)
		if err == nil {
			var cursor firestoreMemory
			if err := cursorDoc.DataTo(&cursor); err == nil && cursor.UserID == filter.UserID {
				query = query.StartAfter(cursor.CreatedAt, cursor.ID)
			}
		} else if status.Code(err) != codes.NotFound {
			return nil, err
		}
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	return r.collect(ctx, query, filter)
}

func (r *FirestoreMemoryRepository) Update(ctx context.Context, memory *domain.Memory) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.client.Collection(memoriesCollection).Doc(memory.ID).Set(ctx, memoryToDocument(memory))
	return err
}

func (r *FirestoreMemoryRepository) Archive(ctx context.Context, userID, memoryID string) (*domain.Memory, error) {
	return r.change(ctx, userID, memoryID, func(memory *domain.Memory) error {
		return memory.MarkArchived()
	})
}

func (r *FirestoreMemoryRepository) DeleteSoft(ctx context.Context, userID, memoryID string) (*domain.Memory, error) {
	return r.change(ctx, userID, memoryID, func(memory *domain.Memory) error {
		return memory.MarkDeleted()
	})
}

func (r *FirestoreMemoryRepository) SearchBasic(ctx context.Context, filter domain.SearchFilter) ([]*domain.Memory, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = domain.DefaultSearchLimit()
	}
	listFilter := domain.ListFilter{
		UserID:     filter.UserID,
		Type:       filter.Type,
		Tag:        filter.Tag,
		Importance: filter.Importance,
		Status:     domain.StatusActive,
		Limit:      limit,
	}
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(memoriesCollection).
		Where("user_id", "==", filter.UserID).
		Where("status", "==", domain.StatusActive)
	if filter.Type != "" {
		query = query.Where("type", "==", filter.Type)
	}
	if filter.Importance != "" {
		query = query.Where("importance", "==", filter.Importance)
	}
	query = query.Limit(limit)
	memories, err := r.collect(ctx, query, listFilter)
	if err != nil {
		return nil, err
	}
	result := memories[:0]
	for _, memory := range memories {
		if matchesSearchFilter(memory, filter) {
			result = append(result, memory)
		}
	}
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *FirestoreMemoryRepository) SearchByTerms(ctx context.Context, userID string, terms []string, limit int) ([]*domain.Memory, error) {
	terms = domain.CapSearchTerms(terms)
	if len(terms) == 0 {
		return []*domain.Memory{}, nil
	}
	if limit <= 0 {
		limit = domain.DefaultSearchLimit()
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(memoriesCollection).
		Where("user_id", "==", userID).
		Where("status", "==", domain.StatusActive).
		Where("search_terms", "array-contains-any", terms).
		Limit(limit * 4)

	memories, err := r.collectQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	return domain.RankByTermMatches(memories, terms, limit), nil
}

func (r *FirestoreMemoryRepository) TouchAccessed(ctx context.Context, userID, memoryID string) error {
	memory, err := r.FindByID(ctx, userID, memoryID)
	if err != nil {
		return err
	}
	if memory == nil {
		return domain.ErrMemoryNotFound
	}
	memory.TouchAccessed()
	return r.Update(ctx, memory)
}

func (r *FirestoreMemoryRepository) change(ctx context.Context, userID, memoryID string, mutate func(*domain.Memory) error) (*domain.Memory, error) {
	memory, err := r.FindByID(ctx, userID, memoryID)
	if err != nil {
		return nil, err
	}
	if memory == nil {
		return nil, domain.ErrMemoryNotFound
	}
	if err := mutate(memory); err != nil {
		return nil, err
	}
	if err := r.Update(ctx, memory); err != nil {
		return nil, err
	}
	return memory, nil
}

func (r *FirestoreMemoryRepository) collect(ctx context.Context, query firestore.Query, filter domain.ListFilter) ([]*domain.Memory, error) {
	iter := query.Documents(ctx)
	defer iter.Stop()

	result := make([]*domain.Memory, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		memory, err := documentToMemory(doc)
		if err != nil {
			return nil, err
		}
		if !matchesListFilter(memory, filter) {
			continue
		}
		result = append(result, memory)
	}
	sortMemories(result)
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *FirestoreMemoryRepository) collectQuery(ctx context.Context, query firestore.Query) ([]*domain.Memory, error) {
	iter := query.Documents(ctx)
	defer iter.Stop()

	result := make([]*domain.Memory, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		memory, err := documentToMemory(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, memory)
	}
	return result, nil
}

func memoryToDocument(memory *domain.Memory) firestoreMemory {
	return firestoreMemory{
		ID:             memory.ID,
		UserID:         memory.UserID,
		Type:           memory.Type,
		Title:          memory.Title,
		Content:        memory.Content,
		Summary:        memory.Summary,
		Source:         memory.Source,
		SourceID:       memory.SourceID,
		Confidence:     memory.Confidence,
		Importance:     memory.Importance,
		Tags:           append([]string(nil), memory.Tags...),
		Visibility:     memory.Visibility,
		Status:         memory.Status,
		SearchTerms:    append([]string(nil), memory.SearchTerms...),
		CreatedAt:      memory.CreatedAt,
		UpdatedAt:      memory.UpdatedAt,
		LastAccessedAt: memory.LastAccessedAt,
	}
}

func documentToMemory(doc *firestore.DocumentSnapshot) (*domain.Memory, error) {
	var stored firestoreMemory
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return stored.toDomain(), nil
}

func (m firestoreMemory) toDomain() *domain.Memory {
	return &domain.Memory{
		ID:             m.ID,
		UserID:         m.UserID,
		Type:           m.Type,
		Title:          m.Title,
		Content:        m.Content,
		Summary:        m.Summary,
		Source:         m.Source,
		SourceID:       m.SourceID,
		Confidence:     m.Confidence,
		Importance:     m.Importance,
		Tags:           append([]string(nil), m.Tags...),
		Visibility:     m.Visibility,
		Status:         m.Status,
		SearchTerms:    append([]string(nil), m.SearchTerms...),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		LastAccessedAt: m.LastAccessedAt,
	}
}
