# v359 As-Built — Presentation Feel Catalog

Date: 2026-06-28  
Spec: [`docs/specs/v359_spec-presentation-feel-catalog.md`](../specs/v359_spec-presentation-feel-catalog.md)

## Shipped

- `shared/assets/combat_feel_presentation.v0.json` + schema centralize client-only feel tuning.
- `CombatFeelPresentationLoader` loads the catalog; `CombatFeelConfig` remains the facade with runtime accessors.
- Consumers (`combat_input_buffer`, `command_retarget_grace`, `melee_lunge_presentation`, `movement_visual_smoothing`, `main.gd`) read values after `ensure_loaded()`.

## Verification

```bash
make validate-shared
make client-unit
make maintainability
```
