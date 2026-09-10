# Sofía — Backend

API en Go que sostiene el bucle de aprendizaje de Sofía. Ver el
[README del monorepo](../../README.md) para la visión completa del sistema.

**Go 1.24 · Firestore · DeepSeek · Vertex AI · Cloud Run**

---

## Arranque rápido

Sin nube, sin emuladores, sin gastar en el proveedor de IA:

```bash
PERSISTENCE_DRIVER=memory JWT_SECRET=dev-secret AI_MODEL_PROVIDER=fake \
  go run ./cmd/sofia

curl localhost:8080/health
```

Contra Firestore real:

```bash
gcloud auth application-default login

PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=sophia-ai-60e87 \
FIRESTORE_DATABASE_ID=default \
JWT_SECRET=dev-secret AI_MODEL_PROVIDER=fake \
  go run ./cmd/sofia
```

> **`FIRESTORE_DATABASE_ID` es obligatorio.** Este proyecto usa una base *nombrada*
> `default`, que no es lo mismo que la base implícita `(default)` de Firestore. Si falta la
> variable, el binario falla al arrancar con un mensaje explícito — a propósito: un valor
> por defecto escondería el error hasta la primera consulta, y en un job programado eso
> significa morir cada noche sin que nadie lo note.

## Tests

```bash
go vet ./... && go test ./...
```

35 paquetes, sin red ni emuladores. Cada repositorio tiene una implementación en memoria
que los tests usan en lugar de Firestore, y el cliente de IA tiene un doble (`fake`).

---

## Estructura

```
cmd/
├── sofia/                    API HTTP
├── workers/
│   ├── reminders/            entrega de recordatorios (loop con lease)
│   ├── synthesis/            destilación nocturna de decisiones → creencias
│   └── ingestion/            importación de conversaciones externas
└── tools/                    utilidades operativas y migraciones

internal/
├── auth/ users/ activities/ reminders/ insights/ memory/
├── learning/                 creencias, prompt versionado, resúmenes diarios
├── ingestion/                importación externa
├── ai/runtime/               construcción de contexto, planificación, política
├── ai/actions/               propuestas, ejecución, autonomía
├── conversations/ tools/ notifications/ privacy/
├── platform/                 jsonschema, exportador y migración de Firestore
├── app/                      composición: repositorios y módulos
└── server/                   rutas y middleware
```

### Las cuatro capas

Cada módulo de `internal/` sigue la misma estructura, sin excepciones:

| Capa | Contiene | Puede depender de |
|---|---|---|
| `domain/` | entidades, reglas, interfaces de repositorio | nada externo |
| `application/` | casos de uso | `domain` |
| `infrastructure/` | `memory_*.go` **y** `firestore_*.go` | `domain` |
| `interfaces/http/` | handlers, requests, responses | `application` |

**Al añadir un método a una interfaz de repositorio hay que implementarlo en las dos
versiones** o los tests dejan de compilar. Es intencional: mantiene la suite rápida y sin
dependencias externas.

El cableado vive en `internal/app/modules.go` y `internal/app/repositories.go`.

---

## Los tres workers

### `reminders`
Loop con lease que entrega recordatorios vencidos por FCM.

### `synthesis` — el corazón
Una vez al día por usuario:

1. Reserva el día creando el `daily_summary` (ID determinista + `Create` de Firestore, que
   falla atómicamente si existe). **Ese fallo es el lease.**
2. Lee **solo las decisiones** del día. Nunca las conversaciones.
3. Una llamada al modelo con salida validada por JSON-Schema.
4. Aplica el delta a las creencias.
5. Recompila el prompt base desde cero, con techo de 1000 tokens.

El orden importa: reservar antes de mutar es lo que impide que un reintento aplique el
mismo delta dos veces e infle la evidencia.

### `ingestion`
Procesa lotes de conversaciones externas en ventanas, con el modelo barato y presupuesto
acotado. Todo lo que produce entra como creencia de **nivel 3** (techo de confianza 0.5) y
se puede deshacer por lote.

---

## Configuración

Todas las variables se leen y validan en `internal/config/config.go`. Ver
[`docs/environment.md`](docs/environment.md) para la lista completa.

### Imprescindibles

| Variable | Notas |
|---|---|
| `PERSISTENCE_DRIVER` | `memory` o `firestore` |
| `FIRESTORE_PROJECT_ID` | requerido con `firestore` |
| `FIRESTORE_DATABASE_ID` | **requerido con `firestore`**, sin valor por defecto |
| `JWT_SECRET` | obligatorio en producción |
| `AI_MODEL_PROVIDER` | `fake` o `deepseek` |

