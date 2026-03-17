#!/bin/bash
# Wait for gofakes3 to be ready for connections on port 9000
# Usage: ./scripts/wait-gofakes3.sh [timeout_seconds]
#
# Environment Variables:
#   GOFAKES3_WAIT_TIMEOUT - Timeout in seconds (default: 30)

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

TIMEOUT=$(get_timeout 30 GOFAKES3_WAIT_TIMEOUT "$1")

log_info "Waiting for gofakes3 to be ready on port 9000..."

if wait_with_timeout "$TIMEOUT" "nc -z localhost 9000"; then
  log_ok "gofakes3 is ready"
  exit 0
else
  log_error "gofakes3 readiness check failed or timeout"
  exit 1
fi
