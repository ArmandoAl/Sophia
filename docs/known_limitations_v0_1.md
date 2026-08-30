# Known Limitations v0.1

Fecha: 2026-07-04

## Alcance General

v0.1 es operable personal, no produccion publica. Sirve para uso controlado, pruebas de persistencia y validacion de arquitectura modular.

## IA

- AI Runtime es proposal-only.
- No hay ejecucion autonoma.
- Gemini existe como provider opcional, pero no es default.
- Falta eval real sostenido con `GEMINI_API_KEY`.
- No hay dashboard de costo/calidad/seguridad.
- No hay RAG, embeddings ni memoria semantica.

## Notificaciones

- FCM esta deshabilitado por default.
- `FCM_DRY_RUN=true` por default.
- No hay envio push productivo habilitado.
- Falta limpieza automatica de tokens invalidos.
- Falta politica completa de preferencias de notificacion por usuario.

## Reminder Worker

- Worker deshabilitado por default.
- Lease transaccional existe en Firestore, pero falta validacion operacional con multiples procesos reales.
- La validacion Sprint 30 es deterministica con repositorio en memoria.
- No hay heartbeat ni extension de lease.
- Provider failure marca reminder como `failed`; retry/backoff programado sigue pendiente.

## Recurrencia

- Solo soporta `none`, `daily`, `weekly`, `monthly`.
- No hay RRULE completo.
- No hay excepciones, calendarios complejos ni manejo avanzado de DST.

## Firestore y Migraciones

- Firestore real puede contener colecciones legacy/conceptuales.
- El backend espera colecciones modulares nuevas.
- Migraciones son dry-run/staging/modular-only por default.
- No borrar colecciones legacy.
- No sobrescribir `users`, `activities` o `reminders` legacy sin aprobacion manual y backup.

## Privacidad

- Export de usuario y delete request existen.
- Borrado fisico requiere comando interno y request `approved`.
- `audit_logs` no se borran por default.
- `users` auth records no se borran por default.
- Falta politica final de retencion, export firmado y borrado de credenciales.

## Observabilidad

- Logging basico existe.
- Audit logs existen para eventos relevantes.
- Falta tracing, metricas externas, alerting y dashboards.

## Seguridad

- JWT protege endpoints privados.
- Ownership por `user_id` desde token esta implementado.
- Falta rate limiting.
- Falta hardening adicional para produccion publica.
- No exponer prompts completos, tokens, secrets ni payloads sensibles.

