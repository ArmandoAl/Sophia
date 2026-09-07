# Model Provider Strategy

Fecha: 2026-07-01

## Estado actual

Sprint 15 agrega Gemini como primer proveedor real detras de la interfaz `ModelClient`.

El proveedor por defecto sigue siendo `fake` para desarrollo seguro y tests.

El backend define:

```go
type ModelClient interface {
    Generate(ctx context.Context, request ModelRequest) (ModelResponse, error)
}
```

Implementaciones actuales:

- `FakeModelClient`, usado para validar arquitectura, endpoint, context builder, tool selection y safety policy sin llamar servicios externos.
- `Gemini ModelClient`, usado solo cuando `AI_MODEL_PROVIDER=gemini`.

## Candidatos

### Gemini

Gemini es candidato natural por el ecosistema Google/Firebase:

- afinidad con Firestore/Firebase,
- posible integracion operacional con Google Cloud,
- buen encaje si Sofia se despliega mayormente en infraestructura Google.

Riesgos a evaluar:

- calidad de structured outputs/tool calling para este caso,
- costo por contexto largo,
- observabilidad,
- controles de seguridad,
- compatibilidad con estrategias futuras de memoria/grounding.

### OpenAI

OpenAI es candidato fuerte por structured outputs y tool calling:

- buen soporte para salidas estructuradas,
- ergonomia para herramientas y reasoning controlado,
- opciones maduras para evaluaciones y safety workflows.

Riesgos a evaluar:

- costo,
- residencia de datos,
- integracion con infraestructura Google/Firebase,
- politicas de retencion/configuracion empresarial.

## Criterios de decision

- calidad para planificacion personal,
- confiabilidad de JSON/structured outputs,
- costo por usuario activo,
- latencia,
- seguridad y controles de datos,
- facilidad de observabilidad/evals,
- soporte para tool calling controlado,
- compatibilidad con privacidad/export/delete.

## Decision actual

Gemini se integra primero por afinidad con Google/Firebase y Firestore.

La arquitectura sigue manteniendo la interfaz generica para que OpenAI u otro proveedor puedan evaluarse despues sin cambiar handlers ni casos de uso de producto.

La decision de proveedor definitivo para produccion sigue pendiente de evals de costo, calidad y seguridad.

Sprint 20 deja Gemini como eval real opcional: el runner siempre ejecuta `fake`, y solo ejecuta `gemini` cuando `GEMINI_API_KEY` existe. La ausencia de API key genera un reporte `skipped`, no un fallo de CI.

## Principio arquitectonico

El backend debe hablar con `ModelClient`, no con Gemini/OpenAI directamente desde handlers o casos de uso de producto.

Los providers reales deberian vivir en:

```text
internal/ai/runtime/infrastructure
```

Y deben cumplir:

- no loggear prompts con datos sensibles sin redaccion,
- timeouts,
- retries seguros,
- metricas,
- pruebas con fake provider,
- feature flag/config explicita.

## Config actual

```sh
AI_MODEL_PROVIDER=fake|gemini
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-2.5-flash|gemini-1.5-flash
AI_RUNTIME_ENABLED=true
AI_RUNTIME_PROPOSAL_ONLY=true
```

`AI_RUNTIME_PROPOSAL_ONLY=false` se rechaza por configuracion.

## Diagnostico Gemini

Para diagnosticar 401/403/404/quota/model access sin exponer secretos:

```sh
go run ./cmd/tools/test-gemini-provider -output-dir=docs
```

El provider clasifica errores como:

- `unauthorized`
- `permission_denied`
- `model_not_found`
- `quota`
- `request_failed`

El body de error se redacta/trunca y no incluye API keys.
