# Local E2E Report

Generated at: `2026-07-05T19:12:12Z`

## Summary

- user_id: `49c4028f-d047-425d-80ae-646f17425b50`
- activity_id: `409c29e0-b1a9-43e2-9da2-246e39385f29`
- reminder_id: `a272ad31-8bc5-4b38-8aaa-90d5e24c507f`
- memory_id: `d5f87516-af59-4a42-aaaf-4213954e9b17`
- proposal_id: `186095a0-eda2-4fe5-a8ca-f78b5584ce75`
- executed_proposal_status: `executed`
- proposal_only_preserved: `true`
- no_automatic_ai_execution: `true`
- explicit_execution_created_activity: `true`

## Export Counts

- activities: `2`
- reminders: `1`
- memories: `1`
- action_proposals: `1`

## AI Runtime Safety

- dry_run_actions_returned: `1`
- proposals_before_dry_run: `0`
- proposals_after_dry_run: `0`

## Steps

| Step | Status | Detail |
| --- | --- | --- |
| register user | pass | auth user created in memory |
| create profile/settings | pass | default user profile and AI settings created |
| configure AI settings | pass | memory/reminders/planning enabled with manual autonomy |
| create activity | pass | 409c29e0-b1a9-43e2-9da2-246e39385f29 |
| create recurring reminder | pass | a272ad31-8bc5-4b38-8aaa-90d5e24c507f |
| create memory | pass | d5f87516-af59-4a42-aaaf-4213954e9b17 |
| call AI runtime fake dry_run | pass | actions=1 persisted_delta=0 |
| validate no automatic AI execution | pass | dry_run did not persist or execute proposals |
| validate proposal-only | pass | runtime stayed proposal-only |
| create proposal | pass | 186095a0-eda2-4fe5-a8ca-f78b5584ce75 |
| confirm proposal | pass | confirmed |
| execute proposal explicitly | pass | created one activity after explicit execute |
| export user data | pass | activities=2 reminders=1 memories=1 proposals=1 |
