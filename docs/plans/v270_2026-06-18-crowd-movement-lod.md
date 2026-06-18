# v270 Plan - Crowd Movement LOD

Status: Complete
Goal: Reduce crowded-room movement/path-planning work for far, low-priority monsters while keeping
nearby and important monsters precise and server-authoritative.
Architecture: Extend `NavigationRules` with deterministic movement LOD settings. The server owns
priority classification and decides whether a monster may do movement/path-planning work on the
current tick. Clients remain presentation-only consumers.
Tech stack: Go sim movement, shared navigation rules/schema, protocol and visual bot probe.

## Baseline

v269 bounded monster pathfinding to `monster_path_requests_per_tick=40` and
`monster_path_nodes_per_tick=1200`, but far monsters can still participate in movement work every
tick once they have cached goals. v270 should reduce low-priority movement work before v271 adds
overload guardrails.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/navigation.v0.json` | Add movement LOD tuning |
| Modify | `shared/rules/navigation.v0.schema.json` | Validate movement LOD fields |
| Modify | shared golden navigation mirrors | Keep golden fixtures schema-valid |
| Modify | `server/internal/game/rules.go` / helper | Load and validate movement LOD fields |
| Add | `server/internal/game/monster_movement_lod.go` | Server-owned priority and LOD tick helpers |
| Modify | `server/internal/game/sim.go` | Defer low-priority chase goal work on LOD skip ticks |
| Modify | `server/internal/game/elite_minion_ai.go` | Defer direct movement for low-priority monsters |
| Add/Modify | focused Go tests | Prove LOD deferral and high-priority precision |
| Add | `docs/as-built/v270_crowd-movement-lod.md` | Record shipped proof |
| Modify | progress docs | Advance lifecycle/current status |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go` is at the current ratchet allowance. v270 avoided net-new lines
  there by gating LOD through the navigation helper path.
- [x] `server/internal/game/rules.go` kept the v269 navigation loader extraction.
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.
- [ ] Defer extraction with rationale.

Verification:

```bash
make maintainability
```

## Task 1 - Data-driven LOD settings

Files:
- Modify: `shared/rules/navigation.v0.json`
- Modify: `shared/rules/navigation.v0.schema.json`
- Modify: shared golden navigation mirrors
- Modify: `server/internal/game/navigation_rules.go`
- Modify: `server/internal/game/rules.go`

- [x] Add `monster_movement_lod_min_live_monsters`,
  `monster_movement_lod_near_distance`, and
  `monster_movement_lod_update_interval_ticks`.
- [x] Validate threshold and update interval as positive and near distance as non-negative.
- [x] Keep defaults conservative so LOD activates in the 36-monster probe but not ordinary small
  fights.

```bash
make validate-shared
```

## Task 2 - Server-owned movement LOD

Files:
- Add: `server/internal/game/monster_movement_lod.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/elite_minion_ai.go`
- Modify: `server/internal/game/monster_navigation_budget.go`

- [x] Count live monsters on the active server level.
- [x] Treat bosses, elites, pack leaders, and monsters within the configured near distance of any
  living player as high precision.
- [x] For low-priority far monsters in crowded rooms, deterministically skip movement and path goal
  work on non-LOD ticks using monster id/tick staggering.
- [x] Keep skipped movement authoritative: clients may interpolate stale state but never decide the
  monster's new position.

```bash
cd server && go test ./internal/game
```

## Task 3 - Proof and docs

Files:
- Add/Modify: focused Go tests
- Modify: `docs/specs/v270_spec-crowd-movement-lod.md`
- Modify: `docs/plans/v270_2026-06-18-crowd-movement-lod.md`
- Add: `docs/as-built/v270_crowd-movement-lod.md`
- Modify: progress docs

- [x] Add tests proving the crowded probe has deferred low-priority monsters and that an important
  monster remains precise.
- [x] Run protocol and visual crowded lightning probes.
- [x] Record perf sample and boundaries in as-built docs.

```bash
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test -count=1 ./internal/game`
- [x] `ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe`
- [x] `ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe`
- [x] `make maintainability`

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
