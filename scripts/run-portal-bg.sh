#!/bin/bash
# shellcheck disable=SC1091

set -euo pipefail

# Run portal in background and capture PID
# Usage: ./scripts/run-portal-bg.sh [port]

PORT="${1:-8080}"

echo "Starting portal on port ${PORT}..."

# Source environment
if [ -f .env ]; then
  # shellcheck disable=SC1091
  . .env
fi

# Run portal in background
PORTAL_PORT="${PORT}" ./portal &
PID=$!

# Save PID
echo "$PID" > .portal.pid

echo "Portal started (PID: ${PID})"
