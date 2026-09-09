package config

import (
	"testing"
	"time"
)

func TestLoadDefaultsDevelopment(t *testing.T) {
	t.Setenv("ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("PERSISTENCE_DRIVER", "")
	t.Setenv("REQUEST_BODY_LIMIT_BYTES", "")
	t.Setenv("CONTEXT_TOKEN_BUDGET", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Env != "development" {
		t.Fatalf("expected development env, got %q", cfg.Env)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected default port, got %q", cfg.Port)
	}
	if cfg.JWTSecret == "" {
		t.Fatal("expected development JWT secret")
	}
	if cfg.PersistenceDriver != "memory" {
		t.Fatalf("expected memory persistence, got %q", cfg.PersistenceDriver)
	}
	if cfg.FirestoreDatabaseID != "" {
		t.Fatalf("expected empty Firestore database ID by default, got %q", cfg.FirestoreDatabaseID)
	}
	if cfg.AIModelProvider != "fake" {
		t.Fatalf("expected fake model provider, got %q", cfg.AIModelProvider)
	}
	if cfg.ContextTokenBudget != 4600 {
		t.Fatalf("expected context token budget 4600, got %d", cfg.ContextTokenBudget)
	}
	if !cfg.AIRuntimeProposalOnly {
		t.Fatal("expected proposal-only runtime by default")
	}
	if cfg.AutonomyThreshold != 0.85 {
		t.Fatalf("expected default autonomy threshold 0.85, got %v", cfg.AutonomyThreshold)
	}
	if cfg.EmbeddingsEnabled || cfg.BeliefDedupeThreshold != 0.85 {
		t.Fatalf("unexpected default embeddings config: %+v", cfg)
	}
}

func TestLoadContextTokenBudget(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("CONTEXT_TOKEN_BUDGET", "3200")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ContextTokenBudget != 3200 {
		t.Fatalf("expected context token budget 3200, got %d", cfg.ContextTokenBudget)
	}
}

func TestLoadProductionRequiresSecureJWTAndFirestore(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("PERSISTENCE_DRIVER", "memory")

	if _, err := Load(); err == nil {
		t.Fatal("expected production config error")
	}
}

func TestLoadFirestoreRequiresProjectID(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("JWT_SECRET", "local-secret")
	t.Setenv("PERSISTENCE_DRIVER", "firestore")
	t.Setenv("FIRESTORE_PROJECT_ID", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing Firestore project ID error")
	}
}

func TestLoadFirestoreDatabaseID(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("JWT_SECRET", "local-secret")
	t.Setenv("PERSISTENCE_DRIVER", "firestore")
	t.Setenv("FIRESTORE_PROJECT_ID", "sofia-local")
	t.Setenv("FIRESTORE_DATABASE_ID", "default")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.FirestoreDatabaseID != "default" {
		t.Fatalf("expected Firestore database ID default, got %q", cfg.FirestoreDatabaseID)
	}
}

func TestLoadTimeouts(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("HTTP_READ_TIMEOUT", "2s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "3s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "4s")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "1s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ReadTimeout != 2*time.Second || cfg.WriteTimeout != 3*time.Second || cfg.IdleTimeout != 4*time.Second || cfg.ReadHeaderTimeout != time.Second {
		t.Fatalf("unexpected timeout config: %+v", cfg)
	}
}

func TestLoadAuthHardeningConfig(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("JWT_ACCESS_TOKEN_TTL", "15m")
	t.Setenv("AUTH_RATE_LIMIT_REQUESTS", "3")
	t.Setenv("AUTH_RATE_LIMIT_WINDOW", "30s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.JWTAccessTokenTTL != 15*time.Minute {
		t.Fatalf("expected JWT TTL 15m, got %s", cfg.JWTAccessTokenTTL)
	}
	if cfg.AuthRateLimitRequests != 3 {
		t.Fatalf("expected auth rate limit 3, got %d", cfg.AuthRateLimitRequests)
	}
	if cfg.AuthRateLimitWindow != 30*time.Second {
		t.Fatalf("expected auth rate window 30s, got %s", cfg.AuthRateLimitWindow)
	}
}

func TestLoadReminderWorkerConfig(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("REMINDER_WORKER_ENABLED", "true")
	t.Setenv("REMINDER_WORKER_ID", "worker-a")
	t.Setenv("REMINDER_WORKER_INTERVAL", "15s")
	t.Setenv("REMINDER_WORKER_BATCH_SIZE", "25")
	t.Setenv("REMINDER_WORKER_LEASE_DURATION", "45s")
	t.Setenv("REMINDER_DELIVERY_PROVIDER", "fcm")
	t.Setenv("FCM_ENABLED", "false")
	t.Setenv("FCM_DRY_RUN", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !cfg.ReminderWorkerEnabled {
		t.Fatal("expected reminder worker enabled")
	}
	if cfg.ReminderWorkerInterval != 15*time.Second {
		t.Fatalf("expected reminder worker interval 15s, got %s", cfg.ReminderWorkerInterval)
	}
	if cfg.ReminderWorkerID != "worker-a" {
		t.Fatalf("expected reminder worker ID worker-a, got %q", cfg.ReminderWorkerID)
	}
	if cfg.ReminderWorkerLease != 45*time.Second {
		t.Fatalf("expected reminder worker lease 45s, got %s", cfg.ReminderWorkerLease)
	}
	if cfg.ReminderWorkerBatchSize != 25 {
		t.Fatalf("expected reminder worker batch size 25, got %d", cfg.ReminderWorkerBatchSize)
	}
	if cfg.ReminderDeliveryProvider != "fcm" || !cfg.FCMDryRun || cfg.FCMEnabled {
		t.Fatalf("unexpected FCM config: %+v", cfg)
	}
}

func TestLoadSynthesisWorkerConfig(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("SYNTHESIS_WORKER_ENABLED", "true")
	t.Setenv("SYNTHESIS_WORKER_ID", "synth-a")
	t.Setenv("SYNTHESIS_WORKER_INTERVAL", "2m")
	t.Setenv("SYNTHESIS_WORKER_LEASE", "90s")
	t.Setenv("SYNTHESIS_RUN_HOUR_LOCAL", "4")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !cfg.SynthesisWorkerEnabled {
		t.Fatal("expected synthesis worker enabled")
	}
	if cfg.SynthesisWorkerID != "synth-a" {
		t.Fatalf("expected synth-a, got %q", cfg.SynthesisWorkerID)
	}
	if cfg.SynthesisWorkerInterval != 2*time.Minute || cfg.SynthesisWorkerLease != 90*time.Second || cfg.SynthesisRunHourLocal != 4 {
		t.Fatalf("unexpected synthesis config: %+v", cfg)
	}
}

func TestLoadRejectsInvalidSynthesisRunHour(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("SYNTHESIS_RUN_HOUR_LOCAL", "24")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid SYNTHESIS_RUN_HOUR_LOCAL error")
	}
}

func TestLoadRejectsInvalidReminderWorkerBatchSize(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("REMINDER_WORKER_BATCH_SIZE", "0")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid REMINDER_WORKER_BATCH_SIZE error")
	}
}

