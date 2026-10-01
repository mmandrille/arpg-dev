# v516 Spec — character-screen-redesign

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)
- **Date:** 2026-10-01
- **Batch base:** `425b9ae4` (detached batch worktree, removed; evidence preserved under `.artifacts/batch-v511-v517-evidence/v516-character-screen/`))
- **Dependency:** v514 `ui-theme-foundation` (`UiTheme` + `shared/assets/ui_theme.v0.json`) — **not yet available**; only the theme-token migration is blocked on it (see Plan).
- **Siblings touching nearby files:** v515 (extracting from `inventory_panel.gd`), v517 (HUD/hotbar), v514 (theme).

## Purpose

The character stats panel (`C`) and the inventory paper-doll are functional but flat: a 23 px font everywhere, three loose text lines (`Level / XP / Points`), a four-row stat table, and one 17-row undifferentiated "Derived" scroll list. Nothing says which class you are beyond a title suffix, and the paper-doll is ten buttons around a blank gray rectangle. v516 redesigns both for readability: class-aware header, grouped derived stats, clearer stat-allocation rows, and a paper-doll with a class-aware backdrop and consistent slot framing.

Display-only. No protocol, server, shared-rules, or golden change. All numbers shown are already in `character_progression` (`level`, `experience`, `experience_to_next_level`, `unspent_stat_points`, `base_stats`, `effective_base_stats`, `derived_stats`, `stat_breakdowns`, `character_class`).

## Non-goals

- No new stats, formulas, tooltips content, or breakdown logic (`CharacterStatsBreakdown` stays the source of tooltip text; only presentation of its output may move).
- No new protocol fields, server events, or character-class data. No change to stat-point allocation rules (`allocate_stat_requested` signal unchanged).
- No inventory bag/grid/tooltip/drag-drop changes (v515 owns those). No HUD/XP bar/hotbar changes (v517 owns those).
- No new 3D character render in the paper-doll (see Decision D3). No new art assets.
- No theme system of our own: colors/fonts/styleboxes move to `UiTheme` when v514 lands, not before.

## Current-state inspection (adopt / borrow / reject)

In-repo findings at base `425b9ae4`:

- `client/scripts/character_stats_panel.gd` (550 lines, **not** grandfathered; ~50 under the 600 cap). Builds a `DraggableWindow` (`layout_key "character_stats"`, 330x585). Hard-coded `Color("#d8c7a6")`, `#c9a227`, font size 23/24/15. `DERIVED_LABELS` is an ordered dict of 17 keys with no grouping. Dual-wield damage adds a third MAIN/OFF column and re-lays out widths in `_apply_derived_header_layout`.
- `character_stats_breakdown.gd` (222): pure formatting/breakdown helpers; no UI. Leave behavior alone.
- `character_panel_styles.gd` (21): single `panel_style()`; the natural seam for later `UiTheme` delegation.
- `stat_tooltip_label.gd`: `StatTooltipLabel` with the custom tooltip panel and `apply_effective_stat_style` (boost/penalty colors).
- `class_icon.gd` + `shared/assets/class_presentations.v0.json`: per-class icon shape, `color`, `accent`, plus a KayKit hero `model` entry. Already the single source of class identity color/shape. `character_select_panel.gd` and `class_creation_summary.gd` already use it.
- Paper-doll lives in `inventory_panel.gd` (**1,623 lines, grandfathered at baseline 1623**, ratchet rule 3/4): `PAPER_DOLL_SLOT_POSITIONS`, a `Panel` named `character_paper_doll` styled by `InventoryPanelStyles.paper_doll_style()` (91-line file, not grandfathered), 10 `InventorySlotButton`s at `EQUIPMENT_SLOT_SIZE (96,58)`, weapon-set tabs from `WeaponSetTabs`.
- ADR-0018: kit visuals, armor = tints, **D9 screenshot harness is the visual gate**. `tools/showme/screenshot_catalog.py` has suites for classes/items/scenes but **no UI-panel suite**.
- Debug-state contract consumed by tests and bot steps: `CharacterStatsPanel.get_debug_state()` keys and string formats (`"STR  25 / 42"`, `"Hit chance  50%"`, `stat_columns`, `derived_columns`, `derived_title`, `derived_scroll`, `stat_tooltips`, `stat_mouse_filters`, `window`), used by `client/tests/test_character_stats_panel.gd`, `test_coop_client.gd`, `test_client_bot.gd`, `assert_character_stats_panel_visible`, `assert_stat_button_enabled`, `click_stat_button`. Paper-doll: `get_debug_state()["paper_doll_slot_ids"|"paper_doll_slots"|"paper_doll_preview"]` used by `assert_paper_doll_layout` and `test_stash_panel.gd`.

