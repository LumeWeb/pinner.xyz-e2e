#!/bin/bash
# E2E Test Runner
# Usage: ./scripts/run-tests.sh [godog_options...]
#
# This script loads environment variables from .env and runs E2E tests using godog.
#
# IMPORTANT: Always use this script to run tests. It loads .env before execution,
# which is required for proper SDK configuration.
#
# Environment Variables:
#   TEST_DEBUG=1  Enable debug mode with Delve debugger (listens on :2345)
#
# Examples:
#   ./scripts/run-tests.sh
#   ./scripts/run-tests.sh --godog.format=pretty
#   ./scripts/run-tests.sh --godog.tags="@delete-api-key"
#   ./scripts/run-tests.sh features/account_management.feature
#
# Debug Mode:
#   TEST_DEBUG=1 ./scripts/run-tests.sh
#   Then connect with: dlv connect :2345

set -euo pipefail

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Source environment variables from .env
log_info "Loading environment from .env..."
# shellcheck disable=SC1091
if [ -f .env ]; then
    . scripts/load-env.sh
else
    log_error ".env file not found. Please run 'make setup-env' first."
    exit 1
fi

log_info "Environment loaded successfully"

# Check if tests exist
if [ ! -d "features" ]; then
    log_error "No tests found. The 'features/' directory does not exist."
    log_error "E2E test infrastructure is not yet set up. Please add your test scenarios to the features/ directory."
    exit 1
fi

# Run godog tests
log_info "Running E2E tests..."

# Check if debug mode is enabled via environment variable
if [ "${TEST_DEBUG:-0}" = "1" ]; then
  log_info "Debug mode enabled: Starting Delve debugger on :2345"
  log_info "Connect with: dlv connect :2345"
  # Run tests with Delve in headless mode
  # This allows attaching a debugger from IDE or terminal
  # Note: Use --build-flags to pass verbose mode to go test, -v is not directly supported by dlv test
  dlv test --headless --listen=:2345 --api-version=2 --build-flags="-v" -- "$@"
else
  # Run tests using go test with godog
  # Pass-through any additional arguments after -args
  go test -v -args "$@"
fi

log_info "Test execution completed"
