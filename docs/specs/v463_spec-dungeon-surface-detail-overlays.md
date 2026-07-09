# v463 Spec - Dungeon Surface Detail Overlays

Status: Implemented
Date: 2026-07-09
Codename: dungeon-surface-detail-overlays
Baseline: v462 `rounded-dungeon-corners`

## Purpose

Make generated dungeon floors and walls feel less flat by adding a presentation-only detail layer
on top of the existing procedural material path. Floors should gain readable signs, scratches, and
ritual-style marks at mixed scales, while eligible walls should gain artifacts such as plaques,
cracks, or inset details that make the environment read as more authored.

This slice is intentionally additive. The authoritative server keeps the same wall layout, floor
bounds, collision, reachability, and protocol. The client only renders more detail.

## Non-goals

- No server, protocol, collision, or pathfinding changes.
- No imported plugin or addon dependency.
- No runtime-downloaded art.
- No town presentation pass.
- No biome-wide lighting or fog overhaul.
- No non-rectangular geometry or collision changes.

## Acceptance Criteria

- Generated dungeon floors render at least two families of floor detail overlays with visible size
  variation on normal dungeon floors.
- Eligible dungeon walls render at least two families of wall detail overlays/artifacts without
  changing authoritative wall positions or occlusion behavior.
- Overlay placement is client-only and leaves movement, stairs reachability, wall counts, and
  authoritative layouts unchanged.
- The detail system is data-driven through a shared client-presentation config rather than
  hardcoded one-off placements.
- Town surfaces remain unchanged.
- Focused client tests prove eligible detail roots attach on dungeon floors, stay off in town, and
  generate mixed-size floor signs plus wall artifacts.
- A dedicated client scenario proves the presentation path on a generated dungeon floor and remains
  suitable for `make bot-visual`.

## Scope and Likely Files

- Client presentation:
  - `client/scripts/dungeon_surface_detail_presentation.gd`
  - `client/scripts/dungeon_surface_detail_loader.gd`
  - `client/scripts/main.gd`
- Shared client-only config:
  - `shared/assets/dungeon_surface_detail_presentation.v0.json`
- Client tests:
  - `client/tests/test_factories.gd`
- Bot / visual proof:
  - `tools/bot/scenarios/client/101_dungeon_surface_detail_overlays.json`
- Docs:
  - `docs/plans/v463_2026-07-09-dungeon-surface-detail-overlays.md`
  - `docs/as-built/v463_dungeon-surface-detail-overlays.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

### Asset / plugin decision

- Adopt: existing procedural ground/wall presentation path, room tint segmentation, and client bot
  rollout scenarios already in repo.
- Borrow: no external assets in the first pass; keep outside CC0 texture sources as a future option
  only if the procedural pass proves too weak.
- Reject: external dungeon kits, shader addons, and runtime asset download.

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
HEADLESS=1 make bot-client scenario=dungeon_surface_detail_overlays
HEADLESS=1 make bot-client scenario=wall_floor_dungeon_rollout
```

Expected visual verification command:

```bash
make bot-visual scenario=dungeon_surface_detail_overlays
```

## Open Questions and Risks

- Placement risk: strong floor signs can look too decorative if every room gets the same motif.
  Default: vary motifs by room cell and wall source while keeping density modest.
- Truthfulness risk: wall artifacts must not imply openings or interactables that do not exist.
- Maintainability risk: keep the new logic out of `wall_renderer.gd` and `main.gd` beyond a single
  sync call.
