# v77 As-built — Main Config Derived Gameplay

Status: Complete

## Shipped

- Server-loaded combat cadence now uses `main_config.gameplay.base_attack_interval_ticks` as the authoritative value.
- Player direct movement and auto-nav movement now use `main_config.gameplay.base_movement_speed`.
- Shared validation no longer requires `combat.v0.json` or `navigation.v0.json` to mirror those two main-config values.
- Focused Go tests load a temp rules directory with only `main_config.v0.json` changed and prove attack interval and movement distance follow that file.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestMainConfig|TestMovement|TestLoadRules'`
- `make ci`
