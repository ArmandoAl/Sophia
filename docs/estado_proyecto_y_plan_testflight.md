# Estado de Sofía y plan para llegar a TestFlight

Fecha de auditoría: 2026-08-30  
Alcance: `sophia_ai` (Flutter), `sophia_ai_backend/sofia-backend` (Go), migración Gemini → DeepSeek, despliegue con Firebase y preparación de TestFlight.

## Resumen ejecutivo

Sofía ya tiene una base V1 funcional en código. El frontend consume el backend para autenticación, sesión, onboarding, perfil, ajustes de IA, diagnóstico, conversaciones persistentes y reminders. El backend ofrece esos flujos y además contiene módulos de actividades, insights, memoria, privacidad, notificaciones, propuestas de IA y herramientas.

El proyecto todavía **no está listo para TestFlight externo** porque faltan cuatro gates:

1. Integrar y evaluar DeepSeek con una API key real.
2. Desplegar el backend Go con Firestore real y secretos seguros.
3. Configurar identidad, firma y metadatos reales de iOS.
4. Ejecutar E2E en staging y pruebas manuales en dispositivos físicos.

La V1 más corta para TestFlight debe incluir: auth/onboarding, chat, reminders, perfil/ajustes y diagnóstico. Smart Home, activities UI, insights reales, memoria administrable, push real y ejecución de propuestas pueden quedar fuera del primer beta.

