# v532 Plan — Item tooltip without icon preview

- **Spec:** [`v532_spec-tooltip-no-icon.md`](../specs/v532_spec-tooltip-no-icon.md) (approved 2026-10-02)
- **Baseline:** `dab60eb5` on `main`. No prerequisite slices. Standalone run, no branch, no worktree.
- **Goal:** `ItemTooltipPanel` shows text only; the text column spans `CONTENT_WIDTH`; no consumer or debug accessor changes behavior.

## Spec review gate (result: pass, two recorded corrections)

- **Scope and non-goals:** client presentation only. No protocol, schema, golden, server, replay, rules or asset-manifest change, so no determinism, `validate_shared` or golden work. Standalone spec-gate exemption does not apply because a spec exists; it is kept.
- **Open question resolved (spec scope item 4):** `setup(...)` has 11 production call sites (`inventory_panel` 2, `stash_panel` 2, `shop_panel` 3, `market_panel` 2, `material_wallet_panel` 2) plus 2 direct test calls and the shop tests. That is above the spec's 12-edit threshold once multi-line calls are counted, and several pass `price`, `affordable` and `affinity_lines` positionally. **Decision: keep the signature, rename the two dead parameters `_item_presentations` and `_fallback_label`, add a doc comment, and record "prune the signature" as a follow-up.** This is the spec's stated default, not a scope change.
- **Correction to the spec's verification list:** there is no existing tooltip screenshot suite (`make regen-screenshots-list` has none, and `showme_rarity_ui_fixtures.gd` does not build tooltips). The plan adds a small `item-tooltip` showme focus (Task 4) rather than assuming one exists.
- **Size ratchet:** `item_tooltip_panel.gd` is 391 lines, under 600, not grandfathered, so no baseline change. The new capture driver stays far below 600.
- **Asset and plugin decision:** adopt `ItemTooltipPanel`, `InventoryPanelStyles` and the showme harness; borrow the `showme_character_screen_capture.gd` driver pattern; reject any replacement preview and any plugin or asset.
- **Sibling overlap:** none. v533 to v537 touch camera, dungeon rules, light, server and capture evidence, not the tooltip. `docs/CODEMAP.md` is the only shared doc and is touched at most once, in Task 6.

## File map

| File | Change |
|---|---|
| `client/scripts/item_tooltip_panel.gd` | Remove `ItemPreview`, its instantiation, dead constants and imports, and the `top_row` wrapper; use `CONTENT_WIDTH` for main lines; underscore the two dead `setup` parameters |
| `client/tests/test_inventory_tooltip_content.gd` | Add assertions: no preview/icon-drawing node, and main-line labels are `CONTENT_WIDTH` wide, for each of the five rarities |
| `client/scripts/showme/showme_item_tooltip_capture.gd` | New capture driver (one tooltip per rarity) |
| `skills/showme/scripts/render_focus.py`, `tools/showme/screenshot_catalog.py`, `tools/test_regen_screenshots.py` | Register the `item-tooltip` focus and suite, with a catalog test |
| `skills/showme/SKILL.md` | One row in the focus table |
| `docs/as-built/v532_tooltip-no-icon.md`, `docs/progress/slice-lifecycle.md`, `PROGRESS.md` | As-built evidence and closeout (Task 6) |

No ignored raw evidence is expected apart from the screenshots under `.artifacts/`; the before/after PNGs referenced by the as-built are copied to `docs/as-built/assets/v532/`.

## Tasks

- [x] **1. Baseline capture and tests (before any edit).**
  - Run the focused Godot gates against unmodified `main` and record they pass: `test_inventory_tooltip_content`, `test_shop_panel`, `test_shop_tooltip_stability`, `test_stash_panel`, `test_character_stats_panel`, `test_look_and_feel_polish`. Run each as the `run_gate` line in `scripts/client_smoke.sh` does (`$GODOT --headless --path client --script res://tests/<name>.gd`).
  - Grep to confirm no other code depends on the removed geometry: `rg "ItemPreview|PREVIEW_SIZE|PREVIEW_GAP|MAIN_STAT_WIDTH|ICON_FONT_SIZE" client`.
  - Check: all six gates print their `[gdtest] PASS` sentinel.
