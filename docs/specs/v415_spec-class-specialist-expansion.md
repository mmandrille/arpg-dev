# v415 Spec — Class Specialist Expansion

Status: Approved for implementation
Date: 2026-07-03
Codename: class-specialist-expansion
Baseline: v414 `class-survival-skills` complete

Related:

- [`v405_spec-class-specialist-gear.md`](v405_spec-class-specialist-gear.md)
- [`../adr/0014-core-progression-and-endgame-design-rules.md`](../adr/0014-core-progression-and-endgame-design-rules.md)

## Purpose

Expand v405's one-slot-per-class specialist catalog to **three class-locked pieces per class**
(two new templates per class) with build-defining roll pools, depth-1 treasure-class drops, and
server-authored **class rows in `requirement_status`** so inventory/stash/shop tooltips show why
wrong-class characters cannot equip specialist gear.

## Non-goals

- Affix grammar, unique/set specialist packages, vendor/mystery pools
- Production art (borrow existing presentation families)
- Class-locked generic weapons or jewelry variety slice
- Protocol version bump beyond optional `class_id` on existing v8 requirement rows

## Acceptance criteria

### Shared rules

- [ ] Ten new `class_specialist` templates (two per class) with `class_required`, no `armor` rolls
- [ ] Each class has ≥3 specialist templates across ≥2 slots (15 total including v405)
- [ ] New templates in `dungeon_mob_tc_depth_1` treasure class and lab world loot spawns
- [ ] `item_presentations.v0.json` and `item_visuals.v0.json` cover all new template IDs
- [ ] `make validate-shared` passes

### Server

- [ ] `requirement_status` includes a `stat: "class"` row with `class_id` when item is class-locked
- [ ] `requirements_met` and `equip_preview.requirements_met` false when class mismatch
- [ ] Equip still rejects with `class_requirement_not_met` (existing path)
- [ ] Go tests cover class requirement status + template validation count

### Client

- [ ] Tooltips show localized class requirement line (green/red) from `requirement_status`
- [ ] Headless unit test covers class requirement line formatting

### Bot

- [ ] Extended `class_specialist_gear_lab`: barbarian sees unmet class row on paladin scepter;
      equips new barbarian specialist piece

## Scope and files

- `shared/rules/item_templates.v0.json`, `treasure_classes.v0.json`, `worlds.v0.json`
- `shared/assets/item_presentations.v0.json`, `item_visuals.v0.json`
- `shared/protocol/state_delta.v8.schema.json`, `session_snapshot.v8.schema.json`
- `server/internal/game/item_class_requirements.go`, `sim.go`, tests
- `client/scripts/item_requirement_views.gd`, `inventory_panel.gd`, tests
- `tools/bot/scenarios/114_class_specialist_gear_lab.json`

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ClassSpecialist|ClassRequirement' -count=1
godot --headless --path client --script res://tests/test_item_requirement_views.gd
make bot scenario=class_specialist_gear_lab
```

## Asset decision

- **Adopt:** existing procedural icon families (`gloves`, `belt`, `mail`, `ring`, `shield`, `boots`, `amulet`)
- **Borrow:** placeholder GLB mounts from same-slot generic gear
- **Reject:** external art pipeline

## Open questions

None — defaults from `/next` brief: two new specialists per class, dungeon drops only, hard class lock.
