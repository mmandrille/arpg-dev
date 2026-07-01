# v397 Spec — Item Archetype Library

Status: Approved
Date: 2026-07-01
Codename: item-archetype-library
Baseline: v396 `game-codex-chapters` complete

Related:

- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) — template + roll model
- [`v191_spec-affix-name-grammar.md`](v191_spec-affix-name-grammar.md) — first affix prefix grammar (superseded naming rules here)
- [`v388_spec-armor-slot-families.md`](v388_spec-armor-slot-families.md) — armor archetype stats (light/medium/heavy *themes*, not schema tiers)
- [`v394_spec-weapon-slot-families.md`](v394_spec-weapon-slot-families.md) — sword/bow archetype stats
- [`v285_spec-damage-type-readability.md`](v285_spec-damage-type-readability.md) — client damage-type labels
- [`../adr/0014-core-progression-and-endgame-design-rules.md`](../adr/0014-core-progression-and-endgame-design-rules.md)

## Purpose

Replace biome-themed `cave_*` equipment with a **generic archetype library** organized by
equipment category. Each template is a distinct gear archetype (Long Sword, Rapier, Hunter Bow,
Buckler, Helm, Ring, …) — not a light/medium/heavy tier field.

Ship in the same slice:

1. **Shared taxonomy** — `equipment_category` on every equipment template.
2. **Template ID + display name cleanup** — remove `cave_` prefix; archetype slug = template key.
3. **Unified display-name grammar** for rolled equipment (common → magic/rare → unique/set).
4. **Weapon elemental rolls** — flat bonus elemental damage on weapons, wired to affix words and
   basic-attack `damage_type` when an elemental bonus is present.
5. **Starter archetype pack** — add missing archetypes from the owner catalog; every new archetype
   appears in treasure classes (borrow same-slot presentation assets).

### Library categories (`equipment_category`)

| Category | Covers | Examples |
|----------|--------|----------|
| `weapon_1h` | One-handed melee or ranged weapons | Long Sword, Dagger, Rapier, Falchion, Hammer, Morningstar, Wand |
| `weapon_2h` | Two-handed weapons | Great Sword, Warhammer, Bow, Hunter Bow, War Bow, Staff, Axe |
| `off_hand` | Off-hand gear | Buckler, Shield (book deferred) |
| `gear` | Armor slots incl. belt | Head, chest, gloves, boots, belt |
| `jewelry` | Rings and amulets only | Ring, Amulet — **no subtypes** (no signet/band/charm variants) |

`item_type` remains the presentation/combat family (`sword`, `bow`, `helm`, `shield`, …).
`equipment_category` is the library shelf; `item_type` is the visual/socket hook.

### Display-name grammar (authoritative)

Names are assembled server-side into `ItemRollPayload.display_name` at roll time (and on reroll).
No rarity word (`Common`, `Rare`, …) appears in the final name.

| Rarity | Pattern | Example |
|--------|---------|---------|
| `common` | `{Archetype}` | `Long Sword` |
| `magic`, `rare` | `{Affix} {Archetype}` | `Freezing Long Sword` |
| `unique` | `{Affix} {Archetype} of {Effect}` | `Freezing Long Sword of Everburning Wound` |
| `set` | `{Affix} {Archetype} of {SetName}` | `Stalwart Helm of Verdant Vanguard` |

Rules:

- `{Archetype}` = template `name` (title case, no biome prefix).
- `{Affix}` = deterministic prefix from rolled stat families (`affix_names.go`); omit when no
  qualifying roll gain (falls back to archetype only, same as common).
- `{Effect}` = `unique_effects.v0.json` `display_name` for the item's primary `effect_ids[0]`.
- `{SetName}` = parent set catalog `display_name`.
- **All** unique and set payloads use this grammar — remove reliance on authored piece
  `display_name` in `unique_items.v0.json` and `set_items.v0.json` (keep piece `id` and
  `base_template_id`; names are generated).

### Elemental weapon damage

Add rollable flat stats on **weapon templates only**:

| Stat | Affix word (priority) | Combat |
|------|----------------------|--------|
| `bonus_cold_damage` | Freezing | Adds flat damage; basic attack `damage_type` = `cold` when this is the dominant elemental bonus |
| `bonus_fire_damage` | Burning | `fire` |
| `bonus_lightning_damage` | Shocking | `lightning` |
| `bonus_poison_damage` | Venomous | `poison` |

Dominant elemental = highest rolled bonus among element stats; ties break by stat key order.
Physical `damage_min`/`damage_max` still apply; elemental bonus adds flat damage to the rolled hit
range before resistance. When no elemental bonus > 0, basic attacks remain `force` (current
behavior for templates).

