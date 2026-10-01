# v516 Plan — character-screen-redesign

- **Spec:** [`docs/specs/v516_spec-character-screen-redesign.md`](../specs/v516_spec-character-screen-redesign.md)
- **Date:** 2026-10-01
- **Baseline commit:** `425b9ae4` (detached worktree, no branch)
- **Prerequisite slices:** v514 `ui-theme-foundation` (blocks Phase P5 only); v515 `inventory-tooltip` (blocks Phase P4's `inventory_panel.gd` wiring only).

## Spec review gate (recorded)

- Scope/non-goals: display-only, no protocol/schema/golden/Go/replay surface. Determinism, server authority, world presets: not touched.
- Shared data: only additive strings in `shared/i18n/en.json`. Group membership is code-owned presentation ordering (spec D2); recorded as a deliberate exception to the data-driven policy because it is not balance tuning.
- Acceptance-to-test mapping: AC1 -> `test_character_stats_header.gd` + captures; AC2 -> `test_character_stat_groups.gd` + existing dual-wield test; AC3/AC4 -> unchanged `test_character_stats_panel.gd`, `test_coop_client.gd`, `test_client_bot.gd`, bot `character_stats_panel` (file `client/09_character_stats_panel.json`); AC5 -> `test_paper_doll_backdrop.gd` + bot `client_inventory_paper_doll`; AC6 -> showme suite; AC7 -> ratchet tests; AC8 -> P5 grep gate.
- Bot scenarios: none added. Display-only; existing scenarios `18_character_stats_and_leveling`, `client/09_character_stats_panel`, `client/13_inventory_paper_doll` are the regression net and stay unmodified (movement-audit untouched).
- Minor spec correction: none needed.
- No product decision blocks planning; Open questions 1-4 have defaults used below.

## What can proceed before v514 / v515

| Phase | Needs | Status |
|-------|-------|--------|
| P0 Baseline captures | nothing | **Can start now** |
| P1 Pure modules (groups, header model) | nothing | **Can start now** |
| P2 Header + groups + allocation rows in `character_stats_panel.gd` | nothing (colors via `CharacterPanelStyles` seam) | **Can start now** |
| P3 Paper-doll backdrop node + slot styles in `inventory_panel_styles.gd` | nothing (node is standalone) | **Can start now** (do not touch `inventory_panel.gd`) |
| P4 Wire backdrop into `inventory_panel.gd` | **v515 integrated** into this worktree | **BLOCKED on v515** |
| P5 `UiTheme` token migration | **v514 integrated** into this worktree | **BLOCKED on v514** |
| P6 Captures, as-built, handoff | P2-P3 (P4/P5 for final) | Interim captures now; final after P4/P5 |

Until blocked phases unblock, all hard-coded colors/fonts introduced in P1-P3 live **only** in `character_panel_styles.gd` / `inventory_panel_styles.gd` (named constants/functions), so P5 is a one-seam swap and not a hunt.

## File map and ownership

New (all <=600 lines, independently importable; typed inputs, no `globals()` helpers):

- `client/scripts/character_stat_groups.gd` — `class_name CharacterStatGroups extends RefCounted`; static `GROUPS` (ordered `[{id, title_key, keys}]`), `group_for(key)`, `ordered_keys()`. Pure.
- `client/scripts/character_stats_header.gd` — `class_name CharacterStatsHeader extends VBoxContainer`; `configure(hero_name, progression)`; owns `ClassIcon`, name/class/level, XP bar, points badge; `get_header_state() -> Dictionary`; pure static `xp_fraction(progression)` and `class_accent(class_id)` (reads `ClassPresentationsLoader`).
- `client/scripts/paper_doll_backdrop.gd` — `class_name PaperDollBackdrop extends Control`; `configure(class_id)`; `_draw` card + silhouette watermark + slot connectors; takes slot anchor points as a constructor/`configure` argument (dictionary of slot -> Vector2), not from `inventory_panel.gd`.
- Optional `client/scripts/showme/character_screen_capture.gd` — headless-less capture driver for the UI suite (mirrors `showme_rarity_cues_capture.gd`).
- Tests: `client/tests/test_character_stat_groups.gd`, `test_character_stats_header.gd`, `test_paper_doll_backdrop.gd`.

Edited:

- `client/scripts/character_stats_panel.gd` (550) — **must end <=600**; moves header/group building out; keeps `get_debug_state()` shape.
- `client/scripts/character_panel_styles.gd` (21) — color/font/section/chip/bar styles seam.
- `client/scripts/inventory_panel_styles.gd` (91) — paper-doll card + slot frame styles.
- `client/scripts/inventory_panel.gd` (**grandfathered 1623**) — P4 only, net <=0 lines.
- `shared/i18n/en.json` — additive strings.
- `scripts/client_smoke.sh` — register 3 new tests.
- `tools/showme/screenshot_catalog.py`, `tools/test_regen_screenshots.py` — UI suite (if Open question 3 default holds).
- `docs/CODEMAP.md` — new files under Classes / Client UI utilities rows.
- `docs/as-built/v516_character-screen-redesign.md` — evidence.

## Shared-file conflicts and handoff notes

| File | Also touched by | Handling |
|------|-----------------|----------|
| `client/scripts/inventory_panel.gd` | v515 (extracting; moves/relines the paper-doll block) | P4 waits for v515. Re-apply the swap on the integrated file; never overwrite. Net <=0 lines. |
| `client/scripts/inventory_panel_styles.gd` | v514 (tokens), likely v515 | Add paper-doll functions only; keep edits additive and localized; rebase on integrated file. |
| `client/scripts/character_panel_styles.gd` | v514 (theme adoption) | v516 adds named style functions; v514 may rewire internals to `UiTheme`. P5 reconciles. |
| `client/scripts/draggable_window.gd` | v514/v517 possibly | v516 does **not** edit it. |
| `client/scripts/stat_tooltip_label.gd` | v514/v515 (tooltip theme) possibly | v516 reads only; if P5 needs token edits, do it after v514. |
| `shared/i18n/en.json` | any slice adding strings | Additive keys under a `character_screen.*` prefix; merge by key. |
| `tools/showme/screenshot_catalog.py`, `tools/test_regen_screenshots.py` | v514/v515/v517 if they add UI suites | Coordinator picks single owner/suite name; else append-only entry. |
| `scripts/client_smoke.sh`, `docs/CODEMAP.md` | all client slices | Append-only lines. |
| `client/scripts/main.gd` | v517 | v516 does not edit it (panel API unchanged). |
| `shared/assets/ui_theme.v0.json` | v514 owns | v516 only consumes. |

## Tasks

Smallest runnable check follows each task. Godot invocation: `godot --headless --path client --script res://tests/<file>.gd` (run `godot --headless --path client --import` once in the fresh worktree and revert rewritten `.glb.import` files before handoff).

### P0 Baseline

- [x] **P0.1** Capture "before" images: character panel (`C`) and inventory paper-doll for all five classes at the base commit, using showme/`regen-screenshots` or a one-off capture driver. Save under `.artifacts/screenshots/v516-before/`. Check: files exist, panel visible and non-blank.
- [x] **P0.2** Record the baseline pass of existing tests: `test_character_stats_panel.gd`, `make client-unit` subset for stats/inventory, bot `character_stats_panel` (file `client/09_character_stats_panel.json`) and `client_inventory_paper_doll`. Check: all green before any edit.

### P1 Pure modules (TDD)

- [x] **P1.1** Write `test_character_stat_groups.gd` (every `CharacterStatsPanel.DERIVED_LABELS` key in exactly one group; no unknown keys; order stable; titles resolve through `TextCatalog`), then `character_stat_groups.gd`. Check: that test passes.
- [x] **P1.2** Write `test_character_stats_header.gd` (XP fraction incl. `experience_to_next_level == null` max-level, zero-xp, empty class fallback, accent per class from `class_presentations`, points badge visibility), then `character_stats_header.gd`. Check: that test passes.
- [x] **P1.3** Add `character_screen.*` strings to `shared/i18n/en.json`. Check: `python tools/validate_i18n.py`.
- [x] **P1.4** Add named style/color constants and style builders to `character_panel_styles.gd` (section header, row card, XP bar bg/fill, points chip, `+` button highlight). No raw color literals anywhere else. Check: panel still constructs (`test_character_stats_panel.gd`).

### P2 Stats panel redesign

- [x] **P2.1** Replace the three text lines with `CharacterStatsHeader` in `_build()`; keep `_hero_name`, `_title_text()`, `set_hero_name`, `set_progression`. `Level/XP/Points` debug fields keep working (`progression` copy in debug state). Check: existing `test_character_stats_panel.gd` green.
- [x] **P2.2** Restyle the four base-stat rows (D3); keep `_stat_value_labels/_stat_base_labels/_stat_effective_labels/_stat_buttons` maps and their debug strings. Check: stat label/`stat_columns`/button assertions in existing tests green.
- [x] **P2.3** Render derived rows under group headers via `CharacterStatGroups`; move the MAIN/OFF header into the Offense header; keep `_derived_name_labels/_derived_labels/_derived_off_labels` keyed by stat key, `derived_title`, `derived_columns`, scroll debug state, and the Ranger tooltip override. Add additive `derived_groups` and `header` debug keys. Check: `test_character_stats_panel.gd` (incl. dual-wield) and `test_coop_client.gd` character tests green.
- [x] **P2.4** Size/typography pass (D4): 3-step type scale, content fits at 1280x720; `layout_key` unchanged. Check: window-chrome tests in `test_coop_client.gd` green; manual capture at 1280x720 has no clipping.
- [x] **P2.5** Maintainability: confirm `wc -l character_stats_panel.gd` <=600; if over, extract rather than compress. Check: `.venv/bin/pytest tools/test_file_size_ratchet.py tools/test_extraction_coupling_ratchet.py`.
- [x] **P2.6** Run bot `character_stats_panel` (file `client/09_character_stats_panel.json`) headless. Check: pass, unmodified scenario.

### P3 Paper-doll node (no `inventory_panel.gd` edits)

- [x] **P3.1** Write `test_paper_doll_backdrop.gd` (constructs standalone, `configure(class_id)` for all five classes + unknown, name/visibility contract, no import of `inventory_panel.gd`), then `paper_doll_backdrop.gd`; add `paper_doll_*` styles to `inventory_panel_styles.gd`. Check: that test passes.
- [x] **P3.2** Register the three new tests in `scripts/client_smoke.sh`. Check: `bash -n scripts/client_smoke.sh` and each test runs via its `[gdtest] PASS` line.

### P4 Paper-doll wiring — BLOCKED until v515 is integrated into this worktree

- [x] **P4.1** Transfer the integrated v515 `inventory_panel.gd` (three-way, without overwriting v516 files). Re-read the paper-doll block at its new location.
- [x] **P4.2** Replace the inline `Panel` construction with `PaperDollBackdrop` (name `character_paper_doll` preserved), pass slot anchors and class id from `character_progression`; re-configure on class change. Net <=0 lines. Check: `wc -l` <= baseline (lower baseline if file shrank per ratchet rule 3), `test_stash_panel.gd` paper-doll assertions, inventory unit tests, bot `client_inventory_paper_doll`.
- [x] **P4.3** If a screenshot shows slot/backdrop overlap, adjust `PAPER_DOLL_SLOT_POSITIONS` only in the same edit; slot ids untouched. Check: `assert_paper_doll_layout` still green.

### P5 UiTheme migration — BLOCKED until v514 is integrated into this worktree

- [x] **P5.1** Transfer integrated v514 (`UiTheme`, `shared/assets/ui_theme.v0.json`, schema/validator). Run its focused dependency checks (`tools/validate_shared.py` theme checks, its unit test).
- [x] **P5.2** Point `character_panel_styles.gd` and the paper-doll style functions at `UiTheme` tokens; drop local constants; map class accents to the theme where v514 defines them, else keep `class_presentations` accents. Check: `test_character_stats_panel.gd`, `test_paper_doll_backdrop.gd` green.
- [x] **P5.3** Grep gate: no `Color("#` literals or numeric `font_size` overrides in `character_stats_panel.gd`, `character_stats_header.gd`, `character_stat_groups.gd`, `paper_doll_backdrop.gd`, `character_panel_styles.gd` (document any justified exception). Check: `rg -n 'Color\("#|font_size", [0-9]' <files>` returns only allowed items.
- [x] **P5.4** Re-run P2.6 and P4.2 checks after theming.

### P6 Visual proof and handoff

- [x] **P6.1** Add the UI capture suite (`character-screen`) to `tools/showme/screenshot_catalog.py` (+ capture driver if needed) and update `tools/test_regen_screenshots.py`. Check: `.venv/bin/pytest tools/test_regen_screenshots.py` and `make regen-screenshots-list`.
- [x] **P6.2** Run `make regen-screenshots SUITE="character-screen"` in a real render window: five classes x (points / no points) x (single / dual-wield for at least two classes) for the panel, and five paper-doll captures; plus 1280x720. Output under `.artifacts/screenshots/<timestamp>/`. Inspect for clipping, contrast, overlap. Interim run after P3; final run after P4/P5.
- [x] **P6.3** Update `docs/CODEMAP.md` (new files), `docs/as-built/v516_character-screen-redesign.md` (before/after, commands, outcomes, limits such as 3D paper-doll deferred), and this plan's checkboxes. Do not edit `PROGRESS.md` or the lifecycle index (coordinator owns).

## Final gate — focused slice verification (no `make ci`)

```bash
godot --headless --path client --script res://tests/test_character_stat_groups.gd
godot --headless --path client --script res://tests/test_character_stats_header.gd
godot --headless --path client --script res://tests/test_paper_doll_backdrop.gd
godot --headless --path client --script res://tests/test_character_stats_panel.gd
make client-unit
make bot-client SCENARIO=character_stats_panel HEADLESS=1
make bot-client SCENARIO=client_inventory_paper_doll HEADLESS=1
.venv/bin/pytest tools/test_file_size_ratchet.py tools/test_extraction_coupling_ratchet.py tools/test_validate_codemap.py tools/test_regen_screenshots.py
python tools/validate_i18n.py && make validate-shared && make maintainability
make regen-screenshots SUITE="character-screen"
git diff --check
```

The coordinator runs the single combined `make ci` after all accepted slices integrate; this slice does not run `make ci`/`make ci-full`, commit, push, or `/finish`.

## Handoff to coordinator

Report: base `425b9ae4`, dependencies used (v514/v515 integrated commit ids if applied), dirty state; complete changed/deleted/untracked list (new scripts, tests, i18n, smoke registration, showme catalog, CODEMAP, as-built); ignored evidence to preserve (`.artifacts/screenshots/v516-*`); exact commands + outcomes; the shared-file conflict table above updated with actual overlaps; any acceptance criterion not met (state AC8 explicitly if v514 never integrated). Revert Godot-rewritten `.glb.import` churn first. Session stays open for fixes.

## Execution notes (P0-P3, 2026-10-01)

- Stat-panel debug state keeps the global NAME/VALUE|MAIN/OFF column header above the scroll; group titles are rows inside the scroll. The spec's "MAIN/OFF in the Offense header" was not done: it would change `derived_columns` semantics for no gain (recorded deviation).
- Capture driver is `client/scripts/showme/showme_character_screen_capture.gd` via a new `--focus character-screen` (+ `--variant`) in `skills/showme/scripts/render_focus.py`; paper-doll variant is an interim standalone render (backdrop + plain slot buttons at `InventoryPanel` constant geometry) until P4.
- Paper-doll finding for P4.3: existing slot geometry overlaps (Head/Amulet, Gloves/Boots) and the central column hides the body card; `paper_doll_backdrop.gd` therefore spans the whole slot field. A slot reflow should be decided when `inventory_panel.gd` is reachable.
- `.glb.import` churn from `--import` was reverted with `git checkout`.
- P5 (2026-10-01): colors/sizes/frames moved to `shared/assets/ui_theme.v0.json` under `character_*` entries (10 colors, 9 spacing, 3 fonts, 6 frames); `CharacterPanelStyles`/`InventoryPanelStyles` are UiTheme delegates plus accent-derived builders. Accent darken/alpha factors stay as named constants in the styles files (class-accent math, not tokens). Grep gate residue: `Color(1,1,1,EMBLEM_ALPHA)` modulate and a `Color(0,0,0,0)` "no accent" sentinel. `stat_tooltip_label.gd` still carries its own colors/font size (v514 left it unmigrated; not in scope).
- P4/P6 (2026-10-01): backdrop wired into `inventory_panel.gd` (1119 -> 1111, baseline lowered to 1111); `PAPER_DOLL_SLOT_POSITIONS` moved to new `paper_doll_layout.gd` (slot reflow to a non-overlapping grid, P4.3); capture driver now renders the real `InventoryPanel`. As-built: `docs/as-built/v516_character-screen-redesign.md`. `make client-unit` run once on the final state: PASS.
