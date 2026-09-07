package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	conversationsdomain "github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
	insightsdomain "github.com/armandoalvarado/sofia-backend/internal/insights/domain"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	"github.com/armandoalvarado/sofia-backend/internal/privacy/domain"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	usersapp "github.com/armandoalvarado/sofia-backend/internal/users/application"
	"github.com/google/uuid"
)

const defaultExportLimit = 1000

type UsersReader interface {
	GetMe(userID string) (*usersapp.Me, error)
}

type ActivityLister interface {
	ListActivities(ctx context.Context, filter activitiesdomain.ListFilter) ([]*activitiesdomain.Activity, error)
}

type ReminderLister interface {
	ListReminders(ctx context.Context, filter remindersdomain.ListFilter) ([]*remindersdomain.Reminder, error)
}

type InsightsReader interface {
	ListMoodEntries(ctx context.Context, filter insightsdomain.ListFilter) ([]*insightsdomain.MoodEntry, error)
	ListActivityOutcomes(ctx context.Context, filter insightsdomain.ListFilter) ([]*insightsdomain.ActivityOutcome, error)
	ListReflections(ctx context.Context, filter insightsdomain.ListFilter) ([]*insightsdomain.Reflection, error)
}

type MemoryLister interface {
	ListMemories(ctx context.Context, filter memorydomain.ListFilter) ([]*memorydomain.Memory, error)
}

type ActionProposalLister interface {
	ListActionProposals(ctx context.Context, filter actionsdomain.ListFilter) ([]*actionsdomain.AIActionProposal, error)
}

type ConversationLister interface {
	List(ctx context.Context, filter conversationsdomain.ListFilter) ([]*conversationsdomain.Conversation, error)
}
type ConversationMessageLister interface {
	List(ctx context.Context, userID, conversationID string, limit int, cursor string) ([]*conversationsdomain.ConversationMessage, error)
}

type Service struct {
	users          UsersReader
	activities     ActivityLister
	reminders      ReminderLister
	insights       InsightsReader
	memories       MemoryLister
	actions        ActionProposalLister
	conversations  ConversationLister
	messages       ConversationMessageLister
	auditLogs      domain.AuditLogRepository
	deleteRequests domain.DeleteRequestRepository
}

type UserDataExport struct {
	GeneratedAt          time.Time                                  `json:"generated_at"`
	User                 any                                        `json:"user"`
	Profile              any                                        `json:"profile"`
	AISettings           any                                        `json:"ai_settings"`
	Activities           []*activitiesdomain.Activity               `json:"activities"`
	Reminders            []*remindersdomain.Reminder                `json:"reminders"`
	MoodEntries          []*insightsdomain.MoodEntry                `json:"mood_entries"`
	ActivityOutcomes     []*insightsdomain.ActivityOutcome          `json:"activity_outcomes"`
	Reflections          []*insightsdomain.Reflection               `json:"reflections"`
	Memories             []*memorydomain.Memory                     `json:"memories"`
	ActionProposals      []*actionsdomain.AIActionProposal          `json:"action_proposals"`
	Conversations        []*conversationsdomain.Conversation        `json:"conversations"`
	ConversationMessages []*conversationsdomain.ConversationMessage `json:"conversation_messages"`
}

func NewService(
	users UsersReader,
	activities ActivityLister,
	reminders ReminderLister,
	insights InsightsReader,
	memories MemoryLister,
	actions ActionProposalLister,
	auditLogs domain.AuditLogRepository,
	deleteRequests domain.DeleteRequestRepository,
	conversationReaders ...any,
) *Service {
	service := &Service{users: users, activities: activities, reminders: reminders, insights: insights, memories: memories, actions: actions, auditLogs: auditLogs, deleteRequests: deleteRequests}
	for _, reader := range conversationReaders {
		switch value := reader.(type) {
		case ConversationLister:
			service.conversations = value
		case ConversationMessageLister:
			service.messages = value
		}
	}
	return service
}