Elemental stats appear in weapon `rollable_stats` pools (magic+), with modest weights — not on
armor/jewelry in this slice.

## Non-goals

- `gear_tier` / light-medium-heavy-specialist schema field (v388/v394 tradeoffs stay in template
  stats, not a tier enum)
- Jewelry archetype subtypes (signet, band, charm, …)
- Off-hand **book** template (future `off_hand` expansion)
- Full affix grammar (multi-affix prefix + suffix, localization catalog)
- Procedural item **IDs**; market/search still keyed by `item_template_id` + instance stats
- Stash migration / `cave_*` alias layer — **breaking rename**; `make db-reset` before play test
- Production art; new archetypes **borrow** existing slot `family` icons/GLBs
- Full depth-band treasure-class rebalance beyond adding entries for every archetype
- Protocol schema version bump (`display_name` string shape changes only)

## Template migration map

Rename existing `cave_*` templates (stats/requirements unchanged unless noted). Template key =
archetype slug.

### Weapons — 1H (`weapon_1h`)

| Old ID | New ID | Display name |
|--------|--------|--------------|
| `cave_blade` | `long_sword` | Long Sword |
| `cave_rapier` | `rapier` | Rapier |
| `cave_heavy_blade` | `falchion` | Falchion |
| `cave_war_sword` | `war_sword` | War Sword |

### Weapons — 1H (new)

| New ID | Display name | Notes |
|--------|--------------|-------|
| `dagger` | Dagger | Fast, low base damage, dex rolls; borrow `sword` / dagger-speed tuning |
| `hammer` | Hammer | Str-biased 1H blunt; borrow `sword` presentation family |
| `morningstar` | Morningstar | Str + armor-pierce roll bias; borrow `sword` family |
| `wand` | Wand | Magic req, skill roll bias; borrow `staff` icon/GLB scaled for 1H |

### Weapons — 2H (`weapon_2h`)

| Old ID | New ID | Display name |
|--------|--------|--------------|
| `cave_greatsword` | `great_sword` | Great Sword |
| `cave_bow` | `bow` | Bow |
| `cave_hunting_bow` | `hunter_bow` | Hunter Bow |
| `cave_war_bow` | `war_bow` | War Bow |

### Weapons — 2H (new)

| New ID | Display name | Notes |
|--------|--------------|-------|
| `warhammer` | Warhammer | Heavy 2H blunt; borrow `greatsword` / axe presentation |

Starter templates (`starter_*`, `training_bow`, class weapons) keep IDs; add `equipment_category`
where missing. No requirement to rename starters in this slice.

### Off-hand (`off_hand`)

| Old ID | New ID | Display name |
|--------|--------|--------------|
| `cave_shield` | `shield` | Shield |

| New ID | Display name | Notes |
|--------|--------------|-------|
| `buckler` | Buckler | Lower armor/block base, dex/evade roll bias; borrow `shield` family |

### Gear (`gear`)

| Old ID | New ID | Display name |
|--------|--------|--------------|
| `cave_helm` | `helm` | Helm |
| `cave_leather_cap` | `leather_cap` | Leather Cap |
| `cave_tiara` | `tiara` | Tiara |
| `cave_mail` | `mail` | Mail |
| `cave_leather_vest` | `leather_vest` | Leather Vest |
| `cave_full_plate` | `full_plate` | Full Plate |
| `cave_gloves` | `gloves` | Gloves |
| `cave_cloth_wraps` | `cloth_wraps` | Cloth Wraps |
| `cave_gauntlets` | `gauntlets` | Gauntlets |
| `cave_boots` | `boots` | Boots |
| `cave_soft_boots` | `soft_boots` | Soft Boots |
| `cave_plate_boots` | `plate_boots` | Plate Boots |
| `cave_belt` | `belt` | Belt |
| `cave_pack_belt` | `war_girdle` | War Girdle |

| New ID | Display name | Notes |
|--------|--------------|-------|
| `sash` | Sash | Light belt; move/hotbar roll bias; borrow `belt` family |

### Jewelry (`jewelry`)

| Old ID | New ID | Display name |
|--------|--------|--------------|
| `cave_ring` | `ring` | Ring |
| `cave_amulet` | `amulet` | Amulet |

## Acceptance criteria

### Shared rules & schema

- [ ] `item_templates.v0.schema.json` requires `equipment_category` enum on every equipment template
- [ ] Zero `cave_*` keys in `item_templates.v0.json`
- [ ] Rollable stat enum includes `bonus_cold_damage`, `bonus_fire_damage`, `bonus_lightning_damage`,
      `bonus_poison_damage` (weapons only in data)
