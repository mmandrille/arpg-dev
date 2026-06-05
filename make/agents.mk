# --- Agents -------------------------------------------------------------------
.PHONY: bot bot-visual replay
bot: tools ## Run all Python protocol bot scenarios against a running server
	$(PY) -m tools.bot.run --base-url "$(BASE_URL)" --dev-token "$(DEV_TOKEN)" --debug-token "$(DEBUG_TOKEN)"

bot-visual: db-up tools ## Record all bot scenarios and open a Godot visual replay playlist
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" ./scripts/bot_visual.sh

replay: ## Verify a recorded session by re-simulating from seed + inputs
	@test -n "$(SESSION_ID)" || { echo "usage: make replay SESSION_ID=<session-id>"; exit 2; }
	cd $(SERVER_DIR) && go run ./cmd/arpg-replay --session-id "$(SESSION_ID)"
