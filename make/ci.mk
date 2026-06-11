# --- CI -----------------------------------------------------------------------
.PHONY: maintainability ci
maintainability: ## Run maintainability ratchet checks
	./scripts/check-file-size-ratchet.sh

ci: ## Run the full local CI suite (quiet; VERBOSE=1 for full logs)
	./scripts/ci.sh
