# Sprint V1.3 Report

Estado: implementado en Flutter.

- Backend conectado: reminders collection, detail, update, cancel, archive y due.
- Mocks removidos: `TaskItem` y los cinco reminders hardcodeados de `RemindersCubit`.
- Pantallas: lista, editor modal y detalle.
- Recurrence simple: none, daily, weekly, monthly.
- Fuera de alcance: FCM, device tokens, worker, RRULE y Activities UI completa.
- Campos no conectados: `processing_by`, `processing_until` y `last_attempt_at` no aparecen en responses del backend; no se modelan para edición.
- Pruebas: parseo, requests, repository, cubit y UI deben cubrirse con el suite existente; no se agregó dependencia nueva.
- Riesgos: timezone local y backend de reminders deshabilitado pueden producir errores `403/409`, que la UI muestra como error recuperable.
