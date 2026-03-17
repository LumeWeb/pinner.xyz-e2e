#!/bin/bash
# shellcheck disable=SC1091

set -euo pipefail

# Start the portal binary in background with logging
# By default, logs are written ONLY to the log file (not to the terminal).
# Use VERBOSE=1 to see output in BOTH terminal and file.
#
# Usage: ./scripts/start-portal.sh [log-path]
#
# Arguments:
#   log-path - Path to log file (default: .portal.log)
#
# Environment Variables:
#   VERBOSE     - Set to 1 to also output to terminal (default: silent, logs to file only)
#   LOGFILE     - Path to log file (overrides positional argument)
#   PORTAL_PORT - Port to run portal on (default: 8080)

# Import shared utility functions
# shellcheck disable=SC1091
. scripts/lib.sh

# Import environment for portal (this will also make PORTAL_PORT available)
load_portal_env quiet

# Use PORTAL_PORT from environment or default to 8080
PORT="${PORTAL_PORT:-8080}"

# Log file path
LOG_PATH=$(setup_log_path ".portal.log" LOGFILE "$1")
# shellcheck disable=SC1091
. scripts/lib.sh

# Check if portal is already running
if is_process_running .portal.pid; then
  log_info "Portal already running (PID: $(cat .portal.pid))"
  exit 0
fi

# Remove stale PID file if present
if [ -f .portal.pid ]; then
  log_warn "Removing stale portal PID file"
  rm -f .portal.pid
fi

# Clean up stale portal configuration
cleanup_portal_config

log_info "Copying portal binary..."
cp ./dist/portal ./portal
chmod +x ./portal

log_info "Waiting for services..."
./scripts/wait-mysql.sh
./scripts/wait-gofakes3.sh
./scripts/wait-ipfs.sh

log_info "Starting portal..."

# Start portal in background with logging
if [ "${VERBOSE:-0}" = "1" ]; then
  # Log to both file and terminal (verbose mode)
  PORTAL_PORT="${PORT}" ./portal > >(tee "${LOG_PATH}") 2>&1 &
else
  # Silent mode - only log to file (default)
  PORTAL_PORT="${PORT}" ./portal >"${LOG_PATH}" 2>&1 &
fi

PID=$!

# Verify the process actually started
if [ -z "$PID" ] || ! kill -0 "$PID" 2>/dev/null; then
  echo "Failed to start portal" >&2
  exit 1
fi

# Save PID for later cleanup
echo "$PID" > .portal.pid

# Wait for portal to be ready (up to 10 seconds)
wait_for_health_check_with_pid "$PORT" "/health" 10 1 "$PID"
result=$?

if [ "$result" -eq 0 ]; then
  log_ok "Portal started (PID: ${PID}) on port ${PORT} logging to ${LOG_PATH}"
  exit 0
elif [ "$result" -eq 2 ]; then
  log_error "Portal process died during startup"
  exit 1
else
  log_error "Portal failed to become ready within 10 seconds"
  exit 1
fi
