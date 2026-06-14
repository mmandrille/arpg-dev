# v143 As-built: Client Bot Facade

Spec: [`docs/specs/v143_spec-client-bot-facade.md`](../specs/v143_spec-client-bot-facade.md)
Plan: [`docs/plans/v143_2026-06-13-client-bot-facade.md`](../plans/v143_2026-06-13-client-bot-facade.md)

## What shipped

- Added `client/scripts/bot_facade.gd` as the focused home for bot panel/action adapter helpers.
- Kept the existing public `main.gd` `bot_*` method names and signatures, with thin delegates to
  `BotFacade` so `BotController` remains unchanged.
- Moved shop, stash, bishop, blacksmith, market, hotbar, stat, skill-bar, and directional skill-cast
  adapter bodies out of `main.gd`.
- Added `client/tests/test_bot_facade.gd` with fake panel/main coverage and wired it into
  `scripts/client_smoke.sh`.
- Lowered the `client/scripts/main.gd` maintainability baseline from 6769 to 6703 lines.
- Updated CODEMAP so Bot / scenarios work points at `bot_facade.gd`.

## Proof

- `make client-unit`
- `.venv/bin/python tools/validate_codemap.py`
- `make maintainability`
- `make bot-client HEADLESS=1` (40 passed, 0 failed)
- `make ci`

## Deferred

- Splitting `client/scripts/bot_scenario_runner.gd` remains v144.
- Rebinding `BotController` directly to the facade is deferred; `main.gd` keeps compatibility
  wrappers for the existing bot API.
