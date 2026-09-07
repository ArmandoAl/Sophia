# Database Schema Snapshot

Generated at: `2026-07-01T06:42:32Z`

Project ID: `sophia-ai-60e87`

## `activities`

- Purpose inferred: planning units: tasks, habits, routines, goals, events and checklists
- Sampled documents: `1`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `ai_action_ref` | `*firestore.DocumentRef` | true | false | false |
| `confirmed` | `bool` | true | false | false |
| `description` | `string` | true | false | false |
| `end_time` | `timestamp` | true | false | false |
| `flexibility` | `number` | true | false | false |
| `origin_message_ref` | `*firestore.DocumentRef` | true | false | false |
| `priority` | `number` | true | false | false |
| `requires_confirmation` | `bool` | true | false | false |
| `reschedulable` | `bool` | true | false | false |
| `start_time` | `timestamp` | true | false | false |
| `title` | `string` | true | false | false |
| `type` | `string` | true | false | false |
| `user_id` | `*firestore.DocumentRef` | true | false | false |

Recommended indexes: `user_id+status`, `user_id+type`, `user_id+priority`, range filters by date

Differences vs expected model:

- Missing fields: `created_by`, `id`, `source`, `status`, `tags`, `timezone`
- Unexpected fields: `_document_id`, `ai_action_ref`, `confirmed`, `description`, `end_time`, `flexibility`, `origin_message_ref`, `requires_confirmation`, `reschedulable`, `start_time`

## `ai_actions_log`

- Purpose inferred: unknown collection; review ownership before relying on it
- Sampled documents: `1`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `action_type` | `string` | true | false | false |
| `approved` | `bool` | true | false | false |
| `body_id` | `*firestore.DocumentRef` | true | false | false |
| `confidence` | `number` | true | false | false |
| `created_at` | `timestamp` | true | false | false |
| `executed` | `bool` | true | false | false |
| `reasoning` | `string` | true | false | false |
| `requires_confirmation` | `bool` | true | false | false |
| `target_refs` | `array` | true | false | false |
| `user_id` | `*firestore.DocumentRef` | true | false | false |

Recommended indexes: review query patterns before creating indexes

## `ai_memory`

- Purpose inferred: unknown collection; review ownership before relying on it
- Sampled documents: `1`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `body_id` | `*firestore.DocumentRef` | true | false | false |
| `confidence` | `number` | true | false | false |
| `enabled` | `bool` | true | false | false |
| `last_seen` | `timestamp` | true | false | false |
| `reinforcement_count` | `number` | true | false | false |
| `source_message_ref` | `*firestore.DocumentRef` | true | false | false |
| `statement` | `string` | true | false | false |
| `type` | `string` | true | false | false |
| `user_id` | `*firestore.DocumentRef` | true | false | false |

Recommended indexes: review query patterns before creating indexes

## `ai_policy`

- Purpose inferred: unknown collection; review ownership before relying on it
- Sampled documents: `1`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `action_type` | `string` | true | false | false |
| `confidence_threshold` | `number` | true | false | false |
| `forbidden_without_confirmation` | `bool` | true | false | false |
| `requires_confirmation` | `bool` | true | false | false |

Recommended indexes: review query patterns before creating indexes

## `bodies`

- Purpose inferred: unknown collection; review ownership before relying on it
- Sampled documents: `1`
- Subcollections detected: `extensions`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `description` | `string` | true | false | false |
| `is_active` | `bool` | true | false | false |
| `location` | `*latlng.LatLng` | true | false | false |
| `name` | `string` | true | false | false |
| `user_id` | `*firestore.DocumentRef` | true | false | false |

Recommended indexes: review query patterns before creating indexes

## `conversations`

- Purpose inferred: unknown collection; review ownership before relying on it
- Sampled documents: `1`
- Subcollections detected: `messages`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `body_id` | `*firestore.DocumentRef` | true | false | false |
| `session_id` | `string` | true | false | false |
| `timestamp` | `timestamp` | true | false | false |
| `user_id` | `*firestore.DocumentRef` | true | false | false |

Recommended indexes: review query patterns before creating indexes

## `daily_summaries`

