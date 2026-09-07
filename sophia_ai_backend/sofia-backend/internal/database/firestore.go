package database

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/config"
)

type Firestore struct {
	Client *firestore.Client
}

func NewFirestore(ctx context.Context, cfg config.Config) (*Firestore, error) {
	if cfg.FirestoreProjectID == "" {
		return nil, fmt.Errorf("firestore project id is required")
	}

	var (
		client *firestore.Client
		err    error
	)
	if cfg.FirestoreDatabaseID != "" {
		client, err = firestore.NewClientWithDatabase(ctx, cfg.FirestoreProjectID, cfg.FirestoreDatabaseID)
	} else {
		client, err = firestore.NewClient(ctx, cfg.FirestoreProjectID)
	}
	if err != nil {
		return nil, fmt.Errorf("create firestore client: %w", err)
	}

	return &Firestore{Client: client}, nil
}

func (f *Firestore) Close() error {
	if f == nil || f.Client == nil {
		return nil
	}
	return f.Client.Close()
}
