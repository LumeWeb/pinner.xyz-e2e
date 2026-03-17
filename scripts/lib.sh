#!/bin/bash
# shellcheck disable=SC2129,SC2312
#
# Shared bash utility functions for E2E testing scripts
# Usage: source scripts/lib.sh

set -euo pipefail

# Color output support
if [ -t 1 ]; then
  COLOR_OK='\033[0;32m'
  COLOR_ERROR='\033[0;31m'
  COLOR_WARN='\033[0;33m'
  COLOR_RESET='\033[0m'
else
  COLOR_OK=''
  COLOR_ERROR=''
  COLOR_WARN=''
  COLOR_RESET=''
fi

# Logging functions
log_ok() {
  echo -e "${COLOR_OK}[OK]${COLOR_RESET} $*"
}

log_error() {
  echo -e "${COLOR_ERROR}[ERROR]${COLOR_RESET} $*" >&2
}

log_warn() {
  echo -e "${COLOR_WARN}[WARN]${COLOR_RESET} $*" >&2
}

log_info() {
  echo "$*" >&2
}

# Path resolution
# Usage: setup_project_path
# Sets SCRIPT_DIR and PROJECT_ROOT, and changes to project root
setup_project_path() {
  # Get the directory of the calling script, not lib.sh
  local caller_script="${BASH_SOURCE[1]:-${BASH_SOURCE[0]:-$0}}"
  SCRIPT_DIR="$(cd "$(dirname "$caller_script")" && pwd)"
  PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
  cd "$PROJECT_ROOT"
}

# Wait loop with timeout
# Usage: wait_with_timeout <timeout_seconds> <test_command>
# Returns: 0 on success, 1 on timeout
# Example: wait_with_timeout 30 "curl -f -s http://localhost:8080/api/meta"
wait_with_timeout() {
  local timeout="$1"
  local test_command="$2"
  local counter=0

  while ! eval "$test_command" >/dev/null 2>&1; do
    counter=$((counter + 1))
    if [ $counter -ge "${timeout}" ]; then
      return 1
    fi
    sleep 1
  done
  return 0
}

# PID file management
# Usage: pid_check <pid_file_name> <process_name>
# Sets PROCESS_PID variable, returns 0 if running, 1 if not running
pid_check() {
  local pid_file="$1"
  local process_name="$2"
  
  if [ ! -f "$pid_file" ]; then
    PROCESS_PID=""
    return 1
  fi
  
  PROCESS_PID=$(cat "$pid_file")
  
  if kill -0 "$PROCESS_PID" 2>/dev/null; then
    return 0
  else
    log_warn "$process_name process $PROCESS_PID not running"
    return 1
  fi
}

# Graceful stop with force kill fallback
# Usage: stop_process <pid> <timeout_seconds> <process_name>
stop_process() {
  local pid="$1"
  local timeout="${2:-10}"
  local process_name="${3:-Process}"
  
  if ! kill -0 "$pid" 2>/dev/null; then
    log_ok "$process_name already stopped (PID: ${pid})"
    return 0
  fi
  
  log_info "Stopping $process_name (PID: ${pid})..."
  kill "$pid" 2>/dev/null || true
  
  local counter=0
  while kill -0 "$pid" 2>/dev/null; do
    counter=$((counter + 1))
    if [ $counter -ge "${timeout}" ]; then
      log_warn "$process_name did not stop gracefully, forcing kill..."
      kill -9 "$pid" 2>/dev/null || true
      break
    fi
    sleep 1
  done
  
  return 0
}

# Ensure command exists
# Usage: require_command <command>
require_command() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    log_error "Required command not found: $cmd"
    exit 1
  fi
}

# Ensure file exists
# Usage: require_file <file_path> <error_message>
require_file() {
  local file="$1"
  local message="${2:-Required file not found: $file}"
  if [ ! -f "$file" ]; then
    log_error "$message"
    exit 1
  fi
}

