# v357 As-built — Town Night Perimeter

## What shipped

Level-0 town is enclosed in a **~15 m wooden palisade** with a **locked south gate**, **dungeon-mood fog lighting**, and a **respaced hub layout** around center `(11, 12)`.

### Shared

- `shared/assets/town_presentation.v0.json` — center, radius, segment count, gate gap tuning.
- `tools/assets/gen_town_perimeter.py` — deterministic wood wall segment generator.
- `shared/rules/worlds.v0.json` — 30 wood perimeter walls, respaced services, spawn `(11, 18)`, `town_exit_gate` at `(11, 27)`.
- `shared/rules/interactables.v0.json` — `town_exit_gate` with `locked_exit_reason: town_exit_locked` and wide closed barrier (`5.5 × 0.5`).
- `shared/golden/dungeon_stairs.json` — town stairs/teleporter pinned to new center layout.

### Server

- `wood` obstacle kind blocks movement, projectiles, and LOS like `wall`.
- Closed gate rejects `action_intent` with `town_exit_locked` before door-open toggle.
- `town_perimeter_test.go` — gate reject, perimeter block, service reachability.

### Client

- Town fog active applies **fog-suppressed dungeon lighting** at level 0 (`dungeon_depth_lighting.gd`, `town_presentation_loader.gd`).
- Wood palisade rendering (`wall_renderer.gd`, `ground_wall_factory.gd`).
- Gate feedback: floating text + `last_intent_reject_reason` for bot assertions.
- `hero_visibility_field.gd` — `wood` walls occlude fog like stone walls.

### Bot / tests

- Extended client `90_town_night_perimeter` — fog, locked gate, vendor reachability, outside click blocked.
- `bot_intent_reject_assertions.gd` — `wait_intent_rejected` / `assert_intent_rejected`.
- `28_reachable_dungeon_obstacles` — uses `generated_wall_lab` (no town cross-map walk under 30-step player budget).
- GDScript: `test_town_night_lighting.gd`, coop client reject reason, fog wood occluder test.

## Manual check

```bash
make play
# Town is dark with fog; wooden fence visible; south gate shows "You can't leave for now."

HEADLESS=1 make bot-client SCENARIO=90_town_night_perimeter
make bot-visual scenario=town_night_perimeter
```