Decisions:

| Candidate | Decision |
|-----------|----------|
| `ClassIcon` + `class_presentations.v0.json` (icon shape/color/accent) as header identity and class accent | **Adopt.** No new class data. |
| `DraggableWindow`, `StatTooltipLabel`, `CharacterStatsBreakdown`, `TextCatalog` | **Adopt** unchanged. |
| KayKit hero model render (`character_visual.gd`, `class_presentations` `model`) as a live 3D paper-doll in a `SubViewport` | **Reject for this slice.** Needs a viewport, lighting, per-frame cost on a panel, and an animated model; it is a presentation feature with perf risk (v495-v497 just fought hitches) and v515 is simultaneously restructuring the same file. Revisit as a follow-up slice. |
| `UiTheme` tokens | **Adopt after v514** (blocked). Until then colors/fonts go through `CharacterPanelStyles`/`InventoryPanelStyles` only. |
| Third-party Godot UI plugins | **Reject.** Nothing needed beyond built-in containers. |
| Collapsible derived groups | **Reject** (adds state and bot-assertion surface; scroll + headers suffice). |

## Design decisions

**D1 — Header (class-aware).** A header card above the stats: `ClassIcon` (48 px, configured from `character_class`), hero name, class display name (`character.class.<id>` via `TextCatalog`, unchanged key), a `Level N` chip, and a thin XP progress bar with `XP x / y` text computed from `experience` + `experience_to_next_level` (`remaining == null` => max level, full bar, no "+n"). An unspent-points badge appears only when `unspent_stat_points > 0` and uses the class accent from `class_presentations`. Class tint is applied only to the icon ring, bar fill, and badge; panel chrome stays neutral so every class remains legible. Unknown/empty class falls back to the existing `fallback_class` behavior of `ClassIcon`/loader and a neutral accent.

**D2 — Grouped derived stats.** Derived stats render under four section headers instead of one flat list:

| Group | Keys (existing `DERIVED_LABELS` keys, unchanged names) |
|-------|------|
| Offense | `damage_min`, `damage_max`, `ranged_damage_bonus_percent`, `attack_speed`, `attack_interval_ticks`, `hit_chance`, `crit_chance`, `crit_damage` |
| Defense | `armor`, `evade_chance`, `block_percent` |
| Vitals | `max_hp`, `max_mana`, `health_regen_per_second`, `mana_regen_per_second` |
| Utility | `movement_speed`, `light_radius` |