func TestLoadRejectsInvalidReminderDeliveryProvider(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("REMINDER_DELIVERY_PROVIDER", "sms")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid REMINDER_DELIVERY_PROVIDER error")
	}
}

func TestLoadRequiresFCMProjectWhenEnabled(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("REMINDER_DELIVERY_PROVIDER", "fcm")
	t.Setenv("FCM_ENABLED", "true")
	t.Setenv("FCM_PROJECT_ID", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing FCM_PROJECT_ID error")
	}
}

func TestLoadRejectsInvalidAIModelProvider(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("AI_MODEL_PROVIDER", "unknown")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid AI_MODEL_PROVIDER error")
	}
}

func TestLoadRejectsGeminiProvider(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("AI_MODEL_PROVIDER", "gemini")

	if _, err := Load(); err == nil {
		t.Fatal("expected unsupported Gemini provider error")
	}
}

func TestLoadDeepSeekRequiresAPIKey(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("AI_MODEL_PROVIDER", "deepseek")
	t.Setenv("DEEPSEEK_API_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing DEEPSEEK_API_KEY error")
	}
}

func TestLoadDeepSeekConfig(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("AI_MODEL_PROVIDER", "deepseek")
	t.Setenv("DEEPSEEK_API_KEY", "test-key")
	t.Setenv("DEEPSEEK_MODEL", "deepseek-test")
	t.Setenv("DEEPSEEK_MODEL_PLAN", "deepseek-plan")
	t.Setenv("DEEPSEEK_MODEL_SYNTHESIZE", "deepseek-synthesize")
	t.Setenv("DEEPSEEK_MODEL_EXTRACT", "")
	t.Setenv("DEEPSEEK_BASE_URL", "https://deepseek.test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AIModelProvider != "deepseek" || cfg.DeepSeekAPIKey != "test-key" || cfg.DeepSeekBaseURL != "https://deepseek.test" {
		t.Fatalf("unexpected DeepSeek config: %+v", cfg)
	}
	if cfg.DeepSeekModels[""] != "deepseek-test" || cfg.DeepSeekModels["plan"] != "deepseek-plan" || cfg.DeepSeekModels["synthesize"] != "deepseek-synthesize" || cfg.DeepSeekModels["extract"] != "deepseek-test" {
		t.Fatalf("unexpected DeepSeek models: %+v", cfg.DeepSeekModels)
	}
}

