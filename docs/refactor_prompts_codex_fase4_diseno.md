# Sofía — Fase 4: configuración pendiente y rediseño de la app

Continúa desde `refactor_prompts_codex_fase3.md` (pasos 14-16, verificados).

Orden: **17 → 18 → 19 → 20 → 21**. El 17 es funcional; del 18 al 21 son el rediseño y
**deben correrse en ese orden**: el 18 crea el sistema del que dependen los demás.

Usa el mismo bloque de **Contexto común** de la fase 2 en todos.

---

## Paso 17 — Cerrar los huecos de configuración

```
La pantalla de ajustes ya tiene Perfil, Ajustes del asistente, "Lo que Sofía sabe de ti",
Contextos y Diagnóstico. Faltan tres secciones cuyos endpoints ya existen en el backend
pero no tienen interfaz.

1. PRIVACIDAD Y DATOS — nueva pantalla en lib/features/settings/
   Consume endpoints que ya existen:
     GET  /users/me/export          → descarga todos tus datos
     POST /users/me/delete-request  → solicita el borrado de la cuenta
   - "Exportar mis datos" descarga el JSON y lo comparte con el share sheet del sistema.
   - "Solicitar borrado" pide confirmación escrita (que el usuario teclee BORRAR) y explica
     en una línea que es una solicitud, no un borrado inmediato.
   - Quita el gate de privacyDeleteRequestEnabled en feature_flags.dart: el backend ya lo
     soporta de verdad.

2. NOTIFICACIONES — nueva pantalla
   Consume:
     POST   /notifications/device-tokens
     DELETE /notifications/device-tokens/{id}
   - Interruptor para registrar o quitar este dispositivo.
   - Lista de dispositivos registrados con su fecha de alta.
   - notificationsEnabled sigue siendo un flag: déjalo, pero que la pantalla exista.

3. IMPORTAR CONVERSACIONES — nueva pantalla, con fricción deliberada
   Consume:
     POST   /ingestion/conversations
     DELETE /ingestion/batches/{id}
   - Lista de lotes importados con su estado y cuántas creencias generó cada uno.
   - Botón para deshacer un lote entero, con confirmación.
   - ANTES de permitir importar, una pantalla de advertencia que el usuario debe leer:
     que importar procesa mensajes de otras personas que no dieron su consentimiento, y
     que lo importado entra como creencia de nivel 3 con confianza máxima de 0.5 hasta
     que una decisión suya lo corrobore.
   - No construyas el selector de archivos todavía: por ahora basta con listar lotes y
     poder deshacerlos. La importación se hará por API hasta que exista el conversor.

4. Todas cuelgan de settings_page.dart. Al terminar, `fvm dart analyze lib` sin issues.
```

---

## Paso 18 — El sistema de diseño y de movimiento

> **Este paso no cambia ninguna pantalla.** Crea los cimientos. Los pasos 19-21 los usan.
> Si Codex intenta rediseñar pantallas aquí, párale.

