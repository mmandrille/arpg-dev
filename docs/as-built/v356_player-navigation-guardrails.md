# v356 As-built — Player Navigation Guardrails

## What shipped

Bounded **player-facing** pathfinding to a **3-second planning horizon** and stopped client retry storms after `no_path`.

### Server (`shared/rules/navigation.v0.json`)

| Field | Value | Meaning |
|-------|-------|---------|
| `path_planning_horizon_seconds` | 3.0 | Documentation anchor for horizon |
| `player_max_auto_steps` | 30 | ~3 s auto-nav per plan (10 Hz × 1 cell/tick) for combat/move |
| `player_path_nodes_per_search` | 480 | Base A* expansions per player query |
| `player_path_nodes_per_tick` | 8192 | Aggregate player search work per tick (distance-scaled per query) |
| `monster_path_nodes_per_search` | 360 | Per-monster search cap |
| `monster_path_nodes_per_tick` | 600 | Down from 1200 |

Distance scaling: per-query node limit is `max(480, octile(start, goal)*128 + 64)` capped by remaining tick budget.

- `player_navigation_budget.go` — `planPlayerPath`, `planPlayerPathForApproach` (failed ring probes charge ≤64 nodes), `planPlayerPathForInteractable` (truncates to dungeon `max_auto_steps`).
- `approach.go` extracted from `sim.go`; interactable approach tries entity-center path first (client-aligned), except **closed-barrier doors** which use ring scan for player-side preference; ring scan for combat.
- `pathfind.go` — count goal on same expansion as limit check; `planPathWithNodeLimit` retries alt-starts after main-search limit.
- `handlers.go` — interactable `action_intent` may use floor `max_auto_steps` for `path_too_long` guard.
- Dungeon generation / `planPath` (unlimited) unchanged for reachability proofs.

### Client

- `path_reject_backoff.gd` — 0.75 s backoff per target/goal after `no_path` / `path_too_long`.
- `main.gd` — clears sticky attack on path reject; gates resends; floor click clears backoff.
- `entity_tick_smoothing_runtime.gd` — projectile facing uses `look_at_from_position`.

### Proof

- Go: `player_navigation_budget_test.go`, `TestV40ObstaclesWoodenDoorActionAcks`, `TestGeneratedWallLabStairsOffsetMoveGoalV40`
- Client: `test_path_reject_backoff.gd`, sticky-clear in `test_coop_client.gd`
- Bot: extended `94_player_path_budget_lab.json`; `reachable_dungeon_obstacles` seed → `wall_seed_00` (v40 cross-map walk died to aggro under 30-step replans)

## Manual check

```bash
make play
# Hold-attack a walled monster: path_nodes_visited should stay in hundreds/low thousands, not 10k+.

make bot-visual scenario=player_path_budget_lab
make bot-client SCENARIO=dungeon_wall_rendering HEADLESS=1
```
