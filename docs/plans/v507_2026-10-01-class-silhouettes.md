# v507 Plan — Class Silhouettes

- **Status:** Complete; evidence review found no presentation-data tuning justified. Combined batch `make ci` passed; see the [as-built](../as-built/v507_class-silhouettes.md).
- **Date:** 2026-10-01
- **Spec:** [`v507_spec-class-silhouettes.md`](../specs/v507_spec-class-silhouettes.md)
- **Assigned base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Worktree:** `/Users/mmandrille/git/arpg-dev-batch/v507-class-silhouettes` (detached)
- **Prerequisite:** coordinator integration/transfer and diff review of v503 Ground Gear Models before any v507 runtime edit if paths overlap.

## Spec review gate

Scope and test mapping are sound: five class outlines, equipped refresh, animation/camera fit, catalog/manifest integrity, live equip path, paired play-camera screenshots, and conditional renderer cost. There is no new protocol, schema of network payloads, gameplay rule, server authority, world preset, replay, or Go determinism surface. Presentation tuning belongs in schema-backed `shared/assets/`, not GDScript constants. The only material limit is committed asset coverage: Rogue and Ranger have no native headgear, so their head item remains tint/no-world-mesh per v483. Adding a model to close that gap would be a separate scope decision. No blocking product question is needed for the accepted constrained slice.

The security router classifies this document-only visual/UI planning work as out of scope: no user-controlled rendering, network/input handling, credential, or data-access change is planned. Re-run its scope check if implementation expands into one of those surfaces.

## Ownership and file map

| Role | Paths | Decision |
|---|---|---|
| Class look data | `shared/assets/class_presentations.v0.json`, `.schema.json` | Add only chosen class-specific model proportions/native-part visibility or stance controls, with bounded schema values; keep class gameplay data in `shared/rules/` unchanged. |
| Armor look data | `shared/assets/armor_look.v0.json`, `.schema.json` | Change only if measured gear readability needs a class-specific head/accessory policy or tint strength. Preserve chest-over-belt precedence and existing item coverage. |
| Client presentation | `client/scripts/class_presentations_loader.gd`; a focused `client/scripts/class_silhouette.gd` if native-part visibility needs logic; `client/scripts/character_visual.gd`; `client/scripts/armor_look.gd` if required | Apply stable class look after model replacement, then equipment look. Keep `EquipmentVisualResolver` as equipment owner and rig/socket wiring in `CharacterVisual`. Avoid `main.gd` unless a verified swap path misses refresh. |
| Tests and validation | `client/tests/test_armor_look.gd`, focused new `client/tests/test_class_silhouette.gd` if needed, `client/tests/equipped_gear_fit_probe.gd`, `tools/validate_shared.py` or a focused validator only if JSON schema cannot prove mesh name coverage | Assert actual imported mesh presence/visibility, idempotent equip/unequip, socket binding, and unchanged animation clips without hardcoding current tuning numbers. |
| Bot and capture | Existing `tools/bot/scenarios/client/97_bone_gear_sockets.json` and `105_eye_view_weapon_presentation.json`; `tools/showme/screenshot_catalog.py`, `skills/showme/scripts/render_focus.py`, and a focused capture helper only if existing focuses cannot show normal play zoom | Prefer the existing scenario and suites. Add a semantic bot assertion/scenario only when a visible state cannot be proved by current probes. Do not promote it into `ci_pack.json` without a distinct merge-blocking gap. |
| Documentation/evidence | `docs/CODEMAP.md`, `docs/as-built/v507_class-silhouettes.md`, `docs/as-built/assets/v507/` | Record exact before/after frames, class pair judgments, validated fit, performance evidence/limits, and final asset decision. Coordinator owns `PROGRESS.md` and lifecycle on integration. |

`assets/manifests/assets.v0.json` and `shared/assets/item_presentations.v0.json` are **read/verify** inputs unless a final v503 handoff proves a v507 change necessary. New art is not planned. Potential v503 overlap is highest in the manifest, `equipment_display`, showme captures, and item model tests; compare its final diff path by path before editing any of them. The final content of every overlapping path must include both slices' intended behavior, not merely a clean merge.

