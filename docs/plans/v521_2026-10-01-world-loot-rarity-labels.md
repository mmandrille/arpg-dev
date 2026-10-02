# v521 Plan - World Loot Rarity Labels

- **Status:** Complete; standalone `make ci` passed
- **Date:** 2026-10-01
- **Base commit:** `f80cc5b5`
- **Goal:** Remove 3D geometric rarity markers from floor loot while retaining rarity-colored names, model tint, lighting, and inventory-slot shapes.
- **Prerequisites:** v518 item models/rarity cues and the current shared catalog are integrated. No sibling slice dependency.

## Spec Review Gate - PASS

- Every acceptance criterion maps to a Godot test, catalog/schema validation, bot interaction, or actual renderer capture.
- Scope is the visual client and its presentation catalog; no server, protocol, gameplay, or authoritative item changes.
- Slot rarity shapes are explicitly out of scope; world-label text, color, model tint, and lighting remain.
- The catalog schema will drop only marker fields that become unused; the existing label-height value remains.
- Existing assets and capture infrastructure are adopted/borrowed; new dependencies and assets are rejected.
- The direct user request authorizes implementation; no unresolved product decision remains.

## File Map and Ownership

| Action | Path | Responsibility |
|---|---|---|
| Modify | `client/scripts/loot_node_factory.gd` | Stop attaching 3D markers to real loot nodes. Preserve labels, rarity text color, model creation/tint. |
| Modify | `client/scripts/rarity_cue_presenter.gd` | Remove world-only mesh/material methods and caches. Keep slot badge rendering and shared shape mapping. |
| Modify | `client/tests/test_rarity_cues.gd` | Cover marker absence, full revealed label text/color, exclusions, and distinct slot rarity shapes. |
| Modify | `client/scripts/showme/showme_rarity_cues_capture.gd` | Make `--reveal true` reveal all five sample labels for visual evidence. |
| Modify | `shared/assets/rarity_cues.v0.json` | Remove unused world marker offset, height, size, colors; retain label height. |
| Modify | `shared/assets/rarity_cues.v0.schema.json` | Match the smaller catalog contract. |
| Add | `docs/as-built/v521_world-loot-rarity-labels.md` | Record implementation, exact validation, visual evidence, and limits. |
| Modify | `docs/progress/slice-lifecycle.md`, `docs/progress/slice-codename-index.md`, `PROGRESS.md` | Register completed slice and current state. |

## Maintenance and Data Ownership

- Remove now-unused code and data instead of retaining a compatibility path.
- Keep rarity colors sourced from the existing `UiTheme` catalog and item display tint from the existing equipment presentation data.
- Do not modify the authoritative game or protocol.
- Check touched files against `.maintainability/file-size-baseline.tsv`; preserve or lower the existing baseline.

## Ordered Tasks

### 1. Remove floor markers, preserve text/tint

- [x] Remove `RarityCuePresenter.add_world_marker` from loot-node creation.
- [x] Delete the unused world marker mesh/material caches and methods; preserve `_shape_points` and slot drawing.
- [x] Remove unused marker fields from the rarity catalog/schema while retaining `revealed_label_height`.
- [x] Update the capture fixture so reveal mode displays every sample rarity label.

### 2. Protect the revised contract

- [x] Update rarity tests: no marker node for supported equipment rarities; labels retain full rarity/item names and expected rarity color; slot shapes remain distinct.
- [x] Preserve existing non-equipment/concealed-item exclusions and model tint coverage.
- [x] Keep CODEMAP accurate; no new source file or scenario is expected.

### 3. Verify and close out

- [x] Run the named Godot tests, `make client-unit`, shared/assets validators, and `make maintainability`.
- [x] Run `make bot-visual scenario=inventory_lab_drop_item BOT_STEP_DELAY=0.05`.
- [x] Capture five drops with `--reveal true`; inspect for no 3D rarity shapes and readable rarity-colored labels.
- [x] Run `make ci`, update the as-built, PROGRESS, lifecycle, and codename index; check `git diff --check` and commit on the current branch.

## Verification Notes

This is standalone work, so the plan retains `make ci` as the final gate. The bot and real-renderer capture are needed because unit tests cannot establish floor-visual appearance. No performance comparison is claimed or required for removal of the marker meshes.
