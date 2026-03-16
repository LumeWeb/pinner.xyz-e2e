#!/bin/bash
# Validate compliance test setup
# Checks that all prerequisites are met without running actual tests

set -euo pipefail

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Source compliance helpers
# shellcheck disable=SC1091
. helpers/compliance_helpers.sh

log_info "Validating IPFS compliance test setup..."

PASS_COUNT=0
FAIL_COUNT=0

check_pass() {
	log_info "✓ $1"
	PASS_COUNT=$((PASS_COUNT + 1))
}

check_fail() {
	log_error "✗ $1"
	FAIL_COUNT=$((FAIL_COUNT + 1))
}

# Check Node.js availability
log_info "Checking Node.js availability..."
if check_command node; then
	NODE_VERSION=$(node --version)
	check_pass "Node.js available: $NODE_VERSION"
else
	check_fail "Node.js not found"
fi

# Check npx availability
log_info "Checking npx availability..."
if check_command npx; then
	NPX_VERSION=$(npx --version)
	check_pass "npx available: $NPX_VERSION"
else
	check_fail "npx not found"
fi

# Check jq availability
log_info "Checking jq availability..."
if check_command jq; then
	JQ_VERSION=$(jq --version)
	check_pass "jq available: $JQ_VERSION"
else
	check_fail "jq not found"
fi

# Check npm package availability
log_info "Checking npm package availability..."
if verify_npm_package_available "@ipfs-shipyard/pinning-service-compliance"; then
	PACKAGE_VERSION=$(npm view @ipfs-shipyard/pinning-service-compliance version)
	check_pass "npm package available: @ipfs-shipyard/pinning-service-completion@$PACKAGE_VERSION"
else
	check_fail "npm package not available or unreachable"
fi

# Check script permissions
log_info "Checking script permissions..."
if [ -x "scripts/run-compliance-tests.sh" ]; then
	check_pass "run-compliance-tests.sh is executable"
else
	check_fail "run-compliance-tests.sh is not executable"
fi

if [ -x "helpers/compliance_helpers.sh" ]; then
	check_pass "compliance_helpers.sh is executable"
else
	check_fail "compliance_helpers.sh is not executable"
fi

# Check shared libraries
log_info "Checking shared libraries..."
if [ -f "scripts/lib.sh" ]; then
	check_pass "lib.sh exists"
else
	check_fail "lib.sh not found"
fi

if [ -f "helpers/compliance_helpers.sh" ]; then
	check_pass "compliance_helpers.sh exists"
else
	check_fail "compliance_helpers.sh not found"
fi

# Makefile integration
log_info "Checking Makefile integration..."
if grep -q "test-compliance:" Makefile; then
	check_pass "test-compliance target found in Makefile"
else
	check_fail "test-compliance target not found in Makefile"
fi

if grep -q "_test-compliance:" Makefile; then
	check_pass "_test-compliance target found in Makefile"
else
	check_fail "_test-compliance target not found in Makefile"
fi

# Summary
log_info ""
log_info "Validation complete: $PASS_COUNT passed, $FAIL_COUNT failed"

if [ $FAIL_COUNT -eq 0 ]; then
	log_info "✓ All checks passed! Compliance tests are ready to run."
	exit 0
else
	log_error "✗ $FAIL_COUNT check(s) failed. Please address the issues above."
	exit 1
fi
