#!/bin/bash
# shellcheck disable=SC2312
set -euo pipefail

# Ensure Python virtual environment with dnserver is installed

# Source shared library
# shellcheck disable=SC1091
. scripts/lib.sh

# Setup project path
setup_project_path

if [ ! -d .venv ]; then
  log_info "Creating Python virtual environment..."
  python3 -m venv .venv
  log_info "Installing dnserver..."
  .venv/bin/pip install dnserver >/dev/null 2>&1
  log_ok "Virtual environment created with dnserver"
else
  log_ok "Virtual environment already exists"
fi

