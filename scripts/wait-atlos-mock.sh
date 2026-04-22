#!/bin/bash
# shellcheck disable=SC2312

set -euo pipefail

# Wait for atlos-mock to be ready
# Usage: ./scripts/wait-atlos-mock.sh [timeout_seconds]
#
# Arguments:
#   timeout_seconds - Maximum time to wait (default: 30)
#
# Environment:
#   ATLOS_MOCK_PORT - Port atlos-mock is running on (default: 8085)

# Import shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

ATLOS_MOCK_PORT="${ATLOS_MOCK_PORT:-8085}"
TIMEOUT=$(get_timeout 30 ATLOS_MOCK_WAIT_TIMEOUT "${1:-}")

log_info "Waiting for atlos-mock on port ${ATLOS_MOCK_PORT}..."

if wait_with_timeout "$TIMEOUT" "atlos_mock_health_check"; then
  log_ok "Atlos-mock is ready"
  exit 0
else
  log_error "Atlos-mock failed to become ready within ${TIMEOUT} seconds"
  exit 1
fi
