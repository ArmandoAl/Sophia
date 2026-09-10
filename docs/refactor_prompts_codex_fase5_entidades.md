# Sofía — Fase 5: memoria de entidades y contexto simbiótico

Continúa desde la fase 4. **Corre esta fase después del rediseño**, no en paralelo: el
paso 27 toca pantallas que la fase 4 rediseña.

Orden: **22 → 23 → 24 → 25 → 26 → 27**. Cada uno depende del anterior.

Usa el mismo bloque de **Contexto común** de la fase 2 en todos.

---

## Objetivo de la fase

Que Sofía sepa de las personas del círculo del usuario, no solo del usuario. Que reconozca
de quién se habla, cambie de contexto sola, recuerde conversaciones pasadas sobre esa
persona, y pueda retomar hilos abiertos.

**Escenario de referencia**, úsalo para juzgar si una decisión de diseño sirve:

> «Cortaron a mi hermana, vamos a casa y nos vendría bien algo que nos anime.»
>
> Sofía debe resolver *hermana* → Diana, saber que le gusta Mon Laferte, saber que la
> ruptura es reciente, recordar que hace dos semanas hablaron de que estaba mal, y saber
> que con Diana el usuario es más protector. Todo eso en menos de 600 tokens.

**Presupuesto de contexto tras esta fase: ~5.000 tokens por petición.** Sube desde 4.600.
Si un paso te lleva por encima, recorta la recuperación, nunca subas el techo.

---

## Paso 22 — Entidades y hechos sobre ellas

```
Hoy el sistema solo modela creencias sobre el usuario. Necesita modelar también hechos
sobre las personas de su círculo. La estructura base ya existe y hay que extenderla, NO
crear un módulo nuevo.

PARTE A — El registro de entidades

learning/domain UserContext ya es el registro de entidades (Kind, Slug, Label, Aliases).
Extiéndelo:

1. Kind admite ahora tres valores: "person" | "mode" | "group".
   - "group" es para conjuntos donde no merece la pena una ficha individual: familia
     extendida, primos, el equipo de trabajo. Sus alias incluyen los nombres de sus
     miembros, de modo que mencionar a un primo resuelve al grupo.
2. Campos nuevos:
     Relationship string  // "hermana", "amiga", "pareja", "madre", "primos"
     CreatedFrom   string // "explicit" | "inferred"
     Status        string // "active" | "pending_review" | "merged" | "archived"
     MergedInto    string // si se fusionó con otra entidad
   Una entidad nunca se borra: se archiva o se fusiona.
3. Un método Merge(target) que reasigna los hechos de una entidad a otra y la marca como
   merged. Es lo que permite corregir cuando el sistema creó dos entidades para la misma
   persona.

PARTE B — Hechos sobre entidades

NO crees una colección nueva. Extiende learning/domain Belief, que ya tiene decaimiento,
contradicciones, niveles de confianza, embeddings, deduplicación y cadenas de supersede.
Reaprovechar toda esa maquinaria es el motivo de esta decisión.

4. Campos nuevos en Belief:
     SubjectType string     // "user" (por defecto) | "entity"
     SubjectID   string     // "" para user; el ID de la entidad si es entity
     FactKind    string     // "trait" | "state" | "relationship"
     ValidUntil  *time.Time // solo para state
     FollowUpAt  *time.Time // cuándo conviene retomar el tema
   Los registros existentes sin SubjectType se tratan como "user": retrocompatible, sin
   migración.

5. LOS TRES TIPOS DE HECHO SE COMPORTAN DISTINTO. Esto es lo más importante del paso:

   - trait  → estable. "Le gusta Mon Laferte", "odia el cilantro". NO DECAE. Un gusto no
              se desvanece en 90 días. Sobrescribe el decaimiento para este tipo.
   - state  → temporal. "Acaban de cortarla", "busca trabajo", "está enferma". Requiere
              ValidUntil obligatorio. Al vencer NO se borra: pasa a status "archived" y
              deja de cargarse. Un state sin ValidUntil debe rechazarse en Validate().
   - relationship → estable, describe el vínculo con el usuario. No decae.

   Sin esta distinción, dentro de un año Sofía seguirá tratando a la hermana como recién
   soltera. Escríbelo como comentario en el código.

6. Recuperación por entidad, con presupuesto:
     GetEntityContext(ctx, userID, entityID, tokenBudget) (EntityContext, error)
   Orden de carga, y este orden importa:
     a) states vigentes  — son la novedad, lo más relevante para la conversación de hoy
     b) relationship
     c) traits ordenados por beliefValue
     d) creencias del USUARIO con scope person:<slug> — cómo es él con esa persona
   Corta al llegar al presupuesto. Techo por defecto: 400 tokens (constante nombrada).

7. UpsertBelief debe deduplicar dentro del mismo (SubjectType, SubjectID, FactKind).
   "Le gusta Mon Laferte" sobre Diana y sobre Samira son hechos distintos.

8. Índices nuevos en firestore.indexes.json:
     user_beliefs (user_id ASC, subject_type ASC, subject_id ASC, status ASC, confidence DESC)
     user_beliefs (user_id ASC, subject_id ASC, fact_kind ASC, valid_until ASC)
     user_contexts (user_id ASC, kind ASC, status ASC)

9. Un job de mantenimiento dentro del worker de síntesis que archive los states vencidos.
   Barato: una consulta por valid_until < now.

10. Tests: un trait no pierde confianza al pasar 200 días simulados; un state sin
    ValidUntil no valida; un state vencido no aparece en GetEntityContext; el presupuesto
    de 400 tokens nunca se supera; dos entidades distintas no comparten hechos.
```

