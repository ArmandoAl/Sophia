package jwt

import (
	"errors"
	"testing"
	"time"
)

func TestExpiredTokenRejected(t *testing.T) {
	service := New("test-secret", -time.Hour)
	token, err := service.Generate("user-1", "user@example.test", "user")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if _, err := service.Validate(token); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected expired token error, got %v", err)
	}
}
