# v446 Plan — Wall Occlusion Fade

Status: Ready for implementation  
Goal: Fade box wall meshes that block the camera view of the hero and rendered monsters.  
Architecture: Client-only segment tests on the XZ plane against `current_wall_layout`; `WallRenderer`
applies per-mesh material alpha from data-driven tuning. No server or protocol changes.  
Tech stack: shared presentation JSON, Godot client, Python client bot.

## Baseline and shortcut decision

Reuses `line_of_sight_blocker_lab` for bot proof. Adopt `wall_occlusion_presentation.v0.json`;
borrow `WallRenderer` mesh registry; reject external addons.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/assets/wall_occlusion_presentation.v0.json` | Tuning data |
| Create | `shared/assets/wall_occlusion_presentation.v0.schema.json` | Schema |
| Create | `client/scripts/wall_occlusion_presentation_loader.gd` | Loader singleton |
| Create | `client/scripts/wall_occlusion_fade.gd` | Segment tests + sync |
| Modify | `client/scripts/wall_renderer.gd` | Mesh registry + alpha apply |
| Modify | `client/scripts/main.gd` | Per-frame sync + debug state |
| Modify | `client/scripts/bot_assertion_handlers.gd` | `assert_wall_occlusion` |
| Modify | `client/scripts/bot_step_catalog.gd` | Register step type |
| Create | `tools/bot/scenarios/client/98_wall_occlusion_fade.json` | Extended visual proof |
| Create | `client/tests/test_wall_occlusion_fade.gd` | Unit tests |
| Modify | `scripts/client_smoke.sh` | Register unit test |
| Create | `docs/as-built/v446_wall-occlusion-fade.md` | As-built |

## Maintenance ratchet

Hotspot files touched:
- [ ] `client/scripts/main.gd` — narrow wiring only (~10 lines)
- [ ] `client/scripts/wall_renderer.gd` — stays within baseline

Decision: extract `wall_occlusion_fade.gd` as new focused module.

Verification:
```bash
make maintainability
```

## Task 1 — Shared presentation catalog

- [x] Step 1.1: Add `wall_occlusion_presentation.v0.json` + schema
- [x] Step 1.2: Add `WallOcclusionPresentationLoader`

## Task 2 — Occlusion logic + wall renderer

- [x] Step 2.1: Implement segment/AABB helpers and fade resolution in `wall_occlusion_fade.gd`
- [x] Step 2.2: Register box wall meshes in `wall_renderer.gd` and apply alpha

```bash
make client-unit
```

## Task 3 — Main integration + debug

- [x] Step 3.1: Wire sync in `main.gd` `_process`
- [x] Step 3.2: Expose `wall_occlusion` debug state for bot

## Task 4 — Bot scenario + unit tests

- [x] Step 4.1: `98_wall_occlusion_fade.json` on `line_of_sight_blocker_lab`
- [x] Step 4.2: `assert_wall_occlusion` handler + catalog entry
- [x] Step 4.3: `test_wall_occlusion_fade.gd` + `client_smoke.sh`

```bash
HEADLESS=1 make bot-visual scenario=98_wall_occlusion_fade
make client-unit
```

## Task 5 — Lifecycle docs

- [x] Update `PROGRESS.md`, lifecycle row, as-built

## Final verification

```bash
make maintainability
make client-unit
HEADLESS=1 make bot-visual scenario=98_wall_occlusion_fade
```
