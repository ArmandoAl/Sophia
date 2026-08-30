# Audit Log

Fecha: 2026-07-01

## Coleccion

`audit_logs`

## Proposito

Registrar eventos sensibles de privacidad y operaciones futuras de AI/tooling sin guardar secretos ni cuerpos sensibles completos.

## Campos

| Campo | Tipo | Descripcion |
| --- | --- | --- |
| `id` | string | ID del evento. |
| `user_id` | string | Usuario propietario del evento. |
| `action` | string | Accion auditada. |
| `resource_type` | string | Tipo de recurso relacionado. |
| `resource_id` | string | ID de recurso relacionado. |
| `created_at` | timestamp | Fecha de creacion. |
| `metadata` | map | Metadata redacted/truncada. |

## Acciones iniciales

| Accion | Cuando se registra |
| --- | --- |
| `user_data_export_requested` | Al ejecutar `GET /users/me/export`. |
| `user_delete_requested` | Al ejecutar `POST /users/me/delete-request`. |
| `user_delete_processed` | Al completar el comando interno de borrado fisico. |
| `user_delete_failed` | Cuando el comando interno de borrado fisico falla para una solicitud. |

## Redaccion

La metadata redacta por nombre de campo:

- `token`
- `secret`
- `password`
- `email`

Y trunca campos textuales sensibles como:

- `reason`
- `content`
- `note`

## Indices sugeridos

- `user_id + created_at`
- `action + created_at`
- `resource_type + resource_id`

## Restricciones actuales

- No hay endpoint publico para consultar audit logs.
- No hay politica de retencion.
- No hay export especifico de audit logs fuera del user export.
- No debe usarse para guardar prompts completos sin redaccion fuerte.
