#!/bin/bash

# wait-ipfs.sh - Wait for IPFS service to be ready
# Usage: ./scripts/wait-ipfs.sh [timeout_seconds]
# Defaults to 30 seconds if not specified

set -e

TIMEOUT="${1:-30}"
IPFS_ENDPOINT="${IPFS_ENDPOINT:-http://localhost:5001}"

echo "Waiting for IPFS service at $IPFS_ENDPOINT to be ready..."

ELAPSED=0
while [ "$ELAPSED" -lt "$TIMEOUT" ]; do
    # Check if IPFS API is responding (built-in health check uses this internally)
    if docker exec portal-ipfs ipfs id > /dev/null 2>&1; then
        echo "✓ IPFS service is ready"
        exit 0
    fi
    
    sleep 1
    ELAPSED=$((ELAPSED + 1))
    echo "   Waiting for IPFS... (${ELAPSED}/${TIMEOUT}s)"
done

echo "✗ IPFS service did not become ready within ${TIMEOUT}s"
exit 1
