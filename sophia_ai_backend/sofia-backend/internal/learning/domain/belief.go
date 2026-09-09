package domain

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
)

const (
	CategorySchedule      = "schedule"
	CategoryCommunication = "communication"
	CategoryPriorities    = "priorities"
	CategoryWorkStyle     = "work_style"
	CategoryPersonal      = "personal"
	CategoryConstraint    = "constraint"

	StatusActive     = "active"
	StatusSuperseded = "superseded"
	StatusRetired    = "retired"

	PromptSlotCore        = "core"
	PromptSlotSituational = "situational"

	InitialConfidence    = 0.4
	ReinforceRate        = 0.2
	ContradictionWeight  = 2.5
	HalfLifeDays         = 90.0
	HighOverlapJaccard   = 0.5
	defaultSearchTermMax = 40
	defaultSearchLimit   = 10
	defaultListLimit     = 50
)

var (
	ErrBeliefNotFound        = errors.New("belief not found")
	ErrInvalidStatement      = errors.New("statement is required")
	ErrInvalidCategory       = errors.New("invalid belief category")
	ErrInvalidStatus         = errors.New("invalid belief status")
	ErrInvalidPromptSlot     = errors.New("invalid prompt slot")
	ErrInvalidSupersede      = errors.New("supersede target is required")
	ErrBeliefNotActive       = errors.New("belief is not active")
	ErrInvalidConfidence     = errors.New("confidence must be between 0 and 1")
	ErrPromptVersionNotFound = errors.New("prompt version not found")
	ErrInvalidPromptContent  = errors.New("prompt content is required")
	ErrDailySummaryNotFound  = errors.New("daily summary not found")
	ErrDailySummaryExists    = errors.New("daily summary already exists for this date")
	ErrInvalidDate           = errors.New("date must be YYYY-MM-DD")
)

type Belief struct {
	ID                 string
	UserID             string
	Statement          string
	Category           string
	Confidence         float64
	EvidenceCount      int
	ContradictionCount int
	FirstObservedAt    time.Time
	LastReinforcedAt   time.Time
	LastContradictedAt *time.Time
	Status             string
	SupersededBy       string
	SupersedeReason    string
	PromptSlot         string
	TokenCost          int
	SearchTerms        []string
}

type BeliefCreate struct {
	Statement  string
	Category   string
	PromptSlot string
}

type BeliefRepository interface {
	Create(ctx context.Context, belief *Belief) error
	Update(ctx context.Context, belief *Belief) error
	FindByID(ctx context.Context, userID, beliefID string) (*Belief, error)
	ListActive(ctx context.Context, userID string, limit int) ([]*Belief, error)
	SearchByTerms(ctx context.Context, userID string, terms []string, limit int) ([]*Belief, error)
	SetPromptSlot(ctx context.Context, userID, beliefID, slot string) (*Belief, error)
}

func NewBelief(id, userID string, input BeliefCreate) (*Belief, error) {
	now := time.Now().UTC()
	belief := &Belief{
		ID:               strings.TrimSpace(id),
		UserID:           strings.TrimSpace(userID),
		Statement:        strings.TrimSpace(input.Statement),
		Category:         strings.TrimSpace(input.Category),
		Confidence:       InitialConfidence,
		EvidenceCount:    1,
		FirstObservedAt:  now,
		LastReinforcedAt: now,
		Status:           StatusActive,
		PromptSlot:       strings.TrimSpace(input.PromptSlot),
	}
	belief.refreshDerived()
	if err := belief.Validate(); err != nil {
		return nil, err
	}
	return belief, nil
}

func (b *Belief) Reinforce() {
	now := time.Now().UTC()
	b.EvidenceCount++
	b.Confidence = clamp01(b.Confidence + (1-b.Confidence)*ReinforceRate)
	b.LastReinforcedAt = now
}

func (b *Belief) Contradict() {
	now := time.Now().UTC()
	b.ContradictionCount++
	b.Confidence = clamp01(b.Confidence * (1 - ReinforceRate*ContradictionWeight))
	b.LastContradictedAt = &now
}

func (b *Belief) Supersede(newID, reason string) error {
	newID = strings.TrimSpace(newID)
	if newID == "" {
		return ErrInvalidSupersede
	}
	if b.Status != StatusActive {
		return ErrBeliefNotActive
	}
	b.Status = StatusSuperseded
	b.SupersededBy = newID
	b.SupersedeReason = strings.TrimSpace(reason)
	return nil
}

