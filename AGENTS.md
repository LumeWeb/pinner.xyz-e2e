# AGENTS.md

This document provides guidance and reference information for AI agents working with code in this repository.

## Project Purpose

This is an end-to-end (E2E) testing environment for the LumeWeb Portal application. The repository provides testing infrastructure, not the portal source code itself. It provides:

1. Infrastructure services via Docker Compose (MySQL, Maildev, Gofakes3)
2. Portal build automation using the external `portal-builder` image
3. Portal configuration via environment variables generated from YAML configs
4. End-to-end test execution against the running portal

The portal application is built externally using `ghcr.io/lumeweb/portal-builder:ubuntu` with a `portal-plugins.yaml` manifest specifying plugins (ipfs, dashboard, core) at `@develop` versions.

## High-Level Architecture

### Directory Structure

- `docker-compose.yml` - Infrastructure services only (MySQL, Maildev, Gofakes3, services-ready sync container)
- `Makefile` - Local development commands mirroring the GitHub Actions workflow
- `.github/workflows/e2e-tests.yml` - CI/CD automation with parallel build and test jobs
- `.github/actions/` - Reusable composite actions that delegate to shared bash scripts
- `.github/config/portal-core.yml` - Minimal core portal configuration template
- `config/portal-mysql.yml` - MySQL configuration reference template
- `scripts/` - Shared bash scripts used by both Makefile and GitHub Actions

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
2. **DB Config**: `config/portal-mysql.yml` is the MySQL configuration reference template
3. **Env Generation**: `scripts/setup-env.sh` and `scripts/yaml_to_env.py` convert YAML configs to `PORTAL__*` environment variables in `.env`
4. **Renterd Override**: `RENTERD_*` env vars are preserved and mapped to `PORTAL__CORE__STORAGE__SIA__*`

**Key mappings:**
- `RENTERD_URL` → `PORTAL__CORE__STORAGE__SIA__URL`
- `RENTERD_API_PASSWORD` → `PORTAL__CORE__STORAGE__SIA__KEY`

### Shared Script Library

Bash scripts in `scripts/` provide shared functionality for multiple environments:

- `scripts/create-plugin-manifest.sh` - Creates the portal-plugins.yaml manifest for portal-builder
- `scripts/setup-env.sh mysql` - Generates environment variables from YAML configs
- `scripts/start-portal.sh [log-path]` - Starts portal in background with logging (uses PORTAL_PORT env var)
- `scripts/lib.sh` - Shared utility library for logging, process management, and wait loops
- `scripts/load-env.sh` - Loads environment from .env.renterd and .env (supports quiet mode)
- `scripts/yaml_to_env.py` - Converts YAML to `PORTAL__*` env vars using double-underscore separator for nested keys
- `scripts/wait-mysql.sh [timeout_seconds]` - Waits for MySQL to be ready for connections (uses MYSQL_WAIT_TIMEOUT env var, detects GitHub Actions vs local automatically)
- `scripts/wait-gofakes3.sh [timeout_seconds]` - Waits for gofakes3 to be ready on port 9000 (uses GOFAKES3_WAIT_TIMEOUT env var)
- `scripts/wait-portal.sh [timeout_seconds]` - Waits for portal HTTP endpoint to be available (uses PORTAL_PORT and PORTAL_WAIT_TIMEOUT env vars)
- `scripts/wait-stop-portal.sh [timeout_seconds]` - Waits for portal process to stop gracefully
- `scripts/start-dns.sh` - Starts dynamic DNS server
- `scripts/stop-dns.sh` - Stops dynamic DNS server
- `scripts/ensure-venv.sh` - Ensures Python venv with dnserver is installed

## Common Commands

### Script Execution

```bash
# Create plugin manifest
./scripts/create-plugin-manifest.sh

# Generate environment variables from configs
./scripts/setup-env.sh mysql

# Start portal in background with logging
# Uses PORTAL_PORT from environment (default: 8080)
./scripts/start-portal.sh

# Wait for services (all timeouts configurable via env vars)
./scripts/wait-mysql.sh
./scripts/wait-gofakes3.sh
./scripts/wait-portal.sh
./scripts/wait-stop-portal.sh 10

# Start/stop DNS server
./scripts/start-dns.sh
./scripts/stop-dns.sh
```

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
make start-portal    # Build and run portal in background
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
# Edit .env with RENTERD_URL, RENTERD_API_PASSWORD

