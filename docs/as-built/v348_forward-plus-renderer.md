# v348 As-Built — Forward Plus Renderer

Date: 2026-06-26  
Spec: [`docs/specs/v348_spec-forward-plus-renderer.md`](../specs/v348_spec-forward-plus-renderer.md)  
Plan: [`docs/plans/v348_2026-06-26-forward-plus-renderer.md`](../plans/v348_2026-06-26-forward-plus-renderer.md)

## Shipped behavior

- **Interactive default:** `client/project.godot` sets `renderer/rendering_method="forward_plus"` for
  `make play`, windowed `make bot-visual`, and normal Godot editor runs.
- **Headless CI override:** `scripts/godot_ci_flags.sh` exports
  `GODOT_HEADLESS_FLAGS="--headless --rendering-method gl_compatibility"`. Sourced by
  `client_smoke.sh`, `bot_client.sh`, and headless `bot_visual.sh` so automated gates do not depend
  on forward_plus GPU features.
- **No gameplay/protocol changes.** Fog overlay, materials, and lighting paths unchanged.

## Renderer evaluation (implementation host, 2026-06-26)

| Path | Renderer | Notes |
|------|----------|-------|
| `make play` / windowed bot | `forward_plus` (project default) | Modern 3D pipeline for interactive play |
| `make client-unit`, headless bot-client | `gl_compatibility` (CLI override) | All unit gates green after migration |
| Fog bot regressions 67/68/73 | `gl_compatibility` (headless) | PASS |

Extended perf probe (`dungeon_combat_perf_probe`) not re-run in this slice commit; v347 samples
remain the baseline until a manual `ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe`
run on forward_plus.

## Boundaries

- No server/protocol/shared/golden changes.
- No Settings renderer toggle (single default + headless override).
- Movement tick interpolation deferred to v349.

## Verification

```bash
make maintainability
make client-unit
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
HEADLESS=1 make bot-visual scenario=73_door_fog_toggle
```
