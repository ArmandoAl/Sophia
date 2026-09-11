# Sofía — Runbook de release

Del código verificado a la app en tu teléfono. Sustituye la sección de despliegue de
`runbook_despliegue.md`, que tenía dos errores (nombre del secreto JWT y comando de build).

**Estado verificado el 11 de septiembre:**

| | |
|---|---|
| Backend | 36 paquetes en verde, 0 fallos |
| Repo | sin secretos rastreados, sin `.env` |
| Secret Manager | `deepseek-api-key` real (`sk-…`, 35 chars) · `sofia-jwt-secret` |
| Cloud Run | `sofia-backend` corriendo la revisión del **1 de septiembre** — código viejo |
| Service account | `sofia-backend@sophia-ai-60e87.iam.gserviceaccount.com` — solo `datastore.user` |
| Scheduler | API no habilitada |
| App | `defaultBaseUrl = localhost:8080` |

```bash
export SOFIA=~/Documents/Sophia/sophia_ai_backend/sofia-backend
export P=sophia-ai-60e87
export R=us-central1
export SA=sofia-backend@sophia-ai-60e87.iam.gserviceaccount.com
```

---

## Sobre tu clave de DeepSeek

Verificado: **está bien**. El secreto `deepseek-api-key` en Secret Manager empieza por
`sk-` y tiene 35 caracteres — es la clave real, no el placeholder. Lo creaste el 31 de
agosto con el valor correcto. El servicio de Cloud Run la lee por `secretKeyRef`, que es
la forma correcta: nunca aparece en texto plano en la configuración.

**La app no necesita ninguna clave.** Solo habla con tu backend usando el JWT del login. Por
eso no hay ni habrá un `.env` en Flutter: el único valor que varía es la URL de la API.

### Copia de seguridad local de tus claves

Guárdalas **fuera del repo**, para que ningún `git add -A` las pueda arrastrar:

```bash
mkdir -p ~/.sofia && chmod 700 ~/.sofia
cat > ~/.sofia/secrets.env <<'EOF'
# Sofía — copia local. Los valores canónicos viven en Secret Manager (sophia-ai-60e87).
DEEPSEEK_API_KEY=$(gcloud secrets versions access latest --secret=deepseek-api-key --project=sophia-ai-60e87)
JWT_SECRET=$(gcloud secrets versions access latest --secret=sofia-jwt-secret --project=sophia-ai-60e87)
EOF
chmod 600 ~/.sofia/secrets.env
```

Para usarlo en local: `set -a && source ~/.sofia/secrets.env && set +a`.

---

## 0 · Commit

Toda la fase 5 y el rediseño están sin commitear.

```bash
cd ~/Documents/Sophia
go -C "$SOFIA" test ./... && (cd sophia_ai && fvm dart analyze lib)
git add -A
git commit -m "Fase 5: entidades, episodios e hilos abiertos; rediseño de la app; cloudbuild"
git push
```

---

## 1 · Permisos que faltan

La service account tiene `datastore.user` y lee secretos, pero le falta Vertex AI para
los embeddings. Sin esto la deduplicación semántica degrada a léxico en silencio.

```bash
gcloud projects add-iam-policy-binding $P --member="serviceAccount:$SA" --role="roles/aiplatform.user"
```

**Verificación:**
```bash
gcloud projects get-iam-policy $P --flatten="bindings[].members" \
  --filter="bindings.members:$SA" --format="value(bindings.role)"
```
Debe listar `roles/aiplatform.user` además de `roles/datastore.user`.

---

## 2 · Construir las imágenes

Una sola orden construye la API y el worker con el target correcto de cada uno:

```bash
cd "$SOFIA"
gcloud builds submit --config cloudbuild.yaml --project $P
```

Tarda unos minutos. **Verificación:** termina en `SUCCESS` y lista dos imágenes:
`sofia-backend:latest` y `sofia-synthesis:latest`.

---

## 3 · Redesplegar la API

Actualiza la imagen y añade las variables nuevas. Las que ya estaban (secretos, Firestore,
proveedor) se conservan porque `--update-env-vars` no borra las existentes.

```bash
gcloud run deploy sofia-backend \
  --project $P --region $R \
  --image gcr.io/$P/sofia-backend:latest \
  --update-env-vars=GOOGLE_CLOUD_PROJECT=$P,EMBEDDINGS_ENABLED=true,EMBEDDINGS_MODEL=text-embedding-004,AUTONOMY_THRESHOLD=0.85,CONTEXT_TOKEN_BUDGET=5000 \
  --update-secrets=DEEPSEEK_API_KEY=deepseek-api-key:latest,JWT_SECRET=sofia-jwt-secret:latest
```

> `--update-secrets` cambia los secretos de la versión fija `1` a `latest`. Así, si un día
> rotas la clave, basta con añadir una versión nueva al secreto y redesplegar.

