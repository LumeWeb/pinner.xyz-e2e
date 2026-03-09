#!/bin/bash
# Run tests with godog, loading environment for configuration

# Import environment variables from .env.renterd (if present) and .env
# shellcheck disable=SC1091
set -a
. scripts/load-env.sh
set +a

# Run godog tests with provided arguments
# The tests are invoked through go test with the TestMain function
go test -v ./... "$@"
