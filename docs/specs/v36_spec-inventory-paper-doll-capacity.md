# Spec: `inventory-paper-doll-capacity`

Status: Draft
Branch: `main`
Slice: v36 - inventory paper-doll tuneup and item-gated bag capacity
Baseline: v35 `boss-floor-gate`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - plugin adoption checklist for inventory UI presentation work
- [`v13_spec-inventory-ui.md`](v13_spec-inventory-ui.md) - inventory panel, equip/unequip/drop intents
- [`v22_spec-character-scoped-persistence.md`](v22_spec-character-scoped-persistence.md) - durable character inventory
- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) - rolled item stats and templates
- [`v26_spec-character-stats-and-leveling.md`](v26_spec-character-stats-and-leveling.md) - character progression and derived stat presentation
- [`v28_spec-full-equipment-and-belt-hotbar.md`](v28_spec-full-equipment-and-belt-hotbar.md) - authoritative equipment slots and hotbar capacity model

## 1. Purpose

The current inventory window is functionally useful but visually weak. Equipment slots render as a
plain two-column list, the bag is a scrolling 4-column list, and the panel does not communicate a
paper-doll character layout. This makes loot comparison and equipment management feel like a debug
tool instead of a core ARPG screen.

This slice improves the human inventory experience while preserving the authoritative inventory
boundary:

- The equipment area becomes a paper-doll layout with slots positioned around the character body:
  head, amulet, chest, gloves/hands, belt, boots/feet, rings, main hand, and off hand.
- Empty equipment slots render as simple gray blocks. No custom empty-slot illustrations are
  required.
- A dynamic 2D character presentation appears behind or between the equipment slots. The first
  implementation may use a muted silhouette, viewport capture, or lightweight 2D rendering of the
  current character model, but it must be structured so future model/gear improvements can improve
  the paper-doll backdrop without redesigning the panel.
- The bag becomes a fixed-slot grid with **5 columns** and **15 base slots**.
- Bag capacity is server-owned and derived from a new character/equipment substat:
  `inventory_rows`.
- Base `inventory_rows` is **3**. Each additional row grants **5** more bag slots.
- `inventory_capacity = inventory_rows * 5`.
- Equipped items use their equipment slots and do **not** consume bag capacity.
- Hotbar-assigned items use hotbar slots and do **not** consume bag capacity while assigned.
- In this slice, only **items** may increase `inventory_rows`. Base stat allocation must not
  increase inventory space. Future skill-based row increases are explicitly deferred.

The proof is: item-derived inventory rows -> authoritative capacity/rejection -> protocol snapshot
and delta sync -> 5-column Godot bag grid -> paper-doll equipment layout -> bot/client proof.

## 2. Non-goals

- No stash, vendors, crafting, item filters, sorting, or item comparison overlay.
- No multi-cell Diablo/Tetris inventory item footprints. Every item remains one inventory entry.
- No client-side inventory authority, local save logic, or client-only capacity calculation.
- No new equipment slot ids beyond the v28 authoritative slot set.
- No production paper-doll art, item icon pack, VFX, animation, or audio.
- No custom empty-slot drawings. Empty equipment slots are gray blocks with labels/tooltips only.
- No stat allocation path for inventory space. `str`, `dex`, `vit`, and `magic` do not affect
  `inventory_rows`.
- No skill tree implementation. Future skills may add rows, but v36 only reserves a clean derived
  stat path.
- No Godot inventory logic plugin as authority. The plan may adopt or borrow presentation patterns
  only after recording the plugin checklist result.
- No Protobuf migration.

## 3. Files to create or modify

```text
docs/specs/v36_spec-inventory-paper-doll-capacity.md       - this slice contract
docs/plans/v36_<YYYY-MM-DD>-inventory-paper-doll-capacity.md - implementation plan
PROGRESS.md                                           - lifecycle update when v36 ships

shared/rules/character_progression.v0.json                 - base inventory_rows if stored with derived stats
shared/rules/character_progression.v0.schema.json          - derived stat schema update if needed
shared/rules/item_templates.v0.json                        - deterministic item/template source for +inventory_rows
shared/rules/item_templates.v0.schema.json                 - allow inventory_rows as an item stat
shared/golden/inventory_capacity.json                      - base capacity, item row bonus, full-bag rejection fixture
shared/golden/inventory_capacity.v0.schema.json            - fixture schema
shared/protocol/session_snapshot.v*.schema.json            - inventory_capacity / inventory_rows view
shared/protocol/state_delta.v*.schema.json                 - inventory capacity/progression update if needed
shared/protocol/examples/session_snapshot.json             - capacity example
shared/protocol/examples/state_delta.json                  - capacity update or rejection example
tools/validate_shared.py                                   - inventory_rows stat validation and golden drift checks

server/internal/game/rules.go                              - item stat parser/validator for inventory_rows
server/internal/game/sim.go                                - capacity derivation and pickup rejection
server/internal/game/types.go                              - protocol view fields if needed
server/internal/game/game_test.go                          - capacity, equip/unequip, item bonus, and rejection tests
server/internal/replay/*                                   - replay parity if snapshot/delta shape changes
server/internal/store/*                                    - persistence updates only if durable state shape changes

client/scripts/inventory_panel.gd                          - paper-doll layout, 5-column bag grid, capacity rows
client/scripts/main.gd                                     - snapshot/delta capacity sync
client/scripts/bot_controller.gd                           - client proof helpers if needed
client/scripts/bot_scenario_runner.gd                      - assertions for bag grid/capacity layout
client/scripts/smoke.gd or client/tests/*                  - focused UI model proof

tools/bot/run.py                                           - full-bag and capacity helper steps if needed
tools/bot/scenarios/25_inventory_capacity_and_paper_doll.json - protocol proof
tools/bot/scenarios/client/14_inventory_paper_doll.json    - Godot client visual/model proof if split
```

