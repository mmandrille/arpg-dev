# v410 Plan — Polearm Archetypes and Family Icons

Status: Ready for implementation
Goal: Add spear-line weapons with extended reach and distinct per-family inventory icons.
Architecture: Data-driven templates in shared rules; reach already authoritative in Go sim;
client icons are presentation-only new shapes in `ItemIconDrawer` backed by `item_presentations`
families keyed by `item_type`.
Tech stack: shared JSON, Go sim tests, Godot client drawer, Python protocol bot.

## Baseline and shortcut decision

Builds on v397 archetype library and v394 weapon families. Adopt procedural icons; borrow GLBs.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.json` | New templates + hammer/morningstar item_type |
| Modify | `shared/rules/item_templates.v0.schema.json` | item_type enum |
| Modify | `shared/rules/treasure_classes.v0.json` | Add four templates to depth-3+ pools |
| Modify | `shared/rules/worlds.v0.json` | `polearm_archetype_lab` preset |
| Modify | `shared/assets/item_presentations.v0.json` + schema | Families + item entries |
| Modify | `shared/assets/item_visuals.v0.json` | GLB borrow mappings |
| Modify | `client/scripts/item_icon_drawer.gd` | New draw shapes |
| Create | `client/tests/test_item_icon_drawer.gd` | Shape resolution smoke |
| Create | `server/internal/game/polearm_reach_test.go` | Reach proof |
| Create | `tools/bot/scenarios/119_polearm_archetype_lab.json` | Extended bot proof |
| Create | `docs/as-built/v410_polearm-archetypes-and-family-icons.md` | As-built |

## Maintenance ratchet

Hotspot files touched: `item_icon_drawer.gd` (~170 lines), `item_templates.v0.json` (grandfathered).

- [ ] No coordinator growth; new test file stays under 600 lines.

## Task 1 — Shared templates and treasure classes

- [ ] Add `spear`, `short_spear`, `halberd`, `mace` templates
- [ ] Update `hammer` / `morningstar` `item_type`
- [ ] Extend schema enum
- [ ] Treasure classes + lab world

```bash
make validate-shared
```

## Task 2 — Presentation and visuals

- [ ] New families and shapes in `item_presentations`
- [ ] `item_visuals` entries for new templates
- [ ] Update hammer/morningstar presentation shapes

```bash
make validate-shared
```

## Task 3 — Go reach tests

- [ ] `polearm_reach_test.go`

```bash
cd server && go test ./internal/game/... -run Polearm -count=1
```

## Task 4 — Client icon shapes

- [ ] `item_icon_drawer.gd` + headless test

```bash
godot --headless --path client --script res://tests/test_item_icon_drawer.gd
```

## Task 5 — Bot scenario

- [ ] `119_polearm_archetype_lab.json`

```bash
make bot scenario=polearm_archetype_lab
```

## Task 6 — Lifecycle docs

- [ ] As-built, PROGRESS, lifecycle row

## Final verification

- [ ] `make validate-shared`
- [ ] `cd server && go test ./internal/game/... -run Polearm -count=1`
- [ ] `godot --headless --path client --script res://tests/test_item_icon_drawer.gd`
- [ ] `make bot scenario=polearm_archetype_lab`
- [ ] `make maintainability`
