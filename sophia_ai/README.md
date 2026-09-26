# Sofía — App

Cliente Flutter de Sofía. Ver el [README del monorepo](../README.md) para la visión
completa del sistema.

**Flutter 3.47.1 (vía FVM) · Cubit · go_router · get_it**

---

## Arranque rápido

```bash
fvm flutter pub get
fvm flutter run --dart-define=SOFIA_API_BASE_URL=http://localhost:8080
```

Sin el `--dart-define`, la app usa el valor por defecto de
`lib/core/config/app_config.dart`.

### Análisis estático

```bash
fvm dart analyze lib
```

> **Usa `dart analyze`, no `flutter analyze`.** El segundo toma el lock de arranque global
> de Flutter: si tienes otro build corriendo en la máquina —aunque sea de otro proyecto—
> se queda esperando indefinidamente. `dart analyze` no lo toma y tarda un par de segundos.

### Si algo se cuelga

Casi siempre es una de estas dos:

1. **El SDK fijado no está instalado.** `.fvmrc` apunta a una versión concreta; si no la
   tienes, `fvm` intenta descargar Flutter entero. Compruébalo con `fvm list`.
2. **Otro proceso tiene el lock.** `ps aux | grep -E "flutter|xcodebuild"` — si hay un build
   de iOS corriendo, espera a que termine.

---

## Arquitectura

Clean Architecture por feature. Cada una repite la misma estructura:

```
lib/features/<feature>/
├── data/           implementación de repositorios contra ApiClient
├── domain/         entidades e interfaces de repositorio
└── presentation/
    ├── cubit/      estado
    └── pages/      pantallas
```

```
lib/core/
├── config/         configuración y feature flags
├── di/             service_locator (get_it)
├── models/         DTOs compartidos con el backend
├── network/        ApiClient, manejo de errores
├── router/         go_router
├── storage/        persistencia local (sesión)
├── theme/          tokens de diseño y primitivas de movimiento
└── widgets/        componentes compartidos
```

**La app no accede a Firestore.** No incluye `cloud_firestore` ni ningún SDK de Firebase
para datos: todo pasa por la API en Go con un JWT. Esa es la razón por la que las reglas de
Firestore pueden ser `deny-all` absolutas.

---

## Pantallas

| Pantalla | Qué hace |
|---|---|
| **Chat** | Conversación con Sofía y tarjetas de propuesta |
| **Dashboard** | Actividades y recordatorios del día |
| **Lo que Sofía sabe de ti** | Creencias, con opción de corregir o retirar |
| **Contextos** | Personas y modos que cambian su comportamiento |
| **Ajustes del asistente** | Nombre, estilo, proactividad, autonomía |
| **Perfil** | Nombre, zona horaria, idioma |
| **Diagnóstico** | Salud del backend y estado del aprendizaje |
| Recordatorios · Insights · Auth · Onboarding | |

### La tarjeta de propuesta

Es el componente más importante de la app: donde el usuario toma la decisión que alimenta
todo el aprendizaje del sistema. Ofrece **tres caminos**, no dos:

| Acción | Envía | Valor |
|---|---|---|
| **Hacerlo** | `approved_direct` | Acertó |
| **Ajustar** | `approved_corrected` + el input corregido | Acertó la intención, falló el detalle |
| **No** | `rejected` + motivo | Falló, y por qué |

«Ajustar» abre el `proposed_input` como formulario editable. El backend compara ambos JSON
y guarda **qué campos cambiaron**, que es una señal mucho más rica que un sí o un no.

Las propuestas se confirman **una a una**, nunca en lote: si Sofía propone tres cosas y dos
son buenas, no tiene sentido obligar a rechazar las tres.

También se mide el tiempo entre que la tarjeta aparece y el usuario decide
(`decision_latency_ms`): es una señal de confianza barata — decidir en dos segundos no es
lo mismo que decidir en cuatro minutos.

### Lo que Sofía sabe de ti

Pantalla emocionalmente sensible: muestra lo que una IA cree saber del usuario. El diseño
tiene que leerse como *«esto es tuyo, corrígeme»* y no como un veredicto.

- El nivel de confianza se traduce a lenguaje humano — «porque lo decidiste», «porque lo
  mencionaste», «deducido de tus conversaciones» — nunca se muestra `trust_tier: 3`.
- Las creencias que están en el prompt base se distinguen: son las que gobiernan el
  comportamiento hoy.
- Retirar una creencia no la borra: queda retirada y deja de contar.

Corregir una creencia equivocada es el canal de aprendizaje de mayor densidad que tiene el
sistema. Una corrección vale más que veinte inferencias.

---

## Feature flags

`lib/core/config/feature_flags.dart` controla qué se muestra:

| Flag | Estado |
|---|---|
| `aiActionExecutionEnabled` | ✅ el backend soporta confirm/reject/execute reales |
| `notificationsEnabled` | registro de dispositivo FCM |
| `privacyDeleteRequestEnabled` | exportación y solicitud de borrado |
| `smartHomeEnabled` | ❌ fuera de alcance, sin backend |

Un flag que promete algo que el backend no soporta es peor que no tener la función: activa
uno solo cuando el endpoint exista y esté probado.

---

## Sistema de diseño

`lib/core/theme/` centraliza tokens y movimiento. Dos reglas:

- **Ningún color literal fuera de los tokens.** Todo sale de tokens semánticos, en tema
  claro y oscuro.
- **Ninguna duración de animación fuera de `motion.dart`.** Las duraciones y curvas se
  definen una vez y se consumen desde ahí.

Principios de movimiento: interrumpible (una animación se cancela sin saltos si el usuario
toca), continuidad (los elementos que persisten se transforman, no reaparecen), deferencia
(si una animación se nota como animación, sobra) y respeto por `disableAnimations`.

---

## Plataformas

iOS, Android y macOS. La configuración de iOS usa Swift Package Manager tras la migración
a UIScene. El bundle ID es `com.aial.sophiav1`.

---

## Convenciones

- Un Cubit por pantalla; el estado es inmutable con `copyWith`.
- Los repositorios se declaran en `domain/` y se implementan en `data/`.
- Los errores de red se traducen a mensajes accionables, nunca se muestra el error crudo.
- `fvm dart analyze lib` sin issues antes de cada commit.
