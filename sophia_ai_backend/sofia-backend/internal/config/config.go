package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultContextTokenBudget = 4600

	defaultEnv               = "development"
	defaultPort              = "8080"
	defaultJWTSecret         = "development-insecure-jwt-secret"
	defaultRequestBodyLimit  = int64(1 << 20) // 1 MiB
	defaultPersistence       = "memory"
	defaultReadTimeout       = 10 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultReadHeaderTimeout = 5 * time.Second
	defaultAIModelProvider   = "fake"
	defaultGeminiModel       = "gemini-1.5-flash"
	defaultDeepSeekModel     = "deepseek-v4-pro"
	defaultJWTAccessTokenTTL = 24 * time.Hour
	defaultAuthRateLimit     = 10
	defaultAuthRateWindow    = time.Minute
	defaultReminderInterval  = 30 * time.Second
	defaultReminderBatchSize = 50
	defaultReminderProvider  = "noop"
	defaultReminderLease     = 2 * time.Minute
	defaultSynthesisInterval = time.Minute
	defaultSynthesisLease    = 2 * time.Minute
	defaultSynthesisRunHour  = 3
	defaultAutonomyThreshold = 0.85
)

type Config struct {
	Env                      string
	Port                     string
	JWTSecret                string
	JWTAccessTokenTTL        time.Duration
	AuthRateLimitRequests    int
	AuthRateLimitWindow      time.Duration
	CORSAllowedOrigins       []string
	RequestBodyLimitBytes    int64
	PersistenceDriver        string
	FirestoreProjectID       string
	FirestoreDatabaseID      string
	FirestoreEmulatorHost    string
	GoogleCredentialsFile    string
	ReadTimeout              time.Duration
	WriteTimeout             time.Duration
	IdleTimeout              time.Duration
	ReadHeaderTimeout        time.Duration
	AIModelProvider          string
	ContextTokenBudget       int
	GeminiAPIKey             string
	GeminiModel              string
	DeepSeekAPIKey           string
	DeepSeekModel            string
	DeepSeekBaseURL          string
	AIRuntimeEnabled         bool
	AIRuntimeProposalOnly    bool
	ReminderWorkerEnabled    bool
	ReminderWorkerID         string
	ReminderWorkerInterval   time.Duration
	ReminderWorkerBatchSize  int
	ReminderWorkerLease      time.Duration
	ReminderDeliveryProvider string
	FCMEnabled               bool
	FCMDryRun                bool
	FCMProjectID             string
	SynthesisWorkerEnabled   bool
	SynthesisWorkerID        string
	SynthesisWorkerInterval  time.Duration
	SynthesisWorkerLease     time.Duration
	SynthesisRunHourLocal    int
	AutonomyThreshold        float64
}

