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
  echo "$*"
}

# Path resolution
# Usage: setup_project_path
# Sets SCRIPT_DIR and PROJECT_ROOT, and changes to project root
setup_project_path() {
  # Get the directory of the calling script, not lib.sh
  local caller_script="${BASH_SOURCE[1]}"
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
  # Suppress stderr to handle permission errors gracefully (may not be root)
  if rm -f /etc/lumeweb/portal/core.yaml "$HOME/.lumeweb/portal/core.yaml" ./core.yaml 2>/dev/null; then
    log_ok "Configuration files cleaned"
  else
    log_info "Configuration cleanup skipped (non-writable)"
  fi
}