Protocol note: this is an additive protocol/schema change unless the plan proves existing
`character_progression.derived_stats` can carry the capacity fields without a wire bump. The plan
must explicitly decide whether to add `inventory_rows` / `inventory_capacity` under
`character_progression.derived_stats`, as top-level snapshot fields, or both. Every producer,
consumer, example, replay path, and client parser must be updated together.

## 4. Inventory capacity model

### 4.1 Definitions

The authoritative capacity contract is:

```text
base_inventory_rows = 3
inventory_rows = base_inventory_rows + item_inventory_row_bonus + future_skill_bonus
inventory_capacity = inventory_rows * 5
```

For v36, `future_skill_bonus` is always zero.

`inventory_rows` is a derived/equipment substat, not an allocatable base character stat. Allocating
`str`, `dex`, `vit`, or `magic` must never change bag capacity.

### 4.2 What counts against capacity

Bag capacity counts only inventory entries that are not equipped and not assigned to a hotbar slot.

Rules:

1. Equipped items occupy their equipment slots and do not count against bag capacity.
2. Unequipping an item requires free bag capacity unless the item is being swapped atomically with
   another item that leaves the bag in the same operation.
3. Hotbar-assigned items occupy their assigned hotbar slots and do not count against bag capacity.
4. Consumables, currency-like items, equipment, and quest/test items all occupy one bag slot when
   they are unequipped and not assigned to a hotbar slot.
5. Pickup into a full bag rejects with a stable reason, expected default:
   `inventory_full`.

### 4.3 Item-derived row bonus

At least one deterministic item/template must prove `inventory_rows`.

Acceptable first implementation:

- Add `inventory_rows` to allowed rolled/base item stats.
- Create or update one deterministic lab item that grants `+1 inventory_rows`.
- Equipping that item changes capacity from 15 to 20.
- Removing/unequipping that item reduces capacity back to 15 if the resulting bag count is legal.

If unequipping or removing the row-granting item would shrink capacity below current bag count, the
server must either reject the action with a stable reason or define an explicit overflow behavior in
the plan. Default: reject with `inventory_full` or `capacity_would_overflow` before mutating state.

### 4.4 Persistence and replay

Inventory contents, equipment, and item rolled stats are already character/session authoritative.
The capacity result should be derived from those persisted facts instead of storing a second mutable
capacity value.

Replay must reconstruct the same capacity after the same ordered inputs. Capacity checks must be
deterministic and must not depend on client viewport, wall-clock time, or unordered map iteration.

## 5. Client presentation

### 5.1 Paper-doll layout

The inventory panel should use a readable ARPG paper-doll composition:

```text
          [head]       [amulet]

 [main_hand]  character presentation  [off_hand]

      [gloves] [chest] [ring_left]
      [belt]   [boots] [ring_right]
```

The exact positions may change in the plan, but acceptance requires that slots are no longer a
plain vertical/two-column list. Slot placement must visually correspond to the body area or common
ARPG expectations.

Empty equipment slots:

- Gray block frame.
- Slot label or tooltip.
- No placeholder weapon/shield/body drawings required.

### 5.2 Dynamic character presentation

The paper-doll backdrop should be dynamic enough to benefit from future model or gear upgrades.
The implementation may choose one of these approaches:

- Render the current `CharacterVisual` into a `SubViewport` / `TextureRect` inside the panel.
- Use a lightweight duplicated model/preview node dedicated to UI display.
- Use a simple silhouette now, but keep the API/path named around `character_paper_doll` or
  equivalent so a model-backed preview can replace it without changing inventory state logic.

The presentation must be client-only. It must not affect equipment authority, hitboxes, combat, or
item stats.

### 5.3 Bag grid

The bag must render as:

