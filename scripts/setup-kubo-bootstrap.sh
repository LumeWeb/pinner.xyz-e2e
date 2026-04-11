#!/bin/bash
# shellcheck disable=SC1091

# setup-kubo-bootstrap.sh - Configure kubo bootstrapping
# This script should be run after kubo is started

set -e

# Load shared configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
# shellcheck source=scripts/config.sh
source "${SCRIPT_DIR}/config.sh"

# Load environment variables from .env for PORTAL_IPFS_PEER_ID
set -a
. scripts/load-env.sh
set +a

echo "Setting up kubo bootstrap configuration..."

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

# Configure static swarm addresses to ensure consistent networking
echo "Configuring static swarm addresses..."
docker exec "$IPFS_CONTAINER" ipfs config --json Addresses.Swarm '["/ip4/0.0.0.0/tcp/4001","/ip4/0.0.0.0/udp/4001/quic-v1","/ip4/0.0.0.0/udp/4001/quic-v1/webtransport"]' > /dev/null 2>&1

# Verify static addresses are configured
echo "Verifying swarm addresses..."
swarm_addrs=$(docker exec "$IPFS_CONTAINER" ipfs config Addresses.Swarm)
echo "Current swarm addresses: $swarm_addrs"

# Configure gateway address
echo "Configuring gateway address..."
docker exec "$IPFS_CONTAINER" ipfs config --json Addresses.Gateway '"/ip4/127.0.0.1/tcp/8082"' > /dev/null 2>&1

# Verify gateway address is configured
echo "Verifying gateway address..."
gateway_addr=$(docker exec "$IPFS_CONTAINER" ipfs config Addresses.Gateway)
echo "Current gateway address: $gateway_addr"

# Add portal peer to kubo bootstrap
# Portal uses port 4002 to avoid conflict with Kubo's port 4001
# Use deterministic portal peer ID from environment variable or shared config
PORTAL_IPFS_PEER_ID="${PORTAL_IPFS_PEER_ID:-$DEFAULT_PORTAL_IPFS_PEER_ID}"
echo "Using portal peer ID: $PORTAL_IPFS_PEER_ID"

# Add portal peer to kubo bootstrap using 127.0.0.1
docker exec "$IPFS_CONTAINER" ipfs bootstrap add "/ip4/127.0.0.1/tcp/4002/p2p/$PORTAL_IPFS_PEER_ID" > /dev/null 2>&1 || true

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

# Check if portal peer is in the bootstrap list
if echo "$bootstrap_list" | grep -q "$PORTAL_IPFS_PEER_ID"; then
    echo "✓ Successfully added portal to kubo bootstrap list"
else
    echo "⚠ Failed to find portal in bootstrap list, but continuing..."
fi

echo "Kubo bootstrap setup complete"
