# --- Agents -------------------------------------------------------------------
.PHONY: bot bot-visual bot-client replay
bot: db-up tools ## Run all Python protocol bot scenarios with local Postgres + temporary server
	DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" \
	SCENARIO="$(or $(SCENARIO),$(scenario),all)" ./scripts/bot_local.sh

bot-visual: db-up tools ## Record bot scenario(s) and open a Godot visual replay playlist; pass scenario=<id|file>
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" SCENARIO="$(or $(SCENARIO),$(scenario),all)" ./scripts/bot_visual.sh

bot-client: db-up ## Run Godot client bot scenarios with a temporary local server; pass scenario=<id|all>, HEADLESS=1
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" \
	SCENARIO="$(or $(SCENARIO),$(scenario),all)" \
	HEADLESS="$(or $(HEADLESS),0)" ./scripts/bot_client_local.sh

replay: ## Verify a recorded session by re-simulating from seed + inputs
	@test -n "$(SESSION_ID)" || { echo "usage: make replay SESSION_ID=<session-id>"; exit 2; }
	cd $(SERVER_DIR) && go run ./cmd/arpg-replay --session-id "$(SESSION_ID)"
