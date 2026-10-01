# v517 Spec — HUD / Hotbar Polish

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)
- **Date:** 2026-10-01
- **Codename:** `hud-hotbar-polish`
- **Batch base:** `425b9ae4` (detached v517 worktree, `/Users/mmandrille/git/arpg-dev-batch/v517-hud-hotbar`)
- **Depends on:** v514 `ui-theme-foundation` (UiTheme + `shared/assets/ui_theme.v0.json`, not yet integrated; see
  "Dependency split"). Soft coordination with v512 `boss-presence` (boss bar ownership), v515
  `inventory-tooltip`, v516 `character-screen` (shared theme/test registries only).
- **ADRs:** ADR-0018 (art direction, D9 screenshot gate), ADR-0001 D2 (client is renderer only), ADR-0007
  (presentation-only state).

## Purpose

The always-on HUD is functional but reads as prototype UI: a small framed panel with two 110 px
horizontal bars for health/mana, a row of ten flat `StyleBoxFlat` slots for the hotbar, a flat-bordered
minimap rectangle, and a boss bar built from default-ish `ColorRect`s. Polish these four surfaces into
one coherent look, using the v514 theme as the single source of colors/frames once it exists:

1. **Health and mana globes** — replace the two horizontal meters with two round, liquid-style globes at the
   bottom corners. Fill level = authoritative ratio. Numeric `cur / max` stays visible (never color-only).
2. **Hotbar** — consistent frame, slot, hover/pressed/empty/cooldown states, hotkey label, stack count, XP bar
   seating; the skill bar slot and character bar slot share the same frame language.
3. **Minimap frame** — frame, border and corner treatment around the existing minimap/full-screen map.
   The map drawing itself (`_draw_map` and children) is not restyled.
4. **Boss health bar frame** — panel frame, trough, fill, phase bar and portrait *frame* styling.

## Current state (inspected, base `425b9ae4`)

| Surface | Code | Notes |
|---|---|---|
| Health + mana | `client/scripts/player_health_bar.gd` (`PlayerHealthBar`, 241 lines) | `PanelContainer` at `(12, vp.y-92)`, identity label (name + Lv), two 110x9 `ColorRect` bars, `_flash()` tween on damage/heal/restore, attack-recovery strip. `get_debug_state()` is read by bot assertions. **There are no globes today.** |
| Hotbar | `client/scripts/consumable_bar.gd` (`ConsumableBar`, 444 lines) | 10 `ConsumableSlotButton`s of 52 px, panel centered at `(vp.x-580)/2, vp.y-78`, XP `ProgressBar` built in `_build_xp_bar()`. Slot/panel/XP styles are inline `StyleBoxFlat` factories. Drag/drop and `use_slot` are bot-facing. |
| Skill / character slot | `skill_bar.gd` (395, panel at `center+302`), `character_bar.gd` (180, panel at `center-366`) | Flank the hotbar; own inline panel styles. |
| Minimap | `client/scripts/discovery_minimap.gd` (393) | `PanelContainer` anchored top-right (`offset_left=-234`), `_panel_style()` frame, `_map` Control draws content. Modes: compact / fullscreen / hidden; opacity setting. |
| Boss bar | `client/scripts/boss_health_bar.gd` (403) | Panel 450 px at top (`PANEL_TOP=58`), portrait, 350x14 fill, 350x6 phase bar, reward panel; driven by `boss_visuals_controller.gd` `sync_boss_health_bar()`. `get_debug_state()` used by client bot scenarios. |
| Wiring | `client/scripts/main.gd` `_build_scene()` (~lines 3904–3980) | Instantiates all of the above under the `ui` `CanvasLayer`. 6645 lines now vs baseline 6632 (grandfathered). |
| Theme | none | No `UiTheme` / `ui_theme.v0.json` exists at the base; v514 is being specced in a sibling worktree. |

## Non-goals

- No protocol, schema, server, rules, golden, or persistence change. No new intent. Display only.
- No change to what the bars *mean*: HP/mana values, flash semantics, cooldown, hotbar assignment,
  drag/drop, hotkeys, and use rules are unchanged.
- No new layout for inventory, character screen, tooltips (v515/v516), status-effect bar, companion bar,
  quest/elite trackers, or the level label, beyond keeping them from overlapping the new globes.
