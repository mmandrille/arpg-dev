# v270 Spec - Crowd Movement LOD

Status: Complete
Date: 2026-06-18
Codename: crowd-movement-lod

## Purpose

Add a cheaper server-authoritative movement mode for low-priority monsters in crowded fights. v269
bounded pathfinding spikes; v270 reduces how often far, low-priority monsters ask for movement goals
or emit movement updates while preserving precise movement for nearby and important threats.

## Non-goals

- No client-authoritative monster navigation and no player-hosted navigation.
- No client-only visual smoothing as the source of truth.
- No production viewport/offscreen telemetry; the server approximates "offscreen" as far from all
  living players for this slice.
- No combat balance or monster-density tuning.
- No multiplayer overload guardrail policy; v271 owns room/session degradation warnings.

## Acceptance Criteria

- Movement LOD settings are data-driven under `shared/rules/navigation.v0.json`.
- Movement LOD activates only in crowded rooms above a configured live-monster threshold.
- Nearby monsters, bosses, elites, and pack leaders keep high-precision movement every sim tick.
- Low-priority far monsters deterministically skip movement/path-planning work on staggered ticks
  instead of pushing authority to clients.
- Focused backend tests prove low-priority monsters are LOD-deferred in the crowded probe while a
  nearby important monster remains precise.
- The crowded lightning protocol and visual bot scenarios still pass and backend perf remains
  bounded during repeated `ligthing` casts.

## Scope and Likely Files

- Server game:
  - `server/internal/game/monster_movement_lod.go`
  - `server/internal/game/monster_navigation_budget.go`
  - `server/internal/game/sim.go`
  - `server/internal/game/elite_minion_ai.go`
  - `server/internal/game/rules.go`
  - focused Go tests
- Shared rules:
  - `shared/rules/navigation.v0.json`
  - `shared/rules/navigation.v0.schema.json`
  - golden mirrors that embed navigation rules
- Docs:
  - `docs/as-built/v270_crowd-movement-lod.md`
  - progress docs

## Test and Bot Proof

```bash
make validate-shared
cd server && go test ./internal/game
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
make maintainability
```

All focused commands passed on 2026-06-18. The selected `$autoloop` batch still owns the final
`make ci` gate after v271.

## Open Questions and Risks

- Without client viewport telemetry, "offscreen" must be approximated by distance from all players.
  This is deterministic and multiplayer-ready, but not a final presentation-aware priority model.
- If low-priority monsters feel too slow visually, tune the data fields rather than bypassing server
  authority.
