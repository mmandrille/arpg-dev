# --- Unit tests ---------------------------------------------------------------
# Fast, local test suite: shared contract validation, Go, Python, and Godot
# headless smoke. Does not start Postgres, the server, or the protocol bot —
# use `make ci` for the full integration pipeline. Use `make test-all` for
# every local suite including CI and headless bot-visual.
.PHONY: test test-py test-all
test: ## Run unit tests (shared validation, Go, Python, client unit)
	$(MAKE) validate-shared
	$(MAKE) test-go
	$(MAKE) test-py
	$(MAKE) client-unit

test-py: tools ## Run Python unit tests (tools/)
	$(PY) -m pytest -q tools

test-all: ## Run every test suite (make test + make ci + headless bot-visual)
	./scripts/test_all.sh
