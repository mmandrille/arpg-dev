# v517 Plan — HUD / Hotbar Polish

- **Spec:** [`v517_spec-hud-hotbar-polish.md`](../specs/v517_spec-hud-hotbar-polish.md)
- **Date:** 2026-10-01
- **Baseline commit:** `425b9ae4` (detached worktree `/Users/mmandrille/git/arpg-dev-batch/v517-hud-hotbar`)
- **Prerequisite slices:** v514 `ui-theme-foundation` (UiTheme + `shared/assets/ui_theme.v0.json`). Not yet
  available. Coordination: v512 `boss-presence` (boss bar ownership), v515, v516 (shared theme/registries).
- **Final gate:** focused slice verification (Phase 5). The coordinator runs the combined `make ci` after all
  accepted slices are integrated; there is intentionally no per-slice `make ci` / `make ci-full` here.

## Review gate result

Checked against the spec: scope/non-goals are display-only; no protocol, schema, golden, server, determinism,
replay, or world-preset impact (no `game/` files touched, so `make lint-determinism` is not needed). Server
authority is untouched: the HUD only renders values already delivered. Shared data: the only shared file is
`shared/assets/ui_theme.v0.json` (owned by v514, additive tokens only, blocked). Asset adopt/borrow/reject is
recorded in spec D5 (no imported art). Maintainability: `main.gd` (6645 lines, baseline 6632) must not grow;
other touched files are far below 600. Corrections recorded: the brief's "health and mana globes" do not
exist at base; the plan treats them as a new component, flagged as open question 1 in the spec. Overlap with
siblings is listed under "Shared-file conflicts". No material gap that blocks starting Phase 1–4.

## What can proceed before v514 integrates

Phases 0–4 and 5 (partially). Everything that does not read `UiTheme` or edit `ui_theme.v0.json`:
layout helper, globe control, `PlayerHealthBar` re-seat, `hud_style.gd` adapter with current literals,
hotbar/skill/character/minimap/boss frame refactors, tests, showme fixtures, "before" captures, frame-pacing
baseline. Tasks tagged **BLOCKED(v514)** wait until the exact integrated v514 is transferred into this worktree
(without overwriting spec/plan); then run its focused dependency checks (`validate-shared`, v514 client tests).

## File map and ownership

| File | Action | Owner / note |
|---|---|---|
| `client/scripts/hud_layout.gd` | new (`HudLayout`, static rect math) | v517 |
| `client/scripts/hud_style.gd` | new (`HudStyle`, single color/size adapter) | v517; reads UiTheme after v514 |
| `client/scripts/hud_globe.gd` | new (`HudGlobe`, `_draw` liquid globe) | v517 |
| `client/scripts/boss_bar_frame.gd` | new (style builders for boss bar panel/trough/fill/phase/portrait frame/reward) | v517 (D1) |
| `client/scripts/player_health_bar.gd` | edit: delegate to two `HudGlobe`s, keep API/debug state | v517 |
| `client/scripts/consumable_bar.gd`, `skill_bar.gd`, `character_bar.gd` | edit: styles + `HudLayout` positions only | v517 (`character_bar.gd` frame only; v516 does not restyle it) |
| `client/scripts/discovery_minimap.gd` | edit: `_panel_style()` + frame only; `_draw_map*` untouched | v517 |
| `client/scripts/boss_health_bar.gd` | edit: replace inline style code with `BossBarFrame` calls; keep `get_debug_state`, public API | v517 owns style section; v512 must not edit it |
| `client/scripts/main.gd` | **no edit expected**; if forced, net zero lines | grandfathered; extract don't grow |
| `client/tests/test_hud_layout.gd`, `test_hud_globe.gd` | new | v517 |
| `client/tests/test_boss_health_bar.gd`, `test_skill_bar.gd`, `test_character_bar.gd`, `test_discovery_minimap.gd` | extend (style/layout asserts; none removed) | v517 |
| `scripts/client_smoke.sh` | register new tests | **shared with v512/v514/v515/v516** |
| `tools/bot/scenarios/client/<next free>_hud_polish_visual.json` | new capture scenario | v517 |
| `tools/showme/screenshot_catalog.py` (+ showme scene entry) | add `hud` suite/focus if cheap | v517 |
| `docs/CODEMAP.md` | HUD/Discovery minimap/Boss rows | **shared with all siblings** |
| `docs/progress/scenario-catalog.md`, `scenario-movement-audit.tsv` | only if the new scenario requires | shared registries |
| `shared/assets/ui_theme.v0.json` (+ schema), `tools/validate_shared.py` | additive HUD tokens, **BLOCKED(v514)** | **owned by v514; shared with v515/v516** |
| `docs/as-built/v517_hud-hotbar-polish.md` | new | v517 |

