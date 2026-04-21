#!/bin/bash
# shellcheck disable=SC2312

set -euo pipefail

# Reset stripe-mock-server state (clear all data)

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

# Check if stripe-mock is running
if ! is_stripe_mock_running; then
  log_error "Stripe-mock not running"
  exit 1
fi

PID=$(get_stripe_mock_pid)

# Check if process is still alive
if ! kill -0 "$PID" 2>/dev/null; then
  log_error "Stripe-mock process not found (PID: ${PID} died)"
  cleanup_pid .stripe-mock.pid
  exit 1
fi

log_info "Resetting stripe-mock-server..."

# Send reset request using shared helper
if reset_stripe_mock_state; then
  log_ok "Stripe-mock state reset successfully"
else
  log_error "Failed to reset stripe-mock-server"
  exit 1
fi
