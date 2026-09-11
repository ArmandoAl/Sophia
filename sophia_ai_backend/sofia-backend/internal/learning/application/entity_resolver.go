package application

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"golang.org/x/text/unicode/norm"
)

const (
	defaultEntityCatalogTTL = time.Minute
	entitySampleLimit       = 120
)

type EntityResolver struct {
	contexts   domain.UserContextRepository
	candidates domain.EntityCandidateRepository
	ttl        time.Duration
	mu         sync.Mutex
	cache      map[string]cachedEntityCatalog
}

type cachedEntityCatalog struct {
	entities  []*domain.UserContext
	expiresAt time.Time
}

func NewEntityResolver(contexts domain.UserContextRepository, candidates domain.EntityCandidateRepository, ttl time.Duration) *EntityResolver {
	if ttl <= 0 {
		ttl = defaultEntityCatalogTTL
	}
	return &EntityResolver{contexts: contexts, candidates: candidates, ttl: ttl, cache: make(map[string]cachedEntityCatalog)}
}

func (r *EntityResolver) Resolve(ctx context.Context, userID, message string) ([]domain.EntityMatch, error) {
	entities, err := r.catalog(ctx, userID)
	if err != nil {
		return nil, err
	}
	matches := resolveKnownEntities(message, entities)
	if r.candidates != nil {
		if err := r.captureCandidates(ctx, userID, message, entities); err != nil {
			return nil, err
		}
	}
	return matches, nil
}

func (r *EntityResolver) ResolveKnown(ctx context.Context, userID, message string) ([]domain.EntityMatch, error) {
	entities, err := r.catalog(ctx, userID)
	if err != nil {
		return nil, err
	}
	return resolveKnownEntities(message, entities), nil
}

func resolveKnownEntities(message string, entities []*domain.UserContext) []domain.EntityMatch {
	normalized, offsets := normalizeEntityText(message)
	matches := matchEntities(normalized, entities)
	for i := range matches {
		matches[i].Start, matches[i].End = offsets[matches[i].Start], offsets[matches[i].End]
		matches[i].Mention = message[matches[i].Start:matches[i].End]
	}
	return matches
}

func (r *EntityResolver) Invalidate(userID string) {
	r.mu.Lock()
	delete(r.cache, userID)
	r.mu.Unlock()
}

func (r *EntityResolver) catalog(ctx context.Context, userID string) ([]*domain.UserContext, error) {
	now := time.Now()
	// ponytail: one global lock is enough for tiny (<15 entity) catalogs; shard it only if contention is measured.
	r.mu.Lock()
	if cached, ok := r.cache[userID]; ok && now.Before(cached.expiresAt) {
		entities := cached.entities
		r.mu.Unlock()
		return entities, nil
	}
	r.mu.Unlock()
	values, err := r.contexts.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	entities := make([]*domain.UserContext, 0, len(values))
	for _, value := range values {
		if value != nil && value.CanLoadContext() {
			entities = append(entities, value)
		}
	}
	r.mu.Lock()
	r.cache[userID] = cachedEntityCatalog{entities: entities, expiresAt: now.Add(r.ttl)}
	r.mu.Unlock()
	return entities, nil
}

func matchEntities(message string, entities []*domain.UserContext) []domain.EntityMatch {
	result := make([]domain.EntityMatch, 0)
	type matchKey struct {
		entityID   string
		start, end int
	}
	seen := make(map[matchKey]struct{})
	for _, entity := range entities {
		aliases := append([]string{entity.Slug, entity.Label}, entity.Aliases...)
		for _, alias := range aliases {
			alias = normalizeEntityMention(alias)
			if alias == "" {
				continue
			}
			for offset := 0; offset < len(message); {
				index := strings.Index(message[offset:], alias)
				if index < 0 {
					break
				}
				start, end := offset+index, offset+index+len(alias)
				if wordBoundary(message, start, end) {
					key := matchKey{entityID: entity.ID, start: start, end: end}
					if _, ok := seen[key]; !ok {
						seen[key] = struct{}{}
						result = append(result, domain.EntityMatch{Entity: entity, Mention: message[start:end], Start: start, End: end})
					}
				}
				offset = start + 1
			}
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Start == result[j].Start {
			if result[i].End == result[j].End {
				return result[i].Entity.ID < result[j].Entity.ID
			}
			return result[i].End > result[j].End
		}
		return result[i].Start < result[j].Start
	})
	// ponytail: quadratic overlap removal is bounded by tiny catalogs; use an interval index only if catalogs grow materially.
	compacted := result[:0]
	for _, candidate := range result {
		overlaps := false
		for _, kept := range compacted {
			if candidate.Entity.ID == kept.Entity.ID && candidate.Start < kept.End && kept.Start < candidate.End {
				overlaps = true
				break
			}
		}
		if !overlaps {
			compacted = append(compacted, candidate)
		}
	}
	result = compacted
	for start := 0; start < len(result); {
		end := start + 1
		for end < len(result) && result[end].Start == result[start].Start && result[end].End == result[start].End {
			end++
		}
		if end-start > 1 {
			for index := start; index < end; index++ {
				result[index].Ambiguous = true
			}
		}
		start = end
	}
	return result
}

