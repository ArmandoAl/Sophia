package app

import (
	"context"
	"fmt"
	"log"

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	activitiesinfra "github.com/armandoalvarado/sofia-backend/internal/activities/infrastructure"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	actionsinfra "github.com/armandoalvarado/sofia-backend/internal/ai/actions/infrastructure"
	authdomain "github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	authinfra "github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	conversationsdomain "github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
	conversationsinfra "github.com/armandoalvarado/sofia-backend/internal/conversations/infrastructure"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	ingestiondomain "github.com/armandoalvarado/sofia-backend/internal/ingestion/domain"
	ingestioninfra "github.com/armandoalvarado/sofia-backend/internal/ingestion/infrastructure"
	insightsdomain "github.com/armandoalvarado/sofia-backend/internal/insights/domain"
	insightsinfra "github.com/armandoalvarado/sofia-backend/internal/insights/infrastructure"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	learninginfra "github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	memoryinfra "github.com/armandoalvarado/sofia-backend/internal/memory/infrastructure"
	notificationsdomain "github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	notificationsinfra "github.com/armandoalvarado/sofia-backend/internal/notifications/infrastructure"
	privacydomain "github.com/armandoalvarado/sofia-backend/internal/privacy/domain"
	privacyinfra "github.com/armandoalvarado/sofia-backend/internal/privacy/infrastructure"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	remindersinfra "github.com/armandoalvarado/sofia-backend/internal/reminders/infrastructure"
	toolsdomain "github.com/armandoalvarado/sofia-backend/internal/tools/domain"
	toolsinfra "github.com/armandoalvarado/sofia-backend/internal/tools/infrastructure"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
	usersinfra "github.com/armandoalvarado/sofia-backend/internal/users/infrastructure"
)

type Repositories struct {
	Users            authdomain.UserRepository
	Profiles         usersdomain.UserProfileRepository
	AISettings       usersdomain.AISettingsRepository
	Activities       activitiesdomain.ActivityRepository
	Reminders        remindersdomain.ReminderRepository
	Moods            insightsdomain.MoodEntryRepository
	Outcomes         insightsdomain.ActivityOutcomeRepository
	Reflections      insightsdomain.ReflectionRepository
	Memories         memorydomain.MemoryRepository
	Beliefs          learningdomain.BeliefRepository
	PromptVersions   learningdomain.PromptVersionRepository
	DailySummaries   learningdomain.DailySummaryRepository
	UserContexts     learningdomain.UserContextRepository
	EntityCandidates learningdomain.EntityCandidateRepository
	Episodes         learningdomain.EpisodeRepository
	IngestionBatches ingestiondomain.BatchRepository
	DeviceTokens     notificationsdomain.DeviceTokenRepository
	Tools            toolsdomain.ToolDefinitionRepository
	ActionProposals  actionsdomain.AIActionProposalRepository
	Conversations    conversationsdomain.ConversationRepository
	Messages         conversationsdomain.ConversationMessageRepository
	AuditLogs        privacydomain.AuditLogRepository
	DeleteRequests   privacydomain.DeleteRequestRepository
	FirestoreStatus  string
	Close            func()
}

