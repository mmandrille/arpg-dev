# v269 Plan - Authoritative Navigation Budget

Status: Complete
Goal: Bound server-owned monster pathfinding in crowded combat through data-driven budgets, path
reuse, deterministic repath throttling, and staggered monster repaths.
Architecture: Keep timing in realtime/perf logs only; `game/` owns deterministic budget counters and
transient monster path cache state. Apply the budget only to monster AI navigation. Player
auto-navigation and companions keep precise pathfinding so existing deterministic coverage remains
stable.
Tech stack: Go sim/pathfinding, shared navigation rules/schema, protocol bot stress probe, SDD docs.

## Baseline

v268 crowded probe produced a representative sampled row:

```text
tick=138 total_ms=37.731 sim_ms=36.134 ai_ms=36.066 pathfind_ms=34.937 path_requests=202 path_nodes_visited=5825 monsters_moved=4
```

v269 should keep the same server-authored reproduction but cap monster pathfinding work per tick.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/navigation.v0.json` | Add monster path budget/throttle settings |
| Modify | `shared/rules/navigation.v0.schema.json` | Validate new navigation budget fields |
| Modify | `server/internal/game/rules.go` | Load and validate navigation budget fields |
| Modify | `server/internal/game/pathfind.go` | Add optional node-limit support to deterministic stats |
| Modify | `server/internal/game/perf_debug.go` | Count cache hits and budgeted path work |
| Add | `server/internal/game/monster_navigation_budget.go` | Monster path cache, budget, throttle, and stagger helpers |
| Modify | `server/internal/game/sim.go` | Add transient budget fields and use cached monster chase goals |
| Modify | `server/internal/game/elite_minion_ai.go` | Use budgeted monster path helper for actual movement |
| Add/Modify | `server/internal/game/*navigation*_test.go` | Prove budget bounds and server-owned movement |
| Add | `docs/as-built/v269_authoritative-navigation-budget.md` | Record shipped proof |
| Modify | progress docs | Advance lifecycle/current status |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go` - kept within current grandfathered baseline allowance.
- [x] `server/internal/game/rules.go` - small schema/load additions only.
- [x] Other over-limit files from `.maintainability/file-size-baseline.tsv`: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.
- [ ] Defer extraction with rationale.

Verification:

```bash
make maintainability
```

## Task 1 - Data-driven navigation budgets

Files:
- Modify: `shared/rules/navigation.v0.json`
- Modify: `shared/rules/navigation.v0.schema.json`
- Modify: `server/internal/game/rules.go`

- [x] Add `monster_path_requests_per_tick`, `monster_path_nodes_per_tick`,
  `monster_path_cache_ticks`, `monster_repath_throttle_ticks`, and
  `monster_repath_stagger_ticks`.
- [x] Validate all fields as non-negative/positive where appropriate.
- [x] Keep values conservative enough for crowded combat and configurable for later tuning.

```bash
make validate-shared
```

## Task 2 - Budgeted monster path planning

Files:
- Modify: `server/internal/game/pathfind.go`
- Modify: `server/internal/game/perf_debug.go`
- Add: `server/internal/game/monster_navigation_budget.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/elite_minion_ai.go`

- [x] Add optional path node limit support to `PlanPathWithStats`.
- [x] Reset per-tick monster path budget before each tick.
- [x] Cache selected monster chase goals and path steps by target player, with deterministic expiry.
- [x] Throttle successful and failed monster repaths, staggered by monster id.
- [x] Enforce per-tick request and node budgets for monster AI pathfinding only.
- [x] Use cached paths when available; if no budget/cache exists, skip movement server-side rather
  than moving authority to clients.

```bash
cd server && go test ./internal/game
```

## Task 3 - Proof and docs

Files:
- Add/Modify: focused Go tests
- Modify: `docs/specs/v269_spec-authoritative-navigation-budget.md`
- Modify: `docs/plans/v269_2026-06-18-authoritative-navigation-budget.md`
- Add: `docs/as-built/v269_authoritative-navigation-budget.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/slice-codename-index.md`
- Modify: `PROGRESS.md`

- [x] Add a crowded-room test that advances the v268 probe world and asserts path requests/nodes
  stay within configured monster budgets while movement remains server-side.
- [x] Run the protocol bot probe and inspect backend perf output for bounded path counters.
- [x] Update docs and mark the spec/plan complete.

```bash
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game`
- [x] `cd server && go test ./internal/realtime`
- [x] `ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe`
- [x] `ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe`
- [x] `make maintainability`

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
