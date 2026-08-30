# Reminder Worker Operational Report

Generated at: `2026-07-04T06:21:22Z`

## Scope

This command runs deterministic synthetic scenarios against the reminder worker lease logic. It does not send notifications and does not require production Firestore credentials.

## Results

| Scenario | Status | Detail |
| --- | --- | --- |
| active lease blocks second worker | pass | worker-b saw the due candidate but ClaimDue rejected the active lease; only worker-a delivered |
| expired lease allows retry | pass | active lease rejected retry before expiry; expired lease was claimed by worker-b |
| recurring reminder moves next_run_at | pass | next_run_at advanced to 2026-01-03T14:59:00Z and status remained pending |
| failed reminder does not duplicate | pass | failed reminder moved out of pending scan after one provider failure |

## Notes

- The simulation uses the same worker and repository lease contract used by runtime code.
- Firestore production correctness still depends on the transactional `ClaimDue` implementation and emulator coverage.
- No visible notifications are sent.
