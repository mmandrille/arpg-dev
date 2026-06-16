# v213 As-Built: Bishop Account Revival

Date: 2026-06-16

## What Shipped

- Bishop respec is now free through shared `main_config.v0.json`.
- Added `bishop_revive_all_intent` and `bishop_revive_all` event.
- Realtime persistence revives every dead character on the actor account by clearing `dead` and `death_level`.
- Added a separate Bishop panel button for reviving account heroes.
- Updated Bishop bot/client scenario assertions for free respec and the revive-all button.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game ./internal/inputdecode ./internal/realtime ./internal/replay`
- `cd server && go test ./internal/store -run 'TestReviveDeadCharactersIsAccountScoped'`
- `make client-unit`
- `make bot scenario=town_bishop_respec BOT_ADDR=:18082 BOT_BASE_URL=http://localhost:18082`
- `make bot-client scenario=town_bishop_respec_panel HEADLESS=1`
- `make ci`

## Scope Limits

No per-character revive picker, new Bishop VFX/audio, revive pricing, or corpse item recovery changes shipped here.
