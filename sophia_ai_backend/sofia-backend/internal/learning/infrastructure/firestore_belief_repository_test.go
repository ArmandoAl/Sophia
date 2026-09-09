package infrastructure

import (
	"testing"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
)

func TestBeliefEmbeddingFirestoreConversion(t *testing.T) {
	original := &domain.Belief{Embedding: []float32{0.25, 0.75}, Scope: domain.ScopePerson, ScopeKey: "person:maria"}
	stored := beliefToDocument(original)
	if len(stored.Embedding) != 2 || stored.Embedding[0] != 0.25 || stored.Embedding[1] != 0.75 {
		t.Fatalf("unexpected stored embedding: %v", stored.Embedding)
	}
	roundTrip := stored.toDomain()
	if len(roundTrip.Embedding) != 2 || roundTrip.Embedding[0] != 0.25 || roundTrip.Embedding[1] != 0.75 {
		t.Fatalf("unexpected domain embedding: %v", roundTrip.Embedding)
	}
	if roundTrip.Scope != domain.ScopePerson || roundTrip.ScopeKey != "person:maria" {
		t.Fatalf("unexpected scope: %q %q", roundTrip.Scope, roundTrip.ScopeKey)
	}
}