---

## Paso 23 — Reconocer de quién se habla, y crear la ficha sola

```
Ahora que las entidades pueden guardar hechos, hay que reconocerlas en la conversación y
crearlas sin que el usuario las dé de alta a mano.

PARTE A — Detección por alias (barata, en el camino caliente)

1. En internal/learning, un EntityResolver:
     Resolve(ctx, userID, message) ([]EntityMatch, error)
   - Normaliza el mensaje (minúsculas, sin acentos) y busca coincidencias contra Slug,
     Label y Aliases de las entidades activas del usuario.
   - Cachea el catálogo de entidades en memoria por usuario con TTL corto: son menos de
     15 entidades, no tiene sentido ir a Firestore en cada mensaje.
   - Coincide por palabra completa, nunca por subcadena: "ana" no debe activar a "Diana".
   - Devuelve las coincidencias con su posición en el texto.
   SIN llamadas al modelo. Esto tiene que ser gratis.

2. AMBIGÜEDAD: si dos entidades coinciden con la misma mención (dos Dianas), NO ADIVINES.
   Incluye ambas candidatas en el contexto marcadas como ambiguas e instruye al modelo,
   en el prompt, a preguntar cuál de las dos antes de dar por hecho nada.
   Cargar el contexto de la persona equivocada es peor que hacer una pregunta.

PARTE B — Creación automática

3. Colección nueva entity_candidates:
     user_id, name, normalized_name, mention_count, first_seen_at, last_seen_at,
     sample_contexts []string  (fragmentos cortos donde apareció, máx. 3, truncados)
4. En el camino caliente, tras resolver entidades conocidas, detecta nombres propios que
   NO estén en el catálogo, con heurística barata: palabra capitalizada que no inicia
   frase, que no es una stopword ni un mes ni un día. SIN modelo.
   Incrementa mention_count del candidato.
5. En el worker de síntesis nocturna, promueve a entidad los candidatos con
   mention_count >= 3 (constante configurable ENTITY_PROMOTION_THRESHOLD):
   - Crea el UserContext con Kind "person", CreatedFrom "inferred",
     Status "pending_review".
   - Una entidad pending_review SÍ acumula hechos y SÍ se carga en contexto: no queremos
     fricción. El review es para corregir, no para autorizar.
   - Los candidatos que no llegan al umbral en 30 días se descartan.
   El umbral de 3 mantiene fuera los nombres sueltos que se mencionan una vez.

6. Endpoints:
     GET    /entities                    lista, con filtro por kind y status
     POST   /entities                    alta manual
     GET    /entities/{id}               ficha con sus hechos
     PATCH  /entities/{id}               label, relationship, aliases, kind, status
     POST   /entities/{id}/merge         body: {"into": "<id>"}
     DELETE /entities/{id}               archiva, no borra
     GET    /entities/pending            las creadas por inferencia sin revisar
   Todo detrás de r.Authn y filtrando por el usuario del token.

7. Tests: "ana" no activa "Diana"; dos entidades homónimas marcan ambigüedad; un candidato
   con 2 menciones no se promueve y con 3 sí; fusionar reasigna todos los hechos.
```

