# --- Client -------------------------------------------------------------------
.PHONY: client-unit client-smoke play play-remote skill-logo-sheet
SKILL_LOGO_SHEET_OUT := $(or $(OUT),.artifacts/skill-logo-sheet.svg)
PLAY_CLIENTS_FROM_GOALS := $(firstword $(filter-out play play-remote,$(MAKECMDGOALS)))
PLAY_CLIENTS ?= $(if $(PLAY_CLIENTS_FROM_GOALS),$(PLAY_CLIENTS_FROM_GOALS),1)
PLAY_MAIL ?= $(mail)
PLAY_MAIL_ENV := $(if $(PLAY_MAIL),ARPG_PLAY_EMAIL="$(PLAY_MAIL)",)
ifneq ($(filter play play-remote,$(MAKECMDGOALS)),)
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

play: db-up ## Play locally: use `mail=user@example.com` for one account, or `make play 3` for co-op
	GODOT="$(GODOT)" ARPG_ADDR="$(PLAY_ADDR)" BASE_URL="$(PLAY_BASE_URL)" PLAY_CLIENTS="$(PLAY_CLIENTS)" $(PLAY_MAIL_ENV) ./scripts/play.sh

play-remote: ## Launch menu client(s) against an existing backend: `BASE_URL=https://... make play-remote 3`
	GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" PLAY_CLIENTS="$(PLAY_CLIENTS)" ./scripts/play_remote.sh