- No boss banner, boss visuals, boss entry presence, or portrait *content* (v512).
- No restyling of the minimap's drawn content (grid, walls, POI/objective markers, player arrow).
- No imported art pack, plugin, or font. HUD is drawn in code (`StyleBox*`, `_draw`) from theme tokens.
- No shaders and no per-frame redraw loops.
- No HUD scale/settings UI, no customizable layout.

## Decisions

### D1 — Boss bar ownership with v512 (clear boundary)

| Owner | Owns |
|---|---|
| **v517** | Everything about the *look of the bar widget*: panel frame/background/border, bar trough and fill style, phase-bar trough/fill style, portrait frame, reward-panel frame, bar sizes/margins/position inside the panel, label fonts/colors. Implemented in a new `boss_bar_frame.gd` style helper plus the style section of `BossHealthBar._build()`. |
| **v512** | The boss *banner* (name/entry presentation), boss world visuals/tint/aura, arena presence, and portrait *content* (`_draw_portrait` drawing). v512 may add new nodes/files; it does not restyle the bar widget. |

Rules to avoid conflicts:
- v512 does not edit the style-construction code in `BossHealthBar._build()` or the constants
  (`BAR_SIZE`, `PANEL_WIDTH`, `PANEL_TOP`, ...). If v512 needs to hide/dim the bar during its banner, it
  calls the existing public API (`hide_live_boss`, `show_boss`) or adds one new public method; it does not
  reposition the panel.
- v517 does not touch `boss_visuals_controller.gd`, `boss_arena_presence.gd`, or the banner file.
- `get_debug_state()` keys of the bar are a stable contract for both; additions only.
- If the bar and v512's banner can be on screen together, the banner owns vertical space above
  `PANEL_TOP`; v517 keeps `PANEL_TOP` as a named constant so v512 can read it, not change it. The
  coordinator decides integration order (recommended: v512 first, v517 re-based on it).

### D2 — Globes are a self-contained `Control` drawn with `_draw`

A new `HudGlobe` control (`client/scripts/hud_globe.gd`, own `class_name`) draws a circular frame, a
ratio-clipped liquid fill with a simple surface line and highlight, and the `cur / max` label. It is
redrawn only when value, theme, or flash state changes (`queue_redraw()`), never from `_process`.
Flash on damage/heal/restore reuses the current tween behavior (colors and 0.05 s / 0.45 s timings come
from the existing code; they move to theme tokens only after v514).
`PlayerHealthBar` keeps its public API (`update_hp`, `update_mana`, `set_identity`,
`start_attack_recovery`, `get_debug_state`) and delegates rendering to two `HudGlobe`s; the identity
label and attack-recovery strip are kept (re-seated, not removed). **Rationale:** no change at the
call sites in `main.gd`, so main.gd growth is zero by construction.

HP color thresholds (green/amber/red) stay, but the globe must also communicate state without color
(numeric label, fill height, and a low-HP pulse-free edge treatment), consistent with v508's color-safe
direction.

### D3 — Layout is data computed by a pure helper

A new `hud_layout.gd` (`class_name HudLayout`, `RefCounted`, static functions) returns the rectangles for
globes, hotbar, skill slot, character slot, minimap, and boss bar for a given viewport size. Existing
`_position_panel`/`_sync_position` methods call it instead of local magic numbers
(`center-366`, `center+302`, `vp.y-78`, `(12, vp.y-92)`). This is the testable seam: a unit test asserts no
overlap and in-bounds rects at 1280x720, 1600x900, 1920x1080 and the project minimum window size.
Globes sit in the bottom corners, flanking the existing hotbar cluster (character slot `center-366`,
hotbar `±290`, skill slot `center+302..+366`); on narrow viewports the globes shrink to a minimum
diameter rather than overlapping the cluster.

### D4 — Style goes through one adapter that v514 later replaces

Until v514 lands, a single file `client/scripts/hud_style.gd` (`class_name HudStyle`, static) is the only
place HUD colors, border widths, radii, and sizes are defined (initially the literals currently scattered
across the five scripts, so default output is visually near-identical). After v514 integrates, `HudStyle`
reads tokens from `UiTheme` and the literals are deleted. HUD scripts never reference raw colors. HUD
colors/sizes are presentation, not gameplay tuning, so the CLAUDE.md data-driven policy is satisfied by
routing them through UiTheme/`ui_theme.v0.json` once available; the interim literals are called out in the
plan as temporary.

