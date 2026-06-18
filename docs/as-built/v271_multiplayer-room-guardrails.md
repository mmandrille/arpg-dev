# v271 As-Built - Multiplayer Room Guardrails

Date: 2026-06-18
Spec: [`docs/specs/v271_spec-multiplayer-room-guardrails.md`](../specs/v271_spec-multiplayer-room-guardrails.md)
Plan: [`docs/plans/v271_2026-06-18-multiplayer-room-guardrails.md`](../plans/v271_2026-06-18-multiplayer-room-guardrails.md)

## Shipped Behavior

- Realtime now evaluates the backend tick budget for every session tick, independent of perf-debug
  sampling.
- Over-budget ticks emit `session_tick_budget_overrun` warnings with bounded per-session fields:
  tick timing, budget/overrun, result shape, client count, live monsters, wall count, path counters,
  moved monsters, and whether degradation was applied.
- Warnings are emitted for every over-budget session tick. Degradation is additionally gated on
  path or monster movement pressure so startup/population spikes do not throttle unrelated rooms.
- `monster_overload_degrade_ticks` is data-driven in `shared/rules/navigation.v0.json`.
- The sim keeps a transient overload-degradation deadline. During that window, low-priority monsters
  skip movement/path work; nearby monsters, bosses, elites, and pack leaders remain high precision.

## Authority Model

- The server owns room truth: AI, navigation, movement eligibility, combat, loot, persistence, and
  authoritative state deltas.
- Clients may predict local presentation, smooth/interpolate remote entities, and replay VFX, but
  client predictions are never accepted as monster navigation, combat, loot, or persistence truth.
- Overload degradation is a server decision. Clients receive fewer or older monster movement deltas
  for low-priority monsters during degradation; they do not decide substitute monster positions.

## Perf Sample

The final crowded lightning protocol bot stayed under the 100 ms backend tick budget and emitted no
`session_tick_budget_overrun` warnings. Representative sampled rows:

```text
tick=0 total_ms=29.832 sim_ms=5.833 ai_ms=5.571 pathfind_ms=4.994 path_requests=40 path_cache_hits=3 path_nodes_visited=480 monsters_moved=3 tick_over_budget=false
tick=22 total_ms=13.582 sim_ms=11.214 ai_ms=11.151 pathfind_ms=10.719 path_requests=11 path_cache_hits=4 path_nodes_visited=1201 monsters_moved=1 tick_over_budget=false
tick=86 total_ms=3.135 sim_ms=0.726 ai_ms=0.655 pathfind_ms=0.000 path_requests=0 path_cache_hits=8 path_nodes_visited=0 monsters_moved=1 tick_over_budget=false
tick=138 total_ms=2.610 sim_ms=0.814 ai_ms=0.754 pathfind_ms=0.000 path_requests=0 path_cache_hits=8 path_nodes_visited=0 monsters_moved=1 tick_over_budget=false
```

Focused realtime tests cover the warning payload, budget decision path, and the path/movement
pressure gate even though the bot's steady-state run did not need to degrade.

## Boundaries

- No client-authoritative or player-hosted room simulation was added.
- No matchmaking, load balancing, split deployables, or production autoscaling was added.
- No startup/population-only overrun applies movement degradation unless path or monster movement
  counters show room pressure.
- No combat, loot, persistence, or high-priority monster work is dropped by the degradation policy.

## Verification

```bash
make validate-shared
cd server && go test -count=1 ./internal/game ./internal/realtime
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
make maintainability
HEADLESS=1 make bot-visual scenario=11_combat_feedback
HEADLESS=1 make bot-visual scenario=69_discovery_minimap_toggle
make ci
```

All focused commands passed on 2026-06-18. Final full `make ci` passed on 2026-06-18 in 10m47s,
including the crowded lightning protocol probe, replay, 74 Godot client bot scenarios, and headless
smoke.
