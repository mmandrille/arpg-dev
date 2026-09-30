# v491 Plan — Town look pass

Status: Implemented
Spec: [`docs/specs/v491_spec-town-look-pass.md`](../specs/v491_spec-town-look-pass.md)

## Baseline and shortcut decision
Reuses `DungeonKitFloor` tile variants/pick, `KitPieceLibrary`, v473/v489/v490 vendoring, and the
existing `town_presentation.v0.json` (centre, radius, gate). Adopt ten staged Remastered props.

## Maintenance ratchet
No grandfathered file touched; `town_node_factory.gd` shrinks (preview cabins/campfire removed).

## Task 1 — Assets + data + validation
- [x] Vendor 10 props; manifest entries. `dressing` block + schema; loader accessor.
- [x] Validator step [9]: dressing asset ids resolve. `tools/test_town_dressing.py` clearance/cross-check.
## Task 2 — Client
- [x] `TownDressing` (plaza MultiMesh, road, props, ambient silhouettes) + `sync`.
- [x] `GroundWallFactory.update_ground_material` → `TownDressing.sync`; preview uses it; drop cabins/campfire.
- [x] `test_town_dressing.gd` registered; `test_look_and_feel_polish.gd` still green.
## Task 3 — Visual gate
- [x] `scenes/town` before/after; tune props; copy to `docs/as-built/assets/v491/`.
## Task 4 — Regression + docs + CI
- [x] Town client scenarios; docs; `make ci`.

## Execution notes

- Tuning from captures dropped the wall-mount props (sword_shield, shield banner, quest-giver banner)
  and gave the plaza its own intact tile variant list.
- The plaza "holes" led to the v471 dungeon floor seating bug (decorated/weeds tiles buried); fixed in
  `DungeonKitFloor.tile_transform`. MultiMesh buffers are not readable headless, hence the pure function.
- The ambient silhouettes, once visible, were placeholder capsules: retired instead of shown.
