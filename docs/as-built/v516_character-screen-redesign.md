# v516 — Character screen redesign (as built)

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)

- **Base:** `425b9ae4`, with v514 (UiTheme) and v515 (inventory/tooltip extraction) transferred into this worktree. Presentation only; no protocol, server, rules, golden or bot-scenario change.
- **Spec / plan:** [spec](../specs/v516_spec-character-screen-redesign.md), [plan](../plans/v516_2026-10-01-character-screen-redesign.md)

## What shipped

- **Class-aware header** (`character_stats_header.gd`): class icon, name, class, level chip, XP bar (fraction from `experience / (experience + experience_to_next_level)`, full at max level) and an unspent-points badge shown only when points are available. Tinted from `class_presentations.v0.json`; neutral accent when no class.
- **Grouped derived stats** (`character_stat_groups.gd`, code-owned presentation order): Offense, Defense, Vitals, Utility as titled rows inside the existing scroll. A unit test asserts every `DERIVED_LABELS` key is in exactly one group.
- **Stats panel** (`character_stats_panel.gd`, 581 lines): header replaces the three text lines; enabled `+` buttons take the class accent; panel is taller (min 330x685) so more derived rows show. `get_debug_state()` keeps all keys and string formats and adds `header` and `derived_groups`.
- **Paper-doll** (`paper_doll_backdrop.gd`, `paper_doll_layout.gd`): class-tinted card, class badge in the empty top-left cell, connector lines; slot geometry reflowed to a non-overlapping 3-column grid (old layout overlapped Head/Amulet and Gloves/Boots). `inventory_panel.gd` wires the backdrop and layout (1,119 -> 1,111 lines; baseline lowered). Node name `character_paper_doll`, slot ids and `assert_paper_doll_layout` contract unchanged.
- **Theme:** all colors, type sizes and frames live in `shared/assets/ui_theme.v0.json` under `character_*` entries; `CharacterPanelStyles`/`InventoryPanelStyles` delegate to `UiTheme` and add accent-derived builders (darken/alpha factors are named constants, not tokens).
- **Showme:** `--focus character-screen` (+ `--variant points|nopoints|dual|nopoints-dual|paper-doll`), driver `client/scripts/showme/showme_character_screen_capture.gd` (paper-doll variant renders the real `InventoryPanel`), and a `character-screen` regen suite (17 captures: 5 classes x points/nopoints/paper-doll, plus dual-wield for barbarian and rogue).
- New strings under `character_screen.*` in `shared/i18n/en.json`.

## Evidence

- Captures (`docs/as-built/assets/v516/`): `panel-paladin-before.png` / `-after.png`, `panel-barbarian-dual-after.png`, `paperdoll-{rogue,ranger}-after.png`, `inventory-before.png` / `-after.png`. Full 17-image suite ran 17/17 ok (ignored, local: `.artifacts/screenshots/20261001-174717/`). Note: "before" panel capture was taken at 960x760 (content-scaled 0.5x) and "after" at 1920x1080, so sizes differ; layout, not scale, is the comparison.
- Focused checks on the final state: `make client-unit` PASS; `test_character_stat_groups` (12), `test_character_stats_header` (19), `test_paper_doll_backdrop` (21, includes layout no-overlap/fit), `test_character_stats_panel` (54, unchanged), `test_inventory_panel` (17), `test_stash_panel` (121), `test_ui_theme` (98); bot scenarios `character_stats_panel`, `client_inventory_paper_doll`, `inventory_equip_unequip`, `client_full_equipment`, `client_equipment_requirements_and_preview`, `account_stash_panel` all pass (headless); `make validate-shared` OK (2250 checks); `make maintainability` file-size and coupling ratchets pass; `pytest tools/test_regen_screenshots.py test_file_size_ratchet.py test_extraction_coupling_ratchet.py test_validate_codemap.py` 19 passed.
- Grep gate (no hex colors / numeric font sizes in character-screen files): two structural residues, `Color(1,1,1,EMBLEM_ALPHA)` (modulate) and a `Color(0,0,0,0)` "no accent" default in `card_style`.

## Limits / not done

- Live 3D paper-doll model rejected for this slice (perf and conflict risk); 2D backdrop only. Connector lines are faint because grid gaps are small.
- `stat_tooltip_label.gd` still hard-codes its colors/font size (v514 left it unmigrated).
- Equipped-item rendering inside the new slot grid was checked by bot scenarios and unit tests, not by a populated-slot capture.
- The MAIN/OFF dual-wield heading stays in the global column header, not the Offense group header.
- Windows persisted with the old `character_stats` layout key keep their saved position; the taller panel is clamped by `DraggableWindow`.
