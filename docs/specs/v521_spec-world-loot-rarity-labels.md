# v521 Spec - World Loot Rarity Labels

- **Status:** Complete
- **Date:** 2026-10-01
- **Codename:** world-loot-rarity-labels
- **Dependencies:** v518 ground item tint and rarity-colored labels; current integrated baseline `f80cc5b5`.

## Purpose

Remove the floating geometric rarity markers from ground loot. Keep loot readability on the actual
item model's rarity tint and the full rarity name in its revealed label, with the existing world
lighting. The user's explicit request authorizes this single-slice change.

## Non-goals

- Do not remove the small geometric rarity cues inside inventory/equipment slots.
- Do not remove or recolor the full rarity name shown with revealed ground-loot labels.
- Do not change the actual ground-model rarity tint, item lighting, loot selection, pickup, or
  authoritative item data.
- No protocol, server, rules, gameplay, camera, or fog changes.
- No new assets, packages, plugins, or asset pipeline.

## Acceptance Criteria

1. Eligible ground equipment in every supported rarity creates no floating geometric `RarityCue`
   node.
2. Revealed ground loot labels continue to show the full rarity name and item name, colored by the
   existing rarity text-color mapping.
3. Rarity tint on ground/equipped models and existing scene lighting remain unchanged.
4. Inventory/equipment slot rarity shapes remain distinct and present.
5. The rarity catalog/schema retain only world-label settings still consumed after marker removal;
   unused world-marker size/position/outline fields are removed.
6. Unit tests cover marker absence for eligible rarities, label text/color preservation, and the
   unchanged slot cues. Existing model tint tests remain green.
7. A real-renderer rarity-cue capture with labels revealed shows tinted drops and colored text with
   no world shapes; the `inventory_lab_drop_item` bot scenario still passes.

## Scope and Likely Files

- `client/scripts/loot_node_factory.gd` — remove the world marker attachment.
- `client/scripts/rarity_cue_presenter.gd` — remove world-mesh creation/material code; retain slot
  shape drawing.
- `client/tests/test_rarity_cues.gd` — assert world-marker absence and retain slot shape checks.
- `client/scripts/showme/showme_rarity_cues_capture.gd` — allow the focused capture to reveal all
  rarity labels for visual inspection.
- `shared/assets/rarity_cues.v0.json` and `shared/assets/rarity_cues.v0.schema.json` — drop unused
  world-marker settings, retain revealed-label height.
- `docs/as-built/v521_world-loot-rarity-labels.md`, `docs/progress/slice-lifecycle.md`,
  `docs/progress/slice-codename-index.md`, and `PROGRESS.md` — record implementation and proof.

## Completion

Implemented and verified; see [plan](../plans/v521_2026-10-01-world-loot-rarity-labels.md) and
[as-built](../as-built/v521_world-loot-rarity-labels.md).

No new bot scenario is required; reuse `inventory_lab_drop_item`.

## Existing Code and Asset Decision

The live path is `LootNodeFactory.make_loot_node` → rarity-colored `Label3D` and model tint →
`RarityCuePresenter.add_world_marker`. The same presenter also draws separate inventory-slot
badges. `showme_rarity_cues_capture.gd` already renders five representative rarities using the
in-repo assets. The test gate explicitly checks the world marker today.

- **Adopt:** existing rarity label/color and model-tint code.
- **Borrow:** in-repo rarity capture fixture and loot bot scenario.
- **Reject:** new icons, meshes, add-ons, and external asset pipelines.

## Focused Verification

- `godot --headless --path client --script res://tests/test_rarity_cues.gd`
- `godot --headless --path client --script res://tests/test_loot_node_factory.gd`
- `make client-unit`
- `make validate-shared`
- `make validate-assets`
- `make maintainability`
- `make bot-visual scenario=inventory_lab_drop_item BOT_STEP_DELAY=0.05`
- `python3 skills/showme/scripts/render_focus.py --focus rarity-cues --ground-tone dark --reveal true`
- Inspect the generated image for tinted drops, readable rarity-colored labels, and no world shapes.
- Run the standalone `make ci` final gate.

## Risks and Open Questions

- The floor label remains hidden until loot reveal/hover by existing design. The focused capture
  reveals the sample labels to verify their colors without changing runtime reveal behavior.
- There are no open product decisions: the user's request explicitly removes the world markers and
  retains text color and lighting.
