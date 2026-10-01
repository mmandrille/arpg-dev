# v515 Plan — Inventory, Slot, Rarity Border, and Tooltip Redesign

- **Spec:** [`v515_spec-inventory-tooltip-redesign.md`](../specs/v515_spec-inventory-tooltip-redesign.md)
- **Baseline:** `425b9ae4` (detached worktree, no branch)
- **Prerequisite:** v514 `ui-theme-foundation` (UiTheme + `shared/assets/ui_theme.v0.json`). Not available yet.
- **Final gate:** focused slice verification only. The coordinator runs the combined `make ci` after every slice is integrated; no per-slice `make ci`, `make ci-full`, `/finish`, commit, or push here.

## Spec review gate (recorded)

| Check | Result |
|-------|--------|
| Scope/non-goals | Display-only; no protocol/server/rules/golden/replay. Matches brief. |
| AC -> check mapping | See "Acceptance map" below. |
| Determinism / replay / world presets | N/A (client only). |
| Server authority | Unchanged; intents and drag payloads untouched; proven by existing client scenarios. |
| Shared data ownership | Theme data owned by v514; v515 adds keys only by coordination. |
| Adopt/borrow/reject | Recorded in spec (in-repo Godot StyleBox, ItemIconDrawer, RarityCuePresenter, showme harness; reject external plugins). |
| Ratchet | `inventory_panel.gd` 1,626 (baseline 1,623); extraction required, baseline lowered. |
| Sibling overlap | v514 (theme), v516 (paper doll + styles), v517 (TBD). See "Shared-file conflicts". |
| Spec correction | Worktree file is 1,626 lines vs baseline 1,623 (+3, within +25 allowance); spec uses the real count. |

Open questions in the spec (border weight, other-panel scope, size target, rarity-visual ownership) do not block Phase A.

## Task order and blocking

**Phase A (can proceed now, no v514):** T1-T4 (characterization, extraction, tooltip hierarchy against current constants, baseline). **Phase B (BLOCKED until v514 is integrated into this worktree):** T5-T6 (UiTheme adoption, rarity visual lookup). **Phase C:** T7-T9 (captures, as-built, handoff; final captures wait for Phase B).

## File map

| File | Action |
|------|--------|
| `client/scripts/inventory_tooltip_content.gd` (+ `.uid` if Godot generates) | New: pure content builder, `InventoryTooltipContext` typed input |
| `client/scripts/inventory_item_pricing.gd` | New, only if needed to reach the size target (T2b) |
| `client/scripts/inventory_panel.gd` | Remove moved functions, build context, delegate; no behavior change |
| `client/scripts/inventory_panel_styles.gd` | Slot/border redesign API; single rarity lookup |
| `client/scripts/item_tooltip_panel.gd`, `item_tooltip_stat_sections.gd` | Hierarchy/layout redesign |
| `client/tests/test_inventory_tooltip_content.gd` | New; register in `scripts/client_smoke.sh` |
| `client/tests/test_inventory_panel.gd`, `test_rarity_cues.gd`, `test_look_and_feel_polish.gd` | Update for new visuals/API |
| `.maintainability/file-size-baseline.tsv` | Lower `inventory_panel.gd` entry to real size |
| `docs/as-built/v515_inventory-tooltip-redesign.md`, `docs/CODEMAP.md` (Loot presentation / Client UI utilities rows) | New / add new modules |
| Screenshot suite registration (showme / regen-screenshots) | Only if no inventory/tooltip suite exists |

## Tasks

### Phase A - DONE (executed 2026-10-01; see handoff for results)

- [x] **T1 Characterization tests (before moving code).** Add `client/tests/test_inventory_tooltip_content.gd` skeleton that, against the *current* `InventoryPanel` instance, snapshots `_tooltip`/`_tooltip_lines`/`_requirement_lines`/`_comparison_entries` for: common weapon, magic armor, rare with affixes, unique with effects, set piece, potion, unmet-requirement item, comparison-delta item, hotbar-assigned item. Assert semantic content (rarity word, level, requirement met/unmet, delta sign, set active/inactive), not pixel/tuning values. Check: `godot --headless --path client --script res://tests/test_inventory_panel.gd` and the new test run green on baseline.
- [x] **T2 Extract tooltip content.** Create `InventoryTooltipContent` + `InventoryTooltipContext` per spec D1; move the listed functions; panel builds the context from its fields and delegates (`_make_item_tooltip`, `_tooltip`, `_fill_slot`, `_item_shows_requirement_warning` callers). No `globals()`/panel reference in the new file. Re-point T1 test to preload the module directly (no `inventory_panel.gd` import). Check: `make client-unit` subset (inventory, shop, stash, look_and_feel tests) and `grep -n "helpers=globals" ` shows no new site; `make maintainability`.
- [x] **T2b Extract pricing if needed.** If `inventory_panel.gd` is still above ~1,100 after T2, move `_item_gold_value/_item_buy_price/_generated_buy_price/_town_vendor_*` into `inventory_item_pricing.gd` with its own direct test. Check: same as T2.
- [x] **T3 Lower ratchet baseline.** Set `inventory_panel.gd` baseline to its real post-extraction size in `.maintainability/file-size-baseline.tsv`. Check: `make maintainability`.
- [x] **T4 Tooltip hierarchy.** Rework `ItemTooltipPanel`/line model per spec Scope 3 using current color constants: header (name + rarity word + item level), type/weapon meta, base stats, rolled stats, set/unique, requirements, comparison, hotbar, price footer; fold `Rarity:` line into header, keep rarity word in text. Update debug accessors and any tests/bot scenarios reading them (AC4). Check: new/updated tests, `test_shop_tooltip_stability.gd`, `test_shop_panel.gd`, `test_stash_panel.gd`; showme `inventory` and `shop` focuses for overflow (AC5).
- [x] **T4b Slot redesign against current constants.** Update `InventoryPanelStyles` slot/border/hover/blocked/invalid styles and slot draw order (icon, v508 cue, hotbar badge, warning). Keep API stable for existing callers. Check: `test_inventory_panel.gd`, `test_rarity_cues.gd`, showme `inventory` focus at actual slot sizes.
- [x] **T4c Bot flow preservation.** `make bot-client SCENARIO=02_inventory_open_close HEADLESS=1`, `03_inventory_equip_unequip`, `04_inventory_lab_drop_item`, `13_inventory_paper_doll`, `v508_rarity_cue_preset_camera`; plus one stash, one vendor, one market client scenario (names from `docs/progress/scenario-catalog.md`). Confirm no new scenario setup via incidental navigation (movement-contract rule: no new bot scenarios are planned; any added must follow it).

