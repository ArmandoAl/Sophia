package server

import (
	"net/http"

	activitiesHTTP "github.com/armandoalvarado/sofia-backend/internal/activities/interfaces/http"
	actionsHTTP "github.com/armandoalvarado/sofia-backend/internal/ai/actions/interfaces/http"
	runtimeHTTP "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/interfaces/http"
	authHTTP "github.com/armandoalvarado/sofia-backend/internal/auth/interfaces/http"
	conversationsHTTP "github.com/armandoalvarado/sofia-backend/internal/conversations/interfaces/http"
	ingestionHTTP "github.com/armandoalvarado/sofia-backend/internal/ingestion/interfaces/http"
	insightsHTTP "github.com/armandoalvarado/sofia-backend/internal/insights/interfaces/http"
	learningHTTP "github.com/armandoalvarado/sofia-backend/internal/learning/interfaces/http"
	memoryHTTP "github.com/armandoalvarado/sofia-backend/internal/memory/interfaces/http"
	notificationsHTTP "github.com/armandoalvarado/sofia-backend/internal/notifications/interfaces/http"
	privacyHTTP "github.com/armandoalvarado/sofia-backend/internal/privacy/interfaces/http"
	remindersHTTP "github.com/armandoalvarado/sofia-backend/internal/reminders/interfaces/http"
	toolsHTTP "github.com/armandoalvarado/sofia-backend/internal/tools/interfaces/http"
	usersHTTP "github.com/armandoalvarado/sofia-backend/internal/users/interfaces/http"
)

type Routes struct {
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
	Health        HealthHandler
	Authn         Middleware
	AuthRateLimit Middleware
}

func (r Routes) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/health", r.Health)
	mux.Handle("/auth/register", r.withAuthRateLimit(http.HandlerFunc(r.Auth.Register)))
	mux.Handle("/auth/login", r.withAuthRateLimit(http.HandlerFunc(r.Auth.Login)))
	mux.Handle("/users/me", r.Authn(http.HandlerFunc(r.Users.Me)))
	mux.Handle("/users/me/export", r.Authn(http.HandlerFunc(r.Privacy.Export)))
	mux.Handle("/users/me/delete-request", r.Authn(http.HandlerFunc(r.Privacy.DeleteRequest)))
	mux.Handle("/users/me/profile", r.Authn(http.HandlerFunc(r.Users.Profile)))
	mux.Handle("/users/me/ai-settings", r.Authn(http.HandlerFunc(r.Users.AISettings)))
	mux.Handle("/users/me/onboarding/complete", r.Authn(http.HandlerFunc(r.Users.CompleteOnboarding)))
	mux.Handle("/activities", r.Authn(http.HandlerFunc(r.Activities.Collection)))
	mux.Handle("/activities/", r.Authn(http.HandlerFunc(r.Activities.Resource)))
	mux.Handle("/reminders", r.Authn(http.HandlerFunc(r.Reminders.Collection)))
	mux.Handle("/reminders/due", r.Authn(http.HandlerFunc(r.Reminders.Due)))
	mux.Handle("/reminders/", r.Authn(http.HandlerFunc(r.Reminders.Resource)))
	mux.Handle("/insights/mood", r.Authn(http.HandlerFunc(r.Insights.Mood)))
	mux.Handle("/insights/outcomes", r.Authn(http.HandlerFunc(r.Insights.Outcomes)))
	mux.Handle("/insights/reflections", r.Authn(http.HandlerFunc(r.Insights.Reflections)))
	mux.Handle("/insights/summary", r.Authn(http.HandlerFunc(r.Insights.Summary)))
	mux.Handle("/memory", r.Authn(http.HandlerFunc(r.Memory.Collection)))
	mux.Handle("/memory/search", r.Authn(http.HandlerFunc(r.Memory.Search)))
	mux.Handle("/memory/", r.Authn(http.HandlerFunc(r.Memory.Resource)))
	mux.Handle("/contexts", r.Authn(http.HandlerFunc(r.Contexts.Collection)))
	mux.Handle("/contexts/", r.Authn(http.HandlerFunc(r.Contexts.Resource)))
	mux.Handle("/learning/beliefs", r.Authn(http.HandlerFunc(r.Contexts.Beliefs)))
	mux.Handle("/learning/beliefs/", r.Authn(http.HandlerFunc(r.Contexts.BeliefResource)))
	mux.Handle("/learning/prompt-version", r.Authn(http.HandlerFunc(r.Contexts.PromptVersion)))
	mux.Handle("/learning/summaries", r.Authn(http.HandlerFunc(r.Contexts.Summaries)))
	mux.Handle("/learning/summaries/", r.Authn(http.HandlerFunc(r.Contexts.SummaryResource)))
	mux.Handle("/notifications/device-tokens", r.Authn(http.HandlerFunc(r.Notifications.Collection)))
	mux.Handle("/notifications/device-tokens/", r.Authn(http.HandlerFunc(r.Notifications.Resource)))
	mux.Handle("/tools", r.Authn(http.HandlerFunc(r.Tools.Collection)))
	mux.Handle("/tools/", r.Authn(http.HandlerFunc(r.Tools.Resource)))
	mux.Handle("/ai/action-proposals", r.Authn(http.HandlerFunc(r.AIActions.Collection)))
	mux.Handle("/ai/action-proposals/", r.Authn(http.HandlerFunc(r.AIActions.Resource)))
	mux.Handle("/ai/runtime/message", r.Authn(http.HandlerFunc(r.AIRuntime.Message)))
	mux.Handle("/conversations", r.Authn(http.HandlerFunc(r.Conversations.Collection)))
	mux.Handle("/conversations/", r.Authn(http.HandlerFunc(r.Conversations.Resource)))
	mux.Handle("/ingestion/conversations", r.Authn(http.HandlerFunc(r.Ingestion.Conversations)))
	mux.Handle("/ingestion/batches", r.Authn(http.HandlerFunc(r.Ingestion.Batches)))
	mux.Handle("/ingestion/batches/", r.Authn(http.HandlerFunc(r.Ingestion.Batches)))
	return mux
}

func (r Routes) withAuthRateLimit(handler http.Handler) http.Handler {
	if r.AuthRateLimit == nil {
		return handler
	}
	return r.AuthRateLimit(handler)
}
