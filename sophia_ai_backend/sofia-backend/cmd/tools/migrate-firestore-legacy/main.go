package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	"github.com/armandoalvarado/sofia-backend/internal/platform/firestoremigration"
	"google.golang.org/api/iterator"
)

var migratedCollections = []string{"users", "activities", "reminders"}

func main() {
	dryRun := flag.Bool("dry-run", true, "generate preview without writing data")
	apply := flag.Bool("apply", false, "write mapped data when dry-run is false")
	validateOnly := flag.Bool("validate-only", false, "generate preview and conflict report without writing data")
	allowManualReview := flag.Bool("allow-manual-review", false, "allow apply even when manual_review targets exist")
	targetMode := flag.String("target-mode", firestoremigration.TargetModeStaging, "target destination mode: staging, modular-only, or in-place")
	limit := flag.Int("limit", 20, "maximum legacy documents to inspect per collection")
	outputDir := flag.String("output-dir", "docs", "directory where preview files are written")
	flag.Parse()

	if err := run(context.Background(), *dryRun, *apply, *validateOnly, *allowManualReview, *targetMode, *limit, *outputDir); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, dryRun, apply, validateOnly, allowManualReview bool, targetMode string, limit int, outputDir string) error {
	if limit <= 0 {
		return errors.New("limit must be greater than zero")
	}
	targetMode = strings.TrimSpace(targetMode)
	if targetMode == "" {
		targetMode = firestoremigration.TargetModeStaging
	}
	if targetMode != firestoremigration.TargetModeStaging && targetMode != firestoremigration.TargetModeInPlace && targetMode != firestoremigration.TargetModeModular {
		return fmt.Errorf("invalid target-mode %q; use staging, modular-only, or in-place", targetMode)
	}
	if validateOnly {
		dryRun = true
		apply = false
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.PersistenceDriver != "firestore" {
		return fmt.Errorf("PERSISTENCE_DRIVER=%q; migration preview requires Firestore configuration", cfg.PersistenceDriver)
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

	docs, err := readLegacyDocuments(ctx, store.Client, limit)
	if err != nil {
		return err
	}
	preview := firestoremigration.BuildPreviewWithTargetMode(docs, limit, dryRun, apply, targetMode, time.Now())
	conflicts, err := detectTargetConflicts(ctx, store.Client, preview.Targets)
	if err != nil {
		return err
	}
	preview.Conflicts = conflicts
	if err := writePreview(outputDir, preview); err != nil {
		return err
	}

	if validateOnly {
		log.Printf("migration validation generated; no writes performed")
		return nil
	}
	if !apply || dryRun {
		log.Printf("migration preview generated; no writes performed dry_run=%t apply=%t", dryRun, apply)
		return nil
	}

	if err := firestoremigration.ValidateApplyAllowed(preview, allowManualReview); err != nil {
		return err
	}
	if err := applyTargets(ctx, store.Client, preview.Targets); err != nil {
		return err
	}
	log.Printf("migration applied non-destructively with merge writes targets=%d", len(preview.Targets))
	return nil
}

func detectTargetConflicts(ctx context.Context, client *firestore.Client, targets []firestoremigration.TargetDocument) ([]firestoremigration.Conflict, error) {
	conflicts := []firestoremigration.Conflict{}
	for _, target := range targets {
		if target.Collection == "" || target.ID == "" {
			conflicts = append(conflicts, firestoremigration.Conflict{
				Collection: target.Collection,
				ID:         target.ID,
				Severity:   "critical",
				Message:    "target collection and id are required",
			})
			continue
		}
		snapshot, err := client.Collection(target.Collection).Doc(target.ID).Get(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "not found") {
				continue
			}
			return nil, fmt.Errorf("check target %s/%s: %w", target.Collection, target.ID, err)
		}
		if !snapshot.Exists() {
			continue
		}
		severity := "medium"
		message := "target document already exists"
		switch target.Action {
		case firestoremigration.ActionCreate:
			severity = "high"
			message = "create target already exists; applying could overwrite existing modular data"
		case firestoremigration.ActionMerge:
			severity = "medium"
			message = "merge target already exists; applying can overwrite normalized fields"
		case firestoremigration.ActionReview:
			severity = "critical"
			message = "manual review target exists and must not be auto-applied"
		}
		conflicts = append(conflicts, firestoremigration.Conflict{
			Collection: target.Collection,
			ID:         target.ID,
			Severity:   severity,
			Message:    message,
		})
	}
	return conflicts, nil
}

func readLegacyDocuments(ctx context.Context, client *firestore.Client, limit int) ([]firestoremigration.LegacyDocument, error) {
	var docs []firestoremigration.LegacyDocument
	for _, collection := range migratedCollections {
		iter := client.Collection(collection).Limit(limit).Documents(ctx)
		for {
			doc, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				iter.Stop()
				return nil, fmt.Errorf("read legacy collection %s: %w", collection, err)
			}
			data := doc.Data()
			data["_document_id"] = doc.Ref.ID
			docs = append(docs, firestoremigration.LegacyDocument{
				Collection: collection,
				ID:         doc.Ref.ID,
				Data:       data,
			})
		}
		iter.Stop()
	}
	return docs, nil
}

func writePreview(outputDir string, preview firestoremigration.Preview) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	raw, err := firestoremigration.MarshalPreview(preview)
	if err != nil {
		return err
	}
	files := map[string][]byte{
		"migration_preview.json": append(raw, '\n'),
		"migration_preview.md":   []byte(firestoremigration.RenderPreviewMarkdown(preview)),
	}
	if preview.TargetMode == firestoremigration.TargetModeStaging {
		files["migration_preview_staging.json"] = append(raw, '\n')
		files["migration_preview_staging.md"] = []byte(firestoremigration.RenderPreviewMarkdown(preview))
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

func applyTargets(ctx context.Context, client *firestore.Client, targets []firestoremigration.TargetDocument) error {
	for _, target := range targets {
		if target.Action == firestoremigration.ActionReview {
			log.Printf("skip manual review target collection=%s id=%s", target.Collection, target.ID)
			continue
		}
		if target.Collection == "" || target.ID == "" {
			return fmt.Errorf("invalid target: collection and id are required")
		}
		if _, err := client.Collection(target.Collection).Doc(target.ID).Set(ctx, target.Data, firestore.MergeAll); err != nil {
			return fmt.Errorf("write %s/%s: %w", target.Collection, target.ID, err)
		}
	}
	return nil
}
