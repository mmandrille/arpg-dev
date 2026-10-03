# v533 — Camera zoom-out

- **Status:** Complete (combined `make ci` passed 2026-10-03, 20m36s)
- **Date:** 2026-10-02
- **Codename:** `camera-zoom-out`
- **Baseline:** `dab60eb5` plus v532 in the working tree. No dependency on v532 beyond sharing the tree.
- **ADRs:** ADR-0001 D2 (display only), ADR-0018 (render baseline; screenshot gate).

## Purpose

Playtest feedback: the isometric camera is too close. Show more of the world around the hero by default, without taking away the ability to zoom in.

## Scope

- Raise `isometric.zoom_default` in `shared/assets/camera_presentations.v0.json` by 25% (12.0 → 15.0, orthographic size, larger = farther) and `zoom_max` proportionally (20.0 → 25.0) so the player can still zoom out beyond the default by the same ratio as before. `zoom_min` (8.0) is unchanged.
- No code constant changes. Tuning stays in the data file (Data-Driven Configuration Policy). The `12.0/8.0/20.0` literals inside `player_camera_controller.gd` are only missing-data fallbacks and are out of scope.

## Non-goals

- Chest-view/perspective mode, default scroll step (`CAMERA_ZOOM_STEP`), fog or light radius (v534), any protocol, server, rules or golden change.

## Acceptance criteria

1. The isometric camera starts at the new default and the whole player-visible world area at that default is at least 25% wider than before (orthographic size ratio ≥ 1.25).
2. The player can still zoom both in (below default) and out (above default) within the data bounds; a rules-derived test asserts `zoom_min < zoom_default < zoom_max` and default within bounds, without pinning the numbers.
3. Existing camera, visual-replay and fog-overlay tests still pass (replay zoom multipliers are relative to `zoom_default`, so captures scale consistently).
4. Real-renderer captures at the default zoom show a wider area with the hero, NPCs and props still legible.
5. Render cost at the new default is measured against the old default and reported, with limits.

## Surfaces and verification

`shared/assets/camera_presentations.v0.json`; `client/tests/test_camera_mode_settings.gd`. Focused: `test_camera_mode_settings`, `test_combat_vfx`, `test_fog_of_war_overlay`, `make validate-shared`, town-play before/after captures, live frame counters. No `make ci`.

## Risks

Larger ortho size draws more geometry (draw calls, fog work); weaker GPUs untested. The pinned `dungeon_frame_pacing_probe` fixture is stale (see as-built), so cost is reported from a single live snapshot, not a controlled A/B.