# 2. Start services
make up

# 3. Build and configure
make build-portal
make setup-env

# 4. Run portal
# Note: This also removes stale configuration files from /etc/lumeweb/portal, $HOME/.lumeweb/portal, and ./
make start-portal

# 5. Run tests manually
# Note: Portal HTTP endpoint may take time to start (plugins load first)
# Source the .env file first to get environment variables
. .env
curl -H "Host: localhost:$PORTAL__CORE__PORT" http://localhost:$PORTAL__CORE__PORT/api/meta

# Run the full e2e test suite
go test -v

# 6. Cleanup
make down
make clean
```

### Running E2E Tests Manually

The e2e tests use godog (Cucumber for Go) with BDD scenarios defined in `features/` and step definitions in `steps/`.

**IMPORTANT:** Always use the `run-tests.sh` script to run tests. This script loads the `.env` file before executing tests, which is required for proper SDK configuration.

```bash
# Run all e2e tests
./scripts/run-tests.sh

# Run tests with verbose output and show godog steps
./scripts/run-tests.sh --godog.format=pretty

# Run specific scenario using its tag
./scripts/run-tests.sh --godog.tags="@delete-api-key"

# Run multiple scenarios with tags
./scripts/run-tests.sh --godog.tags="@create-api-key,@delete-api-key"

# Run specific feature file
./scripts/run-tests.sh features/account_management.feature
```

**Important:** Each scenario must have a unique tag to run it individually. Tags are defined in feature files using `@tagname` syntax. See the cucumber-testing skill documentation for tagging rules.

**Note:** When using `TestMain` (as this project does), there is no direct way to run a single scenario by name. The `-run` flag filters Go test functions, not godog scenarios. To run a specific scenario, you must use tags.

**Why use run-tests.sh:** The `run-tests.sh` script sources the `.env` file to load portal configuration (PORTAL__CORE__* environment variables) before running tests. Without this, the SDK cannot connect to the portal and tests will fail.

**Note:** The portal must be running before executing tests. Use the full test cycle or ensure portal is started:

```bash
# Full test cycle (includes setup, portal start, tests, teardown)
make test

# Quick test cycle (up, test, down)
make e2e

# Manual test execution
make up
make build-portal
make setup-env
make start-portal
# Wait for portal to be ready, then run tests
go test -v
```

### CI/CD

The GitHub Actions workflow triggers on push/PR to `main` or `develop` branches. It uses:

- Composite actions in `.github/actions/` that delegate to shared bash scripts
- GitHub Actions services for MySQL, Maildev, and Gofakes3
- Artifact uploads for the built portal binary

**Required GitHub Secrets:**
- `RENTERD_URL`
- `RENTERD_API_PASSWORD`

### Environment Differences

**Local Development:**
- Reusable environment across sessions
- Services (MySQL, Maildev, Gofakes3) run via Docker Compose
- `make up` / `make down` for persistent services
- `make start-portal` / `make stop-portal` for portal lifecycle
- Config and database state persists between runs

**CI/CD:**
- Fresh environment per workflow run
- MySQL, Maildev, Gofakes3 run as GitHub Actions services (accessible via localhost)
- Services start once per job execution
- Parallel build and test jobs
- No persistent state between runs
- Artifacts uploaded/downloaded between jobs

## Critical Guidelines

Follow these rules when working with this repository:

### HTTP Request Handling

The portal uses vhost routing with different subdomains for different plugin APIs:

**Core/meta endpoints:**
```bash
curl -H "Host: localhost:8080" http://localhost:8080/api/meta
```

**Account/authentication endpoints:**
```bash
curl -H "Host: account.localhost:8080" \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","first_name":"Test","last_name":"User","password":"Test123!"}' \
     http://localhost:8080/api/auth/register
