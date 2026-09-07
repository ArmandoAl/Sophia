package domain

import (
	"testing"
	"time"
)

func TestDailyRecurrenceAddsOneDay(t *testing.T) {
	reminder := mustRecurringReminder(t, RecurrenceDaily, 1, 0, nil)
	if err := reminder.MarkDelivered(); err != nil {
		t.Fatalf("MarkDelivered returned error: %v", err)
	}
	expected := reminder.ScheduledAt.AddDate(0, 0, 1)
	if !reminder.NextRunAt.Equal(expected) || reminder.Status != StatusPending || reminder.RecurrenceOccurrences != 1 {
		t.Fatalf("unexpected daily recurrence: next=%s status=%s occurrences=%d", reminder.NextRunAt, reminder.Status, reminder.RecurrenceOccurrences)
	}
}

func TestWeeklyRecurrenceAddsOneWeek(t *testing.T) {
	reminder := mustRecurringReminder(t, RecurrenceWeekly, 1, 0, nil)
	if err := reminder.MarkDelivered(); err != nil {
		t.Fatalf("MarkDelivered returned error: %v", err)
	}
	expected := reminder.ScheduledAt.AddDate(0, 0, 7)
	if !reminder.NextRunAt.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, reminder.NextRunAt)
	}
}

func TestMonthlyRecurrenceAddsOneMonth(t *testing.T) {
	reminder := mustRecurringReminder(t, RecurrenceMonthly, 1, 0, nil)
	if err := reminder.MarkDelivered(); err != nil {
		t.Fatalf("MarkDelivered returned error: %v", err)
	}
	expected := reminder.ScheduledAt.AddDate(0, 1, 0)
	if !reminder.NextRunAt.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, reminder.NextRunAt)
	}
}

func TestRecurrenceCountStopsRecurrence(t *testing.T) {
	reminder := mustRecurringReminder(t, RecurrenceDaily, 1, 1, nil)
	if err := reminder.MarkDelivered(); err != nil {
		t.Fatalf("MarkDelivered returned error: %v", err)
	}
	if reminder.Status != StatusDelivered || reminder.DeliveredAt == nil {
		t.Fatalf("expected delivered after count reached, got status=%s delivered_at=%v", reminder.Status, reminder.DeliveredAt)
	}
}

func TestRecurrenceUntilStopsRecurrence(t *testing.T) {
	until := time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC)
	reminder := mustRecurringReminder(t, RecurrenceDaily, 1, 0, &until)
	if err := reminder.MarkDelivered(); err != nil {
		t.Fatalf("MarkDelivered returned error: %v", err)
	}
	if reminder.Status != StatusDelivered {
		t.Fatalf("expected delivered when next occurrence passes until, got %s", reminder.Status)
	}
}

func TestNonRecurringReminderStaysDelivered(t *testing.T) {
	reminder := mustRecurringReminder(t, RecurrenceNone, 0, 0, nil)
	if err := reminder.MarkDelivered(); err != nil {
		t.Fatalf("MarkDelivered returned error: %v", err)
	}
	if reminder.Status != StatusDelivered || reminder.DeliveredAt == nil {
		t.Fatalf("expected delivered non-recurring reminder, got %+v", reminder)
	}
}

func TestInvalidRecurrenceValidation(t *testing.T) {
	if _, err := NewReminder("reminder-1", "user-1", ReminderCreate{
		Title:              "Invalid",
		ScheduledAt:        time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC),
		Timezone:           "UTC",
		RecurrenceRule:     RecurrenceDaily,
		RecurrenceInterval: -1,
	}); err != ErrInvalidInterval {
		t.Fatalf("expected ErrInvalidInterval, got %v", err)
	}
	if _, err := NewReminder("reminder-1", "user-1", ReminderCreate{
		Title:           "Invalid",
		ScheduledAt:     time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC),
		Timezone:        "UTC",
		RecurrenceRule:  RecurrenceDaily,
		RecurrenceCount: -1,
	}); err != ErrInvalidCount {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
}

func mustRecurringReminder(t *testing.T, rule string, interval, count int, until *time.Time) *Reminder {
	t.Helper()
	reminder, err := NewReminder("reminder-1", "user-1", ReminderCreate{
		Title:              "Recurring reminder",
		ScheduledAt:        time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC),
		Timezone:           "UTC",
		RecurrenceRule:     rule,
		RecurrenceInterval: interval,
		RecurrenceUntil:    until,
		RecurrenceCount:    count,
	})
	if err != nil {
		t.Fatalf("NewReminder returned error: %v", err)
	}
	return reminder
}
