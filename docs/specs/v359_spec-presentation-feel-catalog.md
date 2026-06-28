# v359 Spec: Presentation Feel Catalog

Status: Approved
Date: 2026-06-28
Codename: `presentation-feel-catalog`
Baseline: v358 `scenario-movement-decoupling`

## Purpose

Centralize client-only combat/movement feel microconstants from `CombatFeelConfig` into
`shared/assets/combat_feel_presentation.v0.json` so presentation tuning has a shared home
without crossing the server authority boundary.

## Non-goals

- No server, protocol, shared rules combat tuning, or golden changes.
- No gameplay balance changes (`base_attack_interval_ticks`, damage, costs).
- No attack-animation speed scaling (v362).

## Acceptance criteria

- `shared/assets/combat_feel_presentation.v0.json` + schema own the values currently in
  `combat_feel_config.gd` (input buffer, retarget grace, movement smoothing, melee lunge,
  level-loading min display).
- `CombatFeelPresentationLoader` loads the catalog; `CombatFeelConfig` remains the stable
  consumer-facing facade for existing scripts.
- `make validate-shared` passes.
- Focused Godot unit test proves loader defaults and JSON override behavior.

## Scope and files

| Area | Files |
|------|-------|
| Shared | `combat_feel_presentation.v0.json`, schema |
| Client | `combat_feel_presentation_loader.gd`, `combat_feel_config.gd` |
| Tests | `test_combat_feel_presentation_loader.gd`, `client_smoke.sh` |
| Docs | plan, as-built, lifecycle |

## Test and bot proof

```bash
make validate-shared
make client-unit
make maintainability
```
