#!/bin/bash
# IPFS Compliance Test Setup
# Ensures npm compliance package is installed and dependencies are available
# Usage: ./scripts/setup-compliance.sh [options]
#
#   Options:
#     -v, --verbose  Enable verbose output
#
# Environment Variables:
#   COMPLIANCE_DEBUG    Enable debug mode (default: false)

set -euo pipefail

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Load environment configuration
# shellcheck disable=SC1091
. scripts/load-env.sh

# Script configuration
COMPLIANCE_PACKAGE="@ipfs-shipyard/pinning-service-compliance"

# Show usage
usage() {
	echo "Usage: $0 [options]"
	echo ""
	echo "Options:"
	echo "  -h, --help      Show this help message"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
	case $1 in
		-h|--help)
			usage
			exit 0
			;;
		*)
			log_error "Unknown option: $1"
			usage
			exit 1
			;;
	esac
done

log_info "Setting up IPFS compliance testing environment..."

# Check if Node.js is available
if ! check_command node || ! check_command npm; then
	log_error "Node.js or npm is not available. Compliance tests require Node.js."
	log_error "Please install Node.js to run compliance tests."
	exit 1
fi

log_info "Node.js found: $(node --version)"
log_info "npm found: $(npm --version)"

# Check if npm package can be fetched
if ! require_npm_package "$COMPLIANCE_PACKAGE"; then
	log_error "Could not verify npm package: $COMPLIANCE_PACKAGE"
	log_error "Please check internet connection and npm registry access."
	exit 1
fi

log_info "npm package verified: $COMPLIANCE_PACKAGE"

# Check if package is already installed
PACKAGE_PATH=$(get_npm_package_path "$COMPLIANCE_PACKAGE" || true)

if [ -z "$PACKAGE_PATH" ]; then
	# Install package globally
	log_info "Installing compliance package globally..."
	install_npm_package_globally "$COMPLIANCE_PACKAGE"

	# Verify installation
	PACKAGE_PATH=$(get_npm_package_path "$COMPLIANCE_PACKAGE" || true)

	if [ -z "$PACKAGE_PATH" ]; then
		log_error "Failed to install or locate $COMPLIANCE_PACKAGE"
		exit 1
	fi

	log_ok "Compliance package installed successfully: $PACKAGE_PATH"
else
	log_ok "Compliance package already installed: $PACKAGE_PATH"
fi

log_ok "Compliance testing environment setup complete"