- 5 columns.
- 3 base rows.
- 15 base slots visible when capacity is 15.
- Additional item/skill rows displayed as full 5-slot rows.
- Locked or unavailable rows may be hidden or shown disabled; the plan must pick one behavior.
  Default: show only currently available slots, plus no fake `+` drop slot that alters grid math.

The client must not optimistically add items beyond capacity. It waits for server snapshot/delta
state and shows rejection feedback when the server rejects pickup or unequip due to capacity.

## 6. Bot and test proof

### 6.1 Protocol bot proof

Add a protocol scenario such as:

```text
tools/bot/scenarios/25_inventory_capacity_and_paper_doll.json
```

Required assertions:

- Fresh character/session reports `inventory_rows == 3` and `inventory_capacity == 15`.
- Filling the bag to 15 unequipped entries succeeds.
- The next pickup rejects with `inventory_full` or the chosen stable rejection reason.
- Equipping an item that grants `+1 inventory_rows` increases capacity to 20.
- After capacity increases, five more unequipped bag entries can be held.
- Replay and reconnect resume report the same inventory rows, capacity, inventory count, and
  equipped state.

If a deterministic fixture world is needed, create an inventory-capacity lab world or use existing
test hooks in the bot. The scenario should avoid combat unless loot drops are the simplest way to
obtain the proof items.

### 6.2 Client proof

Add a Godot/client proof that can run headless and verify the UI model without relying on pixel
inspection:

- Inventory panel debug state reports equipment slot positions/ids for the paper-doll layout.
- Bag grid reports `columns == 5`.
- Base visible slot count is 15.
- Capacity 20 produces 4 rows / 20 available slots.
- Empty equipment slots are present and styled as inactive/empty gray blocks.
- The dynamic character presentation node exists or the fallback silhouette path is present.

Optional visual pass: run `make bot-visual` with the client scenario to inspect the layout manually.
The CI proof must remain data-driven.

### 6.3 Unit and golden tests

Required server tests:

- Base capacity from default character/equipment is 15.
- Item-derived `inventory_rows` changes capacity by 5 per row.
- Pickup rejects when bag is full.
- Equipped items do not count against bag capacity.
- Hotbar-assigned items do not count against bag capacity.
- Unequip rejects when no free bag slot exists.
- Stat allocation does not change `inventory_rows` or `inventory_capacity`.
- Replay reconstructs the same capacity and rejection behavior.

Required shared/golden proof:

- `shared/golden/inventory_capacity.json` covers base capacity, +1 row item capacity, and full-bag
  rejection.

## 7. Acceptance checklist

- [ ] Equipment UI uses a paper-doll layout around a character presentation instead of a plain
  two-column equipment list.
- [ ] Empty equipment slots render as gray blocks with stable slot identity.
- [ ] Bag grid uses 5 columns and 15 base slots.
- [ ] Server exposes authoritative `inventory_rows` and `inventory_capacity`, or an equivalent
  protocol view explicitly chosen by the plan.
- [ ] Base capacity is 15.
- [ ] Equipping an item with `+1 inventory_rows` increases capacity to 20.
- [ ] Allocating base character stats does not increase inventory capacity.
- [ ] Equipped items do not consume bag capacity.
- [ ] Hotbar-assigned items do not consume bag capacity.
- [ ] Pickup/unequip into a full bag rejects with a stable reason and does not mutate inventory.
- [ ] Protocol bot scenario proves capacity, rejection, capacity increase, reconnect, and replay.
- [ ] Godot client proof verifies paper-doll slot ids, 5-column bag grid, and capacity rows.
- [ ] `make ci` is green when the slice ships.

## 8. Open questions for the plan

| # | Question | Default |
|---|----------|---------|
| Q-1 | Should capacity fields live in `character_progression.derived_stats`, top-level snapshot fields, or both? | Put display values in `character_progression.derived_stats` and expose a direct `inventory_capacity` if client code becomes simpler. |
| Q-2 | What item proves `+1 inventory_rows`? | Add a deterministic lab/template item with base or rolled `inventory_rows: 1`. |
| Q-3 | When a row-granting item is removed and capacity would shrink below bag count, reject or overflow? | Reject before mutation. |
| Q-4 | Should locked/unavailable bag rows be hidden or shown disabled? | Hide unavailable rows; show only available slots. |
| Q-5 | Which dynamic character preview approach is cheapest in Godot headless CI? | Start with a structured fallback silhouette/preview node; upgrade to `SubViewport` only if reliable. |

## 9. ADR alignment

- ADR-0001 D2 remains intact: the server owns inventory capacity, pickup rejection, and equipment
  state.
- ADR-0001 D6 remains intact: `inventory_rows` is a shared rules/data stat, not a client-only
  setting.
- ADR-0001 D8 remains intact: capacity behavior must replay deterministically from seed and inputs.
- The Godot plugins checklist must be recorded in the implementation plan. Inventory logic plugins
  remain rejected as authority; UI patterns may be borrowed if the checklist supports it.
