package main

import (
	"testing"

	"github.com/armandoalvarado/sofia-backend/internal/config"
	notificationsinfra "github.com/armandoalvarado/sofia-backend/internal/notifications/infrastructure"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/worker"
)

func TestBuildDeliveryProviderUsesConfiguredFCM(t *testing.T) {
	provider, err := buildDeliveryProvider(config.Config{
		ReminderDeliveryProvider: "fcm",
		FCMProjectID:             "project",
		FCMEnabled:               true,
		FCMDryRun:                true,
	}, notificationsinfra.NewInMemoryDeviceTokenRepository())
	if err != nil {
		t.Fatalf("buildDeliveryProvider returned error: %v", err)
	}
	if _, ok := provider.(*worker.FCMDeliveryProvider); !ok {
		t.Fatalf("expected FCM provider, got %T", provider)
	}
}