func BuildRepositories(cfg config.Config) (*Repositories, error) {
	if cfg.PersistenceDriver != "firestore" {
		return &Repositories{
			Users:            authinfra.NewInMemoryUserRepository(),
			Profiles:         usersinfra.NewInMemoryUserProfileRepository(),
			AISettings:       usersinfra.NewInMemoryAISettingsRepository(),
			Activities:       activitiesinfra.NewInMemoryActivityRepository(),
			Reminders:        remindersinfra.NewInMemoryReminderRepository(),
			Moods:            insightsinfra.NewInMemoryMoodEntryRepository(),
			Outcomes:         insightsinfra.NewInMemoryActivityOutcomeRepository(),
			Reflections:      insightsinfra.NewInMemoryReflectionRepository(),
			Memories:         memoryinfra.NewInMemoryMemoryRepository(),
			Beliefs:          learninginfra.NewInMemoryBeliefRepository(),
			PromptVersions:   learninginfra.NewInMemoryPromptVersionRepository(),
			DailySummaries:   learninginfra.NewInMemoryDailySummaryRepository(),
			UserContexts:     learninginfra.NewInMemoryUserContextRepository(),
			EntityCandidates: learninginfra.NewInMemoryEntityCandidateRepository(),
			Episodes:         learninginfra.NewInMemoryEpisodeRepository(),
			IngestionBatches: ingestioninfra.NewInMemoryBatchRepository(),
			DeviceTokens:     notificationsinfra.NewInMemoryDeviceTokenRepository(),
			Tools:            toolsinfra.NewInMemoryToolDefinitionRepository(),
			ActionProposals:  actionsinfra.NewInMemoryAIActionProposalRepository(),
			Conversations:    conversationsinfra.NewInMemoryConversationRepository(),
			Messages:         conversationsinfra.NewInMemoryConversationMessageRepository(),
			AuditLogs:        privacyinfra.NewInMemoryAuditLogRepository(),
			DeleteRequests:   privacyinfra.NewInMemoryDeleteRequestRepository(),
			FirestoreStatus:  "disabled",
			Close:            func() {},
		}, nil
	}

	store, err := database.NewFirestore(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("firestore bootstrap failed: %w", err)
	}
	log.Printf("firestore bootstrap ok project_id=%s emulator_enabled=%t", cfg.FirestoreProjectID, cfg.FirestoreEmulatorHost != "")

	return &Repositories{
		Users:            authinfra.NewFirestoreUserRepository(store.Client),
		Profiles:         usersinfra.NewFirestoreUserProfileRepository(store.Client),
		AISettings:       usersinfra.NewFirestoreAISettingsRepository(store.Client),
		Activities:       activitiesinfra.NewFirestoreActivityRepository(store.Client),
		Reminders:        remindersinfra.NewFirestoreReminderRepository(store.Client),
		Moods:            insightsinfra.NewFirestoreMoodEntryRepository(store.Client),
		Outcomes:         insightsinfra.NewFirestoreActivityOutcomeRepository(store.Client),
		Reflections:      insightsinfra.NewFirestoreReflectionRepository(store.Client),
		Memories:         memoryinfra.NewFirestoreMemoryRepository(store.Client),
		Beliefs:          learninginfra.NewFirestoreBeliefRepository(store.Client),
		PromptVersions:   learninginfra.NewFirestorePromptVersionRepository(store.Client),
		DailySummaries:   learninginfra.NewFirestoreDailySummaryRepository(store.Client),
		UserContexts:     learninginfra.NewFirestoreUserContextRepository(store.Client),
		EntityCandidates: learninginfra.NewFirestoreEntityCandidateRepository(store.Client),
		Episodes:         learninginfra.NewFirestoreEpisodeRepository(store.Client),
		IngestionBatches: ingestioninfra.NewFirestoreBatchRepository(store.Client),
		DeviceTokens:     notificationsinfra.NewFirestoreDeviceTokenRepository(store.Client),
		Tools:            toolsinfra.NewFirestoreToolDefinitionRepository(store.Client),
		ActionProposals:  actionsinfra.NewFirestoreAIActionProposalRepository(store.Client),
		Conversations:    conversationsinfra.NewFirestoreConversationRepository(store.Client),
		Messages:         conversationsinfra.NewFirestoreConversationMessageRepository(store.Client),
		AuditLogs:        privacyinfra.NewFirestoreAuditLogRepository(store.Client),
		DeleteRequests:   privacyinfra.NewFirestoreDeleteRequestRepository(store.Client),
		FirestoreStatus:  "ok",
		Close: func() {
			if err := store.Close(); err != nil {
				log.Printf("firestore close failed: %v", err)
			}
		},
	}, nil
}
