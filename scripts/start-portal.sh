#!/bin/bash
# shellcheck disable=SC1091

set -euo pipefail

# Start the portal binary in background with logging
# Usage: ./scripts/start-portal.sh [log-path]
#
# Arguments:
#   log-path - Path to log file (default: .portal.log)
#
# Environment Variables:
#   QUIET       - Set to 1 for silent mode (no stdout/stderr from portal)
#   LOGFILE     - Path to log file (overrides positional argument)
#   PORTAL_PORT - Port to run portal on (default: 8080)

# Log file path
LOG_PATH="${1:-.portal.log}"

# Use PORTAL_PORT from environment or default to 8080
PORT="${PORTAL_PORT:-8080}"

# Import environment for portal (this will also make PORTAL_PORT available)
# shellcheck disable=SC1091
set -a
. scripts/load-env.sh
set +a

# Re-read PORTAL_PORT after loading env (env takes priority)
PORT="${PORTAL_PORT:-8080}"

# Import shared utility functions
# shellcheck disable=SC1091
. scripts/lib.sh

# Clean up stale portal configuration
cleanup_portal_config

# Start portal in background with logging
if [ "${QUIET:-0}" = "1" ]; then
  # Silent mode - only log to file
  PORTAL_PORT="${PORT}" ./portal >"${LOG_PATH}" 2>&1 &
else
  # Log to both file and terminal
  PORTAL_PORT="${PORT}" ./portal > >(tee "${LOG_PATH}") 2>&1 &
fi

PID=$!

# Save PID for later cleanup
echo "$PID" > .portal.pid

echo "Portal started (PID: ${PID}) on port ${PORT} logging to ${LOG_PATH}"
