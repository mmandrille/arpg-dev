# v356 Plan — Player Navigation Guardrails

Status: **Implemented**  
Goal: Cap player pathfinding cost, stop client retry storms after `no_path`, and fix projectile `look_at` spam.  
Architecture: Add data-driven `player_path_nodes_per_search` and `player_path_nodes_per_tick` beside the
existing monster budgets. Route all player-facing `planPath` calls through a tick-scoped budget helper
that sets `PathSearchStats.NodeLimit` and tracks aggregate nodes. Approach ring scans early-exit when the
tick budget is spent. Client adds `PathRejectBackoff` to gate resends and clears sticky attack on path
rejects. Projectile facing uses explicit direction from segment motion.  
Tech stack: shared JSON rules, Go sim, Godot client, Python protocol bot, GDScript unit tests.

Spec: [`docs/specs/v356_spec-player-navigation-guardrails.md`](../specs/v356_spec-player-navigation-guardrails.md)  
Baseline: v355 `aura-soft-lights`

## Spec review (gate)

| Area | Result |
|------|--------|
| Baseline v356 / builds on v355 | OK |
| Scope / non-goals | OK — extends v269 player gap only |
| Contracts | Navigation rules JSON + schema; no protocol bump |
| Determinism | Tick-local counters only; no `time.Now()` in `game/` |
| Shared rules | New fields in `navigation.v0.json` |
| Server authority | Budgets bound work; outcomes unchanged |
| Bot proof | New extended protocol lab |
| Client | In-repo helpers; no external assets |
| Maintainability | Extract `player_navigation_budget.go`; thin `main.gd` edits |
| Open questions | Resolved in spec defaults table |

## Baseline and shortcut decision

Reuse patterns from:

- `server/internal/game/monster_navigation_budget.go` — per-tick request/node accounting
- `server/internal/game/perf_debug.go` — `planPath` / `runPathSearch` / `PathSearchStats`
- `docs/specs/v269_spec-authoritative-navigation-budget.md` — budget semantics
- `client/scripts/combat_sticky_target.gd` — sticky target lifecycle
- `client/tests/test_coop_client.gd` — existing `no_path` sustained-click assertions

