package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	activitiesapp "github.com/armandoalvarado/sofia-backend/internal/activities/application"
	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	activitiesinfra "github.com/armandoalvarado/sofia-backend/internal/activities/infrastructure"
	activitiesHTTP "github.com/armandoalvarado/sofia-backend/internal/activities/interfaces/http"
	actionsapp "github.com/armandoalvarado/sofia-backend/internal/ai/actions/application"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	actionsinfra "github.com/armandoalvarado/sofia-backend/internal/ai/actions/infrastructure"
	actionsHTTP "github.com/armandoalvarado/sofia-backend/internal/ai/actions/interfaces/http"
	runtimeapp "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/application"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	runtimeinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure"
	runtimeHTTP "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/interfaces/http"
	"github.com/armandoalvarado/sofia-backend/internal/auth/application"
	"github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure"
	authjwt "github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure/jwt"
	authHTTP "github.com/armandoalvarado/sofia-backend/internal/auth/interfaces/http"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	conversationsapp "github.com/armandoalvarado/sofia-backend/internal/conversations/application"
	conversationsdomain "github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
	conversationsinfra "github.com/armandoalvarado/sofia-backend/internal/conversations/infrastructure"
	conversationsHTTP "github.com/armandoalvarado/sofia-backend/internal/conversations/interfaces/http"
	insightsapp "github.com/armandoalvarado/sofia-backend/internal/insights/application"
	insightsdomain "github.com/armandoalvarado/sofia-backend/internal/insights/domain"
	insightsinfra "github.com/armandoalvarado/sofia-backend/internal/insights/infrastructure"
	insightsHTTP "github.com/armandoalvarado/sofia-backend/internal/insights/interfaces/http"
	memoryapp "github.com/armandoalvarado/sofia-backend/internal/memory/application"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	memoryinfra "github.com/armandoalvarado/sofia-backend/internal/memory/infrastructure"
	memoryHTTP "github.com/armandoalvarado/sofia-backend/internal/memory/interfaces/http"
	privacyapp "github.com/armandoalvarado/sofia-backend/internal/privacy/application"
	privacydomain "github.com/armandoalvarado/sofia-backend/internal/privacy/domain"
	privacyinfra "github.com/armandoalvarado/sofia-backend/internal/privacy/infrastructure"
	privacyHTTP "github.com/armandoalvarado/sofia-backend/internal/privacy/interfaces/http"
	remindersapp "github.com/armandoalvarado/sofia-backend/internal/reminders/application"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	remindersinfra "github.com/armandoalvarado/sofia-backend/internal/reminders/infrastructure"
	remindersHTTP "github.com/armandoalvarado/sofia-backend/internal/reminders/interfaces/http"
	"github.com/armandoalvarado/sofia-backend/internal/server"
	toolsapp "github.com/armandoalvarado/sofia-backend/internal/tools/application"
	toolsdomain "github.com/armandoalvarado/sofia-backend/internal/tools/domain"
	toolsinfra "github.com/armandoalvarado/sofia-backend/internal/tools/infrastructure"
	toolsHTTP "github.com/armandoalvarado/sofia-backend/internal/tools/interfaces/http"
	usersapp "github.com/armandoalvarado/sofia-backend/internal/users/application"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
	usersinfra "github.com/armandoalvarado/sofia-backend/internal/users/infrastructure"
	usersHTTP "github.com/armandoalvarado/sofia-backend/internal/users/interfaces/http"
)

const testBodyLimit = int64(1 << 20)

func TestRegisterSuccess(t *testing.T) {
	handler := newTestHandler()

	rec := performJSON(handler, http.MethodPost, "/auth/register", map[string]string{
		"name":     "Armando",
		"email":    "ARMANDO@example.com",
		"password": "strong-password",
	}, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp authHTTP.UserResponse
	decodeResponse(t, rec, &resp)
	if resp.ID == "" || resp.Name != "Armando" || resp.Email != "armando@example.com" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodPost, "/auth/register", map[string]string{
		"name":     "Armando",
		"email":    "armando@example.com",
		"password": "strong-password",
	}, "")

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, rec.Code, rec.Body.String())
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	handler := newTestHandler()

	rec := performJSON(handler, http.MethodPost, "/auth/register", map[string]string{
		"name":     "Armando",
		"email":    "armando@example.com",
		"password": "weak",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestLoginSuccessReturnsToken(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodPost, "/auth/login", map[string]string{
		"email":    "armando@example.com",
		"password": "strong-password",
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp authHTTP.LoginResponse
	decodeResponse(t, rec, &resp)
	if resp.Token == "" {
		t.Fatal("expected token in login response")
	}
	if resp.User.Email != "armando@example.com" {
		t.Fatalf("unexpected user response: %+v", resp.User)
	}
}

func TestLoginMissingUserReturnsUnauthorized(t *testing.T) {
	handler := newTestHandler()

	rec := performJSON(handler, http.MethodPost, "/auth/login", map[string]string{
		"email":    "missing@example.com",
		"password": "strong-password",
	}, "")

	assertUnauthorized(t, rec)
}

func TestLoginWrongPasswordReturnsUnauthorized(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodPost, "/auth/login", map[string]string{
		"email":    "armando@example.com",
		"password": "wrong-password",
	}, "")

	assertUnauthorized(t, rec)
}

func TestLoginRateLimit(t *testing.T) {
	handler := newTestHandlerWithAuthRateLimit(2, time.Minute)
	registerUser(t, handler, "armando@example.com", "strong-password")

	for i := 0; i < 2; i++ {
		rec := performJSON(handler, http.MethodPost, "/auth/login", map[string]string{
			"email":    "armando@example.com",
			"password": "wrong-password",
		}, "")
		assertUnauthorized(t, rec)
	}

	rec := performJSON(handler, http.MethodPost, "/auth/login", map[string]string{
		"email":    "armando@example.com",
		"password": "wrong-password",
	}, "")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d: %s", http.StatusTooManyRequests, rec.Code, rec.Body.String())
	}
}

func TestRegisterRateLimit(t *testing.T) {
	handler := newTestHandlerWithAuthRateLimit(2, time.Minute)

	for i := 0; i < 2; i++ {
		rec := performJSON(handler, http.MethodPost, "/auth/register", map[string]string{
			"name":     "Armando",
			"email":    "armando" + string(rune('a'+i)) + "@example.com",
			"password": "strong-password",
		}, "")
		if rec.Code != http.StatusCreated {
			t.Fatalf("register %d failed: status %d body %s", i, rec.Code, rec.Body.String())
		}
	}

	rec := performJSON(handler, http.MethodPost, "/auth/register", map[string]string{
		"name":     "Armando",
		"email":    "blocked@example.com",
		"password": "strong-password",
	}, "")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d: %s", http.StatusTooManyRequests, rec.Code, rec.Body.String())
	}
}

func TestLoginFailureCreatesAuditLog(t *testing.T) {
	routes, auditRepo := newTestRoutesWithAudit()
	handler := routes.Handler()
	registerUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodPost, "/auth/login", map[string]string{
		"email":    "armando@example.com",
		"password": "wrong-password",
	}, "")
	assertUnauthorized(t, rec)

	logs, err := auditRepo.ListByUser(context.Background(), "", 10)
	if err != nil {
		t.Fatal(err)
	}
	assertHasAuditAction(t, logs, "auth_login_failed")
}

func TestUsersMeWithoutTokenReturnsUnauthorized(t *testing.T) {
	handler := newTestHandler()

	rec := performJSON(handler, http.MethodGet, "/users/me", nil, "")

	assertUnauthorized(t, rec)
}

