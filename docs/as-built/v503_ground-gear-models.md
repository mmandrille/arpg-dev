# v503 As-Built — Ground gear models at game scale

- **Date:** 2026-10-01
- **Spec:** [`v503_spec-ground-gear-models.md`](../specs/v503_spec-ground-gear-models.md) ·
  **Plan:** [`v503_2026-10-01-ground-gear-models.md`](../plans/v503_2026-10-01-ground-gear-models.md)
- **Baseline:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Status:** Integrated; focused checks and combined batch `make ci` passed (11m41s, 2026-10-01).
- **Scope:** client presentation data and visual coverage only. No item, drop, collider, server,
  protocol, model, manifest, rarity, or lighting changes.

## What changed

The seven existing manifest-backed armor and jewelry fallback models now have per-asset ground
poses in `shared/assets/equipment_display.v0.json`. Each is centered over its loot root and rests
at 0.05 m. Sizes differ by family: helm 1.25×, chest 1×, gloves 1.5×, belt 0.5×, boots 1.2×,
ring 3×, and amulet 2.5× the existing presentation scale. The ring is laid flat and recentered;
its source GLB carries a large baked offset. The amulet's asset rotation compensates for its
existing child rotation so the pendant rests flat. Jewelry caps are 0.14 m (ring) and 0.20 m
(amulet).

`test_loot_node_factory.gd` now derives the seven gear slots from the item templates, checks each
manifest-backed fallback model, and verifies the configured pose, floor-rested bounds, scale, and
absence of the primitive rarity tile. `test_item_visuals.gd` checks family model resolution for
each ground gear slot. The `client_full_equipment` scenario records one player-camera frame after
initial loot is present and before pickup; the capture step skips headless runs.

## Visual review

The 60-item `floor-item` suite passed before and after tuning. The isolated captures show larger,
more legible helm and glove silhouettes, boots remain compact, the belt no longer dominates the
preview, and the centered ring and amulet are visible on the floor. The player-camera scenario
passed before and after tuning. Its captured scene is very dark and some ground models remain small
at that camera distance, so this is a single-frame placement check rather than conclusive proof of
family recognition under normal gameplay lighting. No lighting, glow, or rarity adjustments were
made.

| Family | Before | After |
|---|---|---|
| Player camera | [PNG](assets/v503/ground-gear-camera-before.png) | [PNG](assets/v503/ground-gear-camera-after.png) |
| Head | [PNG](assets/v503/helm-before.png) | [PNG](assets/v503/helm-after.png) |
| Chest | [PNG](assets/v503/chest-before.png) | [PNG](assets/v503/chest-after.png) |
| Gloves | [PNG](assets/v503/gloves-before.png) | [PNG](assets/v503/gloves-after.png) |
| Belt | [PNG](assets/v503/belt-before.png) | [PNG](assets/v503/belt-after.png) |
| Boots | [PNG](assets/v503/boots-before.png) | [PNG](assets/v503/boots-after.png) |
| Ring | [PNG](assets/v503/ring-before.png) | [PNG](assets/v503/ring-after.png) |
| Amulet | [PNG](assets/v503/amulet-before.png) | [PNG](assets/v503/amulet-after.png) |

## Verification

| Check | Result |
|---|---|
| `godot --headless --path client --script res://tests/test_loot_node_factory.gd` | PASS — 536 checks |
| `godot --headless --path client --script res://tests/test_item_visuals.gd` | PASS — Godot printed ObjectDB/resource-leak shutdown warnings after the passing result |
| `make bot-client SCENARIO=client_full_equipment HEADLESS=1` | PASS — pickup/equip scenario |
| `make bot-visual scenario=client_full_equipment` | PASS before and after tuning; see visual limit above |
| `make regen-screenshots SUITE="floor-item"` | PASS before and after — 60/60 captures each run |
| `python3 skills/showme/scripts/render_focus.py --focus floor-item --items ring` | PASS before and after tuning |
| `python3 skills/showme/scripts/render_focus.py --focus floor-item --items amulet` | PASS before and after tuning |
| `make validate-shared` | PASS — 2,208 checks; CODEMAP valid |
| `make validate-assets` | PASS — 451 checks |
| `make maintainability` | PASS — file-size, extraction-coupling, dashboard ratchets |
| `git diff --check` | PASS |

The post-tuning headless integration scenario confirms pickup and equip behavior; the fixed pickup
collider and all rarity visuals remain unchanged. Full `make ci` is reserved for the coordinator's
integrated batch gate.
