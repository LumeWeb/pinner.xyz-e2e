# shellcheck shell=sh
# Makefile for E2E testing environment
# This mirrors the GitHub Actions workflow structure

# Configuration
PORTAL_PORT ?= 8080

# Phony targets (targets that don't represent files)
.PHONY: help up down setup-env start-portal restart-portal stop-portal test clean recreate-mysql recreate-powerdns
.PHONY: verify-services logs ps e2e setup teardown _test build-portal rebuild-portal
.PHONY: start-dns stop-dns dns-logs ensure-venv setup-compliance wait-ipfs wait-powerdns
.PHONY: test-tag debug-tag test-compliance
.PHONY: _test-compliance

help:
	@echo "E2E Testing Environment Commands:"
	@echo ""
	@echo "Service Management:"
	@echo "  make up              - Start all Docker services"
	@echo "  make down            - Stop all Docker services"
	@echo "  make restart         - Restart all Docker services"
	@echo "  make logs            - View service logs"
	@echo "  make ps              - Show running containers"
	@echo "  make wait-ipfs       - Wait for IPFS service to be ready"
	@echo "  make wait-powerdns   - Wait for PowerDNS service to be ready"
	@echo "  make verify-services - Wait for all services to be ready"
	@echo "  make recreate-mysql  - Recreate MySQL container to wipe data"
	@echo "  make recreate-powerdns - Recreate PowerDNS container to wipe data"
	@echo ""
	@echo "DNS Server:"
	@echo "  make start-dns       - Start dynamic DNS server (routes to localhost)"
	@echo "  make stop-dns        - Stop DNS server"
	@echo "  make dns-logs        - View DNS server logs"
	@echo "  make ensure-venv     - Ensure Python venv with dnserver is installed"
	@echo ""
	@echo "Portal Build & Run:"
	@echo "  make build-portal     - Build portal"
	@echo "  make rebuild-portal   - Force rebuild of portal"
	@echo "  make setup-env       - Generate environment variables from configs"
	@echo "  make start-portal    - Build and run portal in background"
	@echo "  make restart-portal  - Restart the portal"
	@echo "  make stop-portal     - Stop running portal"
	@echo ""
	@echo "Testing & Cycles:"
	@echo "  make test            - Full e2e test cycle with teardown (includes compliance)"
	@echo "  make e2e             - Quick e2e test (up -> _test -> down)"
	@echo "  make e2e-debug       - Quick e2e test in debug mode (with Delve on :2345)"
	@echo "  make test-tag TAG=@tag  Run a single tag test (no debugger)"
	@echo "  make debug-tag TAG=@tag Run a single tag test with Delve debugger"
	@echo "  make test-compliance - Run IPFS compliance tests only (requires running portal)"
	@echo "  make setup-compliance - Setup compliance testing environment (npm package)"
	@echo "  make setup           - Setup environment (no teardown)"
	@echo "  make teardown        - Tear down environment"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean           - Clean up all resources"

# Service Management
up:
	@echo "Starting services..."
	docker compose up -d services-ready
	@echo "Waiting for services to be healthy..."
	@docker compose ps
	@echo "[OK] Services are ready!"

wait-ipfs:
	@./scripts/wait-ipfs.sh

wait-powerdns:
	@./scripts/wait-powerdns.sh

verify-services:
	@./scripts/wait-mysql.sh
	@./scripts/wait-gofakes3.sh
	@./scripts/wait-ipfs.sh
	@./scripts/wait-powerdns.sh

recreate-mysql:
	@echo "Recreating MySQL container to wipe data..."
	@-docker compose stop mysql
	@-docker compose rm -f mysql
	@docker compose up -d mysql
	@echo "Waiting for MySQL to be ready..."
	@docker compose ps
	@echo "[OK] MySQL recreated with clean state"

recreate-powerdns:
	@echo "Recreating PowerDNS container to wipe data..."
	@-docker compose stop powerdns
	@-docker compose rm -f powerdns
	@docker compose up -d powerdns
	@echo "Waiting for PowerDNS to be ready..."
	@docker compose ps
	@echo "[OK] PowerDNS recreated with clean state"

down:
	@echo "Stopping services..."
	docker compose down
	@echo "[OK] Services stopped"

restart: down up

logs:
	docker compose logs -f

ps:
	docker compose ps

# Portal Build & Run
./dist/portal: portal-plugins.yaml
	@echo "Building portal with plugins using portal-builder..."
	docker run --rm \
		-v "$(PWD):/workspace" \
		-v "$(PWD)/dist:/dist" \
		ghcr.io/lumeweb/portal-builder:ubuntu \
		build-portal
	@echo "[OK] Portal built successfully"

