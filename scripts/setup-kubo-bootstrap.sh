#!/bin/bash
# shellcheck disable=SC1091

# setup-kubo-bootstrap.sh - Add portal to kubo's bootstrap list
# This script should be run after both kubo and portal are started

set -e

# Load shared configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
# shellcheck source=scripts/config.sh
source "${SCRIPT_DIR}/config.sh"

# Load environment variables from .env
set -a
. scripts/load-env.sh
set +a

IPFS_API_ENDPOINT="${IPFS_API_ENDPOINT:-http://localhost:5001}"
PORTAL_HOST="${PORTAL_HOST:-localhost:8080}"

echo "Setting up kubo bootstrap configuration..."
echo "IPFS API: $IPFS_API_ENDPOINT"
echo "Portal: $PORTAL_HOST"

# Wait for portal to be ready
echo "Waiting for portal to be ready..."
if ! ./scripts/wait-portal.sh; then
    echo "✗ Portal health check failed"
    exit 1
fi

# Add portal peer to kubo's bootstrap list
echo "Adding portal to kubo bootstrap list..."

# Detect if running in CI (GitHub Actions) or locally
if [ -n "${GITHUB_ACTIONS:-}" ]; then
    # In CI, use the docker bridge gateway
    portal_ip="172.17.0.1"
else
    # In local dev, inspect docker network to get gateway IP
    portal_ip=$(docker network inspect bridge --format '{{range .IPAM.Config}}{{.Gateway}}{{end}}' 2>/dev/null || echo "172.17.0.1")
fi

# Find the kubo container
# Try named container first, then find by image if not found
if docker ps --format '{{.Names}}' | grep -q "^portal-ipfs$"; then
    IPFS_CONTAINER="portal-ipfs"
else
    # Find kubo container by image name
    IPFS_CONTAINER=$(docker ps --filter "ancestor=ipfs/kubo:latest" --format "{{.Names}}" | head -1 || true)
fi

if [ -z "$IPFS_CONTAINER" ]; then
    echo "✗ Could not find running IPFS/Kubo container"
    exit 1
fi

echo "Found IPFS container: $IPFS_CONTAINER"

# Add portal peer to kubo bootstrap
# Portal uses port 4002 to avoid conflict with Kubo's port 4001
# Use deterministic portal peer ID from environment variable or shared config
PORTAL_IPFS_PEER_ID="${PORTAL_IPFS_PEER_ID:-$DEFAULT_PORTAL_IPFS_PEER_ID}"
echo "Using portal peer ID: $PORTAL_IPFS_PEER_ID"

docker exec "$IPFS_CONTAINER" ipfs bootstrap add "/ip4/$portal_ip/tcp/4002/p2p/$PORTAL_IPFS_PEER_ID" > /dev/null 2>&1 || \
docker exec "$IPFS_CONTAINER" ipfs bootstrap add "/ip4/$portal_ip/udp/4002/p2p/$PORTAL_IPFS_PEER_ID" > /dev/null 2>&1 || true

# Restart kubo to apply bootstrap changes
echo "Restarting kubo to apply bootstrap configuration..."
docker restart "$IPFS_CONTAINER" > /dev/null 2>&1

# Wait for kubo to be ready after restart
echo "Waiting for kubo to restart..."
sleep 5

# Verify the bootstrap list
echo "Verifying kubo bootstrap list..."
bootstrap_list=$(docker exec "$IPFS_CONTAINER" ipfs bootstrap list)
echo "Bootstrap list: $bootstrap_list"

# Check if portal IP is in the bootstrap list
if echo "$bootstrap_list" | grep -q "$portal_ip"; then
    echo "✓ Successfully added portal to kubo bootstrap list"
else
    echo "⚠ Failed to find portal in bootstrap list, but continuing..."
fi

echo "Kubo bootstrap setup complete"
