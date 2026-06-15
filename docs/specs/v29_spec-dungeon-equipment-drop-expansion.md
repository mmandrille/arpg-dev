# Spec: `dungeon-equipment-drop-expansion`

Status: Implemented
Branch: `main`
Slice: v29 - real dungeon drops for the expanded equipment catalog
Baseline: v28 `full-equipment-and-belt-hotbar`
Related:

- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) - rolled item templates and persistence
- [`v25_spec-treasure-classes-and-guarded-chests.md`](v25_spec-treasure-classes-and-guarded-chests.md) - treasure classes and guarded chest generation
- [`v28_spec-full-equipment-and-belt-hotbar.md`](v28_spec-full-equipment-and-belt-hotbar.md) - paper-doll equipment, all current templates, and belt hotbar
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - dungeon progression and better loot by depth
- [`../../PROGRESS.md`](../../PROGRESS.md)

## 1. Purpose

v28 proved the full paper-doll equipment model, two-hand occupancy, and belt-gated hotbar, but the
normal dungeon reward path still behaves like an early weapon demo. `dungeon_mob_tc_1` primarily
produces `cave_blade`, while the broader catalog - bows, shields, belts, armor, boots, gloves,
helmets, rings, and amulets - is mainly proven through `equipment_lab`.

This slice makes ordinary generated dungeon play use the expanded equipment catalog. Dungeon
monsters and guarded chests should produce varied rolled equipment through server-owned treasure
classes, and the proof should run through real dungeon floors instead of a lab-only preset.

The slice also introduces a deliberately thin first-pass depth band model:

```text
level -1  -> entry dungeon loot
level -2  -> deeper dungeon loot
level -3+ -> broad/deeper dungeon loot
```

This is not the final item progression economy. It is a clear data hook for future improvement:
later slices should replace or enrich these coarse bands with stronger depth curves, item-level
gates, rarity tuning, boss-floor rewards, Magic Find, and/or region-specific tables.

## 2. Non-goals

- No affix grammar, prefix/suffix system, procedural item names, unique/set items, or special
  item-specific drop rules.
- No Magic Find stat, debug Magic Find override, or rarity modifier from player gear.
- No real gold wallet. Money-like rewards may continue using `training_badge` until an economy
  slice exists.
- No armor mitigation, block chance execution, crit/hit/attack-speed gameplay, resistances, dodge,
  mana effects, offhand abilities, or dual-wield behavior.
- No vendors, stash, crafting, trade, item comparison UI, loot filters, or item pickup targeting
  overhaul.
- No production item icons, ground art, chest art, audio, or VFX.
- No full Diablo 2-compatible recursive treasure class system.
- No client-side loot logic. The Godot client remains presentation-only.

## 3. Files to create or modify

```text
docs/specs/v29_spec-dungeon-equipment-drop-expansion.md       - this slice contract
docs/plans/v29_2026-06-07-dungeon-equipment-drop-expansion.md - implementation plan
shared/rules/treasure_classes.v0.json                         - broaden monster/chest equipment pools
shared/rules/loot_tables.v0.json                              - route dungeon sources to depth-aware classes if needed
shared/rules/dungeon_generation.v0.schema.json                - first-pass depth loot band config if implemented in generation rules
shared/rules/dungeon_generation.v0.json                       - -1, -2, -3+ loot band rules
shared/golden/treasure_class_rolls.json                       - varied equipment roll fixture updates
shared/golden/dungeon_equipment_drops.json                    - depth/source equipment drop fixture
shared/golden/dungeon_equipment_drops.v0.schema.json          - fixture schema
tools/validate_shared.py                                      - treasure class, depth band, template reachability checks
server/internal/game/rules.go                                 - parse depth loot band data
server/internal/game/dungeon_gen.go                           - assign generated source loot table/class by level band
server/internal/game/sim.go                                   - use selected source loot class without breaking roll order
server/internal/game/game_test.go                             - depth/source loot, category reachability, determinism tests
client/tests/test_golden.gd                                   - data-only golden checks if fixture is shared with Godot
tools/bot/run.py                                              - assertions for dropped template ids, slots, and equipment categories
tools/bot/scenarios/20_dungeon_equipment_drops.json           - end-to-end real-play proof
PROGRESS.md                                              - lifecycle update when v29 ships
```

