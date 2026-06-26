# v340 Spec — Depth Combat Pressure

Status: Complete
Date: 2026-06-25
Codename: `depth-combat-pressure`

## Purpose

Make deeper dungeon floors feel more dangerous via data-driven `monster_depth_scaling` tuning and a focused depth-pressure server test.

## Acceptance Criteria

- `hp_per_depth`, `damage_per_depth`, and cooldown multiplier updated in shared rules.
- `monster_rarity` golden expected values updated for the new scaling curve.
- Go test proves deeper generated monsters are tougher than shallow ones.
