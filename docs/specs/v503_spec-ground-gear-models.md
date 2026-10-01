# v503 — Ground gear models at game scale

- **Status:** Complete; inspected captures, focused checks, and combined batch `make ci` passed.
- **Date:** 2026-10-01
- **Codename:** `ground-gear-models`
- **Accepted batch assignment:** v503, Ground Gear Models
- **Baseline:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **ADRs:** [ADR-0006](../adr/0006-asset-pipeline.md), [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md)
- **Dependencies:** Ground model resolution from v487 is integrated at the assigned base. v504 quest/key/badge loot and v508 rarity cues may spec and plan in parallel; coordinate their implementation after this slice so the shared ground-loot surface stays reviewable.

## Purpose

Dropped weapons and shields use their KayKit models from v487. Dropped armor and jewelry still use
the existing manifest-backed fallback models, and their ground pose inherits the generic pose and
the `ground` scale used by the loot family. In particular, the ring and amulet family scales are
`0.7`; together with the shared ground multiplier of `0.7`, those models render at `0.49` of their
unscaled mesh size. Their silhouettes are consequently hard to recognize beside the existing
weapon drops at the game camera scale.

Tune the ground presentation of the existing head, chest, glove, belt, boot, ring, and amulet
models so each reads as its own gear family in the world. Keep the established KayKit-led look and
the manifest → model resolution path. This is a client presentation change only.

## Asset decision

- **Adopt:** keep the existing KayKit hand-item models and their `item_visuals.v0.json` resolution.
- **Borrow:** keep using the existing manifest-registered armor/jewelry fallback models for dropped
  gear; they already have provenance and are the explicit v487 fallback because the KayKit kits do
  not provide detachable armor meshes.
- **Reject:** no new downloads, generated meshes, plugins, or asset pipeline. Existing
  `equipment_display.v0.json` per-asset ground-pose settings are sufficient for this tuning.

## Scope

- Tune schema-backed `ground_pose.assets` entries in `shared/assets/equipment_display.v0.json` for
  the existing fallback head, chest, gloves, belt, boots, ring, and amulet model IDs.
- Choose each model's ground scale and pose from real-renderer captures. Center the model over its
  loot root and rest it at floor height where its bounds permit; avoid clipping and avoid one shared
  size cap that erases the distinction between small jewelry and larger armor.
- Extend focused client assertions so every gear family resolves its expected model through
  `assets.v0.json` and its effective pose comes from the shared catalog.
- Add a windowed `capture_frame` to the existing `client_full_equipment` scenario before it picks
  up the ground items. Use that scenario for the player-camera before/after proof and pickup
  regression.
- Capture and inspect the focused `floor-item` suite for armor, plus dedicated `ring` and `amulet`
  previews. (The screenshot suite discovers gear through equipped visuals, so it does not include
  those jewelry slots.)

## Non-goals

- Quest items, keys, or badges; v504 owns those ground models and item catalog entries.
- Rarity color, shape, label, outline, beam, glow, or color-accessibility changes; v508 owns rarity
  cues. Existing rarity presentation stays unchanged.
- Class silhouettes, equipped-gear mounting, hero armor tint, item rules, item stats, loot tables,
  drop/pickup behavior, collider dimensions, server state, protocol, or persistence.
- Importing or changing model geometry, textures, licenses, or the asset manifest.

## Acceptance criteria

1. Each of the seven ground gear families (head, chest, gloves, belt, boots, ring, and amulet)
   continues to instantiate its current manifest-backed model; tests derive the item and asset
   coverage from the shared presentation catalog and manifest rather than duplicating asset IDs.
2. Per-model scale, rotation, floor rest, and centering use the existing schema-backed
   `equipment_display.v0.json` pose path. Focused tests verify the configured transform is applied
   and that the model bounds remain above the floor within the selected visual tolerance.
3. A reviewer can distinguish the gear-family silhouettes in both the focused ground-item captures
   and the player-camera `equipment_lab` capture. Jewelry is visibly legible beside armor without
   turning its model into an oversized prop.
4. The bot scenario still proves that ground loot can be picked up and equipped; model pose or scale
   does not change the server outcome or loot pick collider.
5. The displayed rarity color and existing ground glow/beam behavior are unchanged.
6. Shared catalog validation and asset-manifest validation pass. No server, protocol, gameplay-rule,
   or item-golden changes are introduced.

## Security assessment

The security skill classifies pure presentation and animation as out of scope. The planned change
uses local schema-validated catalogs and manifest paths, with no new input, API, authorization,
database, or user-controlled HTML rendering surface; no security-specific change is indicated.

## Likely surfaces

| Area | Files |
|---|---|
| Shared presentation | `shared/assets/equipment_display.v0.json` (schema already supports per-asset `ground_pose`) |
| Client proof | `client/tests/test_loot_node_factory.gd`, `client/tests/test_item_visuals.gd` |
| Player-camera proof | `tools/bot/scenarios/client/10_full_equipment.json` |
| Evidence | `docs/as-built/v503_ground-gear-models.md`, `docs/as-built/assets/v503/` |

## Focused verification

- Godot: `godot --headless --path client --script res://tests/test_loot_node_factory.gd`
- Godot integration: `godot --headless --path client --script res://tests/test_item_visuals.gd`
- Data: `make validate-shared` and `make validate-assets`
- Pickup regression: `make bot-client SCENARIO=client_full_equipment HEADLESS=1`
- Visual model sweep: `make regen-screenshots SUITE="floor-item"`
- Real player camera: `make bot-visual scenario=client_full_equipment`
- Capture jewelry directly with `python3 skills/showme/scripts/render_focus.py --focus floor-item
  --items ring` and the corresponding `amulet` command. Inspect all generated images and preserve
  representative before/after gear and player-camera captures under `docs/as-built/assets/v503/`.
- Run `make maintainability` and `git diff --check` for the changed files.

The player-camera scenario is intentionally windowed for image proof. Its capture step must use
`skip_if_headless: true` so the existing headless bot path remains usable.

## Integration risks and open questions

- **No material product question remains.** Use the existing KayKit-led direction and current
  manifest assets, as the accepted batch brief specifies.
- v504 and v508 may change `item_presentations.v0.json` or loot presentation code after this slice
  integrates. This slice avoids those shared surfaces by owning only `equipment_display.v0.json`,
  gear assertions, and the existing full-equipment camera scenario.
- The current local Godot executable is newer than the repo's pinned version and rewrites import
  metadata when it imports. Restore incidental `.glb.import` defaults only after verifying they are
  generated sidecar changes; do not include them as slice work.
