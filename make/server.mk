# --- Server -------------------------------------------------------------------
.PHONY: server migrate test-go lint-determinism wall-grid-audit
server: ## Run the Go server against local Postgres
	cd $(SERVER_DIR) && go run ./cmd/arpg-server

migrate: ## Apply database migrations (server also self-migrates on boot)
	cd $(SERVER_DIR) && go run ./cmd/arpg-server -migrate-only

test-go: ## Run all Go tests
	cd $(SERVER_DIR) && go test ./...

lint-determinism: ## Check game/ for determinism violations (time.Now, math/rand, bare map ranges)
	cd $(SERVER_DIR) && go run ./cmd/determinism-lint ./internal/game/...

wall-grid-audit: ## Report dungeon wall-rect grid alignment (ADR-0018 D6) -> .artifacts/wall-grid-audit.json
	@mkdir -p $(ROOT)/.artifacts
	cd $(SERVER_DIR) && ARPG_WALL_GRID_AUDIT_OUT=$(ROOT)/.artifacts/wall-grid-audit.json \
		ARPG_WALL_GRID_EXTRA_TILES="$(EXTRA_TILES)" \
		go test ./internal/game/ -run '^TestDungeonWallGridAudit$$' -count=1

regen-golden: ## Regenerate golden fixtures from current sim output (run after intentional formula changes)
	cd $(SERVER_DIR) && go test ./internal/game/... -update -run Golden -v 2>&1 | grep -E 'updated golden|PASS|FAIL'
