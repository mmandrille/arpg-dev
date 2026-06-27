# v355 As-built — Aura Soft Lights

Codename: `aura-soft-lights`  
Date: 2026-06-26

## Shipped

- `shared/assets/aura_light_presentation.v0.json` (+ schema) owns per-aura color, energy,
  attenuation, height, radius source, hero/monster multipliers, priority order, and holy-shield
  cast-pulse tuning.
- `client/scripts/aura_light_presentation_loader.gd` + `client/scripts/aura_soft_lights.gd` replace
  emissive aura mesh markers with one `OmniLight3D` (`AuraSoftLight`) per entity.
- `player_status_effect_markers.gd` keeps debuff mesh factories only; aura `has_*` checks delegate to
  light state.
- Holy Shield cast feedback is a brief omni energy/range tween (debug:
  `holy_shield_cast_pulses` / `holy_shield_target_pulses`).
- Bot debug exposes `active_aura_id`, `aura_light_range`, and `aura_light_color`.
- Aura lights use low energy, soft falloff, and zero specular so they **tint** the floor/entities
  inside their radius instead of brightly illuminating them.

## Proof

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_aura_soft_lights.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_status_effect_presentation.gd
make client-unit
HEADLESS=1 make bot-client SCENARIO=34_elite_aura_readability
HEADLESS=1 make bot-client SCENARIO=37_elite_aura_radius_preview
HEADLESS=1 make bot-client SCENARIO=89_aura_soft_lights
make maintainability
make ci
```

Visual spot-check (starts in fogged dungeon level -1, hero Rage red aura + elite follower cyan aura):

```bash
make bot-visual scenario=89_aura_soft_lights
AUTOPLAY_STEP_DELAY=0.8 make bot-visual scenario=89_aura_soft_lights
```

## Deferred

- Debuff mesh/tint presentation unchanged.
- Production aura VFX/audio, nameplates, minimap markers.
- Settings quality toggle for aura lights.
- forward_plus perf re-baseline.
