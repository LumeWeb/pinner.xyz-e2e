#!/bin/bash
# shellcheck disable=SC1091

set -euo pipefail

# Shared environment loader
# Sources .env.renterd first (if it exists), then .env
# Usage: . scripts/load-env.sh
#        QUIET=1 . scripts/load-env.sh  # suppress output

# Default to verbose unless QUIET is set
QUIET="${QUIET:-0}"

if [ "$QUIET" = "1" ]; then
  # Silent mode - only source files
  if [ -f .env.renterd ]; then
    # shellcheck disable=SC1091
    set -a; . .env.renterd; set +a
  fi
  if [ -f .env ]; then
    # shellcheck disable=SC1091
    set -a; . .env; set +a
  fi
else
  # Verbose mode
  echo "Loading environment variables..."
  
  if [ -f .env.renterd ]; then
    echo "  Loading .env.renterd (local renterd configuration)"
    # shellcheck disable=SC1091
    set -a; . .env.renterd; set +a
  fi
  
  if [ -f .env ]; then
    echo "  Loading .env (portal configuration)"
    # shellcheck disable=SC1091
    set -a; . .env; set +a
  else
    echo "  Warning: .env file not found"
  fi
  
  echo "[OK] Environment loaded"
fi