---

## Paso 24 — Cambio de contexto y resumen de arrastre

```
Cuando la conversación vira hacia otra persona, Sofía debe cambiar de contexto llevándose
lo que ya se había hablado. Es lo que hace que no se sienta como empezar de cero.

1. En el ContextBuilder, después de resolver entidades:
   - Si hay una entidad detectada distinta de la activa, es un CAMBIO DE CONTEXTO.
   - Umbral: una sola mención de pasada NO debería virar la conversación. Cambia solo si
     la entidad se menciona en el mensaje actual Y (aparece más de una vez, o el mensaje
     es una pregunta o petición dirigida a ese tema). Una heurística simple basta; el
     objetivo es no cambiar por un "ayer vi a mi mamá" suelto.

2. RESUMEN DE ARRASTRE. Al cambiar de contexto:
   - Comprime los turnos previos de la conversación en 2 o 3 frases con una llamada al
     modelo usando Task = "extract" (el modelo barato del paso 10).
   - SOLO en el momento del cambio, nunca por mensaje.
   - Guárdalo en la conversación (campo nuevo carry_forward + carry_forward_entity) para
     no recalcularlo si el usuario sigue en el mismo contexto.
   - Va al SUFIJO del prompt, con la etiqueta "de qué veníamos hablando".

3. Estructura del sufijo tras este paso, respetando el presupuesto:
     carry_forward      ~200 tokens
     active_entity      ~400   (GetEntityContext del paso 22)
     conversation_history ~1000 (bájalo de 1250: el arrastre cubre lo anterior)
     recent_state        ~600
     user_message        ~150
   El PREFIJO no cambia y sigue siendo byte-idéntico. Los hechos de entidad NUNCA van al
   prefijo: invalidarían el caché en cada cambio de contexto y se perdería ~45% del
   ahorro. Hay un test de la fase 2 que lo vigila; asegúrate de que sigue pasando.

4. El campo ActiveContext del RuntimeRequest sigue mandando: si el cliente lo envía
   explícitamente, gana sobre la detección automática. La detección es un respaldo, no un
   sustituto de la voluntad del usuario.

5. Devuelve en la respuesta HTTP qué entidad quedó activa y si hubo cambio, para que la
   app pueda mostrarlo. El usuario debe poder ver de qué contexto está hablando Sofía —
   un cambio silencioso y equivocado es confuso.

6. Tests: una mención de pasada no cambia el contexto; una pregunta dirigida sí; el
   arrastre se genera una vez y se reutiliza; el prefijo sigue byte-idéntico entre dos
   peticiones con entidades distintas; el sufijo nunca supera su presupuesto.
```

---

## Paso 25 — Memoria episódica