- [x] **2. Write the failing tests first.**
  - In `test_inventory_tooltip_content.gd`, build tooltips for a common, magic, rare, unique and set item (reuse the file's existing fixtures) and assert: no descendant is an `ItemPreview` and no `Control` outside `Label`, `HBoxContainer`, `VBoxContainer`, `Control` spacers and rules overrides `_draw` for icons (assert on the absence of the `ItemPreview` class name and of any node exposing `item_presentations`); every main-line label has `custom_minimum_size.x == ItemTooltipPanel.CONTENT_WIDTH`; the empty-item tooltip keeps its single text line at `CONTENT_WIDTH`.
  - Check: the new assertions fail on current code with the preview present.
- [x] **3. Implement in `item_tooltip_panel.gd`.**
  - Delete the `ItemPreview` inner class, the `has_item` preview block, `PREVIEW_SIZE`, `PREVIEW_GAP`, `MAIN_STAT_WIDTH`, `ICON_FONT_SIZE` and `ItemIconDrawerScript`; delete `RarityCuePresenterScript` if `rg` shows no other use in the file.
  - Replace `top_row` with `main_stats` added directly to `root` when it only holds `main_stats`; every use of `MAIN_STAT_WIDTH if has_item else CONTENT_WIDTH` becomes `CONTENT_WIDTH`. Keep the header rule logic unchanged.
  - Rename the unused `setup` parameters to `_item_presentations` and `_fallback_label` and add a one-line comment that they are kept for caller compatibility.
  - Check: Task 2 tests pass; `wc -l client/scripts/item_tooltip_panel.gd` is below 391; no `UNUSED_PARAMETER` or parse warnings in the Godot output.
- [x] **4. Visual capture driver and registration.**
  - Add `client/scripts/showme/showme_item_tooltip_capture.gd` modelled on the character-screen driver: dark backdrop, `ItemTooltipPanel.setup` with real `InventoryTooltipContent` lines for one item per rarity, `--rarity` argument, 8 settle frames, PNG output.
  - Register an `item-tooltip` focus in `render_focus.py`, a `SuiteSpec` plus job discovery in `tools/showme/screenshot_catalog.py` (five jobs, one per rarity), a catalog assertion in `tools/test_regen_screenshots.py`, and a row in `skills/showme/SKILL.md`.
  - Capture the **before** set by stashing Task 3 (`git stash` is not used; instead check out the baseline file into a temp copy, or capture once from `dab60eb5` before Task 3 if Task 1 is extended to run the new driver against the old panel). Capture the **after** set on the final code. Both use the real renderer, not headless.
  - Check: `.venv/bin/pytest tools/test_regen_screenshots.py -v`; `make regen-screenshots SUITE="item-tooltip"` writes five PNGs; viewed images show text only, no empty right-hand strip, no clipped or overlapping lines, and rarity border, name color and `Rarity:` line intact. Godot shutdown leak warnings are noted in the as-built, not suppressed.
- [x] **5. Focused regression run.**
  - Rerun the six Godot gates from Task 1 and `make maintainability`, `git diff --check`.
  - Do **not** run `make ci`, `make ci-full` or `make test-all`: no shared contract, protocol, golden or server change (Testing Scope Policy). `make lint-determinism` is not needed because no `game/` file is touched.
  - Check: all gates pass; maintainability reports no new or grown over-limit file.
- [x] **6. As-built and closeout.**
  - Write `docs/as-built/v532_tooltip-no-icon.md` with: the removal, the kept-signature decision and its follow-up, the before/after captures (copied to `docs/as-built/assets/v532/`), the exact commands and outcomes, and the recorded loss of the tooltip-level non-hue slot cue (the `Rarity:` line, name color and border remain; the hovered slot keeps its v508 cue).
  - Add the lifecycle row, update `PROGRESS.md` **Current status** and the codename index, and touch `docs/CODEMAP.md` only if a role description changed (the new capture driver is indexed there if the showme row lists drivers).
  - Check: `make validate-shared` (cheap; validates CODEMAP) and `git status` shows only the files in the file map.

## Final gate (standalone)

Focused verification only (Task 5), because the change is client-presentation with no shared-contract effect. A pre-PR `make ci` is not claimed. When the whole six-slice set is done, one combined `make ci` runs on the final state, then `/finish`, `/review` and `/refactor`, as in the batch workflow.

## Handoff and risks

- Complete change set is the file map above. Nothing is committed in this slice until the user asks (commit message via `commit-commands:commit`).
- **Risk:** a test or panel that measured the old tooltip height. Task 1 grep plus the six gates cover this; if one fails, fix the assertion to derive from `CONTENT_WIDTH` rather than pin a new pixel value (Test Locking Policy).
- **Risk:** the `item-tooltip` capture plumbing grows beyond the slice. Cap it at the five files listed; if `render_focus.py` needs more than a mapping entry, fall back to a one-off capture command and document it instead of registering a suite.
