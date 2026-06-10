# Spec: `full-equipment-and-belt-hotbar`

Status: Implemented — `make ci` green on 2026-06-07
Branch: `main`
Slice: v28 — full paper-doll equipment, two-hand rules, belt-gated hotbar, and droppable gear templates
Baseline: v26 `character-stats-and-leveling` + v27 `hold-click-controls`
Related:

- [`v13_spec-inventory-ui.md`](v13_spec-inventory-ui.md) — inventory panel, equip/unequip/drop intents
- [`v16_spec-use-consumable.md`](v16_spec-use-consumable.md) — `use_intent` and consumable hotbar UX
- [`v22_spec-character-scoped-persistence.md`](v22_spec-character-scoped-persistence.md) — durable item instances and session-start snapshots
- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) — rolled templates, rarity, and instance payloads
- [`v26_spec-character-stats-and-leveling.md`](v26_spec-character-stats-and-leveling.md) — character progression baseline
- [`v27_spec-hold-click-controls.md`](v27_spec-hold-click-controls.md) — sustained click input (unchanged)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) — authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) — character-scoped progression
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) — plugin adoption checklist (expect **reject** logic plugins; custom UI)
- [`../../PROGRESS.md`](../../PROGRESS.md)

## Resolved implementation notes

- Work landed on the existing `main` checkout; no feature branch was created.
- Hotbar key use is `use_hotbar_intent { slot_index }`; direct bag use remains
  `use_intent { item_instance_id }`.
- Godot inventory logic plugins were rejected for v28. The client uses in-repo `Control` UI and
  server-owned equipment/hotbar authority.

## 1. Purpose

The game has inventory intents, a single weapon equip slot, a client-only 10-slot consumable hotbar,
and one rolled weapon template (`cave_blade`). Players cannot equip armor, jewelry, shields, or
two-handed weapons with correct occupancy rules, and hotbar capacity does not reflect equipped gear.

This slice delivers a **Diablo-style paper-doll** with authoritative multi-slot equipment, **two-hand
occupancy rules**, a **belt-gated consumable hotbar** with **server-persisted layout**, and **at least
one rollable droppable template per equipment category**.

After this slice:

- **`equipped`** on the wire exposes all paper-doll slots; legacy **`weapon`** is removed in favor
  of **`main_hand`** and **`off_hand`**.
- Server enforces slot compatibility, two-hand blocking, and ring/amulet/belt slot rules.
- **Base hotbar capacity is 2** with no belt equipped; equipped belt quality unlocks up to **10**
  usable slots via rolled `hotbar_slots`.
- Hotbar assignments are **character-persisted** and included in session-start snapshots for replay.
- Disabled hotbar slots (index ≥ capacity) **keep their assignment** but are **non-interactive**
  (grayed/blurred client presentation; key press no-ops client-side; `use_hotbar_intent`
  rejects server-side if sent).
- Shield templates may roll **`armor`** and **`block_percent`** — stored and shown in tooltips;
  **not applied in combat** yet.
- New world **`equipment_lab`** and treasure class **`equipment_lab_tc_1`** provide deterministic drops
  covering every equipment category for bot/golden proof.
- Godot inventory UI becomes a paper-doll layout; consumable bar syncs capacity and persisted layout
  from snapshots/deltas.

The proof is **expanded equip model → hotbar persistence → belt capacity → templates/drops → bot
scenario → client UI sync**, not affix grammar, combat mitigation, or production art.

## 2. Non-goals

- No affix grammar, prefix/suffix system, or procedural name generator beyond v23 display-name assembly.
- No unique/set item catalogs, crafting, vendors, stash UI, gold wallet, loot filters, or item comparison UI.
- No authoritative **armor mitigation**, **block chance execution**, dodge, resistances, or shield block
  rolls affecting combat outcomes in v28. Rolled `armor` / `block_percent` are **display + persistence only**.
