# --- CI -----------------------------------------------------------------------
.PHONY: ci
ci: ## Run the full local CI suite (shared validation, Go tests, bot, replay)
	./scripts/ci.sh
