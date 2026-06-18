# v269 Spec - Authoritative Navigation Budget

Status: Complete
Date: 2026-06-18
Codename: authoritative-navigation-budget

## Purpose

Keep monster navigation server-authoritative while preventing crowded rooms from spending hundreds
of path searches in one simulation tick. Build directly on v268, where the crowded lightning probe
showed navigation/pathfinding dominating sampled sim ticks.

## Non-goals

- No client-authoritative monster navigation and no player-hosted navigation.
- No client-only visual smoothing or hiding of backend stalls.
- No broad movement LOD policy for far/offscreen/blocked monsters; v270 owns that.
- No production metrics or dashboards.
- No combat balance or monster density tuning.

## Acceptance Criteria

- Monster pathfinding has data-driven per-tick request and node budgets under
  `shared/rules/navigation.v0.json`.
- Monster repaths are deterministic, throttled, and staggered by server-owned state so a crowd does
  not all repath on the same tick.
- Monster movement can reuse cached server paths/goals when the target and cached route remain valid.
- Player auto-navigation and companion navigation keep the existing precise path behavior for this
  slice.
- When the crowded lightning probe runs, sampled backend perf rows show bounded path requests/nodes
  instead of unbounded candidate scans such as the v268 sample (`path_requests=202`,
  `path_nodes_visited=5825`).
- Focused tests prove crowded monster movement remains server-owned and path counters stay within
  the configured budget.

## Scope and Likely Files

- Server game:
  - `server/internal/game/pathfind.go`
  - `server/internal/game/perf_debug.go`
  - `server/internal/game/monster_navigation_budget.go`
  - `server/internal/game/sim.go`
  - `server/internal/game/elite_minion_ai.go`
  - `server/internal/game/rules.go`
  - focused Go tests
- Shared rules:
  - `shared/rules/navigation.v0.json`
  - `shared/rules/navigation.v0.schema.json`
- Bot/docs:
  - `tools/bot/scenarios/93_crowded_lightning_perf_probe.json` only if timing needs adjustment
  - `docs/as-built/v269_authoritative-navigation-budget.md`
  - progress docs

## Test and Bot Proof

```bash
make validate-shared
cd server && go test ./internal/game
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
make maintainability
```

Visual regression command:

```bash
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
```

All focused commands passed on 2026-06-18. The selected `$autoloop` batch still owns the final
`make ci` gate after v270 and v271.

## Open Questions and Risks

- The budget should be strict enough to cap crowded spikes but loose enough that nearby monsters
  still feel responsive. v269 uses conservative server-side reuse/throttling; v270 can add
  lower-priority steering degradation if the crowded visual still feels too stiff.
- Path caches are transient sim state. They must be deterministic from tick order and not persisted
  as durable gameplay state.
