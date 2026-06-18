# v271 Spec - Multiplayer Room Guardrails

Status: Complete
Date: 2026-06-18
Codename: multiplayer-room-guardrails

## Purpose

Add backend guardrails for future co-op/multiplayer room simulation under crowded combat load. v269
bounded monster pathfinding and v270 added movement LOD; v271 adds per-session tick budget warnings,
an overload degradation window, and documentation of the authority model.

## Non-goals

- No client-authoritative monster navigation, combat, loot, or AI.
- No client-hosted room simulation.
- No matchmaking, load balancing, split-process ownership, or production autoscaling.
- No visual-only freeze masking.
- No full netcode prediction protocol.

## Acceptance Criteria

- Realtime session ticks evaluate the authoritative backend tick budget for every session, not only
  when perf debug sampling is enabled.
- Over-budget ticks emit a per-session warning with tick, total time, budget, overrun, live monster
  count, path counters, client count, and whether degradation was applied.
- Overload degradation is server-owned and data-driven. It temporarily defers only low-priority
  monster movement/path work; nearby/boss/elite/pack-leader monsters keep precision.
- Degradation only applies to over-budget ticks with path or monster movement pressure; startup or
  room population spikes still warn but do not force movement LOD.
- Focused tests prove overload degradation remains server-side and does not imply client authority.
- The crowded lightning bot and visual replay still pass.
- Docs state the model: clients may predict/interpolate presentation, but the server owns AI,
  navigation, combat, loot, persistence, and authoritative room state.

## Scope and Likely Files

- Server realtime:
  - `server/internal/realtime/session_tick.go`
  - `server/internal/realtime/tick_guardrails.go`
  - focused realtime tests
- Server game:
  - `server/internal/game/monster_movement_lod.go`
  - `server/internal/game/sim.go`
  - focused game tests
- Shared rules:
  - `shared/rules/navigation.v0.json`
  - `shared/rules/navigation.v0.schema.json`
  - navigation golden mirrors
- Docs:
  - `docs/as-built/v271_multiplayer-room-guardrails.md`
  - progress docs

## Test and Bot Proof

```bash
make validate-shared
cd server && go test -count=1 ./internal/game ./internal/realtime
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
make maintainability
```

All focused commands passed on 2026-06-18. The selected `$autoloop` batch still owns the final
`make ci` gate.

## Open Questions and Risks

- The first degradation policy is intentionally simple: it protects the authoritative room by
  deferring low-priority monsters, not by dropping combat, loot, or persistence.
- Production load routing and cross-process room ownership remain future architecture work.