# Ensure directory exists
# Usage: require_dir <dir_path> <error_message>
require_dir() {
  local dir="$1"
  local message="${2:-Required directory not found: $dir}"
  if [ ! -d "$dir" ]; then
    log_error "$message"
    exit 1
  fi
}

# Safely source environment file
# Usage: source_env_file <file_path>
# Sets allexport mode before sourcing, ensures set +a runs even on failure
# Returns: 0 on success, 1 on failure
source_env_file() {
  local file="$1"
  
  if [ ! -f "$file" ]; then
    return 1
  fi
  
  set -a
  # shellcheck disable=SC1090
  if ! . "$file"; then
    set +a
    # shellcheck disable=SC2317
    # SC2317 is safe to ignore here: this pattern handles both sourced and executed contexts
    return 1 2>/dev/null || exit 1
  fi
  set +a
  return 0
}

# Check if port is in use
# Usage: is_port_in_use <port>
is_port_in_use() {
  local port="$1"
  nc -z localhost "$port" 2>/dev/null
}

# Wait for port to be available
# Usage: wait_for_port <port> [timeout_seconds]
wait_for_port() {
  local port="$1"
  local timeout="${2:-30}"
  
  if wait_with_timeout "$timeout" "nc -z localhost $port" "Port $port not available"; then
    log_ok "Port $port is available"
    return 0
  else
    log_error "Port $port timeout"
    return 1
  fi
}

# Check if process is running by PID file
# Usage: is_process_running <pid_file>
is_process_running() {
  local pid_file="$1"
  if [ ! -f "$pid_file" ]; then
    return 1
  fi
  
  local pid
  pid=$(cat "$pid_file")
  
  if [ -z "$pid" ]; then
    return 1
  fi
  
  kill -0 "$pid" 2>/dev/null
}

# Clean up PID file
# Usage: cleanup_pid <pid_file>
cleanup_pid() {
  local pid_file="$1"
  if [ -f "$pid_file" ]; then
    rm -f "$pid_file"
  fi
}

# Clean up stale portal configuration files
# Usage: cleanup_portal_config
cleanup_portal_config() {
  log_info "Cleaning up stale configuration files..."
  local failed=0
  rm -f /etc/lumeweb/portal/core.yaml "$HOME/.lumeweb/portal/core.yaml" ./core.yaml 2>/dev/null || failed=1
  if [ $failed -eq 0 ]; then
    log_ok "Configuration files cleaned"
  else
    log_info "Configuration cleanup skipped (non-writable)"
  fi
}

