# AGENTS.md
This file provides guidance to various AI agents when working with code in this repository.

## Project Overview

This is an **E2E testing environment** for the LumeWeb Portal application. This repository does NOT contain the portal source code itself. Instead, it provides infrastructure and automation to:

1. Run infrastructure services (MySQL, Maildev, Gofakes3) via Docker Compose
2. Build the portal application using the external `portal-builder` image
3. Configure the portal via environment variables generated from YAML configs
4. Execute end-to-end tests against the running portal

The portal application is built externally using `ghcr.io/lumeweb/portal-builder:ubuntu` with a `portal-plugins.yaml` manifest specifying plugins (ipfs, dashboard, core) at `@develop` versions.

## High-Level Architecture

### Directory Structure

- `docker-compose.yml` - Infrastructure services only (MySQL, Maildev, Gofakes3, services-ready sync container)
- `Makefile` - Local development commands mirroring the GitHub Actions workflow
- `.github/workflows/e2e-tests.yml` - CI/CD automation with parallel build and test jobs
- `.github/actions/` - Reusable composite actions that delegate to shared bash scripts
- `.github/config/portal-core.yml` - Minimal core portal configuration template
- `scripts/` - Shared bash scripts used by both Makefile and GitHub Actions
- `config/portal-mysql.yml` - Generated MySQL configuration (auto-created by setup-env.sh)

### Service Architecture

**Docker Compose Services:**
- `mysql` - Percona Server 8.4 on port 3306 with health checks
- `maildev` - Email catcher on ports 1025 (SMTP) and 1080 (Web UI)
- `gofakes3` - S3-compatible storage on port 9000
- `services-ready` - Synchronization container that depends on all services being healthy

**External Services:**
- `renterd` - External Sia storage service (NOT in docker-compose), configured via environment variables

**Portal Application:**
- Built externally using `portal-builder` image
- Runs as a binary expecting `PORTAL__*` environment variables
- NOT included in docker-compose; built and run separately

### Configuration Flow

1. **Base Config**: `.github/config/portal-core.yml` provides minimal portal settings
2. **DB Config**: `scripts/setup-env.sh` generates `portal-mysql.yml` for MySQL mode
3. **Env Generation**: `scripts/yaml_to_env.py` converts YAML to `PORTAL__*` environment variables in `.env`
4. **Renterd Override**: `RENTERD_*` env vars are preserved and mapped to `PORTAL__CORE__STORAGE__SIA__*`

**Key mappings:**
- `RENTERD_URL` → `PORTAL__CORE__STORAGE__SIA__URL`
- `RENTERD_API_PASSWORD` → `PORTAL__CORE__STORAGE__SIA__API_PASSWORD`
- `RENTERD_SEED` → `PORTAL__CORE__STORAGE__SIA__SEED`

### Shared Script Architecture

Bash scripts in `scripts/` are designed to work in both local and GitHub Actions environments:

- `scripts/setup-env.sh [mysql|sqlite] [true|false]` - Generates environment variables from YAML configs. Second parameter `true` enables GitHub Actions mode (exports to GITHUB_ENV)
- `scripts/run-portal.sh [true|false]` - Runs portal and verifies port binding. Parameter `true` enables GitHub Actions mode
- `scripts/yaml_to_env.py` - Converts YAML to `PORTAL__*` env vars using double-underscore separator for nested keys

## Common Commands

### Local Development

```bash
# Service management
make up              # Start all Docker services
make down            # Stop all Docker services
make restart         # Restart all Docker services
make logs            # View service logs
make ps              # Show running containers

# Portal build & run
make build-portal    # Build portal with plugins using portal-builder
make setup-env       # Generate environment variables from configs
make run-portal      # Build and run portal in background
make stop-portal     # Stop running portal

# Testing & cycles
make test            # Full e2e test cycle with teardown
make e2e             # Quick e2e test (up -> _test -> down)
make setup           # Setup environment (no teardown)
make teardown        # Tear down environment

# Cleanup
make clean           # Clean up all resources
```

### Manual Testing Steps

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env with RENTERD_URL, RENTERD_API_PASSWORD, RENTERD_SEED

# 2. Start services
make up

# 3. Build and configure
make build-portal
make setup-env

# 4. Run portal
make run-portal

# 5. Run tests manually
curl -f -s http://localhost:8081/api/health

# 6. Cleanup
make down
make clean
```

### CI/CD

The GitHub Actions workflow triggers on push/PR to `main` or `develop` branches. It uses:

- Composite actions in `.github/actions/` that delegate to shared bash scripts
- GitHub Actions services for MySQL, Maildev, Gofakes3
- Artifact uploads for the built portal binary

**Required GitHub Secrets:**
- `RENTERD_URL`
- `RENTERD_API_PASSWORD`
- `RENTERD_SEED`

## Critical Implementation Details

### Environment Variable Generation

The `scripts/yaml_to_env.py` script converts nested YAML paths to environment variables:

- Input YAML key: `core.db.type`
- Output env var: `PORTAL__CORE__DB__TYPE="mysql"`

All generated env vars use `PORTAL__` prefix and double-underscore separators.

### Service Health Checks

Docker Compose uses native health checks to ensure services are ready:

- **MySQL**: `mysqladmin ping -h localhost -u root -prootpassword`
- **Maildev**: HTTP check on port 1080
- **Gofakes3**: Port check on 9000

The `services-ready` container depends on all services being healthy, providing a synchronization point for external processes.

### Portal Plugin Manifest

The `portal-plugins.yaml` manifest is created by both Makefile and GitHub Actions:

```yaml
portalVersion: "develop"
plugins:
  - module: "go.lumeweb.com/portal-plugin-ipfs"
    version: "develop"
  - module: "go.lumeweb.com/portal-plugin-dashboard"
    version: "develop"
  - module: "go.lumeweb.com/portal-plugin-core"
    version: "develop"
```

Plugins must use full Go module paths with `@develop` version.

### Renterd Configuration

Renterd is an external service (NOT in docker-compose). It is configured via environment variables that are automatically mapped to portal configuration by `scripts/setup-env.sh`.

**Important:** The test seed in portal-core.yml is intentionally compromised for testing only. Never use it in production.

### Makefile Pattern

The Makefile mirrors the GitHub Actions workflow structure:
- `build` job → `make build-portal`
- `run` job → `make setup-env` + `make run-portal`
- Uses the same `portal-plugins.yaml` generation and `portal-builder` image

### GitHub Actions Pattern

The workflow uses composite actions to delegate to shared bash scripts:

```yaml
- name: Setup Environment
  uses: ./.github/actions/setup-env
  env:
    RENTERD_URL: ${{ secrets.RENTERD_URL }}
    RENTERD_API_PASSWORD: ${{ secrets.RENTERD_API_PASSWORD }}
    RENTERD_SEED: ${{ secrets.RENTERD_SEED }}

- name: Run Portal
  uses: ./.github/actions/run-portal
```

These actions internally call the same scripts as the Makefile.

## Testing Status

Currently, this repository provides the testing infrastructure but does NOT include actual E2E test implementations. The `make _test` target only performs basic health checks:

```bash
curl -f -s http://localhost:8081/api/health
```

A `scripts/mock-renterd.py` script exists but is not integrated into the main workflow.

When adding actual E2E tests, they should:
1. Be added as a new Makefile target or separate test suite
2. Follow the existing pattern of using shared scripts
3. Integrate with both local Makefile and GitHub Actions workflow
