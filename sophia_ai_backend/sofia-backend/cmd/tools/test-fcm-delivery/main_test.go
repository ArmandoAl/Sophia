package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	notificationsdomain "github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	notificationsinfra "github.com/armandoalvarado/sofia-backend/internal/notifications/infrastructure"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/worker"
)

type recordingSender struct {
	called bool
	dryRun bool
}

func (s *recordingSender) Send(ctx context.Context, projectID string, message worker.FCMMessage, dryRun bool) error {
	s.called = true
	s.dryRun = dryRun
	return nil
}

func TestExecuteTestDryRunDoesNotRequireRealNetworkAndRecordsAttempt(t *testing.T) {
	ctx := context.Background()
	devices := notificationsinfra.NewInMemoryDeviceTokenRepository()
	attempts := notificationsinfra.NewInMemoryDeliveryAttemptRepository()
	token := mustDeviceToken(t, "device-1", "user-1", "abcdef1234567890")
	if err := devices.Save(ctx, token); err != nil {
		t.Fatal(err)
	}
	sender := &recordingSender{}

	result, err := executeTest(ctx, testParams{UserID: "user-1", DryRun: true, Title: "Test"}, "project", devices, attempts, sender)
	if err != nil {
		t.Fatalf("executeTest returned error: %v", err)
	}
	if !sender.called || !sender.dryRun || result.VisibleNotification || !result.ProjectIDValid || !result.TokenExists {
		t.Fatalf("expected dry-run sender path, sender=%+v result=%+v", sender, result)
	}
	if result.ErrorType != "dry_run_success" {
		t.Fatalf("expected dry_run_success, got %+v", result)
	}
	if result.TokenRedacted == token.Token || strings.Contains(result.TokenRedacted, token.Token) {
		t.Fatalf("token leaked in result: %+v", result)
	}
	storedAttempts := attempts.ListAll()
	if len(storedAttempts) != 1 {
		t.Fatalf("expected one attempt, got %+v", storedAttempts)
	}
	if storedAttempts[0].Status != notificationsdomain.DeliveryStatusDelivered || !storedAttempts[0].DryRun {
		t.Fatalf("unexpected attempt: %+v", storedAttempts[0])
	}
}

func TestSelectTokenRejectsCrossUserDeviceToken(t *testing.T) {
	ctx := context.Background()
	devices := notificationsinfra.NewInMemoryDeviceTokenRepository()
	if err := devices.Save(ctx, mustDeviceToken(t, "device-1", "user-1", "abcdef1234567890")); err != nil {
		t.Fatal(err)
	}

	if _, err := selectTokens(ctx, devices, "user-2", "device-1"); err == nil {
		t.Fatal("expected cross-user token lookup to fail")
	}
}

func TestWriteReportRedactsToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fcm_test_report.md")
	fullToken := "abcdef1234567890"
	result := testResult{
		UserID:         "user-1",
		DeviceTokenID:  "device-1",
		TokenRedacted:  worker.RedactToken(fullToken),
		Provider:       notificationsdomain.DeliveryProviderFCM,
		Status:         notificationsdomain.DeliveryStatusDelivered,
		DryRun:         true,
		ProjectIDValid: true,
		TokenExists:    true,
		AttemptID:      "attempt-1",
		ReminderID:     "reminder-1",
	}
	if err := writeReport(path, result, nil); err != nil {
		t.Fatalf("writeReport returned error: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), fullToken) {
		t.Fatalf("report leaked full token:\n%s", raw)
	}
	if !strings.Contains(string(raw), "abcd...7890") {
		t.Fatalf("report missing redacted token:\n%s", raw)
	}
	if !strings.Contains(string(raw), "visible_notification: `false`") || !strings.Contains(string(raw), "fcm_validate_only: `true`") {
		t.Fatalf("report missing dry-run detail:\n%s", raw)
	}
}

func TestExecuteTestRequiresProjectID(t *testing.T) {
	ctx := context.Background()
	devices := notificationsinfra.NewInMemoryDeviceTokenRepository()
	attempts := notificationsinfra.NewInMemoryDeliveryAttemptRepository()
	if err := devices.Save(ctx, mustDeviceToken(t, "device-1", "user-1", "abcdef1234567890")); err != nil {
		t.Fatal(err)
	}

	result, err := executeTest(ctx, testParams{UserID: "user-1", DryRun: true}, "", devices, attempts, &recordingSender{})
	if err == nil {
		t.Fatal("expected missing project ID error")
	}
	if result.ErrorType != worker.FCMErrorAuthError {
		t.Fatalf("expected auth_error, got %+v", result)
	}
}

func mustDeviceToken(t *testing.T, id, userID, rawToken string) *notificationsdomain.DeviceToken {
	t.Helper()
	token, err := notificationsdomain.NewDeviceToken(id, userID, notificationsdomain.DeviceTokenCreate{Platform: notificationsdomain.PlatformIOS, Token: rawToken})
	if err != nil {
		t.Fatalf("NewDeviceToken returned error: %v", err)
	}
	return token
}
