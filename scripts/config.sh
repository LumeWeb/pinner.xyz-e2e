#!/bin/bash
# config.sh - Shared configuration for all scripts
# This file contains common configuration values used across multiple scripts
#
# This file is sourced by setup-env.sh and setup-kubo-bootstrap.sh
# shellcheck disable=SC2034  # Variables are used by scripts that source this file

# Default IPFS peer ID for testing
# This is the same peer ID used by the Kubo container for localhost testing
DEFAULT_PORTAL_IPFS_PEER_ID="12D3KooWJbFjcpUmr4tNfpy2kCwTTZ6kq6G7iUrqNQt7yxNjSLRz"
