# E2E Testing Environment

End-to-end testing infrastructure for the LumeWeb Portal application. This repository provides automated testing for portal functionality using docker-compose services, a shared script library, and integrated GitHub Actions workflows.

## Overview

This repository does not contain the portal source code. Instead, it provides infrastructure and automation to:

1. Run infrastructure services via Docker Compose
2. Build the portal application using the external `portal-builder` image
3. Configure the portal via environment variables generated from YAML configs
4. Execute end-to-end tests against the running portal

The portal application is built using `ghcr.io/lumeweb/portal-builder:ubuntu` with a `portal-plugins.yaml` manifest specifying plugins (ipfs, dashboard, core) at `@develop` versions.

## Architecture

### Directory Structure

- `docker-compose.yml` - Infrastructure service definitions with health checks
- `Makefile` - Local development commands
- `.github/workflows/e2e-tests.yml` - CI/CD automation
- `.github/actions/` - Reusable composite actions
- `.github/config/portal-core.yml` - Core portal configuration template
- `config/portal-mysql.yml` - MySQL configuration reference
- `scripts/` - Shared bash scripts (used by Makefile and GitHub Actions)

### Services

**Docker Compose Services**
- `mysql` - Percona Server 8.4 on port 3306
- `maildev` - Email catcher on ports 1025 (SMTP) and 1080 (Web UI)
- `gofakes3` - S3-compatible storage on port 9000
- `services-ready` - Synchronization container

**External Services**
- `renterd` - External Sia storage service configured via environment variables
- `portal` - Built and run separately using portal-builder image

### Configuration Flow

1. **Base Config**: `.github/config/portal-core.yml` provides minimal portal settings
2. **DB Config**: `config/portal-mysql.yml` provides MySQL-specific settings
3. **Env Generation**: `scripts/setup-env.sh` and `scripts/yaml_to_env.py` convert YAML to `PORTAL__*` environment variables in `.env`
4. **Renterd Override**: `RENTERD_*` env vars are preserved and mapped to `PORTAL__CORE__STORAGE__SIA__*`

**Key mappings:**
- `RENTERD_URL` → `PORTAL__CORE__STORAGE__SIA__URL`
- `RENTERD_API_PASSWORD` → `PORTAL__CORE__STORAGE__SIA__KEY`

## Prerequisites

- Docker and Docker Compose installed
- `yq` installed (YAML processor - available via Go installation)
- Access to external renterd service
- Renterd credentials (URL, API password)

## Quick Start

### Local Development

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env with RENTERD_URL and RENTERD_API_PASSWORD

# 2. Start services
make up

# 3. Build and run portal
make test
```

### Manual Testing

```bash
# Service management
make up              # Start all Docker services
make down            # Stop all Docker services
make restart         # Restart all Docker services
make logs            # View service logs
make ps              # Show running containers

# Portal build & run
make ensure-portal-built  # Build portal with plugins
make setup-env            # Generate environment variables
make start-portal         # Build and run portal in background
make stop-portal          # Stop running portal

# Testing & cycles
make test            # Full e2e test cycle with teardown
make e2e             # Quick e2e test (up -> _test -> down)
make setup           # Setup environment (no teardown)
make teardown        # Tear down environment

# Cleanup
make clean           # Clean up all resources
```

## Configuration Management

**NEVER manually manipulate environment variables.** Always regenerate from YAML configs:

1. Update configuration in `config/portal-mysql.yml` (or `.github/config/portal-core.yml` for base settings)
2. Run `make setup-env` to regenerate `.env`
3. Source `.env` before running commands

This ensures consistent configuration across local and CI/CD environments.

### Configurable Timeouts

Several scripts support configurable timeout values:

**Environment Variable Configuration:**
- **`GOFAKES3_WAIT_TIMEOUT`** - Timeout for `wait-gofakes3.sh` (default: 30 seconds)
- **`PORTAL_WAIT_TIMEOUT`** - Timeout for `wait-portal.sh` (default: 30 seconds)
- **`MYSQL_WAIT_TIMEOUT`** - Timeout for `wait-mysql.sh` (default: 30 seconds)

**Argument-based Configuration:**
- **`wait-ipfs.sh [timeout_seconds]`** - Timeout for IPFS service readiness (default: 30 seconds)

Example:
```bash
# Use custom timeout via environment variable
GOFAKES3_WAIT_TIMEOUT=60 ./scripts/wait-gofakes3.sh
PORTAL_WAIT_TIMEOUT=60 ./scripts/wait-portal.sh
MYSQL_WAIT_TIMEOUT=60 ./scripts/wait-mysql.sh

