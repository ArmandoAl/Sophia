# Desarrollo desde el servidor (Sophia)

Verificado el 2026-10-26 en `hpmeserver` (Ubuntu 26.04) en la rama `migration/jaspr-server-20261026` (base `demo` @ `64de542`; `main` no se tocó).

## Estructura
```text
pubspec.yaml     workspace de Pub: packages/core, apps/mobile, apps/web   (pubspec.lock raíz)
apps/mobile/     app Flutter original (flutter_bloc, GetIt, GoRouter, Firebase, secure storage), movida con git mv; paquete `sophia_ai`
apps/web/        interfaz Jaspr: acceso y conversación (modo demo o API real) + pruebas de navegador en apps/web/e2e
packages/core/   `sophia_core`, Dart puro (bloc, equatable, http): modelos, ApiClient, repositorios (contratos e implementaciones),
                 DemoStore/repos demo, TokenStorage (interfaz), SessionCubit, AuthCubit, ChatMessageCubit
sophia_ai_backend/ sophia_db/ sophia_ai_design/   sin cambios (backend Go, base de datos, diseño)
```
Se movió al core sólo lo que no necesita Flutter: lo demás (widgets, router, DI con GetIt, `SecureTokenStorage`, Firebase Messaging)
sigue en `apps/mobile`. `registerDemoRepositories()` (GetIt) quedó en `apps/mobile/lib/demo/register_demo_repositories.dart`.
`packages/core/test/portability_test.dart` rechaza imports de Flutter/plugins y dependencias distintas de bloc/equatable/http.

## Versiones
Flutter **3.47.1** / Dart 3.13.1 (`.fvmrc`) · Jaspr CLI **0.23.4** (`dart pub global activate jaspr_cli 0.23.4`) · `build_web_compilers` 4.8.5 · `lints ^5`.

## Comandos verificados
```bash
source ~/server/infra/dev-env.sh
cd ~/projects/Sophia && fvm flutter pub get                 # una vez, en la raíz
(cd packages/core && fvm dart test)                          # 4 pruebas (portabilidad y conversación demo)
(cd apps/mobile && fvm flutter analyze && fvm flutter test)  # analyze limpio; 99 pasan y 2 FALLAN YA EN LA BASE (api_client_test: baseUrl por defecto)
(cd apps/web && fvm dart analyze)                            # limpio
scripts/preview_web.sh start                                 # jaspr build (demo) + servidor estático en 127.0.0.1:7703
(cd apps/web/e2e && npm install && npm test)                 # 5 pruebas de navegador (Chrome del sistema)
```
Los 2 fallos de `api_client_test.dart` esperan `http://localhost:8080` y el código apunta a Cloud Run: son previos (idénticos en `demo`).

## Preview
`scripts/preview_web.sh start|status|stop` — por defecto **modo demo** (`DEMO=1`, datos ficticios, sin red). Con la API real:
`DEMO=0 SOFIA_API_BASE_URL=https://… scripts/preview_web.sh start` (el backend debe permitir el origen del preview por CORS; **no probado**, no hay credenciales).
Los procesos no persisten tras reiniciar el servidor.

- Mac: `ssh -N -L 7703:127.0.0.1:7703 armando@100.121.183.5` y abrir `http://localhost:7703`.
- Teléfono (en la tailnet): `tailscale serve --bg --https=7703 http://127.0.0.1:7703`. **No está configurado** (cambia Tailscale del host).
- Un cambio sólo en un widget Flutter no aparece en la web, y la web no tiene hot reload de Flutter: tras editar, `scripts/preview_web.sh start` y recargar.

## Qué se prueba dónde
| En el servidor / navegador | Necesita credenciales, dispositivo real o Mac |
| --- | --- |
| Core, cubits, build Jaspr, flujo de la demo en navegador, accesibilidad claro/oscuro, 390 px | Backend real (login/CORS), Firebase Messaging, notificaciones, iOS/Android reales |

## Publicar para ver desde cualquier lugar (Firebase Hosting, canal de vista previa)
`scripts/deploy_preview.sh [canal]` compila `apps/web` en **modo demo** (datos ficticios, sin secretos) y lo sube a un canal de vista previa del
sitio `sophia-ai-demo` (proyecto `sophia-ai-60e87`), URL temporal de 7 días y `noindex`. No toca el sitio en producción (Flutter Web de la demo).
Requiere `firebase login --no-localhost` una vez. Configuración en `firebase.jaspr.json`.
