#!/usr/bin/env bash
# Preview web desde un servidor Linux (Jaspr, sin hot reload de Flutter).
#   scripts/preview_web.sh build | start | status | stop
# Modo demo (por defecto, sin red ni credenciales): DEMO=1. API real: SOFIA_API_BASE_URL=https://… (el backend debe permitir el origen por CORS).
# `start` compila (jaspr build) y sirve apps/web/build/jaspr en 127.0.0.1:$PORT (7703 por defecto) en segundo plano.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="$ROOT_DIR/apps/web"; OUT_DIR="$WEB_DIR/build/jaspr"
RUN_DIR="${RUN_DIR:-$HOME/projects/.run}"; PORT="${PORT:-7703}"; PID="$RUN_DIR/sophia-web.pid"
export PATH="$HOME/.pub-cache/bin:$PATH"
DEFINES=()
[[ "${DEMO:-1}" == 1 ]] && DEFINES+=(--dart-define=DEMO_MODE=true)
[[ -n "${SOFIA_API_BASE_URL:-}" ]] && DEFINES+=("--dart-define=SOFIA_API_BASE_URL=$SOFIA_API_BASE_URL")
mkdir -p "$RUN_DIR"
alive() { [[ -f "$PID" ]] && kill -0 "$(cat "$PID")" 2>/dev/null; }
build() {
  command -v jaspr >/dev/null || { echo "Falta jaspr: dart pub global activate jaspr_cli 0.23.4" >&2; exit 1; }
  (cd "$WEB_DIR" && jaspr build ${DEFINES[@]+"${DEFINES[@]}"})
  printf '%s %s\n' "$(git -C "$ROOT_DIR" branch --show-current)" "$(git -C "$ROOT_DIR" rev-parse --short HEAD)" > "$OUT_DIR/preview-version.txt"
}
stop() { alive && kill "$(cat "$PID")" 2>/dev/null || true; rm -f "$PID"; [[ "${1:-}" == quiet ]] || echo "Preview detenido."; }
status() {
  alive || { echo "Preview detenido."; return 1; }
  echo "Preview activo: http://127.0.0.1:$PORT (HTTP $(curl -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:$PORT/")) · $(cat "$OUT_DIR/preview-version.txt" 2>/dev/null)"
}
case "${1:-}" in
  build) build ;;
  start) stop quiet; build
         ss -ltn "sport = :$PORT" 2>/dev/null | grep -q LISTEN && { echo "Puerto $PORT ocupado" >&2; exit 1; }
         ROOT="$OUT_DIR" PORT="$PORT" setsid nohup node "$ROOT_DIR/scripts/static_spa.mjs" >"$RUN_DIR/sophia-web.log" 2>&1 &
         echo $! > "$PID"; sleep 1; status ;;
  stop) stop ;;
  status) status ;;
  *) sed -n '2,4p' "$0"; exit 1 ;;
esac
