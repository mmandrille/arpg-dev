# v25 Plan - Treasure Classes and Guarded Chests

Status: Complete - make ci green on 2026-06-07
Goal: Add data-driven treasure classes with multi-attempt monster drops and rare guarded procedural dungeon chests.
Architecture: Treasure classes live in `shared/rules/` and are resolved only by the Go Sim. Existing monster `loot_table` references bridge to treasure classes so old scenarios can keep legacy fixed drops while `dungeon_mob` moves to the new path. Dungeon generation rolls chest presence from the level seed; a chest level applies a monster-count bonus and the chest opens once through existing `action_intent`.
Tech stack: shared JSON schemas/goldens, Go authoritative sim and replay tests, Godot placeholder presentation/golden checks, Python protocol bot scenario.

## Baseline and shortcut decision

v25 builds on v24 `main-menu-and-character-start`, reusing v18 dungeon generation, v21 dungeon mobs, v23 item templates/rolled item persistence, and existing interactable/action handling from v10/v18/v19.
## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Created | `docs/specs/v25_spec-treasure-classes-and-guarded-chests.md` | Slice contract |
| Create | `shared/rules/treasure_classes.v0.schema.json` | Treasure class schema |
| Create | `shared/rules/treasure_classes.v0.json` | First monster/chest treasure classes |
| Modify | `shared/rules/loot_tables.v0.schema.json` | Allow `treasure_class_id` bridge entries |
| Modify | `shared/rules/loot_tables.v0.json` | Point `dungeon_mob_drop` and chest drop to treasure classes |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Add guarded chest generation knobs |
| Modify | `shared/rules/dungeon_generation.v0.json` | Chest chance, placement, and monster-count bonus |
| Modify | `shared/rules/interactables.v0.json` | Add `treasure_chest` |
| Create | `shared/golden/treasure_class_rolls.json` | Pin treasure roll outcomes |
| Create | `shared/golden/guarded_chest_generation.json` | Pin chest/no-chest generation |
| Modify | `tools/validate_shared.py` | Validate treasure classes and guarded chest references |
| Modify | `server/internal/game/rules.go` | Parse treasure classes and validation invariants |
| Modify | `server/internal/game/sim.go` | Resolve multi-attempt drops and open-once chest loot |
| Modify | `server/internal/game/game_test.go` | Unit/golden/replay-style sim coverage |
| Modify | `server/internal/replay/...` | Add replay coverage if sim tests do not cover chest reconstruction |
| Modify | `client/scripts/main.gd` | Placeholder chest presentation/action support if current interactables need mapping |
| Modify | `client/tests/test_golden.gd` | Data-only golden checks |
| Modify | `tools/bot/run.py` | Bot assertions/helpers for chest floor and multi-drop checks |
| Create | `tools/bot/scenarios/17_treasure_classes_and_guarded_chests.json` | End-to-end proof |
| Modify | `PROGRESS.md` | Lifecycle update when slice ships |

## Task 1 - Shared treasure class contracts

Files:
- Create: `shared/rules/treasure_classes.v0.schema.json`
- Create: `shared/rules/treasure_classes.v0.json`
- Modify: `shared/rules/loot_tables.v0.schema.json`
- Modify: `shared/rules/loot_tables.v0.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Define `treasure_classes.v0.schema.json` with `version`, `classes`, ordered `attempts`, `attempt_id`, `success_weight`, `no_drop_weight`, and weighted `entries`.
- [x] Step 1.2: Enforce each treasure entry declares exactly one supported reward source: `item_def_id` or `item_template_id`.
- [x] Step 1.3: Add `dungeon_mob_tc_1` with at least two ordered attempts; make the second attempt lower probability than the first.
- [x] Step 1.4: Include rolled equipment, `red_potion`, and `training_badge` in treasure class entries.
- [x] Step 1.5: Add `guarded_chest_tc_1` with one guaranteed or high-probability primary attempt and one lower-probability bonus attempt.
- [x] Step 1.6: Extend `loot_tables.v0.schema.json` so a table can bridge to `treasure_class_id`; keep legacy `entries` and `drops` valid for old scenarios.
- [x] Step 1.7: Change `dungeon_mob_drop` to resolve through `dungeon_mob_tc_1`; add `guarded_chest_drop` through `guarded_chest_tc_1`.
- [x] Step 1.8: Extend `tools/validate_shared.py` to reject unknown treasure class ids, unknown item defs/templates, invalid weights, empty attempts, empty success-entry sets, and mixed legacy/treasure bridge shapes.

```bash
make validate-shared
```

## Task 2 - Shared guarded chest generation contracts

Files:
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/interactables.v0.json`
- Create: `shared/golden/guarded_chest_generation.json`
- Modify: `tools/validate_shared.py`

