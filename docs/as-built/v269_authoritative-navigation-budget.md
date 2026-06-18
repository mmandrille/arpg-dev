# v269 As-Built - Authoritative Navigation Budget

Date: 2026-06-18
Spec: [`docs/specs/v269_spec-authoritative-navigation-budget.md`](../specs/v269_spec-authoritative-navigation-budget.md)
Plan: [`docs/plans/v269_2026-06-18-authoritative-navigation-budget.md`](../plans/v269_2026-06-18-authoritative-navigation-budget.md)

## Shipped Behavior

- Monster pathfinding now has data-driven request, node, cache, throttle, and stagger settings in
  `shared/rules/navigation.v0.json`.
- Monster navigation remains server-authoritative. When the per-tick budget is exhausted, the
  backend skips that monster movement for the tick instead of delegating navigation to clients.
- Monster chase/follow movement can reuse transient cached server paths and goals while the target,
  cache age, and destination remain valid.
- Repaths are throttled and deterministically staggered by monster id so crowded rooms do not all
  repath in the same tick.
- Player auto-navigation and companion navigation keep their existing precise path behavior.

## Perf Sample

The v268 baseline sampled:

```text
tick=138 total_ms=37.731 sim_ms=36.134 ai_ms=36.066 pathfind_ms=34.937 path_requests=202 path_nodes_visited=5825 monsters_moved=4
```

After v269, the same crowded lightning protocol probe produced bounded sampled rows. The observed
request peak hit the configured cap and node-heavy rows stopped at the node limiter boundary:

```text
tick=0 total_ms=27.588 sim_ms=5.658 ai_ms=5.398 pathfind_ms=4.726 path_requests=40 path_cache_hits=3 path_nodes_visited=480 monsters_moved=3 tick_over_budget=false
tick=21 total_ms=16.644 sim_ms=14.201 ai_ms=14.119 pathfind_ms=13.899 path_requests=11 path_cache_hits=4 path_nodes_visited=1201 monsters_moved=1 tick_over_budget=false
tick=126 total_ms=19.202 sim_ms=15.048 ai_ms=14.874 pathfind_ms=14.269 path_requests=13 path_cache_hits=7 path_nodes_visited=1201 monsters_moved=2 tick_over_budget=false
```

`1201` is the expected one-node overshoot from stopping an in-flight deterministic search after the
configured `monster_path_nodes_per_tick=1200` budget is crossed.

## Boundaries

- No client-hosted or client-authoritative monster navigation was added.
- No movement LOD/degraded steering policy shipped here; v270 owns lower-priority movement modes.
- No production metrics/dashboard work shipped; v268/v269 rely on local perf debug logs.
- Path cache state is transient sim state and is not persisted as durable gameplay state.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game
cd server && go test ./internal/realtime
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
make maintainability
```

All focused commands passed on 2026-06-18. Final full `make ci` remains the enclosing autoloop batch gate.
