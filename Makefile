# shellcheck shell=sh
# Makefile for E2E testing environment
# This mirrors the GitHub Actions workflow structure

# Configuration
PORTAL_PORT ?= 8080

# Phony targets (targets that don't represent files)
.PHONY: help up down setup-env start-portal test clean recreate-mysql
.PHONY: verify-services logs ps e2e setup teardown _test stop-portal ensure-portal-built
.PHONY: start-dns stop-dns dns-logs ensure-venv

help:
	@echo "E2E Testing Environment Commands:"
	@echo ""
	@echo "Service Management:"
	@echo "  make up              - Start all Docker services"
	@echo "  make down            - Stop all Docker services"
	@echo "  make restart         - Restart all Docker services"
	@echo "  make logs            - View service logs"
	@echo "  make ps              - Show running containers"
	@echo "  make recreate-mysql  - Recreate MySQL container to wipe data"
	@echo ""
	@echo "DNS Server:"
	@echo "  make start-dns       - Start dynamic DNS server (routes to localhost)"
	@echo "  make stop-dns        - Stop DNS server"
	@echo "  make dns-logs        - View DNS server logs"
	@echo "  make ensure-venv     - Ensure Python venv with dnserver is installed"
	@echo ""
	@echo "Portal Build & Run:"
	@echo "  make ensure-portal-built - Ensure portal is built"
	@echo "  make setup-env       - Generate environment variables from configs"
	@echo "  make start-portal    - Build and run portal in background"
	@echo "  make stop-portal     - Stop running portal"
	@echo ""
	@echo "Testing & Cycles:"
	@echo "  make test            - Full e2e test cycle with teardown"
	@echo "  make e2e             - Quick e2e test (up -> _test -> down)"
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

recreate-mysql:
	@echo "Recreating MySQL container to wipe data..."
	@-docker compose stop mysql
	@-docker compose rm -f mysql
	@docker compose up -d mysql
	@echo "Waiting for MySQL to be ready..."
	@docker compose ps
	@echo "[OK] MySQL recreated with clean state"

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

ensure-portal-built: ./dist/portal
	@echo "[OK] Portal is ready"

portal-plugins.yaml:
	@./scripts/create-plugin-manifest.sh
setup-env: recreate-mysql
	@echo "Generating environment variables..."
	@./scripts/setup-env.sh mysql false
	@echo "[OK] Environment configured"

start-portal: ensure-portal-built setup-env
	@echo "Starting portal..."
	@cp ./dist/portal ./portal
	@chmod +x ./portal
	@./scripts/wait-mysql.sh
	@./scripts/wait-gofakes3.sh
	@./scripts/start-portal.sh .portal.log
	@sleep 3
	@echo "[OK] Portal started (PID: $(cat .portal.pid))"

# Internal test target
_test:
	@echo "Running e2e tests..."
	@./scripts/wait-portal.sh
	@./scripts/run-tests.sh || true

# Full Cycles
e2e: up
	@echo "Running e2e test cycle..."
	@$(MAKE) _test || true
	@$(MAKE) down || true

# Full complete cycle: setup, test, teardown
test: up ensure-portal-built setup-env start-dns start-portal
	@echo "Running tests against running portal..."
	@$(MAKE) _test || true
	@echo "Tests complete, tearing down..."
	@$(MAKE) stop-portal || true
	@$(MAKE) stop-dns || true
	@$(MAKE) down || true
	@echo "[OK] Full cycle completed"

# Setup only (no teardown)
setup: up ensure-portal-built setup-env start-dns start-portal
	@echo "[OK] Environment is ready for manual testing"

# Teardown only
teardown: down stop-portal stop-dns
	@echo "[OK] Environment torn down"

# Stop portal
stop-portal:
	@./scripts/wait-stop-portal.sh

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