Row order inside a group follows today's `DERIVED_LABELS` order; the dual-wield MAIN/OFF columns stay on the Offense damage rows only, with the column header moved into the Offense header row instead of a global header. `ranged_damage_bonus_percent` keeps its existing conditional-Ranger tooltip override. Group membership is **code-owned** in a small pure module (`character_stat_groups.gd`): it is presentation ordering, not balance tuning, so the Data-Driven Configuration Policy's note requirement is satisfied here by recording this decision; there is no shared-JSON file for it unless the owner asks (Open question 2). Any derived key present in `progression.derived_stats` but not in a group is not shown (matches today's behavior); a validator-free unit test asserts every `DERIVED_LABELS` key is in exactly one group so a new stat cannot silently vanish.

**D3 — Stat allocation rows.** The four base stats become cards/rows with name, `base`, effective (colored boost/penalty via existing `apply_effective_stat_style`), and the `+` button. The `+` stays at the end of the row and keeps `focus_mode NONE`, `pressed -> allocate_stat_requested(stat)`, disabled unless `allocation_enabled and points > 0`. When points are available, rows get a subtle emphasized affordance (button highlight) so the player can see where to spend. Header column labels (`NAME/BASE/EFFECTIVE`) remain reported in `stat_columns`.

**D4 — Sizing/typography.** Replace uniform 23 px with a three-step scale (title / row / caption) and fit the content to the existing `DraggableWindow` min size or a documented new one. Window `layout_key "character_stats"` stays, so saved window positions are not invalidated; if the content size changes, `clamp_to_viewport` already handles smaller viewports. Must be readable at the project's smallest supported viewport (verify at 1280x720 in the capture).

**D5 — Paper-doll (inventory).** Minimum-touch redesign of the *visual* surface only:
1. A new self-contained node `paper_doll_backdrop.gd` replaces the blank `Panel` as `_paper_doll_preview` (still named `character_paper_doll`, still `exists`/`visible` in debug state). It draws: a rounded dark card, a class-tinted `ClassIcon` watermark/silhouette centered, and thin connector lines from each slot to a central body axis. It takes class id via `configure(class_id)`; it imports nothing from `inventory_panel.gd` and is unit-testable alone (Extraction independence rule).
2. Slot frames: label/empty-state/rarity styling stays in `InventoryPanelStyles` (91 lines, not grandfathered; room to grow). Slot geometry (`PAPER_DOLL_SLOT_POSITIONS`, `EQUIPMENT_SLOT_SIZE`) stays unless a screenshot review shows overlap; any move must keep slot ids and the `assert_paper_doll_layout` contract.
3. `inventory_panel.gd` receives only the minimal wiring to construct the backdrop node and pass the class id — a **net-zero or negative line count** edit (swap ~6 lines of inline `Panel` construction for a 2–3 line factory call). Must not grow past baseline 1623.

**D6 — Text/i18n.** New visible strings (group titles, "Points available", "Max level") go in `shared/i18n/en.json` through `TextCatalog.get_text(key, fallback)`; `tools/validate_i18n.py` must pass. Debug-state strings keep today's English formats so existing assertions don't change.

## Acceptance criteria (observable)

1. Opening `C` shows a header with the class icon, name, class name, level, an XP bar whose fill matches `experience / (experience + experience_to_next_level)`, and a points badge visible iff `unspent_stat_points > 0`; verified for all five classes in the capture and by unit test of header state (class id -> accent, XP fraction, max-level case, empty-class fallback).
2. Derived stats appear in four titled groups per D2; every key in `DERIVED_LABELS` is in exactly one group; dual-wield MAIN/OFF columns still show for damage rows with correct main/off values; no row is clipped at the default window size.
3. `get_debug_state()` retains all existing keys and string formats; the existing `test_character_stats_panel.gd`, `test_coop_client.gd` character-panel tests, and `test_client_bot.gd` panel tests pass **unmodified** (additive debug keys such as `header` and `derived_groups` are allowed).
4. Allocation `+` buttons behave as before (enabled/disabled, signal, pause-menu disable); bot scenario `client_character_stats_panel` (`tools/bot/scenarios/client/09_character_stats_panel.json`) passes unmodified.
5. Paper-doll: backdrop is class-tinted for the active class, all ten slot ids render and remain interactive; `client_inventory_paper_doll` (`client/13_inventory_paper_doll.json`) passes unmodified; drag/drop of equipment onto slots unchanged (existing inventory unit tests + `test_stash_panel.gd` paper-doll assertions pass).
6. Screenshots: real-renderer captures of the character panel for each of the five classes (with/without unspent points, dual-wield and single-weapon) and the inventory paper-doll for each class are produced and attached to the as-built via the showme/`regen-screenshots` harness (ADR-0018 D9). Before/after pair for at least one class.
7. Maintainability: `character_stats_panel.gd` ends at or below 600 lines (extract header and group renderers rather than growing it); no new file over 600; `inventory_panel.gd` does not exceed baseline 1623 (target: shrink); no new `helpers=globals()` sites; `make maintainability` focused check passes.
8. After v514 integrates: no raw `Color("#...")` literals or hard-coded font sizes remain in the new/changed character-screen files; they resolve through `UiTheme` tokens (spot-checked by a grep gate in the plan). Visual output matches the pre-theme captures within the token palette.

## Likely surfaces

- **Client (new):** `client/scripts/character_stats_header.gd`, `client/scripts/character_stat_groups.gd`, `client/scripts/paper_doll_backdrop.gd`, optional `client/scripts/showme/character_screen_capture.gd`.
- **Client (edit):** `character_stats_panel.gd`, `character_panel_styles.gd`, `inventory_panel_styles.gd`, `inventory_panel.gd` (minimal, see D5).
- **Shared:** `shared/i18n/en.json` (additive strings only). No `rules/`, `protocol/`, or `golden/` change.
- **Tests:** new `client/tests/test_character_stats_header.gd`, `test_character_stat_groups.gd`, `test_paper_doll_backdrop.gd`, registered in `scripts/client_smoke.sh`; existing character-panel tests untouched.
- **Tools:** `tools/showme/screenshot_catalog.py` + `tools/test_regen_screenshots.py` for a UI-panel capture suite (see Open question 3). No bot-scenario additions (display-only; existing scenarios are the regression net). `docs/CODEMAP.md` rows for Classes/Client UI utilities updated for new files.
- **Docs:** `docs/as-built/v516_character-screen-redesign.md`.

## Focused verification (no per-slice `make ci`)

- `godot --headless --path client --script res://tests/test_character_stats_panel.gd` (existing, unchanged) plus the three new tests.
- `make client-unit` (character-panel and inventory tests inside).
- `make bot-client SCENARIO=...` for `client_character_stats_panel` and `client_inventory_paper_doll` headless; `.venv/bin/pytest tools/test_file_size_ratchet.py tools/test_extraction_coupling_ratchet.py tools/test_validate_codemap.py`; `make maintainability`; `python tools/validate_i18n.py`.
- `make regen-screenshots SUITE="<ui suite>"` for the visual evidence.
- Combined `make ci` is run by the coordinator after the batch integrates.

## Dependencies and integration risks

- **v514 (blocker for theme migration only).** Everything structural proceeds first. Ownership split to agree: v514 owns `UiTheme`, `ui_theme.v0.json`, and may touch `character_panel_styles.gd` / `inventory_panel_styles.gd` / `draggable_window.gd` to adopt tokens. If v514 already rewrites those style files, v516 must rebase onto them, not duplicate.
- **v515 (conflict).** Edits/extracts `inventory_panel.gd`, probably `inventory_panel_styles.gd` and tooltip files. v516 limits itself to the single paper-doll construction block and slot-style functions; paper-doll integration waits for v515 (see Plan P4). Expect the `paper` construction block (lines ~516–532 at base) to move; v516 re-applies its swap on the integrated file, never overwrites it.
- **v517 (low).** HUD XP bar/`character_bar.gd` are v517's; v516 builds its own in-panel XP bar and does not touch the HUD. Possible shared helper overlap if both write an "XP bar" widget — agree on one extraction (Open question 4).
- **Showme/catalog.** If v514/v515/v517 also add screenshot suites, `screenshot_catalog.py`, `test_regen_screenshots.py` and `docs/` capture lists conflict.
- **Debug-state contract** is the main regression risk: any change to label text composition breaks assertions. Mitigated by keeping name/value `Label` pairs as the source of the debug strings.
- **Godot import churn:** fresh worktree needs `--import`; revert rewritten `.glb.import` files before handoff (memory note).

## Open questions (affect planning)

1. **Paper-doll scope:** is a class-tinted 2D backdrop + connectors enough, or does the owner want the live 3D hero model (rejected here for perf/conflict)? Default: 2D now, 3D as a follow-up slice.
2. **Group membership data:** keep code-owned (default) or move to a schema-backed `shared/assets/character_stat_groups.v0.json`? Default: code-owned; ask only if the owner wants non-engineer editing.
3. **Screenshot suite ownership:** one shared `ui` suite (coordinated across v514–v517) vs. a per-slice suite. Default: v516 adds `character-screen` unless the coordinator designates a shared owner.
4. **Shared XP bar widget with v517:** reuse or duplicate? Default: duplicate minimal in-panel bar now; dedupe after both land.
