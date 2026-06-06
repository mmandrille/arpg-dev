# Inventory UI Implementation Plan

> **For agentic workers:** implement task-by-task and keep checkbox status current. Do not accept a UI-only inventory panel: the slice is complete only when protocol contracts, Go sim, replay/resume, bot scenario, and Godot panel state all agree.

**Goal:** Add a Diablo-style inventory panel that mirrors authoritative server inventory, supports equip/unequip/drop gestures, removes the `Q` equip shortcut, and proves dropped items return to the world as pickup-able loot.

**Architecture:** The Go sim owns inventory, equipment, drop placement, loot spawning, persistence through recorded intents, and all reject reasons. The Godot panel is display/input only: it sends protocol intents and waits for snapshot/delta reconciliation.

**Tech stack:** Shared JSON schemas/examples, Go sim/tests, Python bot/replay checks, Godot 4 GDScript UI and smoke tests.

**Spec:** [`docs/specs/v13_spec-inventory-ui.md`](../specs/v13_spec-inventory-ui.md)

**Branch:** `feature/inventory-ui` off the current integration branch.

---

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/protocol/messages.v0.schema.json` | Add `unequip_intent` and `drop_intent` payloads |
| Modify | `shared/protocol/envelope.v0.schema.json` | Allow new client intent types |
| Modify | `shared/protocol/session_snapshot.v0.schema.json` | Allow item `slot: ""`; add event field coverage |
| Modify | `shared/protocol/state_delta.v0.schema.json` | Allow item `slot: ""`, `inventory_remove`, and new events |
| Modify | `shared/protocol/examples/state_delta.json` | Example drop tick with `inventory_remove` + loot spawn |
| Create | `shared/protocol/examples/unequip_intent.json` | Protocol example |
| Create | `shared/protocol/examples/drop_intent.json` | Protocol example |
| Modify | `shared/rules/worlds.v0.json` | Add `inventory_lab` (loot at `(5, 5)` for in-reach pickup from spawn) |
| Create | `shared/golden/inventory_drop.json` | Pin isolated sim drop placement at `(5, 5)` from spawn |
| Create | `shared/golden/inventory_drop.v0.schema.json` | Golden schema |
| Modify | `tools/validate_shared.py` | Validate new golden references/constants |
| Modify | `server/internal/inputdecode/inputdecode.go` | Decode new intents; extend `IsClientIntent` |
| Modify | `server/internal/game/types.go` | Intent structs, `OpInventoryRemove`, `Change.MarshalJSON`, event `item_instance_id` |
| Modify | `server/internal/game/sim.go` | Implement unequip/drop/drop-placement handlers; extend `player_dead` gate |
| Modify | `server/internal/game/game_test.go` | Unit and golden tests |
| Modify | `server/internal/replay/replay_test.go` | Replay/resume parity |
| Modify | `server/internal/realtime/runner.go` | Persist `OpInventoryRemove` via store |
| Modify | `server/internal/store/interfaces.go`, `repos.go` | Add `RemoveInventoryItem` |
| Modify | `server/internal/http/ws_test.go` | Reconnect resume after drop if not covered elsewhere |
| Create | `tools/bot/scenarios/07_inventory_lab.json` | End-to-end scenario |
| Modify | `tools/bot/run.py` | Add unequip/drop actions and assertion |
| Modify | `tools/bot/test_protocol.py` | Unit coverage for new actions/assertions |
| Create | `client/scripts/inventory_panel.gd` | Custom Control UI |
| Modify | `client/scripts/main.gd` | Mount panel, wire intents, remove `Q`, process `inventory_remove` |
| Modify | `client/scripts/smoke.gd` or create `client/tests/test_inventory_ui.gd` | Headless panel/model checks |
| Modify | `client/tests/test_golden.gd` | Validate inventory drop golden if practical |
| Modify | `docs/PROGRESS.md` | Add v13 completion notes when shipped |

## Plugin adoption

- [x] Consult `docs/godot-plugins-and-shortcuts.md`.
- [x] Decision: **reject** GLoot, Godot-Inventory, Expresso Inventory System, Wyvernbox, and RPGNodes for v13.
- [x] Reason: this slice only needs one weapon slot, a small bag grid, tooltips, and drag/drop intent sending. A plugin would add client-side inventory logic surface that must be stripped or wrapped, which is unnecessary for the first authoritative proof.

---

## Task 1: Shared Contracts, Rules, And Golden

- [x] **Step 1.1:** Add `unequip_intent` and `drop_intent` to `messages.v0.schema.json` and `envelope.v0.schema.json`.

- [x] **Step 1.2:** Add protocol examples:

```text
shared/protocol/examples/unequip_intent.json
shared/protocol/examples/drop_intent.json
```

- [x] **Step 1.3:** Extend item wire schemas so `slot` accepts `"weapon"` and `""`.

- [x] **Step 1.4:** Add `inventory_remove { item_instance_id }` to `state_delta` changes. Keep snapshots as full post-state only.

- [x] **Step 1.5:** Extend event schemas and Go wire intent to support:
  - `item_unequipped`
  - `item_dropped`
  - optional event `item_instance_id`, required for `item_dropped`

- [x] **Step 1.6:** Add `inventory_lab` to `shared/rules/worlds.v0.json`. Place initial
  `rusty_sword` loot at `(5, 5)` so pickup from player spawn `(4, 5)` is in unarmed reach without
  a movement step.

- [x] **Step 1.7:** Add `shared/golden/inventory_drop.json` and schema. Pin isolated sim drop from
  spawn `(4, 5)` with no prior movement → expected loot at `(5, 5)`. Bot scenario does **not**
  assert exact drop coordinates (see spec §4.7).

- [x] **Step 1.8:** Update `shared/protocol/examples/state_delta.json` with a representative drop
  tick (`equipped_update`, `inventory_remove`, loot `entity_spawn`, `item_dropped`).

- [x] **Step 1.9:** Register golden validation in `tools/validate_shared.py`.

- [x] **Step 1.10:** Run:

```bash
make validate-shared
```

---

## Task 2: Server Inventory Intents

- [x] **Step 2.1:** Add `UnequipIntent` and `DropIntent` input structs and route them through `inputdecode`.

- [x] **Step 2.2:** Treat `unequip_intent` and `drop_intent` as gameplay intents rejected with
  `player_dead` when the player is dead. Extend the `applyInput` dead-player switch alongside
  `equip_intent`.

- [x] **Step 2.3:** Implement `handleUnequip`:
  - validate `slot == "weapon"`
  - reject `slot_empty`
  - mark item unequipped
  - emit `inventory_update`
  - emit `equipped_update` with `null`
  - emit `item_unequipped`
  - ack

- [x] **Step 2.4:** Implement deterministic `findDropPosition(player.pos)`:
  - offsets in fixed cardinal/diagonal order
  - two rings at `step` and `2*step`
  - use `playerPositionBlocked`
  - reject overlap with existing loot using sorted entity ids
  - return `no_drop_space` without mutation when no candidate works

- [x] **Step 2.5:** Implement `handleDrop` atomically:
  - validate inventory ownership
  - find drop position before mutating inventory/equipment
  - if item was equipped: emit `equipped_update` with `weapon: null` (no `inventory_update` on the removed item)
  - remove inventory item and emit `inventory_remove`
  - spawn loot with deterministic `alloc()`
  - emit `item_dropped` with loot entity id and dropped `item_instance_id`
  - ack

- [x] **Step 2.6:** Add `OpInventoryRemove` to `Change.MarshalJSON` in `types.go` (easy to miss).

- [x] **Step 2.7:** Add `RemoveInventoryItem` to the store layer and handle `OpInventoryRemove` in
  `runner.persistTick`. Resume/`/state` reconstruct from replay inputs, but DB rows must not stale.

- [x] **Step 2.8:** Ensure dropped loot reuses existing loot entity shape so `action_intent` pickup works unchanged.

- [x] **Step 2.9:** Add Go tests:
  - `TestUnequipWeapon`
  - `TestDropInventoryItem`
  - `TestDropEquippedWeapon`
  - `TestDropThenPickup`
  - `TestDropNoSpace`
  - `TestInventoryDropGolden`

- [x] **Step 2.10:** Run:

```bash
cd server && go test ./internal/game/... -run 'Unequip|Drop|Inventory' -v
```

---

## Task 3: Replay, Resume, And HTTP/WebSocket Parity

- [x] **Step 3.1:** Confirm persisted input decode includes new intents so replay reconstruction naturally rebuilds unequip/drop outcomes.

- [x] **Step 3.2:** Add replay/resume coverage for the inventory lab flow if existing bot replay
  assertions do not catch inventory removal plus loot respawn state. Extend `ws_test.go` when
  reconnect resume after drop is not already covered.

- [x] **Step 3.3:** Ensure `/state`, fresh WebSocket snapshots, reconnect snapshots, and replay timeline all serialize:
  - item `slot: ""`
  - cleared `equipped.weapon`
  - `inventory_remove`
  - `item_dropped.item_instance_id`

- [x] **Step 3.4:** Run:

```bash
cd server && go test ./...
```

---

## Task 4: Bot Scenario And Protocol Tests

- [x] **Step 4.1:** Add `tools/bot/scenarios/07_inventory_lab.json` with pickup, equip, unequip, drop, re-pickup, and re-equip.

- [x] **Step 4.2:** Add bot action `unequip_slot`.

- [x] **Step 4.3:** Add bot action `drop_inventory_item`.

- [x] **Step 4.4:** Add assertion **`equipped_weapon_def`** — given `item_def_id`, assert
  `equipped.weapon` points at an inventory row with that def:

```json
{ "type": "equipped_weapon_def", "item_def_id": "rusty_sword" }
```

- [x] **Step 4.5:** Update bot state reconciliation for `inventory_remove` (remove row by instance id).

- [x] **Step 4.6:** Update `tools/bot/test_protocol.py` for new state changes, actions, and reject reason handling.

- [x] **Step 4.7:** Run:

```bash
.venv/bin/python -m pytest tools/bot/test_protocol.py -q
```

---

## Task 5: Godot Inventory Panel

- [x] **Step 5.1:** Create `client/scripts/inventory_panel.gd` as a custom `Control` tree with:
  - `I` toggle from `main.gd`
  - one weapon slot
  - 4-column bag grid
  - empty thumbnail frames
  - Diablo-dark placeholder theme
  - hover tooltips from `shared/rules/items.v0.json`

- [x] **Step 5.2:** Add a small data-binding surface from `main.gd` so the panel receives authoritative inventory/equipped state without duplicating protocol parsing.

- [x] **Step 5.3:** Wire interactions:
  - double-click bag weapon sends `equip_intent`
  - drag bag weapon to weapon slot sends `equip_intent`
  - drag weapon slot to bag sends `unequip_intent`
  - drag bag/equipped item outside panel sends `drop_intent`
  - invalid drops snap back with no intent

- [x] **Step 5.4:** Do not optimistically mutate local inventory. Re-render only from snapshot/delta state.

- [x] **Step 5.5:** Process `inventory_remove` in `main.gd` and forward it to the panel.

- [x] **Step 5.6:** Remove `Q` equip from manual input, autoplay, and debug HUD. Keep autoplay using explicit `equip_intent` where needed.

- [x] **Step 5.7:** Ensure `EquipmentVisualResolver` clears weapon visuals when `equipped_update.item_instance_id == null`.

---

## Task 6: Client Smoke And Visual Checks

- [x] **Step 6.1:** Add headless inventory UI/model smoke coverage:
  - snapshot populates bag count
  - `inventory_update` marks weapon equipped
  - `equipped_update` clears slot
  - `inventory_remove` removes bag row

- [x] **Step 6.2:** Keep visual replay mode and `ARPG_AUTOPLAY` from showing an interactive panel while input is locked.

- [x] **Step 6.3:** Run:

```bash
make client-smoke
```

---

## Task 7: Final Verification And Docs

- [x] **Step 7.1:** Run full gate:

```bash
make ci
```

- [ ] **Step 7.2:** Manual pass:

```bash
make play
```

Verify `I` opens the panel, world pickup works, double-click/drag equip works, drag to bag unequips, drag outside drops loot beside the player, re-pickup and re-equip work, and `Q` does nothing.

- [ ] **Step 7.3:** Optional focused visual replay:

```bash
make bot-visual scenario=07_inventory_lab.json
```

- [x] **Step 7.4:** Update `docs/PROGRESS.md` once the slice ships:
  - latest completed slice v13
  - lifecycle table row
  - “What v13 proved”
  - explicit non-goals/deferred gaps
