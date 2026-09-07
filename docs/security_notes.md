# Security Notes

## Google/Firebase credentials

Se detecto el archivo:

```text
sophia_ai_backend/sophia-ai-486821-b188dd536e22.json
```

No se uso en Sprint 0.5 y no debe ser consumido directamente desde el repositorio.

Acciones recomendadas:

- Mover credenciales reales fuera del repo.
- Rotar la credencial si ya fue compartida o versionada.
- Usar variables de entorno, secret manager o credenciales administradas por el entorno de despliegue.
- Mantener el patron `sophia_ai_backend/*.json` en `.gitignore` para evitar nuevos commits accidentales de credenciales.
