package app

import (
	"context"
	"errors"

	activitiesapp "github.com/armandoalvarado/sofia-backend/internal/activities/application"
	actionsapp "github.com/armandoalvarado/sofia-backend/internal/ai/actions/application"
	vertexinfra "github.com/armandoalvarado/sofia-backend/internal/ai/embeddings/infrastructure/vertex"
	runtimeapp "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/application"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	runtimeinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure"
	deepseekinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure/deepseek"
	authjwt "github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure/jwt"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	conversationsapp "github.com/armandoalvarado/sofia-backend/internal/conversations/application"
	ingestionapp "github.com/armandoalvarado/sofia-backend/internal/ingestion/application"
	insightsapp "github.com/armandoalvarado/sofia-backend/internal/insights/application"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
	memoryapp "github.com/armandoalvarado/sofia-backend/internal/memory/application"
	notificationsapp "github.com/armandoalvarado/sofia-backend/internal/notifications/application"
	privacyapp "github.com/armandoalvarado/sofia-backend/internal/privacy/application"
	remindersapp "github.com/armandoalvarado/sofia-backend/internal/reminders/application"
	toolsapp "github.com/armandoalvarado/sofia-backend/internal/tools/application"
	usersapp "github.com/armandoalvarado/sofia-backend/internal/users/application"
)

type Modules struct {
	TokenService  *authjwt.Service
	Users         *usersapp.Service
	Activities    *activitiesapp.Service
	Reminders     *remindersapp.Service
	Insights      *insightsapp.Service
	Memory        *memoryapp.Service
	Learning      *learningapp.Service
	Notifications *notificationsapp.Service
	Tools         *toolsapp.Service
	Actions       *actionsapp.Service
	Privacy       *privacyapp.Service
	AIRuntime     *runtimeapp.RuntimeService
	Conversations *conversationsapp.Service
	Ingestion     *ingestionapp.Service
}

func BuildModules(cfg config.Config, repositories *Repositories) (*Modules, error) {
	tokenService := authjwt.New(cfg.JWTSecret, cfg.JWTAccessTokenTTL)
	usersService := usersapp.NewService(repositories.Users, repositories.Profiles, repositories.AISettings)
	activitiesService := activitiesapp.NewService(repositories.Activities)
	remindersService := remindersapp.NewService(repositories.Reminders, repositories.Activities, usersService)
	insightsService := insightsapp.NewService(repositories.Moods, repositories.Outcomes, repositories.Reflections, repositories.Activities)
	var embeddingProvider memoryapp.EmbeddingProvider
	if cfg.EmbeddingsEnabled {
		var err error
		embeddingProvider, err = vertexinfra.NewClient(context.Background(), cfg.GoogleCloudProject, cfg.EmbeddingsModel)
		if err != nil {
			return nil, err
		}
	}
	memoryService := memoryapp.NewService(repositories.Memories, usersService, embeddingProvider)
	learningService := learningapp.NewService(repositories.Beliefs, repositories.PromptVersions, repositories.DailySummaries)
	learningService.SetContextRepository(repositories.UserContexts)
	learningService.SetEntityCandidateRepository(repositories.EntityCandidates, cfg.EntityPromotionThreshold)
	learningService.SetEmbeddingProvider(embeddingProvider, cfg.BeliefDedupeThreshold)
	notificationsService := notificationsapp.NewService(repositories.DeviceTokens)
	toolsService := toolsapp.NewService(repositories.Tools)
	if err := toolsService.SeedDefaultTools(context.Background()); err != nil {
		return nil, err
	}
	actionsService := actionsapp.NewService(repositories.ActionProposals, toolsService, usersService, activitiesService, remindersService, memoryService)
	privacyService := privacyapp.NewService(usersService, activitiesService, remindersService, insightsService, memoryService, actionsService, repositories.AuditLogs, repositories.DeleteRequests, repositories.Conversations, repositories.Messages)
	actionsService.SetAuditRecorder(privacyService)
	learningService.SetAuditRecorder(privacyService)
	actionsService.SetBeliefSearcher(learningService)
	actionsService.SetAutonomyThreshold(cfg.AutonomyThreshold)
	activitiesService.SetReminderBridge(remindersService)

	contextBuilder := runtimeapp.NewContextBuilder(usersService, activitiesService, remindersService, insightsService, memoryService, 5, cfg.ContextTokenBudget)
	contextBuilder.SetPromptBaseReader(learningService)
	contextBuilder.SetLearningContextReader(learningService)
	toolSelector := runtimeapp.NewToolSelector(toolsService)
	modelClient, err := BuildModelClient(cfg)
	if err != nil {
		return nil, err
	}
	learningService.SetEpisodeStore(repositories.Episodes, modelClient, cfg.EpisodeMinSalience, cfg.EpisodeMaxPerDay)
	contextBuilder.SetExtractor(modelClient)
	planner := runtimeapp.NewPlanner(modelClient)
	safety := runtimeapp.NewSafetyPolicy()
	safety.SetAutonomy(actionsService, cfg.AutonomyThreshold)
	runtimeService := runtimeapp.NewRuntimeService(contextBuilder, toolSelector, planner, safety, actionsService)
	runtimeService.SetAuditRecorder(privacyService)
	conversationsService := conversationsapp.NewService(repositories.Conversations, repositories.Messages, runtimeService)
	ingestionService := ingestionapp.NewService(repositories.IngestionBatches, learningService)

	return &Modules{
		TokenService:  tokenService,
		Users:         usersService,
		Activities:    activitiesService,
		Reminders:     remindersService,
		Insights:      insightsService,
		Memory:        memoryService,
		Learning:      learningService,
		Notifications: notificationsService,
		Tools:         toolsService,
		Actions:       actionsService,
		Privacy:       privacyService,
		AIRuntime:     runtimeService,
		Conversations: conversationsService,
		Ingestion:     ingestionService,
	}, nil
}

func BuildModelClient(cfg config.Config) (runtimedomain.ModelClient, error) {
	if !cfg.AIRuntimeEnabled {
		return runtimeinfra.NewFakeModelClient(), nil
	}
	switch cfg.AIModelProvider {
	case "fake", "":
		return runtimeinfra.NewFakeModelClient(), nil
	case "deepseek":
		if cfg.DeepSeekBaseURL != "" {
			return deepseekinfra.NewClient(cfg.DeepSeekAPIKey, cfg.DeepSeekModels, deepseekinfra.WithEndpoint(cfg.DeepSeekBaseURL))
		}
		return deepseekinfra.NewClient(cfg.DeepSeekAPIKey, cfg.DeepSeekModels)
	default:
		return nil, errors.New("unsupported AI model provider")
	}
}
