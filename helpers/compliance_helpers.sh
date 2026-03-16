#!/bin/bash
# shellcheck disable=SC2016
# Compliance test helper functions
# Provides utilities for IPFS spec compliance testing

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# get_ipfs_endpoint_from_env - extracts IPFS endpoint from portal config
# Returns the formatted endpoint URL for the IPFS pinning service
# Format: http://ipfs.{domain}:{port}
get_ipfs_endpoint_from_env() {
	local domain="${PORTAL__CORE__DOMAIN:-localhost}"
	local port="${PORTAL__CORE__PORT:-8080}"
	local secure="${PORTAL__CORE__SECURE:-false}"
	
	local protocol="http"
	if [ "$secure" = "true" ] || [ "$secure" = "True" ]; then
		protocol="https"
	fi
	
	echo "${protocol}://ipfs.${domain}:${port}"
}

# register_compliance_user - registers a new compliance test user
# Usage: register_compliance_user <email> <password>
# Returns: 0 on success, 1 on failure
# Note: Creates user account but does NOT return JWT token
register_compliance_user() {
	local email="$1"
	local password="$2"
	local port="${PORTAL__CORE__PORT:-8080}"

	# Build JSON request body
	local request_body
	request_body=$(json_build \
		--arg email "$email" \
		--arg password "$password" \
		--arg first_name "Compliance" \
		--arg last_name "Test" \
		'{
			email: $email,
			first_name: $first_name,
			last_name: $last_name,
			password: $password
		}'
	)

	# Make API call to register user
	local response
	response=$(api_call_account POST "$port" "/api/auth/register" "$request_body")

	# Check if response contains an error field
	local error
	error=$(json_extract "$response" '.error')
	if [ -n "$error" ]; then
		return 1
	fi

	return 0
}

# login_compliance_user - logs in a user to obtain JWT token
# Usage: login_compliance_user <email> <password>
# Returns: JWT token (or empty string on failure)
login_compliance_user() {
	local email="$1"
	local password="$2"
	local port="${PORTAL__CORE__PORT:-8080}"

	# Build JSON request body
	local request_body
	request_body=$(json_build \
		--arg email "$email" \
		--arg password "$password" \
		'{
			email: $email,
			password: $password
		}'
	)

	# Step 1: POST /api/auth/login - get 302 redirect with Location header containing auth_token
	local headers
	headers=$(api_call_account_headers POST "$port" "/api/auth/login" "$request_body")

	# Extract auth_token from Location header
	local auth_token
	auth_token=$(echo "$headers" | grep -i "Location:" | sed 's/.*auth_token=//' | tr -d '\r')

	if [ -z "$auth_token" ]; then
		log_error "No auth_token found in login response"
		return 1
	fi

	log_info "Auth token received: ${auth_token:0:20}..."

	# Step 2: GET /api/auth/complete?auth_token=XXX - get JWT token
	local complete_response
	complete_response=$(api_call_account GET "$port" "/api/auth/complete?auth_token=${auth_token}" "")

	# Extract and return JWT token from response
	local jwt_token
	jwt_token=$(json_extract "$complete_response" '.token')

	if is_valid_token "$jwt_token"; then
		printf '%s' "$jwt_token"
	else
		log_error "Failed to extract JWT token from response"
		return 1
	fi
}

# create_api_key - creates an API key for the given JWT token
# Usage: create_api_key <jwt_token> <key_name>
# Returns: API key (or empty string on failure)
create_api_key() {
	local jwt_token="$1"
	local key_name="$2"
	local port="${PORTAL__CORE__PORT:-8080}"

	# Build JSON request body
	local request_body
	request_body=$(json_build \
		--arg name "$key_name" \
		'{name: $name}'
	)

	# Make API call to create API key
	local response
	response=$(api_call_account \
		POST \
		"$port" \
		"/api/account/keys" \
		"$request_body" \
		"Authorization: Bearer ${jwt_token}"
	)

	# Extract API key from response (CreateAPIKeyResponse has .token field)
	local api_key
	api_key=$(json_extract "$response" '.token')

	if is_valid_token "$api_key"; then
		printf '%s' "$api_key"
	fi
}

# parse_compliance_report_exit - handles compliance report from npm package
# Usage: parse_compliance_report_exit <exit_code> [report_file]
# Returns: 0 if we should exit 0, 1 if we should exit with error code
parse_compliance_report_exit() {
	local exit_code="$1"
	local report_file="${2:-}"

	# If exit code is 0, success
	if [ "$exit_code" -eq 0 ]; then
		return 0
	fi

	# If we have a report file, try to parse it for better error reporting
	if [ -n "$report_file" ] && [ -f "$report_file" ]; then
		# Try to parse JSON report to see if we got partial success
		if command -v jq &> /dev/null; then
			local success
			local failed

			success=$(jq -r '.success // false' "$report_file" 2>/dev/null || echo "false")
			failed=$(jq -r '.failed // 0' "$report_file" 2>/dev/null || echo "0")

			# Check if all tests passed despite non-zero exit code
			if [ "$success" = "true" ] && [ "$failed" -eq 0 ]; then
				return 0
			fi
		fi
	fi

	# Exit with the original error code
	return "$exit_code"
}