# API call helpers
# Usage: api_call <method> <url> <request_body> [extra_headers]
# Returns: HTTP response
# Example: api_call POST http://localhost:8080/api/auth/register '{"email":"test@example.com"}'
api_call() {
  local method="$1"
  local url="$2"
  local body="$3"
  shift 3
  local extra_headers=("$@")

  local curl_args=(
    -s
    -X "$method"
    -H "Content-Type: application/json"
  )

  # Add extra headers if provided
  if [ ${#extra_headers[@]} -gt 0 ]; then
    for header in "${extra_headers[@]}"; do
      curl_args+=(-H "$header")
    done
  fi

  # Add body if provided
  if [ -n "$body" ]; then
    curl_args+=(-d "$body")
  fi

  curl_args+=("$url")
  curl "${curl_args[@]}"
}

# API call with headers included in response
# Usage: api_call_headers <method> <port> <path> <body> [extra_headers]
# Returns: HTTP response including headers
# Similar to api_call but includes headers using curl -i
api_call_headers() {
  local method="$1"
  local port="$2"
  local path="$3"
  local body="$4"
  shift 4
  local extra_headers=("$@")
  local url="http://localhost:${port}${path}"

  local curl_args=(
    -i
    -X "$method"
    -H "Content-Type: application/json"
  )

  # Add extra headers if provided
  if [ ${#extra_headers[@]} -gt 0 ]; then
    for header in "${extra_headers[@]}"; do
      curl_args+=(-H "$header")
    done
  fi

  # Add body if provided
  if [ -n "$body" ]; then
    curl_args+=(-d "$body")
  fi

  curl_args+=("$url")
  curl "${curl_args[@]}"
}

# API call with account vhost header
# Usage: api_call_account <method> <port> <path> <request_body> [extra_headers]
# Returns: HTTP response
# Example: api_call_account POST 8080 /api/auth/register '{"email":"test@example.com"}'
api_call_account() {
  local method="$1"
  local port="$2"
  local path="$3"
  local body="$4"
  shift 4
  local extra_headers=("$@")

  local url="http://localhost:${port}${path}"

  api_call "$method" "$url" "$body" "Host: account.localhost:${port}" "${extra_headers[@]}"
}

# API call with account vhost header, includes response headers
# Usage: api_call_account_headers <method> <port> <path> <request_body> [extra_headers]
# Returns: HTTP response including headers
# Similar to api_call_account but includes headers using curl -i
api_call_account_headers() {
  local method="$1"
  local port="$2"
  local path="$3"
  local body="$4"
  shift 4
  local extra_headers=("$@")
  local url="http://localhost:${port}${path}"

  local curl_args=(
    -i
    -X "$method"
    -H "Content-Type: application/json"
    -H "Host: account.localhost:${port}"
  )

  # Add extra headers if provided
  if [ ${#extra_headers[@]} -gt 0 ]; then
    for header in "${extra_headers[@]}"; do
      curl_args+=(-H "$header")
    done
  fi

  # Add body if provided
  if [ -n "$body" ]; then
    curl_args+=(-d "$body")
  fi

  curl_args+=("$url")
  curl "${curl_args[@]}"
}

# Check command availability
# Usage: check_command <command>
# Returns: 0 if available, 1 if not
check_command() {
  local cmd="$1"
  if command -v "$cmd" &> /dev/null; then
    return 0
  else
    return 1
  fi
}

# Require npm package is available
# Usage: require_npm_package <package_name>
# Returns: 0 if available, 1 if not
require_npm_package() {
  local package_name="$1"
  if npm list -g "$package_name" &> /dev/null; then
    return 0
  else
    return 1
  fi
}

# Build JSON object using jq
# Usage: json_build --arg name1 value1 --arg name2 value2 "{...jq_template...}"
# Returns: JSON string
# Example: json_build --arg email "test@example.com" --arg password "pass" '{email: $email, password: $password}'
json_build() {
  if [ $# -eq 0 ]; then
    log_error "json_build requires at least one argument"
    return 1
  fi
  jq -n "$@"
}

# Extract field from JSON using jq
# Usage: json_extract <json> <field>
# Returns: Field value or empty string
# Example: json_extract '{"token":"abc"}' '.token'
json_extract() {
  local json="$1"
  local field="$2"
  echo "$json" | jq -r "$field // empty"
}

# Validate token string from JSON response
# Returns: 0 if token is valid, 1 if token is null/empty/invalid
# Usage: is_valid_token "$token"
is_valid_token() {
  local token="$1"
  [ -n "$token" ] && [ "$token" != "null" ] && [ "$token" != "empty" ]
}

# Check if current user is root
# Returns: 0 if root, 1 if not
is_root() {
  [ "$(id -u)" -eq 0 ]
}

# Install npm package globally
# Usage: install_npm_package_globally <package_name>
# Returns: 0 on success, 1 on failure
install_npm_package_globally() {
  local package="$1"
  
  if check_command "npm"; then
    log_info "Installing $package globally..."
    npm install -g "$package" || return 1
  else
    log_info "npm command not found"
    return 1
  fi
}

# Get path to globally installed npm package
# Handles both root and non-root installation paths
# Usage: get_npm_package_path <package_name> [entry_point]
# Default entry_point: dist/src/index.js
# Returns: Path to package entry point or empty if not found
get_npm_package_path() {
  local package="$1"
  local entry_point="${2:-dist/src/index.js}"
  
  # Try common global installation paths
  local paths=(
    "/usr/lib/node_modules/${package}/${entry_point}"
    "/usr/local/lib/node_modules/${package}/${entry_point}"
    "$HOME/.npm-global/lib/node_modules/${package}/${entry_point}"
    "$HOME/.local/lib/node_modules/${package}/${entry_point}"
  )
  
  for path in "${paths[@]}"; do
    if [ -f "$path" ]; then
      printf '%s' "$path"
      return 0
    fi
  done
  
  return 1
}

# Environment variable priority helper with fallback
# Usage: get_timeout <default_seconds> <env_var_name> [positional_arg]
# Checks environment variable first, then positional argument, then default
# Returns: Timeout value in seconds
get_timeout() {
  local default="$1"
  local env_var_name="$2"
  local positional="${3:-}"
  
  if [ -n "$env_var_name" ] && [ -n "${!env_var_name:-}" ]; then
    echo "${!env_var_name}"
  elif [ -n "$positional" ]; then
    echo "$positional"
  else
    echo "$default"
  fi
}

# Setup log file path with environment and argument fallback
# Usage: setup_log_path <default_path> [env_var_name] [positional_arg]
# Checks environment variable first, then positional argument, then default
# Returns: Log file path
setup_log_path() {
  local default="$1"
  local env_var_name="${2:-LOGFILE}"
  local positional="${3:-}"
  
  if [ -n "${!env_var_name:-}" ]; then
    echo "${!env_var_name}"
  elif [ -n "$positional" ]; then
    echo "$positional"
  else
    echo "$default"
  fi
}

# Wait for HTTP endpoint health check
# Usage: wait_for_health_check <port> <path> <timeout> [interval]
# Returns: 0 on success, 1 on timeout
wait_for_health_check() {
  local port="$1"
  local path="${2:-/health}"
  local timeout="${3:-10}"
  local interval="${4:-1}"
  local counter=0
  
  while [ "$counter" -lt "$timeout" ]; do
    if curl -sf "http://localhost:${port}${path}" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$interval"
    counter=$((counter + interval))
  done
  return 1
}

# Load portal environment variables
# Usage: load_portal_env [quiet]
# If 'quiet' is passed as first argument, suppresses all output
# Returns: 0 on success, 1 on failure
load_portal_env() {
  local quiet="${1:-}"
  
  if [ "$quiet" = "quiet" ]; then
    set -a
    if [ -f .env ]; then
      # shellcheck disable=SC1091
      QUIET=1 . scripts/load-env.sh
    fi
    set +a
  else
    # shellcheck disable=SC1091
    . scripts/lib.sh
    log_info "Loading environment from .env..."
    if [ -f .env ]; then
      # shellcheck disable=SC1091
      . scripts/load-env.sh
      log_ok "Environment loaded successfully"
      return 0
    else
      log_error ".env file not found"
      return 1
    fi
  fi
}

# Add export line to environment file
# Usage: export_env <file> <var_name> <value>
# Safely exports variable to file with proper escaping
export_env() {
  local file="$1"
  local var_name="$2"
  local value="$3"
  
  # Escape special characters in value
  local escaped_value
  escaped_value=$(printf '%s' "$value" | sed 's/["\\]/\\&/g')
  
  echo "export ${var_name}=\"${escaped_value}\"" >> "$file"
}

# Wait for HTTP endpoint health check with pid verification
# Usage: wait_for_health_check_with_pid <port> <path> <timeout> <interval> [pid]
# If pid is provided, also checks if process is still alive
# Returns: 0 on success, 1 on timeout or process death
wait_for_health_check_with_pid() {
  local port="$1"
  local path="${2:-/health}"
  local timeout="${3:-10}"
  local interval="${4:-1}"
  local pid="${5:-}"
  local counter=0
  
  while [ "$counter" -lt "$timeout" ]; do
    # Check if provided PID is still alive
    if [ -n "$pid" ] && ! kill -0 "$pid" 2>/dev/null; then
      return 2
    fi
    if curl -sf "http://localhost:${port}${path}" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$interval"
    counter=$((counter + interval))
  done
  return 1
}
