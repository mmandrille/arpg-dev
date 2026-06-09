# --- CI -----------------------------------------------------------------------
.PHONY: ci
ci: ## Run the full local CI suite (quiet; VERBOSE=1 for full logs)
	./scripts/ci.sh