- [x] Step 2.1: Add a `chest_placement` block with `enabled`, `chance_weight`, `no_chest_weight`, `interactable_def_id`, `loot_table`, `monster_count_bonus`, `min_stair_distance`, and `max_attempts`.
- [x] Step 2.2: Add `treasure_chest` to `interactables.v0.json` with `initial_state: "closed"` and no transition.
- [x] Step 2.3: Validate chest placement references `treasure_chest`, a loot table that resolves to a treasure class, non-negative bonus count, positive chance/no-chest total, and positive max attempts.
- [x] Step 2.4: Create `guarded_chest_generation.json` with one pinned seed/depth that produces no chest and one pinned seed/depth that produces a chest plus normal monster count plus bonus.
- [x] Step 2.5: Include pinned chest position and monster count in the fixture; keep exact monster positions if generation already exposes stable positions cheaply.

```bash
make validate-shared
```

## Task 3 - Go rules and treasure roll engine

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/game_test.go`
- Create: `shared/golden/treasure_class_rolls.json`

- [x] Step 3.1: Add Go structs for treasure classes, attempts, and entries.
- [x] Step 3.2: Load `treasure_classes.v0.json` with the same validation rules as `tools/validate_shared.py`.
- [x] Step 3.3: Extend internal loot-table resolution so a table can return fixed legacy drops, weighted legacy entries, or a treasure class id.
- [x] Step 3.4: Implement deterministic treasure class resolution: attempt order, success/no-drop roll, entry roll, and item-template roll only after a template entry succeeds.
- [x] Step 3.5: Return zero or more internal drop descriptors from a treasure class; do not emit observable no-drop events unless current event conventions already require them.
- [x] Step 3.6: Add `treasure_class_rolls.json` with cases for no-drop, multi-drop monster kill, fixed potion/money-like drop, and rolled equipment drop.
- [x] Step 3.7: Add Go golden/unit tests proving the fixture outcomes and that Magic Find is absent from rule parsing and roll inputs.

```bash
cd server && go test ./internal/game/... -run 'Treasure|Loot|ItemRoll'
```

## Task 4 - Go dungeon generation and chest entities

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 4.1: Parse `chest_placement` into dungeon generation rules.
- [x] Step 4.2: Extend deterministic dungeon generation to roll chest presence before applying monster count, using the level seed and stable ordered operations.
- [x] Step 4.3: Place the chest away from stairs using `min_stair_distance` and `max_attempts`; if placement fails, treat the level as no-chest and do not apply the monster bonus.
- [x] Step 4.4: Apply `monster_count_bonus` only when the chest is successfully placed.
- [x] Step 4.5: Spawn `treasure_chest` as an interactable in generated dungeon levels with closed state.
- [x] Step 4.6: Add tests for `guarded_chest_generation.json`: no-chest seed, chest seed, monster count delta, and stable chest position.
- [x] Step 4.7: Add tests proving chest generation does not perturb stair/teleporter golden expectations except where the new fixture explicitly covers chest floors.

```bash
cd server && go test ./internal/game/... -run 'Dungeon|Chest'
```

## Task 5 - Go sim chest interaction and replay safety

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/replay/...`
- Modify: `shared/protocol/session_snapshot.v1.schema.json`
- Modify: `shared/protocol/state_delta.v1.schema.json`

