# v84 Spec - Client Bot Step Registry

Status: Complete
Date: 2026-06-11
Codename: `client-bot-step-registry`

## Purpose

The Godot client bot runner keeps step types in category-specific lists and a separate
`ALL_STEP_TYPES` list. The v80 review called out that this duplication makes every new client bot
verb require multiple manual edits. This slice makes `ALL_STEP_TYPES` derived from the wait/assert
and action registries while preserving existing validation behavior.

## Non-goals

- No new client bot step verbs.
- No scenario JSON behavior changes.
- No runtime gameplay, protocol, or UI behavior changes.
- No broad bot runner extraction.

## Acceptance Criteria

- `ALL_STEP_TYPES` is derived from `STEP_TYPES_WAIT`, `STEP_TYPES_ASSERT`, and `STEP_TYPES_ACTION`.
- Unknown step validation still rejects unsupported step types with the existing error style.
- Existing step validation coverage stays green.
- A focused unit assertion proves the derived registry contains all category entries.
- `make client-unit`, `make maintainability`, and `make ci` pass before commit.

## Scope and Likely Files

- `client/scripts/bot_scenario_runner.gd`
- `client/tests/test_client_bot.gd`
- `docs/plans/v84_2026-06-11-client-bot-step-registry.md`
- `docs/as-built/v84_client-bot-step-registry.md`
- `PROGRESS.md`

## Test and Bot Proof

- `make client-unit`
- `make maintainability`
- `make ci`

No protocol bot scenario is required because this is client bot validation infrastructure only.

## Open Questions and Risks

- No blocking questions.
- Risk: GDScript constant expression support for array concatenation must work in headless import.
  If not, use a tiny static helper while keeping validation behavior unchanged.
