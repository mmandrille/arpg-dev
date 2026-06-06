# --- Unit tests ---------------------------------------------------------------
# Fast, local test suite: shared contract validation, Go, Python, and Godot
# headless smoke. Does not start Postgres, the server, or the protocol bot —
# use `make ci` for the full integration pipeline.
.PHONY: test test-py
test: ## Run unit tests (shared validation, Go, Python, client unit)
	$(MAKE) validate-shared
	$(MAKE) test-go
	$(MAKE) test-py
	$(MAKE) client-unit

test-py: tools ## Run Python unit tests (tools/)
	$(PY) -m pytest -q tools
