# Spec: `treasure-classes-and-guarded-chests`

Status: Complete - make ci green on 2026-06-07
Branch: `feature/treasure-classes-and-guarded-chests`
Slice: v25 - treasure classes, multi-attempt monster drops, and guarded procedural chests
Baseline: v24 `main-menu-and-character-start`
Related:

- [`v18_spec-dungeon-levels-and-stairs.md`](v18_spec-dungeon-levels-and-stairs.md) - seeded dungeon level generation
- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) - generated dungeon mobs and proactive attacks
- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) - item templates, rarity rolls, rolled weapon payloads
- [`v24_spec-main-menu-and-character-start.md`](v24_spec-main-menu-and-character-start.md) - current baseline
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - dungeon progression and better loot by depth
- [`../adr/0009-boss-floors-and-timing-mechanics.md`](../adr/0009-boss-floors-and-timing-mechanics.md) - future boss-floor chest and reward pacing
- [`../PROGRESS.md`](../PROGRESS.md)

## 1. Purpose

v23 proved that dungeon mobs can drop rolled weapon instances, but the drop source still resolves
through a narrow loot-table entry. This slice adds a simple Diablo-style treasure class layer that
answers three questions server-side: how many drop attempts a source gets, whether each attempt
drops anything, and which concrete reward category/template is produced.

Monsters should be able to roll more than one possible drop, with later attempts becoming less
likely to succeed. Valid rewards include rolled equipment templates, fixed consumables such as
`red_potion`, and money-like fixed item drops that can continue to use current coin/badge
presentation until a real gold economy exists.

Dungeon chests become rare procedural events. When a generated dungeon level contains a chest, the
same seeded generation also increases monster density on that level, making the chest a guarded
reward rather than free loot.

## 2. Non-goals

- No Magic Find in v25. There is no placeholder MF stat, no debug MF override, and no rarity
  modification from player gear.
- No full Diablo 2-compatible recursive treasure class system.
- No unique/set item catalogs or special drop rules. Existing template rarity behavior remains the
  shipped rarity surface for this slice.
- No affix grammar, procedural names, item comparison UI, loot filters, stash, vendors, crafting,
  trade, or real gold wallet.
- No production chest art, chest animation, or audio. A simple placeholder interactable is enough.
- No boss pattern implementation. v25 may prepare chest loot rules for ADR-0009, but boss floors
  remain deferred.

## 3. Files to create or modify

```text
docs/specs/v25_spec-treasure-classes-and-guarded-chests.md       - this slice contract
docs/plans/v25_2026-06-07-treasure-classes-and-guarded-chests.md - implementation plan
shared/rules/treasure_classes.v0.schema.json                     - treasure class contract
shared/rules/treasure_classes.v0.json                            - first monster/chest drop tables
shared/rules/loot_tables.v0.schema.json                          - route loot tables to treasure classes if needed
shared/rules/loot_tables.v0.json                                 - migrate dungeon mob table to treasure class source
shared/rules/dungeon_generation.v0.schema.json                   - rare guarded chest generation knobs
shared/rules/dungeon_generation.v0.json                          - chest chance and monster-density bonus
shared/rules/interactables.v0.json                               - add `treasure_chest`
shared/protocol/session_snapshot.v1.schema.json                  - chest state/loot entity fields if new shape is needed
shared/protocol/state_delta.v1.schema.json                       - chest opened/drop events if new shape is needed
shared/golden/treasure_class_rolls.json                          - deterministic drop attempt fixture
shared/golden/guarded_chest_generation.json                      - deterministic chest/density fixture
tools/validate_shared.py                                         - treasure class and guarded chest validation
server/internal/game/rules.go                                    - parse treasure classes and chest generation rules
server/internal/game/sim.go                                      - multi-attempt drops and open-once chest drops
server/internal/game/game_test.go                                - treasure class, chest generation, replay determinism tests
server/internal/replay/...                                       - replay tests if chest open inputs/events need coverage
client/scripts/main.gd                                           - placeholder chest presentation/action target if needed
client/tests/test_golden.gd                                      - data-only treasure/chest golden checks
tools/bot/run.py                                                 - assertions/helpers for chest floors and multi-drop metadata
tools/bot/scenarios/17_treasure_classes_and_guarded_chests.json  - end-to-end proof
docs/PROGRESS.md                                                 - lifecycle update when v25 ships
```

