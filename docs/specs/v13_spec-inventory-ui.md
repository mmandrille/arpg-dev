# Spec: `inventory-ui`

Status: Draft
Branch: `feature/inventory-ui`
Slice: v13 — Diablo-style inventory panel with drag/drop equip, unequip, and floor drop
Baseline: slice v12 `ranged-projectile-combat` (complete on current integration branch, `make ci` green on 2026-06-05)
Related:

- [`v12_spec-ranged-projectile-combat.md`](v12_spec-ranged-projectile-combat.md) — ranged weapons; bot already equips via `equip_intent`
- [`v10_spec-click-action-and-melee-range.md`](v10_spec-click-action-and-melee-range.md) — pickup remains `action_intent` on loot entities
- [`v2_spec-equip-and-see-it.md`](v2_spec-equip-and-see-it.md) — `EquipmentVisualResolver`, item visuals pipeline
- [`v8_spec-equipped-weapon-damage.md`](v8_spec-equipped-weapon-damage.md) — weapon stats for tooltips
- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../godot-plugins-and-shortcuts.md`](../godot-plugins-and-shortcuts.md) — **reject** inventory plugins for v13; custom UI only
- ADR-0001 (authoritative server; client sends intents only)
- ADR-0007 (client-only presentation; panel layout and drag feedback are not on the wire)

## 1. Purpose

Give humans a **Diablo-dark inventory panel** that mirrors authoritative server inventory state and
sends the same production intents the bot already uses — plus new **unequip** and **drop** intents
where drag gestures require server outcomes.

After this slice:

- **`I`** toggles an inventory overlay: **weapon equip slot** + **bag list** (all inventory items,
  including non-equippable `training_badge`).
- **Rich tooltips** show item name, slot, damage range, reach, and `attack_mode` from
  `shared/rules/items.v0.json`, with **empty thumbnail placeholders** (no new item art pipeline).
- **Double-click** a bag weapon → `equip_intent`.
- **Drag** bag weapon → weapon slot → `equip_intent`.
- **Drag** equipped weapon → bag area → `unequip_intent { slot: "weapon" }`.
- **Drag** any bag item (or equipped item) **outside** the panel → `drop_intent { item_instance_id }`
  → server spawns a **loot entity** in the **nearest valid tile beside the player** (no wall/body
  overlap) and removes the item from inventory.
- **`Q` equip shortcut is removed** from `main.gd`, autoplay, and debug hints. Equip is UI or
  explicit `equip_intent` only.
- New world **`inventory_lab`** and bot scenario **`07_inventory_lab.json`** prove pickup → equip →
  unequip → drop → re-pickup → re-equip over protocol, plus `/state`, reconnect resume, and replay.
- Headless smoke asserts the inventory UI controller syncs item counts and equipped state from
  snapshot/deltas (no pixel tests).

The proof is **new intents → sim inventory mutations → loot respawn → bot scenario → client panel
sync**, not stash grids, vendors, crafting, or plugin adoption.

## 2. Current Problems

### 2.1 No human inventory surface

`client/scripts/main.gd` keeps `inventory` and `equipped` arrays/dicts and shows counts in the debug
HUD only. Players cannot see `training_badge`, compare weapons, or equip except via **`Q`**, which
always equips `inventory[0]` regardless of choice.

### 2.2 No authoritative unequip or player-initiated drop

`handleEquip` can swap weapons but there is no intent to clear the weapon slot without equipping
another item. Items cannot return to the world except monster loot drops. Drag-to-unequip and
drag-outside-to-drop require new server handlers.

### 2.3 Bot already uses protocol equip; client shortcut diverges

Scenarios use `equip_first_inventory_item` / `equip_inventory_item` → `equip_intent`. The **`Q`**
key is a client-only shortcut that agents and docs still reference (v10 debug line). Removing it
aligns human UX with the bot path.

### 2.4 Protocol item `slot` schema is weapon-only

Non-equippable items like `training_badge` appear in tests with `"slot": ""`. The snapshot item
schema enum is `["weapon"]` only — tighten in this slice so bag rows validate for all item kinds.

## 3. Non-goals

- **No Godot inventory plugins** (GLoot, Godot-Inventory, etc.) — custom `Control` UI only.
- **No stash**, shared storage, vendor, buy/sell, or crafting UI.
- **No stack splitting**, item combining, or quantity > 1.
- **No new equipment slots** beyond existing `weapon` (UI may show a single equip slot only).
- **No equip-on-click** in the 3D world — equip stays inventory panel or `equip_intent`.
- **No drop range gate** — drop is an inventory action, not a targeted world action; placement is
  server-computed beside the player (§4.2.1).
- **No drop while dead** — reject `player_dead` like other gameplay intents.
- **No item destruction** — drop always creates a pickup-able loot entity.
- **No production item icons** — 48×48 empty placeholder frames only; optional tint by item category.
- **No path preview, combat, or monster AI changes.**
- **No character-scoped persistence** — session inventory only (unchanged from v0).
- **No protocol schema filename/version bump** — extend the existing v0 schemas and update all
  producers/consumers in the same slice.

## 4. Required Design

### 4.1 Protocol: `unequip_intent` and `drop_intent`

Add client message types to `shared/protocol/messages.v0.schema.json` and `envelope.v0.schema.json`:

**Unequip**

```json
{
  "type": "unequip_intent",
  "payload": { "slot": "weapon" }
}
```

Server `handleUnequip`:

1. Reject `invalid_payload` if slot missing or not `"weapon"`.
2. Reject `player_dead` if player `hp <= 0`.
3. Reject `slot_empty` if `equipped[slot]` is null/zero.
4. Set item `equipped = false`; emit `inventory_update`.
5. Clear `equipped[slot]`; emit `equipped_update` with `item_instance_id: null`.
6. Emit event `item_unequipped` with `entity_id` = former instance id.
7. Ack intent.

**Drop**

```json
{
  "type": "drop_intent",
  "payload": { "item_instance_id": "1004" }
}
```

Server `handleDrop`:

1. Reject `invalid_payload` if `item_instance_id` missing.
2. Reject `player_dead`.
3. Reject `not_in_inventory` if item not found.
4. Compute **`dropPos := findDropPosition(player.pos)`** (§4.2.1). Reject `no_drop_space` if no
   valid candidate exists. This happens before inventory/equipment mutation so the operation is
   atomic on reject.
5. If item is equipped, clear equipped slot (same as unequip) before removal.
6. Remove item from `s.inventory` slice and emit `inventory_remove { item_instance_id }`.
7. Spawn **loot** entity at `dropPos` (`entity_spawn`, `type: "loot"`, `item_def_id` from item).
   Use deterministic entity id allocation (existing `alloc()`).
8. Emit event `item_dropped` with `entity_id` = new loot id and `item_instance_id` = dropped
   inventory instance id.
9. Ack intent.

Stable reject reasons (bot assertions):

| Reason | When |
|--------|------|
| `slot_empty` | Unequip with nothing equipped |
| `not_in_inventory` | Drop unknown instance |
| `no_drop_space` | No collision-free tile beside player |
| `not_equippable` | Equip unchanged path |
| `wrong_slot` | Equip unchanged path |
| `player_dead` | Any inventory intent while dead |
| `invalid_payload` | Missing fields |

Add protocol examples: `shared/protocol/examples/unequip_intent.json`,
`shared/protocol/examples/drop_intent.json`.

Extend `state_delta` changes with `inventory_remove`. `session_snapshot` does not need a remove
op because snapshots contain the post-drop inventory.

Extend `session_snapshot` / `state_delta` event schemas and Go `Event` serialization to allow
`item_unequipped` and `item_dropped`. Add `item_instance_id` as an optional event field generally,
and require it for `item_dropped`.

**Item slot wire shape:** extend snapshot/delta item `slot` to `enum: ["weapon", ""]` where `""`
means non-equippable / no slot (e.g. `training_badge`).

### 4.2 Server sim integration

In `server/internal/game/sim.go`:

- Route `unequip_intent` and `drop_intent` in the input dispatcher (same tick as equip; rejected
  when dead).
- `handleDrop` reuses loot entity shape from `dropLoot` / pickup path so `action_intent` pickup
  works on dropped items without new pickup logic.
- Replay and resume reconstruct dropped loot entities and inventory removals from recorded intents.

#### 4.2.1 Drop placement: `findDropPosition`

Player-initiated drops must appear **beside** the player at a **collision-free** point. Reuse the
same world collision sources as movement (v9 walls, live monster circles, closed interactable
barriers) plus loot-specific spacing.

**Constants** (match existing sim values; pin expected positions in golden):

| Constant | Value | Source |
|----------|-------|--------|
| `lootDropRadius` | `0.35` | same as `lootInteractionRadius` |
| `playerRadius` | `0.45` | existing |
| `step` | `navigation.cell_size` (`1.0`) | shared rules |

**Candidate generation** — deterministic ring search around `player.pos`:

1. Build offset list in fixed order: cardinals `(+step,0)`, `(0,+step)`, `(-step,0)`, `(0,-step)`,
   then diagonals `(±step,±step)`, then second ring at `2*step` with the same angular ordering.
2. Cap search at **2 rings** (8 candidates at distance `step`, 8 candidates at `2*step`; 16 total)
   for v13; extend later if needed.
3. For each candidate `pos = player.pos + offset`, accept the **first** that passes all checks:

| Check | Rule |
|-------|------|
| Clear of player body | `distance(pos, player.pos) >= playerRadius + lootDropRadius` |
| Walkable footprint | `!playerPositionBlocked(pos)` — same function as movement (walls, live monsters, closed doors) |
| Clear of other loot | no existing `loot` entity where `circlesOverlap(pos, lootDropRadius, loot.pos, lootDropRadius)` |

Monster corpses (`hp == 0`) and open interactables do not block (consistent with v9 movement).
Loot entities do **not** block `playerPositionBlocked` today; the loot-vs-loot check prevents
stacking multiple drops on the same point.

**Reject:** if no candidate passes → `no_drop_space`. Because placement is validated before any
inventory or equipment mutation, rejection leaves inventory and equipped state unchanged.

**Determinism:** fixed offset order + stable entity iteration for loot overlap checks
(`sortedEntityIDs`).

Go tests (minimum):

- `TestUnequipWeapon` — equip then unequip; `equipped.weapon == null`, item `equipped == false`.
- `TestDropInventoryItem` — drop unequipped badge; inventory count decreases; loot entity present
  **adjacent** to player (not co-located with player center).
- `TestDropEquippedWeapon` — drop equipped sword; slot cleared + loot spawned in one tick.
- `TestDropThenPickup` — drop → `action_intent` pickup → inventory restored.
- `TestDropNoSpace` — surround player with blocking walls in a unit test world; `drop_intent`
  rejects `no_drop_space`.
- Replay/resume parity for inventory lab flow (follow existing test patterns).

### 4.3 World: `inventory_lab`

Add to `shared/rules/worlds.v0.json`:

```json
"inventory_lab": {
  "player_spawn": { "x": 4, "y": 5 },
  "initial_loot": [
    { "item_def_id": "rusty_sword", "position": { "x": 5, "y": 5 } }
  ],
  "monsters": []
}
```

No monsters — keeps the scenario focused on inventory intents. Open floor; no walls required.
Place initial loot at `(5, 5)` so pickup from spawn `(4, 5)` succeeds within unarmed reach without
requiring a movement step.

### 4.4 Bot scenario: `07_inventory_lab.json`

New catalog entry (runs after `06_ranged_lab.json` in filename order):

```json
{
  "id": "inventory_lab",
  "world_id": "inventory_lab",
  "title": "Inventory lab",
  "description": "Pickup, equip, unequip, drop, re-pickup, and re-equip via inventory intents.",
  "steps": [
    { "action": "pick_up_loot", "item_def_id": "rusty_sword" },
    { "action": "equip_inventory_item", "item_def_id": "rusty_sword", "slot": "weapon" },
    { "action": "unequip_slot", "slot": "weapon" },
    { "action": "drop_inventory_item", "item_def_id": "rusty_sword" },
    { "action": "pick_up_loot", "item_def_id": "rusty_sword" },
    { "action": "equip_inventory_item", "item_def_id": "rusty_sword", "slot": "weapon" }
  ],
  "assertions": [
    { "type": "inventory_contains", "item_def_id": "rusty_sword", "equipped": true },
    { "type": "equipped_weapon_def", "item_def_id": "rusty_sword" }
  ]
}
```

New bot actions in `tools/bot/run.py`:

| Action | Behavior |
|--------|----------|
| `unequip_slot` | Send `unequip_intent`; wait until `equipped.weapon == null` and item `equipped == false`. |
| `drop_inventory_item` | Resolve instance id by `item_def_id`; send `drop_intent`; wait until item absent from inventory and loot entity exists (or `item_dropped` event). |

Reuse existing `pick_up_loot` action (sends `action_intent` on loot entity by `item_def_id`).

Add assertion **`equipped_weapon_def`** — asserts `equipped.weapon` resolves to an inventory row
with the given `item_def_id`:

```json
{ "type": "equipped_weapon_def", "item_def_id": "rusty_sword" }
```

Update `tools/bot/test_protocol.py` unit tests for new actions and reject reasons.

**Existing scenarios:** no step changes required — they already use `equip_inventory_item` /
`equip_first_inventory_item`, not `Q`.

### 4.5 Client: custom inventory UI (`inventory_panel.gd`)

New script + scene (or programmatic `Control` tree) mounted from `main.gd`:

**Layout**

- Toggle visibility with **`I`** (manual mode only; hidden during `ARPG_AUTOPLAY` / visual replay, or
  read-only visible — prefer **hidden** when `_input_locked()`).
- **Left column:** single **Weapon** slot (large frame, 64×64 thumbnail area).
- **Right column:** scrollable **bag** grid (4 columns, 48×48 cells).
- Panel anchored bottom-right or center-right; does not pause sim or block movement.

**Diablo-dark art direction (placeholder theme, no external assets)**

| Element | Guideline |
|---------|-----------|
| Panel background | `#12100c` – `#1a1510` gradient, 90% opacity |
| Outer border | 2px `#6b5420` with inner 1px `#c9a227` highlight |
| Slot frame | Inset `#0a0908`, border `#5c4a1f`, hover `#8b6914` |
| Title text | `#c9a227` small caps ("Inventory", "Weapon") |
| Item name | `#e8dcc8` |
| Stat lines | `#a89878` (damage, reach, mode) |
| Equipped highlight | `#3d2e10` fill + `#c9a227` corner accents |
| Thumbnail placeholder | `#2a2420` fill, `#1a1510` inner shadow — no icon texture |