- No spell system, mana, or offhand abilities beyond passive shield equipment.
- No stack splitting or consumable quantities > 1.
- No production paper-doll art, item icons, or belt/hotbar VFX. Placeholder shapes/colors only (v15 style).
- No Godot inventory **logic** plugins as authority. Custom `Control` UI only.
- No Protobuf migration.
- No depth-scaled treasure classes or dungeon-wide drop table expansion beyond the dedicated lab proof.

## 3. Files to create or modify

```text
docs/specs/v28_spec-full-equipment-and-belt-hotbar.md       - this slice contract
docs/plans/v28_2026-06-07-full-equipment-and-belt-hotbar.md - implementation plan
shared/rules/items.v0.schema.json                           - slot/hand metadata for static items
shared/rules/items.v0.json                                  - migrate weapon slot ids; handedness on weapons
shared/rules/item_templates.v0.schema.json                  - all slot types, handedness, rollable stat keys
shared/rules/item_templates.v0.json                         - one template per equipment category
shared/rules/treasure_classes.v0.json                       - equipment_lab_tc_1
shared/rules/worlds.v0.json                                 - equipment_lab preset
shared/rules/worlds.v0.schema.json                          - if lab layout needs new fields
shared/protocol/messages.v1.schema.json                     - equip/unequip slot enum; hotbar intents
shared/protocol/session_snapshot.v1.schema.json             - equipped slots; hotbar; hotbar_capacity
shared/protocol/state_delta.v1.schema.json                  - hotbar_update; expanded equipped_update
shared/protocol/examples/assign_hotbar_intent.json          - new example
shared/golden/full_equipment.json                           - two-hand, belt capacity, hotbar, shield rolls
shared/golden/full_equipment.v0.schema.json                 - fixture schema
tools/validate_shared.py                                    - slot/hand/hotbar/template drift checks
server/migrations/0006_character_hotbar_and_equipment.sql    - character hotbar rows; multi-slot equipped
server/internal/game/rules.go                               - parse slots, handedness, hotbar stat keys
server/internal/game/sim.go                                 - multi-slot equip, occupancy, hotbar, capacity
server/internal/game/game_test.go                           - equip/hotbar golden tests
server/internal/store/models.go                             - hotbar persistence models
server/internal/store/interfaces.go                         - hotbar repo methods
server/internal/store/repos.go                              - Postgres hotbar + equipped slot persistence
server/internal/store/store_test.go                         - hotbar persistence tests
server/internal/realtime/runner.go                          - persist hotbar mutations by character
server/internal/replay/replay.go                            - session-start hotbar + equipped snapshot
server/internal/inputdecode/inputdecode.go                  - hotbar intent decode
client/scripts/inventory_panel.gd                           - paper-doll layout and slot routing
client/scripts/consumable_bar.gd                            - capacity, disabled styling, server sync
client/scripts/main.gd                                      - hotbar intents; equipped slot migration
client/scripts/equipment_visuals.gd                         - main_hand/off_hand/bow/shield mounts (stretch)
client/scripts/bot_scenario_runner.gd                       - hotbar + multi-slot bot steps
client/tests/test_golden.gd                                 - full_equipment golden checks
client/tests/test_inventory_equipment.gd                    - optional UI model unit tests
tools/bot/run.py                                            - multi-slot equip + hotbar helpers
tools/bot/scenarios/19_full_equipment.json                  - protocol end-to-end proof
tools/bot/scenarios/client/10_full_equipment.json           - Godot client UI proof
PROGRESS.md                                            - lifecycle update when v28 ships
```

Update **all** existing producers/consumers that reference `equipped.weapon` or equip slot `"weapon"`
(including legacy bot scenarios, smoke tests, golden fixtures, and GDScript parsers) in the same slice.

## 4. Equipment model

### 4.1 Paper-doll slots

Authoritative equip slots (fixed set):

