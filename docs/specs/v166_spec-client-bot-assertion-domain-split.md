# v166 Spec — Client Bot Assertion Domain Split

Status: Complete
Date: 2026-06-14
Codename: `client-bot-assertion-domain-split`

## Purpose

Reduce client bot assertion dispatcher breadth by moving a coherent UI/menu assertion domain out of
`BotAssertionHandlers` into a focused helper module. Existing client bot scenario syntax and
assertion behavior must remain unchanged.

## Non-goals

- No new client bot scenario DSL steps.
- No gameplay, protocol, server, or shared-rule changes.
- No broad rewrite of every assertion domain in one slice.

## Acceptance Criteria

- `client/scripts/bot_assertion_handlers.gd` delegates UI/menu assertion types to a focused helper.
- Existing assertion step names keep the same behavior and failure messages where practical.
- `client/tests/test_client_bot.gd` continues to cover menu/session/panel assertions.
- `make client-unit`, `make maintainability`, and `make ci` pass.

## Scope and Likely Files

- Client bot: `client/scripts/bot_assertion_handlers.gd`
- New helper: `client/scripts/bot_ui_assertion_handlers.gd`
- Client tests: existing `client/tests/test_client_bot.gd`
- Docs: `PROGRESS.md`, `docs/as-built/v166_client-bot-assertion-domain-split.md`

## Test and Bot Proof

This is a behavior-preserving client bot architecture slice. Proof comes from existing client bot
unit coverage and full CI; no new runtime gameplay bot scenario is required.

## Open Questions and Risks

- Risk: a delegated assertion may report a different failure string. Keep moved bodies intact unless
  a helper boundary requires a small wrapper.
