# v338 Spec — Wall/Floor Dungeon Rollout

Status: Ready
Date: 2026-06-25
Codename: `wall-floor-dungeon-rollout`

## Purpose

Extend the v308 wall/floor shader polish from lab-only bot proof to the standard `dungeon_levels`
player route and complete the material pass on dungeon ceiling and obstacle meshes.

## Non-goals

- No server, protocol, navigation, fog, or combat changes.
- No imported assets or ShaderMaterial replacement.
- No town presentation changes.

## Acceptance Criteria

- Dungeon ceiling materials use generated normal maps like dungeon walls.
- Rock/column/rubble obstacle materials use generated normal maps.
- `room_divider` walls get a distinct tint from generated/perimeter walls.
- Focused factory tests prove ceiling and obstacle normal mapping.
- A client bot scenario descends from town via `dungeon_levels` and asserts generated dungeon walls
  on floor -1.

## Scope and Files

- `client/scripts/ground_wall_factory.gd`
- `client/scripts/wall_renderer.gd`
- `client/tests/test_factories.gd`
- `tools/bot/scenarios/client/79_wall_floor_dungeon_rollout.json`
- Docs lifecycle/as-built

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
HEADLESS=1 SCENARIO=wall_floor_dungeon_rollout make bot-client
```