```
Actúa como director de diseño de producto con diez años construyendo interfaces para
aplicaciones de inteligencia artificial aplicadas a la salud mental y el acompañamiento
personal. Has trabajado en equipos donde el detalle de una animación se discute durante
días. Tu referencia de calidad es el software de Apple: no por su estética, sino por su
disciplina — cada transición tiene una razón, ninguna llama la atención sobre sí misma.

CONTEXTO DEL PRODUCTO, léelo antes de decidir nada:
Sofía es un asistente personal que aprende de su usuario hasta poder actuar en su nombre.
Guarda creencias sobre él: cómo trabaja, con quién se comporta distinto, qué le cuesta.
El usuario puede ver esas creencias y corregirlas.

Eso impone un carácter emocional muy concreto, y define lo que NO debe hacer el diseño:
- Nada de gamificación. Ni rachas, ni insignias, ni celebraciones. Una app que sabe cosas
  íntimas de ti y te felicita con confeti es una app en la que no confías.
- Nada de patrones de captura de atención. Sin rojos de urgencia, sin badges que reclaman,
  sin animaciones que piden ser miradas.
- La calma ES la característica. La contención se lee como respeto por la intimidad.
- La transparencia se diseña: cuando Sofía muestra lo que cree saber de ti, la interfaz
  debe transmitir "esto es tuyo, puedes cambiarlo", no "esto es el veredicto del sistema".

TAREA: crea el sistema de diseño y movimiento. NO toques ninguna pantalla existente.

1. lib/core/theme/design_tokens.dart
   - Escala tipográfica completa, con una razón entre pasos coherente. Define los roles
     (display, título, cuerpo, etiqueta, dato) y sus tamaños, pesos, altura de línea y
     letter-spacing. Los datos numéricos usan cifras tabulares.
   - Paleta como tokens semánticos, nunca colores sueltos: superficie, superficie elevada,
     tinta, tinta suave, apagado, línea, acento, y tres colores de estado (positivo,
     atención, crítico) que NO son el acento.
   - Escala de espaciado con una sola progresión. Todo el layout sale de ahí.
   - Radios y elevaciones: pocos valores, cada uno con un uso definido. No pongas el mismo
     radio y la misma sombra en todo — eso aplana la jerarquía.
   - Tema claro y oscuro completos. El oscuro no es el claro invertido: recalcula el
     contraste para que el acento siga funcionando sobre fondo oscuro.

2. lib/core/theme/motion.dart — el corazón de este paso
   Define las primitivas de movimiento como constantes con nombre, para que ninguna
   pantalla invente sus propias duraciones:

   - Duraciones: micro (~120ms, feedback táctil), corta (~220ms, cambios de estado),
     media (~320ms, transiciones de página), larga (~480ms, solo para lo que de verdad
     necesita respirar). Nada supera los 500ms. Una animación lenta se siente como una
     app lenta.
   - Curvas: usa física de resorte, no easing lineal. Define al menos tres resortes —
     uno muy amortiguado para la estructura (sin rebote visible), uno medio para
     contenido, uno con ligero rebote SOLO para elementos que responden al dedo.
     En Flutter esto es SpringDescription con CurvedAnimation o el paquete de animación.
   - Escalonado: cuando entren varios elementos hermanos, entre 30 y 50ms de diferencia
     entre ellos. Más que eso se percibe como lentitud.

3. Instala flutter_animate y añádelo a pubspec.yaml. Es la única dependencia de animación
   que necesitas ahora. NO instales rive ni lottie: requieren archivos de assets que no
   tenemos y que Codex no puede crear.

4. lib/core/widgets/motion/ — envoltorios reutilizables sobre flutter_animate:
   - Un widget de entrada escalonada para listas.
   - Un widget de transición de estado que cruza contenido sin saltos de layout.
   - Un botón con respuesta táctil: escala ligeramente al presionar (no más de 0.97),
     con retorno por resorte, y HapticFeedback.lightImpact en el toque.

5. PRINCIPIOS que deben quedar escritos como comentario de cabecera en motion.dart, y que
   los pasos siguientes deben respetar:
   - INTERRUMPIBLE: toda animación debe poder cancelarse a media ejecución sin saltos.
     Si el usuario toca durante una transición, la interfaz responde de inmediato.
   - CONTINUIDAD: un elemento que existe antes y después de una transición se transforma
     (Hero), no desaparece y reaparece.
   - DEFERENCIA: el movimiento sirve al contenido. Si una animación se nota como
     animación, sobra.
   - ORIGEN: las cosas aparecen desde donde tiene sentido. Un menú desde su botón, una
     hoja desde abajo, un detalle desde la fila que lo abrió.

6. Respeta prefers-reduced-motion: si MediaQuery.disableAnimations está activo, todas las
   duraciones caen a cero. Impleméntalo en un solo sitio, dentro de motion.dart, no
   pantalla por pantalla.

7. Al terminar: `fvm dart analyze lib` sin issues, y dime qué archivos creaste. NO debes
   haber modificado ninguna pantalla.
```

---

## Paso 19 — Navegación y transiciones