## 4. Data shapes

### Treasure classes

New file: `shared/rules/treasure_classes.v0.json`.

Example shape:

```json
{
  "version": 0,
  "classes": {
    "dungeon_mob_tc_1": {
      "attempts": [
        {
          "attempt_id": "primary",
          "success_weight": 70,
          "no_drop_weight": 30,
          "entries": [
            { "item_template_id": "cave_blade", "weight": 30 },
            { "item_def_id": "red_potion", "weight": 45 },
            { "item_def_id": "training_badge", "weight": 25 }
          ]
        },
        {
          "attempt_id": "secondary",
          "success_weight": 25,
          "no_drop_weight": 75,
          "entries": [
            { "item_def_id": "red_potion", "weight": 60 },
            { "item_def_id": "training_badge", "weight": 40 }
          ]
        }
      ]
    },
    "guarded_chest_tc_1": {
      "attempts": [
        {
          "attempt_id": "guaranteed",
          "success_weight": 100,
          "no_drop_weight": 0,
          "entries": [
            { "item_template_id": "cave_blade", "weight": 60 },
            { "item_def_id": "red_potion", "weight": 40 }
          ]
        },
        {
          "attempt_id": "bonus",
          "success_weight": 35,
          "no_drop_weight": 65,
          "entries": [
            { "item_def_id": "training_badge", "weight": 100 }
          ]
        }
      ]
    }
  }
}
```

Every attempt rolls success/no-drop first. On success, it rolls one entry from that attempt's
weighted entries. Later attempts can have lower `success_weight`, which gives monsters multiple
possible drops without making every kill flood the floor.

Each entry must declare exactly one of `item_def_id`, `item_template_id`, or any future bounded
entry kind added by a later spec. v25 supports only fixed item defs and item templates.

### Loot table bridge

Existing monsters already point at `loot_table` ids. v25 may either migrate `loot_tables.v0.json`
entries to point at a treasure class or replace internal loot-table resolution with a treasure-class
lookup while preserving existing monster fields. One acceptable bridge shape is:

```json
{
  "dungeon_mob_drop": {
    "treasure_class_id": "dungeon_mob_tc_1"
  }
}
```

Legacy fixed/template loot entries may remain for old scenarios, but `dungeon_mob` must use the new
treasure class path.

### Guarded chest generation

`shared/rules/dungeon_generation.v0.json` gains a data-driven chest block:

```json
{
  "chest_placement": {
    "enabled": true,
    "chance_weight": 15,
    "no_chest_weight": 85,
    "interactable_def_id": "treasure_chest",
    "loot_table": "guarded_chest_drop",
    "monster_count_bonus": 2,
    "min_stair_distance": 4.0,
    "max_attempts": 32
  }
}
```

Chest presence, chest position, and bonus monster placement are derived from the dungeon level seed.
If the chest roll fails, the level uses the normal monster count. If the chest roll succeeds, the
level spawns the chest and applies `monster_count_bonus` to the same level's generated monster
count.

### Chest interaction

`treasure_chest` is an interactable opened through existing `action_intent`.

Flow:

```text
player action_intent -> closed treasure_chest
  -> Sim validates range and closed state
  -> Sim marks chest opened
  -> Sim rolls chest treasure class once
  -> Sim spawns zero or more loot entities near the chest
  -> Sim emits authoritative events/changes
  -> replay/resume preserves opened state and spawned loot without rerolling
```

The client may render the chest as a simple placeholder and rely on the existing interactable click
path. Loot entities continue to use existing presentation metadata.

## 5. Architecture and determinism

All drop rolls happen in the Go Sim through the seeded RNG discipline already used by v23 item
rolls. The order must be stable: source killed/opened, attempt order as declared, success/no-drop
roll, entry roll, then item-template rarity/stat roll when applicable.

Dungeon chest generation is also deterministic from the level seed. Chest presence must not depend
on map iteration order, wall-clock time, or client state. The same session seed and ordered inputs
must reproduce chest placement, bonus monster count, monster positions, chest opened state, and
spawned drops.

