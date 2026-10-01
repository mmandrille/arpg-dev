# v514 — UI Theme Foundation

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)
- **Date:** 2026-10-01
- **Batch baseline:** `425b9ae4` (v510 refactor follow-ups)
- **Dependencies:** none. v515 (inventory/tooltip), v516 (character screen) and v517 (HUD/hotbar) depend on this slice and consume the API below.
- **Architecture:** ADR-0001 D2 (client is presentation only), ADR-0018 D2/D9 (in-repo assets only, real-renderer screenshot gate). ADR-0014 is not touched. CLAUDE.md GDScript rule 7 (`class_name` static singleton with `ensure_loaded()`).

## Purpose

Panel styling is hand-copied per file. `inventory_panel_styles.gd` (91 lines) and `character_panel_styles.gd` (21) each build `StyleBoxFlat` from literals. The same `#6b5420` frame border, `#0a0908` slot, `#3d2e10` hover, rarity slot backgrounds and 12/14 px margins are repeated again as private `_panel_style`, `_slot_style`, `_item_slot_style` and `_tooltip_style` functions in `shop_panel.gd`, `stash_panel.gd`, `skills_panel.gd`, `consumable_bar.gd` and elsewhere (about 32 client files mention `StyleBoxFlat`, plus ~100 `add_theme_*_override` calls). The next three slices redesign the inventory/tooltip, character screen and HUD. Without a shared token source each of them re-invents the palette and drifts.

v514 introduces one data-owned **`UiTheme`**: colors, frame recipes, font roles and spacing in `shared/assets/ui_theme.v0.json` (+ schema, validated by `tools/validate_shared.py`), loaded by a static `class_name UiTheme` singleton. The two existing styles files become thin facades over it. **No layout redesign and no visual change.**

## Scope and non-goals

In scope:

- New `shared/assets/ui_theme.v0.json` + `ui_theme.v0.schema.json`, and a focused cross-check validator `tools/validate_ui_theme.py` registered in `validate_shared.py` (schema is picked up by the existing `assets/` mapping; the validator adds semantic checks).
- New `client/scripts/ui_theme.gd` (`class_name UiTheme`, `extends RefCounted`, `static var _loaded`, `ensure_loaded()`, `invalidate()`), following the `RarityCueLoader` load pattern (`res://../shared/assets/...`, `push_error` on failure, never crash).
- Migrate `client/scripts/inventory_panel_styles.gd` and `client/scripts/character_panel_styles.gd` to build every style from `UiTheme`. Their public static function names and signatures stay, so `inventory_panel.gd`, `character_stats_panel.gd`, `tests/test_shop_panel.gd` and v515/v516 call sites are unchanged.
- Unit test for the loader and the builders; validator pytest.
- Before/after real-renderer captures proving no visual regression in the panels that consume the migrated styles.

Non-goals:

- No layout, size, anchor, font-size or copy change in any panel. No new art, fonts, textures, 9-patch skins, icons or plugins (in-repo `StyleBoxFlat` and Godot's default font only; ADR-0018 D2).
- No Godot `Theme` resource / `.tres` and no project-wide default theme (`ThemeDB`, `Control.theme`). A global theme would restyle every un-migrated control and violate "no visual regression". Revisit in a later slice.
- No change to protocol, server, rules, goldens, bot scenarios, persistence or replay.
- No migration of the private inline style functions in `shop_panel.gd`, `stash_panel.gd`, `market_panel.gd`, `skills_panel.gd`, `consumable_bar.gd`, bars or tooltips (see Open question 1). Rarity cue colors/shapes stay owned by `rarity_cues.v0.json` (v508).

## Decision: adopt / borrow / reject

Inspected: `inventory_panel_styles.gd`, `character_panel_styles.gd`, inline style helpers in shop/stash/skills/consumable_bar, `rarity_cue_loader.gd`, `shared/assets/*.v0.json` pattern, `client/project.godot` (no theme configured), `client/assets` (no UI fonts or skins).

- **Adopt:** `StyleBoxFlat` as the only frame primitive; the `RarityCueLoader` static-singleton loader; the `shared/assets/*.v0.json` + `.v0.schema.json` + `validate_*.py` convention; existing `showme` focuses for capture.
- **Borrow:** the exact current literal values as the initial catalog contents so output is bit-identical; `rarity_cues.v0.json` as the template for schema strictness (`additionalProperties: false`).
- **Reject:** external UI kits (KayKit UI, Kenney, fonts), Godot `Theme` resources/autoloads (rule 7 and the headless-test autoload problem), and per-panel forks of the palette.

## Catalog design (`ui_theme.v0.json`)

Top-level groups, all required, `additionalProperties: false` at the top and inside each token record:

| Group | Content | Example tokens (initial values = current literals) |
|-------|---------|----------------------------------------------------|
| `colors` | named color tokens. A color is `"#rrggbb"`, `"#rrggbbaa"` or `[r,g,b,a]` floats (floats keep existing alphas such as 0.92 exact). | `panel_bg`, `panel_bg_character`, `frame_border`, `slot_bg`, `slot_border`, `slot_hover_bg`, `slot_hover_border`, `empty_slot_bg`, `empty_slot_hover_bg`, `empty_slot_border`, `empty_slot_hover_border`, `blocked_slot_*`, `paper_doll_bg`, `paper_doll_border`, `text_primary` ... |
| `rarity_slot_backgrounds` | one color per gameplay rarity | `common #343432`, `magic #1b3458`, `rare #5a4520`, `unique #5a2f17`, `set #173f28` |
| `spacing` | integer px scale and named margins | `margin_panel`, `margin_panel_tight`, `margin_slot` |
| `fonts` | named roles: `{size, color (token ref), outline_size?, outline_color?}`. `family` is `"default"` only in v0 (no font assets). | `title`, `body`, `caption`, `value` — reserved for v515-v517; v514 migrates no font call site but defines the shape so the API is stable. |
| `frames` | named frame recipes: `{bg, border, border_width, radius, margin}` where `bg`/`border` are color-token refs, `border_width`/`radius` are an int or `[l,t,r,b]` / `[tl,tr,br,bl]`, `margin` is a spacing-token ref or `[l,t,r,b]`. Optional `states: {hover|disabled|invalid|selected: {bg?, border?}}` and optional `shadow {size, color, offset}`. | `panel`, `panel_character`, `slot`, `empty_slot`, `blocked_slot`, `paper_doll` |
| `state_modifiers` | the numeric tint rules applied by `item_slot_style`: `hover_lighten`, `invalid_darken`, `border_lighten`, `border_hover_lighten`, `invalid_border_*` | current 0.12 / 0.40 / 0.28 / 0.46 values, read from `item_slot_style` |

Rules:

- Unknown token refs fail validation and at load time resolve to a magenta fallback plus `push_error` (never silently to a plausible color).
- Per-state records override only `bg`/`border`; geometry is shared across states, which is true of every current slot/empty/blocked style.
- Rarity keys must equal `item_templates.v0.json` `rarities` (validator cross-check, same pattern as `validate_rarity_cues`).
- Frame/color/spacing/font names are `snake_case`. Sibling slices add their own entries under their own prefix (`inventory_*`, `character_*`, `hud_*`) with one entry per line so concurrent JSON edits merge cleanly.

## Public API (the contract v515-v517 build on)

All static, all call `ensure_loaded()` themselves, none return `null`:

```
UiTheme.ensure_loaded() / UiTheme.invalidate()          # tests + hot reload
UiTheme.color(token: String) -> Color
UiTheme.spacing(token: String) -> int
UiTheme.frame(name: String, state: String = "") -> StyleBoxFlat   # fresh instance each call
UiTheme.item_slot_frame(rarity: String, hover := false, dimmed := false, base := "slot") -> StyleBoxFlat
UiTheme.rarity_background(rarity: String) -> Color      # unknown rarity -> "common", as today
UiTheme.font_size(role: String) -> int
UiTheme.font_color(role: String) -> Color
UiTheme.apply_font(control: Control, role: String) -> void   # theme-override size/color/outline
UiTheme.has_token(kind: String, name: String) -> bool   # for tests and optional-extension probes
```

- `frame()` returns a new `StyleBoxFlat` so callers may mutate it (current behaviour); the catalog is never exposed mutably (`catalog()` returns a duplicate).
- v515/v516/v517 extend by **adding catalog entries and, if a recipe needs a new property, additive optional schema fields** (e.g. `corner_detail`, `border_blend`), never by hard-coding colors in panel code. A removed or renamed token is a breaking change and requires the coordinator.
- The facades (`InventoryPanelStyles`, `CharacterPanelStyles`) stay for call-site stability; v515/v516 may call `UiTheme` directly and delete a facade function when its last caller is gone.

## Observable acceptance criteria

1. `shared/assets/ui_theme.v0.json` validates against `ui_theme.v0.schema.json`; `python3 tools/validate_shared.py` passes including the new semantic checks (all token refs resolve; rarity keys equal `item_templates.v0.json`; every state override refers to defined colors; rarity backgrounds are pairwise distinct). A deliberately broken catalog (unknown ref, missing rarity, bad hex) fails in `tools/test_validate_ui_theme.py`.
2. `UiTheme` loads headlessly (no scene tree, no autoload). Missing/corrupt file logs an error, returns the documented fallbacks and does not crash.
3. Property parity: for every function in the two styles files, the migrated `StyleBoxFlat` equals the pre-migration one on `bg_color`, `border_color`, all four `border_width_*`, all four `corner_radius_*` and all four `content_margin_*` (verified by a one-time legacy dump comparison run during implementation; the committed unit test derives expectations from the loaded catalog instead of duplicating literals, per the Test Locking Policy).
4. `inventory_panel_styles.gd` and `character_panel_styles.gd` contain no color literal, border width, radius or margin literal; they only look up `UiTheme`. The two files stay at or below their current line counts.
5. Real-renderer before/after captures at the base commit and after the change are pixel-identical (or differences explained and < a documented threshold) for the panels that consume the migrated styles: `inventory`, `character-menu`, `corpse-inventory`, `rarity-cues` (UI fixtures), `shop`, `blacksmith`, plus `skills` and `hud` as untouched controls. Captures and the diff report are referenced from the as-built.
6. Existing `test_inventory_panel`, `test_shop_panel`, `test_character_stats_panel` (if present), `test_rarity_cues` and the new `test_ui_theme` pass; the new test is registered in `scripts/client_smoke.sh`.
7. `make maintainability` and `make lint-determinism` are unaffected (no `game/` change); no file grows past its baseline; no new `helpers=globals()` site.
8. `docs/CODEMAP.md` lists the new theme files; `docs/as-built/v514_ui-theme-foundation.md` documents the API and how siblings extend it.

## Likely surfaces

| Area | Paths |
|------|-------|
| Shared data | new `shared/assets/ui_theme.v0.json`, `shared/assets/ui_theme.v0.schema.json` |
| Validation | new `tools/validate_ui_theme.py`, `tools/test_validate_ui_theme.py`; 2-3 line registration in `tools/validate_shared.py` (grandfathered, currently 3170 lines vs baseline 3182) |
| Client | new `client/scripts/ui_theme.gd`; edited `client/scripts/inventory_panel_styles.gd`, `client/scripts/character_panel_styles.gd` |
| Tests | new `client/tests/test_ui_theme.gd`; registration in `scripts/client_smoke.sh` |
| Docs | `docs/CODEMAP.md` (Visual regression / showme row or a new "UI theme" row), `docs/as-built/v514_ui-theme-foundation.md`, captures under `docs/as-built/assets/v514/`. Coordinator owns `PROGRESS.md` and lifecycle index. |

## Verification (focused, no per-slice `make ci`)

- `python3 tools/validate_shared.py` and `.venv/bin/pytest tools/test_validate_ui_theme.py -v`
- `make client-unit` limited to the new and the affected tests (via the smoke script gates), or `godot --headless --path client --script res://tests/test_ui_theme.gd`
- Legacy-vs-migrated property dump comparison (implementation step, evidence in as-built)
- `python3 skills/showme/scripts/render_focus.py --focus <inventory|character-menu|rarity-cues|shop|blacksmith|corpse-inventory|skills|hud>` at the base commit and after; pixel diff
- `make maintainability`
- No bot scenario: no gameplay, protocol, world, inventory-logic or movement change. The existing client inventory scenarios are re-run only if the focused unit tests show a behavioral difference. Real-camera gameplay proof is not applicable (UI-only, nothing in the world view changes); the UI real-renderer captures are the visual gate.

## Integration risks

- **Sibling JSON conflicts.** v515-v517 all add entries to `ui_theme.v0.json`. Mitigation: prefix namespaces, one entry per line, coordinator diffs the merged file against each handoff (batch workflow step 3).
- **API drift.** If a sibling needs a function not listed, it must add it to `ui_theme.gd` additively; a rename is a coordinator decision.
- **Alpha rounding.** Hex alphas cannot represent 0.92/0.93 exactly; the float-array form is required for those colors or parity fails by one 8-bit step.
- **Load-time cost.** Styles are built many times per rebuild (slots). `ensure_loaded()` parses once; `frame()` must not re-parse or re-resolve tokens each call beyond dictionary lookups. Measure inventory rebuild time before/after in the as-built (expected unchanged).
- **Headless tests.** Loader reads via `ProjectSettings.globalize_path("res://")` like existing loaders; the test must not need a scene tree.

## Open questions affecting planning

1. **Scope of "similar styles files".** Only two files are literally named `*_styles.gd`. The near-identical private helpers in `shop_panel.gd`, `stash_panel.gd` (grandfathered 1194 lines, would shrink), `market_panel.gd`, `skills_panel.gd`, `consumable_bar.gd` are the larger duplication. Their panel backgrounds differ slightly (alpha 0.93 vs 0.92, tooltip variant), and v515 is likely to rewrite shop/stash/market slot rendering anyway. Recommendation: keep v514 to the two styles files and let v515/v516/v517 migrate the helpers inside the files they already redesign. The plan carries an optional Task 8 for the exact-duplicate slot helpers if the user wants them now.
2. **Godot `Theme` resource.** Recommendation: not now (non-goal above). Confirm.
