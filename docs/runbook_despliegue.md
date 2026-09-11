# Sofía — Runbook de despliegue

> **Para desplegar a Cloud Run, TestFlight y el repo público usa `runbook_release.md`.**
> La sección 8 de este documento tiene dos errores: el secreto JWT se llama
> `sofia-jwt-secret`, y `gcloud builds submit --tag` no construye el worker (hay que usar
> `cloudbuild.yaml`). Los pasos 1-7 y 9-11 de aquí siguen siendo válidos.

Proyecto GCP: **`sophia-ai-60e87`** · Cuenta: `alvaradoarmandodev@gmail.com`

Sigue el orden. Cada bloque dice **qué hace**, **qué comando** y **cómo saber que salió bien**.
Nada de esto pide que escribas datos: son migraciones que leen lo que ya está en Firestore.

Atajo de rutas:
```bash
export SOFIA=~/Documents/Sophia/sophia_ai_backend/sofia-backend
```

## Dos trampas de este proyecto

**1. La base de datos se llama `default`, no `(default)`.** Firestore crea por convención una
base implícita llamada `(default)` (con paréntesis). Este proyecto **no la tiene**: tiene una
base *nombrada* `default`. Son cosas distintas, y `FIRESTORE_DATABASE_ID` no tiene valor por
defecto en `config.go`, así que si no lo pasas el cliente apunta a `(default)` y falla con
`The database (default) does not exist`. **Toda ejecución del backend o de una herramienta
necesita `FIRESTORE_DATABASE_ID=default`**, incluidos los jobs de Cloud Run.

**2. Hay tres proyectos GCP en el historial de esta máquina.** `sophia-ai-60e87` (el bueno),
`sophia-ai-486821` (clave revocada en agosto) y `aifit-a7f6b` (ADC viejo). Por eso todos los
comandos pasan `--project sophia-ai-60e87` explícito: confiar en el proyecto "activo" es cómo
acabas desplegando en el sitio equivocado.

---

## 0 · Antes de tocar la nube: commitea

Ahora mismo tienes decenas de archivos sin commitear. Si algo sale mal en los pasos
siguientes, no tienes a dónde volver.

```bash
cd ~/Documents/Sophia
go -C "$SOFIA" vet ./... && go -C "$SOFIA" test ./...   # debe salir todo ok
git add -A
git commit -m "Fase 2: idempotencia del worker, solo DeepSeek, embeddings, contextos e ingesta"
git push
```

**Verificación:** `git status` limpio y `git log --oneline -1` muestra tu commit.

---

## 1 · Credenciales locales (sin archivos de clave)

Los backfills corren **desde tu Mac contra Firestore de producción**, así que necesitan
credenciales. **No descargues un JSON de service account** — en agosto tuviste una clave
expuesta por eso. Usa credenciales de usuario, que no dejan archivo permanente:

```bash
gcloud auth application-default login
```

Se abre el navegador. Inicia sesión con `alvaradoarmandodev@gmail.com`.

**Verificación:**
```bash
gcloud auth application-default print-access-token | head -c 20 && echo " ...ok"
```

---

## 2 · Índices de Firestore

**Qué hace:** crea los índices compuestos que necesitan las consultas nuevas (beliefs por
scope, propuestas por rango de fechas, búsqueda por `search_terms`, etc.). Sin ellos las
consultas fallan en producción con un error de índice faltante.

**Hazlo pronto:** construir índices tarda de minutos a horas según el volumen. Lánzalo
ahora y sigue con lo demás mientras se construyen.

```bash
cd "$SOFIA"
firebase deploy --only firestore:indexes --project sophia-ai-60e87
```

**Verificación:**
```bash
gcloud firestore indexes composite list --project=sophia-ai-60e87 --format="table(name.basename(),state)"
```
Todos deben acabar en `READY`. Mientras digan `CREATING`, espera. No corras el paso 5
hasta que estén listos.

---

## 3 · Candado de seguridad en Firestore

**Qué hace:** bloquea todo acceso desde clientes. Tu backend Go usa el Admin SDK, que se
salta las reglas por diseño, así que esto no le afecta. Es un candado por si algún día
alguien añade acceso directo desde la app.

Crea `$SOFIA/firestore.rules`:

```
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Todo el acceso pasa por el backend Go (Admin SDK, que ignora estas reglas).
    // Ningún cliente debe leer ni escribir Firestore directamente.
    match /{document=**} {
      allow read, write: if false;
    }
  }
}
```

