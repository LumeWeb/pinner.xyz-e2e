#!/bin/bash
# Wait for portal process to stop
# Usage: ./scripts/wait-stop-portal.sh [timeout_seconds]

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

TIMEOUT="${1:-10}"

if ! is_process_running .portal.pid; then
  log_ok "Portal is not running"
  exit 0
fi

PID=$(cat .portal.pid)

stop_process "$PID" "$TIMEOUT" "Portal"
cleanup_pid .portal.pid

log_ok "Portal stopped"
exit 0
