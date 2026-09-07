package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/config"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	notificationsdomain "github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	notificationsinfra "github.com/armandoalvarado/sofia-backend/internal/notifications/infrastructure"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/worker"
	"github.com/google/uuid"
)

type deviceTokenRepository interface {
	notificationsdomain.DeviceTokenRepository
}

type testParams struct {
	UserID        string
	DeviceTokenID string
	OutputDir     string
	DryRun        bool
	ProjectID     string
	Title         string
}

type testResult struct {
	UserID              string
	DeviceTokenID       string
	TokenRedacted       string
	Provider            string
	Status              string
	DryRun              bool
	ErrorType           string
	AttemptID           string
	ReminderID          string
	ProjectIDValid      bool
	TokenExists         bool
	VisibleNotification bool
	CreatedAt           time.Time
}

func main() {
	userID := flag.String("user-id", "", "user ID whose enabled device token should receive the test payload")
	deviceTokenID := flag.String("device-token-id", "", "specific device token ID to test")
	outputDir := flag.String("output-dir", "docs", "directory for fcm_test_report.md")
	dryRun := flag.Bool("dry-run", true, "use FCM dry-run path; true by default")
	title := flag.String("title", "Sofia FCM test", "minimal notification title")
	flag.Parse()

	if err := run(context.Background(), testParams{
		UserID:        *userID,
		DeviceTokenID: *deviceTokenID,
		OutputDir:     *outputDir,
		DryRun:        *dryRun,
		Title:         *title,
	}); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, params testParams) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	projectID := strings.TrimSpace(params.ProjectID)
	if projectID == "" {
		projectID = strings.TrimSpace(cfg.FCMProjectID)
	}

	devices, attempts, closeFn, err := buildRepositories(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeFn()

	result, err := executeTest(ctx, params, projectID, devices, attempts, worker.NewHTTPFCMSender())
	if reportErr := writeReport(filepath.Join(params.OutputDir, "fcm_test_report.md"), result, err); reportErr != nil {
		return reportErr
	}
	return err
}

func buildRepositories(ctx context.Context, cfg config.Config) (deviceTokenRepository, notificationsdomain.DeliveryAttemptRepository, func(), error) {
	switch cfg.PersistenceDriver {
	case "memory":
		return notificationsinfra.NewInMemoryDeviceTokenRepository(), notificationsinfra.NewInMemoryDeliveryAttemptRepository(), func() {}, nil
	case "firestore":
		store, err := database.NewFirestore(ctx, cfg)
		if err != nil {
			return nil, nil, nil, err
		}
		return notificationsinfra.NewFirestoreDeviceTokenRepository(store.Client), notificationsinfra.NewFirestoreDeliveryAttemptRepository(store.Client), func() {
			if err := store.Close(); err != nil {
				log.Printf("close firestore: %v", err)
			}
		}, nil
	default:
		return nil, nil, nil, errors.New("unsupported persistence driver")
	}
}

func executeTest(ctx context.Context, params testParams, projectID string, devices deviceTokenRepository, attempts notificationsdomain.DeliveryAttemptRepository, sender worker.FCMSender) (testResult, error) {
	if strings.TrimSpace(params.UserID) == "" && strings.TrimSpace(params.DeviceTokenID) == "" {
		return testResult{Provider: notificationsdomain.DeliveryProviderFCM, Status: notificationsdomain.DeliveryStatusFailed, DryRun: params.DryRun, ErrorType: worker.FCMErrorInvalidToken}, errors.New("user-id or device-token-id is required")
	}
	if strings.TrimSpace(projectID) == "" {
		return testResult{Provider: notificationsdomain.DeliveryProviderFCM, Status: notificationsdomain.DeliveryStatusFailed, DryRun: params.DryRun, ErrorType: worker.FCMErrorAuthError}, worker.ErrFCMProjectID
	}
	tokens, err := selectTokens(ctx, devices, params.UserID, params.DeviceTokenID)
	if err != nil {
		return testResult{Provider: notificationsdomain.DeliveryProviderFCM, Status: notificationsdomain.DeliveryStatusFailed, DryRun: params.DryRun, ProjectIDValid: true, ErrorType: worker.FCMErrorInvalidToken}, err
	}
	if len(tokens) == 0 {
		return testResult{Provider: notificationsdomain.DeliveryProviderFCM, Status: notificationsdomain.DeliveryStatusFailed, DryRun: params.DryRun, ProjectIDValid: true, ErrorType: worker.FCMErrorInvalidToken}, worker.ErrFCMNoDeviceToken
	}
	token := tokens[0]
	reminder := testReminder(token.UserID, strings.TrimSpace(params.Title))
	provider := worker.NewFCMDeliveryProvider(staticDeviceTokens(tokens), worker.FCMDeliveryOptions{
		ProjectID: projectID,
		Enabled:   true,
		DryRun:    params.DryRun,
		Sender:    sender,
	})

	status := notificationsdomain.DeliveryStatusDelivered
	errorType := ""
	err = provider.Deliver(ctx, reminder)
	if err != nil {
		status = notificationsdomain.DeliveryStatusFailed
		errorType = deliveryErrorType(err)
	} else if params.DryRun {
		errorType = "dry_run_success"
	}
	attempt := notificationsdomain.NewDeliveryAttempt(uuid.NewString(), token.UserID, reminder.ID, notificationsdomain.DeliveryProviderFCM, status, params.DryRun, errorType)
	if createErr := attempts.Create(ctx, attempt); createErr != nil && err == nil {
		err = createErr
		status = notificationsdomain.DeliveryStatusFailed
		errorType = "attempt_create"
	}
	return testResult{
		UserID:              token.UserID,
		DeviceTokenID:       token.ID,
		TokenRedacted:       worker.RedactToken(token.Token),
		Provider:            notificationsdomain.DeliveryProviderFCM,
		Status:              status,
		DryRun:              params.DryRun,
		ErrorType:           errorType,
		AttemptID:           attempt.ID,
		ReminderID:          reminder.ID,
		ProjectIDValid:      true,
		TokenExists:         true,
		VisibleNotification: !params.DryRun,
		CreatedAt:           attempt.CreatedAt,
	}, err
}

