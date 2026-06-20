# --- Database -----------------------------------------------------------------
.PHONY: db-up db-down db-reset
db-up: ## Start local Postgres (canonical DB startup path) and wait for readiness
	@if docker exec arpg-postgres pg_isready -U arpg -d arpg >/dev/null 2>&1; then \
		echo "using existing ready arpg-postgres"; \
	else \
		$(COMPOSE) up -d postgres; \
		echo "waiting for postgres to be ready..."; \
		for i in $$(seq 1 30); do \
			docker exec arpg-postgres pg_isready -U arpg -d arpg >/dev/null 2>&1 && break; \
			sleep 1; \
		done; \
		docker exec arpg-postgres pg_isready -U arpg -d arpg; \
	fi

db-down: ## Stop local Postgres (keep data volume)
	$(COMPOSE) down

db-reset: ## Destroy and recreate local Postgres (drops all data)
	$(COMPOSE) down -v
	$(MAKE) db-up
