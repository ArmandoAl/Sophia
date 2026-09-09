package app

import (
	"time"

	activitiesHTTP "github.com/armandoalvarado/sofia-backend/internal/activities/interfaces/http"
	actionsHTTP "github.com/armandoalvarado/sofia-backend/internal/ai/actions/interfaces/http"
	runtimeHTTP "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/interfaces/http"
	authapp "github.com/armandoalvarado/sofia-backend/internal/auth/application"
	authdomain "github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	authjwt "github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure/jwt"
	authHTTP "github.com/armandoalvarado/sofia-backend/internal/auth/interfaces/http"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	conversationsHTTP "github.com/armandoalvarado/sofia-backend/internal/conversations/interfaces/http"
	ingestionHTTP "github.com/armandoalvarado/sofia-backend/internal/ingestion/interfaces/http"
	insightsHTTP "github.com/armandoalvarado/sofia-backend/internal/insights/interfaces/http"
	learningHTTP "github.com/armandoalvarado/sofia-backend/internal/learning/interfaces/http"
	memoryHTTP "github.com/armandoalvarado/sofia-backend/internal/memory/interfaces/http"
	notificationsHTTP "github.com/armandoalvarado/sofia-backend/internal/notifications/interfaces/http"
	privacyHTTP "github.com/armandoalvarado/sofia-backend/internal/privacy/interfaces/http"
	remindersHTTP "github.com/armandoalvarado/sofia-backend/internal/reminders/interfaces/http"
	"github.com/armandoalvarado/sofia-backend/internal/server"
	toolsHTTP "github.com/armandoalvarado/sofia-backend/internal/tools/interfaces/http"
	usersHTTP "github.com/armandoalvarado/sofia-backend/internal/users/interfaces/http"
)

type Handlers struct {
	Auth          *authHTTP.AuthHandler
	Users         *usersHTTP.Handler
	Activities    *activitiesHTTP.Handler
	Reminders     *remindersHTTP.Handler
	Insights      *insightsHTTP.Handler
	Memory        *memoryHTTP.Handler
	Contexts      *learningHTTP.Handler
	Notifications *notificationsHTTP.Handler
	Privacy       *privacyHTTP.Handler
	Tools         *toolsHTTP.Handler
	AIActions     *actionsHTTP.Handler
	AIRuntime     *runtimeHTTP.Handler
	Conversations *conversationsHTTP.Handler
	Ingestion     *ingestionHTTP.Handler
	rateLimit     int
	rateWindow    time.Duration
}

func BuildHandlers(cfg config.Config, repositories *Repositories, modules *Modules) *Handlers {
	registerUC := authapp.NewRegisterUserUseCase(repositories.Users)
	loginUC := authapp.NewLoginUserUseCase(repositories.Users)
	authHandler := authHTTP.NewAuthHandler(registerUC, loginUC, modules.TokenService, cfg.RequestBodyLimitBytes)
	authHandler.SetAuditRecorder(modules.Privacy)
	if _, ok := repositories.Users.(authdomain.TransactionalDefaultSaver); !ok {
		authHandler.SetPostRegisterHook(modules.Users.ProvisionDefaults)
	}

	return &Handlers{
		Auth:          authHandler,
		Users:         usersHTTP.NewHandler(modules.Users, cfg.RequestBodyLimitBytes),
		Activities:    activitiesHTTP.NewHandler(modules.Activities, cfg.RequestBodyLimitBytes),
		Reminders:     remindersHTTP.NewHandler(modules.Reminders, cfg.RequestBodyLimitBytes),
		Insights:      insightsHTTP.NewHandler(modules.Insights, cfg.RequestBodyLimitBytes),
		Memory:        memoryHTTP.NewHandler(modules.Memory, cfg.RequestBodyLimitBytes),
		Contexts:      learningHTTP.NewHandler(modules.Learning, cfg.RequestBodyLimitBytes),
		Notifications: notificationsHTTP.NewHandler(modules.Notifications, cfg.RequestBodyLimitBytes),
		Privacy:       privacyHTTP.NewHandler(modules.Privacy, cfg.RequestBodyLimitBytes),
		Tools:         toolsHTTP.NewHandler(modules.Tools),
		AIActions:     actionsHTTP.NewHandler(modules.Actions, cfg.RequestBodyLimitBytes),
		AIRuntime:     runtimeHTTP.NewHandler(modules.AIRuntime, cfg.RequestBodyLimitBytes),
		Conversations: conversationsHTTP.NewHandler(modules.Conversations, cfg.RequestBodyLimitBytes),
		Ingestion:     ingestionHTTP.NewHandler(modules.Ingestion, cfg.RequestBodyLimitBytes),
		rateLimit:     cfg.AuthRateLimitRequests,
		rateWindow:    cfg.AuthRateLimitWindow,
	}
}

func (h *Handlers) Routes(tokenService *authjwt.Service, environment, firestoreStatus string) server.Routes {
	return server.Routes{
		Auth:          h.Auth,
		Users:         h.Users,
		Activities:    h.Activities,
		Reminders:     h.Reminders,
		Insights:      h.Insights,
		Memory:        h.Memory,
		Contexts:      h.Contexts,
		Notifications: h.Notifications,
		Privacy:       h.Privacy,
		Tools:         h.Tools,
		AIActions:     h.AIActions,
		AIRuntime:     h.AIRuntime,
		Conversations: h.Conversations,
		Ingestion:     h.Ingestion,
		Health:        server.HealthHandler{Environment: environment, FirestoreStatus: firestoreStatus},
		Authn:         server.AuthMiddleware(tokenService),
		AuthRateLimit: server.NewRateLimiter(h.rateLimit, h.rateWindow).Middleware,
	}
}