Añade la referencia en `$SOFIA/firebase.json`, dentro del bloque `"firestore"`:

```json
"firestore": {
  "database": "default",
  "rules": "firestore.rules",
  "indexes": "firestore.indexes.json"
}
```

Despliega:
```bash
cd "$SOFIA"
firebase deploy --only firestore:rules --project sophia-ai-60e87
```

**Verificación:** en la consola de Firebase → Firestore → Reglas, debe verse `if false`.
Luego prueba que **tu backend sigue funcionando** (paso 7): si sigue leyendo y escribiendo,
el Admin SDK está bien configurado.

---

## 4 · Vertex AI para embeddings

**Qué hace:** habilita el servicio que convierte frases en vectores, para que Sofía
entienda que «junta» y «reunión» son lo mismo al deduplicar creencias.

```bash
# 1. Habilitar la API
gcloud services enable aiplatform.googleapis.com --project=sophia-ai-60e87

# 2. Confirmar que quedó habilitada
gcloud services list --enabled --project=sophia-ai-60e87 --filter="config.name:aiplatform" --format="value(config.name)"
```

Debe imprimir `aiplatform.googleapis.com`.

**Permisos.** Tu cuenta necesita `roles/aiplatform.user`. Si eres Owner del proyecto ya lo
tienes. Para comprobarlo:
```bash
gcloud projects get-iam-policy sophia-ai-60e87 \
  --flatten="bindings[].members" \
  --filter="bindings.members:alvaradoarmandodev@gmail.com" \
  --format="value(bindings.role)"
```
Si sale `roles/owner`, listo.

> Cuando despliegues el worker (paso 8), la **service account de Cloud Run** también
> necesita `roles/aiplatform.user`. Está en el paso 8.

---

## 5 · Los backfills — qué son y cómo se corren

### ¿Qué es un backfill?

Los pasos 4 y 11 añadieron **campos nuevos** a documentos que ya existían:

| Campo nuevo | En | Para qué |
|---|---|---|
| `search_terms` | `memories` | que la búsqueda por términos encuentre el recuerdo |
| `embedding` | `user_beliefs` | que la deduplicación entienda paráfrasis |

El código **solo rellena esos campos al crear o actualizar** un documento. Los documentos
guardados **antes** del cambio no los tienen, así que son invisibles para la lógica nueva.
Un backfill los recorre una vez y se los rellena.

**No escribes ni introduces ningún dato.** Lee lo que ya está en Firestore y completa los
campos que faltan. Se corre **una sola vez**, y es seguro repetirlo.

### 5a · `backfill-memory-search-terms`

Las variables van **en la misma línea** que el comando: así aplican a esa ejecución y no
hay forma de olvidarlas. Si las pones en un `export` aparte y copias solo la última línea,
el backfill arranca con `PERSISTENCE_DRIVER=memory` y falla.

```bash
cd "$SOFIA"

# 1. ENSAYO — no escribe nada, solo te dice qué haría
PERSISTENCE_DRIVER=firestore FIRESTORE_PROJECT_ID=sophia-ai-60e87 FIRESTORE_DATABASE_ID=default \
  go run ./cmd/tools/backfill-memory-search-terms -dry-run

# 2. Si el conteo tiene sentido, aplícalo de verdad
PERSISTENCE_DRIVER=firestore FIRESTORE_PROJECT_ID=sophia-ai-60e87 FIRESTORE_DATABASE_ID=default \
  go run ./cmd/tools/backfill-memory-search-terms -dry-run=false -apply
```

**Cómo leer la salida.** El ensayo imprime:
```
backfill preview scanned=N would_update=M dry_run=true apply=false
```
`scanned` = memorias revisadas, `would_update` = a cuántas les falta el campo.
Si `scanned=0`, aún no tienes memorias guardadas y no hay nada que hacer — es normal si
todavía no has usado Sofía.

Al aplicar:
```
backfill applied scanned=N updated=M
```

**Verificación:** vuelve a correr el `-dry-run`. Ahora `would_update` debe ser **0**.

> Prueba con `-limit 10` primero si quieres ver el comportamiento sobre pocos documentos.

### 5b · `backfill-belief-embeddings`

Este **sí gasta dinero** (una llamada de embedding por creencia) y **exige** que Vertex AI
esté habilitado (paso 4).

Este necesita cinco variables. Guárdalas en una variable de shell para no repetirlas:

```bash
cd "$SOFIA"
EMB="PERSISTENCE_DRIVER=firestore FIRESTORE_PROJECT_ID=sophia-ai-60e87 FIRESTORE_DATABASE_ID=default GOOGLE_CLOUD_PROJECT=sophia-ai-60e87 EMBEDDINGS_ENABLED=true EMBEDDINGS_MODEL=text-embedding-004"
# ↑ confirma el nombre del modelo en la consola de Vertex AI antes de aplicar

# 1. ENSAYO — no pide embeddings ni escribe
env $EMB go run ./cmd/tools/backfill-belief-embeddings -dry-run

# 2. Aplicar
env $EMB go run ./cmd/tools/backfill-belief-embeddings -dry-run=false -apply
```

Si se corta a media ejecución, es reanudable:
```bash
env $EMB go run ./cmd/tools/backfill-belief-embeddings -dry-run=false -apply -after <ultimo-id-procesado>
```

**Verificación:** el `-dry-run` de nuevo debe reportar 0 pendientes.

> Si aún no tienes creencias (no has corrido el worker de síntesis), esto reportará 0.
> Es lo esperado: vuelve aquí después de unos días de uso.

---

## 6 · Alerta de presupuesto — antes de automatizar nada

**Qué hace:** te avisa por correo si el gasto se dispara. El worker de síntesis llama al
modelo sin que nadie mire; un bug en bucle se nota en la factura, no en la pantalla.

Por consola (más simple que por CLI):
1. https://console.cloud.google.com/billing → tu cuenta de facturación
2. **Presupuestos y alertas** → **Crear presupuesto**
3. Ámbito: proyecto `sophia-ai-60e87`
4. Importe mensual: pon algo que te duela pero no te arruine (p. ej. 20 USD)
5. Umbrales de alerta: 50 %, 90 %, 100 %
6. Marca **enviar correo a los administradores de facturación**

**Verificación:** debe aparecer en la lista de presupuestos con tu proyecto como ámbito.

> Esto cubre el gasto de GCP (Vertex, Firestore, Cloud Run). **DeepSeek se factura aparte** —
> pon también un límite de gasto en su panel.

---

## 7 · Comprobar que el backend sigue vivo

Antes de automatizar, verifica a mano que las reglas del paso 3 no rompieron nada:

```bash
cd "$SOFIA"
PERSISTENCE_DRIVER=firestore FIRESTORE_PROJECT_ID=sophia-ai-60e87 FIRESTORE_DATABASE_ID=default \
JWT_SECRET=algo-local-de-prueba AI_MODEL_PROVIDER=fake \
  go run ./cmd/sofia
```

En otra terminal:
```bash
curl -s localhost:8080/health
```

**Verificación:** responde OK y en los logs no hay errores de permisos de Firestore.
Si aquí falla algo, **para**: no despliegues workers sobre una base rota.

---

## 8 · Worker de síntesis en Cloud Run

**Qué hace:** el trabajo diario que convierte tus decisiones en creencias y recompila el
prompt base. Es el corazón del sistema.

```bash
cd "$SOFIA"

# 1. Construir y subir la imagen del worker (el Dockerfile tiene el target "synthesis")
gcloud builds submit --project=sophia-ai-60e87 \
  --tag gcr.io/sophia-ai-60e87/sofia-synthesis

# 2. Crear el Job
gcloud run jobs create sofia-synthesis \
  --project=sophia-ai-60e87 \
  --region=us-central1 \
  --image gcr.io/sophia-ai-60e87/sofia-synthesis \
  --set-env-vars=PERSISTENCE_DRIVER=firestore,FIRESTORE_PROJECT_ID=sophia-ai-60e87,FIRESTORE_DATABASE_ID=default,GOOGLE_CLOUD_PROJECT=sophia-ai-60e87,SYNTHESIS_WORKER_ENABLED=true,AI_MODEL_PROVIDER=deepseek,EMBEDDINGS_ENABLED=true \
  --set-secrets=DEEPSEEK_API_KEY=deepseek-api-key:latest,JWT_SECRET=jwt-secret:latest \
  --max-retries=1
```

> El `--set-secrets` requiere que los guardes antes en Secret Manager:
> ```bash
> echo -n "TU_API_KEY" | gcloud secrets create deepseek-api-key --data-file=- --project=sophia-ai-60e87
> echo -n "TU_JWT_SECRET" | gcloud secrets create jwt-secret --data-file=- --project=sophia-ai-60e87
> ```
> Nunca pongas la API key en `--set-env-vars`: queda visible en la descripción del job.

