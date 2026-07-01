# v397 Plan — Item Archetype Library

Status: Ready for implementation
Goal: Generic archetype library, `equipment_category`, breaking `cave_*` rename, unified display-name
grammar, weapon elemental rolls, starter archetype pack, bot proof.
Architecture: Shared-rules + server roll/naming/combat paths; no protocol bump. Breaking template IDs;
`make db-reset` after merge.

Tech stack: shared JSON, Go sim, Python bot, GDScript tests (ID string updates only).

## Baseline

Builds on v396 `game-codex-chapters`. Spec:
[`docs/specs/v397_spec-item-archetype-library.md`](../specs/v397_spec-item-archetype-library.md).

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.schema.json` | `equipment_category`, elemental roll stats |
| Modify | `shared/rules/item_templates.v0.json` | Rename, categories, new archetypes, elemental pools |
| Modify | `shared/rules/set_items.v0.json` + schema | Drop piece `display_name`; update `base_template_id` |
| Modify | `shared/rules/unique_items.v0.json` + schema | Drop `display_name`; update `base_template_id` |
| Modify | `shared/rules/treasure_classes.v0.json` | New archetype weights; rename IDs |
| Modify | `shared/rules/worlds.v0.json` | `item_archetype_lab`; rename loot refs |
| Modify | `shared/assets/item_visuals.v0.json`, `item_presentations.v0.json` | Rename + new mappings |
| Modify | `shared/content/codex_index.v0.json` | Template ID strings |
| Modify | `server/internal/game/affix_names.go` | Grammar + elemental affix words |
| Modify | `server/internal/game/item_rolls.go` | Use unified naming helper for unique rolls |
| Modify | `server/internal/game/unique_chest.go`, `set_items.go` | Generated unique/set names; set match by piece id |
| Modify | `server/internal/game/damage_types.go`, `sim.go` | Elemental bonus damage + damage_type |
| Modify | `server/internal/game/rules.go` | `EquipmentCategory` on `ItemTemplateDef` |
| Modify | `tools/validate_shared.py` | Category + no `cave_` template keys |
| Modify | `shared/golden/*.json` | Regenerate contractual fixtures |
| Bulk | bot scenarios, client tests, server tests | `cave_*` → new IDs |
| Create | `tools/bot/scenarios/113_item_archetype_lab.json` | Extended proof |
| Create | `docs/as-built/v397_item-archetype-library.md` | As-built |

## Maintenance ratchet

- [x] `affix_names.go` extraction stays focused; no `sim.go` growth beyond elemental hook
- [x] `make maintainability` green

## Task 1 — Schema and templates

- [x] Add `equipment_category` enum to template schema
- [x] Add elemental bonus stats to rollable stat enum
- [x] Rename all `cave_*` templates per spec map; add 7 new archetypes
- [x] Add `equipment_category` to all equipment templates (incl. starters)

```bash
make validate-shared
```

## Task 2 — Catalogs and assets

- [x] Update `set_items`, `unique_items` base_template_id; optional display_name removal from schema
- [x] Treasure classes: every new archetype in depth-3+ pool
- [x] `item_archetype_lab` world preset
- [x] Presentation/visual mappings for renamed + new IDs
- [x] Codex template references

```bash
make validate-shared
make validate-assets
```

## Task 3 — Server naming and combat

- [x] Rewrite `affixDisplayName`: common = archetype only; magic/rare = affix + archetype
- [x] `rolledUniqueDisplayName`, `rolledSetDisplayName`: affix + archetype + of + effect/set
- [x] Named unique + set payload builders use grammar (not catalog display_name)
- [x] Set equip matching by piece id, not display name
- [x] Elemental bonus stats in weapon damage range + `playerWeaponDamageTypeForSlot`
- [x] Focused tests: naming grammar, cold weapon hit

```bash
cd server && go test ./internal/game/... -run 'AffixName|Archetype|ElementalWeapon|SetItem' -count=1
```

## Task 4 — Bulk reference migration

- [x] Replace `cave_*` template IDs across goldens, bots, client tests, server tests
- [x] Regenerate goldens where `-update` paths exist

```bash
make validate-shared
cd server && go test ./internal/game/... -run Golden -count=1  # if applicable
```

## Task 5 — Bot scenario

- [x] `113_item_archetype_lab.json` — common long sword name, magic cold affix name, set/unique grammar
- [x] `ci_tier: extended`

```bash
make bot scenario=item_archetype_lab
```

## Task 6 — Docs and finish

- [x] As-built, lifecycle row, PROGRESS.md
- [x] Plan checkboxes complete

Autoloop focused verification:

```bash
make validate-shared
make validate-assets
cd server && go test ./internal/game/... -run 'AffixName|Archetype|ElementalWeapon|SetItem' -count=1
make bot scenario=item_archetype_lab
```

Post-loop: `make ci` (autoloop batch gate).

Local play note: `make db-reset` after pulling template ID changes.