| Option | Decision |
|--------|----------|
| New `player_navigation_budget.go` | **Adopt** |
| Backoff in `main.gd` only | **Reject** — extract `path_reject_backoff.gd` |
| Shared rules for backoff seconds | **Reject** |

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/navigation.v0.json` | Add `player_path_nodes_per_search`, `player_path_nodes_per_tick` |
| Modify | `shared/rules/navigation.v0.schema.json` | Schema for new fields |
| Modify | `shared/golden/auto_path.json`, `monster_chase.json` | Mirror new required nav fields if validated as fixtures |
| Modify | `server/internal/game/navigation_rules.go` | Load/validate new fields |
| Create | `server/internal/game/player_navigation_budget.go` | Tick reset, `planPlayerPath`, budget checks |
| Modify | `server/internal/game/perf_debug.go` | Delegate player `planPath` through budget helper |
| Modify | `server/internal/game/approach.go` | Early-exit ring scans when tick budget exhausted |
| Modify | `server/internal/game/sim.go` | `findRangedApproachGoal`, `findSkillCastApproachGoal` early-exit |
| Create | `server/internal/game/player_navigation_budget_test.go` | Cap + approach early-exit proofs |
| Modify | `server/internal/game/obstacle_variety_nav_test.go` or `game_test.go` | Adjust if nav fixture requires new fields |
| Create | `client/scripts/path_reject_backoff.gd` | Target/goal keyed backoff window |
| Modify | `client/scripts/client_constants.gd` | `PATH_REJECT_BACKOFF_S := 0.75` |
| Modify | `client/scripts/main.gd` | Wire backoff; clear sticky on path reject; gate `_start_attack_move` / sustained attack |
| Modify | `client/scripts/entity_tick_smoothing_runtime.gd` | Safe projectile facing |
| Create | `client/tests/test_path_reject_backoff.gd` | Backoff + sticky-clear behavior |
| Modify | `client/tests/test_sustained_input.gd` | Assert sticky cleared on `no_path` |
| Modify | `client/tests/test_projectile_tick_smoothing.gd` | Near-zero motion facing |
| Modify | `scripts/client_smoke.sh` | Register new test if added |
| Create | `tools/bot/scenarios/94_player_path_budget_lab.json` | Extended protocol proof |
| Modify | `shared/rules/worlds.v0.json` | Compact lab preset if new world needed |
| Modify | `tools/validate_shared.py` | Cross-check new nav fields if pattern exists |
| Create | `docs/as-built/v356_player-navigation-guardrails.md` | On `/finish` |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | On `/finish` |
| Modify | `docs/CODEMAP.md` | Index new modules |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [x] `client/scripts/main.gd` (baseline 6193) — wire-only; extract backoff helper
- [x] `server/internal/game/sim.go` — approach extracted to `approach.go`; baseline lowered
- [x] Other over-limit file: `handlers.go` baseline advanced to 1458

Decision:

- [x] Extract `player_navigation_budget.go` + `path_reject_backoff.gd`
- [x] Touched grandfathered files stay at or below baseline

Verification:

```bash
make maintainability
```

## Task 1 — Shared navigation rules

Files:

- Modify: `shared/rules/navigation.v0.json`, `navigation.v0.schema.json`
- Modify: golden nav copies if `make validate-shared` requires them

- [x] Step 1.1: Add horizon + budgets (all derived from 3 s @ 10 Hz):
  - `path_planning_horizon_seconds: 3.0`
  - `player_max_auto_steps: 30`
  - `player_path_nodes_per_search: 480`
  - `player_path_nodes_per_tick: 8192` (distance-scaled; see as-built)
  - `monster_path_nodes_per_search: 360`
  - `monster_path_nodes_per_tick: 600` (down from 1200)
- [x] Step 1.2: Update schema `required` + property defs; `additionalProperties: false` compliance.
- [x] Step 1.3: Sync any `shared/golden/*` navigation embeds referenced by validators.

```bash
make validate-shared
```

## Task 2 — Server player path budget

Files:

- Create: `server/internal/game/player_navigation_budget.go`
- Modify: `server/internal/game/navigation_rules.go`, `perf_debug.go`
- Create: `server/internal/game/player_navigation_budget_test.go`

- [x] Step 2.1: Add tick fields `playerPathNodesThisTick` with `resetPlayerNavigationBudget()` called
  from existing tick perf reset (same site as monster budget reset).
- [x] Step 2.2: Implement `planPlayerPath(nav, start, goal, blocked) ([]Vec2, bool)`:
  - refuse when `playerPathNodesThisTick >= nav.PlayerPathNodesPerTick`
  - set `stats.NodeLimit = min(remaining tick budget, per-search cap)`
  - increment counters via existing `runPathSearch`
- [x] Step 2.3: Change player intent `planPath` calls to `planPlayerPath`; reject routes longer than
  `player_max_auto_steps` with `path_too_long` (not dungeon-inflated `max_auto_steps`).
- [x] Step 2.4: Add `monster_path_nodes_per_search` cap in `planMonsterPath` (today only tick-remaining
  limit); lower shared `monster_path_nodes_per_tick` to 600.
- [x] Step 2.5: Tests:
  - unreachable goal returns `ok=false` with `NodesVisited <= player_path_nodes_per_search`
  - second search in same tick respects aggregate tick cap
  - deterministic across repeated runs (same seed)

```bash
cd server && go test ./internal/game/... -run PlayerNavigation -count=1
```

## Task 3 — Approach scan early-exit

Files:

- Modify: `server/internal/game/approach.go`, `sim.go` (`findRangedApproachGoal`, `findSkillCastApproachGoal`)
- Extend: `player_navigation_budget_test.go`

- [x] Step 3.1: Add `playerPathBudgetAvailable() bool` accessor.
- [x] Step 3.2: In each ring-candidate loop, break when budget exhausted (return `false` goal).
- [x] Step 3.3: Test: walled monster + melee approach returns false without exceeding tick node cap.
- [x] Step 3.4: Closed-barrier interactables skip entity-center fast path (door side preference).

```bash
cd server && go test ./internal/game/... -run 'ApproachGoal|PlayerNavigation' -count=1
```

## Task 4 — Client path-reject backoff + sticky clear

Files:

- Create: `client/scripts/path_reject_backoff.gd`
- Modify: `client/scripts/client_constants.gd`, `main.gd`
- Create: `client/tests/test_path_reject_backoff.gd`
- Modify: `client/tests/test_sustained_input.gd`

- [x] Step 4.1: `PathRejectBackoff` tracks `until_ms` per `target_id` and optional floor goal key.
- [x] Step 4.2: `_handle_intent_rejected`: on `no_path` / `path_too_long`, clear `_sticky_attack` and
  start backoff from `pending_action_targets` / last move goal when available.
- [x] Step 4.3: Gate `_start_attack_move`, `_repeat_hold_attack` move branch, and direct
  `_send_action_intent` for monsters when backoff active for that target.
- [x] Step 4.4: Floor click or different monster id clears/resets backoff entries.
- [x] Step 4.5: Unit tests for backoff window + sticky clear (mirror `test_coop_client` no_path case).

```bash
godot --headless --path client --script res://tests/test_path_reject_backoff.gd
godot --headless --path client --script res://tests/test_sustained_input.gd
```

## Task 5 — Projectile facing fix

Files:

- Modify: `client/scripts/entity_tick_smoothing_runtime.gd`
- Modify: `client/tests/test_projectile_tick_smoothing.gd`

- [x] Step 5.1: In `_face_projectile_motion`, when `flat.length_squared() <= epsilon`, return early
  **before** any `look_at` call; when non-zero, prefer `look_at_from_position(from, look_target, UP)`
  even when inside tree (avoids post-advance position equality).
- [x] Step 5.2: Test: `tick_entities` on projectile with `prev == to` does not error.

```bash
godot --headless --path client --script res://tests/test_projectile_tick_smoothing.gd
```

## Task 6 — Bot scenarios

Files:

- Create: `tools/bot/scenarios/94_player_path_budget_lab.json`
- Modify: `shared/rules/worlds.v0.json` (if new `player_path_budget_lab` preset required)

- [x] Step 6.1: Compact lab: player spawn, monster behind wall strip, single `action_entity` or
  `move_to` intent toward unreachable approach.
- [x] Step 6.2: Assert `intent_rejected` with `no_path` (or wait for reject event).
- [x] Step 6.3: With `ARPG_PERF_DEBUG=1`, assert `path_nodes_visited` semantic upper bound derived
  from loaded rules (e.g. `<= player_path_nodes_per_tick + small slack`), not 10k+.
- [x] Step 6.4: Classify `"ci_tier": "extended"` (default for new scenarios).

```bash
make bot scenario=player_path_budget_lab
ARPG_PERF_DEBUG=1 make bot scenario=player_path_budget_lab
python3 -c "from tools.bot.ci_pack import validate_ci_pack; validate_ci_pack()"
```

## Task 7 — Lifecycle docs and CI

- [x] Update `docs/CODEMAP.md` with new modules.
- [x] `docs/as-built/v356_player-navigation-guardrails.md`, `PROGRESS.md`, lifecycle row.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `make client-unit`
- [x] `make bot scenario=player_path_budget_lab`
- [x] `make ci`

## Deferred (explicit)

- Client/server reach desync tightening (separate slice if playtest still spams `action_intent`).
- Promoting `94_player_path_budget_lab` into merge CI pack (extended-only unless merge gate needs it).
- Perf overlay new fields for player vs monster node split.
