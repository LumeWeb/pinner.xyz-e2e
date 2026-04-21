#!/bin/bash
# shellcheck disable=SC2312

set -euo pipefail

# Stop stripe-mock-server

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

if is_process_running .stripe-mock.pid; then
  PID=$(cat .stripe-mock.pid)
  stop_process "$PID" 10 "stripe-mock"
  cleanup_pid .stripe-mock.pid
  log_ok "Stripe-mock stopped"
else
  log_ok "Stripe-mock not running"
fi