**Permisos de la service account del job** (para Firestore y Vertex):
```bash
SA=$(gcloud run jobs describe sofia-synthesis --project=sophia-ai-60e87 --region=us-central1 --format="value(template.template.serviceAccount)")
SA=${SA:-$(gcloud projects describe sophia-ai-60e87 --format='value(projectNumber)')-compute@developer.gserviceaccount.com}

for ROLE in roles/datastore.user roles/aiplatform.user roles/secretmanager.secretAccessor; do
  gcloud projects add-iam-policy-binding sophia-ai-60e87 --member="serviceAccount:$SA" --role="$ROLE"
done
```

**3. Probar a mano antes de programarlo:**
```bash
gcloud run jobs execute sofia-synthesis --project=sophia-ai-60e87 --region=us-central1 --wait
gcloud run jobs executions list --job=sofia-synthesis --project=sophia-ai-60e87 --region=us-central1
```

**Verificación:** la ejecución termina en `Succeeded`. Revisa los logs:
```bash
gcloud logging read 'resource.type=cloud_run_job AND resource.labels.job_name=sofia-synthesis' \
  --project=sophia-ai-60e87 --limit=50 --format="value(textPayload)"
```

**4. Solo cuando la ejecución manual funcione**, prográmalo:
```bash
gcloud scheduler jobs create http sofia-synthesis-daily \
  --project=sophia-ai-60e87 \
  --location=us-central1 \
  --schedule="0 4 * * *" \
  --time-zone="America/Mexico_City" \
  --uri="https://us-central1-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/sophia-ai-60e87/jobs/sofia-synthesis:run" \
  --http-method=POST \
  --oauth-service-account-email="$SA"
```

Las 4 de la mañana en tu zona: el día ya cerró y nada compite por recursos.

---

## 9 · Tus contextos iniciales

**Qué hace:** define las personas y modos con los que Sofía ajusta su comportamiento.

Empieza con **3 o 4 reales**. Un catálogo de 20 contextos que no usas es ruido que compite
por espacio en el prompt.

Ejemplo de arranque:

| Kind | Slug | Label | Aliases |
|---|---|---|---|
| `mode` | `work` | Trabajo | trabajando, oficina, chamba |
| `mode` | `chill` | Descanso | relax, descansando |
| `person` | `pareja` | *(su nombre)* | *(apodos que usas)* |

Se crean por la API (paso 12 añadió `/contexts`). Con el backend corriendo:

```bash
TOKEN="<tu access token>"
curl -X POST localhost:8080/contexts \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"kind":"mode","slug":"work","label":"Trabajo","aliases":["trabajando","oficina","chamba"]}'
```

**Verificación:** `GET /contexts` los devuelve, y un belief con `scope_key=mode:work`
aparece en el sufijo del prompt solo cuando ese contexto está activo.

---

## 10 · Usar Sofía a diario

No es relleno. La cadena es:

```
decisiones tuyas (nivel 1) → beliefs con confianza alta → prompt base
                           ↘ ascienden los beliefs importados (nivel 3)
```

Sin decisiones reales no hay nivel 1, y sin nivel 1 **los beliefs importados de WhatsApp
jamás pasan de 0.5 de confianza** y nunca gobiernan nada. Necesitas ~10 decisiones por
herramienta antes de que `ShouldAutoExecute` permita algo: unas 2-3 semanas de uso.

---

## 11 · Ingesta de conversaciones — todavía no

Deja esto para el final, después de dos o tres semanas de uso real. Dos razones:

1. Sin decisiones tuyas acumuladas, los beliefs importados no tienen con qué corroborarse.
2. **Privacidad:** importar tus chats significa procesar mensajes de gente que no dio su
   consentimiento — tu pareja, tus amigos, tus clientes. Para uso estrictamente personal es
   una cosa; el día que Sofía tenga otros usuarios en TestFlight, cambia lo que legalmente
   tienes que declarar. Piénsalo antes de importar 500 conversaciones, no después.

Cuando llegue el momento, el formato está en `docs/ingestion_format.md` y todo lote se
puede deshacer entero con `DELETE /ingestion/batches/{id}`.

---

## Resumen del orden

```
0. commit + push
1. gcloud auth application-default login
2. desplegar índices          ← empieza pronto, tarda
3. reglas deny-all
4. habilitar Vertex AI
5. backfills (ensayo → aplicar)   ← espera a que los índices estén READY
6. alerta de presupuesto
7. probar el backend a mano
8. worker en Cloud Run (manual → programado)
9. crear 3-4 contextos
10. usar Sofía 2-3 semanas
11. ingesta de conversaciones (después)
```
