# --- Agents -------------------------------------------------------------------
.PHONY: bot replay
bot: tools ## Run the Python protocol bot end-to-end against a running server
	$(PY) -m tools.bot.run --base-url "$(BASE_URL)" --dev-token "$(DEV_TOKEN)" --debug-token "$(DEBUG_TOKEN)"

replay: ## Verify a recorded session by re-simulating from seed + inputs
	@test -n "$(SESSION_ID)" || { echo "usage: make replay SESSION_ID=<session-id>"; exit 2; }
	cd $(SERVER_DIR) && go run ./cmd/arpg-replay --session-id "$(SESSION_ID)"
