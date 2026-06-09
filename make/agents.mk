# --- Agents -------------------------------------------------------------------
.PHONY: bot bot-visual bot-client replay
bot: db-up tools ## Run all Python protocol bot scenarios (quiet; VERBOSE=1 for full logs)
	DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" \
	SCENARIO="$(or $(SCENARIO),$(scenario),all)" ./scripts/bot_local.sh

bot-visual: db-up tools ## Record bot scenarios + Godot replay (quiet when headless; VERBOSE=1 for full logs)
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" SCENARIO="$(or $(SCENARIO),$(scenario),all)" ./scripts/bot_visual.sh

bot-client: db-up ## Run Godot client bot scenarios (quiet when HEADLESS=1; VERBOSE=1 for full logs)
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" \
	SCENARIO="$(or $(SCENARIO),$(scenario),all)" \
	HEADLESS="$(or $(HEADLESS),0)" ./scripts/bot_client_local.sh

replay: ## Verify a recorded session by re-simulating from seed + inputs
	@test -n "$(SESSION_ID)" || { echo "usage: make replay SESSION_ID=<session-id>"; exit 2; }
	cd $(SERVER_DIR) && go run ./cmd/arpg-replay --session-id "$(SESSION_ID)"
