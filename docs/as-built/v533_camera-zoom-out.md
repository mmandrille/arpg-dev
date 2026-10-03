# v533 As-built — Camera zoom-out

- **Status:** Complete — focused slice verification plus the combined `make ci` on the integrated v532–v537 state (CI OK, 11 stages, 20m36s, 2026-10-03). `make ci-full` was not run.
- **Date:** 2026-10-02
- **Spec:** [`v533_spec-camera-zoom-out.md`](../specs/v533_spec-camera-zoom-out.md) · **Plan:** [`v533_2026-10-02-camera-zoom-out.md`](../plans/v533_2026-10-02-camera-zoom-out.md)
- **Scope:** shared camera data and one test. No code, protocol, server, golden or rules change.

## What changed

`shared/assets/camera_presentations.v0.json`: isometric `zoom_default` 12.0 → 15.0 (+25%), `zoom_max` 20.0 → 25.0 (same zoom-out headroom ratio as before); `zoom_min` stays 8.0. `test_camera_mode_settings.gd` gained three rules-derived assertions (default inside bounds, room to zoom out, room to zoom in) that do not pin the new numbers.

## Verification

| Command | Result |
|---|---|
| `godot --headless … test_camera_mode_settings.gd` | PASS, 49 passed |
| `… test_combat_vfx.gd`, `… test_fog_of_war_overlay.gd` | PASS, 98 and 147 passed |
| `make validate-shared` | PASS, 2,271 checks plus CODEMAP |
| `render_focus.py --focus town-play` plaza and vendor, before and after | captured |

`make ci` / `make ci-full` not run (data-only, no contract change).

## Visual and cost evidence

Real-renderer captures, 1120×720, balanced, [`assets/v533/`](assets/v533/): `before-plaza.png`, `after-plaza.png`, `before-vendor.png`, `after-vendor.png`. The after frames show visibly more terrain around the plaza; NPCs and props stay legible but smaller.

| Live snapshot, generated dungeon fixture (25 monsters, 1920×1080, vsync) | Before (12.0) | After (15.0) |
|---|---:|---:|
| Draw calls | 125 | 135 (+8%) |
| Primitives | ~132.1k | ~135.3k (+2.4%) |
| avg frame ms (last 8 samples) | 17.4–17.9 | 16.4–20.2 |

## Limits

- **Not a controlled A/B.** `dungeon_frame_pacing_probe` times out at step 2 (`wait_wall_layout`) on unmodified `dab60eb5` as well, so its pinned seed/level no longer produces its fixture after the v522–v531 generation changes (pre-existing finding, not fixed here). The numbers above are one live snapshot per side from the same fixture before the timeout; the after run showed one 118 ms process spike. Frame times are close to the 60 FPS vsync budget on this host either way. A weaker GPU is untested.
- Capture window is smaller than a real play window; compare proportions, not absolute size.
- Hover/click targeting and minimap are unchanged but were not re-exercised beyond the camera/fog tests.
