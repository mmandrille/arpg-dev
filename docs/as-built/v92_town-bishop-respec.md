# v92 As-Built - Town Bishop Respec

## What Shipped

- Added a data-driven `respec_cost_gold` value in shared main config, currently 250 gold.
- Added `town_bishop` as a ready interactable service and placed it in town plus `vendor_lab`.
- Added `bishop_respec_intent`, `bishop_service_opened`, and `bishop_respec` protocol coverage.
- Implemented server-authoritative bishop activation that restores HP and mana and reports respec affordability.
- Implemented paid respec: deducts gold, resets class base stats, refunds level-earned stat points, refunds spent skill ranks, clears skill cooldowns, updates resources, and emits progression/skill/gold/resource changes.
- Added focused Go tests for heal-on-open, successful respec, class baseline reset, and unaffordable rejection.
- Added protocol bot coverage for successful and rejected respec paths.
- Added a red non-merchant bishop primitive model, compact bishop service panel, client intent wiring, and headless client-bot panel proof.
- Updated the file-size ratchet baseline with a documented exception for existing integration hotspots touched by the slice: `main.gd`, `handlers.go`, `sim.go`, and `tools/bot/run.py`.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game/... ./internal/inputdecode/...`
- `make bot scenario=town_bishop_respec`
- `make client-unit`
- `make bot-client scenario=town_bishop_respec_panel HEADLESS=1`
- `make maintainability`
- `make ci`

## Follow-Ups

- Optional visual review: `make bot-visual scenario=45_town_bishop_respec.json`.
- Add richer bishop art/dialog only when the town service set grows beyond this one-option menu.