func wordBoundary(value string, start, end int) bool {
	if start > 0 {
		before, _ := utf8.DecodeLastRuneInString(value[:start])
		if unicode.IsLetter(before) || unicode.IsNumber(before) {
			return false
		}
	}
	if end < len(value) {
		after, _ := utf8.DecodeRuneInString(value[end:])
		if unicode.IsLetter(after) || unicode.IsNumber(after) {
			return false
		}
	}
	return true
}

func normalizeEntityMention(value string) string {
	normalized, _ := normalizeEntityText(value)
	return normalized
}

func normalizeEntityText(value string) (string, []int) {
	var normalized strings.Builder
	offsets := []int{0}
	needsSpace := false
	for originalStart, original := range value {
		originalEnd := originalStart + utf8.RuneLen(original)
		wrote := false
		for _, current := range norm.NFD.String(strings.ToLower(string(original))) {
			if unicode.Is(unicode.Mn, current) || (!unicode.IsLetter(current) && !unicode.IsNumber(current)) {
				continue
			}
			if needsSpace && normalized.Len() > 0 {
				normalized.WriteByte(' ')
				offsets = append(offsets, originalStart)
			}
			if normalized.Len() == 0 {
				offsets[0] = originalStart
			}
			encoded := []byte(string(current))
			for index, value := range encoded {
				normalized.WriteByte(value)
				if index == len(encoded)-1 {
					offsets = append(offsets, originalEnd)
				} else {
					offsets = append(offsets, originalStart)
				}
			}
			needsSpace, wrote = false, true
		}
		if !wrote {
			needsSpace = true
		}
	}
	return normalized.String(), offsets
}

func (r *EntityResolver) captureCandidates(ctx context.Context, userID, message string, entities []*domain.UserContext) error {
	known := make(map[string]struct{})
	for _, entity := range entities {
		for _, value := range append([]string{entity.Slug, entity.Label}, entity.Aliases...) {
			for _, word := range strings.Fields(normalizeEntityMention(value)) {
				known[word] = struct{}{}
			}
		}
	}
	now, sample := time.Now().UTC(), truncateRunes(strings.TrimSpace(message), entitySampleLimit)
	for _, token := range capitalizedWords(message) {
		normalized := normalizeEntityMention(token.value)
		if token.sentenceStart || normalized == "" || ignoredProperName(normalized) {
			continue
		}
		if _, ok := known[normalized]; ok {
			continue
		}
		if err := r.candidates.UpsertMention(ctx, userID, token.value, normalized, sample, now); err != nil {
			return err
		}
	}
	return nil
}

type wordToken struct {
	value         string
	sentenceStart bool
}

func capitalizedWords(value string) []wordToken {
	result := make([]wordToken, 0)
	sentenceStart := true
	for start := 0; start < len(value); {
		current, size := utf8.DecodeRuneInString(value[start:])
		if !unicode.IsLetter(current) {
			if current == '.' || current == '!' || current == '?' {
				sentenceStart = true
			}
			start += size
			continue
		}
		end := start + size
		hasLower := unicode.IsLower(current)
		for end < len(value) {
			next, nextSize := utf8.DecodeRuneInString(value[end:])
			if !unicode.IsLetter(next) && !unicode.Is(unicode.Mn, next) && next != '\'' && next != '’' && next != '-' {
				break
			}
			hasLower = hasLower || unicode.IsLower(next)
			end += nextSize
		}
		if unicode.IsUpper(current) && hasLower {
			result = append(result, wordToken{value: value[start:end], sentenceStart: sentenceStart})
		}
		sentenceStart = false
		start = end
	}
	return result
}

func ignoredProperName(value string) bool {
	_, ok := ignoredEntityCandidateWords[value]
	return ok
}

var ignoredEntityCandidateWords = func() map[string]struct{} {
	words := strings.Fields("a al algo ante aqui asi como con contra cual cuando de del desde donde el ella en entre era es esa ese esto esta fue hacia hasta hay hoy la las le lo los mas me mi mis muy no nos o para pero por porque que se si sin sobre su sus te tu tus un una uno y yo sofia lunes martes miercoles jueves viernes sabado domingo enero febrero marzo abril mayo junio julio agosto septiembre octubre noviembre diciembre")
	result := make(map[string]struct{}, len(words))
	for _, word := range words {
		result[word] = struct{}{}
	}
	return result
}()

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
