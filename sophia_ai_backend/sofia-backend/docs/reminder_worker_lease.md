# Reminder Worker Lease

Fecha: 2026-07-03

## Objetivo

Permitir que multiples instancias del reminder worker compartan la misma coleccion Firestore sin entregar el mismo reminder en paralelo.

## Campos

Cada reminder puede tener metadata de claim:

- `processing_by`: ID del worker que reclama el intento actual.
- `processing_until`: expiracion del lease.
- `last_attempt_at`: timestamp del inicio del intento.

## Flujo

1. El worker lista candidatos con `status=pending` y `next_run_at <= now`.
2. Para cada candidato ejecuta `ClaimDue`.
3. `ClaimDue` valida:
   - el reminder existe,
   - pertenece al usuario esperado,
   - sigue `pending`,
   - sigue due,
   - no tiene lease activo o el lease expiro.
4. Si es claimable, escribe `processing_by`, `processing_until` y `last_attempt_at`.
5. Solo el worker que obtuvo claim llama al delivery provider.
6. `MarkDelivered` o `MarkFailed` limpia `processing_by` y `processing_until`.

## Firestore

En Firestore, `ClaimDue` usa `RunTransaction` sobre `reminders/{id}`. Esto hace que dos workers compitiendo por el mismo documento no puedan reclamarlo simultaneamente.

El scan de candidatos sigue usando:

```text
status asc, next_run_at asc
```

El claim no necesita indice adicional porque opera por document ID.

## Configuracion

| Variable | Default | Descripcion |
| --- | --- | --- |
| `REMINDER_WORKER_ID` | host-derived | Identidad estable del worker. Debe ser unica por instancia en despliegues reales. |
| `REMINDER_WORKER_LEASE_DURATION` | `2m` | Duracion del lease. Debe cubrir latencia esperada del provider. |

## Comportamiento

- Si un worker cae despues de reclamar, otro worker podra reclamar cuando `processing_until` expire.
- Si un reminder ya fue `delivered` o `failed`, no se reclama de nuevo.
- Si el provider falla, el reminder queda `failed`, incrementa `retry_count` y libera el lease.
- La recurrencia simple sigue funcionando: un reminder recurrente entregado avanza `next_run_at` y libera el lease.

## Limitaciones

- No hay heartbeat/lease extension durante entregas largas.
- Retry scheduling sigue siendo basico: los fallos del provider finalizan el reminder como `failed`.
- La seguridad operacional depende de configurar un `REMINDER_WORKER_LEASE_DURATION` mayor que la latencia esperada del provider.
