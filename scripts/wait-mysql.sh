#!/bin/bash
# Wait for MySQL to be ready for connections
# Usage: ./scripts/wait-mysql.sh [timeout_seconds]
#
# Environment Variables:
#   MYSQL_WAIT_TIMEOUT - Timeout in seconds (default: 30)
#
# Reads credentials from .env file (PORTAL__CORE__DB__USERNAME and PORTAL__CORE__DB__PASSWORD)
#
# Works with docker containers (accessible via docker exec)

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

TIMEOUT=$(get_timeout 30 MYSQL_WAIT_TIMEOUT "${1:-}")

log_info "Waiting for MySQL to be ready for connections..."

# Import environment to get portal configuration
load_portal_env quiet

# Extract DB credentials from PORTAL env vars
MYSQL_USER="${PORTAL__CORE__DB__USERNAME}"
MYSQL_PASSWORD="${PORTAL__CORE__DB__PASSWORD}"
MYSQL_HOST="${PORTAL__CORE__DB__HOST}"

log_info "Using docker MySQL container"
export MYSQL_PWD="$MYSQL_PASSWORD"
# Detect MySQL container name (local: portal-mysql, CI: mysql)
MYSQL_CONTAINER="portal-mysql"
if ! docker inspect "$MYSQL_CONTAINER" >/dev/null 2>&1; then
  MYSQL_CONTAINER="mysql"
fi

if wait_with_timeout "$TIMEOUT" "docker exec $MYSQL_CONTAINER mysqladmin ping -h 127.0.0.1 -u ${MYSQL_USER} --silent"; then
  log_ok "MySQL is ready"
  exit 0
else
  log_error "MySQL readiness check failed or timeout"
  exit 1
fi
