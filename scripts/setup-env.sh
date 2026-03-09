#!/bin/bash
# shellcheck disable=SC2312,SC2296

set -euo pipefail

# Generate environment variables from YAML configurations
# Usage: ./scripts/setup-env.sh [db-type mysql|sqlite] [workflow-mode true|false]
# Environment variables expected:
#   RENTERD_URL - URL to renterd instance (e.g., http://localhost:8081)
#   RENTERD_API_PASSWORD - API password for renterd

DB_TYPE="${1:-mysql}"
WORKFLOW_MODE="${2:-false}"

# Config file locations (same for both local and GitHub Actions)
WORKFLOWS_CORE_CONFIG=".github/config/portal-core.yml"
YAML_TO_ENV_SCRIPT="scripts/yaml_to_env.py"
# Preserve RENTERD_* variables from environment or existing .env file
# Environment variables take precedence over .env file
PRESERVED_RENTERD_URL="${RENTERD_URL:-}"
PRESERVED_RENTERD_API_PASSWORD="${RENTERD_API_PASSWORD:-}"

# If not set in environment, try to load from .env file
if [ -z "${PRESERVED_RENTERD_URL}" ] || [ -z "${PRESERVED_RENTERD_API_PASSWORD}" ]; then
  if [ -f .env ]; then
    # shellcheck disable=SC1091
    . .env
    # Use environment values first, then fall back to .env values
    PRESERVED_RENTERD_URL="${RENTERD_URL:-${PRESERVED_RENTERD_URL:-}}"
    PRESERVED_RENTERD_API_PASSWORD="${RENTERD_API_PASSWORD:-${PRESERVED_RENTERD_API_PASSWORD:-}}"
  fi
fi

# Clear existing .env file
: > .env

# Merge configs using yq (later configs override earlier ones)
if [ "$DB_TYPE" = "mysql" ] && [ -f "config/portal-mysql.yml" ]; then
  # Merge portal-core.yml (base) and portal-mysql.yml (overrides)
  # Load base config first, then apply mysql config on top
  yq eval 'load("'$WORKFLOWS_CORE_CONFIG'") * .' config/portal-mysql.yml > portal-mysql.yml
fi

# Merge configs and convert to env vars
CONFIG_FILES=()

if [ "$DB_TYPE" = "mysql" ] && [ -f "portal-mysql.yml" ]; then
  CONFIG_FILES+=("portal-mysql.yml")
fi

# Convert YAML to env vars using Python script
# Use a temporary file to collect all vars, then dedupe with last-wins
TEMP_ENV=$(mktemp)
for config_file in "${CONFIG_FILES[@]}"; do
  python3 "$YAML_TO_ENV_SCRIPT" "$config_file" "$TEMP_ENV"
done

# Dynamic renterd env var overrides
# Format: ENV_VAR_NAME -> PORTAL_CONFIG_PATH
declare -A RENTERD_VARS=(
  ["RENTERD_URL"]="PORTAL__CORE__STORAGE__SIA__URL"
  ["RENTERD_API_PASSWORD"]="PORTAL__CORE__STORAGE__SIA__KEY"
)

# Loop through and set each env var
for env_var in "${!RENTERD_VARS[@]}"; do
  portal_var="${RENTERD_VARS[$env_var]}"
  
  # Get value from preserved renterd variables
  case "$env_var" in
    RENTERD_URL)
      value="${PRESERVED_RENTERD_URL:-}"
      ;;
    RENTERD_API_PASSWORD)
      value="${PRESERVED_RENTERD_API_PASSWORD:-}"
      ;;
    *)
      value=""
      ;;
  esac
  
  if [ -n "$value" ]; then
    # Escape special characters in value to prevent command injection
    escaped_value=$(printf '%s' "$value" | sed 's/["\\]/\\&/g')
    echo "export ${portal_var}=\"${escaped_value}\"" >> "$TEMP_ENV"
  else
    echo "# ${env_var} not set, using empty value" >&2
  fi
done

# Dedupe with last-wins (keep last occurrence of each var)
tac "$TEMP_ENV" | awk -F= '!seen[$1]++' | tac > .env
rm -f "$TEMP_ENV"

# Source the env file to verify using shared loader
# shellcheck disable=SC1091
QUIET=1 . scripts/load-env.sh

# GitHub Actions mode: export all PORTAL__* variables to GITHUB_ENV
# This makes them available to all subsequent steps without needing to source .env
if [ "$WORKFLOW_MODE" = "true" ]; then
  if [ -n "${GITHUB_ENV:-}" ]; then
    # Extract and export all PORTAL__* variables from .env
    # Remove 'export ' prefix and write to GITHUB_ENV
    grep '^export PORTAL__' .env | sed 's/^export //' > "$GITHUB_ENV"
    
    # Export convenience variables for backward compatibility
    echo "PORTAL_PORT=${PORTAL_PORT:-8080}" >> "$GITHUB_ENV"
    echo "PORTAL_HOST=${PORTAL_HOST:-localhost}" >> "$GITHUB_ENV"
  else
    echo "Warning: GITHUB_ENV not set, skipping export to GitHub Actions environment" >&2
  fi
fi

echo "Environment configured (DB: ${DB_TYPE}):"
cat .env
