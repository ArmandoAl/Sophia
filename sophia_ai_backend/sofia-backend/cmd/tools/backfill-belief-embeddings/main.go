package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	vertexinfra "github.com/armandoalvarado/sofia-backend/internal/ai/embeddings/infrastructure/vertex"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	memoryapp "github.com/armandoalvarado/sofia-backend/internal/memory/application"
	"google.golang.org/api/iterator"
)

const beliefsCollection = "user_beliefs"

type storedBelief struct {
	Statement string    `firestore:"statement"`
	Embedding []float64 `firestore:"embedding"`
}

func main() {
	dryRun := flag.Bool("dry-run", true, "preview documents without requesting or writing embeddings")
	apply := flag.Bool("apply", false, "write embeddings when dry-run is false")
	batchSize := flag.Int("batch-size", 100, "documents per Firestore batch (1-500)")
	after := flag.String("after", "", "resume after this document ID")
	limit := flag.Int("limit", 0, "maximum documents to inspect; 0 means all")
	flag.Parse()

	if err := run(context.Background(), *dryRun, *apply, *batchSize, *after, *limit); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, dryRun, apply bool, batchSize int, after string, limit int) error {
	if batchSize <= 0 || batchSize > 500 {
		return errors.New("batch-size must be between 1 and 500")
	}
	if limit < 0 {
		return errors.New("limit must be zero or greater")
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.PersistenceDriver != "firestore" {
		return fmt.Errorf("PERSISTENCE_DRIVER=%q; backfill requires Firestore configuration", cfg.PersistenceDriver)
	}
	if !cfg.EmbeddingsEnabled {
		return errors.New("EMBEDDINGS_ENABLED=true is required for the embeddings backfill")
	}

	store, err := database.NewFirestore(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close firestore: %v", err)
		}
	}()
	provider, err := vertexinfra.NewClient(ctx, cfg.GoogleCloudProject, cfg.EmbeddingsModel)
	if err != nil {
		return err
	}

	scanned, updated, checkpoint, err := backfill(ctx, store.Client, provider, dryRun, apply, batchSize, after, limit)
	if err != nil {
		return err
	}
	log.Printf("backfill scanned=%d updated=%d checkpoint_after=%q dry_run=%t apply=%t", scanned, updated, checkpoint, dryRun, apply)
	return nil
}

func backfill(ctx context.Context, client *firestore.Client, provider memoryapp.EmbeddingProvider, dryRun, apply bool, batchSize int, after string, limit int) (scanned, updated int, checkpoint string, err error) {
	for limit == 0 || scanned < limit {
		pageSize := batchSize
		if limit > 0 && limit-scanned < pageSize {
			pageSize = limit - scanned
		}
		query := client.Collection(beliefsCollection).OrderBy(firestore.DocumentID, firestore.Asc).Limit(pageSize)
		if after != "" {
			query = query.StartAfter(after)
		}
		iter := query.Documents(ctx)
		docs := make([]*firestore.DocumentSnapshot, 0, pageSize)
		for {
			doc, nextErr := iter.Next()
			if errors.Is(nextErr, iterator.Done) {
				break
			}
			if nextErr != nil {
				iter.Stop()
				return scanned, updated, after, fmt.Errorf("read beliefs: %w", nextErr)
			}
			docs = append(docs, doc)
		}
		iter.Stop()
		if len(docs) == 0 {
			return scanned, updated, after, nil
		}

		batch := client.Batch()
		writes := 0
		for _, doc := range docs {
			scanned++
			var stored storedBelief
			if err := doc.DataTo(&stored); err != nil {
				return scanned, updated, after, fmt.Errorf("decode %s: %w", doc.Ref.ID, err)
			}
			if len(stored.Embedding) > 0 {
				continue
			}
			updated++
			if !apply || dryRun {
				log.Printf("would update %s/%s", beliefsCollection, doc.Ref.ID)
				continue
			}
			embedding, err := provider.EmbedText(ctx, stored.Statement)
			if err != nil {
				return scanned, updated, after, fmt.Errorf("embed %s: %w", doc.Ref.ID, err)
			}
			if len(embedding) == 0 {
				return scanned, updated, after, fmt.Errorf("embed %s: empty vector", doc.Ref.ID)
			}
			batch.Set(doc.Ref, map[string]any{"embedding": toFloat64s(embedding)}, firestore.MergeAll)
			writes++
		}
		if writes > 0 {
			if _, err := batch.Commit(ctx); err != nil {
				return scanned, updated, after, fmt.Errorf("commit batch after %q: %w", after, err)
			}
		}
		after = docs[len(docs)-1].Ref.ID
		log.Printf("checkpoint_after=%q", after)
		if len(docs) < pageSize {
			return scanned, updated, after, nil
		}
	}
	return scanned, updated, after, nil
}

func toFloat64s(values []float32) []float64 {
	result := make([]float64, len(values))
	for i, value := range values {
		result[i] = float64(value)
	}
	return result
}
