#!/bin/bash
# shellcheck disable=SC2312

set -euo pipefail

# Stop atlos-mock-server
# Usage: ./scripts/stop-atlos-mock.sh

# Import shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

PID_FILE=".atlos-mock.pid"

if [ ! -f "$PID_FILE" ]; then
  log_info "No atlos-mock PID file found (already stopped?)"
  exit 0
fi

PID=$(cat "$PID_FILE")

if [ -z "$PID" ]; then
  log_warn "Empty PID file, cleaning up"
  rm -f "$PID_FILE"
  exit 0
fi

# Use stop_process from lib.sh for graceful shutdown
stop_process "$PID" 10 "atlos-mock"

# Clean up PID file
rm -f "$PID_FILE"

log_ok "Atlos-mock stopped"
