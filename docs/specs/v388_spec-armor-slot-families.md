# v388 Spec — Armor Slot Families

Status: Complete
Date: 2026-06-30
Codename: armor-slot-families

## Purpose

Add **three armor families per slot** for `head`, `chest`, `gloves`, and `boots` (12 templates total).
Each family expresses a build tradeoff through `requirements`, `base_stats` (armor tier identity),
and **weighted `rollable_stats`** (penalties and bonuses are rolled, not fixed in `base_stats`).

| Slot | Light | Medium (existing `cave_*`) | Heavy / specialist |
|------|-------|---------------------------|-------------------|
| Head | Leather cap — dex-biased rolls | Cave Helm | Tiara — `magic` requirement, skill-affix roll bias |
| Chest | Leather vest — low armor, positive move rolls | Cave Mail (chainmail) | Full plate — ~2× mail armor base, negative move rolls (−25..−10) |
| Gloves | Cloth wraps — skill/cast roll bias | Cave Gloves | Gauntlets — higher armor base, negative attack-speed rolls |
| Boots | Soft boots — positive move rolls | Cave Boots | Plate boots — higher armor base, negative move rolls |

Existing `cave_helm`, `cave_mail`, `cave_gloves`, and `cave_boots` remain the **medium** tier so set
items and current scenarios stay stable.

## Non-goals

- Belt, ring, amulet, shield, or weapon families
- New stats, affix grammar, `armor_family` protocol field, or mystery-seller filtering
- Hard class equip locks
- Per-family client icons or production art (borrow existing slot presentation families)
- Full depth-band treasure-class rebalance (add to representative pools + lab world only)

## Acceptance criteria

- [ ] Eight new templates plus four unchanged medium templates; 12 armor-slot templates total
- [ ] `cave_full_plate` `base_stats.armor` is ~2× `cave_mail`; move penalty only in `rollable_stats`
- [ ] `cave_tiara` requires `magic`; skill-related stats appear in its roll pool with higher weight than armor
- [ ] New templates mapped in `item_visuals.v0.json` / `item_presentations.v0.json` (borrow families)
- [ ] New templates appear in at least one treasure class and an `armor_slot_families_lab` world preset
- [ ] Extended bot scenario: tiara requirement gate, plate rolled move penalty, tiara skill affix on pinned drop
- [ ] Focused Go test for template armor ratio and roll-pool contracts
- [ ] `make validate-shared` passes

## Scope and likely files

- `shared/rules/item_templates.v0.json`
- `shared/rules/treasure_classes.v0.json`
- `shared/rules/worlds.v0.json` — `armor_slot_families_lab`
- `shared/assets/item_visuals.v0.json`, `shared/assets/item_presentations.v0.json`
- `server/internal/game/*_test.go` — focused armor-family tests
- `tools/bot/scenarios/110_armor_slot_families_lab.json`
- Docs: plan, as-built, lifecycle

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ArmorSlotFamil' -count=1
make bot scenario=armor_slot_families_lab
```

Visual (optional): `make bot-visual scenario=armor_slot_families_lab`

## Asset decision

- **Adopt:** existing head/chest/gloves/boots placeholder assets and presentation shapes
- **Borrow:** same `family` mapping as sibling `cave_*` templates per slot
- **Reject:** new icons, GLBs, or external plugins

## Open questions

None — Q1–Q5 resolved in `/next` brief (rolled penalties, defaults, one slice, defer icons).