### Phase B - BLOCKED until v514 integrated into this worktree

- [x] **T5 (blocked: v514) Adopt UiTheme.** Replace hard-coded colors/borders/spacing/font sizes in `inventory_panel_styles.gd`, `item_tooltip_panel.gd`, and the extracted module with `UiTheme` lookups. Consolidate `ITEM_RARITY_BACKGROUNDS`, `_rarity_color`, and `LOOT_LABEL_RARITY_COLORS` rarity usage in UI into one theme-backed rarity lookup. If new theme keys are required, add them to `shared/assets/ui_theme.v0.json` + schema in coordination with v514 (list the contributors for the coordinator's shared-JSON diff). Re-run `make validate-shared`; headless tests must not hit autoload-resolution errors (use the `class_name` static singleton + `ensure_loaded()` pattern; `UiTheme` must satisfy it).
- [x] **T6 (blocked: v514) Contrast review.** Measure border-vs-background and cue contrast for all five rarities on theme colors (>= 3:1 border, per spec AC2); adjust only with recorded ratios.

### Phase C

- [x] **T7 Real-renderer captures.** Before (from baseline, taken in T1 time) and after captures with showme `inventory`, `shop`, `corpse-inventory`; fixture set per AC3 (5 rarities, hover, invalid, blocked, hotbar, comparison, set, unique), grayscale pass, recorded renderer/quality tier; register a regen-screenshots suite if none covers inventory/tooltips (`make regen-screenshots-list` first). Final "after" captures wait for T5/T6; interim captures after T4/T4b are recorded as such.
- [x] **T8 As-built + CODEMAP.** Write `docs/as-built/v515_inventory-tooltip-redesign.md` (proof, limits, measured contrast, baseline before/after, any unreadable cases) and add the new modules to `docs/CODEMAP.md`. Do not edit `PROGRESS.md` or the lifecycle index (coordinator).
- [x] **T9 Handoff to coordinator.** Report: base commit, dependencies used (v514 yes/no), complete changed/deleted/untracked list, ignored raw evidence (screenshots under `.artifacts/`), commands and outcomes, scenarios run, unmet criteria, conflict list.

## Acceptance map

| AC | Check |
|----|-------|
| 1 slot states | T4b tests + showme captures (T7) |
| 2 rarity distinguishability/contrast | grayscale captures + T6 measurements |
| 3 tooltip hierarchy, no info lost | T1 characterization (carried over after T2) + T4 tests + captures |
| 4 accessors/tests | T4 test updates |
| 5 other panels | shop/stash tooltip tests + shop/corpse-inventory captures |
| 6 extraction/ratchet | direct-import test, `make maintainability`, baseline diff |
| 7 no server/protocol change | `git diff --stat` shows client/docs/tests only (+ ui_theme keys); T4c scenarios |
| 8 captures | T7 |

## Final focused verification (no `make ci`)

```
make client-unit
make maintainability
make validate-shared                      # only if ui_theme keys were added
make bot-client SCENARIO=<each scenario in T4c> HEADLESS=1
python3 skills/showme/scripts/render_focus.py --focus inventory   # and shop, corpse-inventory
make regen-screenshots SUITE="<registered inventory/tooltip suite>"
git diff --check
```

Coordinator note: combined `make ci` runs after all accepted slices are integrated.

## Shared-file conflicts and integration notes

- **v514:** `shared/assets/ui_theme.v0.json` + schema, `UiTheme` loader, `client_constants.gd` (rarity colors), possibly `inventory_panel_styles.gd`, `item_tooltip_panel.gd`, `docs/CODEMAP.md`, `tools/validate_shared.py`. Any `shared/**/*.v0.json` touched by both must be diffed against each handoff.
- **v516:** `inventory_panel_styles.gd`, paper-doll block and slot creation in `inventory_panel.gd`, `test_inventory_panel.gd`, `13_inventory_paper_doll` scenario, baseline entry for `inventory_panel.gd` (both lower it; coordinator reconciles to the real final size). Suggest integrating v515's extraction before v516's paper-doll edits.
- **v517:** unknown; coordinator to confirm.
- Also touched: `docs/CODEMAP.md`, `.maintainability/file-size-baseline.tsv`, `scripts/client_smoke.sh`.
