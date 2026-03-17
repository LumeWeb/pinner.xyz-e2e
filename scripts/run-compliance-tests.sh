#!/bin/bash
# IPFS Compliance Test Runner
# Runs @ipfs-shipyard/pinning-service-compliance npm package against the portal
# Usage: ./scripts/run-compliance-tests.sh [options]
#
#   Options:
#     -v, --verbose  Enable verbose output
#     -d, --debug    Enable debug output
#
# Environment Variables:
#   PORTAL_PORT         Portal HTTP port (default: 8080)
#   COMPLIANCE_DEBUG    Enable debug mode (default: false)
#   NO_COMPLIANCE       Skip compliance tests if set to "true"

set -euo pipefail

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Source compliance helpers
# shellcheck disable=SC1091
. helpers/compliance_helpers.sh

# Script configuration
COMPLIANCE_PACKAGE="@ipfs-shipyard/pinning-service-compliance"
DEFAULT_PASSWORD="Test123!"

# Show usage
usage() {
	echo "Usage: $0 [options]"
	echo ""
	echo "Options:"
	echo "  -v, --verbose   Enable verbose output"
	echo "  -d, --debug     Enable debug output"
	echo "  -h, --help      Show this help message"
	echo ""
	echo "Environment Variables:"
	echo "  PORTAL_PORT       Portal HTTP port (default: 8080)"
	echo "  COMPLIANCE_DEBUG  Enable debug mode (default: false)"
	echo "  NO_COMPLIANCE     Skip if set to 'true'"
}

# Parse command line arguments
VERBOSE=false
DEBUG=false

while [[ $# -gt 0 ]]; do
	case $1 in
		-v|--verbose)
			VERBOSE=true
			VERBOSE_FLAG="-v"
			shift
			;;
		-d|--debug)
			DEBUG=true
			DEBUG_FLAG="-d"
			shift
			;;
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

# Check if compliance should be skipped
if [ "${NO_COMPLIANCE:-}" = "true" ]; then
	log_info "Compliance tests skipped (NO_COMPLIANCE=true)"
	exit 0
fi

log_info "Running IPFS compliance tests..."

# Set up cleanup trap for temp directory (function will be called on exit)
# Note: COMPLIANCE_REPORT_DIR and IS_TEMP_DIR are set later in the script
# shellcheck disable=SC2154,SC2317  # Variables are defined later, function invoked via trap
cleanup_temp_dir() {
	if [[ "${IS_TEMP_DIR:-0}" == "1" ]] && [[ -n "${COMPLIANCE_REPORT_DIR:-}" ]]; then
		log_info "Cleaning up temporary directory: $COMPLIANCE_REPORT_DIR"
		rm -rf "$COMPLIANCE_REPORT_DIR"
	fi
}
trap cleanup_temp_dir EXIT

# Verify compliance package is installed
PACKAGE_PATH=$(get_npm_package_path "$COMPLIANCE_PACKAGE")

if [ -z "$PACKAGE_PATH" ]; then
	log_error "Compliance package not installed. Please run 'make setup-compliance' first."
	exit 1
fi

# Ensure portal is running
log_info "Waiting for portal to be ready..."
if ! ./scripts/wait-portal.sh; then
	log_error "Portal is not running or not ready."
	log_error "Please run 'make setup' or 'make up build-portal setup-env start-dns start-portal' first."
	exit 1
fi

log_info "Portal is ready"

# Create compliance test user with unique email
TIMESTAMP=$(date +%s)
COMPLIANCE_EMAIL="compliance-test-${TIMESTAMP}@lumeweb.com"

log_info "Creating compliance test user: $COMPLIANCE_EMAIL"

# Register user (does not return JWT token)
if ! register_compliance_user "$COMPLIANCE_EMAIL" "$DEFAULT_PASSWORD"; then
	log_error "Failed to register compliance test user"
	exit 1
fi

log_info "Compliance user registered, logging in..."

# Login to get JWT token
JWT_TOKEN=$(login_compliance_user "$COMPLIANCE_EMAIL" "$DEFAULT_PASSWORD")

if [ -z "$JWT_TOKEN" ]; then
	log_error "Failed to login to compliance test user"
	exit 1
fi

log_info "Compliance user logged in successfully"

# Create API key for compliance tests
API_KEY_NAME="compliance-test-key-${TIMESTAMP}"

log_info "Creating API key: $API_KEY_NAME"

API_KEY=$(create_api_key "$JWT_TOKEN" "$API_KEY_NAME")

if [ -z "$API_KEY" ]; then
	log_error "Failed to create API key for compliance tests"
	exit 1
fi

log_info "API key created successfully"

