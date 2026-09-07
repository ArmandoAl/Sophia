# Users Domain

Fecha: 2026-06-29

## Objetivo

El dominio users modela la identidad funcional de una persona dentro de Sofia. Se separa de auth para evitar mezclar credenciales con preferencias personales y configuracion de IA.

## Entidades

### Auth User

Vive en `internal/auth/domain`.

Campos principales:

- `id`
- `name`
- `email`
- `password_hash` internamente como `Password`, no serializable
- `role`
- `created_at`
- `updated_at`

Firestore:

- Coleccion: `users`
- Campo persistido: `password_hash`

### UserProfile

Vive en `internal/users/domain`.

Campos:

- `id`
- `user_id`
- `display_name`
- `preferred_name`
- `timezone`
- `locale`
- `avatar_url`
- `onboarding_completed`
- `created_at`
- `updated_at`

Defaults:

- `id`: igual a `user_id`
- `display_name`: nombre usado en register
- `preferred_name`: nombre usado en register
- `timezone`: `America/Tijuana`
- `locale`: `es-MX`
- `onboarding_completed`: `false`

Firestore:

- Coleccion: `user_profiles`
- Documento: `user_id`

### AISettings

Vive en `internal/users/domain`.

Campos:

- `user_id`
- `assistant_name`
- `communication_style`
- `proactivity_level`
- `autonomy_level`
- `memory_enabled`
- `reminders_enabled`
- `planning_enabled`
- `preferred_language`
- `created_at`
- `updated_at`

Defaults:

- `assistant_name`: `Sofía`
- `communication_style`: `warm`
- `proactivity_level`: `medium`
- `autonomy_level`: `suggestive`
- `memory_enabled`: `true`
- `reminders_enabled`: `true`
- `planning_enabled`: `true`
- `preferred_language`: `es`

Valores permitidos:

- `proactivity_level`: `low`, `medium`, `high`
- `autonomy_level`: `manual`, `suggestive`, `semi_autonomous`

Firestore:

- Coleccion: `ai_settings`
- Documento: `user_id`

## Endpoints

### GET /users/me

Auth: requerida.

Response:

```json
{
  "user": {
    "id": "user-id",
    "name": "Armando",
    "email": "armando@example.com",
    "role": "user"
  },
  "profile": {
    "id": "user-id",
    "user_id": "user-id",
    "display_name": "Armando",
    "preferred_name": "Armando",
    "timezone": "America/Tijuana",
    "locale": "es-MX",
    "onboarding_completed": false,
    "created_at": "2026-06-29T00:00:00Z",
    "updated_at": "2026-06-29T00:00:00Z"
  },
  "ai_settings": {
    "user_id": "user-id",
    "assistant_name": "Sofía",
    "communication_style": "warm",
    "proactivity_level": "medium",
    "autonomy_level": "suggestive",
    "memory_enabled": true,
    "reminders_enabled": true,
    "planning_enabled": true,
    "preferred_language": "es",
    "created_at": "2026-06-29T00:00:00Z",
    "updated_at": "2026-06-29T00:00:00Z"
  }
}
```

### PATCH /users/me/profile

Auth: requerida.

Request:

```json
{
  "display_name": "Armando Alvarado",
  "preferred_name": "Mando",
  "timezone": "America/Mexico_City",
  "locale": "es-MX",
  "avatar_url": "https://example.com/avatar.png"
}
```

Validaciones:

- `timezone` no puede ser vacio si se envia.
- `locale` no puede ser vacio si se envia.
- Campos desconocidos se rechazan.

### GET /users/me/ai-settings

Auth: requerida.

Devuelve solo AISettings.

### PATCH /users/me/ai-settings

Auth: requerida.

Request:

```json
{
  "assistant_name": "Sofi",
  "communication_style": "direct",
  "proactivity_level": "high",
  "autonomy_level": "semi_autonomous",
  "memory_enabled": true,
  "reminders_enabled": true,
  "planning_enabled": false,
  "preferred_language": "es-MX"
}
```

Validaciones:

- `assistant_name` no puede ser vacio si se envia.
- `preferred_language` no puede ser vacio si se envia.
- `proactivity_level`: `low`, `medium`, `high`.
- `autonomy_level`: `manual`, `suggestive`, `semi_autonomous`.
- Campos desconocidos se rechazan.

### POST /users/me/onboarding/complete

Auth: requerida.

Marca `onboarding_completed=true` en el perfil autenticado.

## Seguridad

- No se acepta `user_id` en requests.
- Todos los endpoints usan el user ID del JWT.
- No hay acceso cruzado entre usuarios.
- `password_hash` no aparece en responses.

## Deuda tecnica

- Validar timezone contra base IANA.
- Validar locale contra BCP 47.
- Hacer transaccional la creacion auth user + profile + AI settings en Firestore.
- Agregar tests contra Firestore emulator.
