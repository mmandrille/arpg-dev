# --- Client -------------------------------------------------------------------
.PHONY: client-unit client-smoke play play-remote
PLAY_CLIENTS_FROM_GOALS := $(firstword $(filter-out play play-remote,$(MAKECMDGOALS)))
PLAY_CLIENTS ?= $(if $(PLAY_CLIENTS_FROM_GOALS),$(PLAY_CLIENTS_FROM_GOALS),1)
ifneq ($(filter play play-remote,$(MAKECMDGOALS)),)
ifneq ($(PLAY_CLIENTS_FROM_GOALS),)
.PHONY: $(PLAY_CLIENTS_FROM_GOALS)
$(PLAY_CLIENTS_FROM_GOALS):
	@:
endif
endif

client-unit: ## Run Godot headless unit tests (no server required)
	GODOT="$(GODOT)" CLIENT_UNIT_ONLY=1 ./scripts/client_smoke.sh

client-smoke: ## Run Godot headless client smoke test (requires pinned Godot; slice needs server)
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" ./scripts/client_smoke.sh

play: db-up ## Play locally: pass a client count as `make play 3` for menu co-op
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" PLAY_CLIENTS="$(PLAY_CLIENTS)" ./scripts/play.sh

play-remote: ## Launch menu client(s) against an existing backend: `BASE_URL=https://... make play-remote 3`
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" PLAY_CLIENTS="$(PLAY_CLIENTS)" ./scripts/play_remote.sh
