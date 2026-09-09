package application

import (
	"context"
	"testing"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	"github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
	"github.com/armandoalvarado/sofia-backend/internal/conversations/infrastructure"
)

type stubRuntime struct{}

func (stubRuntime) HandleMessage(_ context.Context, request runtimedomain.RuntimeRequest) (*runtimedomain.RuntimeResponse, error) {
	proposal := runtimedomain.ActionProposalOutput{
		ToolName: "create_reminder",
		Status:   "proposed",
	}
	if !request.DryRun {
		proposal.ID = "proposal-1"
	}
	return &runtimedomain.RuntimeResponse{
		AssistantMessage: "Listo, te propongo un recordatorio.",
		ProposedActions:  []runtimedomain.ActionProposalOutput{proposal},
		RequestID:        "runtime-1",
	}, nil
}

func TestSendMessageReturnsProposalIDs(t *testing.T) {
	ctx := context.Background()
	service := NewService(
		infrastructure.NewInMemoryConversationRepository(),
		infrastructure.NewInMemoryConversationMessageRepository(),
		stubRuntime{},
	)

	conversation, err := service.CreateConversation(ctx, "user-1", domain.CreateConversationInput{Title: "Chat"})
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.SendMessage(ctx, "user-1", conversation.ID, "Recuérdame estudiar mañana")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ProposedActions) == 0 {
		t.Fatal("expected proposed actions")
	}
	for _, proposal := range result.ProposedActions {
		if proposal.ID == "" {
			t.Fatalf("expected proposal ID, got %+v", proposal)
		}
	}
}

func TestSendMessagePassesPriorTurnsAsHistory(t *testing.T) {
	ctx := context.Background()
	runtime := &recordingRuntime{}
	service := NewService(
		infrastructure.NewInMemoryConversationRepository(),
		infrastructure.NewInMemoryConversationMessageRepository(),
		runtime,
	)

	conversation, err := service.CreateConversation(ctx, "user-1", domain.CreateConversationInput{Title: "Chat"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendMessage(ctx, "user-1", conversation.ID, "Recuérdame estudiar mañana"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendMessage(ctx, "user-1", conversation.ID, "mejor a las 4", "person:maria"); err != nil {
		t.Fatal(err)
	}
	if len(runtime.requests) != 2 {
		t.Fatalf("expected 2 runtime requests, got %d", len(runtime.requests))
	}
	second := runtime.requests[1]
	if second.ActiveContext != "person:maria" {
		t.Fatalf("active context = %q, want person:maria", second.ActiveContext)
	}
	if len(second.History) < 2 {
		t.Fatalf("expected prior turns in history, got %+v", second.History)
	}
	if second.History[0].Role != domain.RoleUser || second.History[0].Content != "Recuérdame estudiar mañana" {
		t.Fatalf("first history turn should be the previous user message, got %+v", second.History[0])
	}
	if second.History[1].Role != domain.RoleAssistant {
		t.Fatalf("second history turn should be the previous assistant message, got %+v", second.History[1])
	}
	for _, turn := range second.History {
		if turn.Content == "mejor a las 4" {
			t.Fatal("current message should not be duplicated in history")
		}
	}
}

type recordingRuntime struct {
	requests []runtimedomain.RuntimeRequest
}

func (r *recordingRuntime) HandleMessage(_ context.Context, request runtimedomain.RuntimeRequest) (*runtimedomain.RuntimeResponse, error) {
	copied := request
	if request.History != nil {
		copied.History = append([]runtimedomain.Turn(nil), request.History...)
	}
	r.requests = append(r.requests, copied)
	return &runtimedomain.RuntimeResponse{
		AssistantMessage: "Listo, te propongo un recordatorio.",
		RequestID:        "runtime-1",
	}, nil
}
