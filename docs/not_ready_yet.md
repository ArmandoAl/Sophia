# Not Ready Yet

Fecha: 2026-07-03

Este documento lista explicitamente lo que no debe tratarse como listo para produccion o autonomia.

## IA y Autonomia

- No hay ejecucion autonoma.
- El runtime es proposal-only.
- Gemini existe como provider, pero falta corrida real de evals con API key en este entorno.
- No hay decision final de proveedor/modelo para produccion.
- No hay dashboard externo de seguridad/costo/calidad.

## Notificaciones

- FCM existe, pero esta deshabilitado y dry-run por default.
- No hay rollout productivo de push.
- No hay preferencias completas de notificacion por usuario.
- No hay limpieza automatica de tokens invalidos.
- No hay auditoria detallada de cada intento de delivery.

## Reminder Worker

- El worker ya tiene claim/lease transaccional para Firestore.
- Falta validacion operacional con Firestore emulator y multiples procesos reales.
- No hay heartbeat/lease extension para providers lentos.
- Retry scheduling todavia es basico.

## Recurrencia

- Existe recurrencia simple: `none`, `daily`, `weekly`, `monthly`.
- No hay RRULE completo.
- No hay excepciones, calendarios complejos ni manejo avanzado de DST.

## Memoria e IA Contextual

- Memory textual esta usable.
- No hay embeddings.
- No hay vector search.
- No hay RAG completo.
- La importacion de `ai_memory` legacy sigue pendiente por privacidad/calidad.

## Privacidad

- Export y delete request existen.
- Delete fisico requiere comando interno y request aprobada.
- `users` auth records y `audit_logs` se retienen por default.
- Falta politica final de borrado de credenciales, retencion legal y estado visible al usuario.

## Firestore y Datos Legacy

- El backend espera colecciones modulares.
- Firestore real todavia contiene colecciones legacy/conceptuales.
- Migradores existen en modos seguros, pero no se debe aplicar migracion destructiva.
- Falta cerrar emulator coverage e indices deployables como parte de operacion.

## Hardware

- No hay integracion hardware/smart home.
- Colecciones legacy como `bodies` y `locations` no forman parte del runtime actual.
# V1 Release Gate additions (2026-07-16)

- Firestore controlled E2E requires valid external credentials, project/database configuration and deployed V1 indexes.
- Gemini real E2E requires an external `GEMINI_API_KEY` with available quota.
