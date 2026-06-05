# --- Python tooling -----------------------------------------------------------
.PHONY: tools
tools: $(VENV)/.installed ## Create the Python venv and install pinned tooling
$(VENV)/.installed: pyproject.toml
	python3 -m venv $(VENV)
	$(PIP) install --upgrade pip >/dev/null
	$(PIP) install -e ".[dev]"
	touch $(VENV)/.installed
