# v77 Spec — Main config derived gameplay

Status: Complete

## Goal

Make the first `main_config.v0.json` gameplay values operational instead of mirror-only. Changing `base_attack_interval_ticks` or `base_movement_speed` in `main_config.v0.json` should affect server gameplay immediately without also editing older combat/navigation rule files or unrelated test expectations.

## Scope

- Server combat cadence consumes `main_config.gameplay.base_attack_interval_ticks`.
- Server direct and auto movement consume `main_config.gameplay.base_movement_speed`.
- Shared validation no longer requires `combat.v0.json.base_attack_interval_ticks` or `navigation.v0.json.cell_size` to mirror `main_config`.
- Focused Go tests prove changing only `MainConfig` changes derived attack interval and movement distance.
- Existing authored defaults remain unchanged.

## Out of Scope

- Drop profile unification; that is v78.
- Moving class, skill, monster, or item catalogs under `main_config`.
- Rewriting navigation grid semantics. `navigation.cell_size` remains the pathfinding/grid unit.

## Acceptance

- `make validate-shared` passes.
- Focused Go tests prove main-config attack and movement tuning.
- `make ci` passes before commit.