```
Sofía debe recordar conversaciones concretas, no solo generalizaciones. "Hace dos semanas
hablamos de que Diana estaba mal" es un episodio; "Diana tiende a agobiarse" es una
creencia. Ambas cosas hacen falta y son distintas.

REGLA QUE SE MANTIENE: el worker de síntesis (learning/worker) sigue sin leer
conversaciones crudas. Lo que construyes aquí es un PASO SEPARADO que sí las lee pero que
produce EPISODIOS, nunca creencias. Los episodios no entran jamás al prompt base: se
recuperan bajo demanda. Así la regla sigue protegiendo lo que tiene que proteger.
Escríbelo como comentario de cabecera.

1. Colección episodes:
     user_id, occurred_at
     summary       string    // 1-2 frases
     entity_ids    []string  // entidades mencionadas
     topics        []string
     salience      float64   // 0..1
     source        string    // "conversation" | "ingestion"
     conversation_id string
     search_terms  []string
     embedding     []float32

2. Extractor de episodios, worker aparte o fase del de ingesta:
   - Procesa las conversaciones del día que MENCIONAN al menos una entidad. Si no hay
     entidad mencionada, no hay episodio: ese es el filtro principal contra el ruido.
   - Una llamada con Task = "extract" por conversación, no por mensaje.
   - Salida estructurada validada por JSON-Schema.
   - SALIENCIA: el modelo la estima, pero además impón un mínimo (constante
     EPISODE_MIN_SALIENCE, por defecto 0.4). Guardar todo es volver al problema original.
   - Máximo de episodios por día y por usuario, configurable. Un día muy hablado no debe
     generar cincuenta.

3. Recuperación:
     SearchEpisodes(ctx, userID, entityID, terms, limit) ([]*Episode, error)
   - Filtra por entity_ids con array-contains, ordena por recencia y saliencia,
     y refina por coseno cuando hay embeddings.
   - Devuelve como mucho 3-5, con techo de ~300 tokens.

4. En el ContextBuilder: cuando hay entidad activa, añade al sufijo un bloque de episodios
   recientes sobre ella. Si el presupuesto está justo, los episodios ceden antes que los
   hechos de entidad: un hecho vigente pesa más que un recuerdo.

5. Retención: los episodios con saliencia baja se archivan a los 180 días. Los de
   saliencia alta se conservan. Job de mantenimiento junto al de states vencidos.

6. Tests: una conversación sin entidades no genera episodio; una por debajo del mínimo de
   saliencia tampoco; la recuperación por entidad devuelve los suyos y ninguno más; el
   bloque nunca supera su presupuesto; los episodios no aparecen en el prompt base.
```

---

## Paso 26 — Hilos abiertos y asociación proactiva

```
Que Sofía pueda retomar temas por iniciativa propia cuando la conversación lo permita:
"hace tres semanas me dijiste que Samira pensaba cambiar de trabajo, ¿cómo le fue?"

Esto es lo que separa un asistente que responde de uno que acompaña. También es lo que
puede sentirse invasivo si se hace mal, así que las restricciones importan tanto como la
capacidad.

1. El campo FollowUpAt del paso 22 se usa aquí. Se rellena cuando:
   - Se crea un state con ValidUntil → FollowUpAt a mitad de su vigencia.
   - El extractor de episodios detecta algo sin resolver (una decisión pendiente, un
     evento futuro) → FollowUpAt en la fecha que corresponda.

2. En el ContextBuilder añade un bloque "hilos abiertos" al sufijo:
     GetOpenThreads(ctx, userID, limit) ([]OpenThread, error)
   - Hechos y episodios con FollowUpAt <= ahora, no retomados aún.
   - Máximo 3, techo de ~100 tokens.
   - Marca un hilo como retomado cuando aparece en una conversación, para no repetirlo.

3. En el prompt, los hilos abiertos se presentan como OPORTUNIDADES, no como órdenes.
   Instrucción explícita al modelo: puede mencionarlos si encaja de forma natural en la
   conversación, y debe ignorarlos si el usuario está en otra cosa. Nunca debe forzar un
   hilo abierto ni abrir la conversación con él si el usuario preguntó otra cosa.

4. LÍMITES, no negociables:
   - Como mucho un hilo retomado por conversación. Sofía no es un interrogatorio.
   - Nada de hilos sobre states sensibles marcados como tales. Añade un campo Sensitive
     bool a Belief: los hechos marcados así se cargan como contexto pero NUNCA se usan
     para iniciar un tema. Salud, dinero y conflictos entran ahí por defecto.
   - El usuario puede silenciar los hilos de una entidad concreta.

5. Tests: un hilo con FollowUpAt futuro no aparece; uno vencido sí; uno marcado sensible
   nunca aparece como hilo; solo se ofrece uno por conversación; un hilo retomado no se
   vuelve a ofrecer.
```

---

## Paso 27 — Las personas en la app

