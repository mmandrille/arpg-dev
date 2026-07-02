# v409 Plan — Town Training Doll

Status: Ready for implementation
Goal: Town training doll with mirrored defensive stats, death/revive loop, and combat damage log UI.
Architecture: Reuse monster combat pipeline with `training_target` def flag; server emits `damage_breakdown` on doll combat events; client-only silhouette + scrollable log panel.
Tech stack: Go sim, shared JSON rules, protocol v8 additive, Godot client, Python bot.

## Baseline and shortcut decision

- Reuse `training_dummy` combat targeting, `monster_damaged` events, and floating combat text (v31).
- **Adopt** procedural silhouette visual (`training_doll_visual.gd`); reject `monster_dummy` GLB.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.json` | `town_training_doll` def |
| Modify | `shared/rules/monsters.v0.schema.json` | `training_target`, `revive_delay_ticks` |
| Modify | `shared/rules/worlds.v0.json` | Doll in `dungeon_levels` + `town_training_doll_lab` |
| Modify | `shared/assets/monster_visuals.v0.json` | Silhouette visual key |
| Modify | `shared/protocol/state_delta.v8.schema.json` | `damage_breakdown` on events |
| Create | `server/internal/game/training_doll.go` | Mirror stats, revive, kill hook |
| Create | `server/internal/game/combat_breakdown.go` | Formula line builder |
| Create | `server/internal/game/training_doll_test.go` | Unit tests |
| Modify | `server/internal/game/sim.go` | Combat hooks, tick revive |
| Modify | `server/internal/game/types.go` | Breakdown types, entity fields |
| Modify | `tools/bot/run.py` | `min_damage_breakdown_lines` matcher |
| Create | `tools/bot/scenarios/118_town_training_doll.json` | Protocol proof |
| Create | `client/scripts/training_damage_log_panel.gd` | Scrollable log UI |
| Create | `client/scripts/training_doll_visual.gd` | Human silhouette mesh |
| Create | `client/scripts/combat_breakdown_format.gd` | Line formatting |
| Modify | `client/scripts/main.gd` | Panel + event routing |

## Maintenance ratchet

Hotspot: `main.gd` — panel hook only; new files stay under 600 lines.

## Task 1 — Shared contracts

- [x] Step 1.1: Monster def, schema, visuals, world placement, protocol `damage_breakdown`
```bash
make validate-shared
```

## Task 2 — Server

- [x] Step 2.1: Training doll mirror, combat breakdown, death/revive, tests
```bash
cd server && go test ./internal/game/... -run TrainingDoll -count=1
```

## Task 3 — Bot scenarios

- [x] Step 3.1: `118_town_training_doll.json` + matcher extension
```bash
make bot scenario=118_town_training_doll
```

## Task 4 — Client

- [x] Step 4.1: Silhouette visual, damage log panel, main wiring, client scenario
```bash
make client-unit
make bot-client SCENARIO=91_training_damage_log HEADLESS=1
```

## Task 5 — Lifecycle docs

- [x] Update `PROGRESS.md`, as-built, lifecycle row

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `go test ./internal/game/... -run TrainingDoll`
- [x] `make bot scenario=118_town_training_doll`
- [x] `make bot-client SCENARIO=91_training_damage_log HEADLESS=1`
