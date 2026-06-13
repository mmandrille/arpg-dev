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
ADDR ?= :8888
BASE_URL ?= http://localhost:8888
PLAY_ADDR ?= $(ADDR)
PLAY_BASE_URL ?= $(BASE_URL)
TEST_ADDR ?= :18081
TEST_BASE_URL ?= http://localhost:18081
CI_ADDR ?= $(TEST_ADDR)
CI_BASE_URL ?= $(TEST_BASE_URL)
BOT_ADDR ?= $(TEST_ADDR)
BOT_BASE_URL ?= $(TEST_BASE_URL)
DEV_TOKEN ?= local-dev-token
DEBUG_TOKEN ?= local-debug-token
GODOT ?= godot
SESSION_ID ?=

export ARPG_DATABASE_URL = $(DATABASE_URL)
export ARPG_ADDR = $(ADDR)
export ARPG_DEV_TOKEN = $(DEV_TOKEN)
export ARPG_DEBUG_TOKEN = $(DEBUG_TOKEN)

# Quiet output for agent-friendly runs (test/ci/bot). Full logs: VERBOSE=1 or V=1.
VERBOSE ?=
V ?=
ifneq ($(filter 1 true yes on,$(VERBOSE) $(V)),)
export ARPG_VERBOSE := 1
endif
export ARPG_QUIET_TAIL_LINES ?= 100
RUN_QUIET := $(ROOT)/scripts/run_quiet.sh

.PHONY: help
help: ## List available commands
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "  Output: make test/ci/bot/test-all/client-* are quiet by default. Use VERBOSE=1 (or V=1) for full logs."
	@echo "  Ports: make play defaults to $(PLAY_BASE_URL); CI/bot defaults to $(TEST_BASE_URL). Override with PLAY_ADDR/PLAY_BASE_URL or CI_ADDR/CI_BASE_URL/BOT_ADDR/BOT_BASE_URL."

.PHONY: skill-visual
skill-visual: ## Run a bot-visual replay for a skill: make skill-visual skill=holy_shield rank=1
	@if [[ -z "$(skill)" ]]; then echo "usage: make skill-visual skill=<skill_id>"; exit 2; fi
	@$(PY) -m tools.bot.skill_visual "$(skill)" $(if $(rank),--rank "$(rank)",) $(if $(level),--level "$(level)",) $(if $(DRY_RUN),--dry-run,)

.PHONY: skill-visual-list
skill-visual-list: ## List skill visual replay coverage
	@$(PY) -m tools.bot.skill_visual --list

include make/subfiles.mk