func TestUsersMeWithValidTokenReturnsUser(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")
	token := loginUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodGet, "/users/me", nil, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	decodeResponse(t, rec, &resp)
	if resp.User.Email != "armando@example.com" || resp.User.ID == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRegisterCreatesDefaultProfileAndAISettings(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")
	token := loginUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodGet, "/users/me", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		Profile struct {
			DisplayName         string `json:"display_name"`
			PreferredName       string `json:"preferred_name"`
			Timezone            string `json:"timezone"`
			Locale              string `json:"locale"`
			OnboardingCompleted bool   `json:"onboarding_completed"`
		} `json:"profile"`
		AISettings struct {
			AssistantName     string `json:"assistant_name"`
			ProactivityLevel  string `json:"proactivity_level"`
			AutonomyLevel     string `json:"autonomy_level"`
			MemoryEnabled     bool   `json:"memory_enabled"`
			RemindersEnabled  bool   `json:"reminders_enabled"`
			PlanningEnabled   bool   `json:"planning_enabled"`
			PreferredLanguage string `json:"preferred_language"`
		} `json:"ai_settings"`
	}
	decodeResponse(t, rec, &resp)

	if resp.Profile.DisplayName != "Armando" || resp.Profile.Timezone != usersdomain.DefaultTimezone || resp.Profile.Locale != usersdomain.DefaultLocale {
		t.Fatalf("unexpected default profile: %+v", resp.Profile)
	}
	if resp.AISettings.AssistantName != usersdomain.DefaultAssistantName ||
		resp.AISettings.ProactivityLevel != usersdomain.ProactivityMedium ||
		resp.AISettings.AutonomyLevel != usersdomain.AutonomySuggestive ||
		!resp.AISettings.MemoryEnabled ||
		!resp.AISettings.RemindersEnabled ||
		!resp.AISettings.PlanningEnabled {
		t.Fatalf("unexpected default ai settings: %+v", resp.AISettings)
	}
}

func TestUpdateProfileValidFields(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")
	token := loginUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodPatch, "/users/me/profile", map[string]string{
		"display_name":   "Armando Alvarado",
		"preferred_name": "Mando",
		"timezone":       "America/Mexico_City",
		"locale":         "es-MX",
		"avatar_url":     "https://example.com/avatar.png",
	}, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		DisplayName   string `json:"display_name"`
		PreferredName string `json:"preferred_name"`
		Timezone      string `json:"timezone"`
		AvatarURL     string `json:"avatar_url"`
	}
	decodeResponse(t, rec, &resp)
	if resp.DisplayName != "Armando Alvarado" || resp.PreferredName != "Mando" || resp.Timezone != "America/Mexico_City" || resp.AvatarURL == "" {
		t.Fatalf("unexpected profile response: %+v", resp)
	}
}

func TestUpdateProfileRejectsEmptyTimezone(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")
	token := loginUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodPatch, "/users/me/profile", map[string]string{
		"timezone": "",
	}, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestUpdateAISettingsValidFields(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")
	token := loginUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodPatch, "/users/me/ai-settings", map[string]any{
		"assistant_name":      "Sofi",
		"communication_style": "direct",
		"proactivity_level":   usersdomain.ProactivityHigh,
		"autonomy_level":      usersdomain.AutonomySemiAutonomous,
		"memory_enabled":      false,
		"reminders_enabled":   true,
		"planning_enabled":    false,
		"preferred_language":  "es-MX",
	}, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		AssistantName     string `json:"assistant_name"`
		ProactivityLevel  string `json:"proactivity_level"`
		AutonomyLevel     string `json:"autonomy_level"`
		MemoryEnabled     bool   `json:"memory_enabled"`
		PlanningEnabled   bool   `json:"planning_enabled"`
		PreferredLanguage string `json:"preferred_language"`
	}
	decodeResponse(t, rec, &resp)
	if resp.AssistantName != "Sofi" || resp.ProactivityLevel != usersdomain.ProactivityHigh || resp.AutonomyLevel != usersdomain.AutonomySemiAutonomous || resp.MemoryEnabled || resp.PlanningEnabled {
		t.Fatalf("unexpected ai settings response: %+v", resp)
	}
}

func TestGetAISettings(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")
	token := loginUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodGet, "/users/me/ai-settings", nil, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		AssistantName string `json:"assistant_name"`
		AutonomyLevel string `json:"autonomy_level"`
	}
	decodeResponse(t, rec, &resp)
	if resp.AssistantName != usersdomain.DefaultAssistantName || resp.AutonomyLevel != usersdomain.AutonomySuggestive {
		t.Fatalf("unexpected ai settings response: %+v", resp)
	}
}