func (b *Belief) PromptValue(now time.Time) float64 {
	return b.DecayedConfidence(now) * math.Log1p(float64(b.EvidenceCount)) / (1 + float64(b.ContradictionCount))
}

func (b *Belief) DecayedConfidence(now time.Time) float64 {
	if b.LastReinforcedAt.IsZero() {
		return b.Confidence
	}
	days := now.Sub(b.LastReinforcedAt).Hours() / 24
	if days < 0 {
		days = 0
	}
	return clamp01(b.Confidence * math.Pow(0.5, days/HalfLifeDays))
}

func (b *Belief) Validate() error {
	if strings.TrimSpace(b.Statement) == "" {
		return ErrInvalidStatement
	}
	if !IsValidCategory(b.Category) {
		return ErrInvalidCategory
	}
	if !IsValidStatus(b.Status) {
		return ErrInvalidStatus
	}
	if b.PromptSlot != "" && !IsValidPromptSlot(b.PromptSlot) {
		return ErrInvalidPromptSlot
	}
	if b.Confidence < 0 || b.Confidence > 1 {
		return ErrInvalidConfidence
	}
	return nil
}

func (b *Belief) refreshDerived() {
	b.SearchTerms = memorydomain.ExtractTerms(b.Statement, defaultSearchTermMax)
	b.TokenCost = ApproximateTokens(b.Statement)
}

func ApproximateTokens(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	return utf8.RuneCountInString(text)/4 + 1
}

func IsValidCategory(value string) bool {
	switch value {
	case CategorySchedule, CategoryCommunication, CategoryPriorities, CategoryWorkStyle, CategoryPersonal, CategoryConstraint:
		return true
	default:
		return false
	}
}

func IsValidStatus(value string) bool {
	switch value {
	case StatusActive, StatusSuperseded, StatusRetired:
		return true
	default:
		return false
	}
}

func IsValidPromptSlot(value string) bool {
	switch value {
	case PromptSlotCore, PromptSlotSituational:
		return true
	default:
		return false
	}
}

func HasHighTermOverlap(a, b []string) bool {
	left := distinctiveTerms(a)
	right := distinctiveTerms(b)
	if len(left) == 0 || len(right) == 0 {
		left, right = a, b
	}
	if len(left) == 0 || len(right) == 0 {
		return false
	}
	overlap := countTermOverlap(left, right)
	if overlap == 0 {
		return false
	}
	union := len(left) + len(right) - overlap
	if union == 0 {
		return false
	}
	return float64(overlap)/float64(union) >= HighOverlapJaccard
}

func RankBeliefsByTermMatches(beliefs []*Belief, terms []string, limit int) []*Belief {
	type scored struct {
		belief *Belief
		score  int
	}
	items := make([]scored, 0, len(beliefs))
	for _, belief := range beliefs {
		score := countTermOverlap(belief.SearchTerms, terms)
		if score == 0 {
			continue
		}
		items = append(items, scored{belief: belief, score: score})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score == items[j].score {
			return items[i].belief.LastReinforcedAt.After(items[j].belief.LastReinforcedAt)
		}
		return items[i].score > items[j].score
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	result := make([]*Belief, len(items))
	for i, item := range items {
		result[i] = item.belief
	}
	return result
}

func DefaultSearchLimit() int {
	return defaultSearchLimit
}

func DefaultListLimit() int {
	return defaultListLimit
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func countTermOverlap(a, b []string) int {
	owned := make(map[string]struct{}, len(a))
	for _, term := range a {
		owned[term] = struct{}{}
	}
	score := 0
	for _, term := range b {
		if _, ok := owned[term]; ok {
			score++
		}
	}
	return score
}

var beliefBoilerplate = map[string]struct{}{
	"prefiere": {}, "prefer": {}, "prefers": {}, "gusta": {}, "suele": {},
	"tiende": {}, "usuario": {}, "siempre": {}, "nunca": {},
}

func distinctiveTerms(terms []string) []string {
	result := make([]string, 0, len(terms))
	for _, term := range terms {
		if _, skip := beliefBoilerplate[term]; skip {
			continue
		}
		result = append(result, term)
	}
	return result
}
