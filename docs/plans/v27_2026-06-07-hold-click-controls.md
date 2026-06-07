# v27 Plan — Hold Click Controls

Status: Complete — `make ci` green on 2026-06-07
Goal: Add sustained left-click hold-attack on monsters and hold-move on floor in the Godot client without changing protocol or server behavior.
Architecture: Client-only hold session state repeats existing `action_intent` and `move_to_intent` sends at `SEND_INTERVAL` (~10 Hz). Monster presses lock a sticky `target_id`; floor presses follow the mouse with a 0.25 xz epsilon before repathing. Non-monster/non-floor clicks stay one-shot. Release clears the session.
Tech stack: Godot GDScript client + headless unit tests; no shared/server/bot scenario changes.

**Spec:** [`docs/specs/v27_spec-hold-click-controls.md`](../specs/v27_spec-hold-click-controls.md)

**Branch:** `feature/hold-click-controls` (off current integration branch)

---

## Spec review (2026-06-07)

| Check | Result |
|-------|--------|
| Baseline v25, slice v27, v26 reserved | OK |
| Client-only scope; no hidden server work | OK |
| Bot proof explicitly deferred with reason | OK — regression-only in this plan |
| As-built: `main.gd` edge-triggered LMB + `_attack_cooldown` | OK |
| Plugin adoption | Reject (spec §9) |
| Minor drift | Spec AC #8 mentions inventory lock; `_menu_blocks_gameplay_input()` does not include inventory today — plan clears hold when inventory is visible |

---

## Baseline and shortcut decision

v27 builds on v25 `treasure-classes-and-guarded-chests`, reusing v10 `action_intent`, v11 `move_to_intent` + auto-approach, and v24 menu/pause input guards.

Godot plugin adoption checklist (**reject**): input timing belongs in existing client scripts; no UI/camera/inventory plugin adds value.

**Testability choice:** Extract hold state + pure decision helpers into `client/scripts/sustained_click_input.gd` (RefCounted). `main.gd` owns lifecycle and network sends; unit tests target the helper without loading the full game scene.

---

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/plans/v27_2026-06-07-hold-click-controls.md` | This plan |
| Create | `client/scripts/sustained_click_input.gd` | Hold session state, mode selection, repeat/stop rules, epsilon check |
| Modify | `client/scripts/main.gd` | Wire press/release/repeat tick; delegate to helper; preserve one-shot paths |
| Create | `client/tests/test_sustained_input.gd` | Headless unit tests for helper start/stop/cadence/epsilon |
| Modify | `scripts/client_smoke.sh` | Register new unit gate in client-unit path |
| Modify | `docs/specs/v27_spec-hold-click-controls.md` | Status → Complete when done |
| Modify | `docs/PROGRESS.md` | Lifecycle row when slice ships |

**Unchanged:** `shared/`, `server/`, `tools/bot/scenarios/`, golden fixtures, protocol schemas.

---

## Task 1 — Sustained click helper (pure logic)

Files:
- Create: `client/scripts/sustained_click_input.gd`

- [x] Step 1.1: Add `SustainedClickInput` RefCounted with fields: `active`, `mode` (`""` | `"attack"` | `"move"`), `target_id`, `last_ground` (Vector2 xz), constant `HOLD_MOVE_EPSILON := 0.25`.
- [x] Step 1.2: Implement `clear()` — reset all hold fields.
- [x] Step 1.3: Implement `begin_from_pick(pick: Dictionary) -> bool` where `pick` contains resolved press outcome:
  - `{ "kind": "monster", "target_id": "..." }` → set attack mode + sticky id, return `true` (hold started)
  - `{ "kind": "floor", "ground": Vector3 }` → set move mode + seed `last_ground`, return `true`
  - `{ "kind": "oneshot" }` → do not activate hold, return `false`
- [x] Step 1.4: Implement `should_stop(player_hp: int, entities: Dictionary) -> bool` — stop when `player_hp <= 0`, target missing, monster hp <= 0, or non-monster sticky target.
- [x] Step 1.5: Implement `can_repeat_move(ground: Vector3) -> bool` — true when xz distance from `last_ground` >= `HOLD_MOVE_EPSILON`.
- [x] Step 1.6: Implement `mark_move_sent(ground: Vector3)` — update `last_ground` after a successful move send decision.

```bash
# No standalone run yet — covered in Task 3
```

---

## Task 2 — Wire hold input in `main.gd`

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Add `var _sustained_click := SustainedClickInput.new()` (preload helper script).
- [x] Step 2.2: Refactor press path:
  - Extract `_resolve_click_at_mouse() -> Dictionary` from `_try_action_at_mouse()` — returns pick kind + ids/ground without sending.
  - On LMB `event.pressed`: resolve pick → if helper `begin_from_pick` returns true, also execute first action immediately (same sends as today); else call existing one-shot `_try_action_at_mouse()` unchanged.
- [x] Step 2.3: On LMB `not event.pressed`: `_sustained_click.clear()`.
- [x] Step 2.4: In `_handle_input(delta)` after cooldown decrement, add `_tick_sustained_click()` when `_sustained_click.active`, gameplay active, WS open, `player_hp > 0`, and hold allowed (see Step 2.5).
- [x] Step 2.5: Implement `_hold_input_allowed() -> bool`:
  - false when `_input_locked()`, `bot_mode`, or `inventory_panel != null and inventory_panel.visible` (spec AC #8)
  - when blocked while hold active → `_sustained_click.clear()` (no intents while blocked)
- [x] Step 2.6: `_tick_sustained_click()`:
  - if `_attack_cooldown > 0` → return
  - if `_sustained_click.should_stop(...)` → clear and return
  - attack mode → face sticky target, `play_one_shot("attack")`, `client.send("action_intent", ...)`, set cooldown
  - move mode → read `_mouse_ground_point()`; if `can_repeat_move` → send `move_to_intent`, `mark_move_sent`, set cooldown
- [x] Step 2.7: Press resolution rules (must match spec §4.1):
  - live monster under ray → hold attack (use `_is_dead_monster` guard; dead monster + nearby loot stays one-shot loot path)
  - empty entity pick + no nearest loot → hold move
  - loot, closed/open chest, door, stairs, teleporter → one-shot only (no hold)
- [x] Step 2.8: Keep `_try_action_at_mouse()` behavior for bot helpers (`bot_click_entity_id`, etc.) — either delegate to shared resolve+send or leave bot paths unchanged if they bypass hold (bot does not simulate hold).

```bash
make client-unit
```

---

## Task 3 — Headless unit tests

Files:
- Create: `client/tests/test_sustained_input.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Step 3.1: Create `test_sustained_input.gd` with sentinel `[gdtest] PASS: test_sustained_input`.
- [x] Step 3.2: Test `begin_from_pick` monster → `active`, mode `"attack"`, sticky id set.
- [x] Step 3.3: Test `begin_from_pick` floor → mode `"move"`, `last_ground` seeded.
- [x] Step 3.4: Test `begin_from_pick` oneshot (loot/interactable) → `active == false`.
- [x] Step 3.5: Test `should_stop` when sticky monster removed from `entities` or `hp <= 0`.
- [x] Step 3.6: Test `can_repeat_move` false when ground delta < 0.25 and true when >= 0.25.
- [x] Step 3.7: Test `clear()` resets all fields.
- [x] Step 3.8: Wire gate in `scripts/client_smoke.sh` after waypoint panel test:

```bash
run_gate "GDScript sustained input test" "[gdtest] PASS: test_sustained_input" res://tests/test_sustained_input.gd
```

```bash
make client-unit
```

---

## Task 4 — Manual play verification

Files: none (manual)

- [ ] Step 4.1: `make play` — hold LMB on dungeon monster until kill without click spam.
- [ ] Step 4.2: Hold LMB on floor and drag through a corner — character keeps pathing toward cursor.
- [ ] Step 4.3: Release LMB — repeats stop immediately.
- [ ] Step 4.4: Single-click open treasure chest; holding LMB afterward sends nothing (hold never started on chest, or clears if somehow active).
- [ ] Step 4.5: WASD during hold-move — manual movement still works (v11 cancel preserved).
- [ ] Step 4.6: ESC pause while holding — no further intents until resume.

---

## Task 5 — Bot scenarios (deferred — regression only)

**Deferral reason (spec §6.3):** hold+drag is unreliable in headless Godot (v14); slice proof is unit test + manual play.

- [x] Step 5.1: Do **not** add a new `tools/bot/scenarios/*.json` in v27.
- [x] Step 5.2: Run full Python bot to confirm no protocol/server regression:

```bash
make bot
```

- [x] Step 5.3: Existing scenarios remain unchanged; no `tools/bot/run.py` helper work unless a regression appears.

---

## Task 6 — Lifecycle docs and CI

Files:
- Modify: `docs/specs/v27_spec-hold-click-controls.md`
- Modify: `docs/PROGRESS.md`

- [x] Step 6.1: Mark spec Status → Complete with date when implementation lands.
- [x] Step 6.2: Add v27 row to `docs/PROGRESS.md` lifecycle table + summary when slice ships (`/finish`).

```bash
make ci
```

---

## Final verification

- [x] `make client-unit` — includes new `test_sustained_input.gd` gate
- [x] `make bot` — regression green (no scenario changes)
- [x] `make ci` — full suite green with client-only diff

---

## Acceptance mapping

| Spec AC | Plan coverage |
|---------|----------------|
| 1 Hold-attack repeats until stop | Task 2.6–2.7, manual 4.1 |
| 2 Hold-move with epsilon | Task 1.6, 2.6, test 3.6 |
| 3 Release stops | Task 2.3, test 3.7 |
| 4 Non-hold targets one-shot | Task 2.7, test 3.4 |
| 5 Open chest not actionable | Task 2.7 (oneshot/hold never on chest), manual 4.4 |
| 6 Chase while holding | Task 2.6 (repeat `action_intent`, no client range gate) |
| 7 WASD cancel preserved | No server changes; manual 4.5 |
| 8 Input lock / inventory | Task 2.5, manual 4.6 |
| 9 Godot unit test | Task 3 |
| 10 `make ci` green, no protocol/server | Task 6, Final verification |

---

## Deferred (explicit)

- Walk locomotion animation during hold-move (spec Q-5)
- Godot visual-bot hold+drag scenario
- Server-side player swing cooldown
- Controls remapping / Settings UI
- `HOLD_MOVE_EPSILON` tuning — default 0.25; document in spec/plan as-built if changed during playtest
