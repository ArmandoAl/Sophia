package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/tools/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const toolsCollection = "tool_definitions"

type FirestoreToolDefinitionRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreTool struct {
	ID                   string    `firestore:"id"`
	Name                 string    `firestore:"name"`
	Description          string    `firestore:"description"`
	Category             string    `firestore:"category"`
	InputSchema          string    `firestore:"input_schema"`
	OutputSchema         string    `firestore:"output_schema"`
	RequiresConfirmation bool      `firestore:"requires_confirmation"`
	Enabled              bool      `firestore:"enabled"`
	Reversible           bool      `firestore:"reversible"`
	MinAutonomyLevel     string    `firestore:"min_autonomy_level"`
	CreatedAt            time.Time `firestore:"created_at"`
	UpdatedAt            time.Time `firestore:"updated_at"`
}

func NewFirestoreToolDefinitionRepository(client *firestore.Client) *FirestoreToolDefinitionRepository {
	return &FirestoreToolDefinitionRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreToolDefinitionRepository) Save(ctx context.Context, tool *domain.ToolDefinition) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(toolsCollection).Doc(tool.Name).Set(ctx, toolToDocument(tool))
	return err
}

func (r *FirestoreToolDefinitionRepository) FindByName(ctx context.Context, name string) (*domain.ToolDefinition, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc, err := r.client.Collection(toolsCollection).Doc(name).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	return documentToTool(doc)
}

func (r *FirestoreToolDefinitionRepository) List(ctx context.Context, enabledOnly bool) ([]*domain.ToolDefinition, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(toolsCollection).Query
	if enabledOnly {
		query = query.Where("enabled", "==", true)
	}
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.ToolDefinition, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		tool, err := documentToTool(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, tool)
	}
	return result, nil
}

func toolToDocument(tool *domain.ToolDefinition) firestoreTool {
	return firestoreTool{ID: tool.ID, Name: tool.Name, Description: tool.Description, Category: tool.Category, InputSchema: string(tool.InputSchema), OutputSchema: string(tool.OutputSchema), RequiresConfirmation: tool.RequiresConfirmation, Enabled: tool.Enabled, Reversible: tool.Reversible, MinAutonomyLevel: tool.MinAutonomyLevel, CreatedAt: tool.CreatedAt, UpdatedAt: tool.UpdatedAt}
}

func documentToTool(doc *firestore.DocumentSnapshot) (*domain.ToolDefinition, error) {
	var stored firestoreTool
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return &domain.ToolDefinition{ID: stored.ID, Name: stored.Name, Description: stored.Description, Category: stored.Category, InputSchema: []byte(stored.InputSchema), OutputSchema: []byte(stored.OutputSchema), RequiresConfirmation: stored.RequiresConfirmation, Enabled: stored.Enabled, Reversible: stored.Reversible, MinAutonomyLevel: stored.MinAutonomyLevel, CreatedAt: stored.CreatedAt, UpdatedAt: stored.UpdatedAt}, nil
}
