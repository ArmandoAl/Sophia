package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	TypePreference   = "preference"
	TypeFact         = "fact"
	TypeEpisodic     = "episodic"
	TypeSemantic     = "semantic"
	TypeInstruction  = "instruction"
	TypeRelationship = "relationship"
	TypeProject      = "project"
	TypeSystemNote   = "system_note"

	SourceUserManual   = "user_manual"
	SourceConversation = "conversation"
	SourceActivity     = "activity"
	SourceReflection   = "reflection"
	SourceImported     = "imported"
	SourceAIInferred   = "ai_inferred"

	VisibilityPrivate          = "private"
	VisibilityAssistantContext = "assistant_context"
	VisibilityArchived         = "archived"

	StatusActive   = "active"
	StatusArchived = "archived"
	StatusDeleted  = "deleted"

	ConfidenceLow    = "low"
	ConfidenceMedium = "medium"
	ConfidenceHigh   = "high"

	ImportanceLow      = "low"
	ImportanceMedium   = "medium"
	ImportanceHigh     = "high"
	ImportanceCritical = "critical"
)

var (
	ErrMemoryNotFound     = errors.New("memory not found")
	ErrMemoryDisabled     = errors.New("memory is disabled for this user")
	ErrInvalidType        = errors.New("invalid memory type")
	ErrInvalidTitle       = errors.New("title is required")
	ErrInvalidContent     = errors.New("content is required")
	ErrInvalidSource      = errors.New("invalid memory source")
	ErrInvalidVisibility  = errors.New("invalid memory visibility")
	ErrInvalidStatus      = errors.New("invalid memory status")
	ErrInvalidConfidence  = errors.New("invalid confidence")
	ErrInvalidImportance  = errors.New("invalid importance")
	ErrInvalidLimit       = errors.New("limit must be positive")
	ErrInvalidSearchQuery = errors.New("search query or filter is required")
)

type Memory struct {
	ID             string
	UserID         string
	Type           string
	Title          string
	Content        string
	Summary        string
	Source         string
	SourceID       string
	Confidence     string
	Importance     string
	Tags           []string
	Visibility     string
	Status         string
	SearchTerms    []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastAccessedAt *time.Time
}

type MemoryCreate struct {
	Type       string
	Title      string
	Content    string
	Summary    string
	Source     string
	SourceID   string
	Confidence string
	Importance string
	Tags       []string
	Visibility string
}

type MemoryUpdate struct {
	Type       *string
	Title      *string
	Content    *string
	Summary    *string
	Source     *string
	SourceID   *string
	Confidence *string
	Importance *string
	Tags       *[]string
	Visibility *string
	Status     *string
}

type ListFilter struct {
	UserID     string
	Type       string
	Tag        string
	Importance string
	Status     string
	Limit      int
	Cursor     string
}

type SearchFilter struct {
	UserID     string
	Query      string
	Type       string
	Tag        string
	Importance string
	Limit      int
}

type MemoryRepository interface {
	Create(ctx context.Context, memory *Memory) error
	FindByID(ctx context.Context, userID, memoryID string) (*Memory, error)
	List(ctx context.Context, filter ListFilter) ([]*Memory, error)
	Update(ctx context.Context, memory *Memory) error
	Archive(ctx context.Context, userID, memoryID string) (*Memory, error)
	DeleteSoft(ctx context.Context, userID, memoryID string) (*Memory, error)
	SearchBasic(ctx context.Context, filter SearchFilter) ([]*Memory, error)
	SearchByTerms(ctx context.Context, userID string, terms []string, limit int) ([]*Memory, error)
	TouchAccessed(ctx context.Context, userID, memoryID string) error
}

