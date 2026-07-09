# v462 Plan — Rounded Dungeon Corners

Status: Complete
Goal: Make generated dungeon interiors feel less blocky by adding client-only rounded corner presentation to `room_wall` joins, with an explicit mid-slice option review so the owner can choose the final style before rollout.
Architecture: Keep authoritative wall layout unchanged. The server continues to emit rectangular room walls; the client derives eligible `room_wall` L-turns from the normalized wall layout and renders a softer corner treatment on top of the existing geometry. The first pass implements exactly two style variants for review on the same join-detection path, then the chosen variant is kept for rollout. Town, boss floors, collision, and non-`room_wall` obstacle shapes remain unchanged.
Tech stack: Godot client presentation, headless/windowed visual capture, client unit tests, client bot scenario, docs.

## Baseline and shortcut decision

Baseline: v461 `entity-locomotion-polish`

Reuse:
- Adopt: existing procedural dungeon presentation pipeline in `GroundWallFactory` and `WallRenderer`, plus the generated wall layout from the current client state.
- Borrow: existing wall/floor factory tests, `generated_wall_lab`, and the extended client scenario pattern from `wall_floor_dungeon_rollout`.
- Reject: external dungeon kits, imported curved wall assets, Godot addons, and any server-authored geometry format change for this slice.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/dungeon_wall_corner_presentation.gd` | Detect eligible `room_wall` joins and build bevel/rounded review variants. |
| Modify | `client/scripts/wall_renderer.gd` | Invoke the helper after wall layout render and attach presentation nodes without changing authoritative wall state. |
| Modify | `client/tests/test_factories.gd` | Prove eligible joins produce corner nodes, excluded surfaces do not, and both styles render stable meshes. |
| Add | `client/scripts/rounded_dungeon_corner_capture.gd` | Save focused PNG previews for both review variants without touching the main client coordinator. |
| Add | `tools/bot/scenarios/client/100_rounded_dungeon_corners.json` | Client scenario for rounded-corner layout readiness and visual replay. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [ ] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.
- [ ] Defer extraction with rationale: not needed if the helper owns the join/style logic.

Verification:
```bash
make maintainability
```

## Task 1 — Corner helper and eligibility model

Files:
- Add: `client/scripts/dungeon_wall_corner_presentation.gd`

- [x] Step 1.1: Detect deterministic `room_wall` L-turns only and exclude town, perimeter, obstacle kinds, and non-corner junctions.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

- [x] Step 1.2: Build exactly two review variants: `bevel_cap` and `rounded_cap`.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

## Task 2 — Renderer integration

Files:
- Modify: `client/scripts/wall_renderer.gd`

- [x] Step 2.1: Attach rounded-corner nodes after normalized wall layout render while keeping wall-count/layout output unchanged.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

- [x] Step 2.2: Keep the renderer below the ratchet ceiling by delegating all join/style logic to the new helper.
```bash
make client-unit
```

## Task 3 — Variant review capture checkpoint

Files:
- Add: `client/scripts/rounded_dungeon_corner_capture.gd`

- [x] Step 3.1: Save a focused `bevel_cap` PNG for owner review.
```bash
godot --windowed --single-window --resolution 960x640 --path client --script res://scripts/rounded_dungeon_corner_capture.gd -- --style bevel_cap --output .artifacts/showme/rounded-dungeon-corner-bevel.png
```

- [x] Step 3.2: Save a focused `rounded_cap` PNG for owner review.
```bash
godot --windowed --single-window --resolution 960x640 --path client --script res://scripts/rounded_dungeon_corner_capture.gd -- --style rounded_cap --output .artifacts/showme/rounded-dungeon-corner-rounded.png
```

- [x] Step 3.3: Stop for user feedback before final rollout lock.

## Task 4 — Client scenario coverage

Files:
- Add: `tools/bot/scenarios/client/100_rounded_dungeon_corners.json`

- [x] Step 4.1: Add a focused client scenario on `generated_wall_lab` that proves layout readiness on floor `-1`.
```bash
HEADLESS=1 make bot-client scenario=100_rounded_dungeon_corners
```

- [x] Step 4.2: Keep `wall_floor_dungeon_rollout` green as the existing regression proof.
```bash
HEADLESS=1 make bot-client scenario=wall_floor_dungeon_rollout
```

## Final verification

- [x] `godot --headless --path client --script res://tests/test_factories.gd`
- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-client scenario=100_rounded_dungeon_corners`
- [x] `HEADLESS=1 make bot-client scenario=wall_floor_dungeon_rollout`
- [x] `make maintainability`

## Deferred scope

- Final style lock landed as the `rounded_cap` corner treatment plus the approved softened top edge on eligible `room_wall` runs.
- Perimeter/generated wall joins, non-rectangular walkable space, and obstacle silhouette redesign.
