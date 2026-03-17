#!/bin/bash
# shellcheck disable=SC2312

set -euo pipefail

# Create portal-plugins.yaml manifest for portal-builder
# Usage: ./scripts/create-plugin-manifest.sh

# Change to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Initialize manifest with portal version
yq -n '.portalVersion = "develop"' > portal-plugins.yaml

# Add plugins
yq -i '.plugins[0].module = "go.lumeweb.com/portal-plugin-ipfs" | .plugins[0].version = "develop"' portal-plugins.yaml
yq -i '.plugins[1].module = "go.lumeweb.com/portal-plugin-dashboard" | .plugins[1].version = "develop"' portal-plugins.yaml
yq -i '.plugins[2].module = "go.lumeweb.com/portal-plugin-core" | .plugins[2].version = "develop"' portal-plugins.yaml

echo "Created portal-plugins.yaml:"
cat portal-plugins.yaml
