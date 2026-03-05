# E2E Testing Environment

This directory contains the setup for end-to-end testing of the portal with various services. The setup includes both local development (docker-compose + Makefile) and CI/CD (GitHub Actions workflow) configurations.

## Architecture

This e2e testing setup consists of:

1. **Docker Compose**: Infrastructure services only (mysql, maildev, gofakes3)
2. **Makefile**: Local development commands to run e2e tests
3. **GitHub Workflow**: CI/CD automation for running e2e tests in GitHub Actions
4. **Shared Bash Scripts**: Reusable scripts for both local and CI environments

## Services

The following services are included in this e2e environment:

- **MySQL**: Database service running on port 3306
- **Maildev**: Email catching service running on ports 1025 (SMTP) and 1080 (Web UI)
- **Gofakes3**: S3-compatible storage service running on port 4568

**Note**: The portal itself is NOT included in docker-compose. It's built and run separately using portal-builder.

**Note**: Renterd is NOT included in docker-compose. It is expected to be an external service configured via dedicated environment variables (`RENTERD_URL`, `RENTERD_API_PASSWORD`, `RENTERD_SEED`).

## Local Development

### Prerequisites

- Docker and Docker Compose installed
- Python 3 with `yq` installed
- Access to external renterd server
- Renterd credentials (URL, API password, seed)

### Setup

1. **Configure environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your actual RENTERD_URL, RENTERD_API_PASSWORD, and RENTERD_SEED
   ```

2. **Start services:**
   ```bash
   make up
   ```

### Running Tests

The Makefile provides several commands:

```bash
# Service Management
make up              - Start all Docker services
make down            - Stop all Docker services
make restart         - Restart all Docker services
make logs            - View service logs
make ps              - Show running containers

# Portal Build & Run
make build-portal    - Build portal with plugins
make setup-env       - Generate environment variables from configs
make run-portal      - Build and run portal in background
make stop-portal     - Stop running portal

# Testing & Cycles
make test            - Full e2e test cycle with teardown
make e2e             - Quick e2e test (up -> _test -> down)
make setup           - Setup environment (no teardown)
make teardown        - Tear down environment

# Cleanup
make clean           - Clean up all resources
```

### Configuration Files

- `docker-compose.yml`: Defines infrastructure services only with native health checks
- `.github/config/portal-core.yml`: Core portal configuration template
- `.env`: Environment variables (copy from `.env.example`)
- `portal-mysql.yml`: Generated MySQL configuration (auto-created by setup-env.sh)

### Health Checks

Docker Compose uses native health checks for all services:
- MySQL: Built-in health check via `mysqladmin ping`
- Maildev: Checks SMTP port 1025
- Gofakes3: Checks port 4568

The `services-ready` container depends on all services being healthy, providing a synchronization point for external processes.

### Accessing Services (Local)

- **Maildev Web UI**: http://localhost:1080
- **Gofakes3**: http://localhost:4568

## GitHub Actions CI/CD

### Workflow File

The GitHub Actions workflow is located at `.github/workflows/e2e-tests.yml`.

### Modular Actions

The workflow uses reusable composite actions in `.github/actions/` that delegate to shared bash scripts:

- **setup-env/action.yml**: Calls `scripts/setup-env.sh` to generate environment variables
- **build-portal/action.yml**: Builds portal using portal-builder image
- **run-portal/action.yml**: Calls `scripts/run-portal.sh` to run portal and verify startup

### Shared Bash Scripts

The following bash scripts are shared between local Makefile and GitHub Actions:

- **scripts/setup-env.sh**: Generates environment variables from YAML configs
  - Usage: `./scripts/setup-env.sh [mysql|sqlite] [true|false]`
  - Second parameter `true` enables GitHub Actions mode (exports to GITHUB_ENV)
- **scripts/run-portal.sh**: Runs portal and verifies it binds to the expected port
  - Usage: `./scripts/run-portal.sh [true|false]`
  - Parameter `true` enables GitHub Actions mode

### Required Secrets

Configure these secrets in your GitHub repository:

- `RENTERD_URL`: URL to external renterd server
- `RENTERD_API_PASSWORD`: API password for renterd authentication
- `RENTERD_SEED`: Seed for renterd configuration

### Workflow Features

- Triggers on push/PR to develop branch
- Runs all services as GitHub Actions services
- Builds portal using portal-builder with `portal-plugins.yaml` manifest
- Configures MySQL mode with dedicated environment variables
- Waits for services to be healthy
- Runs portal with environment variables
- Executes e2e tests
- Uploads test results as artifacts

### GitHub Actions Services

The workflow uses GitHub Actions services for:
- **MySQL**: percona/percona-server:8.4
- **Maildev**: maildev/maildev:latest
- **Gofakes3**: johannesboyne/gofakes3:latest

## Configuration Details

### MySQL Mode

The portal is configured to run in MySQL mode with the following settings:
- Host: 127.0.0.1
- Port: 3306
- Database: portal
- User: portal
- Password: portal
- Charset: utf8mb4

These settings are generated by `scripts/setup-env.sh` and converted to `PORTAL__CORE__DB__*` environment variables.

### Plugins

The following plugins are configured to use the develop version:
- IPFS plugin: `go.lumeweb.com/portal-plugin-ipfs@develop`
- Dashboard plugin: `go.lumeweb.com/portal-plugin-dashboard@develop`
- Core plugin: `go.lumeweb.com/portal-plugin-core@develop`

These are specified in `portal-plugins.yaml` which is created by the GitHub Actions workflow. The portal-builder image uses this manifest to build portal with the correct plugins.

**Note**: Plugins must use the full Go module path (e.g., `go.lumeweb.com/portal-plugin-ipfs@develop`).

### Renterd Configuration

Renterd is configured via dedicated environment variables that map to portal configuration:

| Environment Variable | Portal Configuration Path |
|---------------------|--------------------------|
| `RENTERD_URL` | `PORTAL__CORE__STORAGE__SIA__URL` |
| `RENTERD_API_PASSWORD` | `PORTAL__CORE__STORAGE__SIA__API_PASSWORD` |
| `RENTERD_SEED` | `PORTAL__CORE__STORAGE__SIA__SEED` |

The `scripts/setup-env.sh` script automatically maps these variables when running in either local or GitHub Actions mode.

## Testing

### Local Testing

```bash
# Full e2e test cycle with teardown
make test

# Quick e2e test (up -> _test -> down)
make e2e

# Setup environment (no teardown)
make setup

# Tear down environment
make teardown

# Manual step-by-step
make up
make build-portal
make setup-env
make run-portal
# ... run tests manually ...
make down
make clean
```

### CI Testing

Push to develop branch or create a PR to trigger the e2e test workflow.

## Troubleshooting

### Services not starting

Check logs with: `docker-compose logs <service-name>`

### Portal not connecting to MySQL

Ensure MySQL is healthy: `docker-compose ps mysql`

### Renterd authentication issues

Verify your `.env` file has the correct `RENTERD_URL`, `RENTERD_API_PASSWORD`, and `RENTERD_SEED`, or that the GitHub secrets are configured.

### Test failures

Check test logs in the uploaded artifacts from GitHub Actions.
