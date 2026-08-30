package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	notificationsdomain "github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

var (
	ErrFCMDisabled      = errors.New("fcm delivery is disabled")
	ErrFCMProjectID     = errors.New("fcm project id is required")
	ErrFCMNoDeviceToken = errors.New("no enabled device tokens for reminder user")
)

const (
	FCMErrorAuthError    = "auth_error"
	FCMErrorInvalidToken = "invalid_token"
	FCMErrorQuota        = "quota"
	FCMErrorNetwork      = "network"
)

type FCMError struct {
	Type         string
	StatusCode   int
	BodyRedacted string
}

func (e *FCMError) Error() string {
	if e == nil {
		return "fcm error"
	}
	if e.BodyRedacted == "" {
		return fmt.Sprintf("fcm %s", e.Type)
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("fcm %s: status=%d body=%s", e.Type, e.StatusCode, e.BodyRedacted)
	}
	return fmt.Sprintf("fcm %s: %s", e.Type, e.BodyRedacted)
}

type DeviceTokenLister interface {
	ListByUser(ctx context.Context, userID string) ([]*notificationsdomain.DeviceToken, error)
}

type FCMSender interface {
	Send(ctx context.Context, projectID string, message FCMMessage, dryRun bool) error
}

type FCMDeliveryProvider struct {
	devices   DeviceTokenLister
	sender    FCMSender
	projectID string
	enabled   bool
	dryRun    bool
}

type FCMDeliveryOptions struct {
	ProjectID string
	Enabled   bool
	DryRun    bool
	Sender    FCMSender
}

type FCMMessage struct {
	Token         string
	ReminderID    string
	Title         string
	Type          string
	TokenRedacted string
}

func NewFCMDeliveryProvider(devices DeviceTokenLister, options FCMDeliveryOptions) *FCMDeliveryProvider {
	sender := options.Sender
	if sender == nil {
		sender = NewHTTPFCMSender()
	}
	return &FCMDeliveryProvider{devices: devices, sender: sender, projectID: strings.TrimSpace(options.ProjectID), enabled: options.Enabled, dryRun: options.DryRun}
}

func (p *FCMDeliveryProvider) Deliver(ctx context.Context, reminder *remindersdomain.Reminder) error {
	if !p.enabled {
		return nil
	}
	if strings.TrimSpace(p.projectID) == "" {
		return ErrFCMProjectID
	}
	tokens, err := p.devices.ListByUser(ctx, reminder.UserID)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return ErrFCMNoDeviceToken
	}
	for _, token := range tokens {
		if token == nil || !token.Enabled || strings.TrimSpace(token.Token) == "" {
			continue
		}
		message := FCMMessage{
			Token:         token.Token,
			ReminderID:    reminder.ID,
			Title:         truncateNotificationTitle(reminder.Title),
			Type:          "reminder_due",
			TokenRedacted: RedactToken(token.Token),
		}
		if err := p.sender.Send(ctx, p.projectID, message, p.dryRun); err != nil {
			return err
		}
	}
	return nil
}

type HTTPFCMSender struct {
	client      *http.Client
	tokenSource oauth2.TokenSource
}

func NewHTTPFCMSender() *HTTPFCMSender {
	return &HTTPFCMSender{client: &http.Client{Timeout: 10 * time.Second}}
}

func (s *HTTPFCMSender) Send(ctx context.Context, projectID string, message FCMMessage, dryRun bool) error {
	tokenSource, err := s.source(ctx)
	if err != nil {
		return &FCMError{Type: FCMErrorAuthError, BodyRedacted: sanitizeFCMError(err.Error(), message.Token)}
	}
	token, err := tokenSource.Token()
	if err != nil {
		return &FCMError{Type: FCMErrorAuthError, BodyRedacted: sanitizeFCMError(err.Error(), message.Token)}
	}
	body, err := json.Marshal(map[string]any{
		"validate_only": dryRun,
		"message": map[string]any{
			"token": message.Token,
			"notification": map[string]any{
				"title": message.Title,
			},
			"data": map[string]string{
				"reminder_id": message.ReminderID,
				"type":        message.Type,
			},
		},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return &FCMError{Type: FCMErrorNetwork, BodyRedacted: sanitizeFCMError(err.Error(), message.Token)}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		redacted := sanitizeFCMError(string(payload), message.Token)
		return &FCMError{Type: classifyFCMHTTPError(resp.StatusCode, redacted), StatusCode: resp.StatusCode, BodyRedacted: redacted}
	}
	return nil
}

func (s *HTTPFCMSender) source(ctx context.Context) (oauth2.TokenSource, error) {
	if s.tokenSource != nil {
		return s.tokenSource, nil
	}
	source, err := google.DefaultTokenSource(ctx, fcmScope)
	if err != nil {
		return nil, err
	}
	s.tokenSource = source
	return source, nil
}

func truncateNotificationTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Reminder due"
	}
	runes := []rune(title)
	if len(runes) <= 80 {
		return title
	}
	return string(runes[:80])
}

func RedactToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 8 {
		return "[REDACTED]"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func classifyFCMHTTPError(statusCode int, body string) string {
	lower := strings.ToLower(body)
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return FCMErrorAuthError
	case http.StatusTooManyRequests:
		return FCMErrorQuota
	case http.StatusBadRequest, http.StatusNotFound:
		if strings.Contains(lower, "token") || strings.Contains(lower, "registration") || strings.Contains(lower, "invalid") || strings.Contains(lower, "not found") {
			return FCMErrorInvalidToken
		}
	}
	if strings.Contains(lower, "quota") || strings.Contains(lower, "rate") {
		return FCMErrorQuota
	}
	if strings.Contains(lower, "auth") || strings.Contains(lower, "permission") || strings.Contains(lower, "credential") {
		return FCMErrorAuthError
	}
	if strings.Contains(lower, "token") || strings.Contains(lower, "registration") || strings.Contains(lower, "invalid") {
		return FCMErrorInvalidToken
	}
	return FCMErrorNetwork
}

func sanitizeFCMError(value, token string) string {
	value = strings.TrimSpace(value)
	if token = strings.TrimSpace(token); token != "" {
		value = strings.ReplaceAll(value, token, "[REDACTED_TOKEN]")
	}
	if len([]rune(value)) > 500 {
		return string([]rune(value)[:500]) + "...[TRUNCATED]"
	}
	return value
}