| Slot id | Accepts |
|---------|---------|
| `head` | helmets |
| `amulet` | amulets |
| `chest` | body armor |
| `gloves` | gloves |
| `belt` | belts |
| `boots` | boots |
| `ring_left` | rings |
| `ring_right` | rings |
| `main_hand` | one-handed weapons, two-handed weapons, bows |
| `off_hand` | one-handed weapons (legacy offhand weapon — defer), shields |

Bag rows remain non-equipped inventory entries with `slot: ""`.

### 4.2 Handedness and occupancy

Templates and static equippable items declare:

```json
{
  "slot": "main_hand",
  "handedness": "one_handed",
  "occupies_hands": ["main_hand"]
}
```

Allowed `handedness` values:

| Value | `occupies_hands` | Examples |
|-------|------------------|----------|
| `one_handed` | `["main_hand"]` or `["off_hand"]` for shields | 1H sword → `main_hand`; shield → `off_hand` |
| `two_handed` | `["main_hand", "off_hand"]` | 2H sword, bow |

Rules (server `handleEquip`):

1. Item must be equippable and target slot must match template/item `slot` **or** valid alternate for
   rings (`ring_left` / `ring_right` only).
2. Equipping a **two-handed** item into `main_hand` clears `off_hand` (unequip former offhand item to bag).
3. Equipping into `off_hand` while `main_hand` holds a two-handed item is rejected (`hands_blocked`).
4. Equipping a two-handed item while **either** hand holds a blocking item is rejected unless the
   equip atomically clears the other hand (two-handed path clears offhand; shield+1H remains valid).
5. **1H sword + shield** may coexist (`main_hand` weapon + `off_hand` shield).
6. **Bow** uses `handedness: two_handed`, `attack_mode: ranged` — blocks both hands like a 2H sword.

### 4.3 Attack and reach resolution

Replace `equippedWeapon()` / `equipped.weapon` with:

- Primary attack item: equipped instance in `main_hand` with `attack_mode` + damage; if empty, check
  `off_hand` (future dual-wield defer — v28 only checks `main_hand` for weapons).
- **Shield** never supplies attack damage or reach.
- **Bow** in `main_hand` with two-hand occupancy uses existing v12 ranged projectile path.
- Rolled weapon damage continues to prefer instance `rolled_stats` (v23), then static item rules (v8).

Reach, projectile speed, and melee/ranged mode read from the primary attack item as today.

### 4.4 Slot migration from v13–v27

Breaking wire change (coordinated v1 schema extension):

| Before | After |
|--------|-------|
| `equipped.weapon` | `equipped.main_hand` |
| `equip_intent.slot: "weapon"` | `equip_intent.slot: "main_hand"` (and other slots) |
| `unequip_intent.slot: "weapon"` | `unequip_intent.slot: "main_hand"` etc. |
| Item/template `slot: "weapon"` | `slot: "main_hand"` for weapons; shields `off_hand` |
| `cave_blade` template | `slot: "main_hand"`, `handedness: one_handed` |

No backward compatibility shim on the wire. Update bots, client, smoke, goldens, and validation together.

## 5. Item templates and rolled stats

### 5.1 Template catalog (minimum)

Add at least **one rollable template per category** in `item_templates.v0.json`:

| Template id | Slot | Handedness | Notes |
|-------------|------|------------|-------|
| `cave_blade` | `main_hand` | `one_handed` | migrate existing template |
| `cave_greatsword` | `main_hand` | `two_handed` | melee 2H |
| `cave_bow` | `main_hand` | `two_handed` | `attack_mode: ranged` |
| `cave_shield` | `off_hand` | `one_handed` | rolls `armor`, `block_percent` |
| `cave_helm` | `head` | — | rolls `armor` |
| `cave_mail` | `chest` | — | rolls `armor` |
| `cave_gloves` | `gloves` | — | rolls `armor` |
| `cave_belt` | `belt` | — | rolls `hotbar_slots`, optional `armor` |
| `cave_boots` | `boots` | — | rolls `armor` |
| `cave_ring` | `ring_left` or either ring | — | rolls `max_hp` or `armor` |
| `cave_amulet` | `amulet` | — | rolls `max_hp` |

