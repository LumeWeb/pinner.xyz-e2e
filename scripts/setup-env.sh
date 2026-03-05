#!/bin/bash
# shellcheck disable=SC2312,SC2296

set -euo pipefail

# Generate environment variables from YAML configurations
# Usage: ./scripts/setup-env.sh [db-type mysql|sqlite] [workflow-mode true|false]
# Environment variables expected:
#   RENTERD_URL - URL to renterd instance (e.g., http://localhost:8081)
#   RENTERD_API_PASSWORD - API password for renterd
#   RENTERD_SEED - Seed phrase for renterd

DB_TYPE="${1:-mysql}"
WORKFLOW_MODE="${2:-false}"

# Config file locations
WORKFLOWS_CORE_CONFIG=""
YAML_TO_ENV_SCRIPT=""

if [ "$WORKFLOW_MODE" = "true" ]; then
  # GitHub Actions mode: checkout workflows repo for configs
  WORKFLOWS_CORE_CONFIG="workflows-config/.github/config/portal-core.yml"
  YAML_TO_ENV_SCRIPT="workflows-config/scripts/yaml_to_env.py"
else
  # Local mode: use local configs
  WORKFLOWS_CORE_CONFIG=".github/config/portal-core.yml"
  YAML_TO_ENV_SCRIPT="scripts/yaml_to_env.py"
fi

# Preserve RENTERD_* variables from existing .env file
if [ -f .env ]; then
  # shellcheck disable=SC1091
  . .env
  # Save renterd values
  PRESERVED_RENTERD_URL="${RENTERD_URL:-}"
  PRESERVED_RENTERD_API_PASSWORD="${RENTERD_API_PASSWORD:-}"
  PRESERVED_RENTERD_SEED="${RENTERD_SEED:-}"
else
  PRESERVED_RENTERD_URL=""
  PRESERVED_RENTERD_API_PASSWORD=""
  PRESERVED_RENTERD_SEED=""
fi

# Clear existing .env file
: > .env

# Create database config based on type
if [ "$DB_TYPE" = "mysql" ]; then
  yq -n '
    .core.db.type = "mysql" |
    .core.db.host = "127.0.0.1" |
    .core.db.port = 3306 |
    .core.db.username = "portal" |
    .core.db.password = "portal" |
    .core.db.name = "portal" |
    .core.db.charset = "utf8mb4"
  ' > portal-mysql.yml
fi

# Merge configs and convert to env vars
CONFIG_FILES=("$WORKFLOWS_CORE_CONFIG")

if [ "$DB_TYPE" = "mysql" ] && [ -f "portal-mysql.yml" ]; then
  CONFIG_FILES+=("portal-mysql.yml")
fi

# Convert YAML to env vars using Python script
for config_file in "${CONFIG_FILES[@]}"; do
  python3 "$YAML_TO_ENV_SCRIPT" "$config_file" .env
done

# Dynamic renterd env var overrides
# Format: ENV_VAR_NAME -> PORTAL_CONFIG_PATH
declare -A RENTERD_VARS=(
  ["RENTERD_URL"]="PORTAL__CORE__STORAGE__SIA__URL"
  ["RENTERD_API_PASSWORD"]="PORTAL__CORE__STORAGE__SIA__API_PASSWORD"
  ["RENTERD_SEED"]="PORTAL__CORE__STORAGE__SIA__SEED"
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
    RENTERD_SEED)
      value="${PRESERVED_RENTERD_SEED:-}"
      ;;
    *)
      value=""
      ;;
  esac
  
  if [ -n "$value" ]; then
    echo "export ${portal_var}=\"${value}\"" >> .env
  else
    echo "# ${env_var} not set, using empty value" >&2
  fi
done

# Source the env file to verify
set -a
# shellcheck disable=SC1091
. .env
set +a

# GitHub Actions mode: export to GITHUB_ENV
if [ "$WORKFLOW_MODE" = "true" ]; then
  echo "PORTAL_PORT=${PORTAL_PORT:-8080}" >> "$GITHUB_ENV"
  echo "PORTAL_HOST=${PORTAL_HOST:-localhost}" >> "$GITHUB_ENV"
fi

echo "Environment configured (DB: ${DB_TYPE}):"
cat .env
