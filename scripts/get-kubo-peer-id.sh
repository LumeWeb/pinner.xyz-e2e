#!/bin/bash

# get-kubo-peer-id.sh - Retrieve Kubo peer ID and generate IPFS bootstrap addresses
# This script must be run AFTER kubo container is started

set -euo pipefail

CONTAINER_NAME="${KUBO_CONTAINER_NAME:-portal-ipfs}"
KUBO_NETWORK="${KUBO_NETWORK:-pinnerxyz-e2e_e2e-network}"

echo "Retrieving Kubo peer ID from container: $CONTAINER_NAME"

# Check if container is running
if ! docker ps --filter "name=$CONTAINER_NAME" --format '{{.Status}}' | grep -q "Up"; then
    echo "Error: Kubo container '$CONTAINER_NAME' is not running"
    echo "Please ensure 'make up' has been executed"
    exit 1
fi

# Get the peer ID
PEER_ID=$(docker exec "$CONTAINER_NAME" ipfs id -f '<id>')
if [ -z "$PEER_ID" ]; then
    echo "Error: Failed to retrieve peer ID from Kubo container"
    exit 1
fi

echo "Retrieved peer ID: $PEER_ID"

# Get the container IP address
# Use docker network inspect to get the IP in the e2e-network
CONTAINER_IP=$(docker network inspect "$KUBO_NETWORK" --format '{{range .Containers}}{{if eq .Name "'"$CONTAINER_NAME"'"}}{{.IPv4Address}}{{end}}{{end}}' 2>/dev/null | cut -d'/' -f1)

# Fallback to localhost if network inspection fails (for testing scenarios)
if [ -z "$CONTAINER_IP" ]; then
    echo "Warning: Could not retrieve container IP, using localhost"
    CONTAINER_IP="127.0.0.1"
fi

echo "Container IP: $CONTAINER_IP"

# Generate bootstrap addresses (with p2p notation which includes peer ID)
BOOTSTRAP_TCP="/ip4/$CONTAINER_IP/tcp/4001/p2p/$PEER_ID"
BOOTSTRAP_UDP="/ip4/$CONTAINER_IP/udp/4001/p2p/$PEER_ID"

# Output in format suitable for parsing
echo "BOOTSTRAP_TCP=$BOOTSTRAP_TCP"
echo "BOOTSTRAP_UDP=$BOOTSTRAP_UDP"
