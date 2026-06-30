# v388 Plan — Armor Slot Families

Status: Ready for implementation
Goal: Add 12 armor-slot templates (3 families × 4 slots) with data-driven tradeoffs and bot proof.
Architecture: Pure shared-rules extension; existing roll/equip/requirement paths unchanged. Tradeoff
penalties live in `rollable_stats`; armor tier identity in `base_stats`. Medium `cave_*` templates
unchanged for set-item compatibility.

Tech stack: shared JSON, Go rules validation, Python protocol bot.

## Baseline and shortcut decision

Builds on v387 class-weapon-affinities (template + lab-world pattern). Reuses placeholder visuals per
slot; no client code changes required.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.json` | 8 new templates |
| Modify | `shared/rules/treasure_classes.v0.json` | Add families to depth-3+ pool |
| Modify | `shared/rules/worlds.v0.json` | `armor_slot_families_lab` preset |
| Modify | `shared/assets/item_visuals.v0.json` | Visual mappings for new templates |
| Modify | `shared/assets/item_presentations.v0.json` | Presentation family mappings |
| Modify | `server/internal/game/item_templates_test.go` | New focused tests |
| Create | `tools/bot/scenarios/110_armor_slot_families_lab.json` | Extended proof |
| Create | `docs/as-built/v388_armor-slot-families.md` | As-built summary |

## Maintenance ratchet

Hotspot files touched: `item_templates.v0.json`, `worlds.v0.json` (grandfathered).

- [ ] No new over-600-line source files
- [ ] Touched grandfathered files stay within baseline

```bash
make maintainability
```

## Task 1 — Item templates

Files:
- Modify: `shared/rules/item_templates.v0.json`

- [ ] Add 8 family templates with requirements, base armor tiers, and roll pools per spec
- [ ] Step 1.1: validate shared rules

```bash
make validate-shared
```

## Task 2 — Loot and lab world

Files:
- Modify: `shared/rules/treasure_classes.v0.json`
- Modify: `shared/rules/worlds.v0.json`

- [ ] Add new template weights to `dungeon_mob_tc_depth_3_plus`
- [ ] Add `armor_slot_families_lab` with pinned loot entities for bot proof

```bash
make validate-shared
```

## Task 3 — Presentation mappings

Files:
- Modify: `shared/assets/item_visuals.v0.json`
- Modify: `shared/assets/item_presentations.v0.json`

- [ ] Map each new template id to the same slot family as its medium sibling

```bash
make validate-assets
```

## Task 4 — Go tests

Files:
- Create: `server/internal/game/item_templates_test.go`

- [ ] Test plate vs mail armor ratio from loaded rules
- [ ] Test family roll pools include expected stats with rule-derived bounds

```bash
cd server && go test ./internal/game/... -run 'ArmorSlotFamil' -count=1
```

## Task 5 — Bot scenario

Files:
- Create: `tools/bot/scenarios/110_armor_slot_families_lab.json`

- [ ] Tiara requirement reject then magic allocate + equip
- [ ] Full plate pickup asserts rolled `movement_speed_percent` in template range
- [ ] Tiara pinned seed asserts skill affix in `rolled_stats`
- [ ] Equip plate vs mail: derived `movement_speed` lower with plate when penalty rolled

```bash
make bot scenario=armor_slot_families_lab
```

## Task 6 — Lifecycle docs

- [ ] `docs/as-built/v388_armor-slot-families.md`
- [ ] `PROGRESS.md` + `slice-lifecycle.md` on `/finish`

## Final verification

- [ ] `make maintainability`
- [ ] `make validate-shared`
- [ ] `cd server && go test ./internal/game/... -run 'ArmorSlotFamil' -count=1`
- [ ] `make bot scenario=armor_slot_families_lab`
