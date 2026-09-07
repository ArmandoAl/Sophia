package application

import "context"

type EmbeddingProvider interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
}