func TestUpdateAISettingsRejectsInvalidAutonomyLevel(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")
	token := loginUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodPatch, "/users/me/ai-settings", map[string]string{
		"autonomy_level": "fully_autonomous",
	}, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCompleteOnboarding(t *testing.T) {
	handler := newTestHandler()
	registerUser(t, handler, "armando@example.com", "strong-password")
	token := loginUser(t, handler, "armando@example.com", "strong-password")

	rec := performJSON(handler, http.MethodPost, "/users/me/onboarding/complete", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		OnboardingCompleted bool `json:"onboarding_completed"`
	}
	decodeResponse(t, rec, &resp)
	if !resp.OnboardingCompleted {
		t.Fatal("expected onboarding to be completed")
	}
}

func TestCreateActivitySuccess(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createActivity(t, handler, token, map[string]any{
		"type":        activitiesdomain.TypeTask,
		"title":       "Plan sprint",
		"priority":    activitiesdomain.PriorityHigh,
		"timezone":    "America/Tijuana",
		"tags":        []string{"Work", "Sofia"},
		"description": "Prepare activities domain",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp activityTestResponse
	decodeResponse(t, rec, &resp)
	if resp.ID == "" || resp.UserID == "" || resp.Title != "Plan sprint" || resp.Status != activitiesdomain.StatusPending || resp.Priority != activitiesdomain.PriorityHigh {
		t.Fatalf("unexpected activity response: %+v", resp)
	}
}

func TestCreateActivityWithoutTitleFails(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createActivity(t, handler, token, map[string]any{
		"type":     activitiesdomain.TypeTask,
		"timezone": "America/Tijuana",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCreateActivityInvalidTypeFails(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createActivity(t, handler, token, map[string]any{
		"type":     "simple_task",
		"title":    "Bad type",
		"timezone": "America/Tijuana",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestListActivitiesOnlyAuthenticatedUser(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")

	createActivity(t, handler, tokenA, map[string]any{"title": "Mine", "timezone": "America/Tijuana"})
	createActivity(t, handler, tokenB, map[string]any{"title": "Other", "timezone": "America/Tijuana"})

	rec := performJSON(handler, http.MethodGet, "/activities", nil, tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp listActivitiesTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Activities) != 1 || resp.Activities[0].Title != "Mine" {
		t.Fatalf("unexpected activities: %+v", resp.Activities)
	}
}

func TestGetActivityExisting(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	activity := createActivityAndDecode(t, handler, token, "Read")

	rec := performJSON(handler, http.MethodGet, "/activities/"+activity.ID, nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp activityTestResponse
	decodeResponse(t, rec, &resp)
	if resp.ID != activity.ID || resp.Title != "Read" {
		t.Fatalf("unexpected activity: %+v", resp)
	}
}

func TestCannotAccessAnotherUsersActivity(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")
	activity := createActivityAndDecode(t, handler, tokenA, "Private")

	rec := performJSON(handler, http.MethodGet, "/activities/"+activity.ID, nil, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestUpdateActivity(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	activity := createActivityAndDecode(t, handler, token, "Draft")

	rec := performJSON(handler, http.MethodPatch, "/activities/"+activity.ID, map[string]any{
		"title":        "Updated",
		"status":       activitiesdomain.StatusActive,
		"priority":     activitiesdomain.PriorityUrgent,
		"has_reminder": true,
		"reminder_at":  "2026-07-01T09:00:00Z",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp activityTestResponse
	decodeResponse(t, rec, &resp)
	if resp.Title != "Updated" || resp.Status != activitiesdomain.StatusActive || resp.Priority != activitiesdomain.PriorityUrgent || !resp.HasReminder || resp.ReminderAt == nil {
		t.Fatalf("unexpected activity: %+v", resp)
	}
}

func TestCompleteCancelArchiveActivity(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	complete := createActivityAndDecode(t, handler, token, "Complete me")
	rec := performJSON(handler, http.MethodPost, "/activities/"+complete.ID+"/complete", nil, token)
	assertActivityStatus(t, rec, activitiesdomain.StatusCompleted)

	cancel := createActivityAndDecode(t, handler, token, "Cancel me")
	rec = performJSON(handler, http.MethodPost, "/activities/"+cancel.ID+"/cancel", nil, token)
	assertActivityStatus(t, rec, activitiesdomain.StatusCanceled)

	archive := createActivityAndDecode(t, handler, token, "Archive me")
	rec = performJSON(handler, http.MethodPost, "/activities/"+archive.ID+"/archive", nil, token)
	assertActivityStatus(t, rec, activitiesdomain.StatusArchived)
}

func TestListActivityFilters(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	createActivity(t, handler, token, map[string]any{
		"title":    "Habit",
		"type":     activitiesdomain.TypeHabit,
		"priority": activitiesdomain.PriorityLow,
		"timezone": "America/Tijuana",
		"tags":     []string{"health"},
	})
	createActivity(t, handler, token, map[string]any{
		"title":    "Goal",
		"type":     activitiesdomain.TypeGoal,
		"priority": activitiesdomain.PriorityHigh,
		"timezone": "America/Tijuana",
		"tags":     []string{"work"},
	})

	rec := performJSON(handler, http.MethodGet, "/activities?type=habit&priority=low&status=pending&tag=health", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp listActivitiesTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Activities) != 1 || resp.Activities[0].Title != "Habit" {
		t.Fatalf("unexpected filtered activities: %+v", resp.Activities)
	}
}

func TestCreateReminderSuccess(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createReminder(t, handler, token, map[string]any{
		"title":        "Call mom",
		"scheduled_at": "2026-07-01T09:00:00Z",
		"timezone":     "America/Tijuana",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp reminderTestResponse
	decodeResponse(t, rec, &resp)
	if resp.ID == "" || resp.Title != "Call mom" || resp.Status != remindersdomain.StatusPending || resp.NextRunAt.IsZero() {
		t.Fatalf("unexpected reminder: %+v", resp)
	}
}

func TestCreateReminderWithoutTitleFails(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createReminder(t, handler, token, map[string]any{
		"scheduled_at": "2026-07-01T09:00:00Z",
		"timezone":     "America/Tijuana",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCreateReminderInvalidScheduledAtFails(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createReminder(t, handler, token, map[string]any{
		"title":        "Bad clock",
		"scheduled_at": "tomorrow",
		"timezone":     "America/Tijuana",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestListRemindersOnlyAuthenticatedUser(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")

	createReminderAndDecode(t, handler, tokenA, "Mine", "2026-07-01T09:00:00Z")
	createReminderAndDecode(t, handler, tokenB, "Other", "2026-07-01T09:00:00Z")

	rec := performJSON(handler, http.MethodGet, "/reminders", nil, tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp listRemindersTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Reminders) != 1 || resp.Reminders[0].Title != "Mine" {
		t.Fatalf("unexpected reminders: %+v", resp.Reminders)
	}
}

func TestCannotAccessAnotherUsersReminder(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")
	reminder := createReminderAndDecode(t, handler, tokenA, "Private", "2026-07-01T09:00:00Z")

	rec := performJSON(handler, http.MethodGet, "/reminders/"+reminder.ID, nil, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestCancelAndArchiveReminder(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	cancel := createReminderAndDecode(t, handler, token, "Cancel me", "2026-07-01T09:00:00Z")
	rec := performJSON(handler, http.MethodPost, "/reminders/"+cancel.ID+"/cancel", nil, token)
	assertReminderStatus(t, rec, remindersdomain.StatusCanceled)

	archive := createReminderAndDecode(t, handler, token, "Archive me", "2026-07-01T09:00:00Z")
	rec = performJSON(handler, http.MethodPost, "/reminders/"+archive.ID+"/archive", nil, token)
	assertReminderStatus(t, rec, remindersdomain.StatusArchived)
}

func TestListDueReminders(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	createReminderAndDecode(t, handler, token, "Due", "2020-01-01T09:00:00Z")
	createReminderAndDecode(t, handler, token, "Future", "2999-01-01T09:00:00Z")

	rec := performJSON(handler, http.MethodGet, "/reminders/due", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp listRemindersTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Reminders) != 1 || resp.Reminders[0].Title != "Due" {
		t.Fatalf("unexpected due reminders: %+v", resp.Reminders)
	}
}

func TestCreateActivityWithReminderCreatesAssociatedReminder(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	activity := createActivityAndDecodePayload(t, handler, token, map[string]any{
		"title":        "With reminder",
		"timezone":     "America/Tijuana",
		"has_reminder": true,
		"reminder_at":  "2026-07-01T09:00:00Z",
	})

	rec := performJSON(handler, http.MethodGet, "/reminders", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp listRemindersTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Reminders) != 1 || resp.Reminders[0].ActivityID != activity.ID || resp.Reminders[0].Source != remindersdomain.SourceActivity {
		t.Fatalf("unexpected associated reminders: %+v", resp.Reminders)
	}
}

func TestCancelActivityCancelsPendingAssociatedReminder(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	activity := createActivityAndDecodePayload(t, handler, token, map[string]any{
		"title":        "Cancelable activity",
		"timezone":     "America/Tijuana",
		"has_reminder": true,
		"reminder_at":  "2026-07-01T09:00:00Z",
	})

	rec := performJSON(handler, http.MethodPost, "/activities/"+activity.ID+"/cancel", nil, token)
	assertActivityStatus(t, rec, activitiesdomain.StatusCanceled)

	rec = performJSON(handler, http.MethodGet, "/reminders", nil, token)
	var resp listRemindersTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Reminders) != 1 || resp.Reminders[0].Status != remindersdomain.StatusCanceled {
		t.Fatalf("expected associated reminder canceled, got %+v", resp.Reminders)
	}
}

func TestRemindersDisabledBlocksCreation(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := performJSON(handler, http.MethodPatch, "/users/me/ai-settings", map[string]any{
		"reminders_enabled": false,
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	rec = createReminder(t, handler, token, map[string]any{
		"title":        "Blocked",
		"scheduled_at": "2026-07-01T09:00:00Z",
		"timezone":     "America/Tijuana",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, rec.Code, rec.Body.String())
	}
}

func TestCreateMoodEntry(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createMood(t, handler, token, map[string]any{
		"mood":         insightsdomain.MoodGood,
		"energy_level": 4,
		"stress_level": 2,
		"focus_level":  5,
		"occurred_at":  "2026-07-01T09:00:00Z",
		"timezone":     "America/Tijuana",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp moodEntryTestResponse
	decodeResponse(t, rec, &resp)
	if resp.ID == "" || resp.Mood != insightsdomain.MoodGood || resp.EnergyLevel != 4 {
		t.Fatalf("unexpected mood entry: %+v", resp)
	}
}

func TestCreateMoodRejectsInvalidMood(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createMood(t, handler, token, map[string]any{
		"mood":         "ecstatic",
		"energy_level": 4,
		"stress_level": 2,
		"focus_level":  5,
		"occurred_at":  "2026-07-01T09:00:00Z",
		"timezone":     "America/Tijuana",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCreateMoodRejectsLevelOutOfRange(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createMood(t, handler, token, map[string]any{
		"mood":         insightsdomain.MoodNeutral,
		"energy_level": 6,
		"stress_level": 2,
		"focus_level":  5,
		"occurred_at":  "2026-07-01T09:00:00Z",
		"timezone":     "America/Tijuana",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCreateOutcomeForOwnActivity(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	activity := createActivityAndDecode(t, handler, token, "Write report")

	rec := createOutcome(t, handler, token, map[string]any{
		"activity_id":        activity.ID,
		"result":             insightsdomain.ResultCompleted,
		"completion_quality": 5,
		"satisfaction_level": 4,
		"difficulty_level":   3,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp outcomeTestResponse
	decodeResponse(t, rec, &resp)
	if resp.ActivityID != activity.ID || resp.Result != insightsdomain.ResultCompleted {
		t.Fatalf("unexpected outcome: %+v", resp)
	}
}

func TestCannotCreateOutcomeForAnotherUsersActivity(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")
	activity := createActivityAndDecode(t, handler, tokenA, "Private activity")

	rec := createOutcome(t, handler, tokenB, map[string]any{
		"activity_id":        activity.ID,
		"result":             insightsdomain.ResultCompleted,
		"completion_quality": 5,
		"satisfaction_level": 4,
		"difficulty_level":   3,
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

func TestCreateReflection(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createReflection(t, handler, token, map[string]any{
		"title":       "Sprint note",
		"content":     "Today I learned where planning got stuck.",
		"sentiment":   "neutral",
		"tags":        []string{"Work", "Sofia"},
		"occurred_at": "2026-07-01T10:00:00Z",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp reflectionTestResponse
	decodeResponse(t, rec, &resp)
	if resp.ID == "" || resp.Content == "" || len(resp.Tags) != 2 || resp.Tags[0] != "work" {
		t.Fatalf("unexpected reflection: %+v", resp)
	}
}

func TestListInsightsOnlyAuthenticatedUser(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")

	createMood(t, handler, tokenA, validMoodPayload("Mine"))
	createMood(t, handler, tokenB, validMoodPayload("Other"))
	createReflection(t, handler, tokenA, validReflectionPayload("Mine reflection"))
	createReflection(t, handler, tokenB, validReflectionPayload("Other reflection"))

	rec := performJSON(handler, http.MethodGet, "/insights/mood", nil, tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var moods moodEntriesListTestResponse
	decodeResponse(t, rec, &moods)
	if len(moods.MoodEntries) != 1 || moods.MoodEntries[0].Note != "Mine" {
		t.Fatalf("unexpected mood entries: %+v", moods.MoodEntries)
	}

	rec = performJSON(handler, http.MethodGet, "/insights/reflections", nil, tokenA)
	var reflections reflectionsListTestResponse
	decodeResponse(t, rec, &reflections)
	if len(reflections.Reflections) != 1 || reflections.Reflections[0].Content != "Mine reflection" {
		t.Fatalf("unexpected reflections: %+v", reflections.Reflections)
	}
}

func TestInsightsSummary(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	activity := createActivityAndDecode(t, handler, token, "Measure")

	createMood(t, handler, token, map[string]any{
		"mood":         insightsdomain.MoodGood,
		"energy_level": 4,
		"stress_level": 2,
		"focus_level":  5,
		"occurred_at":  "2026-07-01T09:00:00Z",
		"timezone":     "America/Tijuana",
	})
	createMood(t, handler, token, map[string]any{
		"mood":         insightsdomain.MoodNeutral,
		"energy_level": 2,
		"stress_level": 4,
		"focus_level":  3,
		"occurred_at":  "2026-07-02T09:00:00Z",
		"timezone":     "America/Tijuana",
	})
	createOutcome(t, handler, token, map[string]any{
		"activity_id":        activity.ID,
		"result":             insightsdomain.ResultCompleted,
		"completion_quality": 5,
		"satisfaction_level": 4,
		"difficulty_level":   3,
	})
	createReflection(t, handler, token, validReflectionPayload("Useful day"))

	rec := performJSON(handler, http.MethodGet, "/insights/summary?from=2026-01-01T00:00:00Z&to=2027-01-01T00:00:00Z", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp summaryTestResponse
	decodeResponse(t, rec, &resp)
	if resp.AverageEnergy != 3 || resp.AverageStress != 3 || resp.AverageFocus != 4 || resp.TotalOutcomes != 1 || resp.TotalReflections != 1 {
		t.Fatalf("unexpected summary: %+v", resp)
	}
}

func TestInsightsEndpointsRequireAuth(t *testing.T) {
	handler := newTestHandler()

	for _, path := range []string{"/insights/mood", "/insights/outcomes", "/insights/reflections", "/insights/summary"} {
		rec := performJSON(handler, http.MethodGet, path, nil, "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("path %s expected status %d, got %d: %s", path, http.StatusUnauthorized, rec.Code, rec.Body.String())
		}
	}
}

func TestCreateMemory(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := createMemory(t, handler, token, validMemoryPayload("Prefers morning planning", []string{"planning", "routine"}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp memoryTestResponse
	decodeResponse(t, rec, &resp)
	if resp.ID == "" || resp.Title != "Prefers morning planning" || resp.Type != memorydomain.TypePreference || resp.Status != memorydomain.StatusActive {
		t.Fatalf("unexpected memory: %+v", resp)
	}
}

func TestMemoryDisabledBlocksCreationAndSearch(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := performJSON(handler, http.MethodPatch, "/users/me/ai-settings", map[string]any{
		"memory_enabled": false,
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	rec = createMemory(t, handler, token, validMemoryPayload("Blocked", []string{"x"}))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, rec.Code, rec.Body.String())
	}

	rec = performJSON(handler, http.MethodPost, "/memory/search", map[string]any{"query": "Blocked"}, token)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, rec.Code, rec.Body.String())
	}
}

func TestListMemoriesOnlyAuthenticatedUser(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")

	createMemoryAndDecode(t, handler, tokenA, "Mine", []string{"private"})
	createMemoryAndDecode(t, handler, tokenB, "Other", []string{"private"})

	rec := performJSON(handler, http.MethodGet, "/memory", nil, tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp memoriesListTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Memories) != 1 || resp.Memories[0].Title != "Mine" {
		t.Fatalf("unexpected memories: %+v", resp.Memories)
	}
}

func TestGetMemoryAndPreventCrossUserAccess(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")
	memory := createMemoryAndDecode(t, handler, tokenA, "Private", []string{"secret"})

	rec := performJSON(handler, http.MethodGet, "/memory/"+memory.ID, nil, tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var own memoryTestResponse
	decodeResponse(t, rec, &own)
	if own.ID != memory.ID || own.LastAccessedAt == nil {
		t.Fatalf("unexpected own memory: %+v", own)
	}

	rec = performJSON(handler, http.MethodGet, "/memory/"+memory.ID, nil, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestUpdateArchiveAndSoftDeleteMemory(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	memory := createMemoryAndDecode(t, handler, token, "Original", []string{"old"})

	rec := performJSON(handler, http.MethodPatch, "/memory/"+memory.ID, map[string]any{
		"title":      "Updated",
		"importance": memorydomain.ImportanceCritical,
		"tags":       []string{"new", "important"},
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var updated memoryTestResponse
	decodeResponse(t, rec, &updated)
	if updated.Title != "Updated" || updated.Importance != memorydomain.ImportanceCritical || updated.Tags[0] != "new" {
		t.Fatalf("unexpected updated memory: %+v", updated)
	}

	rec = performJSON(handler, http.MethodPost, "/memory/"+memory.ID+"/archive", nil, token)
	assertMemoryStatus(t, rec, memorydomain.StatusArchived)

	rec = performJSON(handler, http.MethodDelete, "/memory/"+memory.ID, nil, token)
	assertMemoryStatus(t, rec, memorydomain.StatusDeleted)

	rec = performJSON(handler, http.MethodGet, "/memory/"+memory.ID, nil, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestSearchMemoryByTextTagAndType(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	createMemoryAndDecode(t, handler, token, "Planning preference", []string{"planning"})
	createMemory(t, handler, token, map[string]any{
		"type":       memorydomain.TypeProject,
		"title":      "Sofia backend",
		"content":    "Firestore indexing notes",
		"source":     memorydomain.SourceUserManual,
		"importance": memorydomain.ImportanceHigh,
		"tags":       []string{"backend"},
	})

	rec := performJSON(handler, http.MethodPost, "/memory/search", map[string]any{
		"query": "planning",
		"tag":   "planning",
		"type":  memorydomain.TypePreference,
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp memoriesListTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Memories) != 1 || resp.Memories[0].Title != "Planning preference" {
		t.Fatalf("unexpected search results: %+v", resp.Memories)
	}
}

func TestMemoryEndpointsRequireAuth(t *testing.T) {
	handler := newTestHandler()

	for _, tc := range []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodGet, "/memory", nil},
		{http.MethodPost, "/memory", validMemoryPayload("No auth", []string{"x"})},
		{http.MethodPost, "/memory/search", map[string]any{"query": "x"}},
		{http.MethodGet, "/memory/id", nil},
	} {
		rec := performJSON(handler, tc.method, tc.path, tc.body, "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s expected status %d, got %d: %s", tc.method, tc.path, http.StatusUnauthorized, rec.Code, rec.Body.String())
		}
	}
}

func TestSeedAndListTools(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := performJSON(handler, http.MethodGet, "/tools", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	decodeResponse(t, rec, &resp)
	if len(resp.Tools) < 8 {
		t.Fatalf("expected default tools, got %+v", resp.Tools)
	}

	rec = performJSON(handler, http.MethodGet, "/tools/"+toolsdomain.ToolCreateActivity, nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestCreateRejectAndConfirmActionProposal(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	proposal := createProposalAndDecode(t, handler, token, toolsdomain.ToolCreateActivity, actionsdomain.RiskLow, map[string]any{
		"title":    "Proposed activity",
		"timezone": "America/Tijuana",
	})
	if proposal.Status != actionsdomain.StatusProposed || !proposal.RequiresConfirmation {
		t.Fatalf("unexpected proposal: %+v", proposal)
	}

	rec := performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/reject", nil, token)
	assertProposalStatus(t, rec, actionsdomain.StatusRejected)

	proposal = createProposalAndDecode(t, handler, token, toolsdomain.ToolCreateActivity, actionsdomain.RiskLow, map[string]any{
		"title":    "Confirm me",
		"timezone": "America/Tijuana",
	})
	rec = performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/confirm", nil, token)
	assertProposalStatus(t, rec, actionsdomain.StatusConfirmed)
}

func TestCannotAccessAnotherUsersActionProposal(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")
	proposal := createProposalAndDecode(t, handler, tokenA, toolsdomain.ToolCreateActivity, actionsdomain.RiskLow, map[string]any{
		"title":    "Private proposal",
		"timezone": "America/Tijuana",
	})

	rec := performJSON(handler, http.MethodGet, "/ai/action-proposals/"+proposal.ID, nil, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestExecuteConfirmedCreateActivityProposal(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	proposal := createProposalAndDecode(t, handler, token, toolsdomain.ToolCreateActivity, actionsdomain.RiskLow, map[string]any{
		"title":    "Executed activity",
		"timezone": "America/Tijuana",
	})
	performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/confirm", nil, token)

	rec := performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/execute", nil, token)
	assertProposalStatus(t, rec, actionsdomain.StatusExecuted)

	rec = performJSON(handler, http.MethodGet, "/activities", nil, token)
	var activities listActivitiesTestResponse
	decodeResponse(t, rec, &activities)
	if len(activities.Activities) != 1 || activities.Activities[0].Title != "Executed activity" {
		t.Fatalf("expected executed activity, got %+v", activities.Activities)
	}
}

func TestExecuteConfirmedCreateReminderProposal(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	proposal := createProposalAndDecode(t, handler, token, toolsdomain.ToolCreateReminder, actionsdomain.RiskLow, map[string]any{
		"title":        "Executed reminder",
		"scheduled_at": "2026-07-01T09:00:00Z",
		"timezone":     "America/Tijuana",
	})
	performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/confirm", nil, token)
	rec := performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/execute", nil, token)
	assertProposalStatus(t, rec, actionsdomain.StatusExecuted)
}

func TestExecuteConfirmedCreateMemoryProposalAndBlockWhenDisabled(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	proposal := createProposalAndDecode(t, handler, token, toolsdomain.ToolCreateMemory, actionsdomain.RiskLow, map[string]any{
		"title":   "Executed memory",
		"content": "Remember this safe preference.",
	})
	performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/confirm", nil, token)
	rec := performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/execute", nil, token)
	assertProposalStatus(t, rec, actionsdomain.StatusExecuted)

	rec = performJSON(handler, http.MethodPatch, "/users/me/ai-settings", map[string]any{"memory_enabled": false}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	proposal = createProposalAndDecode(t, handler, token, toolsdomain.ToolCreateMemory, actionsdomain.RiskLow, map[string]any{
		"title":   "Blocked memory",
		"content": "Should fail.",
	})
	performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/confirm", nil, token)
	rec = performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/execute", nil, token)
	assertProposalStatus(t, rec, actionsdomain.StatusFailed)
}

func TestManualAutonomyBlocksUnconfirmedExecution(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	proposal := createProposalAndDecode(t, handler, token, toolsdomain.ToolCreateActivity, actionsdomain.RiskLow, map[string]any{
		"title":    "Needs confirmation",
		"timezone": "America/Tijuana",
	})

	rec := performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/execute", nil, token)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, rec.Code, rec.Body.String())
	}
}

func TestAIRuntimeWithoutTokenReturnsUnauthorized(t *testing.T) {
	handler := newTestHandler()

	rec := performJSON(handler, http.MethodPost, "/ai/runtime/message", map[string]any{
		"message": "Recuérdame estudiar mañana",
		"dry_run": true,
	}, "")

	assertUnauthorized(t, rec)
}

func TestCreateConversationAuthenticated(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")

	conversation := createConversationAndDecode(t, handler, token, "V1 chat")

	if conversation.ID == "" || conversation.Title != "V1 chat" || conversation.Status != conversationsdomain.StatusActive {
		t.Fatalf("unexpected conversation: %+v", conversation)
	}
}

func TestListConversationsOnlyOwn(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")
	createConversationAndDecode(t, handler, tokenA, "Own")
	createConversationAndDecode(t, handler, tokenB, "Other")

	rec := performJSON(handler, http.MethodGet, "/conversations", nil, tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp conversationsListTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Conversations) != 1 || resp.Conversations[0].Title != "Own" {
		t.Fatalf("unexpected conversations: %+v", resp.Conversations)
	}
}

func TestCannotAccessAnotherUsersConversation(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")
	conversation := createConversationAndDecode(t, handler, tokenA, "Private")

	rec := performJSON(handler, http.MethodGet, "/conversations/"+conversation.ID, nil, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestSendConversationMessagePersistsUserAndAssistant(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	conversation := createConversationAndDecode(t, handler, token, "Chat")

	rec := performJSON(handler, http.MethodPost, "/conversations/"+conversation.ID+"/messages", map[string]any{
		"content": "Hola Sofía",
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}
	var resp conversationSendTestResponse
	decodeResponse(t, rec, &resp)
	if resp.UserMessage.Role != conversationsdomain.RoleUser || resp.UserMessage.Content != "Hola Sofía" {
		t.Fatalf("unexpected user message: %+v", resp.UserMessage)
	}
	if resp.AssistantMessage.Role != conversationsdomain.RoleAssistant || resp.AssistantMessage.Content == "" {
		t.Fatalf("unexpected assistant message: %+v", resp.AssistantMessage)
	}
	if resp.RuntimeRequestID == "" || resp.AssistantMessage.RuntimeRequestID != resp.RuntimeRequestID {
		t.Fatalf("expected runtime request id, got %+v", resp)
	}
	if len(resp.ProposedActions) != 0 {
		t.Fatalf("dry-run conversation send should not persist/return auto proposals for generic message: %+v", resp.ProposedActions)
	}
}

func TestConversationMessagesReturnChronologicalOrder(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	conversation := createConversationAndDecode(t, handler, token, "Chat")
	performJSON(handler, http.MethodPost, "/conversations/"+conversation.ID+"/messages", map[string]any{"content": "First"}, token)
	performJSON(handler, http.MethodPost, "/conversations/"+conversation.ID+"/messages", map[string]any{"content": "Second"}, token)

	rec := performJSON(handler, http.MethodGet, "/conversations/"+conversation.ID+"/messages", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp conversationMessagesListTestResponse
	decodeResponse(t, rec, &resp)
	if len(resp.Messages) != 4 {
		t.Fatalf("expected 4 messages, got %+v", resp.Messages)
	}
	if resp.Messages[0].Content != "First" || resp.Messages[1].Role != conversationsdomain.RoleAssistant || resp.Messages[2].Content != "Second" {
		t.Fatalf("messages are not chronological: %+v", resp.Messages)
	}
}

func TestConversationProviderErrorKeepsUserMessageOnly(t *testing.T) {
	conversations := conversationsinfra.NewInMemoryConversationRepository()
	messages := conversationsinfra.NewInMemoryConversationMessageRepository()
	service := conversationsapp.NewService(conversations, messages, failingRuntime{})
	conversation, err := service.CreateConversation(context.Background(), "user-1", conversationsdomain.CreateConversationInput{Title: "Chat"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.SendMessage(context.Background(), "user-1", conversation.ID, "Hola")
	if !errors.Is(err, errRuntimeFailed) {
		t.Fatalf("expected runtime error, got %v", err)
	}
	stored, err := service.ListMessages(context.Background(), "user-1", conversation.ID, 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 || stored[0].Role != conversationsdomain.RoleUser {
		t.Fatalf("expected only persisted user message, got %+v", stored)
	}
}

func TestConversationsRequireAuth(t *testing.T) {
	handler := newTestHandler()

	rec := performJSON(handler, http.MethodPost, "/conversations", map[string]any{"title": "Nope"}, "")

	assertUnauthorized(t, rec)
}

func TestUserExportOnlyIncludesAuthenticatedUsersData(t *testing.T) {
	handler := newTestHandler()
	tokenA := registerAndLogin(t, handler, "armando@example.com")
	tokenB := registerAndLogin(t, handler, "other@example.com")
	createActivityAndDecode(t, handler, tokenA, "Own activity")
	createActivityAndDecode(t, handler, tokenB, "Other activity")
	createMemoryAndDecode(t, handler, tokenA, "Own memory", []string{"own"})
	createMemoryAndDecode(t, handler, tokenB, "Other memory", []string{"other"})

	rec := performJSON(handler, http.MethodGet, "/users/me/export", nil, tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		Activities []struct {
			Title string `json:"title"`
		} `json:"activities"`
		Memories []struct {
			Title string `json:"title"`
		} `json:"memories"`
	}
	decodeResponse(t, rec, &resp)
	if len(resp.Activities) != 1 || resp.Activities[0].Title != "Own activity" {
		t.Fatalf("unexpected exported activities: %+v", resp.Activities)
	}
	if len(resp.Memories) != 1 || resp.Memories[0].Title != "Own memory" {
		t.Fatalf("unexpected exported memories: %+v", resp.Memories)
	}
}

func TestMemoryDeletedDoesNotAppearInExportOrSearch(t *testing.T) {
	handler := newTestHandler()
	token := registerAndLogin(t, handler, "armando@example.com")
	memory := createMemoryAndDecode(t, handler, token, "Delete me", []string{"delete"})

	rec := performJSON(handler, http.MethodDelete, "/memory/"+memory.ID, nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	rec = performJSON(handler, http.MethodPost, "/memory/search", map[string]any{"query": "Delete me"}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var searchResp struct {
		Memories []struct {
			ID string `json:"id"`
		} `json:"memories"`
	}
	decodeResponse(t, rec, &searchResp)
	if len(searchResp.Memories) != 0 {
		t.Fatalf("deleted memory appeared in search: %+v", searchResp.Memories)
	}

	rec = performJSON(handler, http.MethodGet, "/users/me/export", nil, token)
	var exportResp struct {
		Memories []struct {
			ID string `json:"id"`
		} `json:"memories"`
	}
	decodeResponse(t, rec, &exportResp)
	if len(exportResp.Memories) != 0 {
		t.Fatalf("deleted memory appeared in export: %+v", exportResp.Memories)
	}
}

func TestDeleteRequestCreatesAuditLog(t *testing.T) {
	routes, auditRepo := newTestRoutesWithAudit()
	handler := routes.Handler()
	token := registerAndLogin(t, handler, "armando@example.com")

	rec := performJSON(handler, http.MethodPost, "/users/me/delete-request", map[string]any{"reason": "please delete my data"}, token)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, rec.Code, rec.Body.String())
	}

	logs, err := auditRepo.ListByUser(context.Background(), userIDFromMe(t, handler, token), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].Action != "user_delete_requested" {
		t.Fatalf("expected delete audit log, got %+v", logs)
	}
}

func TestUserExportRequiresAuth(t *testing.T) {
	handler := newTestHandler()
	rec := performJSON(handler, http.MethodGet, "/users/me/export", nil, "")
	assertUnauthorized(t, rec)
}

func TestHealthCheck(t *testing.T) {
	handler := newTestHandler()

	rec := performJSON(handler, http.MethodGet, "/health", nil, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp server.HealthResponse
	decodeResponse(t, rec, &resp)
	if resp.Status != "ok" || resp.Environment != "test" || resp.Firestore != "disabled" {
		t.Fatalf("unexpected health response: %+v", resp)
	}
}

func TestCORSAllowedOrigin(t *testing.T) {
	app := server.New(config.Config{
		Env:                "test",
		Port:               "8080",
		CORSAllowedOrigins: []string{"https://app.example.com"},
		ReadTimeout:        time.Second,
		WriteTimeout:       time.Second,
		IdleTimeout:        time.Second,
		ReadHeaderTimeout:  time.Second,
	}, newTestRoutes())

	req := httptest.NewRequest(http.MethodOptions, "/users/me", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("unexpected CORS origin: %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type" {
		t.Fatalf("unexpected CORS headers: %q", got)
	}
}

func TestAIActionLifecycleCreatesAuditLogs(t *testing.T) {
	routes, auditRepo := newTestRoutesWithAudit()
	handler := routes.Handler()
	token := registerAndLogin(t, handler, "armando@example.com")
	userID := userIDFromMe(t, handler, token)

	proposal := createProposalAndDecode(t, handler, token, toolsdomain.ToolCreateActivity, actionsdomain.RiskLow, map[string]any{
		"title":    "Audited activity",
		"timezone": "America/Tijuana",
	})
	performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/confirm", nil, token)
	rec := performJSON(handler, http.MethodPost, "/ai/action-proposals/"+proposal.ID+"/execute", nil, token)
	assertProposalStatus(t, rec, actionsdomain.StatusExecuted)

	logs, err := auditRepo.ListByUser(context.Background(), userID, 20)
	if err != nil {
		t.Fatal(err)
	}
	assertHasAuditAction(t, logs, "ai_action_proposal_generated")
	assertHasAuditAction(t, logs, "ai_action_proposal_confirmed")
	assertHasAuditAction(t, logs, "ai_action_proposal_executed")
}

func newTestHandler() http.Handler {
	return newTestRoutes().Handler()
}

func newTestHandlerWithAuthRateLimit(limit int, window time.Duration) http.Handler {
	routes, _ := newTestRoutesWithAuditAndRateLimit(limit, window)
	return routes.Handler()
}

func newTestRoutes() server.Routes {
	routes, _ := newTestRoutesWithAudit()
	return routes
}

func newTestRoutesWithAudit() (server.Routes, *privacyinfra.InMemoryAuditLogRepository) {
	return newTestRoutesWithAuditAndRateLimit(0, 0)
}

func newTestRoutesWithAuditAndRateLimit(limit int, window time.Duration) (server.Routes, *privacyinfra.InMemoryAuditLogRepository) {
	repo := infrastructure.NewInMemoryUserRepository()
	profiles := usersinfra.NewInMemoryUserProfileRepository()
	settings := usersinfra.NewInMemoryAISettingsRepository()
	activities := activitiesinfra.NewInMemoryActivityRepository()
	reminders := remindersinfra.NewInMemoryReminderRepository()
	moods := insightsinfra.NewInMemoryMoodEntryRepository()
	outcomes := insightsinfra.NewInMemoryActivityOutcomeRepository()
	reflections := insightsinfra.NewInMemoryReflectionRepository()
	memories := memoryinfra.NewInMemoryMemoryRepository()
	toolRepo := toolsinfra.NewInMemoryToolDefinitionRepository()
	actionRepo := actionsinfra.NewInMemoryAIActionProposalRepository()
	conversationRepo := conversationsinfra.NewInMemoryConversationRepository()
	messageRepo := conversationsinfra.NewInMemoryConversationMessageRepository()
	auditRepo := privacyinfra.NewInMemoryAuditLogRepository()
	deleteRequestRepo := privacyinfra.NewInMemoryDeleteRequestRepository()
	usersService := usersapp.NewService(repo, profiles, settings)
	activitiesService := activitiesapp.NewService(activities)
	remindersService := remindersapp.NewService(reminders, activities, usersService)
	insightsService := insightsapp.NewService(moods, outcomes, reflections, activities)
	memoryService := memoryapp.NewService(memories, usersService, nil)
	toolsService := toolsapp.NewService(toolRepo)
	if err := toolsService.SeedDefaultTools(context.Background()); err != nil {
		panic(err)
	}
	actionsService := actionsapp.NewService(actionRepo, toolsService, usersService, activitiesService, remindersService, memoryService)
	privacyService := privacyapp.NewService(usersService, activitiesService, remindersService, insightsService, memoryService, actionsService, auditRepo, deleteRequestRepo)
	actionsService.SetAuditRecorder(privacyService)
	activitiesService.SetReminderBridge(remindersService)
	contextBuilder := runtimeapp.NewContextBuilder(usersService, activitiesService, remindersService, insightsService, memoryService, 5)
	toolSelector := runtimeapp.NewToolSelector(toolsService)
	planner := runtimeapp.NewPlanner(runtimeinfra.NewFakeModelClient())
	runtimeService := runtimeapp.NewRuntimeService(contextBuilder, toolSelector, planner, runtimeapp.NewSafetyPolicy(), actionsService)
	runtimeService.SetAuditRecorder(privacyService)
	conversationsService := conversationsapp.NewService(conversationRepo, messageRepo, runtimeService)
	tokenService := authjwt.New("test-secret")
	registerUC := application.NewRegisterUserUseCase(repo)
	loginUC := application.NewLoginUserUseCase(repo)

	authHandler := authHTTP.NewAuthHandler(registerUC, loginUC, tokenService, testBodyLimit)
	authHandler.SetPostRegisterHook(usersService.ProvisionDefaults)
	authHandler.SetAuditRecorder(privacyService)
	usersHandler := usersHTTP.NewHandler(usersService, testBodyLimit)
	activitiesHandler := activitiesHTTP.NewHandler(activitiesService, testBodyLimit)
	remindersHandler := remindersHTTP.NewHandler(remindersService, testBodyLimit)
	insightsHandler := insightsHTTP.NewHandler(insightsService, testBodyLimit)
	memoryHandler := memoryHTTP.NewHandler(memoryService, testBodyLimit)
	privacyHandler := privacyHTTP.NewHandler(privacyService, testBodyLimit)
	toolsHandler := toolsHTTP.NewHandler(toolsService)
	actionsHandler := actionsHTTP.NewHandler(actionsService, testBodyLimit)
	runtimeHandler := runtimeHTTP.NewHandler(runtimeService, testBodyLimit)
	conversationsHandler := conversationsHTTP.NewHandler(conversationsService, testBodyLimit)

	routes := server.Routes{
		Auth:          authHandler,
		Users:         usersHandler,
		Activities:    activitiesHandler,
		Reminders:     remindersHandler,
		Insights:      insightsHandler,
		Memory:        memoryHandler,
		Privacy:       privacyHandler,
		Tools:         toolsHandler,
		AIActions:     actionsHandler,
		AIRuntime:     runtimeHandler,
		Conversations: conversationsHandler,
		Health:        server.HealthHandler{Environment: "test", FirestoreStatus: "disabled"},
		Authn:         server.AuthMiddleware(tokenService),
	}
	if limit > 0 {
		routes.AuthRateLimit = server.NewRateLimiter(limit, window).Middleware
	}
	return routes, auditRepo
}

func registerUser(t *testing.T, handler http.Handler, email, password string) {
	t.Helper()

	rec := performJSON(handler, http.MethodPost, "/auth/register", map[string]string{
		"name":     "Armando",
		"email":    email,
		"password": password,
	}, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("register failed: status %d body %s", rec.Code, rec.Body.String())
	}
}

func loginUser(t *testing.T, handler http.Handler, email, password string) string {
	t.Helper()

	rec := performJSON(handler, http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: status %d body %s", rec.Code, rec.Body.String())
	}

	var resp authHTTP.LoginResponse
	decodeResponse(t, rec, &resp)
	return resp.Token
}

func userIDFromMe(t *testing.T, handler http.Handler, token string) string {
	t.Helper()
	rec := performJSON(handler, http.MethodGet, "/users/me", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("users/me failed: status %d body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	decodeResponse(t, rec, &resp)
	if resp.User.ID == "" {
		t.Fatal("expected user id")
	}
	return resp.User.ID
}

func performJSON(handler http.Handler, method, path string, payload any, token string) *httptest.ResponseRecorder {
	var body bytes.Buffer
	if payload != nil {
		_ = json.NewEncoder(&body).Encode(payload)
	}

	req := httptest.NewRequest(method, path, &body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}
	if err := json.NewDecoder(rec.Body).Decode(dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func assertUnauthorized(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnauthorized, rec.Code, rec.Body.String())
	}

	var resp struct {
		Message string `json:"message"`
	}
	decodeResponse(t, rec, &resp)
	if resp.Message != "invalid credentials" {
		t.Fatalf("expected generic invalid credentials message, got %q", resp.Message)
	}
}

type activityTestResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	HasReminder bool       `json:"has_reminder"`
	ReminderAt  *time.Time `json:"reminder_at"`
}

type listActivitiesTestResponse struct {
	Activities []activityTestResponse `json:"activities"`
}

type reminderTestResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ActivityID  string    `json:"activity_id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	ScheduledAt time.Time `json:"scheduled_at"`
	NextRunAt   time.Time `json:"next_run_at"`
	Source      string    `json:"source"`
}

type listRemindersTestResponse struct {
	Reminders []reminderTestResponse `json:"reminders"`
}

type moodEntryTestResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	ActivityID  string `json:"activity_id"`
	Mood        string `json:"mood"`
	EnergyLevel int    `json:"energy_level"`
	StressLevel int    `json:"stress_level"`
	FocusLevel  int    `json:"focus_level"`
	Note        string `json:"note"`
}

type outcomeTestResponse struct {
	ID                string `json:"id"`
	UserID            string `json:"user_id"`
	ActivityID        string `json:"activity_id"`
	Result            string `json:"result"`
	CompletionQuality int    `json:"completion_quality"`
	SatisfactionLevel int    `json:"satisfaction_level"`
	DifficultyLevel   int    `json:"difficulty_level"`
}

type reflectionTestResponse struct {
	ID         string   `json:"id"`
	UserID     string   `json:"user_id"`
	ActivityID string   `json:"activity_id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Sentiment  string   `json:"sentiment"`
	Tags       []string `json:"tags"`
}

type moodEntriesListTestResponse struct {
	MoodEntries []moodEntryTestResponse `json:"mood_entries"`
}

type reflectionsListTestResponse struct {
	Reflections []reflectionTestResponse `json:"reflections"`
}

type summaryTestResponse struct {
	AverageEnergy    float64 `json:"average_energy"`
	AverageStress    float64 `json:"average_stress"`
	AverageFocus     float64 `json:"average_focus"`
	TotalOutcomes    int     `json:"total_outcomes"`
	TotalReflections int     `json:"total_reflections"`
}

type memoryTestResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Type           string     `json:"type"`
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	Source         string     `json:"source"`
	Confidence     string     `json:"confidence"`
	Importance     string     `json:"importance"`
	Tags           []string   `json:"tags"`
	Visibility     string     `json:"visibility"`
	Status         string     `json:"status"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
}

type memoriesListTestResponse struct {
	Memories []memoryTestResponse `json:"memories"`
}

type proposalTestResponse struct {
	ID                   string          `json:"id"`
	UserID               string          `json:"user_id"`
	ToolName             string          `json:"tool_name"`
	Status               string          `json:"status"`
	ProposedInput        json.RawMessage `json:"proposed_input"`
	RiskLevel            string          `json:"risk_level"`
	RequiresConfirmation bool            `json:"requires_confirmation"`
	ExecutionError       string          `json:"execution_error"`
}

type conversationTestResponse struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Title         string     `json:"title"`
	Status        string     `json:"status"`
	LastMessageAt *time.Time `json:"last_message_at"`
}

type conversationMessageTestResponse struct {
	ID               string `json:"id"`
	ConversationID   string `json:"conversation_id"`
	UserID           string `json:"user_id"`
	Role             string `json:"role"`
	Content          string `json:"content"`
	RuntimeRequestID string `json:"runtime_request_id"`
}

type conversationsListTestResponse struct {
	Conversations []conversationTestResponse `json:"conversations"`
}

type conversationMessagesListTestResponse struct {
	Messages []conversationMessageTestResponse `json:"messages"`
}

type conversationSendTestResponse struct {
	Conversation     conversationTestResponse        `json:"conversation"`
	UserMessage      conversationMessageTestResponse `json:"user_message"`
	AssistantMessage conversationMessageTestResponse `json:"assistant_message"`
	ProposedActions  []json.RawMessage               `json:"proposed_actions"`
	RuntimeRequestID string                          `json:"runtime_request_id"`
}

func registerAndLogin(t *testing.T, handler http.Handler, email string) string {
	t.Helper()
	registerUser(t, handler, email, "strong-password")
	return loginUser(t, handler, email, "strong-password")
}

func createConversationAndDecode(t *testing.T, handler http.Handler, token, title string) conversationTestResponse {
	t.Helper()
	rec := performJSON(handler, http.MethodPost, "/conversations", map[string]any{"title": title}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create conversation failed: status %d body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Conversation conversationTestResponse `json:"conversation"`
	}
	decodeResponse(t, rec, &resp)
	return resp.Conversation
}

var errRuntimeFailed = errors.New("runtime failed")

type failingRuntime struct{}

func (failingRuntime) HandleMessage(ctx context.Context, request runtimedomain.RuntimeRequest) (*runtimedomain.RuntimeResponse, error) {
	return nil, errRuntimeFailed
}

func createActivity(t *testing.T, handler http.Handler, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return performJSON(handler, http.MethodPost, "/activities", payload, token)
}

func createActivityAndDecode(t *testing.T, handler http.Handler, token, title string) activityTestResponse {
	t.Helper()
	return createActivityAndDecodePayload(t, handler, token, map[string]any{
		"title":    title,
		"timezone": "America/Tijuana",
	})
}

func createActivityAndDecodePayload(t *testing.T, handler http.Handler, token string, payload map[string]any) activityTestResponse {
	t.Helper()
	rec := createActivity(t, handler, token, payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create activity failed: status %d body %s", rec.Code, rec.Body.String())
	}
	var resp activityTestResponse
	decodeResponse(t, rec, &resp)
	return resp
}

func createReminder(t *testing.T, handler http.Handler, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return performJSON(handler, http.MethodPost, "/reminders", payload, token)
}

func createReminderAndDecode(t *testing.T, handler http.Handler, token, title, scheduledAt string) reminderTestResponse {
	t.Helper()
	rec := createReminder(t, handler, token, map[string]any{
		"title":        title,
		"scheduled_at": scheduledAt,
		"timezone":     "America/Tijuana",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create reminder failed: status %d body %s", rec.Code, rec.Body.String())
	}
	var resp reminderTestResponse
	decodeResponse(t, rec, &resp)
	return resp
}

func assertActivityStatus(t *testing.T, rec *httptest.ResponseRecorder, status string) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp activityTestResponse
	decodeResponse(t, rec, &resp)
	if resp.Status != status {
		t.Fatalf("expected activity status %q, got %+v", status, resp)
	}
}

func assertReminderStatus(t *testing.T, rec *httptest.ResponseRecorder, status string) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp reminderTestResponse
	decodeResponse(t, rec, &resp)
	if resp.Status != status {
		t.Fatalf("expected reminder status %q, got %+v", status, resp)
	}
}

func createMood(t *testing.T, handler http.Handler, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return performJSON(handler, http.MethodPost, "/insights/mood", payload, token)
}

func createOutcome(t *testing.T, handler http.Handler, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return performJSON(handler, http.MethodPost, "/insights/outcomes", payload, token)
}

func createReflection(t *testing.T, handler http.Handler, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return performJSON(handler, http.MethodPost, "/insights/reflections", payload, token)
}

func validMoodPayload(note string) map[string]any {
	return map[string]any{
		"mood":         insightsdomain.MoodGood,
		"energy_level": 4,
		"stress_level": 2,
		"focus_level":  5,
		"note":         note,
		"occurred_at":  "2026-07-01T09:00:00Z",
		"timezone":     "America/Tijuana",
	}
}

func validReflectionPayload(content string) map[string]any {
	return map[string]any{
		"title":       "Reflection",
		"content":     content,
		"occurred_at": "2026-07-01T10:00:00Z",
	}
}

func createMemory(t *testing.T, handler http.Handler, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return performJSON(handler, http.MethodPost, "/memory", payload, token)
}

func createMemoryAndDecode(t *testing.T, handler http.Handler, token, title string, tags []string) memoryTestResponse {
	t.Helper()
	rec := createMemory(t, handler, token, validMemoryPayload(title, tags))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create memory failed: status %d body %s", rec.Code, rec.Body.String())
	}
	var resp memoryTestResponse
	decodeResponse(t, rec, &resp)
	return resp
}

func validMemoryPayload(title string, tags []string) map[string]any {
	return map[string]any{
		"type":       memorydomain.TypePreference,
		"title":      title,
		"content":    title + " is useful context for Sofia.",
		"summary":    title,
		"source":     memorydomain.SourceUserManual,
		"confidence": memorydomain.ConfidenceHigh,
		"importance": memorydomain.ImportanceMedium,
		"tags":       tags,
		"visibility": memorydomain.VisibilityAssistantContext,
	}
}

func assertMemoryStatus(t *testing.T, rec *httptest.ResponseRecorder, status string) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp memoryTestResponse
	decodeResponse(t, rec, &resp)
	if resp.Status != status {
		t.Fatalf("expected memory status %q, got %+v", status, resp)
	}
}

func createProposalAndDecode(t *testing.T, handler http.Handler, token, toolName, riskLevel string, input map[string]any) proposalTestResponse {
	t.Helper()
	rec := performJSON(handler, http.MethodPost, "/ai/action-proposals", map[string]any{
		"tool_name":      toolName,
		"risk_level":     riskLevel,
		"reason":         "test proposal",
		"proposed_input": input,
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create proposal failed: status %d body %s", rec.Code, rec.Body.String())
	}
	var resp proposalTestResponse
	decodeResponse(t, rec, &resp)
	return resp
}

func assertProposalStatus(t *testing.T, rec *httptest.ResponseRecorder, status string) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp proposalTestResponse
	decodeResponse(t, rec, &resp)
	if resp.Status != status {
		t.Fatalf("expected proposal status %q, got %+v", status, resp)
	}
}

func assertHasAuditAction(t *testing.T, logs []*privacydomain.AuditLog, action string) {
	t.Helper()
	for _, log := range logs {
		if log.Action == action {
			return
		}
	}
	t.Fatalf("expected audit action %q, got %+v", action, logs)
}
