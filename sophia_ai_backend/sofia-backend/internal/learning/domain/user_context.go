package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrUserContextNotFound = errors.New("user context not found")
	ErrUserContextExists   = errors.New("user context already exists")
	ErrInvalidContextKind  = errors.New("invalid context kind")
	ErrInvalidContextSlug  = errors.New("context slug is required")
	ErrInvalidContextLabel = errors.New("context label is required")
)

type UserContext struct {
	ID        string
	UserID    string
	Kind      string
	Slug      string
	Label     string
	Aliases   []string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserContextCreate struct {
	Kind    string
	Slug    string
	Label   string
	Aliases []string
}

type UserContextUpdate struct {
	Label   *string
	Aliases *[]string
	Active  *bool
}

type UserContextRepository interface {
	Create(ctx context.Context, value *UserContext) error
	List(ctx context.Context, userID string) ([]*UserContext, error)
	FindBySlug(ctx context.Context, userID, kind, slug string) (*UserContext, error)
	Update(ctx context.Context, userID, contextID string, update UserContextUpdate) (*UserContext, error)
	Archive(ctx context.Context, userID, contextID string) (*UserContext, error)
}

func NewUserContext(id, userID string, input UserContextCreate) (*UserContext, error) {
	now := time.Now().UTC()
	value := &UserContext{
		ID:        strings.TrimSpace(id),
		UserID:    strings.TrimSpace(userID),
		Kind:      strings.ToLower(strings.TrimSpace(input.Kind)),
		Slug:      strings.ToLower(strings.TrimSpace(input.Slug)),
		Label:     strings.TrimSpace(input.Label),
		Aliases:   normalizeAliases(input.Aliases),
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := value.Validate(); err != nil {
		return nil, err
	}
	return value, nil
}

func (c *UserContext) Apply(update UserContextUpdate) error {
	if update.Label != nil {
		c.Label = strings.TrimSpace(*update.Label)
	}
	if update.Aliases != nil {
		c.Aliases = normalizeAliases(*update.Aliases)
	}
	if update.Active != nil {
		c.Active = *update.Active
	}
	c.UpdatedAt = time.Now().UTC()
	return c.Validate()
}

func (c *UserContext) Archive() {
	c.Active = false
	c.UpdatedAt = time.Now().UTC()
}

func (c *UserContext) ScopeKey() string {
	return c.Kind + ":" + c.Slug
}

func (c *UserContext) Validate() error {
	if c.Kind != ScopePerson && c.Kind != ScopeMode {
		return ErrInvalidContextKind
	}
	if c.Slug == "" || strings.Contains(c.Slug, ":") {
		return ErrInvalidContextSlug
	}
	if c.Label == "" {
		return ErrInvalidContextLabel
	}
	return nil
}

func normalizeAliases(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
