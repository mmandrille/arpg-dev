# v325 As Built - Town Ambient Life

Date: 2026-06-23

## What Shipped

- `TownAmbientLife` attaches static NPC silhouettes to town ground on level 0.
- Wired through `GroundWallFactory.update_ground_material` and town preview factory.

## Proof

```bash
godot --headless --path client --script res://tests/test_look_and_feel_polish.gd
```
