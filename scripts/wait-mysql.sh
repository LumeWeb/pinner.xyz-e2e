#!/bin/bash
# Wait for MySQL to be ready for connections
# Usage: ./scripts/wait-mysql.sh [timeout_seconds]
#
# Environment Variables:
#   MYSQL_WAIT_TIMEOUT - Timeout in seconds (default: 30)
#
# Reads credentials from .env file (PORTAL__CORE__DB__USERNAME and PORTAL__CORE__DB__PASSWORD)
#
# Works in both environments:
# - GitHub Actions: Uses MySQL service (accessible via localhost)
# - Local: Uses docker compose services

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Use environment variable or positional argument or default to 30
TIMEOUT="${MYSQL_WAIT_TIMEOUT:-${1:-30}}"

log_info "Waiting for MySQL to be ready for connections..."

# Import environment to get portal configuration
# shellcheck disable=SC1091
set -a
. scripts/load-env.sh
set +a

# Extract DB credentials from PORTAL env vars
MYSQL_USER="${PORTAL__CORE__DB__USERNAME}"
MYSQL_PASSWORD="${PORTAL__CORE__DB__PASSWORD}"
MYSQL_HOST="${PORTAL__CORE__DB__HOST}"

# Detect environment and use appropriate MySQL command
if [ "${GITHUB_ACTIONS:-}" = "true" ]; then
  # GitHub Actions: MySQL service is accessible via localhost
  log_info "Using GitHub Actions MySQL service"
  if wait_with_timeout "$TIMEOUT" "mysqladmin ping -h localhost -u $MYSQL_USER -p$MYSQL_PASSWORD --silent"; then
    log_ok "MySQL is ready"
    exit 0
  else
    log_error "MySQL readiness check failed or timeout"
    exit 1
  fi
else
  # Local: Use docker compose
  log_info "Using docker compose MySQL service"
  if wait_with_timeout "$TIMEOUT" "docker compose exec -T -e MYSQL_PWD=${MYSQL_PASSWORD} mysql mysqladmin ping -h ${MYSQL_HOST} -u ${MYSQL_USER} --silent"; then
    log_ok "MySQL is ready"
    exit 0
  else
    log_error "MySQL readiness check failed or timeout"
    exit 1
  fi
fi