> Aclaración de infraestructura: Firebase Hosting no ejecuta directamente el proceso Go. La arquitectura correcta es **Firebase Hosting como fachada HTTPS con rewrite hacia un contenedor Go en Cloud Run**. Firebase documenta oficialmente esta integración para APIs y microservicios: [Hosting + Cloud Run](https://firebase.google.com/docs/hosting/serverless-overview) y [rewrites a Cloud Run](https://firebase.google.com/docs/hosting/full-config#direct_requests_to_a_cloud_run_container).

## Evidencia verificada en esta auditoría

| Verificación | Resultado |
| --- | --- |
| `go test ./...` | PASS |
| `go vet ./...` | PASS |
| `go build ./cmd/sofia` | PASS |
| E2E local del backend | PASS, 13/13 pasos |
| `fvm flutter analyze` | PASS, sin issues |
| `fvm flutter test` | PASS, 97 tests |
| Build iOS release sin firma | PASS, `Runner.app` generado |
| Build firmado / IPA | No probado; requiere identidad y provisioning definitivos |
| Firestore real | No probado en esta auditoría |
| Proveedor de IA real | No probado; Gemini no tiene credencial operativa y DeepSeek aún no está implementado |
| Backend desplegado | No existe configuración de despliegue en el workspace |

El build iOS actual dispara una migración automática del proyecto a la versión reciente de Flutter/UIScene. Debe ejecutarse y revisarse deliberadamente antes de generar el IPA; no debe descubrirse por primera vez durante el upload final.

## Estado real del frontend Flutter

### Implementado y conectado al backend

- Cliente HTTP con `SOFIA_API_BASE_URL`, Bearer token y errores tipados.
- Token en `flutter_secure_storage`.
- Registro, login, logout y restauración de sesión.
- Navegación protegida y onboarding.
- Perfil y ajustes del asistente.
- Diagnóstico mediante `GET /health`.
- Chat real sobre `/conversations` y `/conversations/{id}/messages`.
- Historial persistente y render de propuestas del runtime.
- CRUD, cancelación y archivo de reminders.
- Tests de red, modelos, navegación, chat y reminders.

### Pendiente o incompleto

- Los botones **Confirmar** y **Modificar** de una propuesta en chat solo hacen `debugPrint`; no llaman al backend.
- Smart Home sigue siendo un prototipo local y aparece como rama del shell aunque el backend no lo soporta.
- La UI de insights es prototipo local; no consume el módulo real del backend.
- No hay UI real para activities, memory, notifications, privacy export/delete ni administración de conversaciones.
- No existe registro de token FCM ni recepción de push en iOS.
- Los `FeatureFlags` existen, pero sus defaults no gobiernan de forma consistente las rutas visibles; por ejemplo, chat real está activo aunque `aiRuntimeEnabled` inicia en `false`.
- `README.md` y la descripción de `pubspec.yaml` siguen siendo los textos del template Flutter.
- El bundle ID es `com.example.sophiaAi`; debe reemplazarse antes del primer upload.
- El nombre visible es `Sophia Ai`; falta confirmar naming final, iconos, splash y textos legales.
- La URL por defecto sigue siendo `http://localhost:8080`; el build beta debe llevar explícitamente la URL HTTPS de staging/producción.

### Alcance recomendado para el primer beta

Mantener visibles únicamente:

- Chat.
- Reminders.
- Perfil y ajustes.
- Diagnóstico, preferentemente oculto para testers normales.

Ocultar Smart Home e insights mock. No construir todavía pantallas para todos los módulos que ya existen en backend; no son necesarias para validar la propuesta central de Sofía.

## Estado real del backend Go

### Implementado

- API HTTP modular con auth JWT y rate limit de auth.
- Persistencia seleccionable: memory o Firestore.
- Usuarios, perfil, preferencias y onboarding.
- Activities y reminders, incluida recurrencia básica.
- Conversaciones y mensajes persistentes con ownership.
- Runtime de IA detrás de la interfaz `ModelClient`.
- Context builder, tools, policy/safety y modo proposal-only.
- Propuestas explícitas; no hay ejecución autónoma desde el modelo.
- Insights, memory, notificaciones, export y delete request.
- Worker de reminders y proveedor FCM, apagados por defecto.
- Cobertura amplia de handlers, seguridad, repositorios y rutas.

### Pendiente para operar en producción

- No hay `Dockerfile`, `firebase.json`, `.firebaserc`, `firestore.indexes.json` ni pipeline de deploy.
- No se ha hecho E2E controlado contra Firestore real.
- No están desplegados/verificados los índices requeridos por las consultas V1.
- El worker multi-instancia y FCM real no tienen validación productiva.
- Falta observabilidad externa: logs estructurados, alertas mínimas y medición de latencia/error del proveedor de IA.
- El backend no tiene repositorio Git propio y el workspace raíz tampoco. Solo `sophia_ai` tiene Git, con gran parte de la implementación actual sin commit. Hoy no existe un snapshot reproducible de todo el producto.

### Riesgo de seguridad inmediato

Existe un archivo local de service account en `sophia_ai_backend/sophia-backend-key.json`, con permisos `0644`, dentro del workspace y fuera de un repositorio que permita comprobar su tracking.

Antes de desplegar:

- Rotar la credencial si alguna vez fue compartida, respaldada o subida.
- Retirar el JSON del workspace.
- Usar la identidad del servicio de Cloud Run para Firestore mediante ADC.
- Guardar `DEEPSEEK_API_KEY` y `JWT_SECRET` en Secret Manager.
- Nunca incluir secretos en `firebase.json`, Docker image, Dart defines ni el binario iOS.

## Auditoría de la migración Gemini → DeepSeek

### Conclusión

La arquitectura ya tiene el punto de extensión correcto:

```go
type ModelClient interface {
    Generate(context.Context, ModelRequest) (ModelResponse, error)
}
```

El runtime, conversaciones, handlers, policy engine, tools y Flutter no dependen directamente de Gemini. Por eso el cambio debe limitarse al paquete de infraestructura, configuración, wiring, diagnóstico/evals y documentación.

### Implementación mínima recomendada

1. Crear `internal/ai/runtime/infrastructure/deepseek/client.go` usando `net/http`; no agregar un SDK.
2. Mantener el contrato interno `ModelRequest` → `ModelResponse` sin cambios.
3. Agregar configuración:
   - `AI_MODEL_PROVIDER=deepseek`
   - `DEEPSEEK_API_KEY`
   - `DEEPSEEK_MODEL=deepseek-v4-flash`
   - `DEEPSEEK_BASE_URL=https://api.deepseek.com` solo si se necesita override para tests.
4. Agregar el caso `deepseek` en `BuildModelClient` y en la validación de `config.Load`.
5. Reutilizar las mismas instrucciones proposal-only y la misma validación posterior de tools/policy.
6. Añadir tests HTTP en memoria equivalentes a Gemini: success, JSON inválido/vacío, timeout, 401/403/404/429/5xx, truncamiento y redacción de API key.
7. Añadir DeepSeek al runner de evals y ejecutar casos reales antes de cambiar producción.
8. Mantener Gemini solo durante el rollout como rollback; eliminarlo cuando DeepSeek haya pasado los gates y no haya intención real de usarlo.

DeepSeek ofrece un endpoint compatible con Chat Completions y JSON mode. La llamada debe usar `POST /chat/completions`, Bearer auth y `response_format: {"type":"json_object"}`. El prompt también debe pedir JSON y el cliente debe rechazar contenido vacío o truncado; estas condiciones están señaladas por la propia [documentación de JSON Output](https://api-docs.deepseek.com/guides/json_mode/) y el [contrato de Chat Completions](https://api-docs.deepseek.com/api/create-chat-completion/).

Al 2026-08-30, no se debe configurar `deepseek-chat`: el alias fue retirado. Para el beta, empezar con `deepseek-v4-flash` por latencia/costo y comparar contra `deepseek-v4-pro` en los evals oficiales del proyecto. La disponibilidad y nombres vigentes están en el [changelog oficial de DeepSeek](https://api-docs.deepseek.com/updates/).

### Mapeo de request/response

Request mínimo:

```json
{
  "model": "deepseek-v4-flash",
  "messages": [
    {"role": "system", "content": "Return JSON only. Never execute actions."},
    {"role": "user", "content": "<contexto compacto y mensaje>"}
  ],
  "thinking": {"type": "disabled"},
  "temperature": 0.2,
  "response_format": {"type": "json_object"}
}
```

Leer `choices[0].message.content`, deserializar la forma existente:

```json
{
  "assistant_message": "string",
  "proposed_actions": []
}
```

No usar tool calling nativo de DeepSeek en la primera migración. El proyecto ya tiene un formato de propuestas, validación JSON Schema, ToolSelector y PolicyEngine; reemplazar todo eso añadiría riesgo sin desbloquear TestFlight.

### Gates de aceptación del proveedor

- 100% de los casos safety existentes pasan.
- Ninguna acción se ejecuta automáticamente.
- JSON válido y compatible con el contrato en los evals seleccionados.
- API keys y datos sensibles no aparecen en logs ni errores.
- Timeout y errores de cuota se traducen a respuestas controladas.
- Medición básica de latencia, tasa de error y tokens/costo por caso.
- Prueba sintética real desde la misma región de Cloud Run.

## Arquitectura de despliegue recomendada

```text
Flutter iOS (TestFlight)
        |
        | HTTPS
        v
Firebase Hosting (dominio estable / fachada)
        |
        | rewrite
        v
Cloud Run: API Go
   |              |
   v              v
Firestore      DeepSeek API
```

### Decisiones mínimas

- Un servicio Cloud Run para `cmd/sofia`.
- Un site de Firebase Hosting dedicado al API, con rewrite global al servicio Cloud Run. Así el backend conserva rutas como `/health`, `/auth/*` y `/conversations/*` sin agregar prefijos.
- Misma región compatible para Hosting/Cloud Run y cercana a Firestore.
- Cloud Run autenticado públicamente a nivel de transporte, porque el backend aplica JWT en rutas privadas; los secretos permanecen en Secret Manager.
- `ENV=production`, `PERSISTENCE_DRIVER=firestore`, `AI_MODEL_PROVIDER=deepseek` y `AI_RUNTIME_PROPOSAL_ONLY=true`.
- Para el primer TestFlight, mantener `REMINDER_WORKER_ENABLED=false`, `FCM_ENABLED=false` y delivery `noop`. El proceso polling actual no debe vivir dentro del mismo servicio autoscalable de la API.

Si push es requisito del beta, convertir la ejecución del worker en un trabajo invocable y programarlo por separado; no mantener un loop infinito dentro de Cloud Run API.

### Archivos que faltan

- `Dockerfile` multi-stage para compilar y ejecutar `cmd/sofia` como usuario no root.
- `.dockerignore` que excluya credenciales, caches y documentación innecesaria.
- `firebase.json` con rewrite al servicio Cloud Run.
- `.firebaserc` o selección explícita de proyecto en CI.
- `firestore.indexes.json` validado contra todas las queries V1.
- Runbook de variables, Secret Manager, deploy, smoke test y rollback.

## Preparación de TestFlight

Apple exige crear el registro de la app, firmar con un identificador/provisioning válido y subir un build antes de distribuirlo. Referencias: [flujo de App Store Connect](https://developer.apple.com/help/app-store-connect/get-started/app-store-connect-workflow/), [subir builds](https://developer.apple.com/help/app-store-connect/manage-builds/upload-builds) y [TestFlight](https://developer.apple.com/help/app-store-connect/test-a-beta-version/testflight-overview/).

### Bloqueadores iOS

- Cambiar `com.example.sophiaAi` por el bundle ID definitivo y registrarlo en Apple Developer/App Store Connect.
- Confirmar que `DEVELOPMENT_TEAM=K2NQ868JMG` corresponde a la cuenta que publicará la app.
- Ejecutar y versionar la migración iOS solicitada por la versión actual de Flutter.
- Revisar `Podfile.lock`; el build debe incluir realmente `flutter_secure_storage`.
- Sustituir iconos/splash de template y confirmar el nombre comercial.
- Incrementar `version` y build number en cada upload.
- Generar un IPA firmado con la URL HTTPS final mediante `--dart-define=SOFIA_API_BASE_URL=...`.
- Completar privacidad, export compliance, descripción beta, contacto y “What to Test”.
- Probar login, token seguro, suspensión/reapertura, red lenta, pérdida de sesión y actualización de app en iPhone físico.

## Plan de cierre ordenado

### Gate 0 — Snapshot y secretos

- [ ] Crear un repositorio raíz o repos separados para frontend y backend.
- [ ] Guardar el estado actual en commits revisables.
- [ ] Retirar/rotar el service account JSON.
- [ ] Definir proyecto Firebase staging y producción.

### Gate 1 — DeepSeek

- [ ] Implementar el adapter mínimo.
- [ ] Pasar tests unitarios y safety.
- [ ] Ejecutar evals reales Flash vs Pro.
- [ ] Elegir modelo con evidencia de calidad, latencia y costo.

### Gate 2 — Backend staging

- [ ] Crear contenedor y desplegar Cloud Run.
- [ ] Conectar Firebase Hosting mediante rewrite.
- [ ] Configurar Firestore, índices, ADC y Secret Manager.
- [ ] Ejecutar smoke y E2E con persistencia real.
- [ ] Verificar restart/redeploy sin pérdida de datos.

### Gate 3 — Cliente beta

- [ ] Ocultar Smart Home e insights mock.
- [ ] Decidir si las propuestas solo se muestran o si se implementará confirmación real; no dejar botones falsos.
- [ ] Probar chat/reminders contra staging.
- [ ] Confirmar estados vacíos, offline, 401, 429 y error del proveedor.
- [ ] Actualizar identidad visual y documentación básica.

### Gate 4 — TestFlight

- [ ] Bundle ID, team, certificados y provisioning.
- [ ] IPA firmado con URL de staging.
- [ ] Upload a App Store Connect.
- [ ] Prueba interna en dispositivos reales.
- [ ] Corregir crashes/bloqueos y luego abrir beta externa si aplica.

## Definición de “listo para TestFlight interno”

Sofía está lista cuando un usuario nuevo, desde un iPhone físico y una instalación limpia, puede:

1. Registrarse e iniciar sesión.
2. Completar onboarding.
3. Enviar mensajes y recuperar historial después de reiniciar la app.
4. Recibir respuestas reales de DeepSeek sin ejecución autónoma.
5. Crear, editar, cancelar y archivar reminders.
6. Cerrar sesión y volver a autenticarse.
7. Repetir el flujo tras un redeploy del backend sin perder datos.

Además, el equipo debe poder detectar errores en logs, rotar secretos y regresar a una revisión anterior de Cloud Run/Hosting.

## Lo que no bloquea el primer TestFlight

- Smart Home/hardware.
- RAG, embeddings o vector search.
- RRULE y recurrencia avanzada.
- Streaming de tokens, voz, imágenes o attachments.
- Push real, si se comunica que reminders se validan dentro de la app.
- UI para todos los bounded contexts del backend.
- Ejecución automática de acciones; debe seguir prohibida.

Estas capacidades deben añadirse después de validar que auth, chat, persistencia y reminders funcionan con usuarios beta reales.
