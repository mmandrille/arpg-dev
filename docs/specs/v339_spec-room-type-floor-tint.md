# v339 Spec — Room-Type Floor Tint

Status: Complete
Date: 2026-06-25
Codename: `room-type-floor-tint`

## Purpose

Give partitioned dungeon rooms distinct floor tint overlays so combat, corridor, rest, and treasure areas read more clearly after v330 room dividers.

## Acceptance Criteria

- Data-driven archetype tints in `shared/assets/dungeon_room_presentation.v0.json`.
- Client overlays sync from `room_divider` walls and treasure chest interactables.
- Headless factory test proves tint root and treasure tint.