func Load() (Config, error) {
	cfg := Config{
		Env:                      getEnv("ENV", defaultEnv),
		Port:                     getEnv("PORT", defaultPort),
		JWTSecret:                os.Getenv("JWT_SECRET"),
		JWTAccessTokenTTL:        defaultJWTAccessTokenTTL,
		AuthRateLimitRequests:    defaultAuthRateLimit,
		AuthRateLimitWindow:      defaultAuthRateWindow,
		CORSAllowedOrigins:       splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
		RequestBodyLimitBytes:    defaultRequestBodyLimit,
		PersistenceDriver:        strings.ToLower(getEnv("PERSISTENCE_DRIVER", defaultPersistence)),
		FirestoreProjectID:       strings.TrimSpace(os.Getenv("FIRESTORE_PROJECT_ID")),
		FirestoreDatabaseID:      strings.TrimSpace(os.Getenv("FIRESTORE_DATABASE_ID")),
		FirestoreEmulatorHost:    strings.TrimSpace(os.Getenv("FIRESTORE_EMULATOR_HOST")),
		GoogleCredentialsFile:    strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")),
		ReadTimeout:              defaultReadTimeout,
		WriteTimeout:             defaultWriteTimeout,
		IdleTimeout:              defaultIdleTimeout,
		ReadHeaderTimeout:        defaultReadHeaderTimeout,
		AIModelProvider:          strings.ToLower(getEnv("AI_MODEL_PROVIDER", defaultAIModelProvider)),
		ContextTokenBudget:       DefaultContextTokenBudget,
		GeminiAPIKey:             strings.TrimSpace(os.Getenv("GEMINI_API_KEY")),
		GeminiModel:              getEnv("GEMINI_MODEL", defaultGeminiModel),
		DeepSeekAPIKey:           strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		DeepSeekModel:            getEnv("DEEPSEEK_MODEL", defaultDeepSeekModel),
		DeepSeekBaseURL:          strings.TrimSpace(os.Getenv("DEEPSEEK_BASE_URL")),
		AIRuntimeEnabled:         boolEnv("AI_RUNTIME_ENABLED", true),
		AIRuntimeProposalOnly:    boolEnv("AI_RUNTIME_PROPOSAL_ONLY", true),
		ReminderWorkerEnabled:    boolEnv("REMINDER_WORKER_ENABLED", false),
		ReminderWorkerID:         getEnv("REMINDER_WORKER_ID", defaultReminderWorkerID()),
		ReminderWorkerInterval:   defaultReminderInterval,
		ReminderWorkerBatchSize:  defaultReminderBatchSize,
		ReminderWorkerLease:      defaultReminderLease,
		ReminderDeliveryProvider: strings.ToLower(getEnv("REMINDER_DELIVERY_PROVIDER", defaultReminderProvider)),
		FCMEnabled:               boolEnv("FCM_ENABLED", false),
		FCMDryRun:                boolEnv("FCM_DRY_RUN", true),
		FCMProjectID:             strings.TrimSpace(os.Getenv("FCM_PROJECT_ID")),
		SynthesisWorkerEnabled:   boolEnv("SYNTHESIS_WORKER_ENABLED", false),
		SynthesisWorkerID:        getEnv("SYNTHESIS_WORKER_ID", defaultSynthesisWorkerID()),
		SynthesisWorkerInterval:  defaultSynthesisInterval,
		SynthesisWorkerLease:     defaultSynthesisLease,
		SynthesisRunHourLocal:    defaultSynthesisRunHour,
		AutonomyThreshold:        defaultAutonomyThreshold,
	}

	if !isValidEnv(cfg.Env) {
		return Config{}, fmt.Errorf("ENV must be one of development, test, production")
	}

	if cfg.JWTSecret == "" {
		if cfg.Env == "production" {
			return Config{}, errors.New("JWT_SECRET is required in production")
		}
		cfg.JWTSecret = defaultJWTSecret
	}
	if cfg.Env == "production" && cfg.JWTSecret == defaultJWTSecret {
		return Config{}, errors.New("JWT_SECRET must be secure in production")
	}

	if cfg.PersistenceDriver != "memory" && cfg.PersistenceDriver != "firestore" {
		return Config{}, errors.New("PERSISTENCE_DRIVER must be memory or firestore")
	}
	if cfg.Env == "production" && cfg.PersistenceDriver != "firestore" {
		return Config{}, errors.New("PERSISTENCE_DRIVER=firestore is required in production")
	}
	if cfg.PersistenceDriver == "firestore" && cfg.FirestoreProjectID == "" {
		return Config{}, errors.New("FIRESTORE_PROJECT_ID is required when PERSISTENCE_DRIVER=firestore")
	}
	if cfg.AIModelProvider != "fake" && cfg.AIModelProvider != "gemini" && cfg.AIModelProvider != "deepseek" {
		return Config{}, errors.New("AI_MODEL_PROVIDER must be fake, gemini, or deepseek")
	}
	if cfg.AIModelProvider == "gemini" && cfg.GeminiAPIKey == "" {
		return Config{}, errors.New("GEMINI_API_KEY is required when AI_MODEL_PROVIDER=gemini")
	}
	if cfg.AIModelProvider == "gemini" && strings.TrimSpace(cfg.GeminiModel) == "" {
		return Config{}, errors.New("GEMINI_MODEL is required when AI_MODEL_PROVIDER=gemini")
	}
	if cfg.AIModelProvider == "deepseek" && cfg.DeepSeekAPIKey == "" {
		return Config{}, errors.New("DEEPSEEK_API_KEY is required when AI_MODEL_PROVIDER=deepseek")
	}
	if cfg.AIModelProvider == "deepseek" && strings.TrimSpace(cfg.DeepSeekModel) == "" {
		return Config{}, errors.New("DEEPSEEK_MODEL is required when AI_MODEL_PROVIDER=deepseek")
	}
	if !cfg.AIRuntimeProposalOnly {
		return Config{}, errors.New("AI_RUNTIME_PROPOSAL_ONLY=false is not supported yet")
	}
	if cfg.ReminderDeliveryProvider != "noop" && cfg.ReminderDeliveryProvider != "fcm" {
		return Config{}, errors.New("REMINDER_DELIVERY_PROVIDER must be noop or fcm")
	}
	if cfg.ReminderDeliveryProvider == "fcm" && cfg.FCMEnabled && strings.TrimSpace(cfg.FCMProjectID) == "" {
		return Config{}, errors.New("FCM_PROJECT_ID is required when FCM delivery is enabled")
	}

	if rawLimit := os.Getenv("REQUEST_BODY_LIMIT_BYTES"); rawLimit != "" {
		limit, err := strconv.ParseInt(rawLimit, 10, 64)
		if err != nil || limit <= 0 {
			return Config{}, errors.New("REQUEST_BODY_LIMIT_BYTES must be a positive integer")
		}
		cfg.RequestBodyLimitBytes = limit
	}

	var err error
	if cfg.JWTAccessTokenTTL, err = durationEnv("JWT_ACCESS_TOKEN_TTL", cfg.JWTAccessTokenTTL); err != nil {
		return Config{}, err
	}
	if cfg.AuthRateLimitRequests, err = intEnv("AUTH_RATE_LIMIT_REQUESTS", cfg.AuthRateLimitRequests); err != nil {
		return Config{}, err
	}
	if cfg.AuthRateLimitWindow, err = durationEnv("AUTH_RATE_LIMIT_WINDOW", cfg.AuthRateLimitWindow); err != nil {
		return Config{}, err
	}

	if cfg.ReadTimeout, err = durationEnv("HTTP_READ_TIMEOUT", cfg.ReadTimeout); err != nil {
		return Config{}, err
	}
	if cfg.WriteTimeout, err = durationEnv("HTTP_WRITE_TIMEOUT", cfg.WriteTimeout); err != nil {
		return Config{}, err
	}
	if cfg.IdleTimeout, err = durationEnv("HTTP_IDLE_TIMEOUT", cfg.IdleTimeout); err != nil {
		return Config{}, err
	}
	if cfg.ReadHeaderTimeout, err = durationEnv("HTTP_READ_HEADER_TIMEOUT", cfg.ReadHeaderTimeout); err != nil {
		return Config{}, err
	}
	if cfg.ContextTokenBudget, err = intEnv("CONTEXT_TOKEN_BUDGET", cfg.ContextTokenBudget); err != nil {
		return Config{}, err
	}
	if cfg.ReminderWorkerInterval, err = durationEnv("REMINDER_WORKER_INTERVAL", cfg.ReminderWorkerInterval); err != nil {
		return Config{}, err
	}
	if cfg.ReminderWorkerBatchSize, err = intEnv("REMINDER_WORKER_BATCH_SIZE", cfg.ReminderWorkerBatchSize); err != nil {
		return Config{}, err
	}
	if cfg.ReminderWorkerLease, err = durationEnv("REMINDER_WORKER_LEASE_DURATION", cfg.ReminderWorkerLease); err != nil {
		return Config{}, err
	}
	if cfg.SynthesisWorkerInterval, err = durationEnv("SYNTHESIS_WORKER_INTERVAL", cfg.SynthesisWorkerInterval); err != nil {
		return Config{}, err
	}
	if cfg.SynthesisWorkerLease, err = durationEnv("SYNTHESIS_WORKER_LEASE", cfg.SynthesisWorkerLease); err != nil {
		return Config{}, err
	}
	if cfg.SynthesisRunHourLocal, err = hourEnv("SYNTHESIS_RUN_HOUR_LOCAL", cfg.SynthesisRunHourLocal); err != nil {
		return Config{}, err
	}
	if cfg.AutonomyThreshold, err = ratioEnv("AUTONOMY_THRESHOLD", cfg.AutonomyThreshold); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return ":" + c.Port
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func defaultReminderWorkerID() string {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		return "reminder-worker-local"
	}
	return "reminder-worker-" + strings.TrimSpace(hostname)
}

func defaultSynthesisWorkerID() string {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		return "synthesis-worker-local"
	}
	return "synthesis-worker-" + strings.TrimSpace(hostname)
}

func hourEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 || parsed > 23 {
		return 0, fmt.Errorf("%s must be an hour between 0 and 23", key)
	}
	return parsed, nil
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func boolEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func isValidEnv(env string) bool {
	switch env {
	case "development", "test", "production":
		return true
	default:
		return false
	}
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return duration, nil
}

func ratioEnv(key string, fallback float64) (float64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed < 0 || parsed > 1 {
		return 0, fmt.Errorf("%s must be a number between 0 and 1", key)
	}
	return parsed, nil
}

func intEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}
