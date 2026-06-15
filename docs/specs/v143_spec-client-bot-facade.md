# v143 Spec: Client Bot Facade

Status: Complete
Date: 2026-06-13
Codename: `client-bot-facade`

## Purpose

Extract the client bot panel/action adapter code from `client/scripts/main.gd` into a focused
`client/scripts/bot_facade.gd` helper while preserving the public `bot_*` methods that
`BotController` already calls. This shrinks the main scene coordinator and creates a safer home for
future client-bot adapter work.

## Non-goals

- No client bot scenario semantics or JSON step changes.
- No server, protocol, gameplay, or shared-rule changes.
- No split of `bot_scenario_runner.gd`; that remains v144.

## Acceptance Criteria

- `client/scripts/bot_facade.gd` exists as a focused `class_name BotFacade` helper.
- The shop, stash, bishop, blacksmith, market, hotbar, stat, skill, and skill-direction bot adapter
  implementations move out of `main.gd` into `BotFacade`.
- `main.gd` keeps the same public `bot_*` method names used by `BotController`, but those methods
  delegate to `BotFacade`.
- A headless GDScript unit test covers the facade with fake panels/main objects.
- `client/scripts/main.gd` shrinks and its ratchet baseline is lowered.
- Existing 40 client bot scenarios remain green.

## Scope And Likely Files

- `client/scripts/main.gd` - replace adapter bodies with thin delegating wrappers.
- `client/scripts/bot_facade.gd` - new focused helper.
- `client/tests/test_bot_facade.gd` - fake-panel unit coverage.
- `scripts/client_smoke.sh` - add the new unit test to the client-unit gate.
- `.maintainability/file-size-baseline.tsv` - lower `main.gd` baseline.
- `docs/CODEMAP.md` - point Bot / scenarios at `bot_facade.gd`.
- Lifecycle docs.

## Shortcut Decision

surface without replacing any domain logic.

## Test And Bot Proof

- `make client-unit`
- `make bot-client HEADLESS=1`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking product/design questions.
- Risk: thin wrappers accidentally diverge from current `BotController` method expectations.
  Mitigation: keep wrapper names/signatures unchanged and run all client bot scenarios.