Use Godot `Theme` overrides or inline `StyleBoxFlat` — no PNG skin requirement for v13.

**Data binding**

- Subscribe to the same snapshot/delta path as `main.gd` (extract an `InventoryModel` ref-counted
  helper or callbacks from `main.gd` to avoid duplicating delta parsing).
- Bag lists all `inventory[]` entries; weapon slot reads `equipped.weapon` → resolve row highlight.
- Tooltips on hover: read `shared/rules/items.v0.json` via existing client rules loader (same path
  golden tests use).

**Interactions**

| Gesture | Condition | Intent |
|---------|-----------|--------|
| Double-click bag row | item equippable + weapon slot | `equip_intent` |
| Drag bag → weapon slot | equippable weapon | `equip_intent` on drop |
| Drag weapon slot → bag | item equipped | `unequip_intent` on drop over bag |
| Drag bag row → outside panel | always | `drop_intent` |
| Drag weapon slot → outside panel | equipped | `drop_intent` |
| Double-click non-equippable | — | no-op |
| Drop on invalid target | — | snap back (no intent) |

Use Godot 4 `Control` drag-and-drop (`set_drag_preview`, `_get_drag_data`, `_can_drop_data`,
`_drop_data`). Do not mutate local inventory optimistically — wait for server delta (panel may
show brief drag ghost; reconcile on `inventory_add`, `inventory_update`, `inventory_remove`, and
`equipped_update`).

