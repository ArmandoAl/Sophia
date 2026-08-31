# Deploy de Sofia: Cloud Run detrás de Firebase Hosting

Este runbook prepara el backend Go como un servicio Cloud Run y usa Firebase Hosting como fachada HTTPS mediante un rewrite global. No usa archivos JSON de service account: Cloud Run accede a Firestore y Secret Manager con su identidad de servicio y Application Default Credentials (ADC).

Referencias oficiales: [Cloud Run deploy](https://cloud.google.com/sdk/gcloud/reference/run/deploy), [Cloud Build + Artifact Registry](https://cloud.google.com/build/docs/build-push-docker-image), [Firebase Hosting → Cloud Run](https://firebase.google.com/docs/hosting/full-config#direct_requests_to_a_cloud_run_container), [Firestore IAM](https://cloud.google.com/firestore/docs/security/iam) y [Secret Manager](https://cloud.google.com/secret-manager/docs/create-secret-quickstart).

## Configuración de producción

Variables no secretas de Cloud Run:

| Variable | Valor |
| --- | --- |
| `ENV` | `production` |
| `PERSISTENCE_DRIVER` | `firestore` |
| `FIRESTORE_PROJECT_ID` | `<PROJECT_ID>` |
| `FIRESTORE_DATABASE_ID` | Omitir para la base default; definir el ID real si se usa una base nombrada |
| `AI_MODEL_PROVIDER` | `deepseek` |
| `DEEPSEEK_MODEL` | `deepseek-v4-pro` |
| `AI_RUNTIME_PROPOSAL_ONLY` | `true` |
| `REMINDER_WORKER_ENABLED` | `false` |
| `FCM_ENABLED` | `false` |

Variables secretas, montadas como variables de entorno desde Secret Manager:

| Variable | Secreto |
| --- | --- |
| `DEEPSEEK_API_KEY` | `deepseek-api-key` |
| `JWT_SECRET` | `sofia-jwt-secret` |

Cloud Run inyecta `PORT`; no se configura manualmente. Tampoco se configura `GOOGLE_APPLICATION_CREDENTIALS`: la librería de Firestore usa ADC con la service account de Cloud Run.

## 1. Preparar placeholders y APIs

Ejecutar desde la raíz del backend y reemplazar los placeholders. `<REGION>` debe ser una región soportada por el rewrite de Firebase Hosting. `firebase.json` contiene `us-central1` como placeholder; cambiarlo antes del deploy si Cloud Run usa otra región.

```sh
export PROJECT_ID="<PROJECT_ID>"
export REGION="<REGION>"
export RUNTIME_SA="sofia-backend@${PROJECT_ID}.iam.gserviceaccount.com"
export IMAGE_TAG="$(date -u +%Y%m%d-%H%M%S)"
export IMAGE="${REGION}-docker.pkg.dev/${PROJECT_ID}/sofia/sofia-backend:${IMAGE_TAG}"

gcloud config set project "${PROJECT_ID}"
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com \
  firestore.googleapis.com \
  iam.googleapis.com \
  --project="${PROJECT_ID}"
```

## 2. Service account y Firestore mediante ADC

```sh
gcloud iam service-accounts create sofia-backend \
  --display-name="Sofia backend Cloud Run" \
  --project="${PROJECT_ID}"

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:${RUNTIME_SA}" \
  --role="roles/datastore.user"
```

`roles/datastore.user` concede lectura/escritura de documentos al runtime. El usuario o CI que ejecute `firebase deploy --only firestore:indexes` necesita permisos administrativos de índices por separado; no se deben conceder al runtime.

## 3. Crear secretos y permitir acceso

Los siguientes comandos leen los valores sin mostrarlos y crean la versión `1`. Si un secreto ya existe, agregar una versión nueva con `gcloud secrets versions add` y desplegar usando el número de esa versión.

```sh
read -r -s -p "DeepSeek API key: " DEEPSEEK_API_KEY_VALUE; printf '\n'
printf '%s' "${DEEPSEEK_API_KEY_VALUE}" | gcloud secrets create deepseek-api-key \
  --replication-policy="automatic" \
  --data-file=- \
  --project="${PROJECT_ID}"
unset DEEPSEEK_API_KEY_VALUE

read -r -s -p "JWT secret: " JWT_SECRET_VALUE; printf '\n'
printf '%s' "${JWT_SECRET_VALUE}" | gcloud secrets create sofia-jwt-secret \
  --replication-policy="automatic" \
  --data-file=- \
  --project="${PROJECT_ID}"
unset JWT_SECRET_VALUE

gcloud secrets add-iam-policy-binding deepseek-api-key \
  --member="serviceAccount:${RUNTIME_SA}" \
  --role="roles/secretmanager.secretAccessor" \
  --project="${PROJECT_ID}"

gcloud secrets add-iam-policy-binding sofia-jwt-secret \
  --member="serviceAccount:${RUNTIME_SA}" \
  --role="roles/secretmanager.secretAccessor" \
  --project="${PROJECT_ID}"
```

Rotación posterior:

```sh
read -r -s -p "Nuevo valor: " SECRET_VALUE; printf '\n'
printf '%s' "${SECRET_VALUE}" | gcloud secrets versions add <SECRET_NAME> \
  --data-file=- \
  --project="${PROJECT_ID}"
unset SECRET_VALUE
```

Usar siempre un número de versión explícito en producción, no `latest`.

## 4. Crear el repositorio de imágenes y hacer build + push

El repositorio de Artifact Registry se crea una sola vez:

```sh
gcloud artifacts repositories create sofia \
  --repository-format="docker" \
  --location="${REGION}" \
  --description="Sofia backend images" \
  --project="${PROJECT_ID}"
```

Cloud Build construye el `Dockerfile` y publica la imagen:

```sh
gcloud builds submit . \
  --region="${REGION}" \
  --tag="${IMAGE}" \
  --project="${PROJECT_ID}"
```

## 5. Desplegar Cloud Run

Para la base Firestore default, omitir `FIRESTORE_DATABASE_ID`:

```sh
gcloud run deploy sofia-backend \
  --image="${IMAGE}" \
  --region="${REGION}" \
  --platform="managed" \
  --service-account="${RUNTIME_SA}" \
  --allow-unauthenticated \
  --set-env-vars="ENV=production,PERSISTENCE_DRIVER=firestore,FIRESTORE_PROJECT_ID=${PROJECT_ID},AI_MODEL_PROVIDER=deepseek,DEEPSEEK_MODEL=deepseek-v4-pro,AI_RUNTIME_PROPOSAL_ONLY=true,REMINDER_WORKER_ENABLED=false,FCM_ENABLED=false" \
  --set-secrets="DEEPSEEK_API_KEY=deepseek-api-key:1,JWT_SECRET=sofia-jwt-secret:1" \
  --project="${PROJECT_ID}"
```

Si se usa una base nombrada, añadirla después con su ID real:

```sh
gcloud run services update sofia-backend \
  --region="${REGION}" \
  --update-env-vars="FIRESTORE_DATABASE_ID=<FIRESTORE_DATABASE_ID>" \
  --project="${PROJECT_ID}"
```

El servicio permite invocación no autenticada porque Firebase Hosting necesita alcanzarlo; las rutas privadas de Sofia siguen protegidas por el JWT de la aplicación.

## 6. Conectar Firebase Hosting y desplegar índices

`firebase.json` apunta a `sofia-backend` en `us-central1`. Antes de desplegar, reemplazar esa región si `${REGION}` es distinta. No se genera `.firebaserc` por adelantado porque todavía no existe un project ID.

```sh
firebase login
firebase use --add
# En el selector interactivo: elegir <PROJECT_ID> y asignar el alias deseado.

firebase deploy \
  --only hosting,firestore:indexes \
  --project="${PROJECT_ID}"
```

Los dos casos documentados como “Worker due scan” comparten `status ASC, next_run_at ASC`; `firestore.indexes.json` declara ese índice una sola vez.

## 7. Smoke test

Probar primero Cloud Run y después la fachada de Hosting:

```sh
export RUN_URL="$(gcloud run services describe sofia-backend \
  --region="${REGION}" \
  --project="${PROJECT_ID}" \
  --format='value(status.url)')"

curl --fail --silent --show-error "${RUN_URL}/health"
curl --fail --silent --show-error "https://${PROJECT_ID}.web.app/health"
```

Esperado: HTTP 200 y un JSON de health con `environment=production` y estado de Firestore no degradado.

## Rollback

### Cloud Run

Listar revisiones y enviar todo el tráfico a la revisión anterior:

```sh
gcloud run revisions list \
  --service="sofia-backend" \
  --region="${REGION}" \
  --project="${PROJECT_ID}"

gcloud run services update-traffic sofia-backend \
  --to-revisions="<PREVIOUS_REVISION>=100" \
  --region="${REGION}" \
  --project="${PROJECT_ID}"
```

Para regresar después a la revisión más reciente:

```sh
gcloud run services update-traffic sofia-backend \
  --to-latest \
  --region="${REGION}" \
  --project="${PROJECT_ID}"
```

Referencia: [rollbacks de Cloud Run](https://cloud.google.com/run/docs/rollouts-rollbacks-traffic-migration#rollback).

### Firebase Hosting

La Firebase CLI no ofrece un comando directo para revertir el canal live a una release histórica. En Firebase Console abrir **Hosting & Serverless → Hosting → Release history**, localizar la release anterior, abrir el menú y seleccionar **Roll back**. Esto crea una nueva release que vuelve a servir la versión seleccionada. Los índices de Firestore no se revierten con el rollback de Hosting.

Referencia: [rollback de Firebase Hosting](https://firebase.google.com/docs/hosting/manage-hosting-resources#roll_back).

## Regla de secretos

Nunca poner `DEEPSEEK_API_KEY`, `JWT_SECRET` ni credenciales Google en `firebase.json`, en la imagen Docker, en el repositorio, en logs o en `--dart-define` de Flutter. El cliente móvil solo recibe la URL HTTPS pública de la fachada.
