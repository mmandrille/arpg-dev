# arpg-dev developer commands.
# All canonical dev workflows for the first playable vertical slice run through
# this Makefile so humans and agents share one entrypoint.
SHELL := /bin/bash
ROOT := $(shell pwd)
VENV := $(ROOT)/.venv
PY := $(VENV)/bin/python
PIP := $(VENV)/bin/pip
COMPOSE := docker-compose
SERVER_DIR := $(ROOT)/server

# Connection + runtime settings (override on the command line as needed).
DATABASE_URL ?= postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable
ADDR ?= :8080
DEV_TOKEN ?= local-dev-token
DEBUG_TOKEN ?= local-debug-token
BASE_URL ?= http://localhost:8080
GODOT ?= godot
SESSION_ID ?=

export ARPG_DATABASE_URL = $(DATABASE_URL)
export ARPG_ADDR = $(ADDR)
export ARPG_DEV_TOKEN = $(DEV_TOKEN)
export ARPG_DEBUG_TOKEN = $(DEBUG_TOKEN)

.PHONY: help
help: ## List available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# --- Database -----------------------------------------------------------------
.PHONY: db-up db-down db-reset
db-up: ## Start local Postgres (canonical DB startup path) and wait for readiness
	$(COMPOSE) up -d postgres
	@echo "waiting for postgres to be ready..."
	@for i in $$(seq 1 30); do \
		docker exec arpg-postgres pg_isready -U arpg -d arpg >/dev/null 2>&1 && break; \
		sleep 1; \
	done
	@docker exec arpg-postgres pg_isready -U arpg -d arpg

db-down: ## Stop local Postgres (keep data volume)
	$(COMPOSE) down

db-reset: ## Destroy and recreate local Postgres (drops all data)
	$(COMPOSE) down -v
	$(MAKE) db-up

# --- Python tooling -----------------------------------------------------------
.PHONY: tools
tools: $(VENV)/.installed ## Create the Python venv and install pinned tooling
$(VENV)/.installed: pyproject.toml
	python3 -m venv $(VENV)
	$(PIP) install --upgrade pip >/dev/null
	$(PIP) install -e ".[dev]"
	touch $(VENV)/.installed

# --- Server -------------------------------------------------------------------
.PHONY: server migrate test-go
server: ## Run the Go server against local Postgres
	cd $(SERVER_DIR) && go run ./cmd/arpg-server

migrate: ## Apply database migrations (server also self-migrates on boot)
	cd $(SERVER_DIR) && go run ./cmd/arpg-server -migrate-only

test-go: ## Run all Go tests
	cd $(SERVER_DIR) && go test ./...

# --- Shared contracts ---------------------------------------------------------
.PHONY: validate-shared validate-assets
validate-shared: tools ## Validate all shared JSON (protocol, rules, golden) against schemas
	$(PY) tools/validate_shared.py

validate-assets: tools ## Validate the asset manifest, runtime .glb paths, and GLB nodes
	$(PY) tools/assets/validate_assets.py

gen-assets: tools ## Regenerate committed runtime .glb files (deterministic source-of-truth)
	$(PY) tools/assets/gen_glb.py

# --- Agents -------------------------------------------------------------------
.PHONY: bot replay
bot: tools ## Run the Python protocol bot end-to-end against a running server
	$(PY) -m tools.bot.run --base-url "$(BASE_URL)" --dev-token "$(DEV_TOKEN)" --debug-token "$(DEBUG_TOKEN)"

replay: ## Verify a recorded session by re-simulating from seed + inputs
	@test -n "$(SESSION_ID)" || { echo "usage: make replay SESSION_ID=<session-id>"; exit 2; }
	cd $(SERVER_DIR) && go run ./cmd/arpg-replay --session-id "$(SESSION_ID)"

# --- Client -------------------------------------------------------------------
.PHONY: client-smoke
client-smoke: ## Run the Godot headless client smoke test (requires pinned Godot)
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" ./scripts/client_smoke.sh

# --- CI -----------------------------------------------------------------------
.PHONY: ci
ci: ## Run the full local CI suite (shared validation, Go tests, bot, replay)
	./scripts/ci.sh
