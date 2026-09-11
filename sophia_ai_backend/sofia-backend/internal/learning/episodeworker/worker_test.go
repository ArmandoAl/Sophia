package episodeworker

import (
	"context"
	"fmt"
	"testing"
	"time"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	conversationsdomain "github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
	conversationsinfra "github.com/armandoalvarado/sofia-backend/internal/conversations/infrastructure"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	learninginfra "github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

type model struct {
	body  string
	calls int
}

func (m *model) Generate(context.Context, runtimedomain.ModelRequest) (runtimedomain.ModelResponse, error) {
	m.calls++
	return runtimedomain.ModelResponse{AssistantMessage: m.body}, nil
}

func TestRunDayCallsExtractorOnlyForConversationsWithEntities(t *testing.T) {
	ctx, userID := context.Background(), "user"
	conversations, messages := conversationsinfra.NewInMemoryConversationRepository(), conversationsinfra.NewInMemoryConversationMessageRepository()
	learning := learningapp.NewService(learninginfra.NewInMemoryBeliefRepository(), learninginfra.NewInMemoryPromptVersionRepository(), learninginfra.NewInMemoryDailySummaryRepository())
	learning.SetContextRepository(learninginfra.NewInMemoryUserContextRepository())
	episodes, extractor := learninginfra.NewInMemoryEpisodeRepository(), &model{}
	learning.SetEpisodeStore(episodes, extractor, learningdomain.EpisodeMinSalience, 5)
	runner := New(nil, conversations, messages, learning)
	now := time.Now().UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	conversation, _ := conversationsdomain.NewConversation("plain", userID, conversationsdomain.CreateConversationInput{})
	message, _ := conversationsdomain.NewMessage("m1", conversation.ID, userID, conversationsdomain.RoleUser, "Hoy cociné pasta", "")
	conversation.Touch(now)
	_ = conversations.Create(ctx, conversation)
	_ = messages.Create(ctx, message)
	processed, err := runner.RunDay(ctx, userID, day)
	if err != nil || processed != 0 || extractor.calls != 0 {
		t.Fatalf("plain conversation processed=%d calls=%d err=%v", processed, extractor.calls, err)
	}

	entity, _ := learning.CreateUserContext(ctx, userID, learningdomain.UserContextCreate{Kind: learningdomain.ScopePerson, Slug: "diana", Label: "Diana"})
	extractor.body = fmt.Sprintf(`{"episodes":[{"summary":"Diana aceptó el nuevo trabajo.","entity_ids":[%q],"topics":["trabajo"],"salience":0.8}]}`, entity.ID)
	conversation, _ = conversationsdomain.NewConversation("entity", userID, conversationsdomain.CreateConversationInput{})
	message, _ = conversationsdomain.NewMessage("m2", conversation.ID, userID, conversationsdomain.RoleUser, "Diana aceptó el trabajo", "")
	conversation.Touch(now)
	_ = conversations.Create(ctx, conversation)
	_ = messages.Create(ctx, message)
	processed, err = runner.RunDay(ctx, userID, day)
	if err != nil || processed != 1 || extractor.calls != 1 {
		t.Fatalf("entity conversation processed=%d calls=%d err=%v", processed, extractor.calls, err)
	}
}
