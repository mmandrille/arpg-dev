# v36 Plan - Inventory Paper-Doll Capacity

Status: Ready for implementation
Goal: Replace the current inventory panel with a paper-doll equipment layout and server-owned 5-column bag capacity derived from item-granted `inventory_rows`.
Architecture: Reuse the v28 authoritative equipment/hotbar model and add inventory capacity as a derived server value. The Go sim remains the only authority for pickup, unequip, capacity shrink, and rejection outcomes. The Godot client renders a paper-doll panel and capacity rows from snapshots/deltas only. Bot and replay proof cover the new capacity contract, while client bot/unit proof covers the UI model without pixel assertions.
Tech stack: Shared JSON rules/schemas/goldens, Go deterministic sim, JSON protocol schemas, Python protocol bot, Godot GDScript UI/client bot.

## Baseline and shortcut decision

Baseline is v35 `boss-floor-gate` on `main`, building on v28 full equipment/hotbar and v26 character progression. Current `InventoryPanel` already owns drag/drop slot routing and item drawing, while `Sim.hotbarCapacity()` is the closest server pattern for item-derived capacity.

Godot plugin adoption decision for this slice: **reject inventory logic plugins** as authority. **Borrow pattern only** from the inventory UI resources listed in `docs/researchs/godot-plugins-and-shortcuts.md` if useful for paper-doll composition. A full addon is not justified because the current in-repo panel is already wired to server snapshots/intents and the slice needs layout/capacity, not a new inventory logic framework.

Plan decisions:

- Expose direct snapshot fields `inventory_rows` and `inventory_capacity` for simple client/bot consumption.
- Also include `inventory_rows` / `inventory_capacity` in `character_progression.derived_stats` or stat breakdowns only if the existing progression panel needs them for display.
- Bag capacity counts inventory entries that are **not equipped** and **not assigned to a hotbar slot**.
- Base rows are `3`, with capacity `rows * 5`.
- Add `inventory_rows` as an allowed item base/rolled stat, with one deterministic lab item granting `+1`.
- If capacity shrink would overflow the bag, reject before mutation with `capacity_would_overflow`.
- Pickup into a full bag rejects with `inventory_full`.
- Hide unavailable bag rows; render only available 5-slot rows.
- For the paper-doll backdrop, start with a dedicated `character_paper_doll` preview/silhouette node that is structured for a future model-backed `SubViewport`, but avoid SubViewport work unless it proves reliable in headless client tests.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.schema.json` | Allow `inventory_rows` in item stats. |
| Modify | `shared/rules/item_templates.v0.json` | Add deterministic row-granting item/template. |
| Create | `shared/golden/inventory_capacity.json` | Base capacity, +1 row, full-bag rejection contract. |
| Create | `shared/golden/inventory_capacity.v0.schema.json` | Fixture schema. |
| Modify | `shared/protocol/session_snapshot.v2.schema.json` | Add `inventory_rows` and `inventory_capacity`. |
| Modify | `shared/protocol/state_delta.v2.schema.json` | Add inventory capacity fields on relevant changes if needed. |
| Modify | `shared/protocol/examples/session_snapshot.json` | Snapshot example with capacity fields. |
| Modify | `shared/protocol/examples/state_delta.json` | Delta/rejection example for capacity updates. |
| Modify | `tools/validate_shared.py` | Validate item stat and capacity golden drift. |
| Modify | `server/internal/game/rules.go` | Parse/validate `inventory_rows`. |
| Modify | `server/internal/game/types.go` | Add capacity fields to protocol views. |
| Modify | `server/internal/game/sim.go` | Derive capacity, count bag entries, gate pickup/unequip/capacity shrink. |
| Modify | `server/internal/game/game_test.go` | Capacity, hotbar exemption, equip/unequip, golden, replay-oriented tests. |
| Modify | `server/internal/replay/*` | Update replay expectations only if snapshot/delta shape requires it. |
| Modify | `tools/bot/run.py` | Track/assert inventory rows/capacity and new rejection flow. |
| Modify | `tools/bot/test_protocol.py` | Unit coverage for new bot assertions/state. |
| Create | `tools/bot/scenarios/25_inventory_capacity_and_paper_doll.json` | Protocol proof. |
| Modify | `client/scripts/inventory_panel.gd` | Paper-doll layout, 5-column bag, capacity slots, debug state. |
| Modify | `client/scripts/main.gd` | Sync capacity fields into inventory panel. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Client assertions for paper-doll/capacity UI model. |
| Modify | `client/scripts/bot_controller.gd` | Client bot helpers if needed. |
| Create | `tools/bot/scenarios/client/13_inventory_paper_doll.json` | Client UI proof. |
| Modify | `client/scripts/smoke.gd` or `client/tests/*` | Headless UI model checks. |
| Modify | `docs/PROGRESS.md` | Lifecycle update when slice ships. |

## Task 1 - Shared contracts and golden fixture

Files:
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Create: `shared/golden/inventory_capacity.json`
- Create: `shared/golden/inventory_capacity.v0.schema.json`
- Modify: `shared/protocol/session_snapshot.v2.schema.json`
- Modify: `shared/protocol/state_delta.v2.schema.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add `inventory_rows` to item template `base_stats` / `rollable_stats` schemas with integer bounds `0..20`; keep invalid stats rejected.
```bash
make validate-shared
```

- [x] Step 1.2: Add one deterministic equipment/template source for `inventory_rows: 1`. Prefer an existing equipment category where equipping the item proves capacity without adding new slot ids.
```bash
make validate-shared
```

- [x] Step 1.3: Add `inventory_rows` and `inventory_capacity` to session snapshots; add delta fields only on `equipped_update`, `hotbar_update`, and/or `character_progression_update` if the plan implementation needs live capacity changes without a full snapshot.
```bash
make validate-shared
```

- [x] Step 1.4: Add `shared/golden/inventory_capacity.json` covering base rows `3`, base capacity `15`, +1 row item capacity `20`, hotbar/equipped exemptions, and the chosen rejection reasons.
```bash
make validate-shared
```

## Task 2 - Go sim capacity authority

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/replay/*` if needed

- [x] Step 2.1: Add helpers equivalent to `hotbarCapacity()`: `inventoryRows()`, `inventoryCapacity()`, and `bagOccupancyCount()`.
```bash
cd server && go test ./internal/game/... -run TestInventoryCapacity
```

- [x] Step 2.2: Count only items where `equipped == false` and the instance id is not assigned in any hotbar slot.
```bash
cd server && go test ./internal/game/... -run TestInventoryCapacity
```

- [x] Step 2.3: Gate `pickUpTarget` before removing the loot entity. A rejected full-bag pickup must emit `intent_rejected` and leave world loot plus inventory unchanged.
```bash
cd server && go test ./internal/game/... -run TestInventoryCapacity
```

- [x] Step 2.4: Gate `handleUnequip` before mutating the item/slot when the unequipped item would exceed capacity.
```bash
cd server && go test ./internal/game/... -run TestInventoryCapacity
```

- [x] Step 2.5: Gate capacity-shrinking operations, including unequipping/removing the row-granting item, with `capacity_would_overflow` before mutation.
```bash
cd server && go test ./internal/game/... -run TestInventoryCapacity
```

- [x] Step 2.6: Publish capacity in snapshots and relevant state deltas so client, bot, reconnect, and replay receive identical values.
```bash
cd server && go test ./internal/game/... -run 'TestInventoryCapacity|Test.*Snapshot|Test.*Replay'
```

- [x] Step 2.7: Add tests proving base capacity `15`, +1 row capacity `20`, equipped exemption, hotbar exemption, stat allocation does not change capacity, pickup rejection, unequip rejection, and golden fixture parity.
```bash
cd server && go test ./internal/game/...
```

## Task 3 - Protocol bot proof

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Create: `tools/bot/scenarios/25_inventory_capacity_and_paper_doll.json`

- [x] Step 3.1: Extend runtime state, snapshot parsing, delta parsing, reconnect assertions, and structured assertions with `inventory_rows` and `inventory_capacity`.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -v
```

- [x] Step 3.2: Add bot helpers to fill capacity deterministically and to expect `inventory_full` / `capacity_would_overflow` rejections.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -v
```

- [x] Step 3.3: Add `25_inventory_capacity_and_paper_doll.json` proving base capacity, full-bag rejection, row item equip to capacity 20, five more bag entries, reconnect, and replay.
```bash
make bot
```

## Task 4 - Godot inventory panel model and layout

Files:
- Modify: `client/scripts/inventory_panel.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/smoke.gd` or `client/tests/*`

- [x] Step 4.1: Replace the two-column equipment list with a paper-doll layout using the existing authoritative slot ids: `head`, `amulet`, `chest`, `gloves`, `belt`, `boots`, `ring_left`, `ring_right`, `main_hand`, `off_hand`.
```bash
make client-unit
```

- [x] Step 4.2: Add a dedicated dynamic-preview placeholder node, named around `character_paper_doll`, behind/between the slots. It may render a simple silhouette for v36, but must keep rendering isolated from inventory authority.
```bash
make client-unit
```

- [x] Step 4.3: Change the bag grid to 5 columns and render exactly `inventory_capacity` available cells, with 15 at base capacity and 20 after +1 row.
```bash
make client-unit
```

- [x] Step 4.4: Remove the fake bag `+` drop slot from grid math; preserve drag-to-bag/unequip and drag-outside/drop behavior through explicit drop areas or panel logic that does not alter capacity cells.
```bash
make client-unit
```

- [x] Step 4.5: Extend `InventoryPanel.get_debug_state()` with paper-doll slot ids/positions, `bag_columns`, `available_slot_count`, `inventory_rows`, `inventory_capacity`, empty-slot style markers, and paper-doll preview status.
```bash
make client-unit
```

## Task 5 - Client bot scenario

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_controller.gd` if needed
- Create: `tools/bot/scenarios/client/13_inventory_paper_doll.json`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 5.1: Add client scenario assertions for `assert_inventory_capacity`, `assert_bag_grid`, and `assert_paper_doll_layout`.
```bash
make client-unit
```

- [x] Step 5.2: Add `13_inventory_paper_doll.json` to open the panel, assert 5 columns / 15 slots, assert all equipment slot ids exist in paper-doll layout, equip the row item, and assert 20 slots.
```bash
HEADLESS=1 make bot-client scenario=13_inventory_paper_doll.json
```

- [x] Step 5.3: Keep existing client scenarios green after layout changes, especially inventory drag/drop and full equipment.
```bash
make client-unit
make client-smoke
```

## Task 6 - Bot scenarios

Files:
- Create: `tools/bot/scenarios/25_inventory_capacity_and_paper_doll.json`
- Create: `tools/bot/scenarios/client/13_inventory_paper_doll.json`
- Modify: existing scenarios only if capacity limits require fixtures to assign/equip items before collecting more loot

- [x] Step 6.1: Ensure protocol scenario `25_inventory_capacity_and_paper_doll.json` is selected by default `make bot`.
```bash
make bot
```

- [x] Step 6.2: Ensure client scenario `13_inventory_paper_doll.json` validates through the client scenario loader.
```bash
make client-unit
```

- [x] Step 6.3: Run focused visual/client proof.
```bash
HEADLESS=1 make bot-client scenario=13_inventory_paper_doll.json
```

## Task 7 - Lifecycle docs and CI

Files:
- Modify: `docs/PROGRESS.md`
- Keep: `docs/specs/v36_spec-inventory-paper-doll-capacity.md`
- Keep: `docs/plans/v36_2026-06-08-inventory-paper-doll-capacity.md`

- [x] Step 7.1: After implementation is complete, mark v36 complete in `docs/PROGRESS.md`, add the lifecycle row, and summarize the proof and any deferred polish.
```bash
rg -n "v36|inventory-paper-doll-capacity|Next slice" docs/PROGRESS.md
```

- [x] Step 7.2: Run final CI.
```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -v`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot`
- [x] `HEADLESS=1 make bot-client scenario=13_inventory_paper_doll.json`
- [x] `make ci`

## Deferred scope

- Skill-tree or passive skill sources for `inventory_rows`.
- Stash, vendor, crafting, filters, sorting, or item comparison.
- Multi-cell item footprints.
- Production paper-doll art or a fully model-backed SubViewport preview if the fallback proves sufficient.
- Inventory UI plugin adoption beyond borrowed presentation ideas.
