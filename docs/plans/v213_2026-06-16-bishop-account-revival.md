# v213 Plan: Bishop Account Revival

Date: 2026-06-16

## Tasks

- [x] Set `respec_cost_gold` to `0` and update Bishop respec tests/scenarios.
- [x] Add account-scoped store API to clear `dead` and `death_level` for dead characters.
- [x] Add `bishop_revive_all_intent` decoding, sim validation, event emission, and realtime persistence hook.
- [x] Add protocol schema support for the intent and event.
- [x] Add a separate Bishop panel button and client/bot wiring.
- [x] Verify with targeted Go tests, shared validation, and client unit tests.

## Targeted Verification

- `make validate-shared`
- `cd server && go test ./internal/game ./internal/inputdecode ./internal/realtime ./internal/replay`
- `cd server && go test ./internal/store -run 'TestReviveDeadCharactersIsAccountScoped'`
- `make client-unit`
- `make bot scenario=town_bishop_respec`
- `make bot-client scenario=town_bishop_respec_panel HEADLESS=1`
