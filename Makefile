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

include make/subfiles.mk
