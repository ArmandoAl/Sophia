# Flutter Reminders Integration

Sprint V1.3 conecta `sophia_ai` con el contrato real de reminders.

Endpoints: `GET/POST /reminders`, `GET/PATCH /reminders/{id}`, `POST /reminders/{id}/cancel`, `POST /reminders/{id}/archive` y `GET /reminders/due`.

Modelos: `Reminder`, `CreateReminderRequest`, `UpdateReminderRequest`, `ReminderListResponse`, `ReminderStatus`, `ReminderSource` y `RecurrenceRule`. El modelo refleja los campos que el handler Go devuelve; no edita campos internos de worker.

Cubits: `RemindersCubit`, `ReminderDetailCubit` y `ReminderEditorCubit`. La lista se refresca después de crear, editar, cancelar o archivar.

La pantalla permite listar, refrescar, crear, abrir detalle, cancelar y archivar. La recurrencia disponible es `none`, `daily`, `weekly` y `monthly`. `activity_id` queda opcional porque Activities no tiene UI integrada en este sprint.

No se conecta FCM, no se registran tokens y no se inicia el worker. Push real puede no estar activo todavía.

Riesgos: timezone usa la configuración local de Dart (`timeZoneName`), que puede ser una abreviatura en algunas plataformas; conviene derivarla del profile cuando el bootstrap de sesión la exponga de forma global.
