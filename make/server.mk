# --- Server -------------------------------------------------------------------
.PHONY: server migrate test-go
server: ## Run the Go server against local Postgres
	cd $(SERVER_DIR) && go run ./cmd/arpg-server

migrate: ## Apply database migrations (server also self-migrates on boot)
	cd $(SERVER_DIR) && go run ./cmd/arpg-server -migrate-only

test-go: ## Run all Go tests
	cd $(SERVER_DIR) && go test ./...
