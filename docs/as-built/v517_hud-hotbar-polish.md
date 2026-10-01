# v517 — HUD / Hotbar Polish (as built)

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)

- **Base:** `425b9ae4` + integrated v514 (UiTheme). Display only; no protocol, server, rules, golden or ADR change.
- **Spec / plan:** [spec](../specs/v517_spec-hud-hotbar-polish.md), [plan](../plans/v517_2026-10-01-hud-hotbar-polish.md)

## What shipped

- **Globes:** `client/scripts/hud_globe.gd` (`HudGlobe`): `_draw`-only liquid globe, redraws only when ratio/label/color change, no `_process`. `PlayerHealthBar` now hosts a health and a mana globe at the bottom corners plus a name/level strip; its public API and `get_debug_state()` keys are unchanged (contract test `test_player_health_bar.gd` passed on the base code first). The never-built attack-recovery strip was dropped.
- **Layout:** `hud_layout.gd` (`HudLayout`, pure static rect math) feeds the hotbar, XP bar, character/skill slots, globes, name strip, minimap and boss bar positions; `test_hud_layout.gd` proves no overlap/in-bounds at 1024x576 .. 2560x1440 and legacy slot-cluster positions.
- **Style:** `hud_style.gd` (`HudStyle`) has no literals: every color, spacing and frame resolves from `hud_*` tokens added to `shared/assets/ui_theme.v0.json` (54 colors, 22 spacing entries, 9 frame recipes, one per line; schema unchanged, pressed slot look uses the schema's `selected` state). `boss_bar_frame.gd` builds the boss panel/reward/trough/gloss/portrait frame from the same tokens. `test_hud_style.gd` loads a temp catalog and proves token edits change the produced StyleBoxes/colors.
- **Hotbar / slots / minimap / boss bar:** empty, filled, hover, pressed, disabled slot looks; shared bronze frame with shadow; minimap frame only (map drawing untouched); boss bar frame/trough/fill gloss/phase colors. HP, phase, reward logic untouched.
- **Capture tooling:** `client/scripts/showme/showme_hud_capture.gd` + `hud` focus routing in `skills/showme/scripts/render_focus.py` + `hud` suite in `tools/showme/screenshot_catalog.py` (8 captures). The old `_setup_hud` was removed from the grandfathered `visual_capture.gd` (1259 -> 1248 lines). New client scenario `113_hud_polish_visual.json` (`hud_polish_visual`).
- `main.gd` unchanged (6645 lines).

## Evidence

- Captures in `docs/as-built/assets/v517/`: `hud-{full,low,empty}-{before,after}.png` (1920x1080 fixture, before = base code, after = themed final) and `real-camera-boss-floor.png` (real play camera, boss floor lab). Full set under ignored `.artifacts/v517/{before,after,regen}/` (1280x720 and 1920x1080, full/half/low/empty).
- Pacing (v497 `dungeon_frame_pacing_probe`, balanced, 1920x1080, Apple M4 Pro, Godot 4.7.2). Clean interleaved A/B (base worktree vs v517, 2 runs each, same session): frame interval p50 16.66 vs 16.66 ms, p95 17.20-17.21 vs 17.21-17.22 ms; process p50 17.31-17.38 vs 17.34-17.35 ms; draw calls 108 vs 124 (+16); primitives about 126.4k vs 131.7k (+4%); rendered objects 405 vs 427. Three further themed runs: p95 17.13-17.23 ms. Earlier non-interleaved before/after numbers (`.artifacts/v517/pacing-{before,after}`) showed host noise from concurrent agents and are not used for the verdict. Limits: vsync-capped (60 Hz) so sub-vsync cost is not visible; weaker GPUs untested.
- Checks run: `make client-unit` (incl. new `test_hud_layout`, `test_hud_globe`, `test_hud_style`, `test_player_health_bar`), `make maintainability`, `tools/validate_shared.py`, pytest for `test_validate_ui_theme`, `test_scenario_movement_audit`, `test_regen_screenshots`, 12 existing HUD/boss/minimap/hotbar client scenarios plus `hud_polish_visual` (real renderer). No `make ci`.

## Deviations and limits

- No Balanced/Performance capture duplication (the HUD has no tier-dependent code); no stack count (no quantity in the hotbar model) and no drag-over slot state (no hover hook on the slot button).
- Remaining non-token colors: skill/character bar flash and badge colors, item drag preview label and the boss portrait content drawing (v512 owns portrait content).
- The HUD adds 16 draw calls; merging globe draw commands is a possible future saving.
