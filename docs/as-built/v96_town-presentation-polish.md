# v96 As-Built — Town presentation polish

Date: 2026-06-12
Status: Complete

## What Shipped

- Reworked the `dungeon_levels` town service layout so the stairs, teleporter, vendor, mystery seller, stash, bishop, and market board sit in a wider ring around the town center.
- Added a client-only town preview composition with two procedural wood cabins and a central campfire.
- Improved the level-0 ground texture from a simple grass grid to a noisier grass/dirt surface.
- Added `$showme --focus town` for fast visual captures.

## Proof

- `make validate-shared`
- `godot --headless --path client --script res://tests/test_item_visuals.gd`
- `python3 skills/showme/scripts/render_focus.py --focus town`
- `make client-unit`
- `make maintainability`
- `make ci`

Latest reviewed capture:

```text
.artifacts/showme/20260612-103004-town.png
```

## Scope Limits

- Cabins and campfire are presentation-only and do not affect collision, pathing, service reachability, protocol, or server authority.
- Production town art, imported building assets, collision-aware decorations, ambient NPC movement, fire animation polish, audio, and lighting pass remain deferred.

## Maintainability Note

This slice intentionally updated the grandfathered file-size baseline for `client/scripts/main.gd`
and `skills/showme/scripts/visual_capture.gd`. The added code is focused procedural presentation
and capture setup; extracting a new presentation module is deferred until more town/environment
props accumulate.