```
Mismo rol de director de diseño del paso 18. Usa exclusivamente los tokens y las
primitivas de motion.dart que creaste allí. Si necesitas una duración o una curva que no
existe, añádela a motion.dart primero; nunca la escribas suelta en una pantalla.

TAREA: la navegación. Es lo que más se percibe y donde más se nota el trabajo.

1. TRANSICIONES DE PÁGINA en lib/core/router/app_router.dart
   go_router ya está en uso. Sustituye las transiciones por defecto por CustomTransitionPage
   con dos comportamientos distintos y deliberados:

   - Navegación JERÁRQUICA (entrar a un detalle desde una lista): la pantalla nueva entra
     desde la derecha con un desplazamiento corto — no la pantalla entera, unos 30 píxeles
     — combinado con opacidad. La pantalla anterior se desplaza ligeramente en sentido
     contrario y se oscurece un poco. Eso crea profundidad sin usar sombras.

   - Navegación LATERAL (cambiar de pestaña en la barra inferior): NUNCA un deslizamiento
     horizontal. Eso es lenguaje de Material y se siente barato. Usa un cruce de opacidad
     con una escala mínima: la saliente baja a 0.98 mientras se desvanece, la entrante
     sube desde 1.02. Duración corta. El resultado debe sentirse instantáneo pero no seco.

2. BARRA INFERIOR — aquí está el detalle que pediste
   - El icono seleccionado escala ligeramente (no más de 1.1) con el resorte medio, y su
     color pasa al acento. La transición de color y la de escala tienen que empezar juntas.
   - Un indicador que se DESPLAZA entre pestañas en lugar de aparecer y desaparecer. Que
     el usuario vea de dónde a dónde va. Usa el resorte muy amortiguado.
   - La etiqueta de la pestaña activa entra con opacidad, las inactivas se atenúan.
   - HapticFeedback.selectionClick en cada cambio de pestaña. Es lo que hace que se sienta
     físico.
   - Si se toca la pestaña ya activa, la lista de esa pestaña sube al inicio con animación.
     Es un detalle de iOS que nadie pide y todo el mundo agradece.

3. CONTINUIDAD entre lista y detalle
   Usa Hero para los elementos que persisten. Casos concretos:
   - Del chat a una propuesta: la tarjeta se expande desde su posición en la lista.
   - De la lista de creencias a su edición: el enunciado se mantiene y crece.
   - De contextos a editar contexto: la etiqueta persiste.
   Cuida que el Hero no arrastre el fondo ni provoque parpadeos de tema.

4. HOJAS MODALES
   Cuando se abra una hoja (editar creencia, motivo de rechazo, ajustar propuesta):
   - Sube desde abajo con el resorte medio.
   - La pantalla de detrás baja a 0.94 de escala y se oscurece: da la sensación de que
     retrocede en profundidad.
   - Se puede cerrar arrastrando, con la resistencia elástica al llegar al límite.

5. ESTADOS DE CARGA
   Nada de spinners centrados. Usa esqueletos de contenido que ocupen exactamente el
   espacio del contenido real, con un brillo muy sutil que recorra la superficie. Cuando
   llegan los datos, el esqueleto se cruza con el contenido — no salta.

6. `fvm dart analyze lib` sin issues. Dime qué archivos tocaste.
```

---

## Paso 20 — Rediseño de las pantallas núcleo

```
Mismo rol. Usa los tokens del paso 18 y las transiciones del 19. No inventes valores.

TAREA: rediseña chat y dashboard. Son las que el usuario ve todos los días.

1. CHAT (lib/features/chat/presentation/pages/chat_page.dart)
   - Los mensajes entran desde abajo con un desplazamiento corto y opacidad, con el
     resorte medio. El del usuario y el de Sofía entran distinto: el suyo más directo, el
     de ella con una fracción de retraso, como si estuviera formulándolo.
   - El indicador de que Sofía está pensando NO debe ser tres puntos rebotando. Diseña algo
     más contenido y propio del producto: una respiración lenta, un pulso muy tenue. Que
     comunique proceso, no espera.
   - LA TARJETA DE PROPUESTA es el elemento más importante de la app: es donde el usuario
     toma la decisión que alimenta todo el aprendizaje. Merece el trabajo más fino:
       · Entra con más presencia que un mensaje normal, pero sin estridencia.
       · Los tres botones (No / Ajustar / Hacerlo) tienen peso visual distinto y honesto:
         "Hacerlo" es el primario, "Ajustar" claramente disponible, "No" discreto pero
         nunca escondido. Esconder el rechazo sería un patrón oscuro.
       · Al decidir, la tarjeta confirma con una transición breve a un estado resuelto y
         se asienta en el hilo. No desaparece: el usuario debe poder ver qué decidió.
       · Haptics distintos por acción: impacto medio al confirmar, ligero al rechazar.
   - El campo de escritura crece con el texto de forma fluida, sin saltos de altura.

2. DASHBOARD (lib/features/dashboard/presentation/pages/dashboard_page.dart)
   - Entrada escalonada de las secciones al cargar, entre 30 y 50ms entre ellas.
   - Jerarquía real: lo que requiere atención hoy arriba y con más peso; el resto se
     consulta, no se impone. No pongas todo en tarjetas iguales — eso aplana la jerarquía
     y obliga a leerlo todo.
   - Si muestras números (actividades, recordatorios), usa cifras tabulares y anima el
     cambio de valor con un conteo corto, no con un salto.

3. Aplica el botón con respuesta táctil del paso 18 a TODOS los elementos pulsables de
   ambas pantallas. La consistencia del feedback es lo que hace que una app se sienta
   sólida.

4. `fvm dart analyze lib` sin issues.
```

---

## Paso 21 — Rediseño de las pantallas de configuración y aprendizaje