Ring templates declare `slot: "ring"` as a logical type; server accepts equip into `ring_left` or
`ring_right` when template slot is `ring`.

Non-weapon armor templates omit `attack_mode` / `reach` / damage base stats where not applicable.
Schema must allow category-specific `base_stats` / `rollable_stats` per item type.

### 5.2 Rollable stat keys (v28 catalog)

Extend v23 roll vocabulary:

| Stat key | Used on | Combat in v28 |
|----------|---------|---------------|
| `damage_min`, `damage_max` | weapons | **Yes** (authoritative) |
| `max_hp` | any | Display only |
| `armor` | armor, shield | Display only |
| `block_percent` | shield | Display only |
| `hotbar_slots` | belt | **Yes** for capacity (authoritative) |

`hotbar_slots` rolls an integer in `[3, 10]` on belt templates (exact range pinned in golden).
Capacity formula:

```text
hotbar_capacity = clamp(equipped_belt.hotbar_slots, 2, 10)  // if belt equipped
hotbar_capacity = 2                                         // if no belt
```

Belt with no rolled `hotbar_slots` uses template `base_stats.hotbar_slots` minimum (default 3).

### 5.3 Droppables

New treasure class `equipment_lab_tc_1` with weighted entries covering **every template id above**
(at least one entry per category). World **`equipment_lab`** spawns a deterministic source (chest,
monster, or preset loot pile) using pinned seed in golden/bot so CI can collect each category without
RNG flakiness.

Existing dungeon TCs (`dungeon_mob_tc_1`, etc.) remain unchanged in v28.

## 6. Hotbar persistence and capacity

### 6.1 Snapshot shape

Add to `session_snapshot` (required fields):

```json
{
  "hotbar_capacity": 2,
  "hotbar": [
    { "slot_index": 0, "item_instance_id": "1010" },
    { "slot_index": 1, "item_instance_id": null },
    { "slot_index": 2, "item_instance_id": "1011" },
    { "slot_index": 3, "item_instance_id": null },
    "... slots 4-9 ..."
  ]
}
```

- Fixed **10** entries, indices `0`–`9` (keys `1`–`0` on client).
- `hotbar_capacity` is server-computed each tick from equipped belt; included in snapshot for client UX.
- Assignment persists even when `slot_index >= hotbar_capacity` (disabled slot).

### 6.2 Intents

**Assign hotbar slot**

```json
{
  "type": "assign_hotbar_intent",
  "payload": {
    "slot_index": 2,
    "item_instance_id": "1011"
  }
}
```

- `item_instance_id: null` clears the slot.
- Reject `invalid_payload` if index ∉ `[0, 9]`.
- Reject `not_in_inventory` if instance not in bag.
- Reject `not_consumable` if item category ≠ consumable.
- Reject `player_dead`.
- **Do not reject** disabled slots — player may pre-assign potions before finding a better belt.
- Persist to character; emit `hotbar_update` delta; ack intent.

**Use consumable directly** — unchanged `use_intent { item_instance_id }`:

- On successful use, remove inventory row **and** clear any hotbar entry referencing that instance.

**Use hotbar slot** — new `use_hotbar_intent { slot_index }`:

- Reject `invalid_payload` if index ∉ `[0, 9]`.
- Reject `hotbar_slot_disabled` when `slot_index >= hotbar_capacity`.
- Reject `slot_empty`, `not_in_inventory`, or `not_consumable` when the assigned item is missing or invalid.
- On successful use, consume through the same authoritative path as direct `use_intent`, remove inventory row,
  and clear any hotbar entry referencing that instance.

### 6.3 Persistence

- Durable table `character_hotbar_slots (character_id, slot_index, item_instance_id NULL)` or JSON
  column on character row — implementation choice; contract is character-scoped durability matching v22.
- Session-start snapshot freezes hotbar layout for replay (same pattern as items/waypoints/progression).
- Same-session reconnect reconstructs from inputs + session-start snapshot; hotbar mutations in session
  inputs replay correctly.

