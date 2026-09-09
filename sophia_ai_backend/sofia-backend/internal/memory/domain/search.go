package domain

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	firestoreArrayContainsAnyLimit = 30
	defaultSearchLimit             = 10
)

var stopwords = map[string]struct{}{
	"para": {}, "pero": {}, "como": {}, "este": {}, "esta": {}, "esto": {},
	"estos": {}, "estas": {}, "cuando": {}, "cuándo": {}, "donde": {}, "dónde": {},
	"porque": {}, "tengo": {}, "tienes": {}, "tiene": {}, "tenemos": {}, "tienen": {},
	"quiero": {}, "quieres": {}, "quiere": {}, "quieren": {}, "puedes": {},
	"puede": {}, "pueden": {}, "podemos": {}, "sobre": {}, "desde": {},
	"hasta": {}, "entre": {}, "tambien": {}, "también": {}, "despues": {},
	"después": {}, "antes": {}, "ahora": {}, "quién": {}, "quien": {},
	"cuales": {}, "cuáles": {}, "mucho": {}, "muchos": {}, "muchas": {},
	"todos": {}, "todas": {}, "otro": {}, "otra": {}, "otros": {}, "otras": {},
	"menos": {}, "siempre": {}, "nunca": {}, "hacer": {}, "hacia": {},
	"según": {}, "segun": {}, "cada": {}, "algo": {}, "nada": {}, "mismo": {},
	"misma": {}, "aquí": {}, "aqui": {}, "allí": {}, "alli": {},
	"this": {}, "that": {}, "with": {}, "from": {}, "have": {},
	"been": {}, "were": {}, "will": {}, "would": {}, "could": {},
	"should": {}, "about": {}, "when": {}, "where": {}, "what": {},
	"which": {}, "your": {}, "their": {}, "there": {}, "then": {},
	"than": {}, "some": {}, "more": {}, "just": {}, "like": {},
	"only": {}, "over": {}, "such": {}, "most": {}, "also": {},
	"into": {}, "very": {}, "want": {}, "need": {}, "make": {},
	"made": {}, "being": {}, "them": {}, "they": {}, "does": {},
	"please": {},
}

func ExtractTerms(text string, max int) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	seen := make(map[string]struct{}, len(fields))
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if utf8.RuneCountInString(field) < 4 {
			continue
		}
		if _, skip := stopwords[field]; skip {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		result = append(result, field)
		if max > 0 && len(result) >= max {
			break
		}
	}
	return result
}

func CapSearchTerms(terms []string) []string {
	if len(terms) <= firestoreArrayContainsAnyLimit {
		return terms
	}
	return append([]string(nil), terms[:firestoreArrayContainsAnyLimit]...)
}

func CountTermMatches(memory *Memory, terms []string) int {
	if memory == nil || len(terms) == 0 || len(memory.SearchTerms) == 0 {
		return 0
	}
	owned := make(map[string]struct{}, len(memory.SearchTerms))
	for _, term := range memory.SearchTerms {
		owned[term] = struct{}{}
	}
	score := 0
	for _, term := range terms {
		if _, ok := owned[term]; ok {
			score++
		}
	}
	return score
}

func RankByTermMatches(memories []*Memory, terms []string, limit int) []*Memory {
	type scored struct {
		memory *Memory
		score  int
	}
	items := make([]scored, 0, len(memories))
	for _, memory := range memories {
		score := CountTermMatches(memory, terms)
		if score == 0 {
			continue
		}
		items = append(items, scored{memory: memory, score: score})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score == items[j].score {
			return items[i].memory.UpdatedAt.After(items[j].memory.UpdatedAt)
		}
		return items[i].score > items[j].score
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	result := make([]*Memory, len(items))
	for i, item := range items {
		result[i] = item.memory
	}
	return result
}

func DefaultSearchLimit() int {
	return defaultSearchLimit
}
