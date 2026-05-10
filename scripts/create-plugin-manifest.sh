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
yq -i '.plugins[3].module = "go.lumeweb.com/portal-plugin-quota" | .plugins[3].version = "develop"' portal-plugins.yaml
yq -i '.plugins[4].module = "go.lumeweb.com/portal-plugin-admin" | .plugins[4].version = "develop"' portal-plugins.yaml
yq -i '.plugins[5].module = "go.lumeweb.com/portal-plugin-billing" | .plugins[5].version = "develop"' portal-plugins.yaml
yq -i '.plugins[6].module = "go.lumeweb.com/portal-plugin-sia" | .plugins[6].version = "develop"' portal-plugins.yaml

# Add Go module replacements (equivalent to go.mod replace directives)
# These are passed as --replace old=new flags to xportal build
INDEXD_REPLACEMENT="${INDEXD_REPLACEMENT:-go.lumeweb.com/indexd@v0.0.0-20260504054644-d86c0fc81134}"
yq -i ".replacements[0].old = \"go.sia.tech/indexd\" | .replacements[0].new = \"${INDEXD_REPLACEMENT}\"" portal-plugins.yaml

echo "Created portal-plugins.yaml:"
cat portal-plugins.yaml