The client remains presentation-only. It never decides whether a chest exists, whether a monster or
chest drops loot, or what item/rarity/stat payload appears.

## 6. Acceptance criteria

1. `make validate-shared` validates `treasure_classes.v0.json`, treasure class references, drop
   attempt weights, item def/template references, chest generation fields, and interactable
   references.
2. Monster treasure classes support multiple ordered drop attempts, with later attempts using lower
   success probability.
3. Treasure class entries can resolve to rolled item templates, fixed consumables such as
   `red_potion`, and money-like fixed item defs such as `training_badge`.
4. Magic Find is not implemented and no player stat modifies treasure class or rarity rolls.
5. `dungeon_mob` drops through a treasure class instead of directly rolling a single template entry.
6. A generated dungeon level can rarely place a `treasure_chest` interactable from seeded PCG.
7. If a chest is generated, the same level deterministically spawns additional monsters from
   `monster_count_bonus`.
8. Opening a chest via `action_intent` rolls its treasure class once, spawns loot near the chest,
   and rejects/reuses state on later open attempts without duplicate drops.
9. `shared/golden/treasure_class_rolls.json` pins at least: one no-drop attempt, one multi-drop
   monster kill, one fixed potion/money-like drop, and one rolled equipment drop.
10. `shared/golden/guarded_chest_generation.json` pins one seed with no chest and one seed with a
    chest plus increased monster count.
11. Go tests prove treasure class rolls, generated chest placement, chest open-once semantics, and
    replay determinism.
12. Godot `test_golden.gd` validates the data-only treasure/chest golden fixtures.
13. Bot scenario `17_treasure_classes_and_guarded_chests.json` proves dungeon kill drops, rare chest
    floor behavior, chest open, pickup, `/state`, reconnect, replay, and fresh-session persistence.
14. `make ci` green.

## 7. Testing plan

1. `make validate-shared`
2. `cd server && go test ./internal/game/... -run 'Treasure|Chest|Loot|Dungeon'`
3. `cd server && go test ./internal/http/... ./internal/replay/... -run 'Treasure|Chest|Replay|Item'`
4. `make client-unit`
5. `make bot` - includes `17_treasure_classes_and_guarded_chests.json`
6. `make ci`
7. Manual: `make play`, descend into dungeon floors until a chest appears, fight the guarded floor,
   open the chest, pick up loot, restart play, and confirm character-owned item persistence remains
   intact.

## 8. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | v25 uses treasure classes as data under `shared/rules/`. | Preserves ADR-0001 shared-rules discipline and makes future depth/boss loot data-driven. |
| 2 | Drop attempts roll success/no-drop before entry selection. | Matches the desired simple D2-style flow while making decreasing later attempts explicit. |
| 3 | Magic Find is fully deferred. | Avoids inventing a fake player stat before progression/gear stats exist. |
| 4 | Money is represented by an existing fixed item def in v25. | Proves money-like drops without adding a wallet/economy slice. |
| 5 | Chests are rare procedural dungeon events with bonus monsters. | Makes chest rewards feel guarded and aligns with future ADR-0009 boss-floor chests. |
| 6 | Chest open state is authoritative and single-use. | Prevents duplicate loot through reconnect, replay, or repeated actions. |

## 9. Open questions

| # | Question | Default |
|---|----------|---------|
| Q-1 | Should v25 rename `training_badge` to a more explicit money placeholder? | No. Keep existing def/presentation and document it as money-like until a gold economy exists. |
| Q-2 | Should chest chance be tuned to be very low in normal play? | Use a low default weight, but pin both chest and no-chest seeds in goldens and bot proof. |
| Q-3 | Should old single-entry loot tables be removed? | No. Keep legacy tables if needed for old scenarios, but migrate `dungeon_mob` to treasure classes. |

## 10. Deferred follow-ups

- Real gold wallet, stackable currency, and economy UI.
- Magic Find once character stats/equipment-derived modifiers exist.
- Unique/set item catalogs and special item-specific drop rules.
- Depth-banded treasure class upgrades for deeper dungeon floors.
- Boss-floor treasure chest integration from ADR-0009.
- Production chest art, opening animation, and sound.
