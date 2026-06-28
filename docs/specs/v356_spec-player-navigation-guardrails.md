# v356 Spec: Player Navigation Guardrails

Status: Implemented  
Date: 2026-06-27  
Codename: `player-navigation-guardrails`  
Baseline: v355 `aura-soft-lights`

## Purpose

Prevent combat sessions from stalling when the server spends hundreds of milliseconds on
**unbounded player pathfinding** and the client keeps retrying unreachable attack/move commands.

Observed failure (manual `make play`, paladin melee combat):

- `session_tick_budget_overrun` with `path_nodes_visited: 30291` on a single tick (~570 ms sim time).
- Repeated client `rejected: no_path` while hold-click / sticky attack kept issuing intents.
- Secondary Godot spam: `look_at() failed` in `entity_tick_smoothing_runtime.gd` for near-stationary
  projectiles.

This slice extends v269's monster navigation budget to **player-facing** path searches, stops the
client from hammering the server after `no_path`, and fixes the projectile facing edge case.

## Non-goals

- No change to monster movement LOD or overload degradation policy (v269/v270/v271).
- Monster per-search cap and lower shared tick budget align monsters to the same 3 s planning horizon.
- No protocol or wire-format changes; existing reject reasons (`no_path`, `path_too_long`) stay.
- No new server events or client animation states (ADR-0007).
- No combat reach tuning, damage, cooldown, or paladin-skill behavior changes.
- No NavMesh or client-authoritative pathfinding.
- No broad refactor of `findApproachGoal` ring ordering or approach heuristics beyond budget early-exit.
- No perf overlay UI changes unless required to surface new counters in existing debug payloads.

## Background

v269 intentionally capped **monster** path nodes/requests per tick but left player auto-navigation
with unlimited A*. `planPath` calls `PlanPathWithStats` with an empty `PathSearchStats` (no
`NodeLimit`), so a failed `move_to_intent` or `action_intent` approach can explore the entire
reachable grid (~30k directional states on a dungeon floor).

On the client, v296 sticky attack remembers a clicked monster but `_handle_intent_rejected` only
clears `_sustained_click` on `no_path` — not `_sticky_attack`. Hold-click can also resend approach
`move_to_intent` every `SEND_INTERVAL` until rejection arrives, amplifying server load.

## Acceptance criteria

### Server — data-driven path planning horizon (3 seconds)

Navigation budgets are tied to **how far ahead** entities plan, not to “one node = one step of
walking.” A* **nodes visited** are search expansions (`pathState` = grid cell + incoming direction,
up to ~8× cell count on hard failures). Returned **path steps** are what auto-nav executes.

At **10 Hz** (`tickHz`):

| Movement mode | Per-tick travel | 3 s travel |
|---------------|-----------------|------------|
| Player auto-nav (grid path) | **1.0** world unit / tick (`cell_size`) | **~30 steps** |
| Player WASD (paladin) | **~0.65** units / tick | **~19.5** units |
| Typical dungeon mob | **~0.39** units / tick | **~12** units |

Today `max_auto_steps` is **100** in the base rules file (~10 s of auto-nav), and **dungeon floors
inflate it to `width + height`** (often 60–80+ steps = 6–8 s per plan). That is why a single failed
search can explore tens of thousands of nodes on a full floor grid.

- [x] `shared/rules/navigation.v0.json` (+ schema) adds:
  - `path_planning_horizon_seconds` — default **3.0**; derived `horizon_ticks = floor(seconds × 10)`.
  - `player_max_auto_steps` — cap returned player routes; default **30** (= 3 s auto-nav).
  - `player_path_nodes_per_search` — max A* expansions per player query; default **480**
    (`horizon_ticks × 16` search factor for 8-way + turns).
  - `player_path_nodes_per_tick` — max aggregate player-path nodes per sim tick; default **8192**
    (distance-scaled per query; raised from 600 after maze routes needed ~3k–8k expansions).
  - `monster_path_nodes_per_search` — per-monster search cap; default **360**
    (`horizon_ticks × 12`; monsters move slower than auto-nav).
  - `monster_path_nodes_per_tick` — lower from **1200 → 600** (shared crowd budget for ~3 s horizon).
- [x] Player handlers (`move_to`, `action` approach, skill approach) use `player_max_auto_steps`,
  not dungeon-inflated `max_auto_steps`, except **interactable** `action_intent` may accept routes up
  to floor `max_auto_steps` when the offset move goal is unreachable but the entity center is.
- [x] Player path searches fail fast when either limit is hit and surface existing reject reasons
  (`no_path` when no route found or budget exhausted mid-search).
- [x] `path_nodes_visited` perf counters include player path work and stay ≤ configured tick cap in
  focused tests.
- [x] Deterministic: budgets depend only on tick-local counters and navigation rules, not wall clock.

