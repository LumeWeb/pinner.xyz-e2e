#!/bin/bash
# shellcheck disable=SC2312,SC2296

set -euo pipefail

# Start stripe-mock-server in background with logging
# Must be run with sudo to bind to port 80
#
# Usage: ./scripts/start-stripe-mock.sh [log-path]
#
# Arguments:
#   log-path - Path to log file (default: .stripe-mock.log)
#
# Environment Variables:
#   VERBOSE          - Set to 1 to also output to terminal (default: silent)
#   LOGFILE          - Path to log file (overrides positional argument)
#   STRIPE_MOCK_PORT - Port to run on (default: 80)

# Import shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

# Load environment for portal (needed for webhook URL)
load_portal_env quiet

# Configuration
STRIPE_MOCK_PORT="${STRIPE_MOCK_PORT:-80}"
STRIPE_MOCK_BIN="stripe-mock-server"
STRIPE_MOCK_GO_MODULE="go.lumeweb.com/stripe-mock-server/cmd/stripe-mock-server"
STRIPE_MOCK_LOG=".stripe-mock.log"
STRIPE_MOCK_PID=".stripe-mock.pid"

# Get webhook URL for portal
# Portal webhook endpoint: /api/account/billing/webhooks/:gatewayType
STRIPE_WEBHOOK_URL="${PORTAL__CORE__URL:-http://account.localhost:8080}/api/account/billing/webhooks/stripe"

# Get Stripe API key from shared helpers
STRIPE_API_KEY="$(get_stripe_api_key)"

# Log file path
LOG_PATH=$(setup_log_path "$STRIPE_MOCK_LOG" LOGFILE "${1:-}")

# Initialize MOCK_BINARY variable
MOCK_BINARY=""

# Ensure binary is available
if ! ensure_mock_binary "$STRIPE_MOCK_BIN" "$STRIPE_MOCK_GO_MODULE" "$STRIPE_MOCK_BIN" ""; then
  exit 1
fi

# Check if already running (uses PID file via start_mock_server)
if is_process_running "$STRIPE_MOCK_PID"; then
  log_info "Stripe-mock already running (PID: $(cat "$STRIPE_MOCK_PID"))"

  # If we have a saved webhook secret, ensure it's in .env
  # (avoid re-registering which creates a new secret and invalidates the old one)
  if [ -f .stripe-webhook-secret ]; then
    SAVED_SECRET=$(cat .stripe-webhook-secret)
    if [ -n "$SAVED_SECRET" ]; then
      log_info "Restoring webhook secret from .stripe-webhook-secret"
      export_env .env PORTAL__PLUGIN__BILLING__SERVICE__BILLING__STRIPE__WEBHOOK_SECRET "$SAVED_SECRET"
      log_ok "Webhook secret restored"
    fi
  else
    # No saved secret — register webhook endpoint (idempotent but creates new secret)
    log_info "Registering webhook endpoint..."
    if ! ./scripts/setup-stripe-webhook.sh "$STRIPE_WEBHOOK_URL" "$STRIPE_API_KEY"; then
      log_warn "Failed to register webhook, but stripe-mock is running"
    fi
  fi
  exit 0
fi

# Build command args for stripe-mock
STRIPE_MOCK_ARGS=(
  "-port" "$STRIPE_MOCK_PORT"
)
if [ "${VERBOSE:-0}" = "1" ]; then
  STRIPE_MOCK_ARGS+=("-verbose")
fi

# Start the server (needs sudo for port 80)
MOCK_SERVER_PID=""
if ! start_mock_server "stripe-mock" "$MOCK_BINARY" "$LOG_PATH" "$STRIPE_MOCK_PID" "$STRIPE_MOCK_PORT" \
    --sudo \
    --args "${STRIPE_MOCK_ARGS[@]}"; then
  exit 1
fi

# Wait for server to be ready
if ! wait_for_mock_ready "stripe-mock" "$MOCK_SERVER_PID" "$LOG_PATH" 30 stripe_mock_health_check STRIPE_MOCK_WAIT_TIMEOUT; then
  sudo kill "$MOCK_SERVER_PID" 2>/dev/null || true
  exit 1
fi

# Register webhook endpoint
log_info "Registering webhook endpoint..."
if ! ./scripts/setup-stripe-webhook.sh "$STRIPE_WEBHOOK_URL" "$STRIPE_API_KEY"; then
  log_warn "Failed to register webhook, but stripe-mock is running"
fi
