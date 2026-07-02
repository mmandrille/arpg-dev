# v405 Spec — Class Specialist Gear

Status: Approved
Date: 2026-07-02
Codename: `class-specialist-gear`
Baseline: v404 `character-mercenary-hire` complete

Related:

- [`v70_spec-class-skill-and-item-gates.md`](v70_spec-class-skill-and-item-gates.md) — fixed `class_required` weapons
- [`v388_spec-armor-slot-families.md`](v388_spec-armor-slot-families.md) — armor archetype tradeoffs
- [`v397_spec-item-archetype-library.md`](v397_spec-item-archetype-library.md) — `equipment_category` taxonomy; deferred off-hand book
- [`../adr/0014-core-progression-and-endgame-design-rules.md`](../adr/0014-core-progression-and-endgame-design-rules.md)

## Purpose

Ship **one class-locked rolled equipment archetype per class**. Each specialist item trades
**armor for class-themed stats** at common rarity and is hard-gated to its class at equip time
(same `class_requirement_not_met` contract as v70 fixed weapons).

| Class | Template | Slot | Stat identity (no armor) |
|-------|----------|------|--------------------------|
| Barbarian | Skull Face | `head` | `damage_max` +1..+5 at ilvl 1 common |
| Sorcerer | Magic Book | `off_hand` | `max_mana` / `mana_regen_per_10_seconds` |
| Paladin | Scepter | `main_hand` (1H melee) | `skill_damage_percent` (+ modest weapon damage) |
| Rogue | Shadow Mask | `head` | `evade_chance` (+ `dex` at magic+) |
| Ranger | Hunting Quiver | `belt` | `attack_speed_percent` (+ `dex` at magic+) |

Templates use `equipment_category: class_specialist` and optional template `class_required`.

## Non-goals

- Affix grammar, unique behavior hooks, or set packages for specialist gear
- Protocol/schema version bump
- Production art (borrow existing slot families)
- Full treasure-class economy rebalance
- `class_affinity` rework of existing weapons
- Requirement-status UI for class lock (equip reject + tests prove the gate)

## Acceptance criteria

- [ ] `item_templates.v0.schema.json` adds `equipment_category: class_specialist`, template
  `class_required`, and `item_type: book` for Magic Book presentation
- [ ] Five specialist templates exist; none include `armor` in `base_stats` or `rollable_stats`
- [ ] Barbarian Skull Face common roll at ilvl 1 includes `damage_max` in +1..+5
- [ ] Sorcerer Magic Book rolls mana sustain stats only (no armor)
- [ ] Paladin Scepter is one-handed melee with `skill_damage_percent` in the common roll pool
- [ ] Rolled items with template `class_required` reject wrong-class equip with
  `class_requirement_not_met`
- [ ] Templates in lab world + representative treasure class entry; `make validate-shared` passes
- [ ] Focused Go tests for roll pools, no-armor contract, and equip gates
- [ ] Extended bot scenario proves barbarian skull equip + cross-class reject on sorcerer book

## Scope and likely files

- `shared/rules/item_templates.v0.json` + schema
- `shared/rules/treasure_classes.v0.json`, `shared/rules/worlds.v0.json`
- `shared/assets/item_visuals.v0.json`, `item_presentations.v0.json`
- `server/internal/game/rules.go`, `sim.go`, tests
- `tools/validate_shared.py`
- `tools/bot/scenarios/114_class_specialist_gear_lab.json`
- Docs: plan, as-built, lifecycle

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ClassSpecialist' -count=1
make bot scenario=class_specialist_gear_lab
```

Visual (optional): `make bot-visual scenario=class_specialist_gear_lab`

## Asset decision

- **Adopt:** existing helm/belt/staff/off-hand presentation families
- **Borrow:** `book` uses `staff` presentation shape until production art
- **Reject:** new GLBs, external plugins

## Open questions

Resolved in `/autoloop` batch gate:

- Q1 hard class lock (not affinity)
- Q2 all five classes in one slice
- Q3 paladin **Scepter** 1H with `skill_damage_percent` (not amulet)
- Q4 ilvl scaling via existing finalize path
- Q5 `equipment_category: class_specialist`