## Tasks

### Phase 0 — Baseline evidence (no code change)

- [x] 0.1 Record base captures ("before"): run `make regen-screenshots` for the scenes suite subset if useful
  and a manual `make bot-client SCENARIO=use_potion_hotbar HEADLESS=0` frame; save under
  `.artifacts/v517/before/` (ignored; list in handoff). Check: files exist, non-blank.
- [x] 0.2 Frame-pacing baseline: `make bot-client SCENARIO=dungeon_frame_pacing_probe HEADLESS=0` x3 (balanced),
  record steady-state frame time p50/p95 and process_ms into `.artifacts/v517/pacing-before/`. Check: report
  passes its own counters.
- [x] 0.3 Read how existing bot assertions consume HUD debug state: `grep -rn "get_debug_state\|health_bar\|hotbar"
  client/scripts/bot_*.gd` and list the keys that must stay stable. Check: list written into as-built.
- [x] 0.4 Record `wc -l client/scripts/main.gd` (expect 6645) and the baseline row. Check: matches.

### Phase 1 — Layout helper (unblocked)

- [x] 1.1 `hud_layout.gd`: static functions returning `Rect2` for health globe, mana globe, character slot,
  hotbar, skill slot, minimap compact, boss bar panel (keep `PANEL_TOP` named), for a viewport size; globe
  diameter scales between a min and a max and never intrudes into the character/hotbar/skill cluster.
- [x] 1.2 `client/tests/test_hud_layout.gd`: no overlap among HUD rects and all in-bounds at 1280x720,
  1600x900, 1920x1080, and the project minimum window size (read from `project.godot`); also vs. a stub rect
  for status-effect/companion bars after reading their real positions. Check:
  `godot --headless --path client -s res://tests/test_hud_layout.gd` (or the `client-unit` runner).
- [x] 1.3 Register in `client_smoke.sh`. Check: `make client-unit` lists PASS for it.

### Phase 2 — Style adapter and globe (unblocked)

- [x] 2.1 `hud_style.gd`: move the literal colors/borders/radii/sizes now inline in `player_health_bar.gd`,
  `consumable_bar.gd`, `skill_bar.gd`, `character_bar.gd`, `discovery_minimap.gd`, `boss_health_bar.gd` into
  named getters. Mark in a header comment: "temporary literals; replaced by UiTheme tokens in v517 task 6.x".
  Default values reproduce today's look.
- [x] 2.2 `hud_globe.gd`: Control with `set_ratio`, `set_label`, `set_fill_color`, flash API; draws frame ring,
  ratio-clipped liquid fill (polygon/arc clip, no shader), surface line, highlight; `queue_redraw()` only on
  change; `get_debug_state()`. No `_process`. Check: `test_hud_globe.gd` — ratio 0/0.5/1 produce expected
  fill extent, ratio clamps for `max<=0`, redraw not scheduled when values are unchanged, no
  `_process`/timer.
- [x] 2.3 Register `test_hud_globe.gd`. Check: `make client-unit`.

### Phase 3 — Re-seat player health/mana (unblocked)

- [x] 3.1 `player_health_bar.gd`: build two `HudGlobe`s positioned by `HudLayout` (`_sync_position` on
  viewport resize), keep `update_hp`, `update_mana`, `set_identity`, `start_attack_recovery`,
  `get_debug_state` with unchanged keys/values, keep flash semantics (reuse tween, colors from `HudStyle`),
  keep identity label and attack-recovery strip, re-seated near the globes. Remove old `ColorRect` meters.
- [x] 3.2 Extend/locate the existing player-health-bar test (search `client/tests` for `PlayerHealthBar`; add
  `test_hud_globe`-style asserts there or in `test_hud_globe.gd`): same input -> same debug state as at base
  (compare against values captured in 0.3). Check: `make client-unit`.
- [x] 3.3 Bot: `make bot-client SCENARIO=use_potion_hotbar HEADLESS=1` (heal/mana via potion) passes
  unmodified.

### Phase 4 — Hotbar, skill/character slot, minimap frame, boss bar frame (unblocked, literal style)

- [x] 4.1 `consumable_bar.gd`: slot/panel/XP styles from `HudStyle`; add distinct empty/filled/hover/pressed/
  drag-over visuals, hotkey label and stack count legibility; position via `HudLayout`. Behavior (drag/drop,
  `use_slot`, `set_hotbar_state`) untouched. Check: existing tests plus a new style assertion test; scenarios
  `use_potion_hotbar`, `client_skill_points_and_magic_bolt`.
