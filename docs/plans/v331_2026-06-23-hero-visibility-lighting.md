# v331 Plan — Hero Visibility Lighting

Status: Ready for implementation  
Goal: Unified world-space fog darkness + hero light falloff for all camera modes.  
Architecture: Extract LOS/occluder math into `hero_visibility_field.gd`; drive falloff and ambient
suppression from `fog_presentation.v0.json`; replace screen-radius shader with per-pixel ground-plane
visibility sampling; keep Polygon2D LOS shadows; enable compositor in perspective modes.  
Tech stack: shared JSON, Godot 4 GDScript client, Python/Godot bot scenarios. No server changes.

## Baseline and shortcut decision

Builds on v253–v264 fog stack and v329 camera modes. **Adopt** `fog_presentation.v0.json` and
`hero_visibility_field.gd`. **Borrow** existing `FogOfWarOverlay` shadow polygons. **Reject** external
fog plugins/art.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/assets/fog_presentation.v0.json` | Falloff, shadow reach, ambient suppression, organic edge per mode |
| Create | `shared/assets/fog_presentation.v0.schema.json` | Schema |
| Create | `client/scripts/fog_presentation_loader.gd` | Static loader |
| Create | `client/scripts/hero_visibility_field.gd` | Occluder normalize, LOS shadow polygons |
| Modify | `client/scripts/fog_of_war_overlay.gd` | World-space shader, loader, field delegation |
| Modify | `client/scripts/dungeon_depth_lighting.gd` | Fog ambient suppression helper |
| Modify | `client/scripts/main.gd` | All-camera fog active; ambient hook |
| Modify | `client/tests/test_fog_of_war_overlay.gd` | World falloff + mode flags |
| Modify | `tools/bot/scenarios/client/67_*.json`, `83_*.json` | Updated assertions |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Optional new debug keys |

## Maintenance ratchet

Hotspot files:
- [ ] `client/scripts/fog_of_war_overlay.gd` (~520) — extract field module; net shrink
- [ ] `client/scripts/main.gd` — narrow wiring only

Decision: extract `hero_visibility_field.gd` this slice.

## Task 1 — Shared fog presentation catalog

- [x] Create schema + JSON
- [x] Create `fog_presentation_loader.gd`

## Task 2 — Hero visibility field extraction

- [x] Create `hero_visibility_field.gd` with occluder + shadow polygon APIs
- [x] Unit-test via existing overlay tests

## Task 3 — World-space compositor shader

- [x] Replace radial screen-distance shader with ground-plane world distance + falloff
- [x] Wire loader tuning; keep LOS polygon layer

## Task 4 — Main integration

- [x] Remove isometric-only fog gate
- [x] Ambient suppression when fog active in dungeons

## Task 5 — Tests and bot proof

- [x] Update `test_fog_of_war_overlay.gd`
- [x] Update scenarios 67, 68, 83
```bash
make client-unit
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
HEADLESS=1 make bot-visual scenario=83_camera_mode_setting
make bot scenario=92_fog_of_war_radius
make maintainability
```

## Task 6 — Lifecycle docs and CI

- [ ] as-built + lifecycle on `/finish`
```bash
make ci
```
