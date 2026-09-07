package worker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	notificationsdomain "github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	notificationsinfra "github.com/armandoalvarado/sofia-backend/internal/notifications/infrastructure"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
)

type recordingFCMSender struct {
	called   int
	messages []FCMMessage
}

func (s *recordingFCMSender) Send(ctx context.Context, projectID string, message FCMMessage, dryRun bool) error {
	s.called++
	s.messages = append(s.messages, message)
	return nil
}

func TestFCMDryRunUsesInjectedSenderAndMinimalPayload(t *testing.T) {
	devices := notificationsinfra.NewInMemoryDeviceTokenRepository()
	token, err := notificationsdomain.NewDeviceToken("device-1", "user-1", notificationsdomain.DeviceTokenCreate{Platform: notificationsdomain.PlatformIOS, Token: "abcdef1234567890"})
	if err != nil {
		t.Fatal(err)
	}
	if err := devices.Save(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	sender := &recordingFCMSender{}
	provider := NewFCMDeliveryProvider(devices, FCMDeliveryOptions{ProjectID: "project", Enabled: true, DryRun: true, Sender: sender})

	reminder := mustFCMReminder(t, strings.Repeat("Sensitive title ", 20))
	if err := provider.Deliver(context.Background(), reminder); err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}

	if sender.called != 1 {
		t.Fatalf("expected one sender call, got %d", sender.called)
	}
	message := sender.messages[0]
	if message.Type != "reminder_due" || message.ReminderID != reminder.ID {
		t.Fatalf("unexpected message: %+v", message)
	}
	if len([]rune(message.Title)) > 80 {
		t.Fatalf("expected title truncated, got %q", message.Title)
	}
	if strings.Contains(message.TokenRedacted, "abcdef1234567890") {
		t.Fatalf("token was not redacted: %q", message.TokenRedacted)
	}
}

func TestFCMDisabledIsSafeNoop(t *testing.T) {
	devices := notificationsinfra.NewInMemoryDeviceTokenRepository()
	sender := &recordingFCMSender{}
	provider := NewFCMDeliveryProvider(devices, FCMDeliveryOptions{ProjectID: "project", Enabled: false, DryRun: true, Sender: sender})

	if err := provider.Deliver(context.Background(), mustFCMReminder(t, "Reminder")); err != nil {
		t.Fatalf("expected disabled FCM to be safe noop, got %v", err)
	}
	if sender.called != 0 {
		t.Fatalf("expected sender not called, got %d", sender.called)
	}
}

func TestRedactTokenDoesNotExposeFullToken(t *testing.T) {
	token := "abcdef1234567890"
	redacted := RedactToken(token)
	if strings.Contains(redacted, token) || redacted == token {
		t.Fatalf("token leaked: %q", redacted)
	}
	if !strings.Contains(redacted, "...") {
		t.Fatalf("expected compact redaction, got %q", redacted)
	}
}

func TestClassifyFCMHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       string
	}{
		{name: "auth", statusCode: 403, body: "permission denied", want: FCMErrorAuthError},
		{name: "invalid token", statusCode: 400, body: "invalid registration token", want: FCMErrorInvalidToken},
		{name: "quota", statusCode: 429, body: "quota exceeded", want: FCMErrorQuota},
		{name: "network fallback", statusCode: 500, body: "internal", want: FCMErrorNetwork},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyFCMHTTPError(tt.statusCode, tt.body); got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestSanitizeFCMErrorRedactsToken(t *testing.T) {
	token := "abcdef1234567890"
	body := sanitizeFCMError("invalid token abcdef1234567890", token)
	if strings.Contains(body, token) {
		t.Fatalf("token leaked: %s", body)
	}
}

func TestFCMErrorCanBeInspected(t *testing.T) {
	err := &FCMError{Type: FCMErrorInvalidToken, StatusCode: 400, BodyRedacted: "invalid token"}
	var fcmErr *FCMError
	if !errors.As(err, &fcmErr) {
		t.Fatalf("expected FCMError")
	}
	if fcmErr.Type != FCMErrorInvalidToken {
		t.Fatalf("unexpected type: %+v", fcmErr)
	}
}

func mustFCMReminder(t *testing.T, title string) *remindersdomain.Reminder {
	t.Helper()
	reminder, err := remindersdomain.NewReminder("reminder-1", "user-1", remindersdomain.ReminderCreate{
		Title:       title,
		ScheduledAt: time.Now().Add(-time.Minute),
		Timezone:    "UTC",
	})
	if err != nil {
		t.Fatalf("NewReminder returned error: %v", err)
	}
	return reminder
}
