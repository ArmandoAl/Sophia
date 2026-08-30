package application

import (
	"context"
	"testing"

	"github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	"github.com/armandoalvarado/sofia-backend/internal/notifications/infrastructure"
)

func TestRegisterDeviceToken(t *testing.T) {
	service := NewService(infrastructure.NewInMemoryDeviceTokenRepository())

	token, err := service.RegisterDeviceToken(context.Background(), "user-1", domain.DeviceTokenCreate{Platform: domain.PlatformIOS, Token: "fcm-token-123456"})
	if err != nil {
		t.Fatalf("RegisterDeviceToken returned error: %v", err)
	}
	if token.UserID != "user-1" || token.Platform != domain.PlatformIOS || !token.Enabled {
		t.Fatalf("unexpected token: %+v", token)
	}
}

func TestListDeviceTokensOnlyOwn(t *testing.T) {
	repo := infrastructure.NewInMemoryDeviceTokenRepository()
	service := NewService(repo)
	if _, err := service.RegisterDeviceToken(context.Background(), "user-1", domain.DeviceTokenCreate{Platform: domain.PlatformIOS, Token: "own-token-123456"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RegisterDeviceToken(context.Background(), "user-2", domain.DeviceTokenCreate{Platform: domain.PlatformAndroid, Token: "other-token-123456"}); err != nil {
		t.Fatal(err)
	}

	tokens, err := service.ListDeviceTokens(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ListDeviceTokens returned error: %v", err)
	}
	if len(tokens) != 1 || tokens[0].UserID != "user-1" {
		t.Fatalf("expected only own tokens, got %+v", tokens)
	}
}

func TestDeleteDeviceTokenOnlyOwn(t *testing.T) {
	repo := infrastructure.NewInMemoryDeviceTokenRepository()
	service := NewService(repo)
	token, err := service.RegisterDeviceToken(context.Background(), "user-1", domain.DeviceTokenCreate{Platform: domain.PlatformWeb, Token: "own-token-123456"})
	if err != nil {
		t.Fatal(err)
	}

	if err := service.DeleteDeviceToken(context.Background(), "user-2", token.ID); err == nil {
		t.Fatal("expected deleting another user's token to fail")
	}
	if err := service.DeleteDeviceToken(context.Background(), "user-1", token.ID); err != nil {
		t.Fatalf("DeleteDeviceToken returned error: %v", err)
	}
	tokens, _ := service.ListDeviceTokens(context.Background(), "user-1")
	if len(tokens) != 0 {
		t.Fatalf("expected token deleted, got %+v", tokens)
	}
}
