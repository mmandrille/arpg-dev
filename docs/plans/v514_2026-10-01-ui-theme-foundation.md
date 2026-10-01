# v514 Plan — UI Theme Foundation

- **Status:** Implemented in worktree; awaiting coordinator integration
- **Date:** 2026-10-01
- **Spec:** [`v514_spec-ui-theme-foundation.md`](../specs/v514_spec-ui-theme-foundation.md)
- **Recorded batch base:** `425b9ae4`
- **Prerequisites:** none. v515-v517 consume this slice; they do not block it.

## Review gate and ownership

The spec passes review: client presentation only, no protocol, server, rules, golden, world preset, replay or determinism impact (`game/` untouched, so `lint-determinism` is not in scope). Shared data is a new schema-backed `shared/assets` catalog that must not carry executable logic (CLAUDE.md shared-rules invariant); frame recipes are declarative values, and the only logic (token resolution, tint math) lives in `ui_theme.gd`. Tuning-style numbers (tint factors) live in the catalog, satisfying the data-driven policy. Server authority is unaffected. Acceptance maps to tasks in the table at the end. No product decision blocks planning; the two open questions in the spec have recommended defaults (migrate only the two styles files; no Godot `Theme` resource). Task 8 is skipped unless the user answers Question 1 otherwise.

**Asset/plugin decision:** adopt in-repo `StyleBoxFlat`, the `RarityCueLoader` static-loader pattern and existing `showme` focuses; borrow current literals verbatim as catalog contents; reject external UI kits, fonts, skins, Godot Theme resources and plugins. No manifest or license entry needed.

**Security router:** fixed local JSON read, no network/auth/user input surface; out of scope.

**Maintainability:** `inventory_panel_styles.gd` (91) and `character_panel_styles.gd` (21) are under 600 and not grandfathered; they must not grow. New files stay well below 600 (`ui_theme.gd` target about 180 lines, validator about 120). `validate_shared.py` is grandfathered (3170 lines, baseline 3182): only a 2-3 line import + call is added, which fits the +25 allowance; put all logic in `tools/validate_ui_theme.py`. No `helpers=globals()` use.

## File map and shared-file risk

| Area | Path | Notes |
|------|------|-------|
| Catalog | new `shared/assets/ui_theme.v0.json`, `ui_theme.v0.schema.json` | **Cross-slice hot file.** v515/v516/v517 append entries; coordinator must diff merged result against each handoff. |
| Loader/API | new `client/scripts/ui_theme.gd` | **Cross-slice hot file** if siblings add helper functions (additive only). |
| Facades | `client/scripts/inventory_panel_styles.gd`, `client/scripts/character_panel_styles.gd` | v515 and v516 will edit these (likely conflicts). |
| Validator | new `tools/validate_ui_theme.py`, `tools/test_validate_ui_theme.py`; edit `tools/validate_shared.py` (lines near the `validate_rarity_cues(...)` call at ~3053 and the import block ~36/55) | `validate_shared.py` is touched by many slices; conflicts are small and mechanical. |
| Tests | new `client/tests/test_ui_theme.gd`; `scripts/client_smoke.sh` (add one `run_gate` line near the rarity cues gate ~line 273) | `client_smoke.sh` is a shared registry; every slice adds a line. |
| Docs | `docs/CODEMAP.md`, new `docs/as-built/v514_ui-theme-foundation.md`, `docs/as-built/assets/v514/*` | CODEMAP rows are shared with every slice. Coordinator owns `PROGRESS.md`/lifecycle. |

Not touched: `inventory_panel.gd`, `character_stats_panel.gd`, `shop_panel.gd`, `stash_panel.gd`, `market_panel.gd`, `skills_panel.gd`, `consumable_bar.gd`, `main.gd`, anything under `server/` or `shared/rules`, `shared/protocol`, `shared/golden`.

## Tasks

Tasks 1 and 2 need nothing from other slices and can start immediately; every task in this slice can run now. Only the final integration (merged `ui_theme.v0.json`) depends on siblings and belongs to the coordinator.

