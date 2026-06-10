# Spec: `item-templates-and-rolled-drops`

Status: Complete — `make ci` green on 2026-06-07
Branch: `feature/item-templates-and-rolled-drops`
Slice: v23 — item templates, rarities, rolled stats, and dungeon mob drops
Baseline: v22 `character-scoped-persistence`
Related:

- [`v8_spec-equipped-weapon-damage.md`](v8_spec-equipped-weapon-damage.md) — equipped weapon damage
- [`v13_spec-inventory-ui.md`](v13_spec-inventory-ui.md) — inventory UI and item tooltips
- [`v15_spec-item-visuals-and-loot-presentation.md`](v15_spec-item-visuals-and-loot-presentation.md) — shared item presentation metadata
- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) — dungeon mobs and current `no_drop`
- [`v22_spec-character-scoped-persistence.md`](v22_spec-character-scoped-persistence.md) — durable item instances with `rolled_stats`
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) — shared rules as data and deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) — dungeon progression and better loot loop
- [`../../PROGRESS.md`](../../PROGRESS.md)

## 1. Purpose

The game now has persistent character-owned item instances, but item identity is still fixed by
small static `item_def_id` records. Dungeon mobs are threatening, but v21 deliberately left them
without meaningful loot. This blocks the core ARPG loop: kill monsters, find randomized gear, equip
the stronger item, and keep that item on the character.

This slice introduces a server-authoritative item template and rolled-drop model. Shared item
templates define type, rarity pool, equip slot, requirements, base stats, rollable stat ranges, and
reserved special-effect identifiers. When a loot table references an item template, the Go Sim uses
its seeded RNG to roll a concrete item instance. The rolled instance is persisted through v22
character item storage and sent over the existing inventory/snapshot paths with additive item
instance fields.

The first vertical slice is intentionally narrow: one or two weapon templates, deterministic rarity
and stat rolls, dungeon mob drops, rolled weapon damage in combat, and inventory tooltip display.

## 2. Non-goals

- No full affix economy, prefix/suffix grammar, or procedural item-name generator beyond a simple
  display name assembled from template + rarity.
- No armor slots, rings, amulets, offhand, or stash UI.
- No crafting, vendors, gold, trade, item comparison UI, or loot filters.
- No special-effect execution in combat. v23 may store `effect_ids`, but active/passive effect
  behavior is deferred.
- No character level/attribute requirements beyond schema validation and, at most, a trivial
  `required_level <= 1` equip check.
- No production item art. Existing placeholder item presentation may be reused or minimally mapped.
- No Protobuf migration.

## 3. Files to create or modify

```text
docs/specs/v23_spec-item-templates-and-rolled-drops.md     - this slice contract
docs/plans/v23_2026-06-07-item-templates-and-rolled-drops.md - implementation plan
shared/rules/item_templates.v0.schema.json                 - item template contract
shared/rules/item_templates.v0.json                        - first rolled weapon templates
shared/rules/loot_tables.v0.schema.json                    - allow template entries
shared/rules/loot_tables.v0.json                           - dungeon mob rolled-drop table
shared/rules/monsters.v0.json                              - dungeon_mob uses rolled loot table
shared/protocol/session_snapshot.v1.schema.json            - additive item instance fields
shared/protocol/state_delta.v1.schema.json                 - additive item instance fields
shared/golden/item_rolls.json                              - deterministic item roll fixture
tools/validate_shared.py                                   - template, loot, and golden drift checks
server/internal/game/rules.go                              - parse templates and template loot entries
server/internal/game/sim.go                                - roll item instances and use rolled weapon damage
server/internal/game/game_test.go                          - roll, persistence payload, and combat tests
server/internal/store/models.go                            - expose rolled instance metadata if needed
server/internal/store/repos.go                             - persist/load rolled item metadata
client/scripts/inventory_panel.gd                          - tooltip rarity and rolled stats
client/tests/test_golden.gd                                - data-only item roll/template checks
tools/bot/run.py                                           - rolled item assertions if needed
tools/bot/scenarios/16_rolled_drops.json                   - end-to-end rolled loot proof
PROGRESS.md                                           - lifecycle update when v23 ships
```

## 4. Data shapes

### Item templates

New file: `shared/rules/item_templates.v0.json`.

Example shape:

```json
{
  "version": 0,
  "rarities": {
    "common": { "weight": 70, "stat_rolls": 1, "name_prefix": "Common" },
    "magic": { "weight": 25, "stat_rolls": 2, "name_prefix": "Magic" },
    "rare": { "weight": 5, "stat_rolls": 3, "name_prefix": "Rare" }
  },
  "templates": {
    "cave_blade": {
      "name": "Cave Blade",
      "category": "equipment",
      "item_type": "sword",
      "slot": "weapon",
      "equippable": true,
      "attack_mode": "melee",
      "reach": 1.5,
      "requirements": { "level": 1 },
      "base_stats": {
        "damage_min": 2,
        "damage_max": 4
      },
      "rollable_stats": [
        { "stat": "damage_min", "min": 0, "max": 1, "weight": 3 },
        { "stat": "damage_max", "min": 1, "max": 3, "weight": 3 },
        { "stat": "max_hp", "min": 1, "max": 3, "weight": 1 }
      ],
      "effect_pool": []
    }
  }
}
```

Template data is declarative. Roll behavior is a bounded catalog implemented in Go and checked by
goldens, not an expression language.

### Loot table template entries

`shared/rules/loot_tables.v0.json` gains template entries while retaining fixed item entries for
older scenarios:

```json
{
  "dungeon_mob_drop": {
    "entries": [
      { "item_template_id": "cave_blade", "weight": 1 }
    ]
  }
}
```

An entry must declare exactly one of `item_def_id` or `item_template_id`. Fixed entries keep current
behavior. Template entries roll a concrete item instance when dropped.

### Rolled item instance protocol view

`ItemView` remains backward-compatible and gains optional fields:

```json
{
  "item_instance_id": "1008",
  "item_def_id": "cave_blade",
  "item_template_id": "cave_blade",
  "display_name": "Magic Cave Blade",
  "rarity": "magic",
  "slot": "weapon",
  "equipped": false,
  "rolled_stats": {
    "damage_min": 3,
    "damage_max": 6,
    "max_hp": 2
  },
  "requirements": { "level": 1 },
  "effect_ids": []
}
```

For v23, `item_def_id` may equal the template id for client compatibility. The authoritative value
for combat is the rolled instance payload, not a client-side recomputation from template rules.

### Persistent rolled stats

v22 already provides `character_item_instances.rolled_stats` and session-start item snapshots.
v23 stores enough JSON in that payload to reconstruct the rolled item for fresh sessions and replay:

```json
{
  "item_template_id": "cave_blade",
  "display_name": "Magic Cave Blade",
  "rarity": "magic",
  "stats": {
    "damage_min": 3,
    "damage_max": 6,
    "max_hp": 2
  },
  "requirements": { "level": 1 },
  "effect_ids": []
}
```

## 5. Architecture and flow

```text
dungeon_mob killed
  -> Sim resolves monster loot table
  -> loot entry references item_template_id
  -> Sim rolls rarity through seeded RNG
  -> Sim selects and rolls stat ranges through seeded RNG
  -> Sim creates loot entity carrying rolled item payload
  -> player picks up loot through action_intent
  -> inventory_add sends rolled ItemView
  -> realtime runner persists rolled_stats on character item instance
  -> replay/fresh session reloads session-start or character item payload
  -> equip_intent equips the same item instance
  -> attack damage resolves rolled damage stats first, then falls back to static item rules
```

Determinism requirement: item roll order, selected rarity, selected stats, stat values, and item
instance IDs must reproduce from the same seed and ordered inputs. Go game logic must avoid
wall-clock time, unseeded randomness, and map iteration in roll selection.

## 6. Acceptance criteria

1. `make validate-shared` validates `item_templates.v0.json`, template loot entries, and fixed
   item entries.
2. `shared/golden/item_rolls.json` pins at least two deterministic cases: one common roll and one
   higher-rarity roll for `cave_blade`.
3. Go tests prove the same seed + inputs produce the same rarity, selected stats, stat values, and
   item instance payload.
4. `dungeon_mob` can drop a rolled item from an item template instead of `no_drop`.
5. `inventory_add`, `session_snapshot.inventory`, `/state`, replay timeline, and reconnect resume
   include rolled item fields.
6. Character persistence stores and reloads rolled item payloads through `rolled_stats`; fresh
   sessions preserve the same rolled item.
7. Equipping a rolled weapon makes authoritative attack damage use rolled `damage_min` /
   `damage_max` before falling back to static item definition damage.
8. Godot inventory tooltip displays rarity, display name, and rolled stats from the item instance.
9. Bot scenario `16_rolled_drops.json` proves dungeon mob kill, rolled drop pickup, equip, damage
   use, `/state`, reconnect, replay, and fresh-session persistence.
10. `make ci` green.

## 7. Testing plan

1. `make validate-shared`
2. `cd server && go test ./internal/game/... -run 'ItemTemplate|Rolled|Loot|WeaponDamage'`
3. `cd server && go test ./internal/http/... ./internal/replay/... -run 'Rolled|Item|Persistence|Replay'`
4. `make client-unit`
5. `make bot` — includes `16_rolled_drops.json`
6. `make ci`
7. Manual: `make play`, kill dungeon mobs, pick up a rolled weapon, inspect tooltip, equip it,
   restart play, confirm the same rolled item persists.

## 8. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | v23 starts with weapon templates only. | Weapon damage creates immediate player value and limits UI/equipment scope. |
| 2 | Rolled stats live on item instances, not templates. | Templates define ranges; concrete items need durable per-instance values. |
| 3 | `item_def_id` remains present in protocol. | Keeps existing client and bot paths working while adding template metadata. |
| 4 | Special effects are stored as IDs only. | Avoids mixing the item data model with an effect execution engine in one slice. |
| 5 | Dungeon mobs get the first template loot table. | v21 created threat; v23 adds the missing reward loop. |
| 6 | Roll logic is a bounded Go catalog with golden fixtures. | Preserves ADR-0001 determinism and shared-data discipline. |

## 9. Open questions

| # | Question | Default if unanswered |
|---|----------|----------------------|
| Q-1 | Should v23 include rarities `common`, `magic`, and `rare`, or only `common`/`magic`? | Use `common`, `magic`, and `rare`, with tests pinning common and magic/rare cases. |
| Q-2 | Should requirements be enforced immediately? | Validate schema and enforce only `required_level <= 1`; broader character stats are deferred. |
| Q-3 | Should rolled `max_hp` affect the player when equipped? | No. Display and persist it only; combat uses rolled weapon damage in v23. |
| Q-4 | Should dropped loot entities expose rolled fields before pickup? | Yes, enough for `/state` and future presentation, but client ground art may remain template/category based. |
| Q-5 | Should fixed `items.v0.json` be deleted? | No. Keep fixed items for consumables, currency, quest items, and legacy scenarios until migrated deliberately. |