### Modelo por tarea

`ModelRequest` lleva un campo `Task`, y cada tarea puede usar un modelo distinto:

| Variable | Tarea | Frecuencia |
|---|---|---|
| `DEEPSEEK_MODEL` | por defecto, obligatorio | — |
| `DEEPSEEK_MODEL_PLAN` | camino caliente | cada mensaje |
| `DEEPSEEK_MODEL_SYNTHESIZE` | síntesis nocturna | 1×/día |
| `DEEPSEEK_MODEL_EXTRACT` | ingesta masiva | por lote |

Si una específica está vacía, cae al valor por defecto. **El camino caliente es el 88 % del
gasto**: es ahí donde conviene el modelo barato, no en la síntesis.

### Aprendizaje

| Variable | Efecto |
|---|---|
| `AUTONOMY_THRESHOLD` | probabilidad mínima para actuar sin preguntar (0.85) |
| `CONTEXT_TOKEN_BUDGET` | presupuesto de contexto por petición |
| `EMBEDDINGS_ENABLED` | activa la deduplicación semántica |
| `EMBEDDINGS_MODEL` | modelo de Vertex AI |
| `SYNTHESIS_RUN_HOUR_LOCAL` | hora local de cierre del día |
| `INGESTION_MAX_TOKENS_PER_BATCH` | tope de gasto por lote |

---

## Endpoints

| Grupo | Rutas |
|---|---|
| Auth | `/auth/register`, `/auth/login` |
| Usuario | `/users/me`, `/users/me/profile`, `/users/me/ai-settings`, `/users/me/export`, `/users/me/delete-request` |
| Actividades | `/activities`, `/activities/{id}` (+ complete, cancel, archive) |
| Recordatorios | `/reminders`, `/reminders/due`, `/reminders/{id}` |
| Insights | `/insights/mood`, `/insights/outcomes`, `/insights/reflections`, `/insights/summary` |
| Memoria | `/memory`, `/memory/search`, `/memory/{id}` |
| **Aprendizaje** | `/learning/beliefs`, `/learning/prompt-version`, `/learning/summaries` |
| **Contextos** | `/contexts`, `/contexts/{id}` |
| **Ingesta** | `/ingestion/conversations`, `/ingestion/batches/{id}` |
| Propuestas | `/ai/action-proposals` (+ confirm, reject, execute) |
| Runtime | `/ai/runtime/message` |
| Conversaciones | `/conversations`, `/conversations/{id}/messages` |
| Notificaciones | `/notifications/device-tokens` |
| Herramientas | `/tools`, `/tools/{id}` |

Todo salvo `/health` y `/auth/*` exige `Authorization: Bearer <jwt>` y **filtra siempre por
el usuario del token**. Ningún endpoint acepta un `user_id` del cliente.

Colección de Postman en [`docs/postman_collection.json`](docs/postman_collection.json).

---

## Herramientas

```bash
# Rellenar campos añadidos por migraciones (dry-run por defecto)
go run ./cmd/tools/backfill-memory-search-terms -dry-run
go run ./cmd/tools/backfill-belief-embeddings   -dry-run

# Diagnóstico y evaluación
go run ./cmd/tools/run-local-e2e
go run ./cmd/tools/run-ai-evals
go run ./cmd/tools/export-firestore-schema
```

Los backfills traen `-dry-run` activo por defecto; para escribir hace falta
`-dry-run=false -apply`. El de embeddings es reanudable con `-after <id>`.

---

## Despliegue

`Dockerfile` multi-etapa con dos targets: `sofia` (la API) y `synthesis` (el worker).

```bash
gcloud builds submit --tag gcr.io/sophia-ai-60e87/sofia-synthesis
```

El worker debe desplegarse como **Cloud Run Job**, no como Service: un Job se despierta,
trabaja y muere; un Service corre 24/7 y cuesta.

Su service account necesita `roles/datastore.user`, `roles/aiplatform.user` y
`roles/secretmanager.secretAccessor` — los permisos del operador no se heredan.

Procedimiento completo en [`../../docs/runbook_despliegue.md`](../../docs/runbook_despliegue.md).

---

## Convenciones

- Sin dependencias nuevas salvo necesidad demostrada.
- Sin abstracciones especulativas: nada de interfaces con una sola implementación.
- Los errores de dominio se declaran en `domain/` y se mapean en `infrastructure/`.
- Los audit logs guardan metadatos, **nunca contenido**: los mensajes se registran como
  huella (`len:47`), no como texto.
- `go vet ./... && go test ./...` en verde antes de cada commit.
