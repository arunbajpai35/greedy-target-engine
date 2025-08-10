#!/usr/bin/env bash
set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DB_URL="postgres://postgres:password@localhost:5432/targeting_db?sslmode=disable"
SERVER_PID_FILE="$ROOT_DIR/bin/server.pid"

log() { echo -e "${YELLOW}$*${NC}"; }
ok() { echo -e "${GREEN}$*${NC}"; }
err() { echo -e "${RED}$*${NC}"; }

run_sudo() {
  if command -v sudo >/dev/null 2>&1; then
    sudo -n bash -lc "$*" || sudo bash -lc "$*"
  else
    bash -lc "$*"
  fi
}

run_sudo_as_postgres() {
  local cmd="$1"
  if command -v sudo >/dev/null 2>&1; then
    sudo -n -u postgres bash -lc "$cmd" || sudo -u postgres bash -lc "$cmd"
  else
    # Fallback: try psql directly if running as postgres already
    bash -lc "$cmd"
  fi
}

ensure_psql() {
  if command -v psql >/dev/null 2>&1; then
    ok "psql found"
    return
  fi
  log "Installing PostgreSQL client and server..."
  if command -v apt-get >/dev/null 2>&1; then
    run_sudo "apt-get update"
    run_sudo "apt-get install -y postgresql postgresql-contrib curl"
  else
    err "Unsupported package manager. Please install PostgreSQL manually."
    exit 1
  fi
}

start_postgres() {
  log "Starting PostgreSQL service..."
  if command -v service >/dev/null 2>&1; then
    run_sudo "service postgresql start || true"
  else
    run_sudo "systemctl start postgresql || true"
  fi
}

setup_database() {
  log "Configuring database and user..."
  run_sudo_as_postgres "psql -tAc \"ALTER USER postgres WITH PASSWORD 'password';\"" || true

  local exists
  exists=$(run_sudo_as_postgres "psql -tAc \"SELECT 1 FROM pg_database WHERE datname='targeting_db';\"" | tr -d '[:space:]' || true)
  if [[ "$exists" != "1" ]]; then
    run_sudo_as_postgres "psql -tAc \"CREATE DATABASE targeting_db;\""
  fi

  ok "Database ready"
}

run_migrations() {
  log "Running migrations..."
  psql "$DB_URL" -f "$ROOT_DIR/db/migrations/init.sql"
  psql "$DB_URL" -f "$ROOT_DIR/db/migrations/seed.sql"
  ok "Migrations applied"
}

start_server() {
  log "Starting server in background..."
  mkdir -p "$ROOT_DIR/bin"
  (cd "$ROOT_DIR" && nohup go run ./cmd/server/main.go > "$ROOT_DIR/bin/server.log" 2>&1 & echo $! > "$SERVER_PID_FILE")

  # Wait for health
  log "Waiting for health endpoint..."
  for i in {1..30}; do
    code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/healthz || true)
    if [[ "$code" == "200" ]]; then
      ok "Server is healthy"
      return
    fi
    sleep 1
  done
  err "Server did not become healthy"
  exit 1
}

stop_server() {
  if [[ -f "$SERVER_PID_FILE" ]]; then
    local pid
    pid=$(cat "$SERVER_PID_FILE" || true)
    if [[ -n "${pid:-}" ]] && kill -0 "$pid" 2>/dev/null; then
      log "Stopping server (pid $pid)..."
      kill "$pid" || true
    fi
    rm -f "$SERVER_PID_FILE"
  fi
}

run_api_tests() {
  log "Running API tests..."
  bash "$ROOT_DIR/scripts/test-api.sh"
}

cleanup() {
  stop_server
}
trap cleanup EXIT

# Flow
ensure_psql
start_postgres
setup_database
run_migrations
start_server
run_api_tests
ok "Demo complete! Logs at $ROOT_DIR/bin/server.log"