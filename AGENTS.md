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
- `create-plugin-manifest.sh` - Creates portal-plugins.yaml manifest
- `setup-env.sh` - Generates env vars from YAML configs
- `start-portal.sh` - Starts portal in background with logging
- `lib.sh` - Shared utility library (logging, process management, wait loops)
- `load-env.sh` - Loads env from .env.renterd and .env (quiet mode)
- `yaml_to_env.py` - Converts YAML → `PORTAL__*` env vars (double-underscore nesting)
- `wait-mysql.sh`, `wait-gofakes3.sh`, `wait-portal.sh`, `wait-stop-portal.sh` - Service readiness waits (configurable timeouts)
- `start-dns.sh`, `stop-dns.sh` - DNS server control
- `ensure-venv.sh` - Python venv with dnserver

## Common Commands

**Service management:**
```bash
make up/down/restart    # Docker Compose services
make logs, make ps      # View logs, show containers
```

**Portal lifecycle:**
```bash
make build-portal       # Build with plugins via portal-builder
make setup-env          # Generate PORTAL__* env vars from YAML configs
make start-portal       # Build and run in background
make stop-portal        # Stop running portal
```

**Testing:**
```bash
make test               # Full e2e test cycle with teardown
make e2e                # Quick cycle (up → test → down)
go test -v              # Direct test execution
```

**Scripts:**
```bash
./scripts/create-plugin-manifest.sh
./scripts/setup-env.sh mysql
./scripts/start-portal.sh  # Background start with logging
./scripts/start-dns.sh / ./scripts/stop-dns.sh
```

### Test Execution

The e2e tests use godog (Cucumber for Go) with BDD scenarios in `features/` with step definitions in `steps/`. Portal loads plugins before starting HTTP server; tests wait for availability.

#### IPFS E2E Test Coverage

**Feature Files:**
- `features/ipfs_upload.feature` - Tests IPFS upload functionality:
  - Small file uploads (< 100MB) via standard HTTP
  - Large file uploads (100MB+) via TUS resumable protocol
  - Directory uploads with multiple files
  - Chunking and progress tracking for TUS uploads

- `features/ipfs_pinning.feature` - Tests IPFS pinning operations:
  - Pin existing CIDs to IPFS gateway
  - Track pin status transitions (queued → pinning → pinned/failed)
  - Pin multiple CIDs with size estimates
  - List pins with filtering by status/time
  - Remove pinned content

- `features/ipfs_content_list.feature` - Tests IPFS content management:
  - List uploaded content by CID
  - Filter content by various criteria
  - Verify content persistence across operations
  - Content metadata retrieval

**Step Definition Files:**
- `steps/ipfs_common_steps.go` - Shared IPFS wait/verification steps
  - Status polling and timeout handling
  - Cleanup management
  - Common IPFS assertions

- `steps/ipfs_upload_steps.go` - Upload-specific test logic
  - File routing (small vs large)
  - TUS protocol interactions
  - Upload completion verification

- `steps/ipfs_pinning_steps.go` - Pinning-specific test logic
  - Pin request creation
  - Status change detection
  - Pin list retrieval and filtering

- `steps/ipfs_content_list_steps.go` - Content list/filtering test logic
  - Content listing operations
  - Filter application
  - Metadata validation

**Full test cycle:**
```bash
make test  # Full e2e test with teardown
make e2e   # Quick cycle (up → test → down)
```

**Manual execution:**
```bash
make up && make build-portal && make setup-env && make start-portal
go test -v

# Run specific scenario (requires unique tags):
./scripts/run-tests.sh --godog.tags="@delete-api-key"
```

Each scenario must have a unique tag; use `@tagname` in feature files. When adding E2E tests:
1. Add as new Makefile target or test suite
2. Use shared scripts from `scripts/`
3. Integrate with both Makefile and GitHub Actions workflow

### CI/CD

Triggers on push/PR to `main` or `develop`. Uses:
- Composite actions (`.github/actions/`) → delegate to shared scripts
- GitHub Actions services (MySQL, Maildev, Gofakes3)
- Artifact uploads for built portal binary

**Required secrets:** `RENTERD_URL`, `RENTERD_API_PASSWORD`

### Environment Differences

**Local:** Persistent services via Docker Compose; config/db state persists between runs.

**CI/CD:** Fresh environment per run; GitHub Actions services with parallel build/test; no persistent state.

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

### BDD Step Design Guidelines

When defining godog/Cucumber step definitions, follow these patterns to ensure maintainability and support for multi-service architecture (IPFS, Arweave, S3, etc.).

#### Service-Specific vs Global Operations

**Service-Specific Patterns** - Use service prefixes for operations scoped to specific storage services:
- "the IPFS pin reaches pinned status" (NOT "the pin reaches pinned status")
- "the user uploads a 100MB file to IPFS" (NOT "the user uploads a 100MB file")
- "the user lists their IPFS pins" (NOT "the user lists their pins")

**Why:** Future services (Arweave, S3) will conflict with generic patterns. Service-specific language prevents step registration conflicts.

**Global Operations** - These don't need service prefixes because they're account-service-wide:
- "the operation completes" - operations are global (use account service)

#### Avoid Duplicate Step Patterns

**First-Registered Handler Wins:** If multiple step structs register the same step regex pattern, only the first-registered handler executes. This causes silent failures.

