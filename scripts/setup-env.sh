#!/bin/bash
# shellcheck disable=SC2312,SC2296

set -euo pipefail

# Get script directory first (use $0 as fallback for BASH_SOURCE)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"

# Load shared library
# shellcheck source=scripts/lib.sh
source "${SCRIPT_DIR}/lib.sh"

# Load shared configuration
# shellcheck source=scripts/config.sh
source "${SCRIPT_DIR}/config.sh"

# Generate environment variables from YAML configurations
# Usage: ./scripts/setup-env.sh [db-type mysql|sqlite] [workflow-mode true|false]
# Environment variables expected:
#   RENTERD_URL - URL to renterd instance (e.g., http://localhost:8081)
#   RENTERD_API_PASSWORD - API password for renterd
#   IPFS_API_ENDPOINT - IPFS API endpoint for Kubo RPC client (default: http://127.0.0.1:5001)
#   PORTAL_IPFS_PEER_ID - Deterministic portal IPFS peer ID (from test seed)

DB_TYPE="${1:-mysql}"
WORKFLOW_MODE="${2:-false}"

# Config file locations (same for both local and GitHub Actions)
WORKFLOWS_CORE_CONFIG=".github/config/portal-core.yml"
YAML_TO_ENV_SCRIPT="scripts/yaml_to_env.py"
# Preserve RENTERD_* and IPFS_* variables from environment or existing .env file
# Environment variables take precedence over .env file
PRESERVED_RENTERD_URL="${RENTERD_URL:-}"
PRESERVED_RENTERD_API_PASSWORD="${RENTERD_API_PASSWORD:-}"
PRESERVED_IPFS_API_ENDPOINT="${IPFS_API_ENDPOINT-}"
PRESERVED_PORTAL_IPFS_PEER_ID="${PORTAL_IPFS_PEER_ID:-${DEFAULT_PORTAL_IPFS_PEER_ID}}"

# If not set in environment, try to load from .env file
if [ -z "${PRESERVED_RENTERD_URL}" ] || [ -z "${PRESERVED_RENTERD_API_PASSWORD}" ] || [ -z "${PRESERVED_IPFS_API_ENDPOINT}" ]; then
  # Source .env.renterd first, then .env (in that order)
  if [ -f .env.renterd ]; then
    # shellcheck disable=SC1091
    . .env.renterd
  fi
  if [ -f .env ]; then
    # shellcheck disable=SC1091
    . .env
  fi
  # Use environment values first, then fall back to .env values
  PRESERVED_RENTERD_URL="${RENTERD_URL:-${PRESERVED_RENTERD_URL:-}}"
  PRESERVED_RENTERD_API_PASSWORD="${RENTERD_API_PASSWORD:-${PRESERVED_RENTERD_API_PASSWORD:-}}"
  PRESERVED_IPFS_API_ENDPOINT="${IPFS_API_ENDPOINT:-http://127.0.0.1:5001}"
  PRESERVED_PORTAL_IPFS_PEER_ID="${PORTAL_IPFS_PEER_ID:-${DEFAULT_PORTAL_IPFS_PEER_ID}}"
fi

# Set default for IPFS_API_ENDPOINT if still not set
PRESERVED_IPFS_API_ENDPOINT="${PRESERVED_IPFS_API_ENDPOINT:-http://127.0.0.1:5001}"
# Set default for PORTAL_IPFS_PEER_ID if still not set
PRESERVED_PORTAL_IPFS_PEER_ID="${PRESERVED_PORTAL_IPFS_PEER_ID:-${DEFAULT_PORTAL_IPFS_PEER_ID}}"

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
    # Use export_env helper
    export_env "$TEMP_ENV" "$portal_var" "$value"
  else
    echo "# ${env_var} not set, using empty value" >&2
  fi
done