**Equipment visuals:** existing `EquipmentVisualResolver` continues to react to `equipped_update`;
unequip must clear mounted weapon visual.

**Remove `Q` equip** from `_handle_input` and `_handle_autoplay`. Autoplay equip phase sends
`equip_intent` for the picked-up sword instance id (mirror bot).

Update debug HUD line: remove `Q equip`; add `I inventory`.

### 4.6 Headless smoke extensions

In `client/scripts/smoke.gd` or a focused `client/tests/test_inventory_ui.gd` run via
`make client-smoke`:

- After equip phase, assert inventory panel model reports `equipped == true` and bag count.
- Optionally instantiate `InventoryPanel` headless, feed synthetic snapshot/delta dictionaries,
  assert slot labels and item def ids match (no rendering).

Keep existing equip/resume visual smoke path green.

### 4.7 Golden fixture (recommended)

`shared/golden/inventory_drop.json` — pinned seed + `inventory_lab` player spawn `(4, 5)` after
`drop_intent` for `rusty_sword` with no intervening movement:

```json
{
  "description": "Player drop at inventory_lab spawn places loot at first valid adjacent cell.",
  "world_id": "inventory_lab",
  "player_spawn": { "x": 4, "y": 5 },
  "item_def_id": "rusty_sword",
  "expected_loot_position": { "x": 5, "y": 5 },
  "constants": {
    "loot_drop_radius": 0.35,
    "player_radius": 0.45,
    "drop_step": 1.0
  }
}
```