```
Mismo rol. Cierra el rediseño.

TAREA: ajustes, creencias, contextos y diagnóstico. Aquí el reto no es el movimiento sino
la sensibilidad: son las pantallas donde el usuario ve lo que una IA cree saber de él.

1. "LO QUE SOFÍA SABE DE TI" (lib/features/beliefs/)
   Esta pantalla es emocionalmente delicada. Puede leerse como "aquí está tu perfil
   psicológico" y provocar rechazo, o como "aquí está lo que he entendido, corrígeme".
   Tiene que ser lo segundo, y eso se consigue con diseño:
   - El tono de la interfaz es tentativo, no declarativo. Trabaja los textos: "Creo que
     prefieres..." funciona mejor que "Preferencia: ...".
   - La confianza se muestra de forma legible sin exponer el número crudo. Un indicador
     visual continuo, y el porcentaje solo si el usuario lo busca.
   - Traduce el nivel de confianza a lenguaje humano, nunca muestres "trust_tier: 3":
     "porque lo decidiste" / "porque lo mencionaste" / "deducido de tus conversaciones".
   - Las creencias que están en el prompt base (prompt_slot core) se distinguen con
     claridad: son las que gobiernan su comportamiento hoy, y el usuario debe poder verlo.
   - Retirar una creencia debe sentirse tranquilo y reversible, no destructivo. Nada de
     rojo de alerta ni de animaciones violentas: se atenúa y sale con suavidad.
   - Entrada escalonada agrupada por categoría.

2. CONTEXTOS (lib/features/contexts/)
   - Explica de una línea para qué sirven, sin jerga.
   - Distingue visualmente persona de modo, sin caer en estereotipos ni en iconografía de
     género.
   - Crear un contexto es un formulario breve y agradable; el slug se genera solo desde la
     etiqueta pero se puede editar.

3. AJUSTES (lib/features/settings/) y las tres pantallas nuevas del paso 17
   - Agrupa por afinidad con encabezados de sección, no una lista plana de filas iguales.
   - Los interruptores animan su estado con el resorte, con haptic ligero.
   - La sección de privacidad y datos se lee con calma: es donde el usuario decide sobre
     su propia información y merece espacio, no filas apretadas.

4. DIAGNÓSTICO (lib/features/system/)
   - La sección de estado del aprendizaje que se añadió en el paso 16 merece más que
     texto plano: los últimos 7 días como una serie visual pequeña donde se vea si el
     worker corrió y cuánto costó.
   - Si hay una gráfica, que sus etiquetas salgan de los tokens del tema para que sea
     legible en claro y oscuro, y que ningún texto se salga de su caja.

5. REVISIÓN FINAL de coherencia en toda la app:
   - Ningún color literal fuera de los tokens.
   - Ninguna duración de animación fuera de motion.dart.
   - Todo elemento pulsable tiene el mismo feedback táctil.
   - Todo elemento enfocable tiene un estado de foco visible (accesibilidad).
   - Contraste suficiente en ambos temas.

6. `fvm dart analyze lib` sin issues. Al terminar, dime qué quedó inconsistente y no
   pudiste arreglar dentro del alcance de estos cuatro pasos.
```

---

# Notas sobre el rediseño

**Por qué cuatro pasos y no uno.** Un rediseño completo en un solo prompt produce una app
sin sistema: cada pantalla inventa sus colores y sus tiempos, y el resultado se ve
inconsistente por mucho esfuerzo que haya en cada parte. El paso 18 crea la fuente de
verdad; los tres siguientes solo la consumen. Es la diferencia entre una app rediseñada y
una app con un sistema de diseño.

**Sobre las librerías de animación.** Solo `flutter_animate`. Es declarativa, se integra
con el widget tree y no obliga a gestionar controladores a mano. Descarté a propósito:

- **Rive** y **Lottie** son las herramientas correctas para iconos vectoriales animados,
  pero ambas necesitan archivos de assets (`.riv`, `.json`) que se crean en un editor
  externo. Codex no puede generarlos, y sin los assets la dependencia queda muerta. Si
  algún día tienes esos archivos, añadir Rive es un cambio pequeño y aislado.
- Para animar iconos sin assets, Flutter trae `AnimatedIcon` con un juego limitado, y un
  `CustomPainter` puede interpolar formas. Los pasos 19-21 lo usan donde hace falta.

**Lo que no está en estos prompts a propósito.** Ningún cambio de lógica de negocio, de
llamadas al backend ni de modelos. El rediseño no debe tocar `data/` ni `domain/`: si un
prompt te devuelve cambios ahí, revísalos con cuidado, porque probablemente se salió del
alcance.
