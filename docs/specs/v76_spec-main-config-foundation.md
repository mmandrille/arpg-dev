# v76 Spec — Main config foundation

Status: Complete
Date: 2026-06-11
Codename: `main-config-foundation`
Baseline: v75 `persistent-window-layout` complete on `main`; post-v75 loot tuning is already committed.

## Purpose

Introduce a single shared `main_config.v0.json` rules file for top-level gameplay tuning values that designers should be able to edit directly. This first slice creates the contract, validates it, loads it into the Go rules view, and cross-checks it against the existing combat, navigation, and loot defaults without changing gameplay semantics yet.

## Non-goals

- No removal of existing `combat.v0.json`, `navigation.v0.json`, or treasure-class fields in this slice.
- No protocol/schema version bump.
- No client UI or Godot presentation changes.
- No class, skill, monster, or loot-profile refactor beyond recording the first global values.

## Acceptance criteria

- `shared/rules/main_config.v0.json` exists with a schema and validates through `make validate-shared`.
- The config includes at least:
  - `base_attack_interval_ticks`
  - `base_movement_speed`
  - `base_drop_rate_percent`
- The Go `Rules` struct exposes the loaded main config.
- Rule loading rejects invalid main config values.
- Shared validation fails if the main config drifts from the current combat attack interval, navigation movement speed, or dungeon monster drop rate defaults.
- Existing gameplay behavior remains unchanged.

## Scope and likely files

- `shared/rules/main_config.v0.json`
- `shared/rules/main_config.v0.schema.json`
- `server/internal/game/rules.go`
- `tools/validate_shared.py`
- `server/internal/game/game_test.go`
- `docs/specs/v76_spec-main-config-foundation.md`
- `docs/plans/v76_2026-06-11-main-config-foundation.md`
- `PROGRESS.md`
- `docs/as-built/v76_main-config-foundation.md`

## Test and bot proof

- `make validate-shared`
- Focused Go rules loader tests under `server/internal/game`
- `make ci` during finish

Bot proof is not required because this slice only introduces a validated shared rules file and typed loader view; gameplay behavior is intentionally unchanged.

## Open questions and risks

- The follow-up slices must move consumers to the config and then replace repeated loot rate weights with profiles. Until then, config edits are guarded by drift validation rather than being live tuning knobs.
