# --- CI -----------------------------------------------------------------------
.PHONY: maintainability ci
maintainability: ## Run maintainability ratchet checks
	./scripts/check-file-size-ratchet.sh
	python3 ./scripts/check-extraction-coupling-ratchet.py

ci: ## Run the full local CI suite (quiet; VERBOSE=1 for full logs)
	ARPG_ADDR="$(CI_ADDR)" BASE_URL="$(CI_BASE_URL)" ./scripts/ci.sh
