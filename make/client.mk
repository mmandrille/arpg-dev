# --- Client -------------------------------------------------------------------
.PHONY: client-unit client-smoke play play-debug play-remote skill-logo-sheet model-list model
SKILL_LOGO_SHEET_OUT := $(or $(OUT),.artifacts/skill-logo-sheet.svg)
PLAY_CLIENTS_FROM_GOALS := $(firstword $(filter-out play play-debug play-remote,$(MAKECMDGOALS)))
PLAY_CLIENTS ?= $(if $(PLAY_CLIENTS_FROM_GOALS),$(PLAY_CLIENTS_FROM_GOALS),1)
PLAY_MAIL ?= $(mail)
PLAY_MAIL_ENV := $(if $(PLAY_MAIL),ARPG_PLAY_EMAIL="$(PLAY_MAIL)",)
ifneq ($(filter play play-debug play-remote,$(MAKECMDGOALS)),)
ifneq ($(PLAY_CLIENTS_FROM_GOALS),)
.PHONY: $(PLAY_CLIENTS_FROM_GOALS)
$(PLAY_CLIENTS_FROM_GOALS):
	@:
endif
endif

client-unit: ## Run Godot headless unit tests (quiet; VERBOSE=1 for full logs)
	GODOT="$(GODOT)" CLIENT_UNIT_ONLY=1 ./scripts/client_smoke.sh

client-smoke: ## Run Godot headless smoke against a running TEST_BASE_URL server (quiet; VERBOSE=1 for full logs)
	GODOT="$(GODOT)" BASE_URL="$(TEST_BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" DEBUG_TOKEN="$(DEBUG_TOKEN)" ./scripts/client_smoke.sh

skill-logo-sheet: tools ## Render current skill logos and labels to an SVG image
	@mkdir -p "$$(dirname "$(SKILL_LOGO_SHEET_OUT)")"
	$(PY) tools/assets/skill_logo_sheet.py --out "$(abspath $(SKILL_LOGO_SHEET_OUT))"

model-list: tools ## List previewable character/monster model asset IDs
	@$(PY) -m tools.assets.model_catalog list

model: tools ## Preview a model: make model model=<asset_id> [CHECK=1]
	@if [[ -z "$(model)" ]]; then echo "usage: make model model=<asset_id>"; echo "run: make model-list"; exit 2; fi
	@$(PY) -m tools.assets.model_catalog resolve "$(model)" >/dev/null
	@if [[ "$(CHECK)" == "1" ]]; then \
		MODEL_ASSET_ID="$(model)" MODEL_VIEWER_CHECK=1 "$(GODOT)" --headless --path client --scene res://scenes/model_viewer.tscn; \
	else \
		MODEL_ASSET_ID="$(model)" "$(GODOT)" --path client --scene res://scenes/model_viewer.tscn; \
	fi

play: db-up ## Play locally: use `mail=user@example.com` for one account, or `make play 3` for co-op
	GODOT="$(GODOT)" ARPG_ADDR="$(PLAY_ADDR)" BASE_URL="$(PLAY_BASE_URL)" PLAY_CLIENTS="$(PLAY_CLIENTS)" $(PLAY_MAIL_ENV) ./scripts/play.sh

play-debug: db-up ## Play locally with perf logs captured to /tmp/arpg-perf.log
	ARPG_PERF_DEBUG=true GODOT="$(GODOT)" ARPG_ADDR="$(PLAY_ADDR)" BASE_URL="$(PLAY_BASE_URL)" PLAY_CLIENTS="$(PLAY_CLIENTS)" $(PLAY_MAIL_ENV) ./scripts/play.sh 2>&1 | tee /tmp/arpg-perf.log

play-remote: ## Launch menu client(s) against an existing backend: `BASE_URL=https://... make play-remote 3`
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" PLAY_CLIENTS="$(PLAY_CLIENTS)" ./scripts/play_remote.sh