```

**Host headers by plugin:**
- Core/meta: `Host: localhost:8080`
- Account/auth: `Host: account.localhost:8080`
- Operations: `Host: account.localhost:8080`

**Reference implementation:**
- Check `helpers/steps_common.go::GetPortalHost()` for correct vhost header
- Account/auth operations use `account.localhost:8080`
- Core/meta operations accept both `localhost:8080` and `account.localhost:8080`

### Environment Variable Management

Never manually manipulate environment variables. Use the standard workflow:

1. Update configuration in `config/portal-mysql.yml` (or `.github/config/portal-core.yml` for base settings)
2. Regenerate `.env` from configs:
   ```bash
   make setup-env
   . .env
   # Run commands
   ```

This ensures consistent configuration across all environments and prevents configuration drift.

## Implementation Details

### Environment Variable Generation

The `scripts/yaml_to_env.py` script converts nested YAML paths to environment variables using `PORTAL__` prefix and double-underscore separators:

**Input YAML key:** `core.db.type`  
**Output env var:** `PORTAL__CORE__DB__TYPE="mysql"`

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

Renterd is an external service (not in docker-compose). It is configured via environment variables that are automatically mapped to portal configuration by `scripts/setup-env.sh`.

**WARNING:** The test seed in `portal-core.yml` is intentionally compromised for testing only. Never use it in production.

### Workflow Pattern

The Makefile and GitHub Actions workflow use identical scripts:

**Makefile targets:**
```makefile
build:  ./scripts/create-plugin-manifest.sh + make ensure-portal-built
start:  ./scripts/setup-env.sh + ./scripts/start-portal.sh
```

**GitHub Actions workflow:**
```yaml
- name: Create Plugin Manifest
  run: ./scripts/create-plugin-manifest.sh

- name: Build Portal
  uses: addnab/docker-run-action@v3
  with:
    image: ghcr.io/lumeweb/portal-builder:ubuntu
    options: -v ${{ github.workspace }}:/workspace -v ${{ github.workspace }}/dist:/dist
    run: build-portal

- name: Setup Environment
  uses: ./.github/actions/setup-env
  env:
    RENTERD_URL: ${{ secrets.RENTERD_URL }}
    RENTERD_API_PASSWORD: ${{ secrets.RENTERD_API_PASSWORD }}

- name: Start Portal
  uses: ./.github/actions/start-portal
```

Composite actions in `.github/actions/` delegate to shared scripts, ensuring consistent behavior across environments.

## Environment Variables

### Configurable Timeouts

Several scripts support configurable timeout values via environment variables:

- **`GOFAKES3_WAIT_TIMEOUT`** - Timeout for `wait-gofakes3.sh` (default: 30 seconds)
- **`PORTAL_WAIT_TIMEOUT`** - Timeout for `wait-portal.sh` (default: 30 seconds)
- **`MYSQL_WAIT_TIMEOUT`** - Timeout for `wait-mysql.sh` (default: 30 seconds)

Example:
```bash
# Use custom timeout
GOFAKES3_WAIT_TIMEOUT=60 ./scripts/wait-gofakes3.sh
PORTAL_WAIT_TIMEOUT=60 ./scripts/wait-portal.sh
MYSQL_WAIT_TIMEOUT=60 ./scripts/wait-mysql.sh
```

### Portal Configuration

- **`PORTAL_PORT`** - Port for portal HTTP endpoint (default: 8080)
  - Used by: `start-portal.sh`, `wait-portal.sh`
  - Set via YAML config or environment variable

## Test Execution

The portal HTTP endpoint requires time to become available. The portal loads all plugins before starting the HTTP server, so wait scripts ensure tests run only after readiness.

**Test infrastructure:**
- `make _test` waits for portal availability using `scripts/wait-portal.sh`
- Tests execute godog scenarios from `features/` directory
- Step definitions reside in `steps/` directory

**Note:** A `scripts/mock-renterd.py` script exists but is not integrated into the main workflow.

When adding new E2E tests:
1. Add as a new Makefile target or separate test suite
2. Use shared scripts from `scripts/`
3. Integrate with both Makefile and GitHub Actions workflow