build-portal: ./dist/portal
	@echo "[OK] Portal is ready"

rebuild-portal:
	@echo "Rebuilding portal..."
	@rm -rf dist portal-plugins.yaml
	@$(MAKE) build-portal
	@echo "[OK] Portal rebuilt successfully"

portal-plugins.yaml:
	@./scripts/create-plugin-manifest.sh

setup-env: recreate-mysql
	@echo "Generating environment variables..."
	@./scripts/setup-env.sh mysql false
	@echo "[OK] Environment configured"

start-portal: build-portal setup-env
	@./scripts/setup-kubo-bootstrap.sh
	@./scripts/start-portal.sh .portal.log

restart-portal: stop-portal start-portal

# Internal test target
_test:
	@echo "Running e2e tests..."
	@./scripts/wait-portal.sh
	@./scripts/run-tests.sh $(EXTRA_ARGS) || true

# Internal debug test target (uses Delve debugger)
_test-debug:
	@echo "Running e2e tests in debug mode..."
	@./scripts/wait-portal.sh
	@TEST_DEBUG=1 ./scripts/run-tests.sh $(EXTRA_ARGS) || true

# Internal compliance test target
_test-compliance: setup-compliance
	@echo "Running IPFS compliance tests..."
	@./scripts/wait-portal.sh
	@./scripts/run-compliance-tests.sh || true

# Full Cycles
e2e: up
	@echo "Running e2e test cycle..."
	@$(MAKE) _test || true
	@$(MAKE) down || true

e2e-debug: up
	@echo "Running e2e test cycle in debug mode..."
	@echo "Delve debugger listening on :2345"
	@echo "Connect with: dlv connect :2345"
	@$(MAKE) _test-debug || true
	@$(MAKE) down || true

# Run a single tag without debugger
# Usage: make test-tag TAG=@your-tag
test-tag:
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG parameter is required"; \
		echo "Usage: make test-tag TAG=@your-tag"; \
		exit 1; \
	fi
	@echo "Running test with tag: $(TAG)"
	$(MAKE) _test EXTRA_ARGS="--godog.tags=$(TAG)"

# Run a single tag in debug mode with Delve
# Usage: make debug-tag TAG=@your-tag
debug-tag:
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG parameter is required"; \
		echo "Usage: make debug-tag TAG=@your-tag"; \
		exit 1; \
	fi
	@echo "Running test with tag: $(TAG) in debug mode..."
	@echo "Delve debugger listening on :2345"
	@echo "Connect with: dlv connect :2345"
	$(MAKE) _test-debug EXTRA_ARGS="--godog.tags=$(TAG)"

# Run IPFS compliance tests only
# Usage: make test-compliance
# Prerequisites: Portal must be running (use 'make setup' first)
test-compliance:
	@echo "Running IPFS compliance tests..."
	$(MAKE) _test-compliance || true

# Full complete cycle: setup, test, teardown
test: up build-portal setup-env start-dns start-portal
	@echo "Running tests against running portal..."
	@$(MAKE) _test || true
	@echo "Running compliance tests..."
	@$(MAKE) _test-compliance || true
	@echo "Tests complete, tearing down..."
	@$(MAKE) stop-portal || true
	@$(MAKE) stop-dns || true
	@$(MAKE) down || true
	@echo "[OK] Full cycle completed"

# Setup only (no teardown)
setup: up build-portal setup-env start-dns start-portal
	@echo "[OK] Environment is ready for manual testing"

# Teardown only
teardown: down stop-portal stop-dns
	@echo "[OK] Environment torn down"

# Stop portal
stop-portal:
	@./scripts/wait-stop-portal.sh

# Compliance testing setup
setup-compliance:
	@./scripts/setup-compliance.sh

# DNS Server Management
ensure-venv:
	@./scripts/ensure-venv.sh

start-dns: ensure-venv
	@./scripts/start-dns.sh

stop-dns:
	@./scripts/stop-dns.sh

dns-logs:
	@if [ -f .dns.log ]; then \
		tail -f .dns.log; \
	else \
		echo "No DNS log file found. Start DNS server first with: make start-dns"; \
	fi

# Cleanup
clean:
	@echo "Cleaning up..."
	@$(MAKE) down
	@rm -rf dist .env portal portal-mysql.yml portal-plugins.yaml .venv
	@rm -f .portal.pid .dns.pid .dns.log
	@echo "[OK] Cleanup complete"
