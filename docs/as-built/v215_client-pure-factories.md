# v215 As-Built — Client Pure Factory Extraction

Date: 2026-06-16

## Shipped

- Extracted compile-time client constants into `client/scripts/client_constants.gd`.
- Extracted ground and wall texture/material generation into `GroundWallFactory`.
- Extracted wall layout normalization/rendering into `WallRenderer`.
- Extracted loot node and ground item visual construction into `LootNodeFactory`.
- Updated factory-focused and existing client presentation tests to preload extracted helpers directly.

## Proof

- `godot --headless --path client --script res://tests/test_factories.gd`
- `make client-unit`

## Notes

- The extracted helpers preload their direct dependencies instead of relying on Godot editor class registration, so headless direct-preload tests are quiet.
- `main.gd` remains the scene coordinator and owns label/filter state; pure construction code now lives outside the coordinator.
