package app

// Worker/provider composition currently lives in cmd/workers/reminders because
// reminder workers are separate processes. This file marks the app composition
// boundary for future shared worker builders without introducing a DI framework.
