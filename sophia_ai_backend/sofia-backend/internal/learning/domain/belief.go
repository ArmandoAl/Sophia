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
	StatusArchived   = "archived"

	SubjectUser   = "user"
	SubjectEntity = "entity"

	FactKindTrait        = "trait"
	FactKindState        = "state"
	FactKindRelationship = "relationship"

	PromptSlotCore        = "core"
	PromptSlotSituational = "situational"

	ScopeGlobal = "global"
	ScopePerson = "person"
	ScopeMode   = "mode"

	TrustDecision = 1
	TrustStated   = 2
	TrustInferred = 3

	InitialConfidence               = 0.4
	ReinforceRate                   = 0.2
	ContradictionWeight             = 2.5
	HalfLifeDays                    = 90.0
	HighOverlapJaccard              = 0.5
	defaultSearchTermMax            = 40
	defaultSearchLimit              = 10
	defaultListLimit                = 50
	DefaultEntityContextTokenBudget = 400
	ContextFragmentTokenCeiling     = DefaultEntityContextTokenBudget
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
	ErrInvalidScope          = errors.New("invalid belief scope")
	ErrInvalidScopeKey       = errors.New("invalid belief scope key")
	ErrInvalidTrustTier      = errors.New("invalid belief trust tier")
	ErrInvalidSubjectType    = errors.New("invalid belief subject type")
	ErrInvalidSubjectID      = errors.New("invalid belief subject id")
	ErrInvalidFactKind       = errors.New("invalid belief fact kind")
	ErrStateValidUntil       = errors.New("state facts require valid_until")
	ErrPromptVersionNotFound = errors.New("prompt version not found")
	ErrInvalidPromptContent  = errors.New("prompt content is required")
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
	Embedding          []float32
	Scope              string
	ScopeKey           string
	TrustTier          int
	BatchID            string
	SubjectType        string
	SubjectID          string
	FactKind           string
	ValidUntil         *time.Time
	FollowUpAt         *time.Time
	FollowedUpAt       *time.Time
	Sensitive          bool
}

type BeliefCreate struct {
	Statement   string
	Category    string
	PromptSlot  string
	Scope       string
	ScopeKey    string
	TrustTier   int
	BatchID     string
	SubjectType string
	SubjectID   string
	FactKind    string
	ValidUntil  *time.Time
	FollowUpAt  *time.Time
	Sensitive   bool
}

type BeliefRepository interface {
	Create(ctx context.Context, belief *Belief) error
	Update(ctx context.Context, belief *Belief) error
	List(ctx context.Context, userID string, limit int) ([]*Belief, error)
	FindByID(ctx context.Context, userID, beliefID string) (*Belief, error)
	ListActive(ctx context.Context, userID string, limit int) ([]*Belief, error)
	ListActiveByScope(ctx context.Context, userID, scope, scopeKey string, limit int) ([]*Belief, error)
	SearchByTerms(ctx context.Context, userID string, terms []string, limit int) ([]*Belief, error)
	SetPromptSlot(ctx context.Context, userID, beliefID, slot string) (*Belief, error)
	RetireByBatchID(ctx context.Context, userID, batchID string) (int, error)
	ListActiveBySubject(ctx context.Context, userID, subjectType, subjectID string, limit int) ([]*Belief, error)
	ReassignEntityFacts(ctx context.Context, userID, sourceEntityID, targetEntityID string) error
	ArchiveExpiredStates(ctx context.Context, now time.Time) (int, error)
	ListDueFollowUps(ctx context.Context, userID string, now time.Time, limit int) ([]*Belief, error)
}

func NewBelief(id, userID string, input BeliefCreate) (*Belief, error) {
	now := time.Now().UTC()
	subjectType := strings.ToLower(strings.TrimSpace(input.SubjectType))
	if subjectType == "" {
		subjectType = SubjectUser
	}
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
		Scope:            effectiveScope(input.Scope),
		ScopeKey:         strings.ToLower(strings.TrimSpace(input.ScopeKey)),
		TrustTier:        input.TrustTier,
		BatchID:          strings.TrimSpace(input.BatchID),
		SubjectType:      subjectType,
		SubjectID:        strings.TrimSpace(input.SubjectID),
		FactKind:         strings.ToLower(strings.TrimSpace(input.FactKind)),
		ValidUntil:       cloneTimeValue(input.ValidUntil),
		FollowUpAt:       cloneTimeValue(input.FollowUpAt),
		Sensitive:        input.Sensitive || sensitiveByDefault(input.Statement),
	}
	if belief.FactKind == FactKindState && belief.ValidUntil != nil && belief.FollowUpAt == nil {
		midpoint := now.Add(belief.ValidUntil.Sub(now) / 2)
		belief.FollowUpAt = &midpoint
	}
	belief.refreshDerived()
	if err := belief.Validate(); err != nil {
		return nil, err
	}
	return belief, nil
}