First cardinal offset `(+1, 0)` from spawn is open floor in `inventory_lab` → loot at `(5, 5)`.
Also assert `equipped.weapon == null` and inventory excludes the dropped instance after the intent.

**Scope:** this golden pins **isolated Go sim** drop placement from spawn with no prior movement.
The bot scenario (`07_inventory_lab.json`) proves the full intent loop but does **not** assert an
exact drop coordinate — only that loot respawns and can be re-picked up.

Go + GDScript consume the fixture (pattern from `melee_reach.json` / `ranged_projectile.json`); bot
scenario remains the end-to-end proof.

## 5. File map (expected)

| Action | Path | Notes |
|--------|------|-------|
| Modify | `shared/protocol/messages.v0.schema.json` | Add intents |
| Modify | `shared/protocol/envelope.v0.schema.json` | Add intent types |
| Modify | `shared/protocol/session_snapshot.v0.schema.json` | Item slot `""`, events |
| Modify | `shared/protocol/state_delta.v0.schema.json` | Item slot `""`, `inventory_remove`, events |
| Modify | `shared/protocol/examples/state_delta.json` | Example drop tick with `inventory_remove` + loot spawn |
| Create | `shared/protocol/examples/unequip_intent.json`, `drop_intent.json` | |
| Modify | `server/internal/game/types.go` | `OpInventoryRemove`, `Change.MarshalJSON`, event `item_instance_id`, intent structs |
| Modify | `server/internal/game/sim.go` | Handlers + `findDropPosition` |
| Modify | `server/internal/inputdecode/inputdecode.go` | Decode new intents; extend `IsClientIntent` |
| Modify | `server/internal/game/game_test.go` | Unit + golden tests |
| Modify | `server/internal/replay/replay_test.go` | Replay/resume parity |
| Modify | `server/internal/realtime/runner.go` | Persist `OpInventoryRemove` |
| Modify | `server/internal/store/interfaces.go`, `repos.go` | `RemoveInventoryItem` |
| Modify | `server/internal/http/ws_test.go` | Reconnect resume after drop if not covered elsewhere |
| Modify | `shared/rules/worlds.v0.json` | `inventory_lab` |
| Create | `shared/golden/inventory_drop.json` (+ schema) | |
| Create | `client/scripts/inventory_panel.gd` | UI controller |
| Modify | `client/scripts/main.gd` | Mount panel, remove Q, wire intents, `inventory_remove` |
| Modify | `client/scripts/smoke.gd` or `client/tests/test_inventory_ui.gd` | Smoke |
| Modify | `client/tests/test_golden.gd` | Inventory drop golden |
| Create | `tools/bot/scenarios/07_inventory_lab.json` | |
| Modify | `tools/bot/run.py` | Bot actions + `equipped_weapon_def` |
| Modify | `tools/bot/test_protocol.py` | Unit tests |
| Modify | `tools/validate_shared.py` | Golden/schema |
| Modify | `PROGRESS.md` | On completion |

