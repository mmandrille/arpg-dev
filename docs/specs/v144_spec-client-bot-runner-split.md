# v144 Spec: Client Bot Runner Split

Status: Complete
Date: 2026-06-13
Codename: `client-bot-runner-split`

## Purpose

Split the long `client/scripts/bot_scenario_runner.gd` step dispatch, assertion dispatch, action
dispatch, and static validation tables into focused helper modules while preserving the public
`BotScenarioRunner` API and all existing client bot scenario semantics.

## Non-goals

- No scenario JSON format changes.
- No new bot step types or assertion semantics.
- No server, protocol, gameplay, shared-rule, or Python bot changes.
- No visual/UI behavior changes.
- No direct `BotController` behavior changes beyond loading the same runner class.

## Acceptance Criteria

- `BotScenarioRunner` remains the public frame-tick runner used by `BotController`.
- Step type catalogs and static `validate_scenario` / `validate_step` logic move out of
  `bot_scenario_runner.gd` into a focused helper.
- Wait-step, assertion-step, and action-step dispatch move out of the runner into focused helper
  modules.
- Existing tests that call `BotScenarioRunner.validate_scenario`, `validate_step`, and step type
  constants keep working through compatibility delegates.
- `client/scripts/bot_scenario_runner.gd` shrinks and its maintainability baseline is lowered.
- Existing 40 client bot scenarios remain green.

## Scope And Likely Files

- `client/scripts/bot_scenario_runner.gd` - keep runner state and helper matchers, delegate
  dispatch/validation.
- `client/scripts/bot_step_catalog.gd` - step type categories and static validation.
- `client/scripts/bot_wait_handlers.gd` - wait-step dispatch.
- `client/scripts/bot_assertion_handlers.gd` - assertion-step dispatch.
- `client/scripts/bot_action_handlers.gd` - action-step dispatch.
- `client/tests/test_client_bot.gd` - keep existing unit coverage green; add focused compatibility
  checks only if needed.
- `.maintainability/file-size-baseline.tsv` - lower runner baseline.
- `docs/CODEMAP.md` and lifecycle docs.

## Shortcut Decision

Reject external plugins/addons for this slice. This is internal bot-test infrastructure with no
player UI, art, camera, inventory presentation, or asset work; an addon would add dependency surface
without replacing the current scenario-runner contract.

## Test And Bot Proof

- `make client-unit`
- `make bot-client HEADLESS=1`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking product/design questions.
- Risk: helper classes accidentally break compatibility for tests or `BotController` callers.
  Mitigation: preserve `BotScenarioRunner` constants/static methods as delegates and run
  client-unit plus all 40 headless client bot scenarios.
