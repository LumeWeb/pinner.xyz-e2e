#!/bin/bash

# wait-ipfs.sh - Wait for IPFS service to be ready
# Usage: ./scripts/wait-ipfs.sh [timeout_seconds]
# Defaults to 30 seconds if not specified

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

TIMEOUT=$(get_timeout 30 IPFS_WAIT_TIMEOUT "${1:-}")
IPFS_ENDPOINT="${IPFS_ENDPOINT:-http://localhost:5001}"
IPFS_CONTAINER="${IPFS_CONTAINER:-portal-ipfs}"

log_info "Waiting for IPFS service at $IPFS_ENDPOINT to be ready..."
log_info "Using docker IPFS container ($IPFS_CONTAINER)"

if wait_with_timeout "$TIMEOUT" "docker exec $IPFS_CONTAINER ipfs id > /dev/null 2>&1"; then
  log_ok "IPFS service is ready"
  exit 0
else
  log_error "IPFS service did not become ready within ${TIMEOUT}s"
  exit 1
fi
