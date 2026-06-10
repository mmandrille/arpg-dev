# v29 Plan - Dungeon Equipment Drop Expansion

Status: Implemented
Goal: Make real generated dungeon monsters and guarded chests drop the expanded v28 equipment catalog through deterministic, depth-aware treasure classes.
Architecture: Add a thin shared-data depth band model for levels `-1`, `-2`, and `-3+`, then use it when generated dungeon sources are created. Monster and chest loot remains server-authoritative and replay-stable; no protocol schema change is expected because existing rolled item payloads already carry template, rarity, stats, slot, equipment, and persistence data.
Tech stack: Shared JSON rules and golden fixtures, Go authoritative sim, Python protocol bot, optional Godot golden fixture check only if the new fixture is consumed by client tests.

## Baseline and shortcut decision

Baseline is v28 `full-equipment-and-belt-hotbar`: the game already has all current equipment
templates, paper-doll slots, two-hand occupancy, belt-gated hotbar capacity, rolled item
persistence, and protocol/client bot proofs through `equipment_lab`.

This slice reuses:

- v23 rolled item template payloads and persistence.
- v25 treasure class attempts and guarded chest open-once behavior.
- v28 template catalog and equipment slot assertions.

Godot plugin adoption: **reject for v29**. The slice changes shared loot data, Go source selection,
goldens, and protocol bot proof. It does not add client UI, camera, inventory presentation, or art,
so there is no plugin or asset pack to adopt/borrow.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Add optional/required first-pass `loot_bands` schema for `-1`, `-2`, `-3+`. |
| Modify | `shared/rules/dungeon_generation.v0.json` | Configure monster/chest loot tables by depth and document the temporary band model. |
| Modify | `shared/rules/loot_tables.v0.json` | Add depth-specific monster/chest loot table ids that bridge to treasure classes. |
| Modify | `shared/rules/treasure_classes.v0.json` | Add broadened monster/chest treasure classes covering v28 equipment families. |
| Create | `shared/golden/dungeon_equipment_drops.v0.schema.json` | Schema for depth/source fixture. |
| Create | `shared/golden/dungeon_equipment_drops.json` | Pin representative depth-band/source outcomes. |
| Modify | `shared/golden/treasure_class_rolls.json` | Pin varied direct treasure class rolls for key equipment families. |
| Modify | `tools/validate_shared.py` | Validate loot bands, table/class references, weights, and `-3+` equipment reachability. |
| Modify | `server/internal/game/rules.go` | Parse/validate loot bands and expose typed rules. |
| Modify | `server/internal/game/dungeon_gen.go` | Assign generated source loot table ids from the active depth band. |
| Modify | `server/internal/game/sim.go` | Spawn generated monsters/chests with selected source loot table ids while preserving roll order. |
| Modify | `server/internal/game/game_test.go` | Add depth selection, category reachability, golden, and replay determinism coverage. |
| Modify | `client/tests/test_golden.gd` | Add data-only checks if `dungeon_equipment_drops.json` is included in client golden coverage. |
| Modify | `tools/bot/run.py` | Add assertions/helpers for slot/template category if existing helpers are insufficient. |
| Create | `tools/bot/scenarios/20_dungeon_equipment_drops.json` | End-to-end real dungeon drop proof. |
| Modify | `PROGRESS.md` | Lifecycle update when v29 ships. |

## Task 1 - Shared depth bands and treasure classes

Files:

- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/loot_tables.v0.json`
- Modify: `shared/rules/treasure_classes.v0.json`

- [x] Step 1.1: Add `loot_bands` to dungeon generation rules with exactly the intended coarse bands: depth `1`, depth `2`, and depth `3+`.

```bash
make validate-shared
```

- [x] Step 1.2: Add depth-specific monster loot tables such as `dungeon_mob_drop_depth_1`, `dungeon_mob_drop_depth_2`, and `dungeon_mob_drop_depth_3_plus`.

```bash
make validate-shared
```

- [x] Step 1.3: Add depth-specific guarded chest loot tables such as `guarded_chest_drop_depth_1`, `guarded_chest_drop_depth_2`, and `guarded_chest_drop_depth_3_plus`.

```bash
make validate-shared
```

- [x] Step 1.4: Add treasure classes with monster odds lower than chest odds. By `-3+`, ensure every v28 template is reachable through at least one configured dungeon/chest source.

```bash
make validate-shared
```

- [x] Step 1.5: Document in `dungeon_generation.v0.json` or adjacent validation comments that the `-1`, `-2`, `-3+` bands are a temporary first pass to improve later.

```bash
make validate-shared
```

## Task 2 - Shared validation and golden fixtures

Files:

- Create: `shared/golden/dungeon_equipment_drops.v0.schema.json`
- Create: `shared/golden/dungeon_equipment_drops.json`
- Modify: `shared/golden/treasure_class_rolls.json`
- Modify: `tools/validate_shared.py`

- [x] Step 2.1: Extend shared validation so loot bands must have positive depth ranges, non-overlapping coverage for `1`, `2`, and `3+`, known loot table ids, and table ids that resolve to treasure classes.

```bash
make validate-shared
```

- [x] Step 2.2: Extend validation so `-3+` configured dungeon/chest sources can reach every v28 equipment template: `cave_blade`, `cave_greatsword`, `cave_bow`, `cave_shield`, `cave_helm`, `cave_mail`, `cave_gloves`, `cave_belt`, `cave_boots`, `cave_ring`, and `cave_amulet`.

```bash
make validate-shared
```

- [x] Step 2.3: Add `dungeon_equipment_drops.json` cases that pin depth/source selection and representative rewards for a monster source and a guarded chest source.

```bash
make validate-shared
```

- [x] Step 2.4: Update `treasure_class_rolls.json` or the new fixture to pin varied direct rolls for at least bow, shield, belt, armor, jewelry, potion, and money-like item outcomes.

```bash
make validate-shared
```

## Task 3 - Go rules loader and dungeon generation

Files:

- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 3.1: Add typed Go structures for loot bands and validate the same invariants enforced by `tools/validate_shared.py`.

```bash
cd server && go test ./internal/game/... -run 'Rules|Dungeon'
```

- [x] Step 3.2: Add a deterministic helper for resolving the loot band from `abs(levelNum)` and returning monster/chest loot table ids.

```bash
cd server && go test ./internal/game/... -run 'Dungeon|Loot|Depth'
```

- [x] Step 3.3: Extend generated monster state to carry the selected loot table id instead of always inheriting `dungeon_mob`'s static `LootTable`.

```bash
cd server && go test ./internal/game/... -run 'Dungeon|Loot|Monster'
```

- [x] Step 3.4: Extend generated chest state to use the selected depth-band chest loot table instead of the single `chest_placement.loot_table`.

```bash
cd server && go test ./internal/game/... -run 'Dungeon|Chest|Loot'
```

- [x] Step 3.5: Preserve existing RNG stream order for level geometry, chest presence, monster placement, and loot rolls. Source loot table selection should be data lookup only, not RNG-consuming.

```bash
cd server && go test ./internal/game/... -run 'Dungeon|Chest|Treasure|Replay'
```

## Task 4 - Go test coverage

Files:

- Modify: `server/internal/game/game_test.go`
- Modify: `client/tests/test_golden.gd` only if client golden coverage consumes the new fixture

- [x] Step 4.1: Add tests for depth-band selection at levels `-1`, `-2`, `-3`, and a deeper level such as `-10`.

```bash
cd server && go test ./internal/game/... -run 'Depth|Dungeon'
```

- [x] Step 4.2: Add tests proving generated monsters and generated chests store the expected loot table ids for each band.

```bash
cd server && go test ./internal/game/... -run 'Dungeon|Loot'
```

- [x] Step 4.3: Add golden-backed tests for representative `dungeon_equipment_drops.json` outcomes.

```bash
cd server && go test ./internal/game/... -run 'Equipment|Treasure|Golden'
```

- [x] Step 4.4: Add replay/determinism coverage proving the same seed and inputs reproduce varied equipment drops and loot entity order.

```bash
cd server && go test ./internal/game/... -run 'Replay|Equipment|Loot'
```

- [x] Step 4.5: If the new golden fixture is parsed by Godot tests, add `client/tests/test_golden.gd` assertions for the data-only shape.

```bash
make client-unit
```

## Task 5 - Protocol bot scenario

Files:

- Modify: `tools/bot/run.py`
- Create: `tools/bot/scenarios/20_dungeon_equipment_drops.json`
- Modify: `tools/bot/test_protocol.py` if scenario validation needs new action/assertion tests

- [x] Step 5.1: Add bot assertions only where needed. Prefer reusing existing `kill_monsters`, `action_entity`, `pick_up_loot`, `assert_rolled_inventory_item`, `equip_inventory_item`, and equipment slot assertions.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 5.2: Create `20_dungeon_equipment_drops.json` using `world_id: "dungeon_levels"` and a pinned seed that proves real generated dungeon play, not `equipment_lab`.

```bash
make bot
```

- [x] Step 5.3: Scenario steps should descend to configured depth, kill/open the configured source, pick up representative varied gear, equip at least one non-weapon equipment item plus one hand item where feasible, then assert `/state`, reconnect, replay, and fresh-session persistence through the standard bot flow.

```bash
make bot
```

- [x] Step 5.4: Keep the scenario representative, not exhaustive. Do not force every equipment category into one run; validation/goldens cover full reachability.

```bash
make bot
```

## Task 6 - Regression checks and existing scenarios

Files:

- Modify: existing scenario JSON only if changed loot data breaks hard-coded assumptions.
- Modify: `tools/bot/test_protocol.py` only if scenario catalog expectations need updates.

- [x] Step 6.1: Check `17_treasure_classes_and_guarded_chests.json` after loot-table changes. Update only if the pinned seed now intentionally drops a different template.

```bash
make bot
```

- [x] Step 6.2: Check `19_full_equipment.json` remains lab-proven and still covers full paper-doll/hotbar behavior independent of real dungeon drop randomness.

```bash
make bot
```

- [x] Step 6.3: Run client smoke only if real dungeon drop presentation or scenario-visible client state changed.

```bash
make client-smoke
```

## Task 7 - Lifecycle docs and CI

Files:

- Modify: `PROGRESS.md`
- Modify: `docs/specs/v29_spec-dungeon-equipment-drop-expansion.md` if implementation decisions differ from the draft

- [x] Step 7.1: When implementation is complete, mark the v29 spec as implemented and update any resolved implementation notes.

```bash
rg -n "Status: Draft|v29|dungeon-equipment-drop-expansion" docs/specs PROGRESS.md
```

- [x] Step 7.2: Update `PROGRESS.md` lifecycle table, current status, slice summary, scripted scenario catalog, and deferred follow-ups.

```bash
rg -n "v29|dungeon_equipment_drops|dungeon-equipment-drop-expansion" PROGRESS.md tools/bot/scenarios
```

- [x] Step 7.3: Run full CI.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'Dungeon|Treasure|Loot|Equipment|Depth|Replay'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make client-unit`
- [x] `make bot`
- [x] `make client-smoke` if client presentation/golden parsing changed
- [x] `make ci`

## Deferred scope

- Real item-level and depth progression beyond coarse `-1`, `-2`, `-3+` bands.
- Magic Find, player/equipment-derived drop modifiers, and richer rarity curves.
- Unique/set catalogs, boss-floor rewards, and special item-specific drop rules.
- Real gold wallet, stackable currency, vendors, stash, crafting, trade, loot filters, and comparison UI.
- Armor mitigation, shield block execution, crit/hit/attack-speed gameplay, and offhand abilities.