# Use custom timeout via argument
./scripts/wait-ipfs.sh 60
```

## Manual Testing Steps

1. ```bash
   make up
   make ensure-portal-built
   make setup-env
   make start-portal
   ```

2. Source the environment and verify portal:
   ```bash
   . .env
   curl -H "Host: localhost:$PORTAL__CORE__PORT" http://localhost:$PORTAL__CORE__PORT/api/meta
   ```

3. Run tests manually or via the test cycle:
   ```bash
   make _test
   make teardown
   ```

## Running E2E Tests

### IPFS E2E Features

The E2E test suite includes comprehensive IPFS coverage for upload, pinning, and content management:

**Feature Files:**
- `features/ipfs_upload.feature` - Tests IPFS upload functionality:
  - Upload small files (< 100MB) via direct upload
  - Upload large files (100MB+) via TUS resumable protocol
  - Upload directories with multiple files

- `features/ipfs_pinning.feature` - Tests IPFS pinning operations:
  - Pin existing CIDs to IPFS
  - Track pin status changes (queued, pinning, pinned, failed)
  - Pin multiple CIDs with size estimates
  - List pins with filtering capabilities

- `features/ipfs_content_list.feature` - Tests IPFS content management:
  - List uploaded content by CID
  - Filter content by various criteria
  - Verify content persistence across operations

**Step Definitions:**
- `steps/ipfs_common_steps.go` - Shared IPFS wait/verification steps
- `steps/ipfs_upload_steps.go` - Upload-specific test logic
- `steps/ipfs_pinning_steps.go` - Pinning-specific test logic
- `steps/ipfs_content_list_steps.go` - Content list/filtering test logic

**Supporting Helpers:**
- `helpers/ipfs_common.go` - IPFS context management
- `helpers/ipfs_upload.go` - Upload test utilities
- `helpers/portal_pinning.go` - Portal pinning API integration
- `helpers/kubo_api.go` - Kubo (IPFS node) API helpers for peer ID and bootstrap
- `helpers/ipfs_validation.go` - Content integrity validation
- `helpers/compliance_helpers.sh` - Compliance test orchestration
- `helpers/logging.go` - Structured logging utilities
- `helpers/panic_recovery.go` - Panic recovery for test isolation

**Supporting Scripts:**
- `scripts/wait-ipfs.sh` - Wait for IPFS service readiness
- `scripts/get-kubo-peer-id.sh` - Retrieve Kubo node peer ID
- `scripts/setup-kubo-bootstrap.sh` - Configure Kubo bootstrap peers
- `scripts/run-compliance-tests.sh` - Run Pinning Service API compliance tests
- `scripts/setup-compliance.sh` - Setup compliance test environment
- `scripts/validate-compliance-setup.sh` - Validate compliance prerequisites

### Using the Test Runner

Always use `scripts/run-tests.sh` to run tests. This script loads the `.env` file before execution.

```bash
# Run all e2e tests
./scripts/run-tests.sh

# Run tests with verbose output
./scripts/run-tests.sh --godog.format=pretty

# Run specific scenario using tag
./scripts/run-tests.sh --godog.tags="@delete-api-key"

# Run specific feature file
./scripts/run-tests.sh features/account_management.feature
```

### Test Organization

The e2e tests use godog (Cucumber for Go):
- `features/` - BDD scenarios
- `steps/` - Step definitions
- `helpers/` - Shared test utilities

Each scenario must have a unique tag for individual execution.

### IPFS Compliance Testing

The e2e test suite includes IPFS spec compliance testing using the `@ipfs-shipyard/pinning-service-compliance` npm package.

```bash
# Run compliance tests only (requires running portal)
make test-compliance

# Run full test cycle including compliance
make test

# Skip compliance tests
NO_COMPLIANCE=true make test
```

**Compliance test execution:**
- Runs AFTER godog BDD tests complete
- Creates fresh test user and API key via portal API
- Tests IPFS Pinning Service API spec compliance
- Parses JSON report and reports pass/fail status
- Requires Node.js/npx availability

**Environment variables:**
- `COMPLIANCE_DEBUG` - Enable debug mode (default: false)
- `NO_COMPLIANCE` - Skip compliance tests (default: false)
- `PORTAL_PORT` - Portal HTTP port (default: 8080)

**Validation:**
Run the validation script to check setup prerequisites:
```bash
./scripts/validate-compliance-setup.sh
```

### Kubo IPFS Node Integration

The E2E test suite includes integration with a local Kubo (IPFS) node for advanced IPFS testing:

**Kubo Setup Scripts:**
- `scripts/get-kubo-peer-id.sh` - Retrieves the Kubo node's peer ID
- `scripts/setup-kubo-bootstrap.sh` - Configures bootstrap peers for Kubo

**Kubo Helper Functions:**
- `helpers/kubo_api.go` - Go helpers for Kubo API interactions:
  - Peer ID retrieval
  - Bootstrap configuration
  - Node status checks

**Usage:**
```bash
# Get peer ID from Kubo
./scripts/get-kubo-peer-id.sh

