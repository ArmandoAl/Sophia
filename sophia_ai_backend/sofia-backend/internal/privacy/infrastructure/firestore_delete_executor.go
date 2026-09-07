package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type FirestoreUserDataDeleter struct {
	client  *firestore.Client
	timeout time.Duration
}

func NewFirestoreUserDataDeleter(client *firestore.Client) *FirestoreUserDataDeleter {
	return &FirestoreUserDataDeleter{client: client, timeout: 30 * time.Second}
}

func (d *FirestoreUserDataDeleter) DeleteUserDocuments(ctx context.Context, collection, userID string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	iter := d.client.Collection(collection).Where("user_id", "==", userID).Documents(ctx)
	defer iter.Stop()

	deleted := 0
	batch := d.client.Batch()
	pending := 0
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return deleted, err
		}
		batch.Delete(doc.Ref)
		pending++
		deleted++
		if pending == 450 {
			if _, err := batch.Commit(ctx); err != nil {
				return deleted, err
			}
			batch = d.client.Batch()
			pending = 0
		}
	}
	if pending > 0 {
		if _, err := batch.Commit(ctx); err != nil {
			return deleted, err
		}
	}
	return deleted, nil
}
