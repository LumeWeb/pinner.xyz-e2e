#!/bin/bash
# shellcheck disable=SC2312
set -euo pipefail

# Wait for PowerDNS API to be ready

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Load environment configuration
# shellcheck disable=SC1091
. scripts/load-env.sh

# Get timeout from environment or use default
DNS_WAIT_TIMEOUT="${DNS_WAIT_TIMEOUT:-30}"

# PowerDNS API configuration
DNS_API_URL="${POWERDNS_API_URL:-http://localhost:8081}"
DNS_API_KEY="${POWERDNS_API_KEY:-secret-api-key-for-testing}"

# Normalize API URL - remove trailing /api/v1 if present to avoid duplication
DNS_API_BASE="${DNS_API_URL%/api/v1}"

# Helper function to check PowerDNS API
# Function is called via wait_with_timeout string
# shellcheck disable=SC2317
check_powerdns() {
    local api_url="$1"
    local api_key="$2"
    
    # Try to connect to the PowerDNS API
    # Use curl with the API key to check server status
    curl -f -s -H "X-API-Key: ${api_key}" "${api_url}/api/v1/servers/localhost" > /dev/null 2>&1
}

log_info "Waiting for PowerDNS API at ${DNS_API_BASE}..."

if wait_with_timeout "$(get_timeout 30 DNS_WAIT_TIMEOUT)" "check_powerdns '${DNS_API_BASE}' '${DNS_API_KEY}'"; then
    log_ok "PowerDNS is ready"
    exit 0
else
    log_error "PowerDNS failed to start within ${DNS_WAIT_TIMEOUT:-30} seconds"
    exit 1
fi
