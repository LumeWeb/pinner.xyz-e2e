#!/bin/bash
# shellcheck disable=SC2312
set -euo pipefail

# Start dynamic DNS server

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

# Check if DNS server is already running
if is_process_running .dns.pid; then
  log_info "DNS server already running (PID: $(cat .dns.pid))"
  exit 0
fi

# Remove stale PID file if present
if [ -f .dns.pid ]; then
  log_warn "Removing stale PID file"
  rm -f .dns.pid
fi

# Start DNS server
log_info "Starting DNS server..."
.venv/bin/python scripts/dns-dev-server.py > .dns.log 2>&1 &
echo $! > .dns.pid

# Health check - wait for DNS server to be ready
log_info "Waiting for DNS server to be ready..."
if wait_with_timeout 30 "dig @127.0.0.1 -p 5353 account.localhost +short"; then
  log_ok "DNS server started (PID: $(cat .dns.pid), log: .dns.log)"
  exit 0
else
  log_error "DNS server failed to start within 30 seconds"
  exit 1
fi