- [x] 4.2 `skill_bar.gd` and `character_bar.gd`: share the frame/slot language and `HudLayout` positions;
  `test_skill_bar.gd`, `test_character_bar.gd` pass unmodified plus layout asserts.
- [x] 4.3 `discovery_minimap.gd`: restyle `_panel_style()` / border only; assert map content draw path is
  unchanged (no edits under `_draw_*`). Check: `test_discovery_minimap.gd`; scenarios `discovery_minimap_toggle`,
  `full_screen_map_overlay`, `map_transparency_setting`, `minimap_points_of_interest`,
  `elite_objective_minimap_pin`.
- [x] 4.4 `boss_bar_frame.gd` + `boss_health_bar.gd`: move panel/trough/fill/phase/portrait-frame/reward styles
  into the helper; keep `PANEL_*`/`BAR_SIZE` constants named and the debug-state keys. The edit to
  `boss_health_bar.gd` is limited to the style section (D1) to ease the merge with v512. Check:
  `test_boss_health_bar.gd`; scenarios `client_boss_health_bar_ui`, `client_boss_phase_readability`,
  `boss_portrait_panel`, `boss_reward_panel`.
- [x] 4.5 Layout pass: apply `HudLayout` in the remaining `_position_panel`/`_sync_position` sites; confirm no
  overlap with level label, status-effect bar, companion bar, quest/elite trackers by capture at 1280x720 and
  1920x1080. Fix by adjusting `HudLayout`, not by editing those panels.
- [x] 4.6 Maintainability: `make maintainability` (main.gd must be <= 6645; no new `helpers=globals()`; new
  files <= 600). Verify each new script imports/tests without `main.gd`.

### Phase 5 — Captures, performance, scenario (unblocked parts)

- [x] 5.1 New client scenario `<next free>_hud_polish_visual.json` using `capture_frame` for: globes after
  damage/heal (use a lab/`debug_progression` setup, no incidental navigation per the movement contract),
  hotbar populated, minimap compact, boss bar (reuse the boss-floor lab fixture of
  `client_boss_health_bar_ui`). Register as needed (`scenario-catalog.md`, movement audit). Check:
  `pytest tools/test_scenario_movement_audit.py`; `make bot-client SCENARIO=<id> HEADLESS=0`.
- [x] 5.2 Showme: add a `hud` focus/suite (fixture states: globes 100/50/low/0, hotbar empty/populated,
  minimap compact/fullscreen, boss bar 100/50/10 + phase + reward) if the showme scene pattern (see
  `heal-rain` focus) supports Control fixtures; otherwise document the deferral in as-built and rely on 5.1.
  Check: `pytest tools/test_regen_screenshots.py`; `make regen-screenshots SUITE=hud`.
- [x] 5.3 "After" captures at 1280x720 and 1920x1080, Balanced and Performance tiers, into
  `.artifacts/v517/after/`; copy the curated set into `docs/as-built/assets/v517/` per the repo's as-built
  convention. Inspect directly for overlap, legibility, blank frames.
- [x] 5.4 Frame pacing: re-run the 0.2 probe x3 after the change; compare p50/p95 and process_ms; record
  limits (single host, vsync cap). Any measurable regression blocks completion: investigate redraw frequency
  and node count first.

### Phase 6 — BLOCKED(v514): adopt UiTheme

Unblock when the exact integrated v514 (UiTheme, `ui_theme.v0.json`, schema, `validate_shared` checks) is in
this worktree and its focused checks pass.

- [x] 6.1 BLOCKED(v514): confirm token names/groups with the integrated v514; if per-surface groups are
  allowed, add additive `hud.*` tokens (globe colors, frame/border/radius, slot states, minimap frame, boss
  bar frame) to `shared/assets/ui_theme.v0.json` and its schema; add `validate_shared` cross-check if two
  names/units appear. Check: `make validate-shared`.
- [x] 6.2 BLOCKED(v514): switch `HudStyle` getters to read `UiTheme` (`ensure_loaded()` singleton pattern, no
  duplicated file loading); delete the interim literals and the header comment. Check: a test that changing a
  token in a temp `ui_theme` fixture changes the produced `StyleBox`/globe color (configurability proof).
- [x] 6.3 BLOCKED(v514): re-run Phase 1–5 checks; regenerate "after" captures so they reflect the shared theme;
  re-run the frame-pacing comparison (theme lookup must not run per frame).

### Phase 7 — As-built and handoff

