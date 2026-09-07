# Insights Domain

Fecha: 2026-06-30

## Proposito

`insights` registra informacion personal/productiva del usuario:

- como se sintio,
- que resultado tuvo una actividad,
- que reflexion dejo.

No implementa Memory, IA ni analisis psicologico clinico. Los datos son tracking personal para uso futuro de Sofia.

## Entidades

### MoodEntry

Campos:

- `id`
- `user_id`
- `activity_id`
- `mood`
- `energy_level`
- `stress_level`
- `focus_level`
- `note`
- `occurred_at`
- `timezone`
- `created_at`
- `updated_at`

Mood permitido:

- `very_low`
- `low`
- `neutral`
- `good`
- `great`

Niveles:

- `energy_level`: 1-5
- `stress_level`: 1-5
- `focus_level`: 1-5

### ActivityOutcome

Campos:

- `id`
- `user_id`
- `activity_id`
- `result`
- `completion_quality`
- `satisfaction_level`
- `difficulty_level`
- `note`
- `created_at`
- `updated_at`

Result permitido:

- `completed`
- `partially_completed`
- `skipped`
- `failed`
- `rescheduled`

Niveles:

- `completion_quality`: 1-5
- `satisfaction_level`: 1-5
- `difficulty_level`: 1-5

### Reflection

Campos:

- `id`
- `user_id`
- `activity_id`
- `title`
- `content`
- `sentiment`
- `tags`
- `occurred_at`
- `created_at`
- `updated_at`

## Endpoints

Todos requieren JWT.

### POST /insights/mood

Request:

```json
{
  "activity_id": "optional-activity-id",
  "mood": "good",
  "energy_level": 4,
  "stress_level": 2,
  "focus_level": 5,
  "note": "Felt focused after planning",
  "occurred_at": "2026-07-01T09:00:00Z",
  "timezone": "America/Tijuana"
}
```

### GET /insights/mood

Query params:

- `activity_id`
- `from`
- `to`
- `limit`

Response:

```json
{
  "mood_entries": []
}
```

### POST /insights/outcomes

Request:

```json
{
  "activity_id": "activity-id",
  "result": "completed",
  "completion_quality": 5,
  "satisfaction_level": 4,
  "difficulty_level": 3,
  "note": "Done with minor friction"
}
```

### GET /insights/outcomes

Query params:

- `activity_id`
- `from`
- `to`
- `limit`

Response:

```json
{
  "outcomes": []
}
```

### POST /insights/reflections

Request:

```json
{
  "activity_id": "optional-activity-id",
  "title": "Sprint note",
  "content": "I learned where planning got stuck.",
  "sentiment": "neutral",
  "tags": ["work", "sofia"],
  "occurred_at": "2026-07-01T10:00:00Z"
}
```

### GET /insights/reflections

Query params:

- `activity_id`
- `from`
- `to`
- `limit`

Response:

```json
{
  "reflections": []
}
```

### GET /insights/summary

Query params:

- `from`
- `to`
- `limit`

Response:

```json
{
  "average_energy": 3,
  "average_stress": 3,
  "average_focus": 4,
  "total_outcomes": 1,
  "total_reflections": 1
}
```

## Relacion con Activities

- `MoodEntry.activity_id` es opcional.
- `Reflection.activity_id` es opcional.
- `ActivityOutcome.activity_id` es requerido.
- Si se envia `activity_id`, debe pertenecer al usuario autenticado.
- Completar una Activity no obliga a crear outcome.
- El usuario puede registrar outcome despues.

## Privacidad y seguridad

- Todos los endpoints requieren JWT.
- `user_id` sale del token y no se acepta desde el body.
- Los repositorios filtran por `user_id`.
- No hay acceso cruzado entre usuarios.
- No se generan diagnósticos, perfiles psicologicos ni interpretaciones clinicas.

## Persistencia

Memory:

- `InMemoryMoodEntryRepository`
- `InMemoryActivityOutcomeRepository`
- `InMemoryReflectionRepository`

Firestore:

- `FirestoreMoodEntryRepository`
- `FirestoreActivityOutcomeRepository`
- `FirestoreReflectionRepository`

Colecciones:

- `mood_entries`
- `activity_outcomes`
- `reflections`

## Deuda tecnica

- Agregar paginacion por cursor.
- Agregar indices Firestore para filtros por `user_id`, `activity_id` y rango.
- Agregar tests con Firestore emulator.
- Validar timezone contra IANA.
- Definir si `ActivityOutcome` necesita `occurred_at` ademas de `created_at`.
- Definir politica de export/delete de datos sensibles del usuario.

## Riesgos

- Estos datos pueden ser sensibles aunque no sean clinicos.
- Summary actual es estadistica simple, no insight inteligente.
- `sentiment` es texto libre por ahora; no hay analisis de sentimiento automatico.