**Incorrect:**
```go
// ipfs_upload_steps.go
ctx.Step(`^the pin reaches pinned status$`, s.thePinReachesPinnedStatus)

// ipfs_pinning_steps.go
ctx.Step(`^the pin reaches pinned status$`, s.thePinReachesPinnedStatus)
// ^ Never executes - upload handler wins
```

**Correct:**
```go
// ipfs_common_steps.go - shared by all IPFS step files
ctx.Step(`^the IPFS pin reaches pinned status$`, s.theIPFSPinReachesPinnedStatus)

// Other IPFS step files do NOT register wait steps - they use the common ones
```

#### Common Steps File Pattern

Create shared wait/verification steps in a common file instead of duplicating across multiple step structs:

```go
// steps/ipfs_common_steps.go
ctx.Step(`^the IPFS pin reaches pinned status$`, s.theIPFSPinReachesPinnedStatus)
ctx.Step(`^the operation completes$`, s.theOperationCompletes)
```

**Registration Order:** Register common steps before service-specific step files in `godog_test.go`:

```go
// Register common/wait steps FIRST
ipfsCommonSteps := steps.NewIPFSCommonSteps()
ipfsCommonSteps.InitializeScenario(ctx)

// Then service-specific steps
uploadSteps := steps.NewIPFSUploadSteps()
uploadSteps.InitializeScenario(ctx)
```

#### Tag Naming Conventions

Use service prefixes in scenario tags to support selective execution:

```gherkin
# Tags
@ipfs-upload-small-file
@ipfs-upload-large-file-tus
@ipfs-upload-directory
@ipfs-pin-existing-cid
@ipfs-list-pins
```

**Avoid:**
- Generic tags like `@upload-small-file`, `@list-pins` (ambiguous when adding S3/Arweave)

#### Implicit vs Explicit Waits

**Implicit Waits (Preferred for Upload/Pin Steps):**
- Upload/pin steps should implicitly wait for pin and operation completion
- Feature files stay clean without redundant wait steps
- Example: `the user has 5 uploaded files` handles: upload → pin wait → operation wait

**Explicit Waits (When Needed):**
- Use explicit wait steps when testing granular operations
- Only needed when scenario requires explicit timing validation

#### Design for Multi-Service Expansion

When adding new step patterns, consider future services:

| Pattern                  | Good for Future? | Why?                         |
|--------------------------|------------------|------------------------------|
| `the pin reaches pinned status` | ❌ No | Conflicts with Arweave/S3 |
| `the IPFS pin reaches pinned status` | ✅ Yes | Service-specific |
| `the operation completes` | ✅ Yes | Global (account service) |

**Example Wrong Approach:**
```gherkin
Scenario: User pins an existing CID
  Given the user has a CID
  When the user starts pinning the CID
  And the pin reaches pinned status  # Conflicts with future services
```

**Example Correct Approach:**
```gherkin
Scenario: User pins an existing IPFS CID
  Given the user has an IPFS CID
  When the user starts pinning the IPFS CID
  And the IPFS pin reaches pinned status  # Service-specific
```

#### Step File Structure

Organize step files by functional area, making helpers accessible:

```
steps/
├── ipfs_common_steps.go        # Shared IPFS wait/verification steps
├── ipfs_upload_steps.go        # Upload-specific steps (uses common waits)
├── ipfs_pinning_steps.go       # Pinning-specific steps (uses common waits)
└── ipfs_content_list_steps.go  # Listing/filtering steps (uses common waits)
```

**Key Principles:**
- One step registration per pattern (no duplicates)
- Common steps in separate file, registered first
- Service-specific patterns for service-scoped operations
- Global patterns don't need prefixes (operations, account stuff)

## Implementation Details

### IPFS Helper Utilities

The E2E test suite includes comprehensive IPFS helper libraries:

**Go Helpers (`helpers/`):**
- `helpers/ipfs_common.go` - IPFS context management and state
  - Context key constants for CID, pins, content lists
  - State tracking for uploads, pinning operations, content filtering

- `helpers/ipfs_upload.go` - IPFS upload test utilities
  - TUS protocol upload handling
  - File size detection and routing
  - Upload progress tracking

- `helpers/portal_pinning.go` - Portal IPFS pinning API integration
  - Pin request management
  - Pin status polling and verification
  - Batch pinning operations

- `helpers/kubo_api.go` - Kubo (IPFS) node API helpers
  - Peer ID retrieval
  - Bootstrap configuration
  - Node operations

- `helpers/ipfs_validation.go` - Content integrity validation
  - CID computation and verification
  - Content-hash validation
  - IPFS path resolution

- `helpers/compliance_helpers.sh` - Compliance test orchestration
  - Compliance suite execution
  - Result parsing and reporting
  - Test environment setup

- `helpers/logging.go` - Structured logging utilities
  - Log level management
  - Context-aware logging
  - Test isolation support

- `helpers/panic_recovery.go` - Panic recovery for test isolation
  - Graceful error handling
  - Test cleanup on failure
  - Isolation between test scenarios

**Bash Scripts (`scripts/`):**
- `scripts/wait-ipfs.sh` - Wait for IPFS service readiness
- `scripts/get-kubo-peer-id.sh` - Retrieve Kubo node peer ID
- `scripts/setup-kubo-bootstrap.sh` - Configure Kubo bootstrap peers
- `scripts/run-compliance-tests.sh` - Run IPFS compliance tests
- `scripts/setup-compliance.sh` - Setup compliance test environment
- `scripts/validate-compliance-setup.sh` - Validate compliance prerequisites

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