- [x] 7.1 Write `docs/as-built/v517_hud-hotbar-polish.md`: what shipped, before/after capture links,
  pacing numbers and limits, debug-key stability list, main.gd line count, v514-blocked items still open (if
  any), open questions resolved.
- [x] 7.2 Update `docs/CODEMAP.md` rows (HUD/minimap/boss) and the plan checkboxes. Lifecycle index/`PROGRESS.md`
  are left to the coordinator.
- [x] 7.3 Handoff to the coordinator: base commit, v514 state used, full changed/deleted/untracked list,
  ignored evidence to preserve (`.artifacts/v517/**`), focused commands and results, unmet criteria, and the
  shared-file conflict list below.

## Execution notes (Phases 0-5, 2026-10-01)

- Spec/plan path correction: the smoke registry is `scripts/client_smoke.sh`, not `client/scripts/`.
- Captures use a dedicated `client/scripts/showme/showme_hud_capture.gd` (the `hud` focus in
  `skills/showme/scripts/render_focus.py` now routes to it and the old `_setup_hud` in the grandfathered
  `visual_capture.gd` was removed, shrinking it). `make regen-screenshots SUITE=hud` produces 8 frames.
- Deviations from the spec text, to confirm at review: (a) HUD has no quality-tier dependent code, so
  captures are not duplicated per Balanced/Performance (the real-camera scenario ran with the fixture's
  default tier); (b) the hotbar has no stack count (potions are single items) and no drag-over visual
  (no hover hook exists on the slot button); (c) the dead `_build_attack_recovery` strip of
  `PlayerHealthBar` (never built) was dropped; debug-state keys are unchanged.
- Pacing (v497 probe, balanced, 1920x1080, 3 runs each): see `docs/as-built` when written; raw runs under
  `.artifacts/v517/pacing-{before,after}/`.

- Phase 6/7 (2026-10-01): v514 integrated. Tokens are `hud_*` colors/spacing/frames in `ui_theme.v0.json` (schema
  unchanged; the pressed slot state is stored as the schema state `selected`). Scenario renumbered 113 (v512 holds
  111/112). Final pacing used an interleaved A/B against a temporary detached base worktree (removed). See
  `docs/as-built/v517_hud-hotbar-polish.md`.

## Final focused verification (slice gate; not `make ci`)

```bash
make client-unit
make maintainability
make validate-shared                      # after any ui_theme change (Phase 6)
.venv/bin/pytest tools/test_scenario_movement_audit.py tools/test_regen_screenshots.py
for s in use_potion_hotbar client_boss_health_bar_ui client_boss_phase_readability boss_portrait_panel \
         boss_reward_panel discovery_minimap_toggle full_screen_map_overlay map_transparency_setting \
         minimap_points_of_interest elite_objective_hud elite_objective_minimap_pin \
         client_skill_points_and_magic_bolt <hud_polish_visual_id>; do
  make bot-client SCENARIO=$s HEADLESS=1 || break
done
make bot-client SCENARIO=<hud_polish_visual_id> HEADLESS=0     # real-renderer captures
make bot-client SCENARIO=dungeon_frame_pacing_probe HEADLESS=0  # paired before/after
make regen-screenshots SUITE=hud                                 # if 5.2 landed
git diff --check
```

The coordinator runs one combined `make ci` after every accepted slice is integrated. Do not run it here.

## Shared-file conflicts (for integration)

| File | Contributors | Resolution |
|---|---|---|
| `client/scripts/boss_health_bar.gd` | v517 (style section), v512 (hooks/state, if any) | D1 boundary; v512 first, then v517 rebased; compare final file against both handoffs |
| `client/scripts/boss_visuals_controller.gd`, banner files | v512 only | v517 does not touch |
| `shared/assets/ui_theme.v0.json` + schema, `tools/validate_shared.py` | v514 (owner), v515, v516, v517 (HUD tokens) | additive groups; diff final merged file against each handoff; validate |
| `scripts/client_smoke.sh` | v512, v514, v515, v516, v517 | append-only registration; verify every added gate survives |
| `docs/CODEMAP.md` | all | per-row edits; reconcile manually |
| `docs/progress/scenario-catalog.md`, `scenario-movement-audit.tsv`, `tools/bot/ci_pack.json` | any slice adding a scenario | scenario number collisions: pick next free at integration time |
| `tools/showme/screenshot_catalog.py` | v517 (new `hud` suite); possibly v515/v516 | additive suites |
| `client/scripts/character_bar.gd` | v517 (frame), v516 (must not restyle) | confirm with v516 |
| `client/scripts/main.gd` | none expected | verify zero growth at integration |
| `.maintainability/file-size-baseline.tsv` | none expected | n/a |
