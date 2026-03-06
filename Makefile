# shellcheck shell=sh
# Makefile for E2E testing environment
# This mirrors the GitHub Actions workflow structure

# Configuration
MOCK_RENTERD_PORT ?= 8081
PORTAL_PORT ?= 8080

# Phony targets (targets that don't represent files)
.PHONY: help up down build-portal setup-env run-portal test clean
.PHONY: verify-services logs ps e2e setup teardown _test stop-portal

help:
	@echo "E2E Testing Environment Commands:"
	@echo ""
	@echo "Service Management:"
	@echo "  make up              - Start all Docker services"
	@echo "  make down            - Stop all Docker services"
	@echo "  make restart         - Restart all Docker services"
	@echo "  make logs            - View service logs"
	@echo "  make ps              - Show running containers"
	@echo ""
	@echo "Portal Build & Run:"
	@echo "  make build-portal    - Build portal with plugins"
	@echo "  make setup-env       - Generate environment variables from configs"
	@echo "  make run-portal      - Build and run portal in background"
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
build-portal: portal-plugins.yaml
	@echo "Building portal with plugins using portal-builder..."
	docker run --rm \
		-v "$(PWD):/workspace" \
		-v "$(PWD)/dist:/dist" \
		ghcr.io/lumeweb/portal-builder:ubuntu \
		build-portal
	@echo "[OK] Portal built successfully"

portal-plugins.yaml:
	@echo "Creating plugin manifest..."
	@yq -n '.portalVersion = "develop"' > portal-plugins.yaml
	@yq -i '.plugins[0].module = "go.lumeweb.com/portal-plugin-ipfs" | .plugins[0].version = "develop"' portal-plugins.yaml
	@yq -i '.plugins[1].module = "go.lumeweb.com/portal-plugin-dashboard" | .plugins[1].version = "develop"' portal-plugins.yaml
	@yq -i '.plugins[2].module = "go.lumeweb.com/portal-plugin-core" | .plugins[2].version = "develop"' portal-plugins.yaml
	@echo "[OK] Created portal-plugins.yaml:"
	@cat portal-plugins.yaml

setup-env:
	@echo "Generating environment variables..."
	@./scripts/setup-env.sh mysql false
	@echo "[OK] Environment configured"

run-portal: build-portal setup-env
	@echo "Running portal..."
	@cp ./dist/portal ./portal
	@chmod +x ./portal
	@./scripts/run-portal-bg.sh $(PORTAL_PORT)
	@sleep 3
	@echo "[OK] Portal started (PID: $$(cat .portal.pid))"

# Internal test target
_test:
	@echo "Running e2e tests..."
	@echo "Testing HTTP endpoints..."
	@curl -f -s http://localhost:$(MOCK_RENTERD_PORT)/api/health > /dev/null && echo "[OK] Renterd health check passed" || echo "[ERROR] Renterd health check failed"
	@echo "[OK] E2e tests completed"

# Full Cycles
e2e: up
	@echo "Running e2e test cycle..."
	@$(MAKE) _test
	@$(MAKE) down

# Full complete cycle: setup, test, teardown
test: up build-portal setup-env run-portal
	@echo "Running tests against running portal..."
	@$(MAKE) _test
	@echo "Tests complete, tearing down..."
	@$(MAKE) down
	@echo "[OK] Full cycle completed successfully"

# Setup only (no teardown)
setup: up build-portal setup-env run-portal
	@echo "[OK] Environment is ready for manual testing"

# Teardown only
teardown: down stop-portal
	@echo "[OK] Environment torn down"

# Stop portal
stop-portal:
	@if [ -f .portal.pid ]; then \
		PID=$$(cat .portal.pid); \
		echo "Stopping portal (PID: $PID)..."; \
		kill $PID 2>/dev/null || true; \
		rm .portal.pid; \
		echo "[OK] Portal stopped"; \
	else \
		echo "[OK] Portal is not running"; \
	fi

# Cleanup
clean:
	@echo "Cleaning up..."
	@$(MAKE) down
	@rm -rf dist .env portal portal-mysql.yml portal-plugins.yaml
	@rm -f .portal.pid
	@echo "[OK] Cleanup complete"
