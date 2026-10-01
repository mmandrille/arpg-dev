# v507 As-Built — Class silhouette readability

- **Date:** 2026-10-01
- **Status:** Verified; the existing class assets meet the visual-readability acceptance criteria, so no presentation-data tuning was justified. Combined batch `make ci` passed (11m41s, 2026-10-01).
- **Spec:** [v507 spec](../specs/v507_spec-class-silhouettes.md) · **Plan:** [v507 plan](../plans/v507_2026-10-01-class-silhouettes.md)
- **Baseline:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Scope:** Actual-play-camera visual review and a focused capture helper. No class, armor, equipment, gameplay, or renderer data was changed.

## Baseline matrix and findings

Using the actual gameplay camera helper with Godot 4.7.2 / Metal Forward+ on Apple M4 Pro, the real-renderer matrix captured all five classes in color and grayscale:

- **40 idle frames:** five classes × empty/class-appropriate gear × normal/battle zoom × color/grayscale.
- **30 attack frames:** five classes × class-appropriate gear × normal zoom × color/grayscale × three attack samples (20%, 50%, 80%).

The captures showed distinct silhouettes at normal zoom and the tighter battle view. Rogue and Ranger were the closest unequipped pair, but remained distinguishable: Rogue's lower cape outline differs from Ranger's rounder head outline; Ranger's quiver and bow make the equipped profile clear. The other class cues also remain visible in grayscale:

| Class | Unequipped cue | Equipped cue | Review |
|---|---|---|---|
| Barbarian | Broad body and bear hat | Axe | Distinct at both zooms and in grayscale |
| Sorcerer | Tall hat and cape | Staff | Distinct at both zooms and in grayscale |
| Paladin | Helmet/visor and cape | Sword and shield | Distinct at both zooms and in grayscale |
| Rogue | Low cape; no native headgear | Paired daggers | Distinguishable; closest unequipped pair is Ranger |
| Ranger | Cape and quiver | Bow and quiver | Distinguishable from Rogue in grayscale and while equipped |

The attack samples retained readable weapon silhouettes for axe, staff, sword/shield, daggers, and bow, with no obvious clipping. No class pair failed the grayscale distinction check. Rogue and Ranger still have no native headgear; this existing asset limit remains unchanged.

The 10-image `skeleton gear` suite passed and was inspected for fit. The actual-camera and suite captures are local ignored evidence under `.artifacts/v507-class-silhouettes/` and `.artifacts/screenshots/20261001-110230/`; no PNGs were added to the repository.

## Decision

No class scale, native-part visibility, tint, armor-look, or equipment-selection changes were made. Existing assets already meet the readability criteria in the requested color/grayscale, empty/equipped, normal/battle-zoom matrix, and attack frames. Since no persistent geometry or accessory visibility changed, a renderer-cost comparison was not applicable. These screenshots establish visual distinction only, not a performance gain.

The capture helper at `client/scripts/showme/class_silhouette_play_camera_capture.gd` was corrected so its argument-match branch parses and it waits one process frame before camera setup; this fixed its blank initial capture. This is supporting capture tooling, not a gameplay presentation change.

## Focused verification

| Check | Result |
|---|---|
| Actual-camera idle matrix | PASS — 40 captures inspected |
| Actual-camera attack matrix | PASS — 30 captures inspected at 20%, 50%, and 80% |
| `make regen-screenshots SUITE="skeleton gear"` | PASS — 10/10 images |
| `make bot-client SCENARIO=97_bone_gear_sockets HEADLESS=1` | PASS |
| `make bot-visual scenario=97_bone_gear_sockets` | First run timed out at the scenario's internal 15-second wait on step 10 with the default delay |
| `BOT_STEP_DELAY=0.0 make bot-visual scenario=97_bone_gear_sockets` | PASS — visible renderer retry |
| `test_armor_look.gd` | PASS — 63 checks; one existing unknown-item warning |
| `test_animation.gd` | PASS — Godot emitted existing ObjectDB/resource allocator shutdown warnings |
| `test_item_visuals.gd` | PASS — Godot emitted existing ObjectDB/resource allocator shutdown warnings |
| `make maintainability` | PASS |
| `git diff --check` | PASS |

The default-delay visual-bot timeout is recorded as a timing caveat; the same scenario passed headless and in the visible retry with zero step delay. The transferred v503 dependency baseline had previously passed `make validate-shared` and `make validate-assets`; those validators were not rerun for v507 because no shared catalog or asset mapping changed. Full `make ci` remains the coordinator's integrated-batch gate.