### Server — approach scans respect tick budget

- [x] `findApproachGoalMatching`, `findRangedApproachGoal`, and `findSkillCastApproachGoal` stop
  scanning additional ring candidates once the per-tick player path node budget is exhausted for
  the current tick (return `no_path` rather than continuing expensive probes). Closed-barrier
  interactables (doors) skip the entity-center fast path so ring scans still prefer the player side.

### Client — stop retry storms

- [x] `_handle_intent_rejected` clears `_sticky_attack` (and existing sustained-click clear) on
  `no_path` and `path_too_long`.
- [x] After a path reject, the client enters a short **path-reject backoff** window (client-owned,
  default ~0.75 s) during which it does not resend `move_to_intent` or monster `action_intent` for
  the same target id (or same floor goal for move-only rejects).
- [x] Floor clicks and explicit retargets to a different monster/floor goal bypass or reset backoff
  as appropriate so normal navigation still feels responsive.

### Client — projectile facing fix

- [x] `entity_tick_smoothing_runtime._face_projectile_motion` no longer errors when XZ displacement
  is below the epsilon threshold after the node position was advanced (use direction from `to - prev`
  or `look_at_from_position(from, look_target)` consistently).
- [x] Headless unit test covers the zero/near-zero motion case.

### Regression / proof

- [x] Existing navigation goldens and monster budget tests remain green.
- [x] New Go tests prove player path node caps and approach early-exit.
- [x] Client unit tests prove sticky clear + backoff guards.
- [x] New **extended** protocol bot scenario proves bounded `path_nodes_visited` when approach to a
  walled target is impossible.
- [x] `make ci` green.

## Scope and likely files

| Area | Files |
|------|-------|
| Shared rules | `shared/rules/navigation.v0.json`, `navigation.v0.schema.json`; golden copies if referenced |
| Server game | `server/internal/game/navigation_rules.go`, `perf_debug.go` or new `player_navigation_budget.go`, `approach.go`, `sim.go` (`findRangedApproachGoal`, `findSkillCastApproachGoal`), `handlers.go` (if wiring only), focused `*_test.go` |
| Client | `client/scripts/main.gd`, new `client/scripts/path_reject_backoff.gd` (or extend `combat_sticky_target.gd`), `entity_tick_smoothing_runtime.gd` |
| Client tests | `client/tests/test_sustained_input.gd` and/or new `test_path_reject_backoff.gd`, `test_projectile_tick_smoothing.gd` or `test_entity_tick_smoothing.gd` |
| Bot | `tools/bot/scenarios/NN_player_path_budget_lab.json` (protocol, extended), optional client scenario update for sticky clear |
| Docs | `docs/as-built/v356_player-navigation-guardrails.md`, lifecycle on `/finish` |

## Client asset / plugin decision

| Option | Decision |
|--------|----------|
| In-repo GDScript helpers + `ClientConstants` backoff | **Adopt** |
| Shared rules for client backoff duration | **Reject** — presentation/input feel stays client-owned |
| External plugins | **Reject** |

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'PlayerPath|ApproachGoal|NavigationBudget' -count=1
make client-unit   # or focused godot --headless scripts
make bot scenario=player_path_budget_lab
make maintainability
make ci
```

Manual repro guard (post-fix should not spike):

```bash
make play
# Click/hold attack on monster separated by walls; server logs should not show 10k+ path_nodes_visited.
```

## ADR alignment

- **ADR-0001 D2:** Server remains authoritative for movement outcomes; budgets only bound work,
  not client prediction visuals.
- **ADR-0007:** Projectile facing is client-only presentation.
- **ADR-0014 D7:** Backoff improves input forgiveness without changing combat outcomes.

## Open questions and risks

| Item | Resolution |
|------|------------|
| Planning horizon | **3.0 s** → **30 ticks** at 10 Hz; encoded as `path_planning_horizon_seconds`. |
| Player route length | **`player_max_auto_steps: 30`** (~30 m auto-nav); replan via existing `finishAutoNav`. |
| Player search cap | **480 nodes/search** — enough to find a 30-step route through modest walls, stops ~30k floor scans. |
| Player tick cap | **8192 nodes/tick** (distance-scaled; 600 was too tight for long maze routes). |
| Monster search cap | **360 nodes/search** (new); **`monster_path_nodes_per_tick: 600`**. |
| Why not tie nodes = steps? | A* counts directional expansions, not path length; use horizon ticks for steps, `×12–16` for search. |
| Client backoff duration | **0.75 s** in `ClientConstants`; not shared rules. |
| `path_too_long` vs budget exhaust | Both use `no_path` reject today for failed approach; budget exhaust maps to `no_path` (unchanged wire). |
| Risk: legitimate long paths rejected | Mitigated by per-search limit still allowing multi-step routes within cap; monitor via bot lab + playtest. |
