# --- Client -------------------------------------------------------------------
.PHONY: client-unit client-smoke play
client-unit: ## Run Godot headless unit tests (no server required)
	GODOT="$(GODOT)" CLIENT_UNIT_ONLY=1 ./scripts/client_smoke.sh

client-smoke: ## Run Godot headless client smoke test (requires pinned Godot; slice needs server)
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" ./scripts/client_smoke.sh

play: db-up ## Play the slice: Postgres + server + interactive Godot client (close window to stop)
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" ./scripts/play.sh
