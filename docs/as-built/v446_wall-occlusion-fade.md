# v446 As-Built — Wall Occlusion Fade

Spec: [`docs/specs/v446_spec-wall-occlusion-fade.md`](../specs/v446_spec-wall-occlusion-fade.md)  
Plan: [`docs/plans/v446_2026-07-06-wall-occlusion-fade.md`](../plans/v446_2026-07-06-wall-occlusion-fade.md)

## What it proved

- Client-only box wall meshes (`wall` / `wood`) fade when they lie on the camera→entity segment on the XZ plane, restoring hero and rendered monster readability without changing server fog/LOS authority.
- `WallRenderer` owns occlusion mesh registry + fade sync; `main.gd` piggybacks on the existing monster health-bar refresh tick.
- Extended client bot `wall_occlusion_fade` on `line_of_sight_blocker_lab` asserts `faded_wall_count >= 1` and sub-opaque `min_faded_alpha`.

## Key files

- `shared/assets/wall_occlusion_presentation.v0.json`
- `client/scripts/wall_occlusion_fade.gd`, `wall_occlusion_runtime.gd`, `wall_renderer.gd`
- `tools/bot/scenarios/client/98_wall_occlusion_fade.json`

## Deferred

- User setting to disable/tune fade intensity
- Closed doors, columns, rocks, and other non-box obstacle meshes
- Fog-hidden monster reveal (intentionally out of scope)

## Verification

```bash
make client-unit
make maintainability
HEADLESS=1 make bot-visual scenario=wall_occlusion_fade
```

Manual: `make play` — walk along corridor walls in isometric view.