## 6. Acceptance criteria

1. **`make ci` green** including new bot scenario and smoke checks.
2. **`make play`:** `I` opens Diablo-dark panel; pickup sword via world click; double-click or drag
   equips; weapon visible on character; unequip via drag to bag clears visual; drag outside drops
   loot on the ground **beside** the player (visible offset, pickup-able); re-pickup and re-equip works.
3. **Bot `07_inventory_lab.json`** passes pickup → equip → unequip → drop → re-pickup → equip with
   `/state`, reconnect resume, and replay verification.
4. **`Q` does nothing** in manual and autoplay modes.
5. **No client-side inventory authority** — panel state always matches last snapshot + deltas.
6. **Replay/resume** reconstructs unequip/drop outcomes identically.

## 7. Open questions (resolved)

| # | Question | Decision |
|---|----------|----------|
| 1 | Plugin vs custom UI? | **Custom UI (C)** — no addons |
| 2 | Equip gestures? | **Double-click and drag** both |
| 3 | Panel toggle / layout? | **`I` toggle**, weapon slot + bag, game keeps running |
| 4 | Tooltip depth? | **Rich** stats + **empty thumbnails** |
| 5 | Non-equippable items in bag? | **Yes** (`training_badge` visible, no equip) |
| 6 | Unequip / drop? | **Yes** — drag to bag = unequip; outside = drop (new intents) |
| 7 | Bot + Q? | **New bot scenario**; **remove Q**; bot uses `equip_intent` |
| 8 | Aesthetic? | **Diablo-dark** placeholder theme |
| 9 | Drop position? | **Adjacent valid tile** — ring search, collision-free, deterministic; reject `no_drop_space` |

## 8. Verification commands

```bash
make validate-shared
make test-go
make db-up && make server   # terminal 1
make bot                    # includes 07_inventory_lab
make client-smoke
make ci
make play                   # manual UI check
make bot-visual scenario=07_inventory_lab.json  # optional visual pass
```
