package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	EpisodeSourceConversation = "conversation"
	EpisodeSourceIngestion    = "ingestion"
	EpisodeStatusActive       = "active"
	EpisodeStatusArchived     = "archived"
	EpisodeMinSalience        = 0.4
	EpisodeHighSalience       = 0.8
	EpisodeRetention          = 180 * 24 * time.Hour
	DefaultEpisodeDailyLimit  = 5
	DefaultEpisodeSearchLimit = 5
	EpisodeContextTokenBudget = 300
	OpenThreadTokenBudget     = 100
)

var ErrInvalidEpisode = errors.New("invalid episode")

type Episode struct {
	ID             string
	UserID         string
	OccurredAt     time.Time
	Summary        string
	EntityIDs      []string
	Topics         []string
	Salience       float64
	Source         string
	ConversationID string
	SearchTerms    []string
	Embedding      []float32
	FollowUpAt     *time.Time
	FollowedUpAt   *time.Time
	Status         string
}

type EpisodeRepository interface {
	Create(ctx context.Context, episode *Episode) error
	Update(ctx context.Context, episode *Episode) error
	ListByConversation(ctx context.Context, userID, conversationID string) ([]*Episode, error)
	CountBetween(ctx context.Context, userID string, from, to time.Time) (int, error)
	SearchByEntity(ctx context.Context, userID, entityID string, limit int) ([]*Episode, error)
	ListDueFollowUps(ctx context.Context, userID string, now time.Time, limit int) ([]*Episode, error)
	ArchiveLowSalienceBefore(ctx context.Context, before time.Time, highSalience float64) (int, error)
	ReassignEntity(ctx context.Context, userID, sourceEntityID, targetEntityID string) error
}

func (e *Episode) Validate() error {
	if strings.TrimSpace(e.ID) == "" || strings.TrimSpace(e.UserID) == "" || strings.TrimSpace(e.Summary) == "" || len(e.EntityIDs) == 0 || e.OccurredAt.IsZero() || e.Salience < 0 || e.Salience > 1 {
		return ErrInvalidEpisode
	}
	if e.Source != EpisodeSourceConversation && e.Source != EpisodeSourceIngestion {
		return ErrInvalidEpisode
	}
	if e.Status == "" {
		e.Status = EpisodeStatusActive
	}
	if e.Status != EpisodeStatusActive && e.Status != EpisodeStatusArchived {
		return ErrInvalidEpisode
	}
	return nil
}
