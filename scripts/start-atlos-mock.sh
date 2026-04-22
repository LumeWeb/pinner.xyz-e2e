#!/bin/bash
# shellcheck disable=SC2312,SC2296

set -euo pipefail

# Start atlos-sdk mock server in background with logging
# Usage: ./scripts/start-atlos-mock.sh [log-path]
#
# Arguments:
#   log-path - Path to log file (default: .atlos-mock.log)
#
# Environment Variables:
#   VERBOSE          - Set to 1 to also output to terminal (default: silent)
#   LOGFILE          - Path to log file (overrides positional argument)
#   ATLOS_MOCK_PORT  - Port to run on (default: 8085)

# Import shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

# Load environment for portal (needed for postback URL)
load_portal_env quiet

# Configuration
ATLOS_MOCK_PORT="${ATLOS_MOCK_PORT:-8085}"
ATLOS_MOCK_BIN="atlos-mock-server"
ATLOS_MOCK_GO_MODULE="go.lumeweb.com/atlos-sdk/cmd/atlos-mock-server"
ATLOS_MOCK_LOG=".atlos-mock.log"
ATLOS_MOCK_PID=".atlos-mock.pid"
ATLOS_SHARED_SECRET="${ATLOS_SHARED_SECRET:-test-secret}"
ATLOS_POSTBACK_MODE="${ATLOS_POSTBACK_MODE:-immediate}"

# Get postback URL for portal
# Portal webhook endpoint: /api/account/billing/webhooks/atlos
ATLOS_POSTBACK_URL="${PORTAL__CORE__URL:-http://account.localhost:8080}/api/account/billing/webhooks/atlos"

# Log file path
LOG_PATH=$(setup_log_path "$ATLOS_MOCK_LOG" LOGFILE "${1:-}")

# Initialize MOCK_BINARY variable
MOCK_BINARY=""

# Ensure binary is available (installed as 'server', but we want to call it 'atlos-mock-server')
if ! ensure_mock_binary "$ATLOS_MOCK_BIN" "$ATLOS_MOCK_GO_MODULE" "$ATLOS_MOCK_BIN" "server"; then
  exit 1
fi

# Check if already running
if is_process_running "$ATLOS_MOCK_PID"; then
  log_info "Atlos-mock already running (PID: $(cat "$ATLOS_MOCK_PID"))"
  exit 0
fi

# Build command args
log_info "Postback URL: $ATLOS_POSTBACK_URL"
log_info "Postback mode: $ATLOS_POSTBACK_MODE"

ATLOS_MOCK_ARGS=(
  "--port" "$ATLOS_MOCK_PORT"
  "--postback-url" "$ATLOS_POSTBACK_URL"
  "--shared-secret" "$ATLOS_SHARED_SECRET"
  "--postback-mode" "$ATLOS_POSTBACK_MODE"
)
if [ "${VERBOSE:-0}" = "1" ]; then
  ATLOS_MOCK_ARGS+=("--verbose")
fi

# Start the server
MOCK_SERVER_PID=""
if ! start_mock_server "atlos-mock" "$MOCK_BINARY" "$LOG_PATH" "$ATLOS_MOCK_PID" "$ATLOS_MOCK_PORT" \
    --args "${ATLOS_MOCK_ARGS[@]}"; then
  exit 1
fi

# Wait for server to be ready
if ! wait_for_mock_ready "atlos-mock" "$MOCK_SERVER_PID" "$LOG_PATH" 30 atlos_mock_health_check ATLOS_MOCK_WAIT_TIMEOUT; then
  kill "$MOCK_SERVER_PID" 2>/dev/null || true
  exit 1
fi
