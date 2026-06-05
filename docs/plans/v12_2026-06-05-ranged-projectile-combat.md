# Ranged Projectile Combat Implementation Plan

> **For agentic workers:** implement task-by-task and keep checkbox status current. Do not merge partial work where projectiles are only client-visible or only server-local; the slice is accepted only when shared contracts, Go sim, replay/resume, bot, and Godot presentation are wired together.

**Goal:** Server-authoritative ranged weapon combat: equipping `training_bow` and clicking a monster spawns a traveling projectile, resolves wall/monster collision over ticks, rolls hit/damage only at impact, and auto-moves to a clear line of fire when the direct shot is blocked.

**Architecture:** The Go sim owns projectile spawn, movement, collision, RNG, damage, and entity lifecycle. The client renders only wire-visible `projectile` entities and reconciles to server deltas. Shared JSON rules and golden fixtures pin item validation, projectile constants, and impact outcomes.

**Tech stack:** Go sim/tests, shared JSON schemas + validation, Python protocol bot/replay checks, Godot 4 GDScript client smoke.

**Spec:** [`docs/specs/v12_spec-ranged-projectile-combat.md`](../specs/v12_spec-ranged-projectile-combat.md)

**Branch:** `feature/ranged-projectile-combat` off current integration branch.

---

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/golden/ranged_projectile.json` | Pinned ranged projectile cases |
| Create | `shared/golden/ranged_projectile.v0.schema.json` | Golden schema |
| Create | `tools/bot/scenarios/06_ranged_lab.json` | End-to-end ranged bot proof |
| Modify | `shared/rules/items.v0.json` | Add `training_bow` with ranged fields |
| Modify | `shared/rules/items.v0.schema.json` | Validate `attack_mode` / `projectile_speed` conditionals |
| Modify | `shared/rules/monsters.v0.json` | Add `training_dummy_ranged` |
| Modify | `shared/rules/worlds.v0.json` | Add `ranged_lab` preset |
| Modify | `shared/protocol/state_delta.v0.schema.json` | Add projectile entity and projectile events |
| Modify | `shared/protocol/session_snapshot.v0.schema.json` | Add projectile entity and recent event shapes |
| Modify | `shared/protocol/examples/state_delta.json` | Include representative projectile spawn/update/remove/event |
| Modify | `tools/validate_shared.py` | Ranged item and golden/rules cross-checks |
| Modify | `server/internal/game/rules.go` | Load ranged item fields |
| Modify | `server/internal/game/types.go` | Projectile entity/change/event wire shape |
| Modify | `server/internal/game/sim.go` | Ranged dispatch, projectile state, tick integration, sweep collision |
| Modify | `server/internal/game/game_test.go` | Ranged projectile unit and golden tests |
| Modify | `server/internal/replay/replay_test.go` | Projectile replay determinism coverage |
| Modify | `server/internal/http/ws_test.go` | WebSocket projectile lifecycle/resume coverage if needed |
| Modify | `tools/bot/run.py` | Scenario assertion for never entering melee range |
| Modify | `tools/bot/test_protocol.py` | Ranged scenario catalog/assertion coverage |
| Modify | `client/scripts/main.gd` | Spawn/update/remove placeholder projectile nodes |
| Modify | `client/tests/test_golden.gd` | Validate ranged projectile golden/rules consistency |
| Modify | `client/scripts/smoke.gd` | Cover projectile entity rendering if practical |
| Modify | `docs/PROGRESS.md` | Add v12 completion notes when shipped |

## Plugin adoption

- [x] Consult `docs/godot-plugins-and-shortcuts.md`.
- [x] Decision: **reject** projectile/VFX plugins for v12. The slice needs a simple placeholder mesh and server-authoritative wire entity, not client-side gameplay logic or production art.

---

## Task 1: Shared Contracts And Rules

- [x] **Step 1.1:** Add `attack_mode` and `projectile_speed` support to `items.v0.schema.json`.

  Required conditionals:
  - `attack_mode: "ranged"` requires `slot: "weapon"`, `equippable: true`, `damage`, `reach`, and positive `projectile_speed`.
  - omitted `attack_mode` or `"melee"` forbids `projectile_speed`.
  - `rusty_sword` remains valid without `attack_mode`.

- [x] **Step 1.2:** Add `training_bow` to `items.v0.json`:

```json
"training_bow": {
  "name": "Training Bow",
  "slot": "weapon",
  "equippable": true,
  "attack_mode": "ranged",
  "damage": { "min": 2, "max": 4 },
  "reach": 16.0,
  "projectile_speed": 50.0
}
```

- [x] **Step 1.3:** Add `training_dummy_ranged` to `monsters.v0.json`. Use low enough HP for one successful bow hit in the happy path, or tune bow damage/golden accordingly.

- [x] **Step 1.4:** Add `ranged_lab` to `worlds.v0.json`: player at `(0, 5)`, `training_bow` loot at `(1, 5)`, target monster at `(14, 5)`, walls above/below the y=5 shot line, and a wall-blocking row for the blocked-shot case.

- [x] **Step 1.5:** Extend protocol schemas for `projectile` entity fields: `owner_id`, `target_id`, and `projectile_def_id`.

- [x] **Step 1.6:** Extend event schemas for `projectile_blocked` and `projectile_expired`; ensure `attack_missed` remains valid with `correlation_id` for ranged misses.

- [x] **Step 1.7:** Create `ranged_projectile.json` + schema, then register cross-checks in `tools/validate_shared.py` for referenced world, monster, weapon, constants, and expected event names.

- [x] **Step 1.8:** Run:

```bash
make validate-shared
```

---

## Task 2: Server Ranged Dispatch

- [x] **Step 2.1:** Extend item rule structs with `AttackMode` and `ProjectileSpeed`; default missing `AttackMode` to melee in helper methods, not by mutating loaded rules.

- [x] **Step 2.2:** Add helpers:

```go
playerAttackMode() string
playerActionReach() float64
inActionRange(target *entity) bool
```

- [x] **Step 2.3:** Refactor monster `action_intent` dispatch:
  - melee monster action uses existing instant `attackTarget`.
  - ranged monster action spawns a projectile when in range.
  - loot and interactables continue to require melee range.

- [x] **Step 2.4:** Generalize v11 approach planning so monster + ranged weapon searches for a reachable `inActionRange` standoff cell, while melee, loot, and interactable actions keep the existing melee approach predicate.

- [x] **Step 2.5:** Add `projectile_busy` reject when a player-owned projectile is already in flight.

- [x] **Step 2.6:** Add focused tests for ranged dispatch, ranged auto-approach, loot/door melee behavior while bow-equipped, and `projectile_busy`.

- [x] **Step 2.7:** Add clear-line ranged dispatch:
  - ranged monster actions only fire immediately when in range and the initial shot sweep to the target is clear.
  - blocked direct-line ranged clicks queue auto-navigation to the nearest reachable in-range clear-line standoff.
  - pending ranged actions re-check clear line on arrival before spawning the projectile.
  - no reachable clear-line standoff rejects `no_path`.

---

## Task 3: Server Projectile Simulation

- [x] **Step 3.1:** Add internal projectile state to `Sim`, including owner, target, fixed direction, speed, traveled distance, max distance, source message/correlation ids, and weapon damage snapshot.

- [x] **Step 3.2:** On accepted ranged fire, create a deterministic projectile id, spawn at player position, emit `entity_spawn`, and acknowledge exactly once.

- [x] **Step 3.3:** Advance projectiles once per tick after movement and before tick increment, in ascending projectile id order.

- [x] **Step 3.4:** Implement swept collision:
  - inflate walls and closed-barrier AABBs by `projectileRadius`.
  - test live monsters as circles expanded by `projectileRadius`.
  - choose smallest intersection `t`; tie-break wall, closed interactable, monster, then entity id.
  - dead monsters and open doors do not block.

- [x] **Step 3.5:** Resolve monster impact at impact tick:
  - consume hit RNG.
  - on miss emit `attack_missed` with original correlation id and no retaliation.
  - on hit roll bow damage and reuse existing damage/kill/loot/retaliation path.

- [x] **Step 3.6:** Resolve wall/barrier impact with `projectile_blocked`; resolve max travel with `projectile_expired`; both remove the projectile and deal no damage.

- [x] **Step 3.7:** Emit projectile `entity_update` on movement and `entity_remove` on impact/expiry.

- [x] **Step 3.8:** Add Go tests:
  - `TestRangedProjectileGolden`
  - `TestRangedKillBeyondMeleeRange`
  - `TestProjectileBlockedByWall`
  - `TestRangedImpactMissNoRetaliation`
  - `TestProjectileBusyRejectsSecondFire`
  - `TestMeleeUnchangedWithRustySword`
  - `TestRangedAutoApproachThenFire`
  - `TestRangedBlockedLineAutoMovesUntilClearThenFires`

- [x] **Step 3.9:** Run:

```bash
cd server && go test ./internal/game/... -run 'Ranged|Projectile|Bow' -v
```

---

## Task 4: Replay, Resume, And HTTP/WebSocket Coverage

- [x] **Step 4.1:** Ensure projectile state reconstructs from recorded inputs through existing deterministic replay, including mid-flight snapshots.

- [x] **Step 4.2:** Add replay tests that compare projectile spawn/update/remove ticks and final monster/player state.

- [x] **Step 4.3:** Add WebSocket/session coverage if schemas or snapshot serialization need explicit regression tests for projectile entities.

- [x] **Step 4.4:** Run:

```bash
cd server && go test ./...
```

---

## Task 5: Bot Scenario

- [x] **Step 5.1:** Add `06_ranged_lab.json`: pick up bow, equip bow, click `training_dummy_ranged`, wait for `monster_killed`.

- [x] **Step 5.2:** Add `player_never_in_melee_range_of` assertion, using observed player and monster positions from snapshots/deltas throughout the scenario.

- [x] **Step 5.3:** Extend `06_ranged_lab.json` so the monster click starts behind straight-line cover and proves auto-navigation moves the player until the shot is clear before firing.

- [x] **Step 5.4:** Keep `01` through `05` unchanged and green; do not alter global `combat.base_hit_chance`.

- [x] **Step 5.5:** Run:

```bash
.venv/bin/python -m pytest tools/bot/test_protocol.py -q
```

---

## Task 6: Godot Presentation

- [x] **Step 6.1:** Add `projectile` handling in `main.gd` entity spawn/update/remove paths. Use a simple in-repo primitive mesh keyed by `projectile_def_id == "training_arrow"`.

- [x] **Step 6.2:** Snap or lightly lerp projectile nodes to authoritative positions. Do not add client-side projectile collision or hit prediction.

- [x] **Step 6.3:** Ensure visual replay manifests and timelines render projectile lifecycle through existing delta handlers.

- [x] **Step 6.4:** Update `test_golden.gd` to validate `ranged_projectile.json` references and constants. Add a lightweight evaluator only if it stays small.

- [x] **Step 6.5:** Run:

```bash
make client-smoke
```

---

## Task 7: Final Verification And Docs

- [x] **Step 7.1:** Run full gate:

```bash
make ci
```

- [x] **Step 7.2:** Optionally inspect the visual scenario:

```bash
make bot-visual
```

- [x] **Step 7.3:** Update `docs/PROGRESS.md` with v12 lifecycle row, what the slice proved, explicit non-goals, and any deferred gaps.

- [x] **Step 7.4:** Mark the spec status complete only after `make ci` is green.

## Implementation Notes

- Keep projectile RNG isolated to impact. Projectile spawn and flight must not consume RNG.
- Do not add a projectile catalog in v12; use inline `"training_arrow"` as specified.
- Be careful with pending auto-nav dispatch: ranged pending action must re-check target state and equipped mode on arrival, then spawn the projectile without a second ack.
- Preserve deterministic iteration: sorted projectile ids and sorted entity ids for collision tie-breaks.
- Treat protocol v0 schema extension as coordinated server/client work; do not introduce backward-compat shims unless tests require them.
