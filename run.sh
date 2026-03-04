#!/usr/bin/env bash
set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_PID=""
FRONTEND_PID=""

CONFIG_FILE="${CONFIG_PATH:-$ROOT_DIR/configs/config.yaml}"
BACKEND_PORT=$(grep -A1 '^server:' "$CONFIG_FILE" | grep 'port:' | awk '{print $2}')
BACKEND_PORT="${BACKEND_PORT:-8888}"

cleanup() {
  echo ""
  echo "Shutting down..."
  [ -n "$BACKEND_PID" ] && kill "$BACKEND_PID" 2>/dev/null
  [ -n "$FRONTEND_PID" ] && kill "$FRONTEND_PID" 2>/dev/null
  wait 2>/dev/null
  echo "Done."
}
trap cleanup EXIT INT TERM

# Backend
echo "=== Starting Backend (Go) ==="
cd "$ROOT_DIR/backend"
export CONFIG_PATH="$CONFIG_FILE"
go run cmd/server/main.go &
BACKEND_PID=$!
echo "Backend PID: $BACKEND_PID"

# Frontend
echo "=== Starting Frontend (React) ==="
cd "$ROOT_DIR/frontend"
VITE_PROXY_TARGET="http://localhost:$BACKEND_PORT" npm run dev &
FRONTEND_PID=$!
echo "Frontend PID: $FRONTEND_PID"

echo ""
echo "=== Running ==="
echo "  Backend:  http://localhost:$BACKEND_PORT"
echo "  Frontend: http://localhost:5173"
echo ""
echo "Press Ctrl+C to stop."

wait