func (s *Service) ExportUserData(ctx context.Context, userID string) (*UserDataExport, error) {
	me, err := s.users.GetMe(userID)
	if err != nil {
		return nil, err
	}
	activities, err := s.activities.ListActivities(ctx, activitiesdomain.ListFilter{UserID: userID, Limit: defaultExportLimit})
	if err != nil {
		return nil, err
	}
	reminders, err := s.reminders.ListReminders(ctx, remindersdomain.ListFilter{UserID: userID, Limit: defaultExportLimit})
	if err != nil {
		return nil, err
	}
	moods, err := s.insights.ListMoodEntries(ctx, insightsdomain.ListFilter{UserID: userID, Limit: defaultExportLimit})
	if err != nil {
		return nil, err
	}
	outcomes, err := s.insights.ListActivityOutcomes(ctx, insightsdomain.ListFilter{UserID: userID, Limit: defaultExportLimit})
	if err != nil {
		return nil, err
	}
	reflections, err := s.insights.ListReflections(ctx, insightsdomain.ListFilter{UserID: userID, Limit: defaultExportLimit})
	if err != nil {
		return nil, err
	}
	memories, err := s.memories.ListMemories(ctx, memorydomain.ListFilter{UserID: userID, Status: memorydomain.StatusActive, Limit: defaultExportLimit})
	if err != nil {
		return nil, err
	}
	proposals, err := s.actions.ListActionProposals(ctx, actionsdomain.ListFilter{UserID: userID, Limit: defaultExportLimit})
	if err != nil {
		return nil, err
	}
	conversations := []*conversationsdomain.Conversation{}
	conversationMessages := []*conversationsdomain.ConversationMessage{}
	if s.conversations != nil && s.messages != nil {
		conversations, err = s.conversations.List(ctx, conversationsdomain.ListFilter{UserID: userID, Limit: defaultExportLimit})
		if err != nil {
			return nil, err
		}
		for _, conversation := range conversations {
			messages, listErr := s.messages.List(ctx, userID, conversation.ID, defaultExportLimit, "")
			if listErr != nil {
				return nil, listErr
			}
			conversationMessages = append(conversationMessages, messages...)
		}
	}

	export := &UserDataExport{
		GeneratedAt:          time.Now(),
		User:                 me.User,
		Profile:              me.Profile,
		AISettings:           me.AISettings,
		Activities:           activities,
		Reminders:            reminders,
		MoodEntries:          moods,
		ActivityOutcomes:     outcomes,
		Reflections:          reflections,
		Memories:             memories,
		ActionProposals:      proposals,
		Conversations:        conversations,
		ConversationMessages: conversationMessages,
	}

	_ = s.RecordAuditLog(ctx, userID, domain.AuditActionExportRequested, "user", userID, map[string]any{
		"activity_count":     len(activities),
		"memory_count":       len(memories),
		"conversation_count": len(conversations),
	})
	return export, nil
}

func (s *Service) CreateDeleteRequest(ctx context.Context, userID, reason string) (*domain.DeleteRequest, error) {
	now := time.Now()
	request := &domain.DeleteRequest{
		ID:        uuid.New().String(),
		UserID:    userID,
		Status:    domain.DeleteRequestStatusPending,
		Reason:    strings.TrimSpace(reason),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.deleteRequests.Create(ctx, request); err != nil {
		return nil, err
	}
	if err := s.RecordAuditLog(ctx, userID, domain.AuditActionDeleteRequested, "delete_request", request.ID, map[string]any{"reason": request.Reason}); err != nil {
		return nil, err
	}
	return request, nil
}

func (s *Service) RecordAuditLog(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) error {
	if s.auditLogs == nil {
		return nil
	}
	return s.auditLogs.Create(ctx, &domain.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedAt:    time.Now(),
		Metadata:     redactMetadata(metadata),
	})
}

func redactMetadata(metadata map[string]any) map[string]any {
	redacted := map[string]any{}
	for key, value := range metadata {
		normalized := strings.ToLower(key)
		if strings.Contains(normalized, "token") || strings.Contains(normalized, "secret") || strings.Contains(normalized, "password") || strings.Contains(normalized, "email") {
			redacted[key] = "[REDACTED]"
			continue
		}
		if strings.Contains(normalized, "reason") || strings.Contains(normalized, "content") || strings.Contains(normalized, "note") {
			redacted[key] = truncate(value)
			continue
		}
		redacted[key] = value
	}
	return redacted
}

func truncate(value any) any {
	text := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprint(value), "\n", " "), "\t", " "))
	if len(text) <= 80 {
		return text
	}
	return text[:80] + "...[TRUNCATED]"
}
