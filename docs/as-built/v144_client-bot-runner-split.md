# v144 As-built: Client Bot Runner Split

Spec: [`docs/specs/v144_spec-client-bot-runner-split.md`](../specs/v144_spec-client-bot-runner-split.md)
Plan: [`docs/plans/v144_2026-06-13-client-bot-runner-split.md`](../plans/v144_2026-06-13-client-bot-runner-split.md)

## What shipped

- Added `client/scripts/bot_step_catalog.gd` for client bot step type categories and static
  scenario/step validation.
- Added `client/scripts/bot_wait_handlers.gd`, `client/scripts/bot_assertion_handlers.gd`, and
  `client/scripts/bot_action_handlers.gd` for the runner's wait/assert/action dispatch chains.
- Kept `BotScenarioRunner` as the public class used by `BotController`, with compatibility
  constants and static validation delegates.
- Left the existing runner-owned matchers, memory, pending action storage, and failure formatting
  intact behind helper calls.
- Lowered the `client/scripts/bot_scenario_runner.gd` maintainability baseline from 2376 to 1665
  lines.
- Updated CODEMAP so bot/scenario work points at the new helper modules.

## Proof

- `make client-unit`
- `.venv/bin/python tools/validate_codemap.py`
- `make maintainability`
- `make bot-client HEADLESS=1` (40 passed, 0 failed)
- `make ci`

## Deferred

- Domain-specific matcher extraction for shop/stash/market/progression helper groups remains
  future paydown.
- Python `tools/bot/run.py` assertion dispatch remains v145.
