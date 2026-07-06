# v446 Spec — Wall Occlusion Fade

**Status:** Complete  
**Date:** 2026-07-06  
**Codename:** wall-occlusion-fade

---

## Purpose

When solid dungeon/town wall geometry sits between the gameplay camera and the local hero (or other
rendered combat entities), fade those wall meshes so the player can still read positions and combat.
Fog-of-war (v253/v255/v331) hides **information** behind walls; this slice fixes **mesh occlusion**
blocking the camera view. Presentation-only — server visibility rules are unchanged.

## Non-goals

- No server gameplay visibility, fog radius, monster filtering, aggro, protocol, replay, or golden
  changes.
- No revealing fog-hidden monsters through faded walls (only entities the client already renders).
- No fade for water, holes, rubble, rocks, or column obstacle meshes in v1 (box `wall` + `wood`
  only).
- No closed-door interactable fade (wall layout rows only).
- No user-facing settings toggle (fixed presentation tuning; setting can follow v266 pattern later).
- No production wall art, imported shaders, or Godot addons.

## Acceptance criteria

### Shared presentation data

- `shared/assets/wall_occlusion_presentation.v0.json` + schema define client-only tuning:
  `faded_alpha`, `opaque_alpha`, `segment_inflate`, `move_epsilon`, and
  `min_rebuild_interval_frames`.
- `WallOcclusionPresentationLoader` (`class_name` + `ensure_loaded()`) loads the catalog.

### Client occlusion fade

- `WallOcclusionFade` tests camera→entity segments on the XZ plane against authoritative
  `current_wall_layout` box walls (`kind` absent, `wall`, or `wood`).
- Matching walls fade to `faded_alpha`; non-matching walls restore `opaque_alpha`.
- Targets: local hero plus living monsters with rendered nodes in the client entity map.
- Recompute is throttled when camera, targets, and wall layout are unchanged within epsilon.
- `wall_layout_update` / level transitions reset fade state cleanly.

### Bot / tests

- Extended client scenario on `line_of_sight_blocker_lab` asserts `faded_wall_count_min >= 1` and
  `min_faded_alpha_max` below opaque.
- `client/tests/test_wall_occlusion_fade.gd` covers segment/AABB intersection and fade target
  selection helpers.
- `make client-unit` and `make maintainability` pass.

## Scope and likely files

| Area | Path |
|------|------|
| Shared | `shared/assets/wall_occlusion_presentation.v0.json`, schema |
| Loader | `client/scripts/wall_occlusion_presentation_loader.gd` |
| Logic | `client/scripts/wall_occlusion_fade.gd` |
| Walls | `client/scripts/wall_renderer.gd` |
| Integration | `client/scripts/main.gd` |
| Bot | `tools/bot/scenarios/client/98_wall_occlusion_fade.json`, assertion handlers |
| Tests | `client/tests/test_wall_occlusion_fade.gd`, `scripts/client_smoke.sh` |
| Docs | plan, as-built, lifecycle on `/finish` |

### Asset / plugin decision

| Choice | Decision |
|--------|----------|
| `wall_occlusion_presentation.v0.json` | **Adopt** |
| `hero_visibility_field.gd` AABB patterns | **Borrow** (new segment helper, no import cycle) |
| `WallRenderer` mesh registry | **Borrow** |
| External occlusion addons | **Reject** |

## Test and bot proof

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=98_wall_occlusion_fade
make maintainability
```

Manual: `make play` — stand on either side of corridor walls in isometric mode.

## Open questions

| # | Item | Resolution |
|---|------|------------|
| Q-1 | Hero-only vs hero+monsters | **Resolved:** hero + rendered living monsters |
| Q-2 | Partial fade vs cutaway | **Resolved:** partial material alpha |
| Q-3 | Segment test plane | **Resolved:** XZ plane (matches fog wall layout) |

## ADR alignment

- **ADR-0001 D2:** presentation-only.
- **ADR-0007:** animation unaffected.
- **ADR-0008:** reuses rectangular wall layout; no new server occluder semantics.
