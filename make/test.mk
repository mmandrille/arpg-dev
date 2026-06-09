# --- Unit tests ---------------------------------------------------------------
# Fast, local test suite: shared contract validation, Go, Python, and Godot
# headless smoke. Does not start Postgres, the server, or the protocol bot —
# use `make ci` for the full integration pipeline. Use `make test-all` for
# every local suite including CI and headless bot-visual.
.PHONY: test test-py test-all
test: ## Run unit tests (quiet; VERBOSE=1 for full logs)
	@$(RUN_QUIET) --label validate-shared $(MAKE) validate-shared
	@$(RUN_QUIET) --label test-go $(MAKE) test-go
	@$(RUN_QUIET) --label test-py $(MAKE) test-py
	@$(RUN_QUIET) --label client-unit $(MAKE) client-unit
	@echo "test OK"

test-py: tools ## Run Python unit tests (tools/)
	$(PY) -m pytest -q tools

test-all: ## Run every test suite (quiet; VERBOSE=1 for full logs)
	./scripts/test_all.sh
