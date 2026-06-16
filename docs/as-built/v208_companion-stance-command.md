# v208 As-Built: Companion Stance Command

Date: 2026-06-16
Status: Complete

## What shipped

- Added server-authoritative `companion_command_intent` with `assist`, `defend`, and `passive` stance validation.
- Companions now expose `companion_stance` in entity snapshots/deltas and default to `assist`.
- Successful stance commands update all living owned companions on the active level and emit `companion_stance_changed` with stance and affected count.
- Companion AI now honors stance:
  - `assist` keeps the existing near-companion targeting behavior.
  - `defend` targets monsters near the owner.
  - `passive` clears targets and prevents companion attacks while preserving follow behavior.
- Added shared protocol v8 schema/example coverage for the new command, entity field, and event payload.
- Added protocol bot support for `set_companion_stance` and a `companion_stance_command` scenario.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/inputdecode ./internal/game -run 'TestDecodeCompanionCommandIntent|TestCompanionStance'`
- `make bot scenario=companion_stance_command`
- `make bot scenario=mercenary_hiring_board`
- `make maintainability`

- `make ci`

## Deferred

Godot stance controls, per-companion commands, hold-position/retreat behavior, durable stance persistence, mercenary death/loss rules, gear snapshot refresh, loot/XP/potion behavior, and pricing/listing models remain deferred.