- [ ] All weapon templates include at least one elemental stat in `rollable_stats` (low weight ok)
- [ ] `unique_items.v0.json` — `display_name` removed or ignored; `base_template_id` updated to new IDs
- [ ] `set_items.v0.json` — piece `display_name` removed or ignored; `base_template_id` updated
- [ ] Every **new** archetype (`dagger`, `hammer`, `morningstar`, `wand`, `warhammer`, `buckler`,
      `sash`) appears in at least one depth-3+ treasure class entry
- [ ] `item_visuals.v0.json` / `item_presentations.v0.json` — all renamed + new IDs mapped (borrow
      same-slot family)
- [ ] `shared/content/codex_index.v0.json` — template references updated to generic names/IDs
- [ ] `make validate-shared` passes

### Naming

- [ ] Common rolled item name equals template `name` only (no rarity prefix)
- [ ] Magic/rare name = `{Affix} {Archetype}` when affix applies; no `Magic`/`Rare` in string
- [ ] Unique chest + named unique payloads use `{Affix} {Archetype} of {Effect}`
- [ ] Set piece payloads use `{Affix} {Archetype} of {SetName}`
- [ ] Set piece equip matching uses piece `id` (or template + set), **not** generated display name

### Combat

- [ ] Weapon with `bonus_cold_damage` > 0 in rolled stats deals basic-attack damage with
      `damage_type: cold` (dominant-element rule)
- [ ] Elemental bonus flat-added to attack damage range before resistance
- [ ] Focused Go test: pinned cold roll → `Freezing` affix name + cold damage type on hit

### Tests & bot

- [ ] Goldens referencing `cave_*` regenerated (`item_rolls`, `shop_offers`, `dungeon_equipment_drops`,
      etc. where template IDs are contractual)
- [ ] Extended bot scenario `item_archetype_lab` (or update family lab scenarios): common long sword
      plain name, magic cold long sword affix name, unique/set grammar spot-check
- [ ] `make db-reset` documented in plan for local verification after merge

## Scope and likely files

- `shared/rules/item_templates.v0.json` + `.schema.json`
- `shared/rules/set_items.v0.json`, `unique_items.v0.json`
- `shared/rules/treasure_classes.v0.json`, `worlds.v0.json` — `item_archetype_lab` preset
- `shared/assets/item_visuals.v0.json`, `item_presentations.v0.json`
- `shared/content/codex_index.v0.json`
- `shared/golden/*.json` — ID/name drift
- `server/internal/game/affix_names.go` — grammar rewrite + elemental words
- `server/internal/game/item_rolls.go`, `item_reroll.go`
- `server/internal/game/unique_chest.go`, `set_items.go` — generated names
- `server/internal/game/damage_types.go` or `sim.go` — elemental weapon damage type + bonus damage
- `server/internal/game/rules.go` — `ItemTemplateDef.EquipmentCategory`
- `tools/validate_shared.py` — category + rename invariants
- Bulk update: bot scenarios, client tests, `showme` probes (~200 `cave_*` references)
- Docs: plan, as-built, lifecycle

## Test and bot proof

```bash
make validate-shared
make db-reset   # required after template ID breaking rename
cd server && go test ./internal/game/... -run 'AffixName|Archetype|ElementalWeapon' -count=1
make bot scenario=item_archetype_lab
```

Optional visual: `make bot-visual scenario=item_archetype_lab`

## Asset decision

- **Adopt:** existing per-slot presentation families (`sword`, `bow`, `shield`, `helm`, `belt`, `ring`, …)
- **Borrow:** new archetypes reuse sibling slot GLB/icon with optional color/accent override in
  `item_presentations.v0.json` (same pattern as `rapier` vs `long_sword`)
- **Reject:** new art pipeline, external plugins, per-archetype production models

## Resolved decisions (owner 2026-07-01)

| # | Decision |
|---|----------|
| Q1 | Ship taxonomy + rename + naming grammar together |
| Q2 | Each existing template = distinct archetype; drop `cave_` prefix only |
| Q3 | Add starter-pack archetypes; **all** new entries in treasure classes |
| Q4 | `equipment_category` + keep `item_type` |
| Q5 | Belt is `gear` |
| Q6 | Uniques and sets use `{Affix} {Archetype} of {Effect/SetName}` — no authored piece names |
| Q7 | Elemental bonus damage rollable on weapons + combat + affix words |
| Q8 | Breaking rename; `make db-reset` |
| Q9 | No ring/amulet subtypes |

## Open questions

None — ready for `/plan`.