func NewMemory(id, userID string, input MemoryCreate) (*Memory, error) {
	now := time.Now()
	memory := &Memory{
		ID:         id,
		UserID:     userID,
		Type:       defaultString(input.Type, TypeFact),
		Title:      strings.TrimSpace(input.Title),
		Content:    strings.TrimSpace(input.Content),
		Summary:    strings.TrimSpace(input.Summary),
		Source:     defaultString(input.Source, SourceUserManual),
		SourceID:   strings.TrimSpace(input.SourceID),
		Confidence: defaultString(input.Confidence, ConfidenceMedium),
		Importance: defaultString(input.Importance, ImportanceMedium),
		Tags:       normalizeTags(input.Tags),
		Visibility: defaultString(input.Visibility, VisibilityAssistantContext),
		Status:     StatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	memory.refreshSearchTerms()
	if err := memory.Validate(); err != nil {
		return nil, err
	}
	return memory, nil
}

func (m *Memory) Apply(update MemoryUpdate) error {
	if update.Type != nil {
		m.Type = strings.TrimSpace(*update.Type)
	}
	if update.Title != nil {
		m.Title = strings.TrimSpace(*update.Title)
	}
	if update.Content != nil {
		m.Content = strings.TrimSpace(*update.Content)
	}
	if update.Summary != nil {
		m.Summary = strings.TrimSpace(*update.Summary)
	}
	if update.Source != nil {
		m.Source = strings.TrimSpace(*update.Source)
	}
	if update.SourceID != nil {
		m.SourceID = strings.TrimSpace(*update.SourceID)
	}
	if update.Confidence != nil {
		m.Confidence = strings.TrimSpace(*update.Confidence)
	}
	if update.Importance != nil {
		m.Importance = strings.TrimSpace(*update.Importance)
	}
	if update.Tags != nil {
		m.Tags = normalizeTags(*update.Tags)
	}
	if update.Visibility != nil {
		m.Visibility = strings.TrimSpace(*update.Visibility)
	}
	if update.Status != nil {
		m.Status = strings.TrimSpace(*update.Status)
	}
	m.UpdatedAt = time.Now()
	m.refreshSearchTerms()
	return m.Validate()
}

func (m *Memory) MarkArchived() error {
	now := time.Now()
	m.Status = StatusArchived
	m.Visibility = VisibilityArchived
	m.UpdatedAt = now
	return m.Validate()
}

func (m *Memory) MarkDeleted() error {
	now := time.Now()
	m.Status = StatusDeleted
	m.UpdatedAt = now
	return m.Validate()
}

func (m *Memory) TouchAccessed() {
	now := time.Now()
	m.LastAccessedAt = &now
	m.UpdatedAt = now
}

func (m *Memory) Validate() error {
	if !IsValidType(m.Type) {
		return ErrInvalidType
	}
	if strings.TrimSpace(m.Title) == "" {
		return ErrInvalidTitle
	}
	if strings.TrimSpace(m.Content) == "" {
		return ErrInvalidContent
	}
	if !IsValidSource(m.Source) {
		return ErrInvalidSource
	}
	if !IsValidVisibility(m.Visibility) {
		return ErrInvalidVisibility
	}
	if !IsValidStatus(m.Status) {
		return ErrInvalidStatus
	}
	if !IsValidConfidence(m.Confidence) {
		return ErrInvalidConfidence
	}
	if !IsValidImportance(m.Importance) {
		return ErrInvalidImportance
	}
	return nil
}

func (m *Memory) refreshSearchTerms() {
	m.SearchTerms = ExtractTerms(strings.Join([]string{m.Title, m.Summary, m.Content, strings.Join(m.Tags, " ")}, " "), 40)
}

func ValidateListFilter(filter ListFilter) error {
	if filter.Type != "" && !IsValidType(filter.Type) {
		return ErrInvalidType
	}
	if filter.Importance != "" && !IsValidImportance(filter.Importance) {
		return ErrInvalidImportance
	}
	if filter.Status != "" && !IsValidStatus(filter.Status) {
		return ErrInvalidStatus
	}
	if filter.Limit < 0 {
		return ErrInvalidLimit
	}
	return nil
}

func ValidateSearchFilter(filter SearchFilter) error {
	if filter.Type != "" && !IsValidType(filter.Type) {
		return ErrInvalidType
	}
	if filter.Importance != "" && !IsValidImportance(filter.Importance) {
		return ErrInvalidImportance
	}
	if filter.Limit < 0 {
		return ErrInvalidLimit
	}
	if strings.TrimSpace(filter.Query) == "" && filter.Type == "" && filter.Tag == "" && filter.Importance == "" {
		return ErrInvalidSearchQuery
	}
	return nil
}

func IsValidType(value string) bool {
	switch value {
	case TypePreference, TypeFact, TypeEpisodic, TypeSemantic, TypeInstruction, TypeRelationship, TypeProject, TypeSystemNote:
		return true
	default:
		return false
	}
}

func IsValidSource(value string) bool {
	switch value {
	case SourceUserManual, SourceConversation, SourceActivity, SourceReflection, SourceImported, SourceAIInferred:
		return true
	default:
		return false
	}
}

func IsValidVisibility(value string) bool {
	switch value {
	case VisibilityPrivate, VisibilityAssistantContext, VisibilityArchived:
		return true
	default:
		return false
	}
}

func IsValidStatus(value string) bool {
	switch value {
	case StatusActive, StatusArchived, StatusDeleted:
		return true
	default:
		return false
	}
}

func IsValidConfidence(value string) bool {
	switch value {
	case ConfidenceLow, ConfidenceMedium, ConfidenceHigh:
		return true
	default:
		return false
	}
}

func IsValidImportance(value string) bool {
	switch value {
	case ImportanceLow, ImportanceMedium, ImportanceHigh, ImportanceCritical:
		return true
	default:
		return false
	}
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
