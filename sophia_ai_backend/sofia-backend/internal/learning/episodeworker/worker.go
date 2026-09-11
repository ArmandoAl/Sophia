// Package episodeworker is the only learning worker allowed to read raw conversations.
// It creates episodes only; belief synthesis remains isolated from conversation text.
package episodeworker

import (
	"context"
	"time"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	conversationsdomain "github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
)

type UserIDLister interface{ ListIDs() ([]string, error) }

type Worker struct {
	users         UserIDLister
	conversations conversationsdomain.ConversationRepository
	messages      conversationsdomain.ConversationMessageRepository
	learning      *learningapp.Service
	now           func() time.Time
}

func New(users UserIDLister, conversations conversationsdomain.ConversationRepository, messages conversationsdomain.ConversationMessageRepository, learning *learningapp.Service) *Worker {
	return &Worker{users: users, conversations: conversations, messages: messages, learning: learning, now: time.Now}
}

func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	ids, err := w.users.ListIDs()
	if err != nil {
		return 0, err
	}
	now := w.now().UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	processed := 0
	for _, userID := range ids {
		count, err := w.RunDay(ctx, userID, day)
		if err != nil {
			return processed, err
		}
		processed += count
	}
	return processed, nil
}

func (w *Worker) RunDay(ctx context.Context, userID string, day time.Time) (int, error) {
	// ponytail: 100 daily conversations is the existing repository ceiling; paginate only if real usage reaches it.
	conversations, err := w.conversations.List(ctx, conversationsdomain.ListFilter{UserID: userID, Limit: conversationsdomain.MaxListLimit})
	if err != nil {
		return 0, err
	}
	end, processed := day.AddDate(0, 0, 1), 0
	for _, conversation := range conversations {
		if conversation.LastMessageAt == nil || conversation.LastMessageAt.Before(day) || !conversation.LastMessageAt.Before(end) {
			continue
		}
		messages, err := w.messages.List(ctx, userID, conversation.ID, 0, "")
		if err != nil {
			return processed, err
		}
		turns, entityIDs, seen := make([]runtimedomain.Turn, 0, len(messages)), make([]string, 0), make(map[string]bool)
		for _, message := range messages {
			turns = append(turns, runtimedomain.Turn{Role: message.Role, Content: message.Content})
			matches, err := w.learning.ResolveKnownEntities(ctx, userID, message.Content)
			if err != nil {
				return processed, err
			}
			for _, match := range matches {
				if !match.Ambiguous && match.Entity != nil && !seen[match.Entity.ID] {
					seen[match.Entity.ID] = true
					entityIDs = append(entityIDs, match.Entity.ID)
				}
			}
		}
		if len(entityIDs) == 0 {
			continue
		}
		if _, err := w.learning.ExtractEpisodes(ctx, userID, conversation.ID, learningdomain.EpisodeSourceConversation, *conversation.LastMessageAt, turns, entityIDs); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}