# Get IPFS endpoint
IPFS_ENDPOINT=$(get_ipfs_endpoint_from_env)

log_info "IPFS Pinning Service endpoint: $IPFS_ENDPOINT"

# Report what we're about to do
log_info "Running compliance tests..."
log_info "  Package: $COMPLIANCE_PACKAGE"
log_info "  Endpoint: $IPFS_ENDPOINT"
if [ "$VERBOSE" = true ]; then
	log_info "  Email: $COMPLIANCE_EMAIL"
	log_info "  API Key Name: $API_KEY_NAME"
fi

# Get output directory from package path
# Remove /src/index.js suffix and add /docs to find report location
COMPLIANCE_OUTPUT_DIR="${PACKAGE_PATH%/src/index.js}/docs"
IS_TEMP_DIR=0

if [ ! -d "$COMPLIANCE_OUTPUT_DIR" ]; then
	log_warn "Output directory does not exist: $COMPLIANCE_OUTPUT_DIR"
	COMPLIANCE_REPORT_DIR=$(mktemp -d)
	# Mark as temp dir for cleanup
	IS_TEMP_DIR=1
else
	COMPLIANCE_REPORT_DIR="$COMPLIANCE_OUTPUT_DIR"
	IS_TEMP_DIR=0
fi

log_info "Compliance output directory: $COMPLIANCE_REPORT_DIR"

# Run the compliance tests
log_info "Starting compliance test execution..."

if [ "$VERBOSE" = true ]; then
	log_info "Running in verbose mode"
fi

if [ "$DEBUG" = true ] || [ "${COMPLIANCE_DEBUG:-}" = "true" ]; then
	log_info "Running in debug mode"
	DEBUG_FLAG="-d"
fi

# Install package globally and run compliance tests
# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
DNS_PRELOAD_PATH="${SCRIPT_DIR}/dns-preload.js"

# Install package globally if not already installed
PACKAGE_PATH=$(get_npm_package_path "$COMPLIANCE_PACKAGE")

if [ -z "$PACKAGE_PATH" ]; then
	log_info "Package not found, installing globally..."
	install_npm_package_globally "$COMPLIANCE_PACKAGE"
	PACKAGE_PATH=$(get_npm_package_path "$COMPLIANCE_PACKAGE")
	
	if [ -z "$PACKAGE_PATH" ]; then
		log_error "Failed to install or locate $COMPLIANCE_PACKAGE"
		exit 1
	fi
fi

log_info "Running compliance tests from: $PACKAGE_PATH"

# Run compliance tests with DNS preload script
# Capture both stdout and stderr
# Use environment variable to avoid exposing API key in process listings
COMPLIANCE_OUTPUT=$(API_KEY="$API_KEY" node --require "$DNS_PRELOAD_PATH" "$PACKAGE_PATH" \
	-s "$IPFS_ENDPOINT" \
	${VERBOSE_FLAG:+$VERBOSE_FLAG} \
	${DEBUG_FLAG:+$DEBUG_FLAG} \
	2>&1)
COMPLIANCE_EXIT_CODE=$?

# Save output to file for debugging
echo "$COMPLIANCE_OUTPUT" > "$COMPLIANCE_REPORT_DIR/compliance-output.txt"

# Display a summary of the output
log_info "Compliance test output (last 20 lines):"
echo "$COMPLIANCE_OUTPUT" | tail -n 20

# Since portal is always localhost, use fixed report path
COMPLIANCE_REPORT_MD="${COMPLIANCE_REPORT_DIR}/ipfs.localhost.md"

log_info "Looking for compliance report at: $COMPLIANCE_REPORT_MD"

if [ -f "$COMPLIANCE_REPORT_MD" ]; then
	log_info "Found compliance report: $COMPLIANCE_REPORT_MD"
	
	# Display summary from markdown report
	log_info "Compliance test summary:"
	head -n 30 "$COMPLIANCE_REPORT_MD"
fi

# Check the exit code
if [ "$COMPLIANCE_EXIT_CODE" -eq 0 ]; then
	log_info "Compliance tests passed successfully"
	
	if [ -f "$COMPLIANCE_REPORT_MD" ]; then
		log_info "Compliance report saved to: $COMPLIANCE_REPORT_MD"
	fi
	
	exit 0
else
	log_error "Compliance tests failed with exit code: $COMPLIANCE_EXIT_CODE"
	log_error "Full output saved to: $COMPLIANCE_REPORT_DIR/compliance-output.txt"
	
	if [ -f "$COMPLIANCE_REPORT_MD" ]; then
		log_error "Compliance report: $COMPLIANCE_REPORT_MD"
	fi
	
	exit "$COMPLIANCE_EXIT_CODE"
fi
