#!/bin/bash
# shellcheck disable=SC1091

set -euo pipefail

# Run the built portal binary and verify it starts successfully
# Usage: ./scripts/run-portal.sh [--workflow-mode]

WORKFLOW_MODE="${1:-false}"

# Setup environment
if [ "$WORKFLOW_MODE" = "true" ]; then
  # GitHub Actions mode: download artifacts first
  echo "Downloading build artifacts..."
  # This is handled by the GitHub Actions workflow
fi

# Make portal executable
chmod +x ./portal

# Run the built portal binary with environment variables
# shellcheck disable=SC1091
. .env

# Run portal in background
./portal &
PID=$!

# Wait for port binding
TIMEOUT=30
count=0
while [ $count -lt $TIMEOUT ]; do
  if nc -z localhost "${PORTAL_PORT}" 2>/dev/null; then
    echo "✓ Portal successfully bound to port ${PORTAL_PORT}"
    break
  fi
  sleep 2
  count=$((count + 2))
done

# Check if we timed out
if [ $count -ge $TIMEOUT ]; then
  echo "✗ Portal failed to bind to port ${PORTAL_PORT} within ${TIMEOUT}s"
  kill -KILL $PID 2>/dev/null || true
  exit 1
fi

# Graceful shutdown
kill -TERM "$PID" 2>/dev/null || true
wait $PID
exit $?