func TestLoadRejectsNonProposalOnlyRuntime(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("AI_RUNTIME_PROPOSAL_ONLY", "false")

	if _, err := Load(); err == nil {
		t.Fatal("expected AI_RUNTIME_PROPOSAL_ONLY=false error")
	}
}

func TestLoadAutonomyThreshold(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("AUTONOMY_THRESHOLD", "0.9")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AutonomyThreshold != 0.9 {
		t.Fatalf("expected 0.9, got %v", cfg.AutonomyThreshold)
	}
}

func TestLoadRejectsInvalidAutonomyThreshold(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("AUTONOMY_THRESHOLD", "1.5")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid AUTONOMY_THRESHOLD error")
	}
}

func TestLoadEmbeddingsConfig(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("EMBEDDINGS_ENABLED", "true")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "sofia-test")
	t.Setenv("EMBEDDINGS_MODEL", "embedding-test")
	t.Setenv("BELIEF_DEDUPE_THRESHOLD", "0.9")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EmbeddingsEnabled || cfg.GoogleCloudProject != "sofia-test" || cfg.EmbeddingsModel != "embedding-test" {
		t.Fatalf("unexpected embeddings config: %+v", cfg)
	}
	if cfg.BeliefDedupeThreshold != 0.9 {
		t.Fatalf("expected threshold 0.9, got %v", cfg.BeliefDedupeThreshold)
	}
}

func TestLoadEmbeddingsRequireProjectAndModel(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("EMBEDDINGS_ENABLED", "true")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("FIRESTORE_PROJECT_ID", "")
	t.Setenv("EMBEDDINGS_MODEL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing embeddings configuration error")
	}
}

func TestLoadRejectsInvalidBeliefDedupeThreshold(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("BELIEF_DEDUPE_THRESHOLD", "1.5")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid BELIEF_DEDUPE_THRESHOLD error")
	}
}

func TestLoadIngestionTokenBudget(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("INGESTION_MAX_TOKENS_PER_BATCH", "1234")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IngestionMaxTokensPerBatch != 1234 {
		t.Fatalf("got %d, want 1234", cfg.IngestionMaxTokensPerBatch)
	}
}