func selectTokens(ctx context.Context, devices deviceTokenRepository, userID, deviceTokenID string) ([]*notificationsdomain.DeviceToken, error) {
	userID = strings.TrimSpace(userID)
	deviceTokenID = strings.TrimSpace(deviceTokenID)
	if userID == "" && deviceTokenID == "" {
		return nil, errors.New("user-id or device-token-id is required")
	}
	if deviceTokenID != "" {
		token, err := devices.FindByID(ctx, deviceTokenID)
		if err != nil {
			return nil, err
		}
		if userID != "" && token.UserID != userID {
			return nil, notificationsdomain.ErrDeviceTokenNotFound
		}
		if !token.Enabled {
			return nil, worker.ErrFCMNoDeviceToken
		}
		return []*notificationsdomain.DeviceToken{token}, nil
	}
	return devices.ListByUser(ctx, userID)
}

func testReminder(userID, title string) *remindersdomain.Reminder {
	if title == "" {
		title = "Sofia FCM test"
	}
	return &remindersdomain.Reminder{
		ID:          "fcm-test-" + uuid.NewString(),
		UserID:      userID,
		Title:       title,
		Status:      remindersdomain.StatusPending,
		ScheduledAt: time.Now(),
		Timezone:    "UTC",
		NextRunAt:   time.Now(),
	}
}

func writeReport(path string, result testResult, runErr error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# FCM Test Report\n\n")
	b.WriteString(fmt.Sprintf("Generated at: `%s`\n\n", time.Now().UTC().Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("- provider: `%s`\n", result.Provider))
	b.WriteString(fmt.Sprintf("- status: `%s`\n", result.Status))
	b.WriteString(fmt.Sprintf("- dry_run: `%t`\n", result.DryRun))
	b.WriteString(fmt.Sprintf("- fcm_validate_only: `%t`\n", result.DryRun))
	b.WriteString(fmt.Sprintf("- visible_notification: `%t`\n", result.VisibleNotification))
	b.WriteString(fmt.Sprintf("- project_id_valid: `%t`\n", result.ProjectIDValid))
	b.WriteString(fmt.Sprintf("- device_token_exists: `%t`\n", result.TokenExists))
	b.WriteString(fmt.Sprintf("- user_id: `%s`\n", result.UserID))
	b.WriteString(fmt.Sprintf("- device_token_id: `%s`\n", result.DeviceTokenID))
	b.WriteString(fmt.Sprintf("- token_redacted: `%s`\n", result.TokenRedacted))
	b.WriteString(fmt.Sprintf("- reminder_id: `%s`\n", result.ReminderID))
	b.WriteString(fmt.Sprintf("- attempt_id: `%s`\n", result.AttemptID))
	if result.ErrorType != "" {
		b.WriteString(fmt.Sprintf("- error_type: `%s`\n", result.ErrorType))
	}
	if runErr != nil {
		b.WriteString(fmt.Sprintf("- error: `%s`\n", sanitizeReportValue(runErr.Error())))
	}
	b.WriteString("\nFull device tokens are intentionally excluded from this report.\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func deliveryErrorType(err error) string {
	var fcmErr *worker.FCMError
	if errors.As(err, &fcmErr) && fcmErr.Type != "" {
		return fcmErr.Type
	}
	switch {
	case errors.Is(err, worker.ErrFCMProjectID):
		return worker.FCMErrorAuthError
	case errors.Is(err, worker.ErrFCMNoDeviceToken):
		return worker.FCMErrorInvalidToken
	default:
		return worker.FCMErrorNetwork
	}
}

func sanitizeReportValue(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\n", " ")
	if len([]rune(value)) > 200 {
		return string([]rune(value)[:200]) + "...[TRUNCATED]"
	}
	return value
}

type staticDeviceTokens []*notificationsdomain.DeviceToken

func (s staticDeviceTokens) ListByUser(ctx context.Context, userID string) ([]*notificationsdomain.DeviceToken, error) {
	result := make([]*notificationsdomain.DeviceToken, 0, len(s))
	for _, token := range s {
		if token != nil && token.UserID == userID && token.Enabled {
			result = append(result, token)
		}
	}
	return result, nil
}