### D5 — Adopt / borrow / reject (ADR-0018)

- **Adopt:** the in-repo code-drawn UI approach (`StyleBoxFlat`, `_draw`) and the v514 UiTheme tokens.
- **Borrow:** the flash-tween and ratio logic already in `PlayerHealthBar`; the `capture_frame` bot step and
  `make regen-screenshots` scenes workflow for before/after evidence.
- **Reject:** importing a UI art kit or fonts (ADR-0018 scopes KayKit to 3D; D2 allow-list permits free art
  but it is not needed to meet the criteria and adds license/asset-manifest churn); custom canvas shaders
  for liquid (frame-pacing risk, v497 baseline); a second parallel theme system.

## Dependency split (v514)

- **Can proceed before v514 integrates (no UiTheme symbol used):** `hud_layout.gd` + test; `HudGlobe` drawing
  and test; `PlayerHealthBar` re-seat onto globes with unchanged debug state; `hud_style.gd` adapter with
  current literals; hotbar/skill/character/minimap/boss frame refactors onto `HudStyle` and
  `boss_bar_frame.gd` (still literal values); showme/capture fixtures ("before" captures);
  frame-pacing baseline measurement; bot/client-unit registration.
- **Blocked until v514 is integrated in this worktree:** switching `HudStyle` to read `UiTheme` tokens,
  adding HUD token groups to `shared/assets/ui_theme.v0.json` (+ schema, `validate_shared`) and removing
  literals, theme-driven globe/frame palettes, and final "after" captures that must reflect the shared
  theme. These are explicitly marked `BLOCKED(v514)` in the plan.

## Observable acceptance criteria

1. **Globes.** In the real renderer, a left (health) and right (mana) globe appear at the bottom corners.
   Fill height tracks the ratio; at 0%, 50%, 100% it is visibly empty, half, full. `cur / max` text is
   always readable. Damage, heal, and mana-restore flashes still occur. The old horizontal meters no longer
   render. Name/level and the attack-recovery indicator remain visible.
2. **No regression of values.** `PlayerHealthBar.get_debug_state()` keys and values are unchanged for the
   same inputs (existing bot assertions pass unmodified).
3. **Hotbar.** Ten slots show distinct empty / filled / hover / pressed / drag-over states, a visible hotkey
   label, stack count, and the XP bar seated under the frame; cooldown/disabled presentation of the skill
   slot remains legible. Assignment, drag/drop, hotkeys, and `use_slot` behave exactly as before
   (`use_potion_hotbar` and hotbar assertions pass unmodified).
4. **Minimap.** Compact, full-screen, and hidden modes and the `map_opacity` setting work as before; frame
   and border are restyled; map content rendering is pixel-identical apart from the frame.
5. **Boss bar.** The bar frame, trough, fill, phase bar, portrait frame, and reward panel use the new style;
   HP, phase, and reward displays and `BossHealthBar.get_debug_state()` are unchanged; `boss_health_bar`
   client scenarios pass unmodified. v512 banner work is not required for this to hold.
6. **Layout.** `HudLayout` unit test proves no overlap/in-bounds at the listed viewport sizes; no existing
   HUD element (level label, status-effect bar, companion bar, trackers) is covered by a globe at 1280x720
   and 1920x1080 in a captured frame.
7. **Zero main.gd growth.** `client/scripts/main.gd` line count is <= its value at the base (6645) and does
   not exceed `.maintainability/file-size-baseline.tsv`+25; preferably net-negative. All new files are
   <= 600 lines; no new `helpers=globals()` sites.
8. **Frame pacing (v497).** Paired before/after runs of `dungeon_frame_pacing_probe` (same host, vsync on,
   same fixture) show no statistically meaningful regression in steady-state frame time and no new
   per-frame HUD work; the HUD draws only on change. Result and measurement limits recorded in as-built.
9. **Captures.** Before/after real-renderer captures via the showme/`regen-screenshots` workflow (a new
   `hud` focus/suite or fixtures in the `scenes` suite) cover: globes at full / half / low / empty, hotbar
   empty and populated, minimap compact + full-screen, boss bar at 100% / 50% / 10% with a phase bar and
   the reward panel, at 1280x720 and 1920x1080, in both Balanced and Performance quality tiers. The as-built
   attaches them.