### 6.4 Disabled slot behavior (Q-11)

When `slot_index >= hotbar_capacity`:

- Assignment **remains** in authoritative `hotbar` array.
- Client renders slot **grayed/blurred** (reduced opacity + desaturate; no new art assets).
- Hotkey press for that slot: **no intent sent** (client gate).
- Server rejects `use_hotbar_intent { slot_index }` if sent for a disabled slot: reason
  `hotbar_slot_disabled`.
- When belt raises capacity, previously disabled assignments become active without re-assign.

## 7. Client presentation

### 7.1 Inventory paper-doll (`inventory_panel.gd`)

Replace single weapon column with Diablo-style layout:

```text
        [ head ]
[ ring L ] [ amulet ] [ ring R ]
        [ chest ]
[ main_hand ] [ off_hand ]
[ gloves ] [ belt ] [ boots ]
        [ bag grid ]
```

- Toggle **`I`**; same dark theme as v13.
- Each slot accepts drag/drop from bag when item slot matches; double-click equips from bag.
- Drag equipped item to bag area → `unequip_intent`; drag outside → `drop_intent`.
- Tooltips show template/display name, rarity, rolled stats including `armor`, `block_percent`,
  `hotbar_slots` where present.
- Placeholder icons per category from `item_presentations.v0.json` extensions (minimal new entries).

Plugin checklist outcome (plan gate): **reject** GLoot/Wyvernbox logic; optional layout reference only.

### 7.2 Consumable bar (`consumable_bar.gd`)

- Sync `hotbar` + `hotbar_capacity` from snapshot/deltas — **not** local-only state.
- Drag from bag calls `assign_hotbar_intent` (not local assign).
- Keys `1`–`9`, `0` map to indices `0`–`9`:
  - Enabled slot with assignment → send `use_hotbar_intent { slot_index }`.
  - Disabled slot → **no-op** (visual gray/blur).
- Clearing assignment sends `assign_hotbar_intent` with null id.
- XP bar from v26 stays below hotbar; layout must accommodate disabled slot styling.

### 7.3 Equipment visuals (stretch)

- Migrate weapon mount from `equipped.weapon` to `equipped.main_hand`.
- Optional v28 stretch: shield on offhand socket, bow on two-hand mount — may defer to placeholder
  icons only if GLB pipeline not ready; spec minimum is **inventory + ground loot presentation**.

## 8. Bot and golden proof

### 8.1 Golden `shared/golden/full_equipment.json`

Pin on a fixed seed/world (e.g. `equipment_lab`):

1. Equip 1H sword + shield simultaneously.
2. Equip 2H sword → offhand cleared.
3. Equip bow → both hands occupied.
4. Pinned equipment-lab belt with rolled `hotbar_slots: 10` → `hotbar_capacity: 10`.
5. Assign potions to slots 0, 2, 5; unequip belt → capacity 2; slots 2 and 5 disabled but assigned.
6. Re-equip belt → slots 2 and 5 enabled again.
7. Shield roll includes `armor` and `block_percent` in instance payload.

Go sim test + `test_golden.gd` data drift checks.

### 8.2 Protocol bot `19_full_equipment.json`

Flow (high level):

1. Enter `equipment_lab` with pinned seed.
2. Collect and equip one item per slot category from TC drops.
3. Assert 1H + shield, then 2H swap, then bow.
4. Pick up belt + potions; assign hotbar; assert capacity changes on belt unequip/re-equip.
5. Use potion from enabled hotbar slot via bot helper.
6. `/state`, reconnect resume, replay, fresh-session persistence for equipped + hotbar.

### 8.3 Client bot `10_full_equipment.json`

- Open paper-doll; drag equip to each slot type.
- Assign hotbar via drag; assert disabled styling when capacity < slot index.
- Press key on disabled slot → no HP change.
- Press key on enabled slot → heal/use succeeds.

