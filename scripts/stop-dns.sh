#!/bin/bash
# shellcheck disable=SC2312
set -euo pipefail

# Stop dynamic DNS server

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

if is_process_running .dns.pid; then
  PID=$(cat .dns.pid)
  stop_process "$PID" 10 "DNS server"
  cleanup_pid .dns.pid
  log_ok "DNS server stopped"
else
  log_ok "DNS server not running"
fi

