#!/bin/bash
# shellcheck disable=SC2312

set -euo pipefail

# Register webhook endpoint with stripe-mock-server

# Usage: ./scripts/setup-stripe-webhook.sh <webhook_url> [api_key]
# Example: ./scripts/setup-stripe-webhook.sh "http://localhost:8080/api/account/billing/webhooks/stripe" "sk_test_mock"

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

# Check arguments
if [ $# -lt 1 ]; then
  log_error "Usage: $0 <webhook_url> [api_key]"
  exit 1
fi

WEBHOOK_URL="$1"
API_KEY="${2:-$(get_stripe_api_key)}"

# Validate webhook URL
if [ -z "$WEBHOOK_URL" ]; then
  log_error "Webhook URL is required"
  exit 1
fi

# Validate API key
if [ -z "$API_KEY" ]; then
  log_error "API key is required"
  exit 1
fi

log_info "Registering webhook endpoint at ${WEBHOOK_URL}..."

# Register webhook endpoint with stripe-mock
# Uses localhost URL since DNS is not spoofed in bash
STRIPE_MOCK_URL="$(get_stripe_mock_url)"

# Debug: show the request details
log_info "POST ${STRIPE_MOCK_URL}/v1/webhook_endpoints"
log_info "  url=${WEBHOOK_URL}"

# Register webhook endpoint with stripe-mock
RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${STRIPE_MOCK_URL}/v1/webhook_endpoints" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "url=${WEBHOOK_URL}" \
  -d "enabled_events[]=*" \
  -d "description=E2E test webhook endpoint" 2>&1)

# Extract HTTP status and body
HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d':' -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')

log_info "Response status: ${HTTP_STATUS}"
log_info "Response body: ${BODY}"

if [ -z "$BODY" ] || [ "$HTTP_STATUS" != "200" ]; then
  log_error "Failed to register webhook endpoint (status: ${HTTP_STATUS})"
  log_error "Response: ${BODY}"
  exit 1
fi

RESPONSE="$BODY"

# Extract webhook secret
SECRET=$(echo "$RESPONSE" | grep -o '"secret":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SECRET" ]; then
  log_error "Webhook secret not found in response"
  log_error "Response: $RESPONSE"
  exit 1
fi

log_ok "Webhook registered successfully"
log_info "Webhook secret: ${SECRET}"

# Persist webhook secret to dot file for restart durability
# This file is cleaned up on stop/teardown/clean
echo "${SECRET}" > .stripe-webhook-secret

# Export webhook secret for use in .env
export_env .env PORTAL__PLUGIN__BILLING__SERVICE__BILLING__STRIPE__WEBHOOK_SECRET "${SECRET}"

# Output the secret for immediate use
echo "STRIPE_WEBHOOK_SECRET=\"${SECRET}\""

log_info "Webhook secret exported to PORTAL__PLUGIN__BILLING__SERVICE__BILLING__STRIPE__WEBHOOK_SECRET and .env"