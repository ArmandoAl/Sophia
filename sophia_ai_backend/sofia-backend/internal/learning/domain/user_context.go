package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrUserContextNotFound  = errors.New("user context not found")
	ErrUserContextExists    = errors.New("user context already exists")
	ErrInvalidContextKind   = errors.New("invalid context kind")
	ErrInvalidContextSlug   = errors.New("context slug is required")
	ErrInvalidContextLabel  = errors.New("context label is required")
	ErrInvalidCreatedFrom   = errors.New("invalid context creation source")
	ErrInvalidContextStatus = errors.New("invalid context status")
	ErrInvalidContextMerge  = errors.New("invalid context merge target")
)

const (
	ContextKindGroup                = "group"
	DefaultEntityPromotionThreshold = 3
	EntityCandidateRetention        = 30 * 24 * time.Hour

	CreatedFromExplicit = "explicit"
	CreatedFromInferred = "inferred"

	ContextStatusActive        = "active"
	ContextStatusPendingReview = "pending_review"
	ContextStatusMerged        = "merged"
	ContextStatusArchived      = "archived"
)

type UserContext struct {
	ID           string
	UserID       string
	Kind         string
	Slug         string
	Label        string
	Aliases      []string
	Relationship string
	CreatedFrom  string
	Status       string
	MergedInto   string
	ThreadsMuted bool
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserContextCreate struct {
	Kind         string
	Slug         string
	Label        string
	Aliases      []string
	Relationship string
	CreatedFrom  string
	Status       string
	ThreadsMuted bool
}

type UserContextUpdate struct {
	Kind         *string
	Label        *string
	Aliases      *[]string
	Active       *bool
	Relationship *string
	Status       *string
	ThreadsMuted *bool
}

type EntityMatch struct {
	Entity    *UserContext
	Mention   string
	Start     int
	End       int
	Ambiguous bool
}

type EntityCandidate struct {
	UserID         string
	Name           string
	NormalizedName string
	MentionCount   int
	FirstSeenAt    time.Time
	LastSeenAt     time.Time
	SampleContexts []string
}

type EntityCandidateRepository interface {
	UpsertMention(ctx context.Context, userID, name, normalizedName, sampleContext string, now time.Time) error
	List(ctx context.Context, userID string) ([]*EntityCandidate, error)
	Delete(ctx context.Context, userID, normalizedName string) error
}

type UserContextRepository interface {
	Create(ctx context.Context, value *UserContext) error
	List(ctx context.Context, userID string) ([]*UserContext, error)
	FindBySlug(ctx context.Context, userID, kind, slug string) (*UserContext, error)
	FindByID(ctx context.Context, userID, contextID string) (*UserContext, error)
	Update(ctx context.Context, userID, contextID string, update UserContextUpdate) (*UserContext, error)
	Archive(ctx context.Context, userID, contextID string) (*UserContext, error)
	Merge(ctx context.Context, userID, sourceID string, target *UserContext) (*UserContext, error)
}

func NewUserContext(id, userID string, input UserContextCreate) (*UserContext, error) {
	now := time.Now().UTC()
	createdFrom := strings.ToLower(strings.TrimSpace(input.CreatedFrom))
	if createdFrom == "" {
		createdFrom = CreatedFromExplicit
	}
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = ContextStatusActive
		if createdFrom == CreatedFromInferred {
			status = ContextStatusPendingReview
		}
	}
	value := &UserContext{
		ID:           strings.TrimSpace(id),
		UserID:       strings.TrimSpace(userID),
		Kind:         strings.ToLower(strings.TrimSpace(input.Kind)),
		Slug:         strings.ToLower(strings.TrimSpace(input.Slug)),
		Label:        strings.TrimSpace(input.Label),
		Aliases:      normalizeAliases(input.Aliases),
		Relationship: strings.TrimSpace(input.Relationship),
		CreatedFrom:  createdFrom,
		Status:       status,
		ThreadsMuted: input.ThreadsMuted,
		Active:       status == ContextStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := value.Validate(); err != nil {
		return nil, err
	}
	return value, nil
}

func (c *UserContext) Apply(update UserContextUpdate) error {
	if update.Kind != nil {
		c.Kind = strings.ToLower(strings.TrimSpace(*update.Kind))
	}
	if update.Label != nil {
		c.Label = strings.TrimSpace(*update.Label)
	}
	if update.Aliases != nil {
		c.Aliases = normalizeAliases(*update.Aliases)
	}
	if update.Active != nil {
		c.Active = *update.Active
		if c.Active {
			c.Status = ContextStatusActive
		} else {
			c.Status = ContextStatusArchived
		}
	}
	if update.Relationship != nil {
		c.Relationship = strings.TrimSpace(*update.Relationship)
	}
	if update.Status != nil {
		c.Status = strings.ToLower(strings.TrimSpace(*update.Status))
		c.Active = c.Status == ContextStatusActive
	}
	if update.ThreadsMuted != nil {
		c.ThreadsMuted = *update.ThreadsMuted
	}
	c.UpdatedAt = time.Now().UTC()
	return c.Validate()
}

func (c *UserContext) Archive() {
	c.Active = false
	c.Status = ContextStatusArchived
	c.UpdatedAt = time.Now().UTC()
}

func (c *UserContext) Merge(target *UserContext) error {
	if target == nil || target.ID == "" || target.ID == c.ID || target.UserID != c.UserID || !target.CanLoadContext() {
		return ErrInvalidContextMerge
	}
	c.Active = false
	c.Status = ContextStatusMerged
	c.MergedInto = target.ID
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *UserContext) EffectiveStatus() string {
	if c.Status != "" {
		return c.Status
	}
	if c.Active {
		return ContextStatusActive
	}
	return ContextStatusArchived
}

func (c *UserContext) EffectiveCreatedFrom() string {
	if c.CreatedFrom == "" {
		return CreatedFromExplicit
	}
	return c.CreatedFrom
}

func (c *UserContext) CanLoadContext() bool {
	status := c.EffectiveStatus()
	return status == ContextStatusActive || status == ContextStatusPendingReview
}

func (c *UserContext) ScopeKey() string {
	return c.Kind + ":" + c.Slug
}

func (c *UserContext) Validate() error {
	if !IsValidContextKind(c.Kind) {
		return ErrInvalidContextKind
	}
	if c.Slug == "" || strings.Contains(c.Slug, ":") {
		return ErrInvalidContextSlug
	}
	if c.Label == "" {
		return ErrInvalidContextLabel
	}
	if c.EffectiveCreatedFrom() != CreatedFromExplicit && c.EffectiveCreatedFrom() != CreatedFromInferred {
		return ErrInvalidCreatedFrom
	}
	switch c.EffectiveStatus() {
	case ContextStatusActive, ContextStatusPendingReview, ContextStatusMerged, ContextStatusArchived:
	default:
		return ErrInvalidContextStatus
	}
	if c.EffectiveStatus() == ContextStatusMerged && strings.TrimSpace(c.MergedInto) == "" {
		return ErrInvalidContextMerge
	}
	return nil
}

func IsValidContextKind(kind string) bool {
	return kind == ScopePerson || kind == ScopeMode || kind == ContextKindGroup
}

func IsValidContextStatus(status string) bool {
	switch status {
	case ContextStatusActive, ContextStatusPendingReview, ContextStatusMerged, ContextStatusArchived:
		return true
	default:
		return false
	}
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
