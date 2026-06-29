# v366 Spec: Path Cache Effectiveness

Status: Draft  
Date: 2026-06-28  
Codename: `path-cache-effectiveness`  
Baseline: v365 `path-budget-retune`

## Purpose

Improve monster path cache reuse under crowded combat via longer cache windows, goal tolerance, and
aligned repath throttling — reducing path requests without changing authority.

## Non-goals

- No protocol performance schema changes.
- No client changes.

## Acceptance criteria

- `monster_path_cache_ticks` and `monster_repath_throttle_ticks` retuned.
- New `monster_path_cache_goal_tolerance` (world units) allows reuse when chase goals drift slightly.
- Focused test proves crowded probe achieves cache hits >= path requests over a warmup window.

## Test proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'CrowdedLightning|PathCache' -count=1
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
```
