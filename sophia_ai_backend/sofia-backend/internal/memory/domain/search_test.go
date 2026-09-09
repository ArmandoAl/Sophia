package domain

import (
	"strings"
	"testing"
	"time"
)

func TestExtractTermsLowercasesAndSplitsOnNonAlphanumeric(t *testing.T) {
	got := ExtractTerms("Café-planning: 2024!!", 10)
	want := []string{"café", "planning", "2024"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ExtractTerms = %v, want %v", got, want)
	}
}

func TestExtractTermsDropsShortTokensAndStopwords(t *testing.T) {
	got := ExtractTerms("para pero como este esta esto cuando donde porque tengo quiero puedes sobre café", 20)
	if len(got) != 1 || got[0] != "café" {
		t.Fatalf("expected only café, got %v", got)
	}
}

func TestExtractTermsDedupesPreservingOrderAndRespectsMax(t *testing.T) {
	got := ExtractTerms("coffee planning coffee morning planning", 2)
	if strings.Join(got, ",") != "coffee,planning" {
		t.Fatalf("ExtractTerms = %v, want coffee,planning", got)
	}
}

func TestExtractTermsReturnsEmptyForStopwordOnlyText(t *testing.T) {
	got := ExtractTerms("para pero como this that with from", 12)
	if len(got) != 0 {
		t.Fatalf("expected no terms, got %v", got)
	}
}

func TestRankByTermMatchesPrefersMoreOverlap(t *testing.T) {
	now := time.Now()
	low := &Memory{ID: "low", SearchTerms: []string{"coffee"}, UpdatedAt: now}
	high := &Memory{ID: "high", SearchTerms: []string{"coffee", "planning"}, UpdatedAt: now.Add(-time.Hour)}
	ranked := RankByTermMatches([]*Memory{low, high}, []string{"coffee", "planning"}, 10)
	if len(ranked) != 2 || ranked[0].ID != "high" {
		t.Fatalf("expected high score first, got %+v", ranked)
	}
}
