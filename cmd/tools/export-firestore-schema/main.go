package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	"github.com/armandoalvarado/sofia-backend/internal/platform/firestoreexporter"
	"google.golang.org/api/iterator"
)

func main() {
	limit := flag.Int("limit", 20, "maximum documents to sample per collection")
	outputDir := flag.String("output-dir", "docs", "directory where snapshot files are written")
	redact := flag.Bool("redact", true, "redact sensitive sample values")
	force := flag.Bool("force", false, "allow running even when PERSISTENCE_DRIVER is not firestore")
	flag.Parse()

	if err := run(context.Background(), *limit, *outputDir, *redact, *force); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, limit int, outputDir string, redact bool, force bool) error {
	if limit <= 0 {
		return errors.New("limit must be greater than zero")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.PersistenceDriver != "firestore" && !force {
		return fmt.Errorf("PERSISTENCE_DRIVER=%q; set PERSISTENCE_DRIVER=firestore or pass -force for explicit schema export", cfg.PersistenceDriver)
	}
	if cfg.FirestoreProjectID == "" {
		return errors.New("FIRESTORE_PROJECT_ID is required")
	}

	db, err := database.NewFirestore(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("close firestore client: %v", closeErr)
		}
	}()

	collections, err := inspectFirestore(ctx, db.Client, limit, redact)
	if err != nil {
		return err
	}

	expected, err := loadExpectedModel(outputDir)
	if err != nil {
		return err
	}

	snapshot := firestoreexporter.BuildSnapshot(cfg.FirestoreProjectID, collections, expected, time.Now())
	return writeOutputs(outputDir, snapshot)
}

func inspectFirestore(ctx context.Context, client *firestore.Client, limit int, redact bool) ([]firestoreexporter.CollectionSnapshot, error) {
	var collections []firestoreexporter.CollectionSnapshot
	iter := client.Collections(ctx)
	for {
		collectionRef, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list root collections: %w", err)
		}

		docs, subcollections, err := sampleCollection(ctx, collectionRef, limit)
		if err != nil {
			return nil, fmt.Errorf("sample collection %s: %w", collectionRef.ID, err)
		}
		collection := firestoreexporter.InferCollection(collectionRef.ID, docs, redact)
		collection.Subcollections = subcollections
		collections = append(collections, collection)
	}
	return collections, nil
}

func sampleCollection(ctx context.Context, collectionRef *firestore.CollectionRef, limit int) ([]map[string]any, []string, error) {
	var docs []map[string]any
	subcollectionSet := map[string]struct{}{}

	iter := collectionRef.Limit(limit).Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		data := doc.Data()
		data["_document_id"] = doc.Ref.ID
		docs = append(docs, data)

		subIter := doc.Ref.Collections(ctx)
		for {
			subcollection, subErr := subIter.Next()
			if errors.Is(subErr, iterator.Done) {
				break
			}
			if subErr != nil {
				return nil, nil, subErr
			}
			subcollectionSet[subcollection.ID] = struct{}{}
		}
	}

	subcollections := make([]string, 0, len(subcollectionSet))
	for name := range subcollectionSet {
		subcollections = append(subcollections, name)
	}
	return docs, subcollections, nil
}

func loadExpectedModel(outputDir string) (firestoreexporter.ExpectedModel, error) {
	path := filepath.Join(outputDir, "data_model_expected.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return firestoreexporter.ExpectedModel{}, fmt.Errorf("read expected data model %s: %w", path, err)
	}
	return firestoreexporter.ParseExpectedModel(string(content)), nil
}

func writeOutputs(outputDir string, snapshot firestoreexporter.Snapshot) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	snapshotJSON, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	files := map[string][]byte{
		"database_snapshot.json":          append(snapshotJSON, '\n'),
		"database_schema.md":              []byte(firestoreexporter.RenderSchemaMarkdown(snapshot)),
		"database_collections_summary.md": []byte(firestoreexporter.RenderCollectionsSummary(snapshot)),
	}
	for name, content := range files {
		path := filepath.Join(outputDir, name)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		log.Printf("wrote %s", path)
	}
	return nil
}