**Verificación:**
```bash
curl -s https://sofia-backend-m737njuj6q-uc.a.run.app/health
gcloud run services describe sofia-backend --project $P --region $R --format="value(status.latestReadyRevisionName)"
```
El health debe devolver `"firestore":"ok"` y la revisión debe ser nueva (`00006` o mayor).

Si el despliegue falla, los logs dicen por qué:
```bash
gcloud run services logs read sofia-backend --project $P --region $R --limit 50
```
Lo más probable si falla al arrancar: una variable obligatoria nueva que falta. El mensaje
de `config.Load()` la nombra.

---

## 4 · El worker de síntesis como Job

Se ejecuta a demanda y muere. **Nunca lo despliegues como Service**: correría 24/7 y
costaría.

```bash
gcloud run jobs create sofia-synthesis \
  --project $P --region $R \
  --image gcr.io/$P/sofia-synthesis:latest \
  --service-account $SA \
  --set-env-vars=ENV=production,PERSISTENCE_DRIVER=firestore,FIRESTORE_PROJECT_ID=$P,FIRESTORE_DATABASE_ID=default,GOOGLE_CLOUD_PROJECT=$P,SYNTHESIS_WORKER_ENABLED=true,AI_MODEL_PROVIDER=deepseek,DEEPSEEK_MODEL=deepseek-v4-pro,EMBEDDINGS_ENABLED=true,EMBEDDINGS_MODEL=text-embedding-004,SYNTHESIS_RUN_HOUR_LOCAL=4 \
  --set-secrets=DEEPSEEK_API_KEY=deepseek-api-key:latest,JWT_SECRET=sofia-jwt-secret:latest \
  --max-retries=1 --task-timeout=15m
```

Si el job ya existía, cambia `create` por `update`.

**Pruébalo a mano antes de programarlo:**
```bash
gcloud run jobs execute sofia-synthesis --project $P --region $R --wait
```

**Verificación:** termina en `Succeeded`. Con la base casi vacía, el worker no tiene nada
que sintetizar y sale rápido — eso es correcto. Lo que confirma es que arranca, conecta con
Firestore y no falla por configuración. Sus logs:
```bash
gcloud logging read 'resource.type=cloud_run_job AND resource.labels.job_name=sofia-synthesis' \
  --project $P --limit 30 --format="value(textPayload)"
```

---

## 5 · Programarlo cada noche

```bash
gcloud services enable cloudscheduler.googleapis.com --project $P

gcloud scheduler jobs create http sofia-synthesis-daily \
  --project $P --location $R \
  --schedule="0 4 * * *" --time-zone="America/Mexico_City" \
  --uri="https://$R-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/$P/jobs/sofia-synthesis:run" \
  --http-method=POST \
  --oauth-service-account-email="$SA"
```

La service account necesita permiso para invocar el job:
```bash
gcloud run jobs add-iam-policy-binding sofia-synthesis \
  --project $P --region $R --member="serviceAccount:$SA" --role="roles/run.invoker"
```

**Verificación:** fuerza una ejecución desde el scheduler y comprueba que llega al job:
```bash
gcloud scheduler jobs run sofia-synthesis-daily --project $P --location $R
gcloud run jobs executions list --job sofia-synthesis --project $P --region $R --limit 3
```

---

## 6 · Apuntar la app a producción

Hoy `defaultBaseUrl` es `localhost:8080`. Para un release eso es una trampa: cualquier
build que olvide el `--dart-define` apunta a una máquina que no existe.

**Invierte el valor por defecto.** En `sophia_ai/lib/core/config/app_config.dart`:

```dart
static const String defaultBaseUrl = 'https://sofia-backend-m737njuj6q-uc.a.run.app';
```

Y para desarrollo local pasas el define al revés:
```bash
fvm flutter run --dart-define=SOFIA_API_BASE_URL=http://localhost:8080
```

Así la dirección segura es la de producción: un build sin defines funciona, y solo el
desarrollo requiere acción explícita.

> No es un secreto: la URL de tu API es pública por naturaleza. Lo que la protege es el
> JWT, no que la URL sea difícil de adivinar.

**Verificación:** con el backend local apagado, `fvm flutter run` sin defines debe iniciar
sesión contra producción.

---

## 7 · Prueba de humo contra producción

Desde la app, en este orden. Cada paso confirma una pieza distinta:

| # | Acción | Confirma |
|---|---|---|
| 1 | Registro o login | auth + JWT + Firestore |
| 2 | Escribe «recuérdame llamar a mi hermana mañana a las 5» | runtime + DeepSeek + propuesta persistida |
| 3 | Pulsa **Ajustar** y cambia la hora | corrección de tres estados |
| 4 | Pulsa **Hacerlo** | ejecución real → aparece en Recordatorios |
| 5 | Ajustes → Contextos → crea `work` | entidades / contextos |
| 6 | Escribe algo mencionando a alguien por su nombre | detección de entidad |
| 7 | Diagnóstico → sección de aprendizaje | endpoints `/learning/*` |

