# v168 As-built — Bot Step Validation Split

Date: 2026-06-14

## What shipped

- Added `client/scripts/bot_action_step_validator.gd` for client bot action-step validation.
- Kept `BotStepCatalog.validate_step` as the public validation API and delegated handled action
  steps to the helper.
- Preserved existing scenario DSL step names and validation error strings for moved action checks.

## Proof

- `make client-unit` passed.
- `client/scripts/bot_step_catalog.gd` dropped from 322 to 250 lines; the new action validator is
  106 lines.

## Deferred

- Wait-step and assertion-step validation remain in `BotStepCatalog.validate_step` and can be split
  into focused helpers later.
