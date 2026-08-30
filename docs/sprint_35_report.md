# Sprint 35 Report - Manual Test Runbook v0.1

Fecha: 2026-07-05

## Objetivo

Crear un runbook final para que Armando pruebe Sofia Backend v0.1 manualmente de forma segura.

## Entregables

- `docs/manual_test_runbook_v0_1.md`
- `docs/v0_1_test_results_template.md`
- `docs/sprint_35_report.md`

## Contenido del Runbook

- Como arrancar backend con memory driver.
- Como arrancar backend con Firestore.
- Como importar la Postman collection.
- Orden recomendado de pruebas.
- Resultados esperados.
- Errores comunes.
- Como resetear datos locales.
- Que NO activar todavia.

## Decisiones QA

- La prueba principal recomendada usa `PERSISTENCE_DRIVER=memory`.
- Firestore queda como segunda pasada controlada.
- Gemini, FCM real, migraciones apply y delete executor quedan fuera de la prueba manual principal.
- El cierre de prueba usa una plantilla separada para capturar resultados sin secretos.

## Tests

Comando requerido:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: PASS.