Al terminar, en Firestore deben existir documentos en `ai_action_proposals`, `reminders`,
`user_contexts` y `conversations`:
```bash
TOKEN=$(gcloud auth application-default print-access-token)
for C in ai_action_proposals reminders user_contexts entity_candidates; do
  N=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "https://firestore.googleapis.com/v1/projects/$P/databases/default/documents/$C?pageSize=50" \
    | python3 -c "import sys,json;print(len(json.load(sys.stdin).get('documents',[])))")
  printf "%-22s %s\n" "$C" "$N"
done
```

Si algo falla, el orden te dice dónde: si el paso 2 no propone nada, revisa los logs del
servicio — suele ser DeepSeek (clave, modelo o cuota).

---

## 8 · TestFlight

Requisitos previos que solo tú tienes: cuenta de Apple Developer activa, el bundle ID
`com.aial.sophiav1` registrado en App Store Connect, y un perfil de firma.

```bash
cd ~/Documents/Sophia/sophia_ai
fvm flutter build ipa --release
```

Sin `--dart-define`: tras el paso 6 el default ya es producción. El `.ipa` queda en
`build/ios/ipa/`. Súbelo con Transporter o desde Xcode (Product → Archive → Distribute).

Recuerda tu restricción conocida: el simulador de iPhone 17 Pro Max necesita modo release.
Un dispositivo físico no tiene esa limitación.

---

## 9 · Repo público — cuando la app ya funcione

### Lo que verifiqué

El repo **no contiene secretos**: ni en archivos rastreados ni en `.env`. Lo que sí contiene
es **configuración específica de tu proyecto**, que no es secreta pero hace que el repo
no sea portable:

| Archivo | Contenido específico | ¿Secreto? |
|---|---|---|
| `.firebaserc` | `sophia-ai-60e87` | No — un ID de proyecto no da acceso a nada |
| `firebase.json` | rutas de Cloud Run | No |
| `docs/runbook_*.md` | IDs, URLs, nombres de service account | No |
| `app_config.dart` | URL de tu API | No |
| `ios/` | bundle ID `com.aial.sophiav1` | No |

**Un ID de proyecto de GCP no es una credencial.** Con él nadie puede leer tu Firestore: las
reglas son `deny-all` y el backend exige tu JWT. La razón para limpiarlo es que quien clone
el repo pueda usar el suyo, no seguridad.

### Cómo hacer el split

**No copies la historia de git.** Aunque HEAD esté limpio, un commit antiguo podría
contener algo. Un repo público empieza de cero:

```bash
# 1. Clona el privado a una carpeta aparte, sin .git
cd ~/Documents
git clone --depth 1 https://github.com/ArmandoAl/Sophia.git Sophia-public
cd Sophia-public && rm -rf .git

# 2. Quita lo específico de tu proyecto
rm -f sophia_ai_backend/sofia-backend/.firebaserc
rm -rf sophia_ai_design/              # capturas con tu diseño real; ponlas en el README
rm -rf docs/runbook_*.md              # contienen tus IDs y URLs
rm -rf sophia_ai_backend/docs/        # duplicado con reportes de sprint internos

# 3. Parametriza
#    - app_config.dart: defaultBaseUrl vuelve a 'http://localhost:8080'
#    - firebase.json: quita el bloque "hosting" o deja el serviceId como placeholder
#    - ios bundle ID: documenta que hay que cambiarlo, no lo cambies tú

# 4. Escribe SETUP.md: "trae tu propio Firebase y tu clave de DeepSeek"
#    Lista exacta de lo que necesita quien lo clone:
#      - proyecto de GCP con Firestore (base nombrada, o ajustar FIRESTORE_DATABASE_ID)
#      - habilitar Vertex AI si quiere embeddings
#      - DEEPSEEK_API_KEY
#      - JWT_SECRET propio
#      - las variables de docs/environment.md

# 5. Repo nuevo, historia limpia
git init -b main && git add -A && git commit -m "Sofía — código abierto"
gh repo create ArmandoAl/sofia --public --source=. --push
```

**El privado (`ArmandoAl/Sophia`) sigue siendo tu repo de trabajo.** Cada tanto vuelcas al
público lo que quieras compartir. Nunca al revés.

### Antes de hacerlo público, una última pasada

```bash
cd ~/Documents/Sophia-public
grep -rIE "sk-[a-zA-Z0-9]{20,}|BEGIN [A-Z ]*PRIVATE KEY|sophia-ai-60e87|m737njuj6q" . | head
```
Debe salir vacío. Si `sophia-ai-60e87` aparece, es un runbook o un README que se te pasó.

---

## Resumen del orden

```
0. commit + push
1. rol aiplatform.user a la service account
2. gcloud builds submit --config cloudbuild.yaml
3. gcloud run deploy sofia-backend          ← código nuevo en producción
4. gcloud run jobs create sofia-synthesis   ← probar a mano
5. scheduler                                ← solo cuando el 4 salga Succeeded
6. defaultBaseUrl → producción
7. prueba de humo desde la app
8. TestFlight
9. repo público — cuando todo lo anterior funcione
```
