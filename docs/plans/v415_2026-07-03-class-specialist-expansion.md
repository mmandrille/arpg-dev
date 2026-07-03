# v415 Plan — Class Specialist Expansion

Status: Ready for implementation
Goal: Add ten class-specialist templates, wire drops, and expose class lock in requirement_status UI.
Architecture: Data-only template expansion on v405 pipeline; server appends `stat: "class"` rows;
client formats via shared ItemRequirementViews helper. No new intents or combat paths.
Tech stack: shared JSON, Go sim, Godot client, Python bot.

## Baseline and shortcut decision

Reuses v405 `class_specialist` validation, equip gate, and lab world. Borrow presentation families
from generic slot gear; reject new 3D assets.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.json` | 10 new specialist templates |
| Modify | `shared/rules/treasure_classes.v0.json` | depth-1 drop entries |
| Modify | `shared/rules/worlds.v0.json` | lab loot spawns |
| Modify | `shared/assets/item_presentations.v0.json` | icon families |
| Modify | `shared/assets/item_visuals.v0.json` | mount mappings |
| Modify | `shared/protocol/state_delta.v8.schema.json` | class requirement row |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | class requirement row |
| Modify | `shared/rules/items.v0.schema.json` | add rogue to class_required enum |
| Create | `server/internal/game/item_class_requirements.go` | class status helpers |
| Modify | `server/internal/game/sim.go` | annotate views + equip preview |
| Modify | `server/internal/game/types.go` | ClassID on RequirementStatusView |
| Modify | `server/internal/game/class_specialist_gear_test.go` | expansion tests |
| Modify | `client/scripts/item_requirement_views.gd` | class line formatting |
| Modify | `client/scripts/inventory_panel.gd` | use shared formatter |
| Create | `client/tests/test_item_requirement_views.gd` | headless unit test |
| Modify | `tools/bot/scenarios/114_class_specialist_gear_lab.json` | class status proof |
| Modify | `scripts/client_smoke.sh` | register new test |
| Docs | `docs/as-built/v415_class-specialist-expansion.md`, lifecycle, PROGRESS |

## Maintenance ratchet

Hotspot files touched: `sim.go` (grandfathered — touch-to-shrink only), new `item_class_requirements.go` ≤600 lines.

Decision: extract class requirement logic to new file; no sim.go growth beyond annotate calls.

## Task 1 — Shared templates and drops

- [ ] Add 10 class_specialist templates with roll pools
- [ ] Treasure class + lab world entries
- [ ] Presentations and visuals

```bash
make validate-shared
make validate-assets
```

## Task 2 — Protocol + server class requirement status

- [ ] Extend v8 requirement_status schema (`class` stat + `class_id`)
- [ ] `item_class_requirements.go` + sim annotate + equip preview
- [ ] Go tests

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ClassSpecialist|ClassRequirement' -count=1
```

## Task 3 — Client tooltip class row

- [ ] ItemRequirementViews + inventory_panel integration
- [ ] Headless test + client_smoke registration

```bash
godot --headless --path client --script res://tests/test_item_requirement_views.gd
```

## Task 4 — Bot proof

- [ ] Extend `114_class_specialist_gear_lab.json`

```bash
make bot scenario=class_specialist_gear_lab
```

## Task 5 — Lifecycle docs

- [ ] as-built, lifecycle row, PROGRESS current status

## Final verification

- [ ] `make validate-shared`
- [ ] `cd server && go test ./internal/game/... -run 'ClassSpecialist|ClassRequirement' -count=1`
- [ ] `make bot scenario=class_specialist_gear_lab`
- [ ] `make maintainability`