Update existing client scenarios (`06_use_potion_hotbar.json`, etc.) for `main_hand` migration and
server-backed hotbar intents.

## 9. Architecture flow

```text
pick up rolled loot
  -> inventory_add with slot metadata + rolled_stats
  -> equip_intent { slot: main_hand | off_hand | head | ... }
  -> Sim validates slot/hand occupancy; may unequip conflicting hands
  -> equipped_update + inventory_update deltas
  -> runner persists character item location/slot/equipped
  -> hotbar_capacity recomputed from equipped belt rolled hotbar_slots

assign_hotbar_intent
  -> validate consumable in bag
  -> persist character_hotbar_slots
  -> hotbar_update delta

use_hotbar_intent (from enabled hotbar key)
  -> resolve slot assignment server-side
  -> validate capacity, consumable, HP, alive
  -> apply heal; inventory_remove; clear hotbar references
  -> player_healed + item_used events

fresh session / replay
  -> load session-start equipped + hotbar snapshot
  -> apply session inputs in order
```

Determinism: equip order, two-hand clears, hotbar assigns, and capacity changes must reproduce from
seed + ordered inputs. No wall-clock or unseeded RNG in `game/`.

## 10. Acceptance criteria

1. `make validate-shared` validates expanded slot/hand schemas, templates, TC entries, and golden drift.
2. Wire format exposes full `equipped` slot map; **`weapon` is gone** from schemas and examples.
3. Server enforces two-hand, bow, and shield occupancy rules with stable reject reasons.
4. `hotbar_capacity` is **2** without belt; belt rolled `hotbar_slots` raises cap to max **10**.
5. Hotbar layout persists character-scoped across fresh sessions and session-start replay snapshots.
6. Disabled hotbar slots retain assignments, appear grayed/blurred client-side, and ignore hotkey use.
7. Shield instances may roll and display `armor` and `block_percent` without changing combat outcomes.
8. At least one template per equipment category is droppable from `equipment_lab_tc_1`.
9. Godot paper-doll equips/unequips/drops through existing + expanded intents without local authority.
10. Bot `19_full_equipment.json` and client bot `10_full_equipment.json` pass with reconnect + replay.
11. All legacy scenarios updated for `main_hand`; `make ci` green.

## 11. Testing plan

1. `make validate-shared`
2. `cd server && go test ./internal/game/... -run 'Equip|Hotbar|Hand|Occup|Belt|FullEquipment'`
3. `cd server && go test ./internal/store/... -run 'Hotbar|Equipment'`
4. `cd server && go test ./internal/http/... ./internal/replay/... -run 'Hotbar|Equip|Replay'`
5. `make client-unit`
6. `make bot` — includes `19_full_equipment.json`
7. `make bot-client` — includes `10_full_equipment.json`
8. `make ci`
9. Manual: `make play` — equip full paper-doll, swap 1H/2H/bow, test belt hotbar expand/collapse,
   restart Continue, confirm gear + hotbar layout persist.

## 12. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Migrate `weapon` → `main_hand` + `off_hand` | Matches ARPG hand semantics and user direction |
| 2 | `ring_left` / `ring_right` distinct slots | Diablo paper-doll clarity |
| 3 | Hotbar layout is server-persisted | User requirement; aligns with v22 character ownership |
| 4 | Base capacity 2 without belt | User requirement |
| 5 | Disabled slots keep assignments, grayed UI, no hotkey effect | Q-11; rewards planning ahead for better belt |
| 6 | v23 roll model only; no affix grammar | Keeps slice bounded |
| 7 | Shield `armor` / `block_percent` display-only | Prepares combat slice without scope creep |
| 8 | `equipment_lab` + dedicated TC for CI | Deterministic proof without changing dungeon economy |
| 9 | Custom inventory UI; reject logic plugins | ADR-0001 + v13 precedent |
| 10 | Coordinate breaking equip schema in one slice | Avoids half-migrated wire state |

## 13. Open questions

None — all slice questions resolved during `/next` discovery (2026-06-07).
