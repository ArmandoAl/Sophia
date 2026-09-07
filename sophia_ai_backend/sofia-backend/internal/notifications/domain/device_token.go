package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformWeb     = "web"
)

var (
	ErrDeviceTokenNotFound = errors.New("device token not found")
	ErrInvalidPlatform     = errors.New("invalid platform")
	ErrInvalidToken        = errors.New("token is required")
)

type DeviceToken struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Platform   string    `json:"platform"`
	Token      string    `json:"token,omitempty"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type DeviceTokenCreate struct {
	Platform string
	Token    string
}

type DeviceTokenRepository interface {
	Save(ctx context.Context, token *DeviceToken) error
	FindByID(ctx context.Context, tokenID string) (*DeviceToken, error)
	ListByUser(ctx context.Context, userID string) ([]*DeviceToken, error)
	Delete(ctx context.Context, userID, tokenID string) error
}

func NewDeviceToken(id, userID string, input DeviceTokenCreate) (*DeviceToken, error) {
	now := time.Now()
	token := &DeviceToken{
		ID:         strings.TrimSpace(id),
		UserID:     strings.TrimSpace(userID),
		Platform:   strings.ToLower(strings.TrimSpace(input.Platform)),
		Token:      strings.TrimSpace(input.Token),
		Enabled:    true,
		CreatedAt:  now,
		UpdatedAt:  now,
		LastSeenAt: now,
	}
	if err := token.Validate(); err != nil {
		return nil, err
	}
	return token, nil
}

func (t *DeviceToken) Validate() error {
	if !IsValidPlatform(t.Platform) {
		return ErrInvalidPlatform
	}
	if strings.TrimSpace(t.Token) == "" {
		return ErrInvalidToken
	}
	return nil
}

func IsValidPlatform(platform string) bool {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case PlatformIOS, PlatformAndroid, PlatformWeb:
		return true
	default:
		return false
	}
}
