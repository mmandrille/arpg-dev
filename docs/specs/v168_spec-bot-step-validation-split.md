# v168 Spec — Bot Step Validation Split

Status: Complete
Date: 2026-06-14
Codename: `bot-step-validation-split`

## Purpose

Split client bot action-step validation out of `BotStepCatalog.validate_step` into a focused helper.
Existing client scenario DSL validation behavior and error messages must remain unchanged.

## Non-goals

- No new client bot step types.
- No scenario JSON, gameplay, protocol, server, or shared-rule changes.
- No full validation rewrite; wait/assert validation can be split later.

## Acceptance Criteria

- Action-step validations delegate to a focused helper module.
- Existing invalid-step unit coverage in `client/tests/test_client_bot.gd` keeps passing.
- `client/scripts/bot_step_catalog.gd` shrinks while preserving `validate_step` as the public API.
- `make client-unit`, `make maintainability`, and `make ci` pass.

## Scope and Likely Files

- Client bot: `client/scripts/bot_step_catalog.gd`
- New helper: `client/scripts/bot_action_step_validator.gd`
- Tests: existing `client/tests/test_client_bot.gd`
- Docs: `PROGRESS.md`, `docs/as-built/v168_bot-step-validation-split.md`

## Test and Bot Proof

This is behavior-preserving client bot validation work. Existing client bot unit tests cover invalid
action steps and full CI covers scenario loading.

## Open Questions and Risks

- Risk: helper dispatch can accidentally skip an action validation. Preserve exact error messages
  and keep `validate_step` as the only public caller.
