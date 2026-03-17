#!/bin/bash
# Wait for portal HTTP endpoint to be available
# Usage: ./scripts/wait-portal.sh [timeout_seconds]
#
# Environment Variables:
#   PORTAL_PORT         - Port to check (default: 8080)
#   PORTAL_WAIT_TIMEOUT - Timeout in seconds (default: 30)

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Load environment if not already set
if [ -z "${PORTAL_PORT:-}" ]; then
  load_portal_env quiet
fi

PORT="${PORTAL_PORT:-8080}"
TIMEOUT=$(get_timeout 30 PORTAL_WAIT_TIMEOUT "$1")

log_info "Waiting for HTTP endpoint to be available (plugins may take time to load)..."

if wait_with_timeout "$TIMEOUT" "curl -f -s http://localhost:${PORT}/api/meta"; then
  log_ok "Portal health check passed"
  exit 0
else
  log_error "Portal health check failed or timeout"
  exit 1
fi