- Purpose inferred: unknown collection; review ownership before relying on it
- Sampled documents: `1`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `accepted_suggestions` | `number` | true | false | false |
| `date` | `timestamp` | true | false | false |
| `decisions` | `array` | true | false | false |
| `focus_index` | `number` | true | false | false |
| `memory_written_refs` | `array` | true | false | false |
| `stress_index` | `number` | true | false | false |
| `topics` | `array` | true | false | false |
| `user_id` | `*firestore.DocumentRef` | true | false | false |

Recommended indexes: review query patterns before creating indexes

## `locations`

- Purpose inferred: unknown collection; review ownership before relying on it
- Sampled documents: `1`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `coordinates` | `*latlng.LatLng` | true | false | false |
| `description` | `string` | true | false | false |
| `geo_hash` | `string` | true | false | false |
| `name` | `string` | true | false | false |
| `privacy_level` | `string` | true | false | false |
| `type` | `string` | true | false | false |

Recommended indexes: review query patterns before creating indexes

## `message_embeddings`

- Purpose inferred: unknown collection; review ownership before relying on it
- Sampled documents: `1`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `message_ref` | `*firestore.DocumentRef` | true | false | false |
| `vector` | `array` | true | false | false |

Recommended indexes: review query patterns before creating indexes

## `reminders`

- Purpose inferred: scheduled reminder records
- Sampled documents: `1`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `convertible_to_activity` | `bool` | true | false | false |
| `source` | `string` | true | false | false |
| `text` | `string` | true | false | false |
| `time` | `timestamp` | true | false | false |
| `user_id` | `*firestore.DocumentRef` | true | false | false |

Recommended indexes: `user_id+status`, `user_id+activity_id`, `user_id+status+next_run_at`

Differences vs expected model:

- Missing fields: `activity_id`, `created_by`, `id`, `next_run_at`, `scheduled_at`, `status`, `timezone`, `title`
- Unexpected fields: `_document_id`, `convertible_to_activity`, `text`, `time`

## `users`

- Purpose inferred: auth credentials and identity
- Sampled documents: `1`

| Field | Types | Required in sample | Optional | Sensitive |
| --- | --- | --- | --- | --- |
| `_document_id` | `string` | true | false | false |
| `ai_auto_action_conf_threshold` | `number` | true | false | false |
| `ai_autonomy_enabled` | `bool` | true | false | false |
| `base_personality_prompt` | `string` | true | false | false |
| `communication_style_score` | `number` | true | false | false |
| `context_mode` | `string` | true | false | false |
| `createdAt` | `timestamp` | true | false | false |
| `default_body_id` | `*firestore.DocumentRef` | true | false | false |
| `email` | `string` | true | false | true |
| `humor_tolerance_score` | `number` | true | false | false |
| `name` | `string` | true | false | false |
| `personality_version` | `number` | true | false | false |
| `phone` | `string` | true | false | false |
| `preferred_tone` | `string` | true | false | false |
| `safety_mode_enabled` | `bool` | true | false | false |
| `stress_baseline_score` | `number` | true | false | false |
| `timezone` | `string` | true | false | false |
| `tone_adaptivity_enabled` | `bool` | true | false | false |

Recommended indexes: `email`

Differences vs expected model:

- Missing fields: `created_at`, `id`, `password_hash`, `role`, `updated_at`
- Unexpected fields: `_document_id`, `ai_auto_action_conf_threshold`, `ai_autonomy_enabled`, `base_personality_prompt`, `communication_style_score`, `context_mode`, `createdAt`, `default_body_id`, `humor_tolerance_score`, `personality_version`, `phone`, `preferred_tone`, `safety_mode_enabled`, `stress_baseline_score`, `timezone`, `tone_adaptivity_enabled`

## Collection Differences

- Missing expected collections: `activity_outcomes`, `ai_action_proposals`, `ai_settings`, `memories`, `mood_entries`, `reflections`, `tool_definitions`, `user_profiles`
- Unexpected collections: `ai_actions_log`, `ai_memory`, `ai_policy`, `bodies`, `conversations`, `daily_summaries`, `locations`, `message_embeddings`
