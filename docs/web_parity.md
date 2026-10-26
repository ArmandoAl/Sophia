# Matriz de paridad: app móvil ↔ web (Jaspr)

Estado al 2026-10-26. Prueba: `apps/web/e2e/sophia.e2e.mjs` (E), `packages/core/test` (C). Todo ✔ se probó **en modo demo** (datos ficticios), no contra el backend real.

| Flujo | App móvil | Web | Estado | Prueba |
| --- | --- | --- | --- | --- |
| Acceso (correo + contraseña) | ✔ | ✔ mismo `AuthCubit`/`SessionCubit`; token en `sessionStorage` | ✔ demo · ⏳ real | E |
| Cerrar sesión | ✔ | ✔ | ✔ | E |
| Conversación: historial, envío, escritura, error con reintento | ✔ | ✔ mismo `ChatMessageCubit` | ✔ demo | E, C |
| Respuestas sugeridas | ✔ | ✔ | ✔ | E, C |
| Propuestas de acción (aprobar / rechazar) | ✔ | **sólo lectura** («se confirma desde la app móvil») | parcial | E |
| Contexto manual / entidades, voz | ✔ | no | pendiente | — |
| Recordatorios, actividades, perfil, ajustes, creencias, insights | ✔ | no | pendiente (los repositorios ya están en el core) | — |
| Notificaciones (Firebase Messaging) | ✔ | no | requiere adaptación de navegador | — |
| Teléfono 390 px, tema claro/oscuro, accesibilidad (axe A/AA) | n/a | ✔ | ✔ | E |

## Decisiones
- El token web vive en `sessionStorage` (dura la pestaña): no hay Keychain en el navegador y no debe tratarse como bóveda.
- Registro de usuarios no está en la web todavía.
- `sophia.riv` del avatar de `sofia_prompt` mide 0 bytes: no hay avatar en la web.