## Asset / plugin choice

- **Adopt:** the five committed KayKit hero GLBs and native accessory meshes, existing KayKit weapons/shields, and their manifest provenance.
- **Borrow:** `ClassPresentationsLoader`, `ArmorLook`, `EquipmentVisualResolver`, sockets, kit clips, fit probes, and screenshot capture tooling.
- **Reject:** external art/plugins, uncommitted staged `Rogue_Hooded`, new rig or armor mesh pipeline, and a separate first-person-only model. Respect ADR-0018 D5; do not turn chest/glove/boot armor into attachments.

## Ordered tasks

### 0. Dependency and matched baseline

- [x] Received coordinator notice that v503 is integrated. Transferred its changed `shared/assets/equipment_display.v0.json`, `client/tests/test_item_visuals.gd`, `client/tests/test_loot_node_factory.gd`, `tools/bot/scenarios/client/10_full_equipment.json`, spec/plan/as-built notes, and 16 before/after PNGs from the coordinator's integrated checkout. These paths do not overlap v507's likely class-presentation files. The coordinator CODEMAP included unrelated v501/v505/v506 rows whose source files are not transferred into this worktree, so it was restored to the assigned base; add the v507 row only after new files exist. The v503 dependency baseline passes `make validate-shared` (2,208), `make validate-assets` (451), and `make maintainability`; all copied files retain their exact source hashes. Keep this worktree detached; no sibling cherry-pick/merge.
- [x] Capture the current five-class no-equipment and representative equipped states in the real renderer from the actual play camera, fixed angle/zoom, lighting, quality tier, and pose. The 40-frame idle matrix covers color/grayscale and both zooms; the 30-frame attack matrix names Rogue/Ranger as the closest pair, which remained distinguishable.
- [x] Record a matched current cost baseline if the likely change adds a persistent visible mesh or instancing. N/A: the visual review justified no persistent geometry or presentation-data changes; screenshots are not treated as performance evidence.

Smallest checks: `make regen-screenshots SUITE="skeleton gear"`; inspect `.artifacts/screenshots/latest/` and paired play-camera PNGs. Capture tooling changes, if needed, get their own focused `tools/test_regen_screenshots.py` check.

Source-only asset inventory (not a rendered readability result): `python3 tools/assets/inspect_kit.py client/assets/characters/kaykit/{barbarian,knight,mage,rogue,ranger}.glb` confirms the five models share a 23-joint rig and ~1.94-unit width. Source bounds range from 2.18 to 2.65 units high; Knight/Mage carry cape plus helmet/hat, Barbarian has BearHat, Rogue has cape, and Ranger has cape plus quiver. These findings identify available silhouette cues but do not select a tuning change; make that choice after the matched actual-camera captures are available.

The existing focused screenshot suites use isolated staging cameras (including a 5.2-unit gear-matrix view), so they cannot establish normal play-zoom readability. Added a standalone `client/scripts/showme/class_silhouette_play_camera_capture.gd` harness for one class per capture. It reuses `PlayerCameraController` and reads normal `zoom_default` or tighter battle `zoom_min` from `camera_presentations.v0.json`; both options remain catalog-backed. It supports empty/class-appropriate gear, idle/walk/attack/hit/death poses, grayscale, and saves three time-based attack frames. The harness has not been run or syntax-checked yet because the coordinator has reserved the renderer for v502.

### 1. Select minimal silhouette changes from the committed models

- [x] Compare five class models at identical scale and matched equipment, then choose the smallest data change that separates them. Outcome: current assets already distinguish all five in the matched matrix, so no data change was justified; preserve existing head-item semantics.
- [x] Inspect the selected class parts and rig poses through the play-camera matrix and skeleton gear suite. No weapon/loot obstruction or clipping was observed; Rogue and Ranger retain the documented no-headgear limit.
- [x] If the desired improvement cannot be achieved with committed nodes and existing items, stop and report the unmet criterion. N/A: the measured baseline met the slice's readability criterion; no outside art was needed.

