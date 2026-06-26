# v348 Plan — Forward Plus Renderer

Status: Ready for implementation
Goal: Set `forward_plus` as the interactive Godot default while headless CI/bot paths keep `gl_compatibility` via explicit CLI overrides.
Architecture: Project default moves to modern 3D pipeline for `make play` / windowed `make bot-visual`. A shared `GODOT_HEADLESS_FLAGS` constant (`--headless --rendering-method gl_compatibility`) is sourced by launch scripts so CI never depends on forward_plus GPU features. No protocol/server/shared changes. Adopt/borrow/reject: **reject** external render plugins; **borrow** existing fog overlay, depth lighting, GLB materials.
Tech stack: Godot 4 client, bash launch scripts, client bot scenarios for visual regression.

## Baseline and shortcut decision

Builds on v347 dungeon-render-performance (fog cache, graphics quality preset). v347 deferred renderer migration to this slice. Interactive play uses project default; automated paths override renderer only at launch.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/project.godot` | `renderer/rendering_method="forward_plus"` |
| Create | `scripts/godot_ci_flags.sh` | Shared headless CI flags with gl_compatibility override |
| Modify | `scripts/client_smoke.sh` | Use shared headless flags for gates + import |
| Modify | `scripts/bot_client.sh` | Headless scenario runs use shared flags |
| Modify | `scripts/bot_visual.sh` | Headless replay + import use shared flags |
| Modify | `make/shared.mk` | `gen-anims` headless import/script |
| Modify | `scripts/test_all.sh` | Headless bot-visual flags include compatibility renderer |
| Create | `docs/as-built/v348_forward-plus-renderer.md` | FPS sample + regression notes |
| Modify | `docs/specs/v348_spec-forward-plus-renderer.md` | Mark complete |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Lifecycle close-out |

## Maintenance ratchet

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd` — not touched
- [ ] Other over-limit file: none expected

Decision:
- [x] Defer extraction — launch-script flag helper is new focused file under 600 lines

Verification:
```bash
make maintainability
```

## Task 1 — Project default renderer

Files:
- Modify: `client/project.godot`

- [x] Step 1.1: Set `renderer/rendering_method="forward_plus"`.

## Task 2 — Headless CI flag helper

Files:
- Create: `scripts/godot_ci_flags.sh`
- Modify: `scripts/client_smoke.sh`, `scripts/bot_client.sh`, `scripts/bot_visual.sh`, `make/shared.mk`, `scripts/test_all.sh`

- [x] Step 2.1: Add `GODOT_HEADLESS_FLAGS` defaulting to `--headless --rendering-method gl_compatibility`.
- [x] Step 2.2: Source helper in launch scripts; replace bare `--headless` on CI/bot/test paths.
```bash
make client-unit
```

## Task 3 — Client bot visual regressions

Files:
- (no scenario JSON changes)

- [x] Step 3.1: Run fog/door client bot scenarios headless.
```bash
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
HEADLESS=1 make bot-visual scenario=73_door_fog_toggle
```

## Task 4 — Lifecycle docs

Files:
- Create: `docs/as-built/v348_forward-plus-renderer.md`
- Modify: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, spec status

- [x] Step 4.1: Record renderer decision, headless override pattern, FPS note (if perf probe run).
- [x] Step 4.2: Update lifecycle row.

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay`
- [x] `HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask`
- [x] `HEADLESS=1 make bot-visual scenario=73_door_fog_toggle`

Deferred: `ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe` — manual extended check; capture in as-built when host allows.