# Setup bootstrap peers
./scripts/setup-kubo-bootstrap.sh
```

## CI/CD

### GitHub Actions Workflow

The workflow triggers on push/PR to `main` or `develop` branches:

- Uses shared bash scripts from `scripts/`
- Runs MySQL, Maildev, and Gofakes3 as GitHub Actions services
- Uploads build artifacts
- Runs full e2e test suite

### Required Secrets

Configure these in your GitHub repository:
- `RENTERD_URL` - URL to external renterd service
- `RENTERD_API_PASSWORD` - API password for renterd authentication

## Shared Script Architecture

All scripts in `scripts/` work in both local and GitHub Actions environments:

```bash
# Create plugin manifest
./scripts/create-plugin-manifest.sh

# Generate environment variables
./scripts/setup-env.sh mysql

# Start portal in background with logging
# Uses PORTAL_PORT from environment (default: 8080)
./scripts/start-portal.sh

# Wait for services (timeouts configurable via env vars)
./scripts/wait-mysql.sh
./scripts/wait-gofakes3.sh
./scripts/wait-portal.sh
./scripts/wait-stop-portal.sh 10

# DNS server
./scripts/start-dns.sh
./scripts/stop-dns.sh
```

## Important Notes

### Portal HTTP Endpoint

The portal HTTP endpoint (port 8080) may take several seconds to become available. This is because the portal loads all plugins before starting the HTTP server. The wait scripts ensure tests only run after the portal is ready.

### vhost Routing

The portal uses vhost routing with different subdomains for different plugin APIs:

```bash
# Core/meta endpoints
curl -H "Host: localhost:8080" http://localhost:8080/api/meta

# Account/authentication endpoints
curl -H "Host: account.localhost:8080" \
       -H "Content-Type: application/json" \
       -d '{"email":"test@example.com","password":"Test123!"}' \
       http://localhost:8080/api/auth/register
```

**Host headers by plugin:**
- Core/meta: `Host: localhost:8080`
- Account/auth: `Host: account.localhost:8080`

## Service Access

### Local Development

- **Maildev Web UI**: http://localhost:1080
- **Gofakes3**: http://localhost:9000

## Troubleshooting

### Services Not Starting

```bash
# Check service logs
docker-compose logs <service-name>

# Check service status
docker-compose ps

# Verify service health
docker-compose ps
```

### Portal Not Connecting to MySQL

```bash
# Verify MySQL is healthy
docker-compose ps mysql
```

### Renterd Authentication Issues

Verify your `.env` file contains correct values:
- `RENTERD_URL`
- `RENTERD_API_PASSWORD`

For CI/CD, ensure GitHub secrets are configured.

### Test Failures

Check test logs in GitHub Actions artifacts or review locally:

```bash
make _test
```

## Service Health Checks

**Local (Docker Compose):**
- **MySQL**: `mysqladmin ping -h localhost -u root -prootpassword`
- **Maildev**: HTTP check on port 1080
- **Gofakes3**: Port check on 9000

The `services-ready` container depends on all services being healthy, providing a synchronization point for external processes.

**GitHub Actions:**
- MySQL runs as a service with health checks configured in the workflow
- `scripts/wait-mysql.sh` detects environment and uses appropriate connection method:
  - GitHub Actions: `mysqladmin ping -h localhost` (service connection)
  - Local: `docker compose exec -T mysql mysqladmin ping`
- `scripts/wait-gofakes3.sh` checks exposed port 9000 on localhost for gofakes3 readiness (works in both CI and local)

## Plugin Configuration

The following plugins are configured at `@develop` versions:
- IPFS plugin: `go.lumeweb.com/portal-plugin-ipfs`
- Dashboard plugin: `go.lumeweb.com/portal-plugin-dashboard`
- Core plugin: `go.lumeweb.com/portal-plugin-core`

Plugins are specified in `portal-plugins.yaml`, created by `scripts/create-plugin-manifest.sh`. The portal-builder image uses this manifest to build the portal with the correct plugins.
