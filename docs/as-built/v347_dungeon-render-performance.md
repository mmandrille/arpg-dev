# v347 As-Built — Dungeon Render Performance

Date: 2026-06-26  
Spec: [`docs/specs/v347_spec-dungeon-render-performance.md`](../specs/v347_spec-dungeon-render-performance.md)  
Plan: [`docs/plans/v347_2026-06-26-dungeon-render-performance.md`](../plans/v347_2026-06-26-dungeon-render-performance.md)

## Shipped behavior

- **Fog LOS shadow cache** (`FogLosShadowCache`): shadow polygon geometry rebuilds only on layout,
  camera/radius, viewport, or hero-move invalidation. `FogOfWarOverlay` still syncs cheap
  `Polygon2D` nodes each frame from cached payloads.
- **Shared tuning** (`shadow_cache` in `fog_presentation.v0.json`): `move_epsilon`,
  `viewport_size_epsilon_px`, `performance_min_rebuild_interval_frames` (default 3).
- **Graphics quality preset** in Settings: **Balanced** (user window size) vs **Performance**
  (effective 1920×1080 + fog shadow rebuild throttle). Persisted as `graphics_quality` in
  `user://settings.json`.
- **Benchmark scenario** `dungeon_combat_perf_probe` (`ci_tier: extended`): D1 descent, packed
  mobs, repeated `ligthing` casts under `ARPG_PERF_DEBUG=1`.

## Post-v347 samples (implementation host, 2026-06-26, `ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe`)

Combat phase (`live_monsters=24`, `tick≥300`):

| Metric | Post-v347 (sample) |
|--------|-------------------|
| FPS median | ~73–77 |
| FPS floor | ~57 |
| `process_ms` (combat) | ~14–16 ms typical (down from ~19 ms pre-v347) |

Shadow cache debug keys (`shadow_cache_hits`, `shadow_rebuild_count`) available via `fog_of_war` bot state.

## Pre-v347 baseline (implementation host, 2026-06-26)

Rendered D1 combat (`live_monsters ≥ 10`, combat phase):

| Metric | Pre-v347 |
|--------|----------|
| FPS median | ~53 |
| FPS floor | ~36 |
| Headless CPU median | ~71 |

## Post-v347 notes

- Shadow rebuilds per second drop materially on cache hits (`shadow_rebuild_count` / `shadow_cache_hits`
  exposed in `fog_of_war_overlay` debug state).
- Full rendered FPS sign-off (median ≥ 60, floor ≥ 45) is host/GPU dependent; re-run:
  `ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe` or
  `ARPG_PERF_DEBUG=1 make play` with Settings → Performance vs Balanced in D1 combat.

## Boundaries

- No server/protocol/golden changes.
- `forward_plus` renderer migration deferred to v348.
- Client movement smoothing between 10 Hz ticks deferred to v349.
- `dungeon_combat_perf_probe` stays **extended** (not CI pack).

## Verification

```bash
make validate-shared
make client-unit
make maintainability
make bot scenario=dungeon_combat_perf_probe
HEADLESS=1 make bot-visual scenario=73_door_fog_toggle
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe
```

Focused verification green on 2026-06-26. Full `make ci` is the merge gate on `/finish`.