Smallest check: one `showme` gear/skeleton capture for the changed class plus an actual play-camera capture, inspected in grayscale and color.

### 2. Put the chosen look in data and apply it on model refresh

- [x] Add narrowly named class visual parameters and schema constraints only if the review shows a required change. N/A: no class tuning or new asset reference was needed.
- [x] Preserve idempotent class/equipment refresh and the first-person rig. No gameplay presentation path changed; existing armor, item, and animation tests plus the live equipment scenario passed.
- [x] Avoid modifying `main.gd` and `showme/visual_capture.gd` without a concrete missing path. Neither was changed; the capture helper is a focused new file.

Smallest checks: `make validate-shared`; `make validate-assets` if asset IDs/manifest are touched; focused Godot `test_class_silhouette.gd` or `test_armor_look.gd` through `godot --headless --rendering-method gl_compatibility --path client --script res://tests/<test>.gd`.

### 3. Verify gear, motion, and the live client path

- [x] Run existing focused coverage for five-class equipment presentation and animation (`test_armor_look.gd`, `test_animation.gd`, and `test_item_visuals.gd`), plus the real-camera empty/equipped/attack matrix. No new class look code required regression assertions for a modified path.
- [x] Run the live Godot client equipment scenario: headless and visible `97_bone_gear_sockets` passed. The first-person eye-view route was not needed because scale, head visibility, and camera clearance were unchanged.

Smallest checks: focused `test_armor_look.gd`, `test_animation.gd`, and `test_item_visuals.gd`/fit probe; `make bot-client SCENARIO=97_bone_gear_sockets HEADLESS=1`; `make bot-visual scenario=97_bone_gear_sockets` for user-visible visual inspection; `make bot-client SCENARIO=105_eye_view_weapon_presentation HEADLESS=1` when the first-person path is affected. If a scenario is added, use an unused ID after reconciling the integrated catalog and default it to `ci_tier: extended`.

### 4. Real-camera review, performance, and handoff

- [x] Regenerate and inspect the ten-image `skeleton gear` suite and the 40 idle / 30 attack actual-camera matrix in color and grayscale; record that Rogue/Ranger are closest but distinguishable and that no visible pair remains ambiguous.
- [x] If persistent geometry or model scale changed, compare matched renderer cost. N/A: neither changed; the as-built explicitly does not claim a performance gain.
- [x] Run `make maintainability`, focused client tests and bot checks; shared/asset validation also passed in the coordinator's combined `make ci`. Write the as-built with exact results and headgear limits, update CODEMAP for the capture helper, and report the complete handoff manifest.

Final worktree gate is focused slice verification only. The coordinator integrates the ready slice after dependencies, runs combined `make ci` on the fully integrated state, owns `/finish` and commits, then reviews/refactors the batch. No per-slice `make ci`, `make ci-full`, commit, push, branch, or worktree cleanup is part of this worker plan.

## Maintainability and handoff checklist

- `armor_look.gd` is 127 lines, `equipment_visuals.gd` 483, and `character_visual.gd` 124 at the assigned base; keep each edit cohesive. `showme/visual_capture.gd` is 1,252 lines against a 1,235-line grandfathered baseline, and `main.gd` is 6,628 lines; avoid both unless behavior demands it, then meet touch-to-shrink with a real extraction. Recheck the current baseline after v503.
- Before implementation, review v503's integrated result for shared JSON/schema, manifest, equipment display, screenshots and tests. Note any exact overlaps and resolution in the as-built.
- Preserve proof needed for coordinator integration: changed/untracked file manifest; ignored screenshot paths; before/after play-camera captures; class/gear matrix; focused test/bot output; renderer metrics with method and sample counts. A clean `git status` is not sufficient evidence of visual quality.