No protocol schema change is expected. Existing rolled item fields from v23/v28 should carry every
new reward.

Implementation note: the protocol bot proof remains representative rather than exhaustive. It
uses real generated dungeon play and generic rolled-equipment assertions, while full category and
template reachability is pinned by shared validation plus Go/golden tests.

## 4. Data model

### 4.1 Depth loot bands

Add a data-driven way to select monster and chest loot by generated dungeon level. The exact
location may be `dungeon_generation.v0.json` or another shared rules file if the plan finds a
cleaner existing boundary.

Acceptable shape:

```json
{
  "loot_bands": [
    {
      "min_depth": 1,
      "max_depth": 1,
      "monster_loot_table": "dungeon_mob_drop_depth_1",
      "chest_loot_table": "guarded_chest_drop_depth_1"
    },
    {
      "min_depth": 2,
      "max_depth": 2,
      "monster_loot_table": "dungeon_mob_drop_depth_2",
      "chest_loot_table": "guarded_chest_drop_depth_2"
    },
    {
      "min_depth": 3,
      "max_depth": null,
      "monster_loot_table": "dungeon_mob_drop_depth_3_plus",
      "chest_loot_table": "guarded_chest_drop_depth_3_plus"
    }
  ]
}
```

Depth is `abs(level)` for negative dungeon floors. Level `0` town does not use dungeon loot bands.

This is intentionally coarse. The implementation must document in the rules or spec follow-ups that
`-1`, `-2`, and `-3+` are temporary bands for v29 proof only, not the final progression model.

### 4.2 Monster treasure classes

Dungeon monster drops should broaden gradually. Monsters should not drop every category constantly,
but normal dungeon kills must be able to produce more than `cave_blade`.

Suggested direction:

| Band | Monster reward intent |
|------|-----------------------|
| `-1` | Basic weapons, potions, money-like item, occasional armor/accessory |
| `-2` | Broader one-hand/two-hand/bow/shield/armor mix |
| `-3+` | Full current v28 template catalog reachable, still with no-drop/bonus attempt tuning |

The exact weights are implementation-tuned, but the resulting classes must include reachable
entries for:

- `cave_blade`
- `cave_greatsword`
- `cave_bow`
- `cave_shield`
- at least one armor piece
- `cave_belt`
- at least one jewelry item
- `red_potion`
- `training_badge` or the current money-like placeholder

### 4.3 Guarded chest treasure classes

Chests should have better equipment odds than normal monsters. A guarded chest can be rarer than
a monster kill, but when it appears and is opened it should feel like a better reward source.

Suggested direction:

- Primary chest attempt: guaranteed or near-guaranteed equipment.
- Bonus chest attempt: potion, money-like item, or a second equipment roll.
- Deeper chest bands: broader pool and/or slightly better odds for belt, jewelry, shield, bow, and
  armor categories.

Chest open-once semantics from v25 remain unchanged. Opening a chest must roll its configured
depth-band class exactly once and preserve spawned loot through reconnect and replay.

### 4.4 Template availability

All v28 templates may drop at character level `1` in v29:

- `cave_blade`
- `cave_greatsword`
- `cave_bow`
- `cave_shield`
- `cave_helm`
- `cave_mail`
- `cave_gloves`
- `cave_belt`
- `cave_boots`
- `cave_ring`
- `cave_amulet`

Real stat requirements, item levels, and difficulty-tier gating remain deferred. Validation should
still ensure every referenced template exists and every current equipment slot has at least one
reachable configured source by `-3+`.

## 5. Architecture and determinism

All reward decisions remain in the Go Sim. The client never chooses a loot band, reward class,
drop count, item template, rarity, or rolled stat payload.

Deterministic order matters:

```text
generated level number
  -> choose depth band from shared data
  -> spawn monster/chest with selected loot table/class id
  -> source killed/opened by ordered input
  -> treasure class attempts in declared order
  -> success/no-drop roll
  -> entry roll
  -> item-template rarity/stat roll when applicable
  -> stable loot entity spawn order
```

The implementation must avoid wall-clock time, unseeded randomness, map iteration order, and any
client state in loot selection. Same seed plus same ordered inputs must reproduce the same source
classes, categories, rolled payloads, loot entity ids, pickup/equip results, reconnect state, and
replay output.

