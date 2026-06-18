# v270 As-Built - Crowd Movement LOD

Date: 2026-06-18
Spec: [`docs/specs/v270_spec-crowd-movement-lod.md`](../specs/v270_spec-crowd-movement-lod.md)
Plan: [`docs/plans/v270_2026-06-18-crowd-movement-lod.md`](../plans/v270_2026-06-18-crowd-movement-lod.md)

## Shipped Behavior

- Movement LOD is now data-driven in `shared/rules/navigation.v0.json` through:
  `monster_movement_lod_min_live_monsters`, `monster_movement_lod_near_distance`, and
  `monster_movement_lod_update_interval_ticks`.
- LOD activates only in crowded active levels. The default threshold is 24 live monsters, so ordinary
  small fights keep full precision.
- Bosses, elites, pack leaders, and monsters near any living player remain high precision every tick.
- Low-priority far monsters deterministically skip cached-goal reuse, movement, and path-planning
  work on non-LOD ticks. This stays server-authoritative; clients only render/interpolate state they
  receive.

## Perf Sample

The final protocol bot run with 36 live monsters and repeated `ligthing` casts stayed under budget
and showed low-work LOD samples:

```text
tick=0 total_ms=24.374 sim_ms=5.664 ai_ms=5.427 pathfind_ms=4.750 path_requests=40 path_cache_hits=3 path_nodes_visited=480 monsters_moved=3 tick_over_budget=false
tick=21 total_ms=12.950 sim_ms=10.733 ai_ms=10.632 pathfind_ms=10.245 path_requests=10 path_cache_hits=4 path_nodes_visited=1201 monsters_moved=1 tick_over_budget=false
tick=105 total_ms=8.601 sim_ms=0.842 ai_ms=0.661 pathfind_ms=0.000 path_requests=0 path_cache_hits=8 path_nodes_visited=0 monsters_moved=2 tick_over_budget=false
tick=138 total_ms=6.355 sim_ms=1.217 ai_ms=1.045 pathfind_ms=0.000 path_requests=0 path_cache_hits=8 path_nodes_visited=0 monsters_moved=1 tick_over_budget=false
```

v269's path request/node caps still apply; v270 adds deterministic low-priority movement deferral
around those caps.

## Boundaries

- No client-authoritative or player-hosted monster navigation was added.
- No viewport/offscreen protocol was added; the server approximates offscreen priority as distance
  from all living players.
- No v271 overload guardrail policy shipped here.

## Verification

```bash
make validate-shared
cd server && go test -count=1 ./internal/game
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
make maintainability
```

All focused commands passed on 2026-06-18. Final full `make ci` remains the enclosing autoloop batch gate.
