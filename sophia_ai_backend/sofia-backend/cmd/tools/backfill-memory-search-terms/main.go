package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	"google.golang.org/api/iterator"
)

const memoriesCollection = "memories"

type memorySearchBackfill struct {
	Title   string   `firestore:"title"`
	Summary string   `firestore:"summary"`
	Content string   `firestore:"content"`
	Tags    []string `firestore:"tags"`
}

func main() {
	dryRun := flag.Bool("dry-run", true, "generate preview without writing data")
	apply := flag.Bool("apply", false, "write search_terms when dry-run is false")
	limit := flag.Int("limit", 0, "maximum memories to inspect; 0 means all")
	flag.Parse()

	if err := run(context.Background(), *dryRun, *apply, *limit); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, dryRun, apply bool, limit int) error {
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

	store, err := database.NewFirestore(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close firestore: %v", err)
		}
	}()

	scanned, updated, err := backfillSearchTerms(ctx, store.Client, dryRun, apply, limit)
	if err != nil {
		return err
	}
	if !apply || dryRun {
		log.Printf("backfill preview scanned=%d would_update=%d dry_run=%t apply=%t", scanned, updated, dryRun, apply)
		return nil
	}
	log.Printf("backfill applied scanned=%d updated=%d", scanned, updated)
	return nil
}

func backfillSearchTerms(ctx context.Context, client *firestore.Client, dryRun, apply bool, limit int) (scanned, updated int, err error) {
	iter := client.Collection(memoriesCollection).Documents(ctx)
	if limit > 0 {
		iter = client.Collection(memoriesCollection).Limit(limit).Documents(ctx)
	}
	defer iter.Stop()

	for {
		doc, nextErr := iter.Next()
		if errors.Is(nextErr, iterator.Done) {
			break
		}
		if nextErr != nil {
			return scanned, updated, fmt.Errorf("read memories: %w", nextErr)
		}
		scanned++

		var stored memorySearchBackfill
		if err := doc.DataTo(&stored); err != nil {
			return scanned, updated, fmt.Errorf("decode %s: %w", doc.Ref.ID, err)
		}
		terms := memorydomain.ExtractTerms(strings.Join([]string{stored.Title, stored.Summary, stored.Content, strings.Join(stored.Tags, " ")}, " "), 40)
		updated++
		if !apply || dryRun {
			log.Printf("would update memories/%s search_terms=%d", doc.Ref.ID, len(terms))
			continue
		}
		if _, err := doc.Ref.Set(ctx, map[string]any{"search_terms": terms}, firestore.MergeAll); err != nil {
			return scanned, updated, fmt.Errorf("write %s: %w", doc.Ref.ID, err)
		}
		log.Printf("updated memories/%s search_terms=%d", doc.Ref.ID, len(terms))
	}
	return scanned, updated, nil
}
