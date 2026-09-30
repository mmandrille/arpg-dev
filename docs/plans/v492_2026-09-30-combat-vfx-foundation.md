# v492 Plan — Combat VFX foundation

Status: Implemented
Spec: [`docs/specs/v492_spec-combat-vfx-foundation.md`](../specs/v492_spec-combat-vfx-foundation.md)

## Baseline and shortcut decision
Hooks the existing `GameplayFeedbackPresentation.play_entity_reaction` (hit/death) and
`SceneLightingRig.sync` (quality), so `main.gd` is not touched. `HealRainEffect` keeps its API.

## Maintenance ratchet
No grandfathered file touched.

## Task 1 — Data
- [x] `vfx_presentation.v0.json` + schema (effects, damage_type colours); `particle_scale` on quality tiers.
```bash
make validate-shared
```
## Task 2 — Client
- [x] Shader; `CombatVfx` (loader, `make_burst`, `spawn_for_reaction`, `set_quality`).
- [x] Hook reaction + lighting rig; rewrite `HealRainEffect` on particles.
- [x] `test_combat_vfx.gd` registered; heal-rain tests green.
```bash
make client-unit
```
## Task 3 — Visual gate
- [x] Throwaway capture of sparks per damage type + death burst; `scenes/heal-rain` before/after.
## Task 4 — Regression + docs + CI
- [x] Combat client scenarios; docs; `make ci`.

## Execution notes

- Capture tuning: lower glow energies and more saturated colours (additive colour × energy clipped to white under the tonemapper); death burst darker and wider; heal rain prewarmed with `preprocess` and larger motes.
- The spark/death capture is a throwaway scratch script (not committed).