## 6. Acceptance criteria

1. `make validate-shared` validates new depth loot bands, treasure class references, item def
   references, item template references, weights, and `-3+` reachability.
2. Dungeon level `-1`, `-2`, and `-3+` each resolve monster and chest reward sources through
   explicit shared data.
3. The spec and/or rules comments document that the `-1`, `-2`, `-3+` depth model is a temporary
   first pass to improve later.
4. Normal dungeon monsters can roll more than `cave_blade`; at least one pinned case produces a
   non-weapon equipment item.
5. Guarded chests have better equipment odds than normal monsters and can roll multiple v28
   equipment families.
6. By `-3+`, the configured dungeon/chest reward set can reach every v28 equipment template.
7. All v28 templates remain valid at level `1`; stat requirement and item-level progression are
   deferred.
8. `shared/golden/dungeon_equipment_drops.json` pins representative depth/source outcomes,
   including a monster drop and a guarded chest drop.
9. `shared/golden/treasure_class_rolls.json` or the new fixture pins varied rolled outcomes for at
   least: bow, shield, belt, armor, jewelry, potion/money-like item.
10. Go tests prove depth-band selection, template reachability, seeded roll determinism, and replay
    stability for varied equipment drops.
11. Godot golden tests validate any shared data-only fixture that client code consumes.
12. Protocol bot scenario `20_dungeon_equipment_drops.json` proves real generated dungeon play:
    descend, kill/open a configured source, pick up varied gear, equip representative categories,
    `/state`, reconnect, replay, and fresh-session persistence.
13. The bot scenario does not need to collect every category in one run; exhaustive category
    reachability is covered by validation and unit/golden tests.
14. `make ci` green.

## 7. Testing plan

1. `make validate-shared`
2. `cd server && go test ./internal/game/... -run 'Dungeon|Treasure|Loot|Equipment|Depth'`
3. `cd server && go test ./internal/http/... ./internal/replay/... -run 'Replay|Equipment|Loot'`
4. `make client-unit`
5. `make bot` - includes `20_dungeon_equipment_drops.json`
6. `make bot-client` only if the implementation changes client pickup/equipment assertions
7. `make ci`
8. Manual: `make play`, descend through several generated dungeon floors, kill mobs, open a guarded
   chest when one appears, pick up varied gear, equip a belt/shield/bow/jewelry item, restart from
   the main menu, and confirm character-owned items persist.

## 8. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | v29 uses a thin `-1`, `-2`, `-3+` depth-band model. | Gives real dungeon play a progression hook now without pretending this is the final economy. |
| 2 | The temporary depth model must be documented clearly. | Future slices should improve depth scaling rather than inherit the coarse bands by accident. |
| 3 | Chests have better equipment odds than monsters. | Guarded chests are rarer and already cost risk/time, so their reward source should feel better. |
| 4 | All v28 templates can drop at level `1`. | Keeps the slice focused on making the catalog reachable; real requirements/item levels are a later progression slice. |
| 5 | Bot proof is representative, not exhaustive. | One scenario should prove the real path; validation/goldens are better for exhaustive category coverage. |
| 6 | No protocol bump is expected. | Existing rolled item views already carry template, rarity, stat, slot, equipment, and persistence data. |

## 9. Open questions

Resolved before drafting:

| # | Question | Resolution |
|---|----------|------------|
| Q-1 | Should v29 implement true depth-banded treasure classes, or just broaden current classes? | Add a thin `-1`, `-2`, `-3+` depth-band config and document it as a future improvement point. |
| Q-2 | Should chests have better odds than monsters? | Yes. |
| Q-3 | Should all v28 templates be available immediately at level `1`? | Yes. |
| Q-4 | Should the bot require every category in one scenario? | No. Bot proves representative categories; validation/goldens cover reachability. |

## 10. Deferred follow-ups

- Real depth/item-level progression, including stronger tier bands beyond `-3+`.
- Better rarity curves by depth and source type.
- Magic Find and player/equipment-derived drop modifiers.
- Unique/set item catalogs and special item-specific drop rules.
- Boss-floor chest integration and reward pacing.
- Real gold wallet and stackable currency.
- Armor mitigation, shield block execution, crit/hit/attack-speed gameplay.
- Item comparison UI, loot filters, stash, vendors, crafting, and trade.