- [x] **1. Capture the baseline (before touching code).** At base `425b9ae4`, run `python3 skills/showme/scripts/render_focus.py --focus <f>` for `inventory`, `character-menu`, `corpse-inventory`, `rarity-cues`, `shop`, `blacksmith`, `skills`, `hud`; store PNGs under the scratchpad then `docs/as-built/assets/v514/<focus>-before.png`. Fresh worktrees need `godot --import` first (see memory: Godot import churn; revert rewritten `.glb.import` files afterwards). Also dump a **legacy style table**: a throwaway headless script (scratchpad, not committed) that calls every function in both styles files across their parameter combinations (hover, rarities x invalid, etc.) and prints `bg_color`, `border_color`, `border_width_*`, `corner_radius_*`, `content_margin_*` as JSON. Check: both outputs exist and the dump is deterministic across two runs.
- [x] **2. Schema + catalog.** Write `ui_theme.v0.schema.json` (strict, `additionalProperties: false`, color = `oneOf` hex6/hex8/4-float array, token-ref patterns, five required rarity keys) and `ui_theme.v0.json` populated from the legacy dump, including `state_modifiers` for the item-slot tint rules and a `fonts` section with the reserved roles. Colors needing exact alpha use float arrays. Check: `python3 tools/validate_shared.py` schema step passes for the new file.
- [x] **3. Semantic validator.** `tools/validate_ui_theme.py` (`validate_ui_theme(report, theme, item_templates)`): all frame/font color refs and spacing refs resolve; state overrides reference known colors; rarity keys equal `item_templates["rarities"]` and backgrounds are pairwise distinct; fonts `family == "default"`; sizes positive. Register in `validate_shared.py` next to `validate_rarity_cues`. `tools/test_validate_ui_theme.py` mirrors `test_validate_rarity_cues.py`: green catalog plus broken variants (unknown ref, missing rarity, bad hex, unknown state). Check: `.venv/bin/pytest tools/test_validate_ui_theme.py -v` and `python3 tools/validate_shared.py`.
- [x] **4. `UiTheme` loader and API** (`client/scripts/ui_theme.gd`). Implement the API listed in the spec: static `_loaded`/`_catalog`, `ensure_loaded()`, `invalidate()`, `color`, `spacing`, `frame`, `item_slot_frame`, `rarity_background`, `font_size`, `font_color`, `apply_font`, `has_token`. Unknown token returns magenta/0 plus `push_error` once per token (dedupe via a static set); `frame()` builds a fresh `StyleBoxFlat` from the parsed recipe (resolve colors at load into a cached dictionary to avoid per-call parsing). `item_slot_frame` reproduces the existing `lightened/darkened` logic using `state_modifiers`. Check: `godot --headless --path client --script res://tests/test_ui_theme.gd` after Task 5.
- [x] **5. Unit test** `client/tests/test_ui_theme.gd` (`extends SceneTree`, no scene tree use): loads catalog, asserts every frame builds with values derived from the catalog (not literals), all five rarities resolve, unknown rarity falls back to common, unknown token returns fallback without crash, `invalidate()` reloads, `frame()` returns independent instances (mutating one does not affect the next), corrupt-path handling via a temp-override hook if feasible, else documented as covered by the validator. Register in `scripts/client_smoke.sh` as `run_gate "GDScript ui theme test" "[gdtest] PASS: test_ui_theme" res://tests/test_ui_theme.gd`. Check: the test prints its PASS line.
- [x] **6. Migrate the styles files.** Rewrite `inventory_panel_styles.gd` (`panel_style`, `slot_style`, `item_slot_style`, `empty_slot_style`, `blocked_slot_style`, `paper_doll_style`, `ITEM_RARITY_BACKGROUNDS` if referenced elsewhere: grep first and keep a thin accessor if so) and `character_panel_styles.gd` (`panel_style`) as delegating one-liners with the same signatures. Re-run the legacy dump script against the new code and diff against Task 1's output: must be identical. Check: dump diff empty; `godot --headless` runs of `test_inventory_panel.gd`, `test_shop_panel.gd`, `test_rarity_cues.gd` pass (via their `client_smoke.sh` gates).
- [x] **7. Real-renderer comparison and as-built.** Re-run the Task 1 captures after the change, pixel-diff each pair with a small Python/PIL (or `ImageChops`) script (scratchpad), save `*-after.png` and the diff summary to `docs/as-built/assets/v514/`, record renderer/quality tier and inventory-rebuild timing before/after. Write `docs/as-built/v514_ui-theme-foundation.md` (API reference, extension rules for v515-v517, token inventory, limits, what was deliberately not migrated). Update `docs/CODEMAP.md`. Check: zero differing pixels, or each difference explained; `make maintainability` passes.
- [ ] **8. (Skipped by design, left open) (Optional, only if the user answers Open Question 1 "yes")** Point `shop_panel.gd` and `stash_panel.gd` private `_slot_style`/`_item_slot_style`/`_panel_style` at `UiTheme`, add the 0.93-alpha panel as a distinct catalog frame, shrink and lower the `stash_panel.gd` baseline in `.maintainability/file-size-baseline.tsv`. Re-run the shop/stash captures. Skipped by default to avoid collisions with v515.

## Final focused verification (batch worker gate; no per-slice `make ci`)

```bash
python3 tools/validate_shared.py
.venv/bin/pytest tools/test_validate_ui_theme.py tools/test_validate_rarity_cues.py -v
godot --headless --path client --script res://tests/test_ui_theme.gd      # plus the inventory, shop and rarity gates in scripts/client_smoke.sh
make maintainability
git diff --check
```

The coordinator runs the combined `make ci` only after every accepted slice is integrated. Bot scenarios are not required: no gameplay, protocol, world, inventory-logic, movement, combat or replay change; the UI real-renderer comparison (Task 7) is the visual gate and the spec records why no camera gameplay capture applies.

## Acceptance mapping

| Acceptance | Task / check |
|------------|--------------|
| 1 schema + validator | 2, 3; validate_shared, pytest |
| 2 headless load + fallback | 4, 5 |
| 3 property parity | 1 (dump), 6 (diff) |
| 4 no literals in styles files, no growth | 6; `grep` for `Color(`/`border_width`; `wc -l` |
| 5 screenshots | 1, 7 |
| 6 existing + new tests | 5, 6 |
| 7 ratchet | `make maintainability` |
| 8 docs | 7 |

## Handoff to the coordinator

Report: base commit and dirty state; complete changed/untracked list (expected: 2 new `shared/assets` files; `ui_theme.gd`; `validate_ui_theme.py` + test; `test_ui_theme.gd`; edits to the two styles files, `validate_shared.py`, `client_smoke.sh`, `CODEMAP.md`; spec, plan, as-built, `docs/as-built/assets/v514/`); focused commands and outcomes; pixel-diff result; shared files for merge attention: `shared/assets/ui_theme.v0.json`, `ui_theme.gd`, the two styles files, `validate_shared.py`, `client_smoke.sh`, `docs/CODEMAP.md`, `docs/progress/scenario-catalog.md` (not edited here).
