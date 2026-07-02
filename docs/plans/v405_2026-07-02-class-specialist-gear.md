# v405 Plan — Class Specialist Gear

Status: Complete
Goal: Add five class-locked rolled specialist templates that trade armor for class-themed stats.
Architecture: Extend `item_templates` with `class_specialist` category and template `class_required`; enforce class gate in `itemClassAllowed` for rolled payloads; data-only roll pools with no armor; lab world + extended bot proof.
Tech stack: shared JSON, Go sim, Python validate/bot.

## Baseline and shortcut decision

Reuses v70 `class_requirement_not_met` equip path and v397 `equipment_category` taxonomy.
**Assets:** borrow helm/belt/staff shapes; Magic Book borrows `book` presentation family (staff shape).

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.schema.json` | `class_specialist`, `class_required`, `book` item_type |
| Modify | `shared/rules/item_templates.v0.json` | Five specialist templates |
| Modify | `shared/rules/treasure_classes.v0.json` | Lab TC entries |
| Modify | `shared/rules/worlds.v0.json` | `class_specialist_gear_lab` |
| Modify | `shared/assets/item_presentations.v0.json` + schema | `book` family |
| Modify | `shared/assets/item_visuals.v0.json` | Template → family mapping |
| Modify | `server/internal/game/rules.go` | Load/validate `class_required` on templates |
| Modify | `server/internal/game/class_item_affinities.go` | Rolled template class equip gate |
| Create | `server/internal/game/class_specialist_gear_test.go` | Roll + gate contracts |
| Modify | `tools/validate_shared.py` | Specialist validation |
| Create | `tools/bot/scenarios/114_class_specialist_gear_lab.json` | Extended bot proof |
| Docs | spec, as-built, lifecycle, PROGRESS | Close-out |

## Maintenance ratchet

Hotspot files: `rules.go`, `sim.go`, `item_templates.v0.json`, `validate_shared.py` — touch-to-shrink only; moved `itemClassAllowed` to `class_item_affinities.go`.

## Task 1 — Shared contracts

- [x] Schema: `class_specialist`, `class_required`, `book`
- [x] Five templates in `item_templates.v0.json`
- [x] Presentations + visuals for new template ids
- [x] Treasure class + lab world preset

## Task 2 — Server gates and tests

- [x] `ItemTemplateDef.ClassRequired` + rules validation
- [x] `itemClassAllowed` checks template class for rolled items
- [x] `class_specialist_gear_test.go`
- [x] Empty `base_stats` roll nil-map fix

## Task 3 — Bot scenario

- [x] `114_class_specialist_gear_lab.json` (`ci_tier: extended`)

## Task 4 — Lifecycle docs

- [x] `PROGRESS.md`, lifecycle row, as-built

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'ClassSpecialist' -count=1`
- [x] `make bot scenario=class_specialist_gear_lab`
- [x] `make maintainability`
