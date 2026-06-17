# v255 Plan - Fog LOS Shadow Mask

Status: Complete
Goal: Make the fog overlay visually obscure floor and objects behind rectangular walls in the
hero's line of sight.
Architecture: Keep fog LOS shadows client-presentational and driven from the existing
authoritative wall layout. `FogOfWarOverlay` owns screen-space shadow geometry and debug state;
`main.gd` only synchronizes the current normalized wall layout. No protocol, shared rules, server,
or replay contract changes are required.
Tech stack: Godot GDScript client, Godot client bot scenario, docs.

## Baseline and Shortcut Decision

Builds on v253 fog-of-war radius and v254 area-scaled dungeon density. The server already hides
living monsters behind rectangular walls; this slice fixes the remaining visual leak where the
client radial overlay makes behind-wall floor space look visible. Existing wall layouts already
arrive through snapshots/deltas and are normalized by `WallRenderer`.

Asset/plugin decision: reject external assets, imported fog art, shader plugins, and Godot addons.
Borrow the existing `FogOfWarOverlay` CanvasLayer, `WallRenderer` normalized wall views,
`input_shadow_overlay.gd`/code-native overlay patterns, and client bot debug/assertion conventions.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/fog_of_war_overlay.gd` | Store wall layout, project wall corners, render opaque LOS shadow polygons, expose debug state |
| Modify | `client/scripts/main.gd` | Synchronize `current_wall_layout` into the fog overlay with net-neutral line count |
| Modify | `client/tests/test_fog_of_war_overlay.gd` | Cover no-wall fallback, one-wall shadows, range filtering, and multiple shadows |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Let `assert_fog_of_war` check shadow/occluder debug counts |
| Add | `tools/bot/scenarios/client/68_fog_los_shadow_mask.json` | Client visual proof using an existing wall lab |
| Modify | `docs/specs/v255_spec-fog-los-shadow-mask.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v255 lifecycle row |
| Add | `docs/as-built/v255_fog-los-shadow-mask.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and deferred fog/dungeon presentation gaps |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`: not planned
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none planned
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Keep `client/scripts/main.gd` changes net-neutral or shrinking by folding existing wall
  layout assignment code while adding the fog sync call.
- [x] Keep LOS geometry inside `client/scripts/fog_of_war_overlay.gd`; it is well under the
  600-line target and is the owning presentation module.

Verification:
```bash
make maintainability
```

## Task 1 - Fog Overlay LOS Geometry

Files:
- Modify: `client/scripts/fog_of_war_overlay.gd`
- Modify: `client/tests/test_fog_of_war_overlay.gd`

- [x] Step 1.1: Add `set_wall_layout(walls)` to store normalized rectangular walls without
  depending on wall renderer nodes.
- [x] Step 1.2: Add screen-space shadow polygon generation from hero position, camera projection,
  `light_radius`, `gloom_radius`, and wall rectangle corners.
- [x] Step 1.3: Render opaque black `Polygon2D` shadows above the radial ColorRect while keeping
  the blocking wall readable by starting the mask just beyond the wall silhouette/far side.
- [x] Step 1.4: Expose debug state for `occluder_count`, `shadow_count`, and representative
  polygon points/bounds.
- [x] Step 1.5: Add unit coverage for no-wall fallback, one in-range wall shadow, out-of-range
  walls, and multiple wall shadows.

```bash
make client-unit
```

## Task 2 - Client Wall Layout Sync

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Synchronize current wall layout into the fog overlay after preset fallback wall
  rendering, snapshot wall layout rendering, and wall layout deltas.
- [x] Step 2.2: Keep `main.gd` at or below its maintainability allowance by making the touched wall
  render helpers net-neutral or smaller.
- [x] Step 2.3: Preserve existing wall rendering and `current_wall_layout` debug state.

```bash
make client-unit
```

## Task 3 - Client Bot Proof

Files:
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Add: `tools/bot/scenarios/client/68_fog_los_shadow_mask.json`

- [x] Step 3.1: Extend `assert_fog_of_war` with optional shadow/occluder count expectations.
- [x] Step 3.2: Add a focused client visual scenario using `collision_lab`, which already has an
  interior rectangular wall near the player.
- [x] Step 3.3: Prove both the new LOS shadow scenario and the existing radial fog scenario.

```bash
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
```

## Task 4 - Lifecycle Docs

Files:
- Modify: `docs/specs/v255_spec-fog-los-shadow-mask.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v255_fog-los-shadow-mask.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the v255 spec complete.
- [x] Step 4.2: Add v255 lifecycle and as-built notes.
- [x] Step 4.3: Update `PROGRESS.md` current status and defer remaining non-rectangular,
  doorway/high-obstacle, durable map-memory, and production lighting work.

```bash
make maintainability
```

## Final Verification

- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask`
- [x] `HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay`
- [x] `make maintainability`
- [x] `make ci`

Manual visual proof, if desired:

```bash
make bot-visual scenario=68_fog_los_shadow_mask
```
