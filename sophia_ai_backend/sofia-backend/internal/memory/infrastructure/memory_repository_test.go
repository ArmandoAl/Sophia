package infrastructure

import (
	"context"
	"strings"
	"testing"

	"github.com/armandoalvarado/sofia-backend/internal/memory/domain"
)

func TestSearchByTermsRecoversMemoryFromNaturalLanguagePhrase(t *testing.T) {
	repo := NewInMemoryMemoryRepository()
	ctx := context.Background()
	memory, err := domain.NewMemory("mem-coffee", "user-1", domain.MemoryCreate{
		Title:   "Coffee preference",
		Content: "Likes coffee before planning the day",
		Tags:    []string{"planning"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, memory); err != nil {
		t.Fatal(err)
	}

	phrase := "Hey can you please remind me what I like about coffee when I am planning my morning?"
	foundBasic, err := repo.SearchBasic(ctx, domain.SearchFilter{UserID: "user-1", Query: phrase, Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(foundBasic) != 0 {
		t.Fatalf("phrase contains-search should miss, got %+v", foundBasic)
	}

	terms := domain.ExtractTerms(phrase, 12)
	if len(terms) == 0 {
		t.Fatal("expected searchable terms from the natural language phrase")
	}
	found, err := repo.SearchByTerms(ctx, "user-1", terms, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].ID != memory.ID {
		t.Fatalf("expected coffee memory, got %+v (terms=%v search_terms=%v)", found, terms, memory.SearchTerms)
	}
}

func TestSearchByTermsReturnsEmptyWithoutScanningWhenTermsEmpty(t *testing.T) {
	repo := NewInMemoryMemoryRepository()
	ctx := context.Background()
	memory, err := domain.NewMemory("mem-coffee", "user-1", domain.MemoryCreate{
		Title:   "Coffee preference",
		Content: "Likes coffee before planning the day",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, memory); err != nil {
		t.Fatal(err)
	}

	found, err := repo.SearchByTerms(ctx, "user-1", nil, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 0 {
		t.Fatalf("expected no results for empty terms, got %+v", found)
	}
}

func TestNewMemoryIndexesSearchTermsOnWrite(t *testing.T) {
	memory, err := domain.NewMemory("mem-coffee", "user-1", domain.MemoryCreate{
		Title:   "Coffee preference",
		Content: "Likes coffee before planning",
		Tags:    []string{"planning"},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(memory.SearchTerms, ",")
	for _, term := range []string{"coffee", "preference", "planning"} {
		if !containsTerm(memory.SearchTerms, term) {
			t.Fatalf("SearchTerms %q missing %s", joined, term)
		}
	}
}

func containsTerm(terms []string, want string) bool {
	for _, term := range terms {
		if term == want {
			return true
		}
	}
	return false
}