- [x] Step 5.1: Route `action_intent` on `treasure_chest` through the existing interactable path and range validation.
- [x] Step 5.2: On first valid open, mark chest state opened, resolve its loot table/treasure class once, and spawn zero or more loot entities near the chest using existing cluster placement rules.
- [x] Step 5.3: Emit an authoritative event such as `interactable_activated` or `chest_opened`; use existing schema fields if sufficient, otherwise add additive event fields and examples.
- [x] Step 5.4: Reject or no-op repeated opens without duplicating loot; choose the behavior that best matches existing door/interactable semantics and cover it in tests.
- [x] Step 5.5: Ensure snapshot, `/state`, reconnect resume, replay timeline, and fresh-session reconstruction preserve opened chest state and already-spawned loot without rerolling.
- [x] Step 5.6: Add sim/replay tests for open once, repeated action, loot pickup after chest open, and deterministic replay from the same seed/input sequence.

```bash
cd server && go test ./internal/game/... ./internal/replay/... -run 'Chest|Treasure|Replay'
```

## Task 6 - Client presentation and golden checks

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_golden.gd`

- [x] Step 6.1: Verify current interactable rendering handles unknown/non-door interactables acceptably; if not, add a small placeholder branch for `treasure_chest`.
- [x] Step 6.2: Keep the client display-only: it should render/click the chest but never compute chest presence or drops.
- [x] Step 6.3: Add data-only Godot golden checks for `treasure_class_rolls.json` and `guarded_chest_generation.json` references and expected fixture shape.
- [x] Step 6.4: Ensure any new event name used for chest open does not require animation wiring beyond basic presentation.

```bash
make client-unit
```

## Task 7 - Bot scenario

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/scenarios/17_treasure_classes_and_guarded_chests.json`

- [x] Step 7.1: Add bot assertions/helpers as needed for generated level chest presence, monster count, loot count, item defs/templates, and interactable state.
- [x] Step 7.2: Create scenario `17_treasure_classes_and_guarded_chests.json` using pinned seed/session setup that reaches a chest-producing dungeon level.
- [x] Step 7.3: In the scenario, descend to the pinned level, assert chest presence and monster count includes `monster_count_bonus`.
- [x] Step 7.4: Kill at least one `dungeon_mob` and assert treasure-class-driven loot can include multiple drops and valid fixed/template rewards.
- [x] Step 7.5: Open the chest through `action_entity`, assert loot appears, pick up at least one fixed reward and one rolled reward when the seed supports it.
- [x] Step 7.6: Attempt to open the same chest again and assert no duplicate drops.
- [x] Step 7.7: Run `/state`, reconnect resume, replay, and fresh-session persistence assertions using existing bot scenario patterns.

```bash
make bot
```

## Task 8 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v25_spec-treasure-classes-and-guarded-chests.md`
- Modify: `docs/plans/v25_2026-06-07-treasure-classes-and-guarded-chests.md`

- [x] Step 8.1: When implementation ships, mark the spec `Complete - make ci green on 2026-06-07` or the actual completion date.
- [x] Step 8.2: Add v25 to the `PROGRESS.md` lifecycle table and update latest completed slice.
- [x] Step 8.3: Document as-built behavior: treasure classes, multiple attempts, potion/money-like drops, rare guarded chest generation, chest open-once behavior, bot scenario `17`, and deferred Magic Find.
- [x] Step 8.4: Keep deferred follow-ups explicit: gold wallet, Magic Find, unique/set catalogs, depth-banded treasure classes, boss-floor chest integration, and production chest art.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'Treasure|Chest|Loot|Dungeon|ItemRoll'`
- [x] `cd server && go test ./internal/http/... ./internal/replay/... -run 'Treasure|Chest|Replay|Item'`
- [x] `make client-unit`
- [x] `make bot`
- [x] `make ci`

## Deferred scope

- Magic Find and any player stat that modifies rarity or drop rolls.
- Real gold wallet, stackable currency, vendors, stash, crafting, trade, and economy UI.
- Unique/set item catalogs and special drop rules.
- Depth-banded treasure class upgrades for deeper levels.
- ADR-0009 boss timing and boss-floor locked-stairs implementation.
- Production chest art, animation, and audio.
