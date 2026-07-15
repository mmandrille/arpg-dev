# v466 Spec: Eye-View Weapon Presentation

Status: Complete
Date: 2026-07-14
Codename: `eye-view-weapon-presentation`
Baseline: v465 `combat-impact-confirmation`

## Purpose

Improve the existing perspective camera mode so it reads as the hero's eye view: the camera sits
near the hero head, equipped hand items stay visible in front of the player, and basic attacks
produce readable first-person-style weapon motion. The slice is presentation-only and keeps the
server-authoritative combat, inventory, and movement contracts unchanged.

## Non-goals

- No server, protocol, combat math, cooldown, hit validation, damage, loot, or movement tuning.
- No full FPS controller rewrite, new collision model, or alternate local-only gameplay path.
- No production weapon art or external camera/controller plugin.
- No perfect per-class/per-weapon-family first-person animation pass; prove a small reusable
  framing path and defer bespoke polish.
- No multiplayer-specific camera synchronization.

## Acceptance criteria

- Perspective camera mode positions the camera at a believable eye/head height and no longer feels
  like a chest-mounted camera.
- In perspective camera mode, the local hero's equipped main-hand item is visible from the active
  camera when the item has an existing equipment visual.
- Off-hand/shield visuals remain compatible with existing two-handed/off-hand blocking rules and do
  not regress isometric equipment presentation.
- Basic attack input produces visible hand/weapon motion from the perspective camera using existing
  client-side attack presentation signals.
- Isometric camera mode, isometric equipment mounting, and existing attack clip selection continue
  to behave as before.
- Client debug/bot surfaces expose enough state for a headless Godot scenario to assert that eye
  view is active, a hand item is mounted, and attack motion was observed.

## Scope and likely files

- `shared/assets/camera_presentations.v0.json` and schema: data-owned eye-view offsets/framing.
- `client/scripts/player_camera_controller.gd`: refine perspective rig setup and expose debug state.
- `client/scripts/equipment_visuals.gd`: add a presentation-only eye-view hand mount or transform
  path if existing skeleton-mounted gear is clipped or hidden from camera.
- `client/scripts/combat_local_attack_presentation.gd` or `client/scripts/melee_lunge_presentation.gd`:
  reuse or lightly extend client-side attack feedback for first-person weapon motion.
- `client/tests/test_camera_mode_settings.gd` and `client/tests/test_item_visuals.gd`: focused
  coverage for camera position and mounted visible item state.
- `tools/bot/scenarios/client/105_eye_view_weapon_presentation.json`: Godot visual proof.
- `docs/as-built/v466_eye-view-weapon-presentation.md`, `PROGRESS.md`, and
  `docs/progress/slice-lifecycle.md`: lifecycle close-out.

## Test and bot proof

- `make validate-shared` if shared asset schema/data changes.
- `godot --headless --path client --script res://tests/test_camera_mode_settings.gd`
- `godot --headless --path client --script res://tests/test_item_visuals.gd`
- `make client-unit`
- `HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation`
- Manual visual check: `make bot-visual scenario=105_eye_view_weapon_presentation`

## Asset and plugin decision

- Adopt: existing `chest_view`/perspective camera mode, `PlayerCameraController`, equipment visual
  resolver, humanoid hand sockets, generated GLB weapon assets, and attack presentation clips.
- Borrow: recent camera smoothing, melee lunge, combat impact, and equipment visual test/bot
  patterns.
- Reject: external first-person camera plugins, new controller dependencies, and new asset
  pipelines for this slice.

## Open questions and risks

- The implementation should improve the existing perspective mode in place instead of adding a new
  persisted setting unless code inspection shows that separate `eye_view` and `chest_view` modes are
  materially simpler.
- Hero body clipping may require local-only visibility masking or a duplicated view-model mount.
  Default: keep the smallest presentation-only solution that makes equipped hand items visible
  without changing authoritative character state.
- If some weapon families do not frame cleanly, prove one-handed main-hand plus a non-crashing
  fallback for other families and defer bespoke offsets.
