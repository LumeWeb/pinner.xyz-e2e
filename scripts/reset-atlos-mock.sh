#!/bin/bash
# shellcheck disable=SC2312

set -euo pipefail

# Reset atlos-mock state by restarting the server
# The atlos-sdk mock doesn't have a reset endpoint like stripe-mock,
# so we must kill and restart the process to clear state.
#
# Usage: ./scripts/reset-atlos-mock.sh

# Import shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

log_info "Resetting atlos-mock state..."

# Stop the mock if running
if [ -f .atlos-mock.pid ]; then
  ./scripts/stop-atlos-mock.sh
fi

# Clear log file if desired
if [ "${CLEAR_ATLOS_LOG:-0}" = "1" ]; then
  rm -f .atlos-mock.log
  log_info "Cleared atlos-mock log file"
fi

# Start fresh
./scripts/start-atlos-mock.sh

log_ok "Atlos-mock reset complete"
