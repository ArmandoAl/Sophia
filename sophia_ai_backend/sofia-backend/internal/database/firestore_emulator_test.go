package database_test

import (
	"context"
	"testing"
	"time"

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	activitiesinfra "github.com/armandoalvarado/sofia-backend/internal/activities/infrastructure"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	memoryinfra "github.com/armandoalvarado/sofia-backend/internal/memory/infrastructure"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	remindersinfra "github.com/armandoalvarado/sofia-backend/internal/reminders/infrastructure"
)

func TestFirestoreEmulatorRepositories(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("PERSISTENCE_DRIVER", "firestore")
	if testing.Short() {
		t.Skip("skipping Firestore emulator test in short mode")
	}
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Firestore emulator config not available: %v", err)
	}
	if cfg.FirestoreEmulatorHost == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST is required for emulator tests")
	}

	ctx := context.Background()
	store, err := database.NewFirestore(ctx, cfg)
	if err != nil {
		t.Fatalf("NewFirestore: %v", err)
	}
	defer store.Close()

	activities := activitiesinfra.NewFirestoreActivityRepository(store.Client)
	activity, err := activitiesdomain.NewActivity("emulator-activity", "emulator-user", activitiesdomain.ActivityCreate{Title: "Emulator activity", Timezone: "UTC"})
	if err != nil {
		t.Fatal(err)
	}
	if err := activities.Create(ctx, activity); err != nil {
		t.Fatalf("create activity: %v", err)
	}
	listedActivities, err := activities.List(ctx, activitiesdomain.ListFilter{UserID: "emulator-user", Limit: 10})
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(listedActivities) == 0 {
		t.Fatal("expected activity from emulator")
	}

	reminders := remindersinfra.NewFirestoreReminderRepository(store.Client)
	reminder, err := remindersdomain.NewReminder("emulator-reminder", "emulator-user", remindersdomain.ReminderCreate{Title: "Emulator reminder", ScheduledAt: time.Now().Add(time.Hour), Timezone: "UTC"})
	if err != nil {
		t.Fatal(err)
	}
	if err := reminders.Create(ctx, reminder); err != nil {
		t.Fatalf("create reminder: %v", err)
	}

	memories := memoryinfra.NewFirestoreMemoryRepository(store.Client)
	memory, err := memorydomain.NewMemory("emulator-memory", "emulator-user", memorydomain.MemoryCreate{Title: "Emulator memory", Content: "safe test content"})
	if err != nil {
		t.Fatal(err)
	}
	if err := memories.Create(ctx, memory); err != nil {
		t.Fatalf("create memory: %v", err)
	}
	if _, err := memories.DeleteSoft(ctx, "emulator-user", "emulator-memory"); err != nil {
		t.Fatalf("delete memory: %v", err)
	}
	found, err := memories.SearchBasic(ctx, memorydomain.SearchFilter{UserID: "emulator-user", Query: "safe", Limit: 10})
	if err != nil {
		t.Fatalf("search memory: %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("deleted memory appeared in search: %+v", found)
	}
}
