#!/usr/bin/env bash
# Start the Go backend for local mini-program joint debugging.
#   1) docker compose up -d mysql
#   2) scripts/dev-backend.sh
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f .env ]; then
  echo "missing .env at repo root (DASHSCOPE_API_KEY etc.)" >&2
  exit 1
fi
set -a
# shellcheck disable=SC1091
source .env
set +a

export DB_HOST="${DB_HOST:-127.0.0.1}"
export DB_PORT="${DB_PORT:-3306}"
export DB_USER="${DB_USER:-travelgd}"
export DB_PASSWORD="${DB_PASSWORD:-travelgd}"
export DB_NAME="${DB_NAME:-travel_gd}"
export JWT_SECRET="${JWT_SECRET:-local-dev-secret-please-change-0123456789}"
export PORT="${PORT:-8081}"
# Keep uploads inside the repo for local runs; production uses /opt/travel-gd/uploads.
export UPLOAD_DIR="${UPLOAD_DIR:-$(pwd)/.local-uploads}"
mkdir -p "$UPLOAD_DIR"

echo "backend: http://127.0.0.1:${PORT}  (uploads: ${UPLOAD_DIR})"
cd server-go
exec go run .
