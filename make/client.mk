# --- Client -------------------------------------------------------------------
.PHONY: client-smoke play
client-smoke: ## Run the Godot headless client smoke test (requires pinned Godot)
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" ./scripts/client_smoke.sh

play: db-up ## Play the slice: Postgres + server + interactive Godot client (close window to stop)
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" ./scripts/play.sh
