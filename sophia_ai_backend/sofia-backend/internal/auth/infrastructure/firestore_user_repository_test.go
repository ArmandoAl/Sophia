package infrastructure

import (
	"testing"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/auth/domain"
)

func TestFirestoreUserMapping(t *testing.T) {
	now := time.Now().UTC()
	user := &domain.User{
		ID:        "user-1",
		Name:      "Armando",
		Email:     " ARMANDO@example.com ",
		Password:  "hash",
		Role:      "user",
		CreatedAt: now,
		UpdatedAt: now,
	}

	stored := domainToFirestoreUser(user)
	if stored.Email != "armando@example.com" {
		t.Fatalf("expected normalized email, got %q", stored.Email)
	}
	if stored.PasswordHash != "hash" {
		t.Fatal("expected password hash to be mapped")
	}

	roundTrip := firestoreUserToDomain(stored)
	if roundTrip.Email != "armando@example.com" || roundTrip.Password != "hash" || roundTrip.ID != "user-1" {
		t.Fatalf("unexpected round trip user: %+v", roundTrip)
	}
}
