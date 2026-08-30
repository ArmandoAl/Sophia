# Gemini Troubleshooting

Fecha: 2026-07-04

## Objetivo

Diagnosticar errores del provider Gemini sin loggear API keys, prompts completos, contexto completo ni raw payloads sensibles.

## Comando De Diagnostico

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/test-gemini-provider -output-dir=docs
```

Con modelo explicito:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/test-gemini-provider -model=gemini-2.5-flash -output-dir=docs
```

El comando escribe:

```text
docs/gemini_provider_diagnostic.md
```

## Modelos Configurables

Valores documentados para `GEMINI_MODEL`:

- `gemini-2.5-flash`
- `gemini-1.5-flash`

El backend mantiene `gemini-1.5-flash` como default historico para compatibilidad, pero el comando permite probar explicitamente `gemini-2.5-flash`.

## Tipos De Error

| Tipo | Significado probable | Accion sugerida |
| --- | --- | --- |
| `unauthorized` | API key ausente, invalida o revocada. | Revisar `GEMINI_API_KEY` sin imprimirla. |
| `permission_denied` | Key sin acceso, API no habilitada o restriccion de proyecto/referrer/IP. | Revisar permisos del proyecto Google AI/Gemini y restricciones de la key. |
| `model_not_found` | Modelo no disponible para esa key/proyecto/endpoint. | Probar `GEMINI_MODEL=gemini-2.5-flash` o `gemini-1.5-flash`; verificar disponibilidad del modelo. |
| `quota` | Cuota agotada o rate limit. | Revisar cuotas, billing y limites por minuto/dia. |
| `request_failed` | Error HTTP no clasificado. | Revisar `error_body_redacted` en el diagnostico. |

## Seguridad

El provider y el comando:

- no loggean API keys,
- no guardan prompts completos,
- no guardan contexto completo,
- no guardan raw provider payload completo,
- truncan/redactan body de error.

## Lectura Del 403 Actual

Los evals anteriores muestran `status=403`. Con el provider actualizado, una nueva corrida deberia clasificarlo como `permission_denied` o `quota` segun el body redacted que devuelva Gemini.