# Get Kubo peer ID and generate IPFS bootstrap addresses
# This requires the kubo container to be running (started via 'make up')
echo "Retrieving Kubo peer ID for IPFS bootstrap..."
if KUBO_PEER_ID_OUTPUT=$(./scripts/get-kubo-peer-id.sh 2>&1); then
    # Parse the output to extract bootstrap addresses
    BOOTSTRAP_TCP=$(echo "$KUBO_PEER_ID_OUTPUT" | grep '^BOOTSTRAP_TCP=' | cut -d'=' -f2-)
    BOOTSTRAP_UDP=$(echo "$KUBO_PEER_ID_OUTPUT" | grep '^BOOTSTRAP_UDP=' | cut -d'=' -f2-)
    
    if [ -n "$BOOTSTRAP_TCP" ] && [ -n "$BOOTSTRAP_UDP" ]; then
        # Set PORTAL__PLUGIN__IPFS__PROTOCOL__BOOTSTRAP_PEERS to list of bootstrap addresses
        # This overrides the hardcoded values from YAML config
        echo "Setting IPFS bootstrap peers from Kubo: $BOOTSTRAP_TCP, $BOOTSTRAP_UDP"

        # Generate env var that matches the YAML path: plugin.ipfs.protocol.bootstrap_peers
        # PORTAL__PLUGIN__IPFS__PROTOCOL__BOOTSTRAP_PEERS should be in CSV format
        # Format: PORTAL__PLUGIN__IPFS__PROTOCOL__BOOTSTRAP_PEERS="addr1,addr2"
        BOOTSTRAP_CSV="$BOOTSTRAP_TCP,$BOOTSTRAP_UDP"
        export_env "$TEMP_ENV" "PORTAL__PLUGIN__IPFS__PROTOCOL__BOOTSTRAP_PEERS" "$BOOTSTRAP_CSV"
    else
        echo "Warning: Could not retrieve Kubo bootstrap addresses, using YAML defaults" >&2
    fi
else
    echo "Warning: Failed to retrieve Kubo peer ID (container may not be running yet). Using YAML defaults." >&2
    echo "Error details: $KUBO_PEER_ID_OUTPUT" >&2
fi

# Dedupe with last-wins (keep last occurrence of each var)
tac "$TEMP_ENV" | awk -F= '!seen[$1]++' | tac > .env
rm -f "$TEMP_ENV"

# Source the env file to verify using shared loader
# shellcheck disable=SC1091
QUIET=1 . scripts/load-env.sh

# Export IPFS_API_ENDPOINT environment variable (not a PORTAL__ variable)
# This is used by the IPFS API test helpers
export_env .env IPFS_API_ENDPOINT "${PRESERVED_IPFS_API_ENDPOINT}"

# Export PORTAL_IPFS_PEER_ID environment variable for kubo bootstrap setup
export_env .env PORTAL_IPFS_PEER_ID "${PRESERVED_PORTAL_IPFS_PEER_ID}"

# GitHub Actions mode: export all PORTAL__* variables to GITHUB_ENV
# This makes them available to all subsequent steps without needing to source .env
if [ "$WORKFLOW_MODE" = "true" ]; then
  if [ -n "${GITHUB_ENV:-}" ]; then
    # Extract and export all PORTAL__* variables from .env
    # Remove 'export ' prefix and write to GITHUB_ENV
    grep '^export PORTAL__' .env | sed 's/^export //' > "$GITHUB_ENV"
    
    # Export convenience variables for backward compatibility
    {
        echo "PORTAL_PORT=${PORTAL_PORT:-8080}"
        echo "PORTAL_HOST=${PORTAL_HOST:-localhost}"
        echo "IPFS_API_ENDPOINT=${PRESERVED_IPFS_API_ENDPOINT}"
        echo "PORTAL_IPFS_PEER_ID=${PRESERVED_PORTAL_IPFS_PEER_ID}"
    } >> "$GITHUB_ENV"
  else
    echo "Warning: GITHUB_ENV not set, skipping export to GitHub Actions environment" >&2
  fi
fi

echo "Environment configured (DB: ${DB_TYPE}):"
cat .env