10. **Verification gates.** `make client-unit` and the affected client bot scenarios stay green (see plan).
    `make validate-shared` stays green after any `ui_theme` token addition (post-v514).

## Likely surfaces and ownership

| Surface | Paths | Notes |
|---|---|---|
| New client | `client/scripts/hud_globe.gd`, `hud_layout.gd`, `hud_style.gd`, `boss_bar_frame.gd` | Each independently importable/testable; no `main.gd` import (extraction independence). |
| Edited client | `player_health_bar.gd`, `consumable_bar.gd`, `skill_bar.gd`, `character_bar.gd`, `discovery_minimap.gd` (frame only), `boss_health_bar.gd` (style only) | Constants/style methods only; behavior untouched. |
| `main.gd` | Expected: no edit. If a call-site tweak is unavoidable it must be net-zero lines. | Grandfathered; extract don't grow. |
| Shared (post-v514) | `shared/assets/ui_theme.v0.json` (+ schema), `tools/validate_shared.py` | **Shared with v514/v515/v516** — additive token groups only. Blocked until v514. |
| Tests | `client/tests/test_hud_globe.gd`, `test_hud_layout.gd`, extend `test_boss_health_bar.gd`, `test_skill_bar.gd`, `test_character_bar.gd`, `test_discovery_minimap.gd`; register in `client/scripts/client_smoke.sh` (**shared with all siblings**) | |
| Bot / capture | New client scenario `tools/bot/scenarios/client/<next>_hud_polish_visual.json` using `capture_frame`; scenario catalog + `docs/progress/scenario-movement-audit.tsv` if the audit requires; showme `tools/showme/screenshot_catalog.py` | |
| Docs | `docs/CODEMAP.md` (HUD row, new files), plan/as-built | CODEMAP is **shared with all siblings**. |

No file under `server/`, `shared/protocol/`, `shared/rules/`, or `shared/golden/` changes.

## Focused verification (summary; commands in the plan)

- `make client-unit` (all new/extended HUD tests registered in `client_smoke.sh`).
- Client bot: `use_potion_hotbar`, `client_boss_health_bar_ui`, boss phase/portrait/reward scenarios,
  minimap/map-overlay/opacity scenarios, `elite_objective_hud` + `elite_objective_minimap_pin`, new
  `hud_polish_visual`.
- `pytest tools/test_scenario_movement_audit.py`, `pytest tools/test_regen_screenshots.py` if the showme
  catalog changes; `make maintainability`.
- `make regen-screenshots SUITE=<hud suite>` before/after; `dungeon_frame_pacing_probe` paired runs.
- No per-slice `make ci`; the coordinator runs the combined gate.

## Integration risks

- **v512 / `boss_health_bar.gd`:** same file. Mitigated by D1 ownership and by moving frame styling into
  `boss_bar_frame.gd` so v517's edit to the shared file is a handful of call lines.
- **v514 / theme + `ui_theme.v0.json`:** v517 consumes and extends it; token names must be agreed with v514
  (open question 3). Until then literals live only in `hud_style.gd`.
- **v515 / v516:** may restyle the inventory panel/tooltips and character panel with the same theme and add
  `client_smoke.sh` / CODEMAP rows; `character_bar.gd` (the HUD slot that opens the character panel) is
  owned by v517 for its frame only — v516 must not restyle it.
- **Layout collisions** with status-effect bar, companion bar, trackers, level label at bottom/top corners
  (verified by captures and the layout test).
- **Frame pacing:** globe redraw cost and added `Control` count; mitigated by redraw-on-change and measured.
- **Bot assertions** that read HUD debug state or node paths could break on re-parenting; mitigated by
  keeping public API/debug keys and running the listed scenarios.

## Open questions affecting planning

1. **Globes vs. enhanced bars.** The prompt says "health and mana globes" but none exist. Default in this
   spec: replace the meters with bottom-corner globes (D2). Confirm the owner wants globes (a layout and
   silhouette change) rather than just polished bars.
2. **Quality tiers.** Should globes have a simplified variant in the Performance tier (no highlight/surface
   line)? Default: same geometry in both tiers; measure first.
3. **Token names with v514.** Need v514's final token schema (color roles, frame/border/radius groups) before
   the blocked tasks; v517 will add `hud.*` groups only if v514 allows per-surface groups.
4. **Integration order for the boss bar** with v512 (recommended v512 first).