```
Depende del rediseño de la fase 4: usa sus tokens y sus primitivas de movimiento.

1. Nueva feature lib/features/entities/
   - Lista de personas y grupos, agrupada por tipo de relación.
   - Las que están pending_review se distinguen con claridad y se pueden confirmar,
     renombrar, fusionar o archivar con un toque. Es la pantalla que corrige los errores
     de la creación automática, así que tiene que ser rápida de usar.

2. FICHA DE PERSONA — la pantalla nueva más importante de esta fase:
   - Nombre, relación y alias, todo editable.
   - Sus hechos separados visualmente por tipo, porque significan cosas distintas:
       · Lo que le pasa ahora (states, con su vigencia visible)
       · Cómo es (traits)
       · Cómo eres tú con ella (creencias del usuario con ese scope)
   - Cada hecho se puede corregir o retirar, igual que en "Lo que Sofía sabe de ti".
   - Episodios recientes sobre esa persona, en orden cronológico.
   - Un interruptor para silenciar los hilos abiertos de esa persona.

3. TONO. Aplica el mismo criterio de la pantalla de creencias: esto es un perfil de una
   persona real que no está presente ni lo pidió. La interfaz debe leerse como notas
   personales del usuario, no como un expediente. Cuida los textos: "lo que sé de Diana"
   funciona; "perfil de Diana" no.

4. En el chat, muestra de forma discreta qué contexto está activo, y permite cambiarlo a
   mano. El usuario debe poder ver y corregir de quién cree Sofía que están hablando.

5. Cuelga la lista de entidades de la pantalla de ajustes, junto a Contextos.

6. `fvm dart analyze lib` sin issues.
```

---

# Lo que tienes que hacer tú

1. **Da de alta a mano tus 5 o 6 personas principales** en cuanto exista el paso 23, con
   sus alias reales (apodos incluidos: son los que de verdad usas al hablar). El resto que
   se creen solas.

2. **Crea 1 o 2 grupos** — «familia extendida», «primos» — con los nombres de sus miembros
   como alias. Así una mención de un primo resuelve al grupo sin generar una ficha por
   cada uno.

3. **Revisa las entidades pendientes una vez por semana.** Es donde se corrigen los
   duplicados y los falsos positivos de la detección automática. Cinco minutos.

4. **Marca como sensibles los hechos que no quieras que Sofía saque por iniciativa
   propia.** Salud, dinero y conflictos entran por defecto, pero solo tú sabes qué más.

5. **Decide el umbral de promoción.** Empieza en 3 menciones. Si aparecen entidades basura,
   súbelo a 4 o 5 con `ENTITY_PROMOTION_THRESHOLD`.

6. **Cuando montes el repo público**, recuerda que esta fase multiplica lo que hay que
   limpiar: la base contendrá nombres, relaciones y estados de personas reales de tu
   círculo. El repo público debe llevar solo código y esquema, jamás un volcado de datos
   ni capturas con nombres reales.

---

# Notas de diseño

**Por qué se extiende `Belief` en vez de crear `entity_facts`.** Un hecho sobre una persona
y una creencia sobre el usuario comparten toda la maquinaria de ciclo de vida: evidencia,
contradicciones, niveles de confianza, embeddings, deduplicación, cadenas de supersede. Lo
que cambia es la política de recuperación y el decaimiento, y eso son parámetros, no un
modelo nuevo. Duplicar la colección habría significado duplicar seis mecanismos para ganar
limpieza semántica que no aporta nada funcional.

**Por qué los hechos de entidad no van al prefijo cacheable.** Es la decisión con más
consecuencias de la fase. El prefijo tiene que ser byte-idéntico durante todo el día para
que el proveedor lo cachee; si el bloque de entidad viviera ahí, cada cambio de contexto
invalidaría el caché y se perdería alrededor del 45 % del ahorro del camino caliente. Por
eso el bloque de entidad vive en el sufijo, y por eso el test de prefijo byte-idéntico de
la fase 2 tiene que seguir en verde después de esta fase.

**Por qué la memoria episódica puede leer conversaciones y la síntesis no.** La regla
existe para que el ruido del día a día no contamine el prompt base, que es lo que gobierna
todo el comportamiento. Los episodios no entran al prompt base: se recuperan solo cuando
la entidad está activa y hay coincidencia. Son dos caminos con garantías distintas sobre
el mismo dato de entrada, y la separación mantiene intacta la garantía que importa.

**Lo que esta fase NO resuelve.** El escenario de referencia tiene dos mitades. Esta fase
construye el conocimiento: Sofía sabrá que a Diana le gusta Mon Laferte y que la ruptura
es reciente. Pero *poner* la música, *buscar* una receta o *sugerir* una película necesita
herramientas que hoy no existen — las ocho actuales son todas internas. Es un proyecto
aparte, y va después: sin el conocimiento, las herramientas dan respuestas genéricas que
daría cualquier asistente.
