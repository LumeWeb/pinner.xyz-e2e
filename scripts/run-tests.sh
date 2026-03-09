#!/bin/bash
# E2E Test Runner
# Usage: ./scripts/run-tests.sh [godog_options...]
#
# This script loads environment variables from .env and runs E2E tests using godog.
# 
# IMPORTANT: Always use this script to run tests. It loads .env before execution,
# which is required for proper SDK configuration.
#
# Examples:
#   ./scripts/run-tests.sh
#   ./scripts/run-tests.sh --godog.format=pretty
#   ./scripts/run-tests.sh --godog.tags="@delete-api-key"
#   ./scripts/run-tests.sh features/account_management.feature

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

# Check if godog is installed
if ! command -v godog &> /dev/null; then
    log_error "godog is not installed. Install it with: go install github.com/cucumber/godog/cmd/godog@latest"
    exit 1
fi

# Run tests with default godog.strict=false and pass-through any additional arguments
godog --godog.strict=false "$@"

log_info "Test execution completed"
