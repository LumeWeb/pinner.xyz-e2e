#!/bin/bash
# shellcheck disable=SC2312

set -euo pipefail

# Wait for stripe-mock-server to be ready

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

if ! is_stripe_mock_running; then
  log_error "Stripe-mock not running (no PID file found)"
  exit 1
fi

PID=$(get_stripe_mock_pid)

# Check if process is still alive
if ! kill -0 "$PID" 2>/dev/null; then
  log_error "Stripe-mock process not found (PID: ${PID} died)"
  cleanup_pid .stripe-mock.pid
  exit 1
fi

log_info "Waiting for stripe-mock to be ready..."

STRIPE_MOCK_WAIT_TIMEOUT=$(get_timeout 30 STRIPE_MOCK_WAIT_TIMEOUT)
TIMEOUT_START=$(date +%s)

while true; do
  # Check if process is still alive
  if ! kill -0 "$PID" 2>/dev/null; then
    log_error "Stripe-mock process died (PID: ${PID})"
    cleanup_pid .stripe-mock.pid
    exit 1
  fi

  # Check if health endpoint is responding
  if stripe_mock_health_check; then
    log_ok "Stripe-mock is ready"
    exit 0
  fi

  # Check timeout
  ELAPSED=$(($(date +%s) - TIMEOUT_START))
  if [ "$ELAPSED" -ge "$STRIPE_MOCK_WAIT_TIMEOUT" ]; then
    log_error "Stripe-mock failed to become ready within ${STRIPE_MOCK_WAIT_TIMEOUT} seconds"
    if [ -f .stripe-mock.log ]; then
      log_error "Last 20 lines from .stripe-mock.log:"
      tail -20 .stripe-mock.log | sed 's/^/  /' >&2
    fi
    exit 1
  fi

  sleep 1
done