func (b *Belief) Reinforce(incomingTier ...int) {
	tier := TrustDecision
	if len(incomingTier) > 0 {
		tier = incomingTier[0]
	}
	if tier == TrustDecision {
		b.TrustTier = TrustDecision
	}
	now := time.Now().UTC()
	b.EvidenceCount++
	ceiling := trustCeiling(b.EffectiveTrustTier())
	b.Confidence = math.Min(ceiling, clamp01(b.Confidence+(1-b.Confidence)*ReinforceRate))
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
	// Traits and relationships are stable. Only temporary states expire; otherwise Sofía could treat a sister as newly single a year later.
	if b.FactKind == FactKindTrait || b.FactKind == FactKindRelationship {
		return b.Confidence
	}
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
	if !IsValidScope(b.Scope) {
		return ErrInvalidScope
	}
	if !isValidScopeKey(b.EffectiveScope(), b.ScopeKey) {
		return ErrInvalidScopeKey
	}
	if !IsValidTrustTier(b.TrustTier) {
		return ErrInvalidTrustTier
	}
	if b.Confidence > trustCeiling(b.EffectiveTrustTier()) {
		return ErrInvalidConfidence
	}
	if !IsValidSubjectType(b.EffectiveSubjectType()) {
		return ErrInvalidSubjectType
	}
	if (b.EffectiveSubjectType() == SubjectEntity && strings.TrimSpace(b.SubjectID) == "") || (b.EffectiveSubjectType() == SubjectUser && strings.TrimSpace(b.SubjectID) != "") {
		return ErrInvalidSubjectID
	}
	if (b.FactKind != "" && !IsValidFactKind(b.FactKind)) || (b.EffectiveSubjectType() == SubjectEntity && b.FactKind == "") {
		return ErrInvalidFactKind
	}
	if b.FactKind == FactKindState && b.ValidUntil == nil {
		return ErrStateValidUntil
	}
	if b.FactKind != FactKindState && b.ValidUntil != nil {
		return ErrInvalidFactKind
	}
	return nil
}

func (b *Belief) EffectiveSubjectType() string {
	if value := strings.ToLower(strings.TrimSpace(b.SubjectType)); value != "" {
		return value
	}
	return SubjectUser
}

func (b *Belief) EffectiveScope() string {
	return effectiveScope(b.Scope)
}

func (b *Belief) EffectiveTrustTier() int {
	if b.TrustTier == 0 {
		return TrustDecision
	}
	return b.TrustTier
}

func (b *Belief) SetInitialConfidence(confidence float64) {
	b.Confidence = math.Min(clamp01(confidence), trustCeiling(b.EffectiveTrustTier()))
}

func (b *Belief) UpdateStatement(statement string) error {
	statement = strings.TrimSpace(statement)
	if statement == "" {
		return ErrInvalidStatement
	}
	b.Statement = statement
	b.Sensitive = b.Sensitive || sensitiveByDefault(statement)
	b.refreshDerived()
	return nil
}

func (b *Belief) refreshDerived() {
	b.SearchTerms = memorydomain.ExtractTerms(b.Statement, defaultSearchTermMax)
	b.TokenCost = ApproximateTokens(b.Statement)
}

func sensitiveByDefault(statement string) bool {
	statement = strings.ToLower(statement)
	for _, term := range []string{"salud", "enferm", "médic", "medic", "hospital", "dinero", "deuda", "finanz", "conflicto", "pelea", "discusión", "discusion"} {
		if strings.Contains(statement, term) {
			return true
		}
	}
	return false
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
	case StatusActive, StatusSuperseded, StatusRetired, StatusArchived:
		return true
	default:
		return false
	}
}

func IsValidSubjectType(value string) bool {
	return value == SubjectUser || value == SubjectEntity
}

func IsValidFactKind(value string) bool {
	switch value {
	case FactKindTrait, FactKindState, FactKindRelationship:
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

func IsValidScope(value string) bool {
	switch effectiveScope(value) {
	case ScopeGlobal, ScopePerson, ScopeMode:
		return true
	default:
		return false
	}
}

func IsValidTrustTier(value int) bool {
	return value == 0 || value == TrustDecision || value == TrustStated || value == TrustInferred
}

func trustCeiling(tier int) float64 {
	switch tier {
	case TrustStated:
		return 0.8
	case TrustInferred:
		return 0.5
	default:
		return 1
	}
}

func effectiveScope(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ScopeGlobal
	}
	return value
}

func isValidScopeKey(scope, key string) bool {
	key = strings.TrimSpace(key)
	if scope == ScopeGlobal {
		return key == ""
	}
	return strings.HasPrefix(key, scope+":") && strings.TrimPrefix(key, scope+":") != "" && !strings.Contains(strings.TrimPrefix(key, scope+":"), ":")
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

func cloneTimeValue(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
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
